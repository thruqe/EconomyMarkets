use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;

use crate::company::Company;
use crate::market::{Account, MarketState, Order, OrderSource, Side};
use super::agent::signed_position_value;

pub fn conviction(abs_mispricing: f64, min_trade_threshold: f64, full_conviction_threshold: f64) -> f64 {
    if abs_mispricing < min_trade_threshold {
        return 0.0;
    }
    if abs_mispricing >= full_conviction_threshold {
        return 1.0;
    }
    let span = full_conviction_threshold - min_trade_threshold;
    (abs_mispricing - min_trade_threshold) / span
}

pub fn conviction_sized_order(
    agent_id: &str,
    acct: &Account,
    target: &Company,
    state: &MarketState,
    min_trade_threshold: f64,
    full_conviction_threshold: f64,
    rebalance_threshold: f64,
    risk_scale: f64,
    skepticism_discount: f64,
) -> Option<Order> {
    let market_price = match state.mid {
        Some(p) if p > 0.0 => p,
        _ => return None,
    };

    let perceived_value =
        target.reported_value - skepticism_discount * (target.reported_value - market_price);
    let mispricing = (perceived_value - market_price) / market_price;
    let abs_mispricing = mispricing.abs();

    let conv = conviction(abs_mispricing, min_trade_threshold, full_conviction_threshold);

    let mut current_prices = HashMap::new();
    current_prices.insert(state.symbol.clone(), market_price);

    let available_buying_power = acct.available_buying_power(&current_prices);
    let current_position_value = signed_position_value(acct, &state.symbol, market_price);

    let max_total_exposure = available_buying_power + current_position_value.abs();
    let mut target_value = conv * risk_scale * max_total_exposure;
    if mispricing < 0.0 {
        target_value = -target_value;
    }

    let delta = target_value - current_position_value;
    let rebalance_floor = rebalance_threshold * (available_buying_power + current_position_value.abs() + 1.0);
    if delta.abs() < rebalance_floor {
        return None;
    }

    let side = if delta < 0.0 { Side::Sell } else { Side::Buy };
    let quantity = delta.abs() / market_price;
    if quantity <= 0.0 {
        return None;
    }

    Some(Order {
        id: 0,
        agent_id: agent_id.to_string(),
        side,
        price: 0.0,
        quantity,
        is_market: true,
    })
}

pub struct HedgeFund {
    pub id: String,
    pub account: Arc<RwLock<Account>>,
    pub company: Company,
    pub min_trade_threshold: f64,
    pub full_conviction_threshold: f64,
    pub rebalance_threshold: f64,
    pub skepticism_discount: f64,
    pub execution_rate: f64,
    pub max_order_shares: f64,
}

impl HedgeFund {
    pub fn new(id: impl Into<String>, account: Arc<RwLock<Account>>, target: Company) -> Self {
        Self {
            id: id.into(),
            account,
            company: target,
            min_trade_threshold: 0.03,
            full_conviction_threshold: 0.25,
            rebalance_threshold: 0.05,
            skepticism_discount: 0.35,
            execution_rate: 1.0,
            max_order_shares: 0.0,
        }
    }
}

impl OrderSource for HedgeFund {
    fn id(&self) -> &str {
        &self.id
    }

    fn next_orders(&mut self, state: &MarketState) -> Vec<Order> {
        if state.fundamental_value > 0.0 {
            self.company.reported_value = state.fundamental_value;
            self.company.true_value = state.fundamental_value;
        }
        let acct = self.account.read();
        let mut order = match conviction_sized_order(
            &self.id,
            &acct,
            &self.company,
            state,
            self.min_trade_threshold,
            self.full_conviction_threshold,
            self.rebalance_threshold,
            1.0,
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
