-- Initial migration for SentiGraph-Analyst
CREATE TABLE IF NOT EXISTS market_data (
    timestamp TIMESTAMPTZ NOT NULL,
    symbol TEXT NOT NULL,
    price NUMERIC NOT NULL,
    sentiment NUMERIC NOT NULL
);

-- Convert to Hypertable for TimescaleDB optimization
SELECT create_hypertable('market_data', 'timestamp', if_not_exists => TRUE);

-- Create index for faster history retrieval
CREATE INDEX IF NOT EXISTS idx_market_data_timestamp ON market_data (timestamp DESC);
