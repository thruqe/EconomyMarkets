use std::sync::Arc;
use parking_lot::RwLock;

use crate::company::Company;
use crate::market::{MarketState, OrderBook, OrderSource};

pub struct CompanyMarket {
    pub co: Company,
    pub book: OrderBook,
    pub price_history: Vec<f64>,
    pub max_history: usize,
    pub participants: Vec<Arc<RwLock<dyn OrderSource>>>,
}

impl CompanyMarket {
    pub fn new(co: Company, max_history: usize) -> Self {
        Self {
            co,
            book: OrderBook::new(),
            max_history,
            price_history: Vec::with_capacity(max_history),
            participants: Vec::new(),
        }
    }

    pub fn record_price(&mut self) {
        if let Some(mid) = self.book.mid_price() {
            self.price_history.push(mid);
            if self.price_history.len() > self.max_history {
                let excess = self.price_history.len() - self.max_history;
                self.price_history.drain(0..excess);
            }
        }
    }

    pub fn recent_volatility(&self) -> f64 {
        if self.price_history.len() < 2 {
            return 0.0;
        }

        let mut returns = Vec::with_capacity(self.price_history.len() - 1);
        for i in 1..self.price_history.len() {
            let prev = self.price_history[i - 1];
            if prev <= 0.001 {
                continue;
            }
            let mut ret = (self.price_history[i] - prev) / prev;
            ret = ret.clamp(-0.10, 0.10);
            returns.push(ret);
        }

        if returns.len() < 2 {
            return 0.0;
        }

        let mean = returns.iter().sum::<f64>() / returns.len() as f64;
        let variance = returns.iter().map(|r| (r - mean).powi(2)).sum::<f64>() / returns.len() as f64;

        variance.sqrt().min(0.05)
    }

    pub fn state(&self, tick: usize, depth_levels: usize) -> MarketState {
        MarketState::new(
            &self.co.symbol,
            tick,
            &self.book,
            depth_levels,
            self.price_history.clone(),
            self.recent_volatility(),
            self.co.reported_value,
        )
    }
}
