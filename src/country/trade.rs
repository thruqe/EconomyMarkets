use std::collections::HashMap;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TradeEconomy {
    pub annual_exports: f64,
    pub annual_imports: f64,
    pub trade_balance: f64,
    pub dollar_index_dxy: f64,
    pub average_tariff_rate: f64,
    pub sector_export_exposure: HashMap<String, f64>,
    pub sector_import_sensitivity: HashMap<String, f64>,
}

impl Default for TradeEconomy {
    fn default() -> Self {
        let exp = 2_800_000_000.0;
        let imp = 3_100_000_000.0;

        let mut export_exp = HashMap::new();
        export_exp.insert("Information Technology".to_string(), 1.30);
        export_exp.insert("Industrials".to_string(), 1.25);
        export_exp.insert("Energy".to_string(), 1.20);
        export_exp.insert("Materials".to_string(), 1.15);
        export_exp.insert("Financials".to_string(), 1.10);
        export_exp.insert("Health Care".to_string(), 1.05);
        export_exp.insert("Consumer Discretionary".to_string(), 0.90);
        export_exp.insert("Consumer Staples".to_string(), 0.85);
        export_exp.insert("Communication Services".to_string(), 1.10);
        export_exp.insert("Utilities".to_string(), 0.50);
        export_exp.insert("Real Estate".to_string(), 0.40);

        let mut import_sens = HashMap::new();
        import_sens.insert("Consumer Discretionary".to_string(), 1.30);
        import_sens.insert("Information Technology".to_string(), 1.20);
        import_sens.insert("Industrials".to_string(), 1.10);
        import_sens.insert("Materials".to_string(), 1.00);

        Self {
            annual_exports: exp,
            annual_imports: imp,
            trade_balance: exp - imp,
            dollar_index_dxy: 100.0,
            average_tariff_rate: 0.028,
            sector_export_exposure: export_exp,
            sector_import_sensitivity: import_sens,
        }
    }
}

impl TradeEconomy {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn tick(&mut self, domestic_interest_rate: f64, _tick_fraction_of_year: f64) {
        let rate_gap = domestic_interest_rate - 0.045;
        let target_dxy = (100.0 + rate_gap * 150.0).clamp(85.0, 120.0);
        self.dollar_index_dxy = 0.98 * self.dollar_index_dxy + 0.02 * target_dxy;

        let dxy_factor = self.dollar_index_dxy / 100.0;
        let baseline_exports = 2_800_000_000.0;
        let baseline_imports = 3_100_000_000.0;

        self.annual_exports = baseline_exports * (1.0 / dxy_factor.powf(0.4));
        self.annual_imports = baseline_imports * dxy_factor.powf(0.3) * (1.0 - self.average_tariff_rate * 0.5);
        self.trade_balance = self.annual_exports - self.annual_imports;
    }

    pub fn export_demand_factor(&self, sector: &str) -> f64 {
        let exposure = self.sector_export_exposure.get(sector).copied().unwrap_or(1.0);
        let currency_competitiveness = 100.0 / self.dollar_index_dxy;
        1.0 + (exposure - 1.0) * 0.5 + (currency_competitiveness - 1.0) * 0.3
    }
}
