use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize)]
pub struct BinanceStreamWrapper<T> {
    pub stream: String,
    pub data: T,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct BinanceTicker {
    #[serde(rename = "e")]
    pub event_type: String,
    #[serde(rename = "E")]
    pub event_time: u64,
    #[serde(rename = "s")]
    pub symbol: String,
    #[serde(rename = "a")]
    pub agg_trade_id: u64,
    #[serde(rename = "p")]
    pub price: String,
    #[serde(rename = "q")]
    pub quantity: String,
    #[serde(rename = "f")]
    pub first_trade_id: u64,
    #[serde(rename = "l")]
    pub last_trade_id: u64,
    #[serde(rename = "T")]
    pub trade_time: u64,
    #[serde(rename = "m")]
    pub is_buyer_market_maker: bool,
    #[serde(rename = "M")]
    pub ignore: bool,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct BinanceDepth {
    #[serde(rename = "u")]
    pub last_update_id: u64,
    #[serde(rename = "b")]
    pub bids: Vec<[String; 2]>,
    #[serde(rename = "a")]
    pub asks: Vec<[String; 2]>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct BinanceLiquidation {
    #[serde(rename = "e")]
    pub event_type: String,
    #[serde(rename = "E")]
    pub event_time: u64,
    #[serde(rename = "o")]
    pub order: Order,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct Order {
    #[serde(rename = "s")]
    pub symbol: String,
    #[serde(rename = "S")]
    pub side: String,
    #[serde(rename = "o")]
    pub order_type: String,
    #[serde(rename = "f")]
    pub time_in_force: String,
    #[serde(rename = "q")]
    pub original_quantity: String,
    #[serde(rename = "p")]
    pub price: String,
    #[serde(rename = "ap")]
    pub average_price: String,
    #[serde(rename = "X")]
    pub order_status: String,
    #[serde(rename = "l")]
    pub last_filled_quantity: String,
    #[serde(rename = "z")]
    pub cumulative_filled_quantity: String,
    #[serde(rename = "T")]
    pub order_trade_time: u64,
}
