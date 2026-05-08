use redis::aio::ConnectionManager;
use redis::pipe;
use std::sync::Arc;
use tracing::{debug, error};

pub struct Storage {
    client: ConnectionManager,
}

impl Storage {
    pub async fn new(redis_url: &str) -> Self {
        let client = redis::Client::open(redis_url).expect("invalid redis url");
        let manager = ConnectionManager::new(client)
            .await
            .expect("failed to create redis manager");
        Self { client: manager }
    }

    pub async fn send_batch(self: Arc<Self>, batch: Vec<(String, String)>) {
        if batch.is_empty() { return }

        let mut pipeline = pipe();
        for (channel, data) in &batch {
            pipeline.publish(channel, data);
            
            if let Some(symbol) = channel.strip_prefix("ticker:") {
                pipeline.cmd("XADD")
                    .arg(format!("market:stream:{}", symbol))
                    .arg("MAXLEN").arg("~").arg(1000)
                    .arg("*")
                    .arg("data").arg(data);
            }
        }

        let mut conn = self.client.clone();
        if let Err(e) = pipeline.query_async::<_, ()>(&mut conn).await {
            error!(err = %e, "failed to send redis pipeline");
        } else {
            debug!(count = batch.len(), "sent batch to redis");
        }
    }
}
