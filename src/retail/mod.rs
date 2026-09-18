pub mod bot;
pub mod crowd;
pub mod human;
pub mod reaction;
pub mod taxonomy;

pub use bot::{SimulatedRetailTrader, StopLoss, TakeProfit, WatchState};
pub use crowd::{attention_score, generate_watchlist_pool};
pub use human::HumanTrader;
pub use reaction::{reaction_for, ActionProbabilities, Signal};
pub use taxonomy::{
    get_tier_profile, Archetype, InformationStyle, SkillTier, StopPlacementStyle, TierProfile,
};
