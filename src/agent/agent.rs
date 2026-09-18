use crate::market::{Account, PositionSide};

/// NetInventory reads an account's signed position size for a symbol:
/// positive for net long, negative for net short, zero if flat.
pub fn net_inventory(acct: &Account, symbol: &str) -> f64 {
    match acct.positions.get(symbol) {
        None => 0.0,
        Some(pos) => match pos.side {
            PositionSide::Short => -pos.quantity,
            PositionSide::Long => pos.quantity,
        },
    }
}

/// SignedPositionValue returns the current position's value signed by direction:
/// positive for long, negative for short, zero if flat.
pub fn signed_position_value(acct: &Account, symbol: &str, price: f64) -> f64 {
    match acct.positions.get(symbol) {
        None => 0.0,
        Some(pos) => {
            let val = pos.quantity * price;
            match pos.side {
                PositionSide::Short => -val,
                PositionSide::Long => val,
            }
        }
    }
}

/// EMA tracks a running exponential moving average.
#[derive(Debug, Clone)]
pub struct EMA {
    pub alpha: f64,
    value: f64,
    initialized: bool,
}

impl EMA {
    pub fn new(alpha: f64) -> Self {
        Self {
            alpha,
            value: 0.0,
            initialized: false,
        }
    }

    pub fn update(&mut self, observation: f64) -> f64 {
        if !self.initialized {
            self.value = observation;
            self.initialized = true;
        } else {
            self.value = self.alpha * observation + (1.0 - self.alpha) * self.value;
        }
        self.value
    }

    pub fn value(&self) -> f64 {
        self.value
    }

    pub fn initialized(&self) -> bool {
        self.initialized
    }
}

pub fn variance(recent_volatility: f64) -> f64 {
    recent_volatility * recent_volatility
}
