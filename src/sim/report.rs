use std::collections::HashMap;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

use crate::citizen::CitizenReport;
use crate::country::NationalReport;
use super::events::Event;

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct PriceQuote {
    pub symbol: String,
    pub mid: f64,
    pub bid: f64,
    pub ask: f64,
    pub spread: f64,
    pub has_mid: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Trade {
    pub tick: usize,
    pub symbol: String,
    pub taker_id: String,
    pub maker_id: String,
    pub side: String,
    pub price: f64,
    pub quantity: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AccountSnapshot {
    pub agent_id: String,
    pub cash: f64,
    pub equity: f64,
    pub margin_ratio: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TickReport {
    pub tick: usize,
    pub timestamp: DateTime<Utc>,
    pub prices: HashMap<String, PriceQuote>,
    pub trades: Vec<Trade>,
    pub events: Vec<Event>,
    pub accounts: Vec<AccountSnapshot>,
    pub citizen: CitizenReport,
    pub national: NationalReport,
}
