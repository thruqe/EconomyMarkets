use super::orderbook::{Order, OrderBook};

#[derive(Debug, Clone)]
pub struct MarketState {
    pub symbol: String,
    pub tick: usize,
    pub best_bid: Option<f64>,
    pub best_ask: Option<f64>,
    pub mid: Option<f64>,
    pub bid_depth: f64,
    pub ask_depth: f64,
    pub price_history: Vec<f64>,
    pub recent_volatility: f64,
    pub fundamental_value: f64,
}

impl MarketState {
    pub fn new(
        symbol: impl Into<String>,
        tick: usize,
        book: &OrderBook,
        depth_levels: usize,
        price_history: Vec<f64>,
        recent_volatility: f64,
        fundamental_value: f64,
    ) -> Self {
        let (bid_depth, ask_depth) = book.depth_at_levels(depth_levels);
        Self {
            symbol: symbol.into(),
            tick,
            best_bid: book.best_bid(),
            best_ask: book.best_ask(),
            mid: book.mid_price(),
            bid_depth,
            ask_depth,
            price_history,
            recent_volatility,
            fundamental_value,
        }
    }
}

pub trait OrderSource: Send + Sync {
    fn id(&self) -> &str;
    fn next_orders(&mut self, state: &MarketState) -> Vec<Order>;
}
