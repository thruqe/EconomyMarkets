use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CentralBank {
    pub fed_funds_rate: f64,
    pub ten_year_yield: f64,
    pub cpi_inflation_rate: f64,
    pub inflation_target: f64,
    pub natural_unemployment: f64,
    pub policy_stance: String,
    pub meeting_ticks_remain: usize,
    pub last_decision: String,
}

impl Default for CentralBank {
    fn default() -> Self {
        Self {
            fed_funds_rate: 0.0450,
            ten_year_yield: 0.0420,
            cpi_inflation_rate: 0.0240,
            inflation_target: 0.0200,
            natural_unemployment: 0.0400,
            policy_stance: "Neutral".to_string(),
            meeting_ticks_remain: 150,
            last_decision: "FOMC maintained target rate unchanged at 4.50%".to_string(),
        }
    }
}

impl CentralBank {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn taylor_rule_rate(&self, unemployment_rate: f64) -> f64 {
        let neutral_rate = 0.025;
        let inflation_gap = self.cpi_inflation_rate - self.inflation_target;
        let unemp_gap = unemployment_rate - self.natural_unemployment;

        let taylor = neutral_rate + self.cpi_inflation_rate + 0.5 * inflation_gap - 0.5 * unemp_gap;
        taylor.clamp(0.0025, 0.090)
    }

    pub fn tick(
        &mut self,
        unemployment_rate: f64,
        wage_growth: f64,
        _tick_fraction_of_year: f64,
    ) -> (bool, String) {
        let wage_pressure = (wage_growth - 0.035) * 0.40;
        let interest_drag = -(self.fed_funds_rate - 0.025) * 0.15;
        let target_inflation = (self.inflation_target + wage_pressure + interest_drag).clamp(0.005, 0.085);

        self.cpi_inflation_rate = 0.98 * self.cpi_inflation_rate + 0.02 * target_inflation;

        let target_10y = self.fed_funds_rate * 0.85 + 0.0080;
        self.ten_year_yield = 0.95 * self.ten_year_yield + 0.05 * target_10y;

        if self.meeting_ticks_remain > 0 {
            self.meeting_ticks_remain -= 1;
            return (false, String::new());
        }

        self.meeting_ticks_remain = 200;
        let taylor = self.taylor_rule_rate(unemployment_rate);
        let rate_diff = taylor - self.fed_funds_rate;

        if self.cpi_inflation_rate > 0.030 {
            self.policy_stance = "Hawkish".to_string();
        } else if unemployment_rate > 0.055 {
            self.policy_stance = "Dovish".to_string();
        } else {
            self.policy_stance = "Neutral".to_string();
        }

        if rate_diff >= 0.0025 {
            let hike = if rate_diff >= 0.0060 { 0.0050 } else { 0.0025 };
            self.fed_funds_rate += hike;
            let announcement = format!(
                "FOMC raised federal funds rate by {} to {:.2}%",
                bps_string(hike),
                self.fed_funds_rate * 100.0
            );
            self.last_decision = announcement.clone();
            (true, announcement)
        } else if rate_diff <= -0.0025 {
            let cut = if rate_diff <= -0.0060 { 0.0050 } else { 0.0025 };
            self.fed_funds_rate -= cut;
            let announcement = format!(
                "FOMC lowered federal funds rate by {} to {:.2}%",
                bps_string(cut),
                self.fed_funds_rate * 100.0
            );
            self.last_decision = announcement.clone();
            (true, announcement)
        } else {
            self.last_decision = format!(
                "FOMC held federal funds rate steady at {:.2}%",
                self.fed_funds_rate * 100.0
            );
            (false, String::new())
        }
    }
}

fn bps_string(rate: f64) -> String {
    let bps = (rate * 10000.0).round() as i64;
    format!("{}bps", bps)
}
