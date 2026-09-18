use serde::{Deserialize, Serialize};

use super::centralbank::CentralBank;
use super::fiscal::{FiscalSystem, POLICY_TARIFF_SHIELD};
use super::labor::LaborMarket;
use super::trade::TradeEconomy;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SovereignBond {
    pub id: String,
    pub name: String,
    pub maturity_years: u32,
    pub coupon_rate: f64,
    pub yield_to_maturity: f64,
    pub price: f64,
    pub outstanding_amount: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NationalReport {
    pub gdp: f64,
    pub real_gdp_growth: f64,
    pub fed_funds_rate: f64,
    pub ten_year_yield: f64,
    pub cpi_inflation_rate: f64,
    pub policy_stance: String,
    pub fomc_announcement: String,
    pub unemployment_rate: f64,
    pub labor_force: f64,
    pub employed_workers: f64,
    pub average_hourly_wage: f64,
    pub annual_wage_growth: f64,
    pub net_monthly_payrolls: f64,
    pub annual_exports: f64,
    pub annual_imports: f64,
    pub trade_balance: f64,
    pub dollar_index_dxy: f64,
    pub federal_revenue: f64,
    pub federal_outlays: f64,
    pub budget_deficit: f64,
    pub national_debt: f64,
    pub treasury_cash: f64,
    pub reserve_currency_reserves: f64,
    pub credit_rating: String,
    pub borrowing_yield: f64,
    pub infrastructure_level: f64,
    pub healthcare_level: f64,
    pub education_level: f64,
    pub enterprise_grants_level: f64,
    pub export_capacity_level: f64,
    pub fdi_inflow: f64,
    pub export_revenue: f64,
    pub exchange_chartered: bool,
    pub maintenance_spending_rate: f64,
    pub depreciation_rate: f64,
    pub tourism_revenue: f64,
    pub productivity_index: f64,
    pub bonds: Vec<SovereignBond>,
    pub active_policies: Vec<String>,
    pub tier: String,          // "small", "mid", "powerhouse"
    pub tier_name: String,     // "Developing Nation", "Emerging State", "Global Powerhouse"
    pub tier_progress: f64,    // 0.0 to 100.0%

    // Demographic, Labor & Geopolitical Properties
    pub population: f64,
    pub gdp_per_capita: f64,
    pub job_openings: f64,
    pub geopolitical_power: f64,
}

#[derive(Debug, Clone)]
pub struct NationalEconomy {
    pub labor: LaborMarket,
    pub trade: TradeEconomy,
    pub fiscal: FiscalSystem,
    pub central_bank: CentralBank,
    pub bonds: Vec<SovereignBond>,

    pub nominal_gdp: f64,
    pub real_gdp_growth: f64,
    pub prev_gdp: f64,
}

impl Default for NationalEconomy {
    fn default() -> Self {
        // Starting as a developing sovereign nation ($45B GDP)
        let base_gdp = 45_000_000_000.0;
        let base_yield = 0.042;

        let bonds = vec![
            SovereignBond {
                id: "BOND-1Y".to_string(),
                name: "1-Year Sovereign T-Bill".to_string(),
                maturity_years: 1,
                coupon_rate: 0.035,
                yield_to_maturity: (base_yield - 0.005f64).max(0.01f64),
                price: 99.90,
                outstanding_amount: 1_200_000_000.0,
            },
            SovereignBond {
                id: "BOND-5Y".to_string(),
                name: "5-Year Sovereign Treasury Note".to_string(),
                maturity_years: 5,
                coupon_rate: 0.040,
                yield_to_maturity: base_yield,
                price: 100.00,
                outstanding_amount: 3_500_000_000.0,
            },
            SovereignBond {
                id: "BOND-10Y".to_string(),
                name: "10-Year Benchmark Sovereign Bond".to_string(),
                maturity_years: 10,
                coupon_rate: 0.045,
                yield_to_maturity: base_yield + 0.006,
                price: 100.00,
                outstanding_amount: 4_500_000_000.0,
            },
            SovereignBond {
                id: "BOND-30Y".to_string(),
                name: "30-Year Sovereign Century Bond".to_string(),
                maturity_years: 30,
                coupon_rate: 0.052,
                yield_to_maturity: base_yield + 0.015,
                price: 96.90,
                outstanding_amount: 2_500_000_000.0,
            },
        ];

        Self {
            labor: LaborMarket::default(),
            trade: TradeEconomy::default(),
            fiscal: FiscalSystem::default(),
            central_bank: CentralBank::default(),
            bonds,
            nominal_gdp: base_gdp,
            real_gdp_growth: 0.045, // 4.5% starting growth rate for developing nation
            prev_gdp: base_gdp,
        }
    }
}

impl NationalEconomy {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn issue_bond(&mut self, maturity_years: u32, amount: f64) -> Result<(), &'static str> {
        if amount <= 0.0 {
            return Err("issuance amount must be positive");
        }
        let bond = self.bonds.iter_mut().find(|b| b.maturity_years == maturity_years)
            .ok_or("bond tranche with requested maturity not found")?;
        bond.outstanding_amount += amount;
        self.fiscal.treasury_cash += amount;
        self.fiscal.national_debt += amount;
        Ok(())
    }

    pub fn buyback_bond(&mut self, maturity_years: u32, amount: f64) -> Result<(), &'static str> {
        if amount <= 0.0 {
            return Err("buyback amount must be positive");
        }
        if self.fiscal.treasury_cash < amount {
            return Err("insufficient treasury cash for bond buyback");
        }
        let bond = self.bonds.iter_mut().find(|b| b.maturity_years == maturity_years)
            .ok_or("bond tranche with requested maturity not found")?;
        let actual_amount = amount.min(bond.outstanding_amount);
        bond.outstanding_amount -= actual_amount;
        self.fiscal.treasury_cash -= actual_amount;
        self.fiscal.national_debt = (self.fiscal.national_debt - actual_amount).max(0.0);
        Ok(())
    }

    pub fn current_tier(&self) -> (&'static str, &'static str, f64) {
        if self.nominal_gdp < 60_000_000_000.0 {
            let prog = ((self.nominal_gdp / 60_000_000_000.0) * 100.0).clamp(0.0, 99.9);
            ("small", "Developing Nation", prog)
        } else if self.nominal_gdp < 300_000_000_000.0 {
            let prog = (((self.nominal_gdp - 60_000_000_000.0) / 240_000_000_000.0) * 100.0).clamp(0.0, 99.9);
            ("mid", "Emerging State", prog)
        } else {
            ("powerhouse", "Global Powerhouse", 100.0)
        }
    }

    pub fn tick(
        &mut self,
        consumer_spending: f64,
        corporate_investment: f64,
        total_corporate_headcount: f64,
        corporate_taxes_paid: f64,
        citizen_taxes_paid: f64,
        tick_fraction_of_year: f64,
    ) -> NationalReport {
        let (rate_changed, fomc_announcement) = self.central_bank.tick(
            self.labor.unemployment_rate,
            self.labor.annual_wage_growth,
            tick_fraction_of_year,
        );

        self.labor.tick(
            total_corporate_headcount,
            self.central_bank.cpi_inflation_rate,
            tick_fraction_of_year,
        );

        if self.fiscal.is_policy_active(POLICY_TARIFF_SHIELD) {
            self.trade.average_tariff_rate = 0.065;
        } else {
            self.trade.average_tariff_rate = 0.028;
        }
        self.trade.tick(self.central_bank.fed_funds_rate, tick_fraction_of_year);

        self.fiscal.update_credit_rating(self.nominal_gdp, self.central_bank.cpi_inflation_rate);

        let mut fdi_base = self.nominal_gdp * 0.025 * (self.fiscal.infrastructure_level / 25.0);
        if self.central_bank.cpi_inflation_rate < 0.045 {
            fdi_base *= 1.20;
        }
        self.fiscal.fdi_inflow = fdi_base;

        self.fiscal.export_revenue = self.nominal_gdp * 0.12 * (self.fiscal.export_capacity_level / 20.0);

        let tariff_rev = self.trade.annual_imports * self.trade.average_tariff_rate;
        self.fiscal.tick(
            citizen_taxes_paid,
            corporate_taxes_paid,
            tariff_rev,
            self.fiscal.borrowing_yield,
            tick_fraction_of_year,
        );

        let c = if consumer_spending <= 0.0 {
            self.nominal_gdp * 0.62
        } else {
            consumer_spending
        };

        let i = if corporate_investment <= 0.0 {
            self.nominal_gdp * 0.20
        } else {
            corporate_investment
        };

        let g = self.fiscal.mandatory_spending * 0.60 + self.fiscal.discretionary_spending;
        let net_exports = self.trade.trade_balance;

        let computed_gdp = (c + i + g + net_exports).clamp(15_000_000_000.0, 500_000_000_000.0);

        self.nominal_gdp = 0.95 * self.nominal_gdp + 0.05 * computed_gdp;

        if self.prev_gdp > 0.0 && tick_fraction_of_year > 0.0 {
            let nominal_growth = (self.nominal_gdp - self.prev_gdp) / self.prev_gdp / tick_fraction_of_year;
            let real_growth = (nominal_growth - self.central_bank.cpi_inflation_rate).clamp(-0.06, 0.12);
            self.real_gdp_growth = 0.90 * self.real_gdp_growth + 0.10 * real_growth;
        }
        self.prev_gdp = self.nominal_gdp;

        let (tier, tier_name, tier_progress) = self.current_tier();
        let active_policies = self.fiscal.get_active_policies();

        let base_yield = self.fiscal.borrowing_yield;
        for bond in &mut self.bonds {
            match bond.maturity_years {
                1 => bond.yield_to_maturity = (base_yield - 0.005).max(0.01),
                5 => bond.yield_to_maturity = base_yield,
                10 => bond.yield_to_maturity = base_yield + 0.008,
                30 => bond.yield_to_maturity = base_yield + 0.018,
                _ => bond.yield_to_maturity = base_yield,
            }
            let duration = (bond.maturity_years as f64).min(10.0);
            bond.price = (100.0 * (1.0 + (bond.coupon_rate - bond.yield_to_maturity) * duration)).clamp(40.0, 160.0);
        }

        NationalReport {
            gdp: self.nominal_gdp,
            real_gdp_growth: self.real_gdp_growth,
            fed_funds_rate: self.central_bank.fed_funds_rate,
            ten_year_yield: self.central_bank.ten_year_yield,
            cpi_inflation_rate: self.central_bank.cpi_inflation_rate,
            policy_stance: self.central_bank.policy_stance.clone(),
            fomc_announcement: if rate_changed { fomc_announcement } else { String::new() },
            unemployment_rate: self.labor.unemployment_rate,
            labor_force: self.labor.labor_force,
            employed_workers: self.labor.employed_workers,
            average_hourly_wage: self.labor.average_hourly_wage,
            annual_wage_growth: self.labor.annual_wage_growth,
            net_monthly_payrolls: self.labor.net_monthly_payrolls,
            annual_exports: self.trade.annual_exports,
            annual_imports: self.trade.annual_imports,
            trade_balance: self.trade.trade_balance,
            dollar_index_dxy: self.trade.dollar_index_dxy,
            federal_revenue: self.fiscal.total_revenue,
            federal_outlays: self.fiscal.total_outlays,
            budget_deficit: self.fiscal.budget_deficit,
            national_debt: self.fiscal.national_debt,
            treasury_cash: self.fiscal.treasury_cash,
            reserve_currency_reserves: self.fiscal.reserve_currency_reserves,
            credit_rating: self.fiscal.credit_rating.clone(),
            borrowing_yield: self.fiscal.borrowing_yield,
            infrastructure_level: self.fiscal.infrastructure_level,
            healthcare_level: self.fiscal.healthcare_level,
            education_level: self.fiscal.education_level,
            enterprise_grants_level: self.fiscal.enterprise_grants_level,
            export_capacity_level: self.fiscal.export_capacity_level,
            fdi_inflow: self.fiscal.fdi_inflow,
            export_revenue: self.fiscal.export_revenue,
            exchange_chartered: self.fiscal.exchange_chartered,
            maintenance_spending_rate: self.fiscal.maintenance_spending_rate,
            depreciation_rate: self.fiscal.depreciation_rate,
            tourism_revenue: self.fiscal.tourism_revenue,
            productivity_index: self.fiscal.productivity_index,
            bonds: self.bonds.clone(),
            active_policies,
            tier: tier.to_string(),
            tier_name: tier_name.to_string(),
            tier_progress,
            population: 8_000_000.0,
            gdp_per_capita: self.nominal_gdp / 8_000_000.0,
            job_openings: (self.labor.labor_force * 0.042).max(10_000.0),
            geopolitical_power: match tier {
                "powerhouse" => 88.0,
                "mid" => 56.0,
                _ => 28.0,
            },
        }
    }

    pub fn borrow_sovereign_capital(&mut self, amount: f64) -> Result<String, &'static str> {
        self.fiscal.borrow_debt(amount)?;
        if let Some(b) = self.bonds.iter_mut().find(|b| b.maturity_years == 5) {
            b.outstanding_amount += amount;
        }
        Ok(format!("Successfully authorized ${:.1}M in sovereign borrowing.", amount / 1_000_000.0))
    }

    pub fn composite_sector_demand(&self, sector: &str, consumer_demand_factor: f64) -> f64 {
        let export_factor = self.trade.export_demand_factor(sector);
        let procurement_factor = self.fiscal.sector_procurement_demand(sector);

        let composite = 0.60 * consumer_demand_factor + 0.20 * export_factor + 0.20 * procurement_factor;
        composite.clamp(0.50, 2.50)
    }
}
