use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LaborMarket {
    pub labor_force: f64,
    pub employed_workers: f64,
    pub unemployment_rate: f64,
    pub natural_unemployment_rate: f64,
    pub average_hourly_wage: f64,
    pub annual_wage_growth: f64,
    pub job_openings: f64,
    pub net_monthly_payrolls: f64,
}

impl Default for LaborMarket {
    fn default() -> Self {
        let lf = 8_784_000.0;
        let unemp = 0.053;
        let employed = lf * (1.0 - unemp);

        Self {
            labor_force: lf,
            employed_workers: employed,
            unemployment_rate: unemp,
            natural_unemployment_rate: 0.048,
            average_hourly_wage: 3.25,
            annual_wage_growth: 0.038,
            job_openings: 420_000.0,
            net_monthly_payrolls: 18_500.0,
        }
    }
}

impl LaborMarket {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn tick(
        &mut self,
        total_corporate_headcount: f64,
        inflation_rate: f64,
        tick_fraction_of_year: f64,
    ) {
        if total_corporate_headcount > 0.0 {
            let target_employed = total_corporate_headcount
                .clamp(self.labor_force * 0.85, self.labor_force * 0.98);
            let delta = (target_employed - self.employed_workers) * 0.05;
            self.employed_workers += delta;
            self.net_monthly_payrolls = delta * (1.0 / (tick_fraction_of_year * 12.0).max(0.001));
        }

        let unemployed = (self.labor_force - self.employed_workers).max(0.0);
        self.unemployment_rate = unemployed / self.labor_force;

        let unemp_gap = self.natural_unemployment_rate - self.unemployment_rate;
        let target_wage_growth = (0.035 + unemp_gap * 0.75 + inflation_rate * 0.40).clamp(0.01, 0.10);
        self.annual_wage_growth = 0.95 * self.annual_wage_growth + 0.05 * target_wage_growth;

        self.average_hourly_wage *= 1.0 + self.annual_wage_growth * tick_fraction_of_year;
        self.job_openings = 100_000.0 + (0.07 - self.unemployment_rate).max(0.0) * 2_500_000.0;
    }
}
