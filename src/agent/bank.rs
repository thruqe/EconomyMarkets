use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;

use crate::company::Company;
use crate::market::{Account, MarketState, Order, OrderSource};
use super::hedgefund::conviction_sized_order;

pub struct Bank {
    pub id: String,
    pub account: Arc<RwLock<Account>>,
    pub coverage: Vec<Company>,
    pub min_trade_threshold: f64,
    pub full_conviction_threshold: f64,
    pub rebalance_threshold: f64,
    pub skepticism_discount: f64,
    pub equity_high_water_mark: f64,
    pub hwm_initialized: bool,
    pub drawdown_de_risk_threshold: f64,
    pub drawdown_full_cutoff: f64,
    pub external_prices: Option<HashMap<String, f64>>,
    pub execution_rate: f64,
    pub max_order_shares: f64,
}

impl Bank {
    pub fn new(id: impl Into<String>, account: Arc<RwLock<Account>>, coverage: Vec<Company>) -> Self {
        Self {
            id: id.into(),
            account,
            coverage,
            min_trade_threshold: 0.06,
            full_conviction_threshold: 0.30,
            rebalance_threshold: 0.05,
            skepticism_discount: 0.50,
            equity_high_water_mark: 0.0,
            hwm_initialized: false,
            drawdown_de_risk_threshold: 0.10,
            drawdown_full_cutoff: 0.30,
            external_prices: None,
            execution_rate: 1.0,
            max_order_shares: 0.0,
        }
    }

    pub fn set_external_prices(&mut self, prices: HashMap<String, f64>) {
        self.external_prices = Some(prices);
    }

    pub fn risk_scale(&mut self, current_equity: f64) -> f64 {
        if !self.hwm_initialized || current_equity > self.equity_high_water_mark {
            self.equity_high_water_mark = current_equity;
            self.hwm_initialized = true;
            return 1.0;
        }
        if self.equity_high_water_mark <= 0.0 {
            return 1.0;
        }

        let drawdown = (self.equity_high_water_mark - current_equity) / self.equity_high_water_mark;
        if drawdown <= self.drawdown_de_risk_threshold {
            return 1.0;
        }
        if drawdown >= self.drawdown_full_cutoff {
            return 0.0;
        }
        let span = self.drawdown_full_cutoff - self.drawdown_de_risk_threshold;
        1.0 - (drawdown - self.drawdown_de_risk_threshold) / span
    }

    fn current_prices_across_coverage(&self, state: &MarketState) -> HashMap<String, f64> {
        let mut prices = HashMap::with_capacity(self.coverage.len());
        for c in &self.coverage {
            if let Some(ref ext) = self.external_prices {
                if let Some(&p) = ext.get(&c.symbol) {
                    prices.insert(c.symbol.clone(), p);
                    continue;
                }
            }
            if c.symbol == state.symbol && state.mid.is_some() {
                prices.insert(c.symbol.clone(), state.mid.unwrap());
            } else {
                prices.insert(c.symbol.clone(), c.reported_value);
            }
        }
        prices
    }
}

impl OrderSource for Bank {
    fn id(&self) -> &str {
        &self.id
    }

    fn next_orders(&mut self, state: &MarketState) -> Vec<Order> {
        let target = match self.coverage.iter().find(|c| c.symbol == state.symbol) {
            Some(t) => t.clone(),
            None => return Vec::new(),
        };

        let prices = self.current_prices_across_coverage(state);
        let equity = {
            let acct = self.account.read();
            acct.equity(&prices)
        };
        let scale = self.risk_scale(equity);

        let acct = self.account.read();
        let mut order = match conviction_sized_order(
            &self.id,
            &acct,
            &target,
            state,
            self.min_trade_threshold,
            self.full_conviction_threshold,
            self.rebalance_threshold,
            scale,
            self.skepticism_discount,
        ) {
            Some(o) => o,
            None => return Vec::new(),
        };

        if self.execution_rate > 0.0 && self.execution_rate < 1.0 {
            order.quantity *= self.execution_rate;
        }
        if self.max_order_shares > 0.0 && order.quantity > self.max_order_shares {
            order.quantity = self.max_order_shares;
        }
        if order.quantity <= 0.0 {
            return Vec::new();
        }

        vec![order]
    }
}
