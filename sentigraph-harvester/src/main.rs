mod config;
mod harvester;
mod models;
mod storage;

use crate::config::Config;
use crate::harvester::Harvester;
use crate::storage::Storage;
use std::sync::Arc;
use tracing_subscriber;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let cfg = Config::from_env();
    let storage = Arc::new(Storage::new(&cfg.redis_url).await);
    
    let harvester = Harvester::new(&cfg.binance_ws_url, &cfg.symbol, storage.clone());
    harvester.run().await;
}
