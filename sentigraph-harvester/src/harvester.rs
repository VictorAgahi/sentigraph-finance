use crate::models::BinanceTicker;
use crate::storage::Storage;
use futures_util::StreamExt;
use tracing::{error, info};
use std::sync::Arc;
use std::time::Duration;
use tokio::time::{sleep, interval};
use tokio_tungstenite::{connect_async, tungstenite::protocol::Message};
use url::Url;

const BATCH_SIZE: usize = 50;
const BATCH_INTERVAL_MS: u64 = 10;

pub struct Harvester {
    url: String,
    storage: Arc<Storage>,
}

impl Harvester {
    pub fn new(url: &str, storage: Arc<Storage>) -> Self {
        Self {
            url: url.to_string(),
            storage,
        }
    }

    pub async fn run(&self) {
        let url = Url::parse(&self.url).expect("Invalid Binance WebSocket URL");
        let mut retry_delay = Duration::from_secs(1);
        let max_retry_delay = Duration::from_secs(60);

        loop {
            info!("Connecting to Binance WebSocket: {}", self.url);

            match connect_async(url.clone()).await {
                Ok((mut ws_stream, _)) => {
                    info!("Successfully connected to Binance");
                    retry_delay = Duration::from_secs(1);

                    let mut buffer = Vec::with_capacity(BATCH_SIZE);
                    let mut flush_interval = interval(Duration::from_millis(BATCH_INTERVAL_MS));

                    loop {
                        tokio::select! {
                            // High-priority: Receive from WebSocket
                            message = ws_stream.next() => {
                                match message {
                                    Some(Ok(Message::Text(text))) => {
                                        if let Some(json_string) = self.process_message(&text) {
                                            buffer.push(json_string);
                                            if buffer.len() >= BATCH_SIZE {
                                                self.flush_batch(&mut buffer);
                                            }
                                        }
                                    }
                                    Some(Ok(_)) => (),
                                    Some(Err(e)) => {
                                        error!("WebSocket error: {}. Reconnecting...", e);
                                        break;
                                    }
                                    None => break,
                                }
                            }
                            // Time-based flush
                            _ = flush_interval.tick() => {
                                if !buffer.is_empty() {
                                    self.flush_batch(&mut buffer);
                                }
                            }
                        }
                    }
                }
                Err(e) => {
                    error!("Failed to connect: {}. Retrying in {:?}...", e, retry_delay);
                }
            }

            sleep(retry_delay).await;
            retry_delay = std::cmp::min(retry_delay * 2, max_retry_delay);
        }
    }

    #[inline(always)]
    fn process_message(&self, text: &str) -> Option<String> {
        // Zero-copy parsing: we only check if it's valid and then use the raw string
        // or re-serialize if needed. Here we re-serialize for Redis format consistency.
        if let Ok(ticker) = serde_json::from_str::<BinanceTicker>(text) {
            // Re-serialization is needed because we can't trust the raw input for our DB schema
            // but we use the existing string buffer for efficiency where possible.
            return serde_json::to_string(&ticker).ok();
        }
        None
    }

    #[inline(always)]
    fn flush_batch(&self, buffer: &mut Vec<String>) {
        let batch = std::mem::take(buffer);
        let storage = Arc::clone(&self.storage);
        
        // Spawn the storage task to keep the harvester loop reactive
        tokio::spawn(async move {
            storage.send_batch(batch).await;
        });
    }
}
