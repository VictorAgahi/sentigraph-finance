use crate::models::{BinanceDepth, BinanceLiquidation};
use crate::storage::Storage;
use futures_util::StreamExt;
use std::sync::Arc;
use std::time::Duration;
use tokio::time::{interval, sleep};
use tokio_tungstenite::{connect_async, tungstenite::protocol::Message};
use tracing::{debug, error, info};
use url::Url;

const BATCH_SIZE: usize = 50;
const BATCH_INTERVAL_MS: u64 = 50;

pub struct Harvester {
    url: String,
    symbol: String,
    storage: Arc<Storage>,
}

impl Harvester {
    pub fn new(url: &str, symbol: &str, storage: Arc<Storage>) -> Self {
        Self {
            url: url.to_string(),
            symbol: symbol.to_string().to_lowercase(),
            storage,
        }
    }

    pub async fn run(&self) {
        let url = Url::parse(&self.url).expect("invalid ws url");
        let mut retry_delay = Duration::from_secs(1);
        let max_retry_delay = Duration::from_secs(60);

        loop {
            info!(url = %self.url, "connecting to websocket");

            match connect_async(url.clone()).await {
                Ok((mut stream, _)) => {
                    info!("connected to binance");
                    retry_delay = Duration::from_secs(1);

                    let mut buffer = Vec::with_capacity(BATCH_SIZE);
                    let mut flush_ticker = interval(Duration::from_millis(BATCH_INTERVAL_MS));

                    loop {
                        tokio::select! {
                            msg = stream.next() => {
                                match msg {
                                    Some(Ok(Message::Text(text))) => {
                                        if let Some(res) = self.process(&text) {
                                            buffer.push(res);
                                            if buffer.len() >= BATCH_SIZE {
                                                self.flush(&mut buffer);
                                            }
                                        }
                                    }
                                    Some(Ok(_)) => (),
                                    Some(Err(e)) => {
                                        error!(err = %e, "websocket error");
                                        break;
                                    }
                                    None => break,
                                }
                            }
                            _ = flush_ticker.tick() => {
                                if !buffer.is_empty() {
                                    self.flush(&mut buffer);
                                }
                            }
                        }
                    }
                }
                Err(e) => {
                    error!(err = %e, "connection failed");
                }
            }

            sleep(retry_delay).await;
            retry_delay = std::cmp::min(retry_delay * 2, max_retry_delay);
        }
    }

    fn process(&self, text: &str) -> Option<(String, String)> {
        let val: serde_json::Value = serde_json::from_str(text).ok()?;
        let stream = val["stream"].as_str()?;
        let data = val["data"].clone();
        
        if stream.contains("@ticker") || stream.contains("@aggTrade") {
            let price = if stream.contains("@ticker") { data["c"].as_str() } else { data["p"].as_str() };
            if let Some(p) = price {
                let simplified = serde_json::json!({ "s": self.symbol, "p": p });
                return Some((format!("ticker:{}", self.symbol), simplified.to_string()));
            }
        } else if stream.contains("@depth") {
            let _ = serde_json::from_value::<BinanceDepth>(data.clone()).ok()?;
            return Some((format!("depth:{}", self.symbol), data.to_string()));
        } else if stream.contains("@forceOrder") {
            let _ = serde_json::from_value::<BinanceLiquidation>(data.clone()).ok()?;
            return Some((format!("liquidation:{}", self.symbol), data.to_string()));
        }
        None
    }

    fn flush(&self, buffer: &mut Vec<(String, String)>) {
        let batch = std::mem::take(buffer);
        let storage = Arc::clone(&self.storage);
        tokio::spawn(async move {
            storage.send_batch(batch).await;
        });
    }
}
