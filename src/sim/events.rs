use serde::{Deserialize, Serialize};
use crate::company::{JumpKind, ReportingProfile};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum EventKind {
    Fundamental,
    Restatement,
    Liquidation,
    IPO,
    Bailout,
    Acquisition,
    Restructuring,
    ReverseSplit,
    Chapter11,
    Macro,
}

impl EventKind {
    pub fn as_str(&self) -> &'static str {
        match self {
            EventKind::Fundamental => "Fundamental",
            EventKind::Restatement => "Restatement",
            EventKind::Liquidation => "Liquidation",
            EventKind::IPO => "IPO",
            EventKind::Bailout => "Bailout",
            EventKind::Acquisition => "Acquisition",
            EventKind::Restructuring => "Restructuring",
            EventKind::ReverseSplit => "ReverseSplit",
            EventKind::Chapter11 => "Chapter11",
            EventKind::Macro => "Macro",
        }
    }
}

impl std::fmt::Display for EventKind {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Event {
    pub tick: usize,
    pub kind: EventKind,
    pub symbol: String,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub fundamental_kind: Option<JumpKind>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub multiplier: Option<f64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub restatement_profile: Option<ReportingProfile>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub prior_gap_percent: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub severity: Option<f64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub liquidated_agent_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub liquidation_qty: Option<f64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub ipo_revenue: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ipo_price: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ipo_shares: Option<f64>,

    #[serde(skip_serializing_if = "Option::is_none")]
    pub distress_details: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub macro_headline: Option<String>,
}
