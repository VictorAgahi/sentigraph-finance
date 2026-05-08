package models

import "time"

type Config struct {
	RedisURL    string
	OllamaURL   string
	DatabaseURL string
	Port        string
	DiscordURL  string
}

type MarketState struct {
	Symbol    string    `json:"symbol"`
	Price     string    `json:"price"`
	Sentiment float64   `json:"sentiment"`
	Imbalance float64   `json:"imbalance"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NewsItem struct {
	Title  string `json:"title"`
	Source string `json:"source"`
	Symbol string `json:"symbol"`
}

type LiquidationItem struct {
	Type   string  `json:"type"`
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"`
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
}
