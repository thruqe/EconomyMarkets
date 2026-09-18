use std::collections::HashMap;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConsumerEngine {
    pub total_consumer_spending: f64,
    pub sector_demand_factors: HashMap<String, f64>,
}

impl Default for ConsumerEngine {
    fn default() -> Self {
        let mut factors = HashMap::new();
        factors.insert("Information Technology".to_string(), 1.0);
        factors.insert("Health Care".to_string(), 1.0);
        factors.insert("Financials".to_string(), 1.0);
        factors.insert("Consumer Discretionary".to_string(), 1.0);
        factors.insert("Consumer Staples".to_string(), 1.0);
        factors.insert("Industrials".to_string(), 1.0);
        factors.insert("Energy".to_string(), 1.0);
        factors.insert("Materials".to_string(), 1.0);
        factors.insert("Communication Services".to_string(), 1.0);
        factors.insert("Utilities".to_string(), 1.0);
        factors.insert("Real Estate".to_string(), 1.0);

        Self {
            total_consumer_spending: 32_000_000_000.0, // $32B annual consumer spending for developing nation
            sector_demand_factors: factors,
        }
    }
}

impl ConsumerEngine {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn tick(&mut self, disposable_income: f64, spending_propensity: f64, savings_rate: f64) {
        let effective_savings = (savings_rate / spending_propensity).clamp(0.02, 0.15);
        self.total_consumer_spending = disposable_income * (1.0 - effective_savings) * spending_propensity;

        let delta = spending_propensity - 1.0;

        self.sector_demand_factors.insert("Consumer Staples".to_string(), 1.0 + delta * 0.15);
        self.sector_demand_factors.insert("Health Care".to_string(), 1.0 + delta * 0.10);
        self.sector_demand_factors.insert("Utilities".to_string(), 1.0 + delta * 0.08);
        self.sector_demand_factors.insert("Energy".to_string(), 1.0 + delta * 0.35);
        self.sector_demand_factors.insert("Communication Services".to_string(), 1.0 + delta * 0.40);
        self.sector_demand_factors.insert("Financials".to_string(), 1.0 + delta * 0.50);
        self.sector_demand_factors.insert("Materials".to_string(), 1.0 + delta * 0.60);
        self.sector_demand_factors.insert("Industrials".to_string(), 1.0 + delta * 0.75);
        self.sector_demand_factors.insert("Real Estate".to_string(), 1.0 + delta * 0.85);
        self.sector_demand_factors.insert("Information Technology".to_string(), 1.0 + delta * 1.10);
        self.sector_demand_factors.insert("Consumer Discretionary".to_string(), 1.0 + delta * 1.40);

        for val in self.sector_demand_factors.values_mut() {
            *val = val.clamp(0.50, 1.80);
        }
    }

    pub fn demand_for_sector(&self, sector: &str) -> f64 {
        self.sector_demand_factors.get(sector).copied().unwrap_or(1.0)
    }
}
