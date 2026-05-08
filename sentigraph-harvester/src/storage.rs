use redis::aio::ConnectionManager;
use redis::pipe;
use tracing::{error, debug};
use std::sync::Arc;

pub struct Storage {
    client: ConnectionManager,
}

impl Storage {
    pub async fn new(redis_url: &str) -> Self {
        let client = redis::Client::open(redis_url).expect("Invalid Redis URL");
        let manager = ConnectionManager::new(client).await.expect("Failed to create Redis connection manager");
        Self { client: manager }
    }

    pub async fn send_batch(self: Arc<Self>, batch: Vec<String>) {
        if batch.is_empty() {
            return;
        }

        let mut pipeline = pipe();
        for json_trade in &batch {
            pipeline.publish("ticker:btc", json_trade);
            
            pipeline.cmd("XADD")
                .arg("market:stream")
                .arg("MAXLEN")
                .arg("~")
                .arg(1000)
                .arg("*")
                .arg("data")
                .arg(json_trade);
        }

        let mut conn = self.client.clone();
        match pipeline.query_async::<_, ()>(&mut conn).await {
            Ok(_) => debug!("Successfully sent batch of {} trades", batch.len()),
            Err(e) => error!("Failed to send Redis pipeline: {}", e),
        }
    }
}
