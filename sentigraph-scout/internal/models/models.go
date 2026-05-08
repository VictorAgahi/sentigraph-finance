package models

import "time"

type NewsItem struct {
	Symbol    string    `json:"symbol"`
	Title     string    `json:"title"`
	Source    string    `json:"source"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
}
