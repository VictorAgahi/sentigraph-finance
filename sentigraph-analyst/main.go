package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	RedisURL    string
	OllamaURL   string
	DatabaseURL string
	Port        string
}

type MarketState struct {
	Price     string    `json:"price"`
	Sentiment float64   `json:"sentiment"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SafeState struct {
	sync.RWMutex
	MarketState
}

type NewsItem struct {
	Title string
}

var (
	state = &SafeState{
		MarketState: MarketState{
			Price:     "0.00",
			Sentiment: 0.0,
			UpdatedAt: time.Now(),
		},
	}
	newsStream = []string{
		"Bitcoin hits new all-time high as institutional interest surges",
		"US Federal Reserve hints at potential interest rate hikes",
		"Tech stocks rally after strong earnings reports from major players",
		"Global supply chain disruptions continue to impact trade volume",
		"New regulations for cryptocurrency markets proposed by EU commission",
		"Oil prices stabilize after sudden drop in global production",
	}
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func LoadConfig() Config {
	godotenv.Load()

	cfg := Config{
		RedisURL:    os.Getenv("REDIS_URL"),
		OllamaURL:   os.Getenv("OLLAMA_URL"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}

	if cfg.RedisURL == "" || cfg.OllamaURL == "" || cfg.DatabaseURL == "" {
		slog.Error("Missing required environment variables")
		os.Exit(1)
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg
}

func ollamaWorker(id int, ollamaURL string, newsChan <-chan NewsItem) {
	client := &http.Client{Timeout: 15 * time.Second}

	for news := range newsChan {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		
		prompt := fmt.Sprintf(`### ROLE: EXPERT FINANCIAL ANALYST
### TASK: Sentiment Scoring for High-Frequency Trading
### INSTRUCTION: Analyze the macroeconomic and technical impact of the news title below.
### CRITERIA: 
- Score 1.0: Explosive bullish (institutional entry, rate cuts, ATH break)
- Score 0.0: Neutral or ambiguous
- Score -1.0: Critical bearish (regulatory crackdown, rate hikes, major hack)
### FORMAT: Return ONLY a single numeric float between -1.0 and 1.0. No text.

NEWS TITLE: %s
SCORE:`, news.Title)
		
		body, _ := json.Marshal(map[string]interface{}{
			"model":  "llama3:8b",
			"prompt": prompt,
			"stream": false,
			"options": map[string]interface{}{
				"temperature": 0.1,
				"num_predict": 10,
				"top_p":       0.9,
			},
		})

		req, _ := http.NewRequestWithContext(ctx, "POST", ollamaURL, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		slog.Info("Sending news to AI for analysis", "worker", id, "title", news.Title)
		resp, err := client.Do(req)
		if err != nil {
			slog.Error("Failed to reach Ollama", "error", err)
			cancel()
			continue
		}

		var result struct {
			Response string `json:"response"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			slog.Error("Failed to decode Ollama response", "error", err)
			resp.Body.Close()
			cancel()
			continue
		}
		resp.Body.Close()
		cancel()

		slog.Info("AI Analysis complete", "worker", id, "response", result.Response)
		score, err := strconv.ParseFloat(result.Response, 64)
		if err == nil {
			state.Lock()
			state.Sentiment = score
			state.UpdatedAt = time.Now()
			state.Unlock()
			slog.Info("Sentiment updated", "score", score)
		} else {
			slog.Warn("Could not parse AI score", "response", result.Response)
		}
	}
}

func startNewsTicker(newsChan chan<- NewsItem) {
	// Send first item immediately
	newsChan <- NewsItem{Title: newsStream[0]}
	
	ticker := time.NewTicker(5 * time.Second)
	idx := 1
	for range ticker.C {
		newsChan <- NewsItem{Title: newsStream[idx%len(newsStream)]}
		idx++
	}
}

func handleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, ws)
		clientsMu.Unlock()
	}()

	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			break
		}
	}
	return nil
}

func broadcast(data interface{}) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	payload, _ := json.Marshal(data)
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, payload)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config := LoadConfig()

	newsChan := make(chan NewsItem, 10)
	dbChan := make(chan MarketState, 100)

	db := NewDatabase(context.Background(), config.DatabaseURL)
	go db.BatchInsertWorker(context.Background(), dbChan)

	for i := 1; i <= 3; i++ {
		go ollamaWorker(i, config.OllamaURL, newsChan)
	}

	go startNewsTicker(newsChan)

	go func() {
		rdb := redis.NewClient(&redis.Options{Addr: config.RedisURL})
		pubsub := rdb.Subscribe(context.Background(), "ticker:btc")
		for msg := range pubsub.Channel() {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &data); err == nil {
				if price, ok := data["p"].(string); ok {
					state.Lock()
					state.Price = price
					state.UpdatedAt = time.Now()
					current := state.MarketState
					state.Unlock()
					
					slog.Info("Price update received", "price", price, "sentiment", current.Sentiment)
					broadcast(current)
					
					select {
					case dbChan <- current:
					default:
					}
				}
			}
		}
	}()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/ws", handleWebSocket)

	e.GET("/dashboard", func(c echo.Context) error {
		state.RLock()
		defer state.RUnlock()
		return c.JSON(http.StatusOK, state.MarketState)
	})

	e.GET("/api/history", func(c echo.Context) error {
		history, err := db.GetHistory(c.Request().Context(), 100)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, history)
	})

	slog.Info("Starting SentiGraph-Analyst", "port", config.Port)
	if err := e.Start(":" + config.Port); err != nil {
		slog.Error("Server crash", "error", err)
		os.Exit(1)
	}
}
