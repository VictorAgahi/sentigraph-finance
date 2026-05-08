use std::env;

pub struct Config {
    pub redis_url: String,
    pub binance_ws_url: String,
    pub symbol: String,
}

impl Config {
    pub fn from_env() -> Self {
        dotenvy::dotenv().ok();

        Self {
            redis_url: env::var("REDIS_URL").unwrap_or_else(|_| "redis://127.0.0.1:6379".to_string()),
            binance_ws_url: env::var("BINANCE_WS_URL").unwrap_or_else(|_| {
                "wss://stream.binance.com:9443/stream?streams=btcusdt@aggTrade/btcusdt@depth20@100ms".to_string()
            }),
            symbol: env::var("SYMBOL").unwrap_or_else(|_| "btc".to_string()),
        }
    }
}
