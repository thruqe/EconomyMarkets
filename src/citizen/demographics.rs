use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Demographics {
    pub total_population: f64,
    pub working_age_population: f64,
    pub labor_force_participation: f64,
    pub employed_count: f64,
    pub median_household_income: f64,
    pub average_hourly_earnings: f64,
    pub aggregate_disposable_income: f64,
    pub personal_savings_rate: f64,
    pub population_growth_rate: f64,
}

impl Default for Demographics {
    fn default() -> Self {
        let pop = 18_500_000.0;
        let working_age = 12_200_000.0;
        let part_rate = 0.72;
        let labor_force = working_age * part_rate;
        let employed = labor_force * 0.947; // ~5.3% unemployment

        let hourly_wage = 3.25; // ~$6,500/year developing nation wage
        let annual_wage = hourly_wage * 2000.0;
        let disposable = employed * annual_wage * 0.82;

        Self {
            total_population: pop,
            working_age_population: working_age,
            labor_force_participation: part_rate,
            employed_count: employed,
            median_household_income: 6_200.0,
            average_hourly_earnings: hourly_wage,
            aggregate_disposable_income: disposable,
            personal_savings_rate: 0.08,
            population_growth_rate: 0.012,
        }
    }
}

impl Demographics {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn labor_force(&self) -> f64 {
        self.working_age_population * self.labor_force_participation
    }

    pub fn unemployment_rate(&self) -> f64 {
        let lf = self.labor_force();
        if lf <= 0.0 {
            return 0.0;
        }
        let unemployed = (lf - self.employed_count).max(0.0);
        unemployed / lf
    }

    pub fn tick(
        &mut self,
        employment_count: f64,
        wage_rate: f64,
        effective_tax_rate: f64,
        tick_fraction_of_year: f64,
    ) {
        let pop_growth = self.total_population * self.population_growth_rate * tick_fraction_of_year;
        self.total_population += pop_growth;
        self.working_age_population += pop_growth * 0.65;

        if employment_count > 0.0 {
            self.employed_count = employment_count.min(self.labor_force());
        }

        if wage_rate > 0.0 {
            self.average_hourly_earnings = wage_rate;
        }

        let annual_hours = 2000.0;
        let annual_wage_per_worker = self.average_hourly_earnings * annual_hours;
        let gross_income = self.employed_count * annual_wage_per_worker;
        let tax_rate = effective_tax_rate.clamp(0.05, 0.35);
        self.aggregate_disposable_income = gross_income * (1.0 - tax_rate);
    }

    pub fn set_public_factors(&mut self, healthcare_level: f64, education_level: f64) {
        self.population_growth_rate = 0.008 + (healthcare_level / 100.0) * 0.020;
        let target_wage = 12.0 + (education_level / 100.0) * 30.0;
        if self.average_hourly_earnings < target_wage {
            self.average_hourly_earnings = target_wage;
        }
    }
}
