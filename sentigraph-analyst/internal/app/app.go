package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"sentigraph-analyst/internal/models"
	"sentigraph-analyst/internal/repository"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type App struct {
	config     models.Config
	db         *repository.Database
	rdb        *redis.Client
	states     map[string]*sync.RWMutex
	market     map[string]*models.MarketState
	marketMu   sync.RWMutex
	clients    map[*websocket.Conn]struct{}
	clientsMu  sync.Mutex
	lastAlert  map[string]time.Time
	alertsMu   sync.Mutex
	upgrader   websocket.Upgrader
	httpClient *http.Client
}

func NewApp(cfg models.Config, db *repository.Database, rdb *redis.Client) *App {
	return &App{
		config:    cfg,
		db:        db,
		rdb:       rdb,
		market:    make(map[string]*models.MarketState),
		states:    make(map[string]*sync.RWMutex),
		clients:   make(map[*websocket.Conn]struct{}),
		lastAlert: make(map[string]time.Time),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}
}

func (a *App) GetState(symbol string) (*models.MarketState, *sync.RWMutex) {
	a.marketMu.Lock()
	defer a.marketMu.Unlock()

	if _, ok := a.market[symbol]; !ok {
		a.market[symbol] = &models.MarketState{
			Symbol:    symbol,
			Price:     "0.00",
			Sentiment: 0.0,
			Imbalance: 1.0,
			UpdatedAt: time.Now(),
		}
		a.states[symbol] = &sync.RWMutex{}
	}
	return a.market[symbol], a.states[symbol]
}

func (a *App) AnalyzeSentiment(ctx context.Context, item models.NewsItem) {
	symbol := item.Symbol
	if symbol == "" { symbol = "btc" }

	state, mu := a.GetState(symbol)

	prompt := fmt.Sprintf(`[INST] You are a Sentiment Analysis Engine. Return ONLY a single number between -1.0 and 1.0 (0.0 is neutral) for this title: "%s" [/INST]`, item.Title)
	
	body, _ := json.Marshal(map[string]interface{}{
		"model":   "llama3.2:1b",
		"prompt":  prompt,
		"stream":  false,
		"options": map[string]interface{}{"temperature": 0.1},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", a.config.OllamaURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		slog.Error("ollama error", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { return }

	var res struct { Response string `json:"response"` }
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil { return }

	score, err := strconv.ParseFloat(strings.TrimSpace(res.Response), 64)
	if err != nil { return }

	mu.Lock()
	state.Sentiment = score
	state.UpdatedAt = time.Now()
	current := *state
	mu.Unlock()

	a.Broadcast(current)
	a.CheckAlerts(current)
}

func (a *App) CheckAlerts(s models.MarketState) {
	if a.config.DiscordURL == "" { return }

	a.alertsMu.Lock()
	if last, ok := a.lastAlert[s.Symbol]; ok && time.Since(last) < 5*time.Minute {
		a.alertsMu.Unlock()
		return
	}
	a.lastAlert[s.Symbol] = time.Now()
	a.alertsMu.Unlock()

	var msg string
	if s.Sentiment > 0.4 && s.Imbalance > 1.3 {
		msg = fmt.Sprintf("🚀 **BULLISH CONVERGENCE: %s**\nPrice: $%s\nSentiment: %.2f\nImbalance: %.2fx", strings.ToUpper(s.Symbol), s.Price, s.Sentiment, s.Imbalance)
	} else if s.Sentiment < -0.4 && s.Imbalance < 0.7 {
		msg = fmt.Sprintf("⚠️ **BEARISH CONVERGENCE: %s**\nPrice: $%s\nSentiment: %.2f\nImbalance: %.2fx", strings.ToUpper(s.Symbol), s.Price, s.Sentiment, s.Imbalance)
	}

	if msg != "" {
		go func() {
			payload, _ := json.Marshal(map[string]string{"content": msg})
			a.httpClient.Post(a.config.DiscordURL, "application/json", bytes.NewBuffer(payload))
		}()
	}
}

func (a *App) Broadcast(data interface{}) {
	payload, _ := json.Marshal(data)
	a.clientsMu.Lock()
	defer a.clientsMu.Unlock()

	for conn := range a.clients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			conn.Close()
			delete(a.clients, conn)
		}
	}
}

func (a *App) HandleMarketData(ctx context.Context, dbChan chan<- models.MarketState) {
	pubsub := a.rdb.Subscribe(ctx, 
		"ticker:btc", "ticker:eth", "ticker:sol",
		"depth:btc", "depth:eth", "depth:sol",
		"liquidation:btc", "liquidation:eth", "liquidation:sol",
	)

	for msg := range pubsub.Channel() {
		symbol := ""
		if strings.HasPrefix(msg.Channel, "ticker:") {
			symbol = strings.TrimPrefix(msg.Channel, "ticker:")
			var data struct { Price string `json:"p"` }
			if err := json.Unmarshal([]byte(msg.Payload), &data); err == nil {
				state, mu := a.GetState(symbol)
				mu.Lock()
				state.Price = data.Price
				state.UpdatedAt = time.Now()
				current := *state
				mu.Unlock()
				a.Broadcast(current)
				dbChan <- current
			}
		} else if strings.HasPrefix(msg.Channel, "depth:") {
			symbol = strings.TrimPrefix(msg.Channel, "depth:")
			var depth struct {
				Bids [][]string `json:"b"`
				Asks [][]string `json:"a"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &depth); err == nil {
				var bidVol, askVol float64
				for _, b := range depth.Bids {
					if v, err := strconv.ParseFloat(b[1], 64); err == nil { bidVol += v }
				}
				for _, a := range depth.Asks {
					if v, err := strconv.ParseFloat(a[1], 64); err == nil { askVol += v }
				}
				state, mu := a.GetState(symbol)
				mu.Lock()
				if askVol > 0 { state.Imbalance = bidVol / askVol }
				state.UpdatedAt = time.Now()
				current := *state
				mu.Unlock()
				a.Broadcast(current)
				dbChan <- current
			}
		} else if strings.HasPrefix(msg.Channel, "liquidation:") {
			symbol = strings.TrimPrefix(msg.Channel, "liquidation:")
			var data struct {
				Order struct {
					Side  string `json:"S"`
					Price string `json:"p"`
					Qty   string `json:"q"`
				} `json:"o"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &data); err == nil {
				p, _ := strconv.ParseFloat(data.Order.Price, 64)
				q, _ := strconv.ParseFloat(data.Order.Qty, 64)
				a.Broadcast(models.LiquidationItem{
					Type:   "liquidation",
					Symbol: symbol,
					Side:   data.Order.Side,
					Price:  p,
					Amount: q,
				})
			}
		}
	}
}

func (a *App) UpgradeWS(w http.ResponseWriter, r *http.Request) error {
	ws, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil { return err }
	defer ws.Close()

	a.clientsMu.Lock()
	a.clients[ws] = struct{}{}
	a.clientsMu.Unlock()

	defer func() {
		a.clientsMu.Lock()
		delete(a.clients, ws)
		a.clientsMu.Unlock()
	}()

	for {
		if _, _, err := ws.ReadMessage(); err != nil { break }
	}
	return nil
}
