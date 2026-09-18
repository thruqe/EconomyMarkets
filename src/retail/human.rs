use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;

use crate::market::{Account, MarketState, Order, OrderSource};

pub struct HumanTrader {
    pub id: String,
    pub account: Arc<RwLock<Account>>,
    pub pending: Vec<Order>,
    pub pending_by_symbol: HashMap<String, Vec<Order>>,
}

impl HumanTrader {
    pub fn new(id: impl Into<String>, account: Arc<RwLock<Account>>) -> Self {
        Self {
            id: id.into(),
            account,
            pending: Vec::new(),
            pending_by_symbol: HashMap::new(),
        }
    }

    pub fn submit_order(&mut self, mut o: Order) {
        o.agent_id = self.id.clone();
        self.pending.push(o);
    }

    pub fn submit_order_for_symbol(&mut self, symbol: impl Into<String>, mut o: Order) {
        o.agent_id = self.id.clone();
        self.pending_by_symbol
            .entry(symbol.into())
            .or_default()
            .push(o);
    }
}

impl OrderSource for HumanTrader {
    fn id(&self) -> &str {
        &self.id
    }

    fn next_orders(&mut self, state: &MarketState) -> Vec<Order> {
        let mut result = Vec::new();

        if let Some(orders) = self.pending_by_symbol.remove(&state.symbol) {
            result.extend(orders);
        }

        if !self.pending.is_empty() {
            result.extend(std::mem::take(&mut self.pending));
        }

        result
    }
}
