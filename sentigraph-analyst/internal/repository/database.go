package repository

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"sentigraph-analyst/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context, connString string) *Database {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		slog.Error("database config error", "err", err)
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		slog.Error("database connection error", "err", err)
		os.Exit(1)
	}

	db := &Database{pool: pool}
	db.Init(ctx)
	return db
}

func (db *Database) Init(ctx context.Context) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS market_data (
			timestamp TIMESTAMPTZ NOT NULL,
			symbol TEXT NOT NULL,
			price NUMERIC NOT NULL,
			sentiment NUMERIC NOT NULL,
			imbalance NUMERIC NOT NULL
		);`,
		`SELECT create_hypertable('market_data', 'timestamp', if_not_exists => TRUE);`,
	}

	for _, q := range queries {
		if _, err := db.pool.Exec(ctx, q); err != nil {
			slog.Warn("init query failed", "err", err)
		}
	}
}

func (db *Database) BatchInsertWorker(ctx context.Context, dataChan <-chan models.MarketState) {
	const batchSize = 100
	const flushInterval = 5 * time.Second

	batch := make([]models.MarketState, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 { return }
		if err := db.insertBatch(ctx, batch); err != nil {
			slog.Error("batch insert failed", "err", err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case data := <-dataChan:
			batch = append(batch, data)
			if len(batch) >= batchSize { flush() }
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}

func (db *Database) insertBatch(ctx context.Context, batch []models.MarketState) error {
	rows := make([][]interface{}, len(batch))
	for i, d := range batch {
		rows[i] = []interface{}{d.UpdatedAt, strings.ToUpper(d.Symbol), d.Price, d.Sentiment, d.Imbalance}
	}

	_, err := db.pool.CopyFrom(
		ctx,
		pgx.Identifier{"market_data"},
		[]string{"timestamp", "symbol", "price", "sentiment", "imbalance"},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (db *Database) GetHistory(ctx context.Context, symbol string, limit int) ([]models.MarketState, error) {
	rows, err := db.pool.Query(ctx, 
		"SELECT timestamp, symbol, price, sentiment, imbalance FROM (SELECT timestamp, symbol, price, sentiment, imbalance FROM market_data WHERE symbol = $1 ORDER BY timestamp DESC LIMIT $2) sub ORDER BY timestamp ASC", 
		strings.ToUpper(symbol), limit,
	)
	if err != nil { return nil, err }
	defer rows.Close()

	var history []models.MarketState
	for rows.Next() {
		var d models.MarketState
		var price float64
		if err := rows.Scan(&d.UpdatedAt, &d.Symbol, &price, &d.Sentiment, &d.Imbalance); err != nil {
			return nil, err
		}
		d.Price = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", price), "0"), ".")
		history = append(history, d)
	}
	return history, nil
}
