mod config;
mod harvester;
mod models;
mod storage;

use config::AppConfig;
use harvester::Harvester;
use storage::Storage;
use std::sync::Arc;
use tracing::info;

#[tokio::main]
async fn main() {
    let config = AppConfig::load();

    tracing_subscriber::fmt()
        .json()
        .with_env_filter(tracing_subscriber::EnvFilter::new(&config.rust_log))
        .init();

    info!("Starting SentiGraph-Harvester...");

    let storage = Arc::new(Storage::new(&config.redis_url).await);

    let harvester = Harvester::new(&config.binance_ws_url, storage);
    
    harvester.run().await;
}
