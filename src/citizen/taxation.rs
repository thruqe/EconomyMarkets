use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Taxation {
    pub effective_income_tax_rate: f64,
    pub capital_gains_tax_rate: f64,
    pub sales_tax_rate: f64,
    pub total_income_taxes_paid: f64,
    pub total_cap_gains_taxes_paid: f64,
    pub total_sales_taxes_paid: f64,
}

impl Default for Taxation {
    fn default() -> Self {
        Self {
            effective_income_tax_rate: 0.185,
            capital_gains_tax_rate: 0.150,
            sales_tax_rate: 0.065,
            total_income_taxes_paid: 0.0,
            total_cap_gains_taxes_paid: 0.0,
            total_sales_taxes_paid: 0.0,
        }
    }
}

impl Taxation {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn compute_income_tax(&self, gross_income: f64) -> f64 {
        gross_income * self.effective_income_tax_rate
    }

    pub fn compute_sales_tax(&self, consumer_spending: f64) -> f64 {
        consumer_spending * self.sales_tax_rate
    }

    pub fn compute_capital_gains_tax(&self, realized_gains: f64) -> f64 {
        if realized_gains <= 0.0 {
            0.0
        } else {
            realized_gains * self.capital_gains_tax_rate
        }
    }

    pub fn record_taxes(&mut self, income_tax: f64, cap_gains_tax: f64, sales_tax: f64) {
        self.total_income_taxes_paid += income_tax;
        self.total_cap_gains_taxes_paid += cap_gains_tax;
        self.total_sales_taxes_paid += sales_tax;
    }

    pub fn total_collected(&self) -> f64 {
        self.total_income_taxes_paid + self.total_cap_gains_taxes_paid + self.total_sales_taxes_paid
    }
}
