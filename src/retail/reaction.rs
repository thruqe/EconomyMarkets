use super::taxonomy::Archetype;

#[derive(Debug, Clone, Copy, Default)]
pub struct Signal {
    pub momentum: f64,
    pub volatility: f64,
    pub fundamental_gap: f64,
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct ActionProbabilities {
    pub buy: f64,
    pub sell: f64,
    pub hold: f64,
}

fn normalize(buy: f64, sell: f64, hold: f64) -> ActionProbabilities {
    let total = buy + sell + hold;
    if total <= 0.0 {
        ActionProbabilities {
            buy: 0.0,
            sell: 0.0,
            hold: 1.0,
        }
    } else {
        ActionProbabilities {
            buy: buy / total,
            sell: sell / total,
            hold: hold / total,
        }
    }
}

fn sigmoid(x: f64) -> f64 {
    1.0 / (1.0 + (-x).exp())
}

pub fn reaction_for(archetype: Archetype, s: Signal, base_activity: f64) -> ActionProbabilities {
    match archetype {
        Archetype::MomentumChaser => {
            let strength = sigmoid(s.momentum * 20.0);
            let buy = base_activity * strength;
            let sell = base_activity * (1.0 - strength);
            let hold = 1.0 - base_activity;
            normalize(buy, sell, hold)
        }
        Archetype::Contrarian => {
            let strength = sigmoid(s.momentum * 20.0);
            let buy = base_activity * (1.0 - strength);
            let sell = base_activity * strength;
            let hold = 1.0 - base_activity;
            normalize(buy, sell, hold)
        }
        Archetype::PanicProne => {
            if s.momentum >= 0.0 {
                let buy = base_activity * (0.10 + 0.30 * sigmoid(s.momentum * 10.0));
                let sell = base_activity * 0.05;
                let hold = 1.0 - buy - sell;
                normalize(buy, sell, hold)
            } else {
                let amplified = (s.momentum.abs() * 40.0).min(1.0);
                let sell = base_activity * (0.20 + 0.80 * amplified);
                let hold = 1.0 - sell;
                normalize(0.0, sell, hold)
            }
        }
        Archetype::Disciplined => {
            let dampened = base_activity * 0.35;
            let strength = sigmoid(s.momentum * 20.0);
            let buy = dampened * strength;
            let sell = dampened * (1.0 - strength);
            let hold = 1.0 - dampened;
            normalize(buy, sell, hold)
        }
        Archetype::Degenerate => {
            let vol_driven = base_activity * (0.35 + 0.65 * (s.volatility * 8.0).min(1.0));
            let direction_lean = sigmoid(s.momentum * 10.0);
            let buy = vol_driven * direction_lean;
            let sell = vol_driven * (1.0 - direction_lean);
            let hold = 1.0 - vol_driven;
            normalize(buy, sell, hold)
        }
    }
}
