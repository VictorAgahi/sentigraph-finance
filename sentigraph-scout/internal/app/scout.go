package app

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"sentigraph-scout/internal/models"

	"github.com/mmcdole/gofeed"
	"github.com/redis/go-redis/v9"
)

type Scout struct {
	rdb             *redis.Client
	parser          *gofeed.Parser
	nitterInstances []string
	nitterIdx       int
	feeds           map[string][]string
	seen            map[string]bool
}

func NewScout(redisURL string) *Scout {
	return &Scout{
		rdb: redis.NewClient(&redis.Options{Addr: redisURL}),
		parser: gofeed.NewParser(),
		nitterInstances: []string{
			"https://nitter.net",
			"https://nitter.cz",
			"https://nitter.it",
			"https://nitter.privacydev.net",
		},
		feeds: map[string][]string{
			"btc":     {"https://cointelegraph.com/rss/tag/bitcoin", "https://www.coindesk.com/arc/outboundfeeds/rss/"},
			"eth":     {"https://cointelegraph.com/rss/tag/ethereum"},
			"sol":     {"https://cointelegraph.com/rss/tag/solana"},
			"general": {"https://www.yahoo.com/news/rss", "nitter/binance/rss", "nitter/WhaleAlert/rss", "nitter/elonmusk/rss"},
		},
		seen: make(map[string]bool),
	}
}

func (s *Scout) Run(ctx context.Context) {
	for {
		for symbol, urls := range s.feeds {
			for _, url := range urls {
				s.fetch(ctx, symbol, url)
				time.Sleep(2 * time.Second)
			}
		}

		if len(s.seen) > 5000 {
			s.seen = make(map[string]bool)
		}
		time.Sleep(5 * time.Minute)
	}
}

func (s *Scout) fetch(ctx context.Context, symbol, url string) {
	fetchURL := url
	if strings.HasPrefix(url, "nitter/") {
		fetchURL = strings.Replace(url, "nitter/", s.nitterInstances[s.nitterIdx]+"/", 1)
		s.nitterIdx = (s.nitterIdx + 1) % len(s.nitterInstances)
	}

	fctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	feed, err := s.parser.ParseURLWithContext(fetchURL, fctx)
	if err != nil {
		slog.Error("fetch error", "url", fetchURL, "err", err)
		return
	}

	for _, item := range feed.Items {
		id := item.GUID
		if id == "" { id = item.Link }
		if s.seen[id] { continue }

		news := models.NewsItem{
			Symbol:    symbol,
			Title:     item.Title,
			Source:    feed.Title,
			URL:       item.Link,
			Timestamp: time.Now(),
		}

		payload, _ := json.Marshal(news)
		if err := s.rdb.Publish(ctx, "market:news", payload).Err(); err != nil {
			slog.Error("redis publish error", "err", err)
		}
		s.seen[id] = true
	}
}
