package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"

	"sentigraph-analyst/internal/app"
	"sentigraph-analyst/internal/models"
	"sentigraph-analyst/internal/repository"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	godotenv.Load()

	cfg := models.Config{
		RedisURL:    os.Getenv("REDIS_URL"),
		OllamaURL:   os.Getenv("OLLAMA_URL"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		DiscordURL:  os.Getenv("DISCORD_WEBHOOK_URL"),
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	db := repository.NewDatabase(context.Background(), cfg.DatabaseURL)
	analyst := app.NewApp(cfg, db, rdb)

	dbChan := make(chan models.MarketState, 1000)
	newsChan := make(chan models.NewsItem, 100)

	go db.BatchInsertWorker(context.Background(), dbChan)
	go analyst.HandleMarketData(context.Background(), dbChan)

	for i := 0; i < 3; i++ {
		go func() {
			for item := range newsChan {
				analyst.AnalyzeSentiment(context.Background(), item)
			}
		}()
	}

	go func() {
		pubsub := rdb.Subscribe(context.Background(), "market:news")
		for msg := range pubsub.Channel() {
			var item models.NewsItem
			if err := json.Unmarshal([]byte(msg.Payload), &item); err == nil {
				title := strings.ToLower(item.Title)
				switch {
				case strings.Contains(title, "eth"): item.Symbol = "eth"
				case strings.Contains(title, "sol"): item.Symbol = "sol"
				default: item.Symbol = "btc"
				}
				newsChan <- item
			}
		}
	}()

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover(), middleware.Logger(), middleware.CORS())

	e.GET("/ws", func(c echo.Context) error {
		return analyst.UpgradeWS(c.Response(), c.Request())
	})

	e.GET("/api/history/:symbol", func(c echo.Context) error {
		history, err := db.GetHistory(c.Request().Context(), c.Param("symbol"), 1000)
		if err != nil { return c.JSON(500, map[string]string{"err": err.Error()}) }
		return c.JSON(200, history)
	})

	port := cfg.Port
	if port == "" { port = "8080" }
	e.Start(":" + port)
}
