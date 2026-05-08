use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize)]
pub struct BinanceStreamWrapper<T> {
    pub stream: String,
    pub data: T,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(from = "RawTicker")]
pub struct BinanceTicker {
    pub event_type: String,
    pub event_time: u64,
    pub symbol: String,
    pub agg_trade_id: u64,
    pub price: String,
    pub quantity: String,
    pub first_trade_id: u64,
    pub last_trade_id: u64,
    pub trade_time: u64,
    pub is_buyer_market_maker: bool,
    pub ignore: bool,
}

#[derive(Deserialize)]
struct RawTicker {
    e: String, E: u64, s: String, a: u64, p: String, q: String,
    f: u64, l: u64, T: u64, m: bool, M: bool,
}

impl From<RawTicker> for BinanceTicker {
    fn from(r: RawTicker) -> Self {
        Self {
            event_type: r.e, event_time: r.E, symbol: r.s, agg_trade_id: r.a,
            price: r.p, quantity: r.q, first_trade_id: r.f, last_trade_id: r.l,
            trade_time: r.T, is_buyer_market_maker: r.m, ignore: r.M,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(from = "RawDepth")]
pub struct BinanceDepth {
    pub last_update_id: u64,
    pub bids: Vec<[String; 2]>,
    pub asks: Vec<[String; 2]>,
}

#[derive(Deserialize)]
struct RawDepth {
    u: u64, b: Vec<[String; 2]>, a: Vec<[String; 2]>,
}

impl From<RawDepth> for BinanceDepth {
    fn from(r: RawDepth) -> Self {
        Self { last_update_id: r.u, bids: r.b, asks: r.a }
    }
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(from = "RawLiquidation")]
pub struct BinanceLiquidation {
    pub event_type: String,
    pub event_time: u64,
    pub order: Order,
}

#[derive(Deserialize)]
struct RawLiquidation {
    e: String, E: u64, o: Order,
}

impl From<RawLiquidation> for BinanceLiquidation {
    fn from(r: RawLiquidation) -> Self {
        Self { event_type: r.e, event_time: r.E, order: r.o }
    }
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(from = "RawOrder")]
pub struct Order {
    pub symbol: String,
    pub side: String,
    pub order_type: String,
    pub time_in_force: String,
    pub original_quantity: String,
    pub price: String,
    pub average_price: String,
    pub order_status: String,
    pub last_filled_quantity: String,
    pub cumulative_filled_quantity: String,
    pub order_trade_time: u64,
}

#[derive(Deserialize)]
struct RawOrder {
    s: String, S: String, o: String, f: String, q: String,
    p: String, ap: String, X: String, l: String, z: String, T: u64,
}

impl From<RawOrder> for Order {
    fn from(r: RawOrder) -> Self {
        Self {
            symbol: r.s, side: r.S, order_type: r.o, time_in_force: r.f,
            original_quantity: r.q, price: r.p, average_price: r.ap,
            order_status: r.X, last_filled_quantity: r.l,
            cumulative_filled_quantity: r.z, order_trade_time: r.T,
        }
    }
}
