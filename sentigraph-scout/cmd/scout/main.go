package main

import (
	"context"
	"log/slog"
	"os"

	"sentigraph-scout/internal/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" { redisURL = "localhost:6379" }

	scout := app.NewScout(redisURL)
	scout.Run(context.Background())
}
