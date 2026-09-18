use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Sentiment {
    pub index: f64,
    pub happiness: f64,
    pub spending_propensity: f64,
    pub trailing_market_return: f64,
}

impl Default for Sentiment {
    fn default() -> Self {
        Self {
            index: 85.0,
            happiness: 82.0,
            spending_propensity: 1.0,
            trailing_market_return: 0.0,
        }
    }
}

impl Sentiment {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn tick(
        &mut self,
        unemployment_rate: f64,
        inflation_rate: f64,
        wage_growth: f64,
        market_return: f64,
        tax_burden: f64,
    ) {
        self.trailing_market_return = 0.9 * self.trailing_market_return + 0.1 * market_return;

        let real_wage_growth = wage_growth - inflation_rate;
        let wage_component = real_wage_growth * 300.0;

        let unemp_gap = unemployment_rate - 0.040;
        let unemp_component = -unemp_gap * 400.0;

        let wealth_component = self.trailing_market_return * 40.0;

        let inflation_gap = inflation_rate - 0.020;
        let inflation_penalty = if inflation_gap > 0.0 {
            -inflation_gap * 250.0
        } else {
            0.0
        };

        let tax_gap = tax_burden - 0.185;
        let tax_penalty = -tax_gap * 100.0;

        let target_index = (85.0 + wage_component + unemp_component + wealth_component + inflation_penalty + tax_penalty)
            .clamp(20.0, 120.0);
        self.index = 0.95 * self.index + 0.05 * target_index;

        let target_happiness = (self.index * 0.80 + (1.0 - unemployment_rate) * 15.0 - (inflation_rate - 0.02).max(0.0) * 100.0)
            .clamp(15.0, 98.0);
        self.happiness = 0.95 * self.happiness + 0.05 * target_happiness;

        self.spending_propensity = (0.70 + (self.index - 50.0) * (0.50 / 50.0)).clamp(0.60, 1.40);
    }
}
