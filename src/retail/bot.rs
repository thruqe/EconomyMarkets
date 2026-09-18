use std::collections::HashMap;
use std::sync::Arc;
use parking_lot::RwLock;
use rand::rngs::StdRng;
use rand::Rng;

use crate::company::Company;
use crate::market::{Account, MarketState, Order, OrderSource, Side};
use crate::technicals::{atr, bollinger_bands, momentum, swing_high_low, Aggregator, Bar};
use super::reaction::{reaction_for, Signal};
use super::taxonomy::{
    get_tier_profile, Archetype, InformationStyle, SkillTier, StopPlacementStyle,
};

#[derive(Debug, Clone, Copy, Default)]
pub struct StopLoss {
    pub active: bool,
    pub trigger_at: f64,
    pub is_long: bool,
}

#[derive(Debug, Clone, Copy, Default)]
pub struct TakeProfit {
    pub active: bool,
    pub trigger_at: f64,
    pub is_long: bool,
}

#[derive(Debug, Clone)]
pub struct WatchState {
    pub agg: Aggregator,
    pub cooldown_remaining: usize,
    pub stop: StopLoss,
    pub target: TakeProfit,
    pub ticks_held: usize,
}

pub struct SimulatedRetailTrader {
    pub id: String,
    pub account: Arc<RwLock<Account>>,
    pub watchlist: HashMap<String, Company>,
    pub state: HashMap<String, WatchState>,
    pub tier: SkillTier,
    pub style: InformationStyle,
    pub archetype: Archetype,
    pub rng: StdRng,
}

impl SimulatedRetailTrader {
    pub fn new(
        id: impl Into<String>,
        account: Arc<RwLock<Account>>,
        watchlist: Vec<Company>,
        tier: SkillTier,
        style: InformationStyle,
        archetype: Archetype,
        rng: StdRng,
    ) -> Self {
        let tp = get_tier_profile(tier);
        let mut companies = HashMap::with_capacity(watchlist.len());
        let mut states = HashMap::with_capacity(watchlist.len());

        for c in watchlist {
            states.insert(
                c.symbol.clone(),
                WatchState {
                    agg: Aggregator::new(tp.aggregator_window, 200),
                    cooldown_remaining: 0,
                    stop: StopLoss::default(),
                    target: TakeProfit::default(),
                    ticks_held: 0,
                },
            );
            companies.insert(c.symbol.clone(), c);
        }

        Self {
            id: id.into(),
            account,
            watchlist: companies,
            state: states,
            tier,
            style,
            archetype,
            rng,
        }
    }

    pub fn watches(&self, symbol: &str) -> bool {
        self.watchlist.contains_key(symbol)
    }

    fn net_inventory(&self, symbol: &str) -> f64 {
        let acct = self.account.read();
        crate::agent::net_inventory(&acct, symbol)
    }

    fn check_stop_loss(&mut self, symbol: &str, current_price: f64) -> Option<Order> {
        let ws = self.state.get_mut(symbol)?;
        if !ws.stop.active {
            return None;
        }

        let mut triggered = false;
        if ws.stop.is_long && current_price <= ws.stop.trigger_at {
            triggered = true;
        }
        if !ws.stop.is_long && current_price >= ws.stop.trigger_at {
            triggered = true;
        }
        if !triggered {
            return None;
        }

        let is_long = ws.stop.is_long;
        let tp = get_tier_profile(self.tier);
        let cooldown = tp.cooldown_ticks_min
            + self.rng.gen_range(0..=(tp.cooldown_ticks_max - tp.cooldown_ticks_min));

        let qty = self.net_inventory(symbol).abs();
        let ws = self.state.get_mut(symbol)?;
        if qty <= 0.0 {
            ws.stop.active = false;
            return None;
        }

        let side = if is_long { Side::Sell } else { Side::Buy };
        ws.stop.active = false;
        ws.cooldown_remaining = cooldown;

        Some(Order {
            id: 0,
            agent_id: self.id.clone(),
            side,
            price: 0.0,
            quantity: qty,
            is_market: true,
        })
    }

    fn place_stop_for_ws(
        tier: SkillTier,
        ws: &mut WatchState,
        entry_price: f64,
        is_long: bool,
        bars: &[Bar],
    ) {
        let tp = get_tier_profile(tier);
        let distance = match tp.stop_placement {
            StopPlacementStyle::RoundNumberStop => {
                let round_to = round_number_increment(entry_price);
                let dist = if is_long {
                    entry_price - round_down(entry_price, round_to)
                } else {
                    round_up(entry_price, round_to) - entry_price
                };
                if dist <= 0.0 {
                    entry_price * 0.02
                } else {
                    dist
                }
            }
            StopPlacementStyle::SwingLevelStop => {
                let window = bars.len().min(10);
                if let Some((high, low)) = swing_high_low(bars, window) {
                    if is_long && low < entry_price {
                        entry_price - low
                    } else if !is_long && high > entry_price {
                        high - entry_price
                    } else {
                        entry_price * 0.03
                    }
                } else {
                    entry_price * 0.03
                }
            }
            StopPlacementStyle::BufferedStop => {
                let window = bars.len().saturating_sub(1).min(14);
                if let Some(a) = atr(bars, window) {
                    if a > 0.0 {
                        a * 2.5
                    } else {
                        entry_price * 0.05
                    }
                } else {
                    entry_price * 0.05
                }
            }
        };

        if is_long {
            ws.stop = StopLoss {
                active: true,
                trigger_at: entry_price - distance,
                is_long: true,
            };
        } else {
            ws.stop = StopLoss {
                active: true,
                trigger_at: entry_price + distance,
                is_long: false,
            };
        }
    }

    fn check_take_profit(&mut self, symbol: &str, current_price: f64) -> Option<Order> {
        let ws = self.state.get_mut(symbol)?;
        if !ws.target.active {
            return None;
        }

        let mut triggered = false;
        if ws.target.is_long && current_price >= ws.target.trigger_at {
            triggered = true;
        }
        if !ws.target.is_long && current_price <= ws.target.trigger_at {
            triggered = true;
        }
        if !triggered {
            return None;
        }

        let is_long = ws.target.is_long;
        let tp = get_tier_profile(self.tier);
        let cooldown = tp.cooldown_ticks_min
            + self.rng.gen_range(0..=(tp.cooldown_ticks_max - tp.cooldown_ticks_min));

        let qty = self.net_inventory(symbol).abs();
        let ws = self.state.get_mut(symbol)?;
        if qty <= 0.0 {
            ws.target.active = false;
            return None;
        }

        let side = if is_long { Side::Sell } else { Side::Buy };
        ws.target.active = false;
        ws.stop.active = false;
        ws.cooldown_remaining = cooldown;

        Some(Order {
            id: 0,
            agent_id: self.id.clone(),
            side,
            price: 0.0,
            quantity: qty,
            is_market: true,
        })
    }

    fn check_time_decay_rotation(&mut self, symbol: &str) -> Option<Order> {
        let inv = self.net_inventory(symbol);
        let ws = self.state.get_mut(symbol)?;
        if inv == 0.0 {
            ws.ticks_held = 0;
            return None;
        }

        ws.ticks_held += 1;
        if ws.ticks_held < 40 {
            return None;
        }

        if self.rng.gen::<f64>() >= 0.15 {
            return None;
        }

        ws.ticks_held = 0;
        ws.stop.active = false;
        ws.target.active = false;

        let qty = inv.abs();
        let side = if inv < 0.0 { Side::Buy } else { Side::Sell };

        let tp = get_tier_profile(self.tier);
        ws.cooldown_remaining = tp.cooldown_ticks_min
            + self.rng.gen_range(0..=(tp.cooldown_ticks_max - tp.cooldown_ticks_min));

        Some(Order {
            id: 0,
            agent_id: self.id.clone(),
            side,
            price: 0.0,
            quantity: qty,
            is_market: true,
        })
    }

    fn place_take_profit_for_ws(
        archetype: Archetype,
        rng: &mut StdRng,
        ws: &mut WatchState,
        entry_price: f64,
        is_long: bool,
    ) {
        let pct = match archetype {
            Archetype::Disciplined => 0.04 + rng.gen::<f64>() * 0.04,
            Archetype::MomentumChaser => 0.08 + rng.gen::<f64>() * 0.10,
            Archetype::Contrarian => 0.05 + rng.gen::<f64>() * 0.06,
            Archetype::PanicProne => 0.03 + rng.gen::<f64>() * 0.05,
            Archetype::Degenerate => 0.02 + rng.gen::<f64>() * 0.15,
        };

        if is_long {
            ws.target = TakeProfit {
                active: true,
                trigger_at: entry_price * (1.0 + pct),
                is_long: true,
            };
        } else {
            ws.target = TakeProfit {
                active: true,
                trigger_at: entry_price * (1.0 - pct),
                is_long: false,
            };
        }
    }

    fn compute_signal(
        style: InformationStyle,
        tier: SkillTier,
        rng: &mut StdRng,
        target: &Company,
        bars: &[Bar],
        state: &MarketState,
    ) -> Option<Signal> {
        let tp = get_tier_profile(tier);

        let mut s = Signal::default();
        let mut have_technical = false;

        if let Some(mom) = momentum(bars, tp.momentum_window) {
            s.momentum = mom;
            have_technical = true;
        }

        let bb_window = bars.len().min(20);
        if let Some(bb) = bollinger_bands(bars, bb_window, 2.0) {
            if bb.middle > 0.0 {
                s.volatility = (bb.upper - bb.lower) / bb.middle;
            }
        }

        let mut have_fundamental = false;
        if tp.reacts_to_fundamental_events {
            if let Some(mid) = state.mid {
                if mid > 0.0 {
                    s.fundamental_gap = (target.reported_value - mid) / mid;
                    have_fundamental = true;
                }
            }
        }

        match style {
            InformationStyle::TechnicalOnly => {
                if have_technical {
                    Some(s)
                } else {
                    None
                }
            }
            InformationStyle::FundamentalOnly => {
                if have_fundamental {
                    s.momentum = s.fundamental_gap;
                    Some(s)
                } else {
                    None
                }
            }
            InformationStyle::Blended => {
                if !have_technical && !have_fundamental {
                    return None;
                }
                if have_technical && have_fundamental {
                    s.momentum = (s.momentum + s.fundamental_gap) / 2.0;
                } else if have_fundamental {
                    s.momentum = s.fundamental_gap;
                }
                Some(s)
            }
            InformationStyle::Confused => {
                let roll = rng.gen::<f64>();
                if roll < 0.4 && have_technical {
                    Some(s)
                } else if roll < 0.7 && have_fundamental {
                    s.momentum = s.fundamental_gap;
                    Some(s)
                } else {
                    None
                }
            }
        }
    }
}

impl OrderSource for SimulatedRetailTrader {
    fn id(&self) -> &str {
        &self.id
    }

    fn next_orders(&mut self, state: &MarketState) -> Vec<Order> {
        let target = match self.watchlist.get(&state.symbol) {
            Some(t) => t.clone(),
            None => return Vec::new(),
        };

        let current_mid = match state.mid {
            Some(m) if m > 0.0 => m,
            _ => return Vec::new(),
        };

        if let Some(ws) = self.state.get_mut(&state.symbol) {
            ws.agg.add_tick(current_mid);
        }

        if let Some(order) = self.check_stop_loss(&state.symbol, current_mid) {
            return vec![order];
        }

        if let Some(order) = self.check_take_profit(&state.symbol, current_mid) {
            return vec![order];
        }

        if let Some(order) = self.check_time_decay_rotation(&state.symbol) {
            return vec![order];
        }

        let (bars, cooldown) = match self.state.get(&state.symbol) {
            Some(w) => (w.agg.all_bars(), w.cooldown_remaining),
            None => return Vec::new(),
        };

        if cooldown > 0 {
            if let Some(ws) = self.state.get_mut(&state.symbol) {
                ws.cooldown_remaining -= 1;
            }
            return Vec::new();
        }

        let signal = match Self::compute_signal(self.style, self.tier, &mut self.rng, &target, &bars, state) {
            Some(s) => s,
            None => return Vec::new(),
        };

        let tp = get_tier_profile(self.tier);
        let base_activity = 0.15;
        let probs = reaction_for(self.archetype, signal, base_activity);

        let roll = self.rng.gen::<f64>();
        let side = if roll < probs.buy {
            Side::Buy
        } else if roll < probs.buy + probs.sell {
            Side::Sell
        } else {
            return Vec::new(); // hold
        };

        let size_fraction = tp.position_size_fraction_min
            + self.rng.gen::<f64>() * (tp.position_size_fraction_max - tp.position_size_fraction_min);
        let inv = self.net_inventory(&state.symbol);

        let quantity: f64;

        if side == Side::Sell && inv > 0.0 {
            let q = (inv * size_fraction).floor().max(1.0);
            quantity = q.min(inv);
            if quantity <= 0.0 {
                return Vec::new();
            }
        } else if side == Side::Buy && inv < 0.0 {
            let short_qty = -inv;
            let q = (short_qty * size_fraction).floor().max(1.0);
            quantity = q.min(short_qty);
            if quantity <= 0.0 {
                return Vec::new();
            }
        } else {
            let buying_power = {
                let mut prices = HashMap::new();
                prices.insert(state.symbol.clone(), current_mid);
                let acct = self.account.read();
                acct.available_buying_power(&prices)
            };
            let notional = buying_power * size_fraction;
            if notional <= 0.0 {
                return Vec::new();
            }
            quantity = (notional / current_mid).floor();
            if quantity <= 0.0 {
                return Vec::new();
            }

            let tier = self.tier;
            let arch = self.archetype;
            if let Some(ws) = self.state.get_mut(&state.symbol) {
                Self::place_stop_for_ws(tier, ws, current_mid, side == Side::Buy, &bars);
                Self::place_take_profit_for_ws(arch, &mut self.rng, ws, current_mid, side == Side::Buy);
            }
        }

        if let Some(ws) = self.state.get_mut(&state.symbol) {
            ws.cooldown_remaining = tp.cooldown_ticks_min
                + self.rng.gen_range(0..=(tp.cooldown_ticks_max - tp.cooldown_ticks_min));
        }

        vec![Order {
            id: 0,
            agent_id: self.id.clone(),
            side,
            price: 0.0,
            quantity,
            is_market: true,
        }]
    }
}

fn round_number_increment(price: f64) -> f64 {
    if price < 10.0 {
        0.5
    } else if price < 100.0 {
        5.0
    } else if price < 1000.0 {
        50.0
    } else {
        500.0
    }
}

fn round_down(price: f64, increment: f64) -> f64 {
    ((price / increment).floor()) * increment
}

fn round_up(price: f64, increment: f64) -> f64 {
    let down = round_down(price, increment);
    if (down - price).abs() < 1e-9 {
        price
    } else {
        down + increment
    }
}
