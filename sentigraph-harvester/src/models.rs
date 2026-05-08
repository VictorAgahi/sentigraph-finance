use serde::{Deserialize, Serialize};

#[allow(dead_code)]
#[derive(Debug, Deserialize, Serialize)]
pub struct BinanceTicker<'input> {
    #[serde(rename = "e")]
    pub event_type: &'input str,
    
    #[serde(rename = "E")]
    pub event_time: u64,
    
    #[serde(rename = "s")]
    pub symbol: &'input str,
    
    #[serde(rename = "a")]
    pub agg_trade_id: u64,
    
    #[serde(rename = "p")]
    pub price: &'input str,
    
    #[serde(rename = "q")]
    pub quantity: &'input str,
    
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
