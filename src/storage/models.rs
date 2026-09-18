use serde::{Deserialize, Serialize};

/// TradeRecord represents an individual matched fill between taker and maker orders.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TradeRecord {
    pub id: i64,
    pub tick: i64,
    pub timestamp: i64, // Unix timestamp in milliseconds
    pub symbol: String,
    pub taker_id: String,
    pub maker_id: String,
    pub side: String, // "BUY" or "SELL"
    pub price: f64,
    pub quantity: f64,
}

/// PriceRecord represents top-of-book market data for a symbol at a specific tick.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PriceRecord {
    pub tick: i64,
    pub timestamp: i64,
    pub symbol: String,
    pub mid: f64,
    pub bid: f64,
    pub ask: f64,
    pub spread: f64,
}

/// CandleRecord represents an aggregated OHLCV bar across a timeframe window.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CandleRecord {
    pub symbol: String,
    pub timeframe: String, // e.g. "1s", "5s", "1m", "5m"
    pub start_time: i64,
    pub end_time: i64,
    pub open: f64,
    pub high: f64,
    pub low: f64,
    pub close: f64,
    pub volume: f64,
    pub tick_count: i32,
}

/// EventRecord captures fundamental shocks, restatements, or liquidation events.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EventRecord {
    pub id: i64,
    pub tick: i64,
    pub timestamp: i64,
    pub kind: String, // "Fundamental", "Restatement", "Liquidation"
    pub symbol: String,
    pub details: String,
}

/// AccountRecord records periodic financial telemetry of a participant account.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AccountRecord {
    pub tick: i64,
    pub timestamp: i64,
    pub agent_id: String,
    pub cash: f64,
    pub equity: f64,
    pub margin_ratio: f64,
}

/// Batch bundles market records to be committed together in an atomic SQLite transaction.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Batch {
    pub trades: Vec<TradeRecord>,
    pub prices: Vec<PriceRecord>,
    pub candles: Vec<CandleRecord>,
    pub events: Vec<EventRecord>,
    pub accounts: Vec<AccountRecord>,
}

impl Batch {
    pub fn is_empty(&self) -> bool {
        self.trades.is_empty()
            && self.prices.is_empty()
            && self.candles.is_empty()
            && self.events.is_empty()
            && self.accounts.is_empty()
    }
}
