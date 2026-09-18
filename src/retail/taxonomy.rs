use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum SkillTier {
    Beginner,
    Intermediate,
    Pro,
}

impl SkillTier {
    pub fn as_str(&self) -> &'static str {
        match self {
            SkillTier::Beginner => "Beginner",
            SkillTier::Intermediate => "Intermediate",
            SkillTier::Pro => "Pro",
        }
    }
}

impl std::fmt::Display for SkillTier {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum InformationStyle {
    TechnicalOnly,
    FundamentalOnly,
    Blended,
    Confused,
}

impl InformationStyle {
    pub fn as_str(&self) -> &'static str {
        match self {
            InformationStyle::TechnicalOnly => "TechnicalOnly",
            InformationStyle::FundamentalOnly => "FundamentalOnly",
            InformationStyle::Blended => "Blended",
            InformationStyle::Confused => "Confused",
        }
    }
}

impl std::fmt::Display for InformationStyle {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Archetype {
    MomentumChaser,
    Contrarian,
    PanicProne,
    Disciplined,
    Degenerate,
}

impl Archetype {
    pub fn as_str(&self) -> &'static str {
        match self {
            Archetype::MomentumChaser => "MomentumChaser",
            Archetype::Contrarian => "Contrarian",
            Archetype::PanicProne => "PanicProne",
            Archetype::Disciplined => "Disciplined",
            Archetype::Degenerate => "Degenerate",
        }
    }
}

impl std::fmt::Display for Archetype {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum StopPlacementStyle {
    RoundNumberStop,
    SwingLevelStop,
    BufferedStop,
}

#[derive(Debug, Clone, Copy)]
pub struct TierProfile {
    pub aggregator_window: usize,
    pub momentum_window: usize,
    pub uses_advanced_indicators: bool,
    pub reacts_to_fundamental_events: bool,
    pub position_size_fraction_min: f64,
    pub position_size_fraction_max: f64,
    pub stop_placement: StopPlacementStyle,
    pub cooldown_ticks_min: usize,
    pub cooldown_ticks_max: usize,
}

pub fn get_tier_profile(tier: SkillTier) -> TierProfile {
    match tier {
        SkillTier::Beginner => TierProfile {
            aggregator_window: 5,
            momentum_window: 3,
            uses_advanced_indicators: false,
            reacts_to_fundamental_events: false,
            position_size_fraction_min: 0.15,
            position_size_fraction_max: 0.60,
            stop_placement: StopPlacementStyle::RoundNumberStop,
            cooldown_ticks_min: 2,
            cooldown_ticks_max: 6,
        },
        SkillTier::Intermediate => TierProfile {
            aggregator_window: 15,
            momentum_window: 8,
            uses_advanced_indicators: false,
            reacts_to_fundamental_events: false,
            position_size_fraction_min: 0.05,
            position_size_fraction_max: 0.25,
            stop_placement: StopPlacementStyle::SwingLevelStop,
            cooldown_ticks_min: 4,
            cooldown_ticks_max: 12,
        },
        SkillTier::Pro => TierProfile {
            aggregator_window: 50,
            momentum_window: 20,
            uses_advanced_indicators: true,
            reacts_to_fundamental_events: true,
            position_size_fraction_min: 0.02,
            position_size_fraction_max: 0.08,
            stop_placement: StopPlacementStyle::BufferedStop,
            cooldown_ticks_min: 8,
            cooldown_ticks_max: 20,
        },
    }
}
