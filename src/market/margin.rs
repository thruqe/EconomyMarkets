use std::collections::HashMap;
use serde::{Deserialize, Serialize};
use super::orderbook::{Order, Side};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum PositionSide {
    Long,
    Short,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Position {
    pub side: PositionSide,
    pub quantity: f64,
    pub entry_cost: f64, // total cash paid/received at entry
    pub opened_at: i64,   // Unix timestamp nanos/millis
}

impl Position {
    pub fn new(side: PositionSide, quantity: f64, entry_cost: f64) -> Self {
        Self {
            side,
            quantity,
            entry_cost,
            opened_at: chrono::Utc::now().timestamp_millis(),
        }
    }

    pub fn position_value(&self, current_price: f64) -> f64 {
        self.quantity * current_price
    }

    pub fn unrealized_pnl(&self, current_price: f64) -> f64 {
        let market_value = self.quantity * current_price;
        match self.side {
            PositionSide::Long => market_value - self.entry_cost,
            PositionSide::Short => self.entry_cost - market_value,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Account {
    pub agent_id: String,
    pub cash: f64,
    pub positions: HashMap<String, Position>,
    pub max_leverage: f64,
    pub maintenance_margin_ratio: f64,
}

impl Account {
    pub fn new(
        agent_id: impl Into<String>,
        starting_cash: f64,
        max_leverage: f64,
        maintenance_margin_ratio: f64,
    ) -> Self {
        Self {
            agent_id: agent_id.into(),
            cash: starting_cash,
            positions: HashMap::new(),
            max_leverage,
            maintenance_margin_ratio,
        }
    }

    pub fn equity(&self, current_prices: &HashMap<String, f64>) -> f64 {
        let mut equity = self.cash;
        for (symbol, pos) in &self.positions {
            if let Some(&price) = current_prices.get(symbol) {
                equity += pos.unrealized_pnl(price);
            }
        }
        equity
    }

    pub fn total_position_value(&self, current_prices: &HashMap<String, f64>) -> f64 {
        let mut total = 0.0;
        for (symbol, pos) in &self.positions {
            if let Some(&price) = current_prices.get(symbol) {
                total += pos.position_value(price);
            }
        }
        total
    }

    pub fn margin_ratio(&self, current_prices: &HashMap<String, f64>) -> Option<f64> {
        let pos_val = self.total_position_value(current_prices);
        if pos_val <= 0.0 {
            None
        } else {
            Some(self.equity(current_prices) / pos_val)
        }
    }

    pub fn is_under_maintenance_margin(&self, current_prices: &HashMap<String, f64>) -> bool {
        match self.margin_ratio(current_prices) {
            Some(ratio) => ratio < self.maintenance_margin_ratio,
            None => false,
        }
    }

    pub fn available_buying_power(&self, current_prices: &HashMap<String, f64>) -> f64 {
        let eq = self.equity(current_prices);
        let max_exposure = eq * self.max_leverage;
        let used = self.total_position_value(current_prices);
        (max_exposure - used).max(0.0)
    }
}

pub fn apply_settled_fill(acct: &mut Account, symbol: &str, side: Side, quantity: f64, price: f64) {
    let position_side = match side {
        Side::Buy => PositionSide::Long,
        Side::Sell => PositionSide::Short,
    };

    let pos = match acct.positions.get_mut(symbol) {
        None => {
            acct.positions.insert(
                symbol.to_string(),
                Position {
                    side: position_side,
                    quantity,
                    entry_cost: quantity * price,
                    opened_at: chrono::Utc::now().timestamp_millis(),
                },
            );
            return;
        }
        Some(p) => p,
    };

    if pos.side == position_side {
        pos.quantity += quantity;
        pos.entry_cost += quantity * price;
        return;
    }

    // Opposing fill: reduces existing position, or fully closes and flips
    let avg_cost = pos.entry_cost / pos.quantity;
    let closed_qty = quantity.min(pos.quantity);

    let realized_pnl = match pos.side {
        PositionSide::Long => closed_qty * (price - avg_cost),
        PositionSide::Short => closed_qty * (avg_cost - price),
    };
    acct.cash += realized_pnl;

    if quantity < pos.quantity {
        pos.quantity -= quantity;
        pos.entry_cost = pos.quantity * avg_cost;
        return;
    }

    let remainder = quantity - pos.quantity;
    if remainder <= 1e-9 {
        acct.positions.remove(symbol);
    } else {
        let pos = acct.positions.get_mut(symbol).unwrap();
        pos.side = position_side;
        pos.quantity = remainder;
        pos.entry_cost = remainder * price;
        pos.opened_at = chrono::Utc::now().timestamp_millis();
    }
}

#[derive(Debug, Clone)]
pub struct ForcedOrder {
    pub symbol: String,
    pub order: Order,
    pub agent_id: String,
}

#[derive(Debug, Default, Clone, Copy)]
pub struct LiquidationEngine;

impl LiquidationEngine {
    pub fn scan_for_liquidations(
        accounts: &[&Account],
        current_prices: &HashMap<String, f64>,
    ) -> Vec<ForcedOrder> {
        let mut forced = Vec::new();
        for acct in accounts {
            if !acct.is_under_maintenance_margin(current_prices) {
                continue;
            }
            for (symbol, pos) in &acct.positions {
                if pos.quantity <= 1e-9 {
                    continue;
                }
                let side = match pos.side {
                    PositionSide::Long => Side::Sell,
                    PositionSide::Short => Side::Buy,
                };
                forced.push(ForcedOrder {
                    symbol: symbol.clone(),
                    order: Order {
                        id: 0,
                        agent_id: acct.agent_id.clone(),
                        side,
                        price: 0.0,
                        quantity: pos.quantity,
                        is_market: true,
                    },
                    agent_id: acct.agent_id.clone(),
                });
            }
        }
        forced
    }
}
