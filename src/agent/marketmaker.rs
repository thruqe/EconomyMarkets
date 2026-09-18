use std::sync::Arc;
use parking_lot::RwLock;

use crate::company::Company;
use crate::market::{Account, MarketState, Order, OrderSource, Side};
use super::agent::{net_inventory, variance, EMA};

pub struct MarketMaker {
    pub id: String,
    pub account: Arc<RwLock<Account>>,
    pub fair_value: EMA,
    pub company: Option<Company>,
    pub ladder_levels: usize,
    pub risk_aversion: f64,
    pub base_half_spread: f64,
    pub spread_vol_coefficient: f64,
    pub max_inventory: f64,
    pub quote_size: f64,
}

impl MarketMaker {
    pub fn new(
        id: impl Into<String>,
        starting_cash: f64,
        max_leverage: f64,
        maintenance_margin_ratio: f64,
    ) -> Self {
        let id_str = id.into();
        Self {
            account: Arc::new(RwLock::new(Account::new(
                id_str.clone(),
                starting_cash,
                max_leverage,
                maintenance_margin_ratio,
            ))),
            id: id_str,
            fair_value: EMA::new(0.1),
            company: None,
            ladder_levels: 1,
            risk_aversion: 0.1,
            base_half_spread: 0.02,
            spread_vol_coefficient: 50.0,
            max_inventory: 1000.0,
            quote_size: 100.0,
        }
    }

    pub fn set_company(&mut self, company: Company) {
        self.company = Some(company);
    }

    pub fn reservation_price(&self, inventory: f64, variance_val: f64) -> f64 {
        let effective_var = variance_val.max(0.0002);
        self.fair_value.value() - inventory * self.risk_aversion * effective_var
    }

    pub fn half_spread(&self, variance_val: f64) -> f64 {
        self.base_half_spread + self.spread_vol_coefficient * self.risk_aversion * variance_val
    }

    pub fn effective_spread(&self, recent_volatility: f64) -> f64 {
        2.0 * self.half_spread(variance(recent_volatility))
    }

    pub fn effective_reservation_price(&self, symbol: &str, recent_volatility: f64) -> f64 {
        let acct = self.account.read();
        self.reservation_price(net_inventory(&acct, symbol), variance(recent_volatility))
    }
}

impl OrderSource for MarketMaker {
    fn id(&self) -> &str {
        &self.id
    }

    fn next_orders(&mut self, state: &MarketState) -> Vec<Order> {
        if let Some(mid) = state.mid {
            self.fair_value.update(mid);
        }

        let mut center = 0.0;
        let fund = if state.fundamental_value > 0.0 {
            state.fundamental_value
        } else if let Some(ref co) = self.company {
            if co.reported_value > 0.0 { co.reported_value } else { co.true_value }
        } else {
            0.0
        };

        if fund > 0.0 {
            center = if self.fair_value.initialized() {
                0.70 * fund + 0.30 * self.fair_value.value()
            } else {
                fund
            };
        }
        if center <= 0.0 {
            if !self.fair_value.initialized() {
                return Vec::new();
            }
            center = self.fair_value.value();
        }

        let var_val = variance(state.recent_volatility);
        let inventory = {
            let acct = self.account.read();
            net_inventory(&acct, &state.symbol)
        };

        let effective_var = var_val.max(0.0002);
        let mut skew = inventory * self.risk_aversion * effective_var;
        let max_skew = center * 0.04;
        skew = skew.clamp(-max_skew, max_skew);

        let r = center - skew;
        let max_half_spread = 0.10f64.max(center * 0.02);
        let half_spread = self.half_spread(var_val).min(max_half_spread);

        let mut buy_size = self.quote_size;
        let mut sell_size = self.quote_size;

        if inventory > self.max_inventory * 0.75 {
            let excess = (inventory - self.max_inventory * 0.75) / (self.max_inventory * 0.25);
            buy_size = 1.0f64.max((self.quote_size * (1.0 - 0.9 * excess.min(1.0))).floor());
            sell_size = (self.quote_size * (1.0 + 0.5 * excess.min(2.0))).floor();
        } else if inventory < -self.max_inventory * 0.75 {
            let excess = (-inventory - self.max_inventory * 0.75) / (self.max_inventory * 0.25);
            sell_size = 1.0f64.max((self.quote_size * (1.0 - 0.9 * excess.min(1.0))).floor());
            buy_size = (self.quote_size * (1.0 + 0.5 * excess.min(2.0))).floor();
        }

        let can_buy = inventory < self.max_inventory;
        let can_sell = inventory > -self.max_inventory;

        if self.ladder_levels <= 1 {
            let mut bid_price = ((r - half_spread) * 100.0).round() / 100.0;
            let mut ask_price = ((r + half_spread) * 100.0).round() / 100.0;
            if bid_price < 0.01 {
                bid_price = 0.01;
            }
            if ask_price <= bid_price {
                ask_price = bid_price + 0.01;
            }

            let mut orders = Vec::new();
            if can_buy {
                orders.push(Order {
                    id: 0,
                    agent_id: self.id.clone(),
                    side: Side::Buy,
                    price: bid_price,
                    quantity: buy_size,
                    is_market: false,
                });
            }
            if can_sell {
                orders.push(Order {
                    id: 0,
                    agent_id: self.id.clone(),
                    side: Side::Sell,
                    price: ask_price,
                    quantity: sell_size,
                    is_market: false,
                });
            }
            return orders;
        }

        let ladder_steps = [
            (1.0, 1.0),
            (2.5, 1.5),
            (5.0, 2.5),
            (9.0, 4.0),
            (15.0, 6.0),
        ];
        let steps_to_use = &ladder_steps[..self.ladder_levels.min(ladder_steps.len())];

        let mut orders = Vec::new();
        for (offset_mult, size_mult) in steps_to_use {
            let step_offset = half_spread * offset_mult;
            let mut bp = ((r - step_offset) * 100.0).round() / 100.0;
            let mut ap = ((r + step_offset) * 100.0).round() / 100.0;
            if bp < 0.01 {
                bp = 0.01;
            }
            if ap <= bp {
                ap = bp + 0.01;
            }

            if can_buy {
                orders.push(Order {
                    id: 0,
                    agent_id: self.id.clone(),
                    side: Side::Buy,
                    price: bp,
                    quantity: (buy_size * size_mult).round(),
                    is_market: false,
                });
            }
            if can_sell {
                orders.push(Order {
                    id: 0,
                    agent_id: self.id.clone(),
                    side: Side::Sell,
                    price: ap,
                    quantity: (sell_size * size_mult).round(),
                    is_market: false,
                });
            }
        }

        orders
    }
}
