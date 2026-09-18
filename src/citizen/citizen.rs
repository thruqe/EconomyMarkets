use std::collections::HashMap;
use serde::{Deserialize, Serialize};

use super::consumer::ConsumerEngine;
use super::demographics::Demographics;
use super::sentiment::Sentiment;
use super::taxation::Taxation;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CitizenReport {
    pub population: f64,
    pub labor_force: f64,
    pub employed_count: f64,
    pub unemployment_rate: f64,
    pub average_hourly_earnings: f64,
    pub disposable_income: f64,
    pub consumer_confidence: f64,
    pub happiness: f64,
    pub consumer_spending: f64,
    pub taxes_paid_this_tick: f64,
    pub sector_demand: HashMap<String, f64>,
}

#[derive(Debug, Clone)]
pub struct CitizenEconomy {
    pub demographics: Demographics,
    pub sentiment: Sentiment,
    pub taxation: Taxation,
    pub consumer: ConsumerEngine,
}

impl Default for CitizenEconomy {
    fn default() -> Self {
        Self {
            demographics: Demographics::default(),
            sentiment: Sentiment::default(),
            taxation: Taxation::default(),
            consumer: ConsumerEngine::default(),
        }
    }
}

impl CitizenEconomy {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn tick(
        &mut self,
        employment_count: f64,
        wage_rate: f64,
        inflation_rate: f64,
        wage_growth: f64,
        market_return: f64,
        tick_fraction_of_year: f64,
    ) -> CitizenReport {
        let effective_tax = self.taxation.effective_income_tax_rate;
        self.demographics.tick(employment_count, wage_rate, effective_tax, tick_fraction_of_year);

        let unemp = self.demographics.unemployment_rate();
        self.sentiment.tick(unemp, inflation_rate, wage_growth, market_return, effective_tax);

        self.consumer.tick(
            self.demographics.aggregate_disposable_income,
            self.sentiment.spending_propensity,
            self.demographics.personal_savings_rate,
        );

        let gross_income = self.demographics.employed_count
            * self.demographics.average_hourly_earnings
            * 2000.0
            * tick_fraction_of_year;
        let income_tax = self.taxation.compute_income_tax(gross_income);
        let sales_tax = self.taxation.compute_sales_tax(self.consumer.total_consumer_spending * tick_fraction_of_year);
        self.taxation.record_taxes(income_tax, 0.0, sales_tax);

        CitizenReport {
            population: self.demographics.total_population,
            labor_force: self.demographics.labor_force(),
            employed_count: self.demographics.employed_count,
            unemployment_rate: unemp,
            average_hourly_earnings: self.demographics.average_hourly_earnings,
            disposable_income: self.demographics.aggregate_disposable_income,
            consumer_confidence: self.sentiment.index,
            happiness: self.sentiment.happiness,
            consumer_spending: self.consumer.total_consumer_spending,
            taxes_paid_this_tick: income_tax + sales_tax,
            sector_demand: self.consumer.sector_demand_factors.clone(),
        }
    }

    pub fn set_public_factors(&mut self, healthcare_level: f64, education_level: f64) {
        self.demographics.set_public_factors(healthcare_level, education_level);
    }
}
