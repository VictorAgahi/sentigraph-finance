package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context, connString string) *Database {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		slog.Error("DATABASE_URL parse error", "error", err)
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		slog.Error("Database connection error", "error", err)
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
			sentiment NUMERIC NOT NULL
		);`,
		`SELECT create_hypertable('market_data', 'timestamp', if_not_exists => TRUE);`,
	}

	for _, q := range queries {
		_, err := db.pool.Exec(ctx, q)
		if err != nil {
			slog.Warn("Init query failed (possibly already exists)", "query", q, "error", err)
		}
	}
	slog.Info("Database initialized successfully")
}

func (db *Database) BatchInsertWorker(ctx context.Context, dataChan <-chan MarketState) {
	const batchSize = 100
	const flushInterval = 5 * time.Second

	batch := make([]MarketState, 0, batchSize)
	ticker := time.NewTicker(flushInterval)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		err := db.insertBatch(ctx, batch)
		if err != nil {
			slog.Error("Batch insert failed", "error", err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case data := <-dataChan:
			batch = append(batch, data)
			if len(batch) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}

func (db *Database) insertBatch(ctx context.Context, batch []MarketState) error {
	rows := make([][]interface{}, len(batch))
	for i, d := range batch {
		rows[i] = []interface{}{d.UpdatedAt, "BTCUSDT", d.Price, d.Sentiment}
	}

	_, err := db.pool.CopyFrom(
		ctx,
		pgx.Identifier{"market_data"},
		[]string{"timestamp", "symbol", "price", "sentiment"},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (db *Database) GetHistory(ctx context.Context, limit int) ([]MarketState, error) {
	rows, err := db.pool.Query(ctx, 
		"SELECT timestamp, price, sentiment FROM market_data ORDER BY timestamp DESC LIMIT $1", 
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []MarketState
	for rows.Next() {
		var d MarketState
		var price float64
		if err := rows.Scan(&d.UpdatedAt, &price, &d.Sentiment); err != nil {
			return nil, err
		}
		d.Price = fmt.Sprintf("%.2f", price)
		history = append(history, d)
	}
	return history, nil
}
