use rand::Rng;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
pub struct JumpParams {
    pub lambda_down: f64,
    pub lambda_up: f64,
    pub down_jump_min: f64,
    pub down_jump_max: f64,
    pub up_jump_min: f64,
    pub up_jump_max: f64,
    pub mania_probability: f64,
    pub mania_multiplier_min: f64,
    pub mania_multiplier_max: f64,
}

impl Default for JumpParams {
    fn default() -> Self {
        Self {
            lambda_down: 1.0 / 2000.0,
            lambda_up: 1.0 / 2500.0,
            down_jump_min: 0.06,
            down_jump_max: 0.25,
            up_jump_min: 0.06,
            up_jump_max: 0.25,
            mania_probability: 0.03,
            mania_multiplier_min: 1.5,
            mania_multiplier_max: 3.5,
        }
    }
}

impl JumpParams {
    pub fn scaled(&self, sector_multiplier: f64, cap_multiplier: f64) -> Self {
        let mut scaled = *self;
        scaled.lambda_down = self.lambda_down * sector_multiplier * cap_multiplier;
        scaled.lambda_up = self.lambda_up * sector_multiplier * cap_multiplier;
        scaled
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum JumpKind {
    NoJump,
    CrisisJump,
    GoodNewsJump,
    ManiaJump,
}

impl JumpKind {
    pub fn as_str(&self) -> &'static str {
        match self {
            JumpKind::NoJump => "none",
            JumpKind::CrisisJump => "crisis",
            JumpKind::GoodNewsJump => "good_news",
            JumpKind::ManiaJump => "mania",
        }
    }
}

impl std::fmt::Display for JumpKind {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct JumpResult {
    pub kind: JumpKind,
    pub multiplier: f64,
}

pub fn roll_jump<R: Rng + ?Sized>(p: &JumpParams, rng: &mut R) -> JumpResult {
    if rng.gen::<f64>() < p.lambda_down {
        let frac = p.down_jump_min + rng.gen::<f64>() * (p.down_jump_max - p.down_jump_min);
        return JumpResult {
            kind: JumpKind::CrisisJump,
            multiplier: 1.0 - frac,
        };
    }

    if rng.gen::<f64>() < p.lambda_up {
        if rng.gen::<f64>() < p.mania_probability {
            let mult = p.mania_multiplier_min
                + rng.gen::<f64>() * (p.mania_multiplier_max - p.mania_multiplier_min);
            return JumpResult {
                kind: JumpKind::ManiaJump,
                multiplier: mult,
            };
        }
        let frac = p.up_jump_min + rng.gen::<f64>() * (p.up_jump_max - p.up_jump_min);
        return JumpResult {
            kind: JumpKind::GoodNewsJump,
            multiplier: 1.0 + frac,
        };
    }

    JumpResult {
        kind: JumpKind::NoJump,
        multiplier: 1.0,
    }
}
