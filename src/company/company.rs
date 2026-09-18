use std::collections::HashSet;
use rand::rngs::StdRng;
use rand::{Rng, SeedableRng};
use rand_distr::{Distribution, StandardNormal};
use serde::{Deserialize, Serialize};

use super::jump::{roll_jump, JumpKind, JumpParams};
use super::naming::{generate_name, generate_symbol};
use super::reporting::{
    choose_reporting_profile, reporting_params_for, ReportingParams, ReportingProfile,
};
use super::taxonomy::{
    get_cap_tier_profile, get_sector_profile, CapTier, Sector, ALL_SECTORS,
};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FundamentalEvent {
    pub symbol: String,
    pub tick: usize,
    pub kind: JumpKind,
    pub multiplier: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RestatementEvent {
    pub symbol: String,
    pub tick: usize,
    pub profile: ReportingProfile,
    pub prior_gap_percent: f64,
    pub severity: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Company {
    pub symbol: String,
    pub name: String,
    pub sector: Sector,
    pub cap_tier: CapTier,
    pub true_value: f64,
    pub reported_value: f64,
    pub shares_outstanding: f64,
    pub float_shares: f64,

    pub annual_revenue: f64,
    pub net_margin: f64,
    pub sector_multiple: f64,
    pub ipo_price: f64,
    pub is_ipo: bool,
    pub ipo_tick: usize,

    pub headcount: f64,
    pub average_wage: f64,
    pub labor_expense: f64,
    pub debt_outstanding: f64,
    pub interest_expense: f64,
    pub corporate_tax_paid: f64,
    pub capex: f64,
    pub macro_demand_factor: f64,

    pub is_public: bool,
    pub stage: String,
    pub private_valuation: f64,

    pub drift: f64,
    pub volatility: f64,
    pub jump_params: JumpParams,
    pub reporting: ReportingParams,

    #[serde(skip)]
    pub rng: Option<StdRng>,
    pub reporting_bias_accum: f64,
    pub session_ticks_remaining: usize,
    pub session_drift: f64,
    #[serde(default)]
    pub momentum: f64,
    pub current_tick: usize,
}

impl Company {
    pub fn market_cap(&self) -> f64 {
        self.true_value * self.shares_outstanding
    }

    pub fn reported_market_cap(&self) -> f64 {
        self.reported_value * self.shares_outstanding
    }

    pub fn net_income(&self) -> f64 {
        self.annual_revenue * self.net_margin
    }

    pub fn earnings_per_share(&self) -> f64 {
        if self.shares_outstanding <= 0.0 {
            0.0
        } else {
            self.net_income() / self.shares_outstanding
        }
    }

    pub fn price_to_earnings(&self, market_price: f64) -> f64 {
        let eps = self.earnings_per_share();
        if eps <= 0.0 {
            0.0
        } else {
            market_price / eps
        }
    }

    pub fn price_to_sales(&self, market_price: f64) -> f64 {
        if self.annual_revenue <= 0.0 {
            0.0
        } else {
            (market_price * self.shares_outstanding) / self.annual_revenue
        }
    }

    pub fn operating_income(&self) -> f64 {
        let op_margin = self.net_margin * 1.35;
        self.annual_revenue * op_margin
    }

    pub fn update_macro(
        &mut self,
        hourly_wage: f64,
        corporate_tax_rate: f64,
        borrowing_rate: f64,
        demand_multiplier: f64,
        tick_fraction_of_year: f64,
    ) {
        if demand_multiplier > 0.0 {
            self.macro_demand_factor = demand_multiplier;
        } else {
            self.macro_demand_factor = 1.0;
        }

        if hourly_wage > 0.0 && self.headcount > 0.0 {
            let mut sector_premium = self.average_wage / 70000.0;
            if sector_premium < 0.6 {
                sector_premium = 0.6;
            } else if sector_premium > 2.5 {
                sector_premium = 2.5;
            }
            self.average_wage = hourly_wage * 2000.0 * sector_premium;
            self.labor_expense = self.headcount * self.average_wage;
        }

        let credit_spread = match self.cap_tier {
            CapTier::SmallCap => 0.035,
            CapTier::MidCap => 0.022,
            _ => 0.015,
        };
        self.interest_expense = self.debt_outstanding * (borrowing_rate + credit_spread);

        let ebit = self.operating_income();
        let taxable_income = ebit - self.interest_expense;
        if taxable_income > 0.0 {
            self.corporate_tax_paid = taxable_income * corporate_tax_rate;
        } else {
            self.corporate_tax_paid = 0.0;
        }

        let max_revenue = match self.cap_tier {
            CapTier::MegaCap => 3_500_000_000.0,
            CapTier::LargeCap => 1_200_000_000.0,
            CapTier::MidCap => 350_000_000.0,
            CapTier::SmallCap => 95_000_000.0,
        };
        let revenue_growth = ((self.macro_demand_factor - 1.0) * 0.05).clamp(-0.02, 0.02);
        self.annual_revenue = (self.annual_revenue * (1.0 + revenue_growth * tick_fraction_of_year)).clamp(1_000_000.0, max_revenue);

        if self.headcount > 0.0 {
            let rev_per_emp = self.annual_revenue / self.headcount;
            if rev_per_emp > 0.0 {
                let target_headcount = self.annual_revenue / rev_per_emp;
                self.headcount = 0.99 * self.headcount + 0.01 * target_headcount;
            }
        }
    }

    pub fn init_runtime(&mut self, rng: Option<StdRng>) {
        if let Some(r) = rng {
            self.rng = Some(r);
        } else if self.rng.is_none() {
            let mut seed: u64 = 0;
            for b in self.symbol.as_bytes() {
                seed = seed.wrapping_mul(31).wrapping_add(*b as u64);
            }
            if seed == 0 {
                seed = chrono::Utc::now().timestamp_nanos_opt().unwrap_or(42) as u64;
            }
            self.rng = Some(StdRng::seed_from_u64(seed));
        }

        let sp = get_sector_profile(self.sector);
        let cp = get_cap_tier_profile(self.cap_tier);

        if self.volatility == 0.0 {
            let vol_mult = if cp.vol_multiplier <= 0.0 { 1.0 } else { cp.vol_multiplier };
            let vol_span = sp.vol_max - sp.vol_min;
            let rand_val = self.rng.as_mut().map(|r| r.gen::<f64>()).unwrap_or(0.5);
            self.volatility = (sp.vol_min + rand_val * vol_span) * vol_mult;
        }

        if self.drift == 0.0 {
            let drift_span = sp.drift_max - sp.drift_min;
            let rand_val = self.rng.as_mut().map(|r| r.gen::<f64>()).unwrap_or(0.5);
            self.drift = sp.drift_min + rand_val * drift_span;
        }

        if self.jump_params.lambda_down == 0.0 && self.jump_params.lambda_up == 0.0 {
            self.jump_params = JumpParams::default().scaled(sp.jump_risk_multiplier, cp.jump_multiplier);
        }

        if self.reporting.noise_std_dev == 0.0 && self.reporting.persistent_bias == 0.0 {
            if let Some(r) = self.rng.as_mut() {
                let profile = choose_reporting_profile(self.sector, r);
                self.reporting = reporting_params_for(profile, r);
            }
        }

        if self.macro_demand_factor <= 0.0 {
            self.macro_demand_factor = 1.0;
        }
        if self.true_value <= 0.0 {
            self.true_value = 10.0;
        }
        if self.reported_value <= 0.0 {
            self.reported_value = self.true_value;
        }
    }

    pub fn tick(&mut self) -> (FundamentalEvent, Option<RestatementEvent>) {
        if self.rng.is_none() {
            self.init_runtime(None);
        }
        self.current_tick += 1;

        if self.session_ticks_remaining == 0 {
            if let Some(rng) = self.rng.as_mut() {
                // Multi-step market regimes: 35 to 110 ticks per wave/regime
                self.session_ticks_remaining = 35 + rng.gen_range(0..75);
                let roll = rng.gen::<f64>();
                if roll < 0.48 {
                    // Bullish trend regime / markup wave: persistent positive drift +0.06% to +0.18%
                    self.session_drift = 0.0006 + rng.gen::<f64>() * 0.0012;
                } else if roll < 0.82 {
                    // Bearish correction / profit-taking pullback: drift -0.05% to -0.15%
                    self.session_drift = -0.0005 - rng.gen::<f64>() * 0.0010;
                } else {
                    // Range-bound accumulation / base
                    self.session_drift = (rng.gen::<f64>() - 0.50) * 0.0004;
                }
            }
        } else {
            self.session_ticks_remaining -= 1;
        }

        let jump = {
            let rng = self.rng.as_mut().unwrap();
            roll_jump(&self.jump_params, rng)
        };

        let event = FundamentalEvent {
            symbol: self.symbol.clone(),
            tick: self.current_tick,
            kind: jump.kind,
            multiplier: jump.multiplier,
        };

        if jump.kind != JumpKind::NoJump {
            self.true_value = (self.true_value * jump.multiplier).clamp(2.0, 450.0);
            self.momentum = (jump.multiplier - 1.0) * 0.35; // Jump event injects directional momentum
        } else {
            let z: f64 = {
                let rng = self.rng.as_mut().unwrap();
                StandardNormal.sample(rng)
            };
            let macro_drift = (self.macro_demand_factor - 1.0) * 0.0012;
            let interest_drag = if self.annual_revenue > 0.0 {
                (self.interest_expense / self.annual_revenue) * 0.0003
            } else {
                0.0
            };

            // Fundamental valuation target based on operational earnings and sector multiple
            let target_price = {
                let ebit = self.operating_income().max(50_000.0);
                let per_share_val = (ebit * self.sector_multiple) / self.shares_outstanding.max(1_000.0);
                per_share_val.clamp(5.0, 350.0)
            };
            // Mean-reversion pull prevents runaway exponential drift
            let reversion_speed = 0.004;
            let reversion_pull = ((target_price - self.true_value) / self.true_value.max(1.0)).clamp(-0.03, 0.03) * reversion_speed;

            // Autoregressive momentum update: AR(1) persistence (rho = 0.82)
            // Replaces independent white-noise flip-flop with smooth, organic price waves
            let target_drift = self.session_drift + macro_drift - interest_drag;
            let noise_shock = self.volatility * 0.15 * z;
            self.momentum = 0.82 * self.momentum + 0.18 * target_drift + noise_shock;
            self.momentum = self.momentum.clamp(-0.015, 0.015);

            let step_change = (self.momentum + reversion_pull).clamp(-0.025, 0.025);
            self.true_value = (self.true_value * (1.0 + step_change)).clamp(2.0, 450.0);
        }

        let restatement = self.tick_reporting();
        (event, restatement)
    }

    fn tick_reporting(&mut self) -> Option<RestatementEvent> {
        self.reporting_bias_accum += self.reporting.persistent_bias;
        if self.reporting_bias_accum > 0.45 {
            self.reporting_bias_accum = 0.45;
        } else if self.reporting_bias_accum < -0.25 {
            self.reporting_bias_accum = -0.25;
        }

        let (noise, roll, sev_sample) = {
            let rng = self.rng.as_mut().unwrap();
            let z: f64 = StandardNormal.sample(rng);
            let n = z * self.reporting.noise_std_dev;
            let r = rng.gen::<f64>();
            let s = rng.gen::<f64>();
            (n, r, s)
        };

        self.reported_value = self.true_value * (1.0 + self.reporting_bias_accum + noise);
        if self.reported_value < 0.01 {
            self.reported_value = 0.01;
        }

        if roll >= self.reporting.restatement_probability || self.reporting_bias_accum.abs() < 1e-9 {
            return None;
        }

        let prior_gap_percent = self.reporting_bias_accum;
        let severity = self.reporting.restatement_severity_min
            + sev_sample
                * (self.reporting.restatement_severity_max - self.reporting.restatement_severity_min);

        self.reporting_bias_accum -= self.reporting_bias_accum * severity;
        self.reported_value = self.true_value * (1.0 + self.reporting_bias_accum + noise);
        if self.reported_value < 0.01 {
            self.reported_value = 0.01;
        }

        Some(RestatementEvent {
            symbol: self.symbol.clone(),
            tick: self.current_tick,
            profile: self.reporting.profile,
            prior_gap_percent,
            severity,
        })
    }
}

#[derive(Debug, Clone)]
pub struct GenerationParams {
    pub starting_value_min: f64,
    pub starting_value_max: f64,
    pub base_jump_params: JumpParams,
}

impl Default for GenerationParams {
    fn default() -> Self {
        Self {
            starting_value_min: 5.0,
            starting_value_max: 500.0,
            base_jump_params: JumpParams::default(),
        }
    }
}

pub fn generate_company(
    sector: Sector,
    tier: CapTier,
    params: &GenerationParams,
    rng: &mut StdRng,
    used_symbols: &mut HashSet<String>,
) -> Company {
    let sp = get_sector_profile(sector);
    let cp = get_cap_tier_profile(tier);

    let drift = sp.drift_min + rng.gen::<f64>() * (sp.drift_max - sp.drift_min);
    let vol = (sp.vol_min + rng.gen::<f64>() * (sp.vol_max - sp.vol_min)) * cp.vol_multiplier;

    let jump_params = params.base_jump_params.scaled(sp.jump_risk_multiplier, cp.jump_multiplier);
    let shares = cp.shares_outstanding_min
        + rng.gen::<f64>() * (cp.shares_outstanding_max - cp.shares_outstanding_min);
    let float_fraction =
        cp.float_fraction_min + rng.gen::<f64>() * (cp.float_fraction_max - cp.float_fraction_min);
    let starting_value =
        params.starting_value_min + rng.gen::<f64>() * (params.starting_value_max - params.starting_value_min);

    let name = generate_name(sector, rng);
    let symbol = generate_symbol(&name, used_symbols);

    let reporting_profile = choose_reporting_profile(sector, rng);
    let reporting_params = reporting_params_for(reporting_profile, rng);

    let margin = sp.net_margin_min + rng.gen::<f64>() * (sp.net_margin_max - sp.net_margin_min);
    let multiple = sp.ps_multiple_min + rng.gen::<f64>() * (sp.ps_multiple_max - sp.ps_multiple_min);
    let revenue = (starting_value * shares) / multiple;

    let mut co = Company {
        symbol,
        name,
        sector,
        cap_tier: tier,
        true_value: starting_value,
        reported_value: starting_value,
        shares_outstanding: shares,
        float_shares: shares * float_fraction,
        annual_revenue: revenue,
        net_margin: margin,
        sector_multiple: multiple,
        ipo_price: starting_value,
        is_ipo: false,
        ipo_tick: 0,
        headcount: 0.0,
        average_wage: 0.0,
        labor_expense: 0.0,
        debt_outstanding: 0.0,
        interest_expense: 0.0,
        corporate_tax_paid: 0.0,
        capex: 0.0,
        macro_demand_factor: 1.0,
        is_public: true,
        stage: "Public".to_string(),
        private_valuation: 0.0,
        drift,
        volatility: vol,
        jump_params,
        reporting: reporting_params,
        rng: None,
        reporting_bias_accum: 0.0,
        session_ticks_remaining: 0,
        session_drift: 0.0,
        momentum: 0.0,
        current_tick: 0,
    };
    init_macro_metrics(&mut co, rng);
    co.rng = Some(rng.clone());
    co
}

pub fn generate_ipo_company(
    sector: Sector,
    tier: CapTier,
    custom_revenue: f64,
    tick: usize,
    rng: &mut StdRng,
    used_symbols: &mut HashSet<String>,
) -> Company {
    let sp = get_sector_profile(sector);
    let cp = get_cap_tier_profile(tier);

    let drift = sp.drift_min + rng.gen::<f64>() * (sp.drift_max - sp.drift_min);
    let vol = (sp.vol_min + rng.gen::<f64>() * (sp.vol_max - sp.vol_min)) * cp.vol_multiplier;

    let base_jump = JumpParams::default();
    let jump_params = base_jump.scaled(sp.jump_risk_multiplier, cp.jump_multiplier);

    let shares = cp.shares_outstanding_min
        + rng.gen::<f64>() * (cp.shares_outstanding_max - cp.shares_outstanding_min);
    let float_fraction =
        cp.float_fraction_min + rng.gen::<f64>() * (cp.float_fraction_max - cp.float_fraction_min);

    let revenue = if custom_revenue <= 0.0 {
        cp.revenue_min + rng.gen::<f64>() * (cp.revenue_max - cp.revenue_min)
    } else {
        custom_revenue
    };

    let margin = sp.net_margin_min + rng.gen::<f64>() * (sp.net_margin_max - sp.net_margin_min);
    let multiple = sp.ps_multiple_min + rng.gen::<f64>() * (sp.ps_multiple_max - sp.ps_multiple_min);

    let implied_market_cap = revenue * multiple;
    let mut starting_value = implied_market_cap / shares;
    if starting_value < 5.0 {
        starting_value = 5.0;
    } else if starting_value > 1500.0 {
        starting_value = 1500.0;
    }

    let name = generate_name(sector, rng);
    let symbol = generate_symbol(&name, used_symbols);

    let reporting_profile = choose_reporting_profile(sector, rng);
    let reporting_params = reporting_params_for(reporting_profile, rng);

    let mut co = Company {
        symbol,
        name,
        sector,
        cap_tier: tier,
        true_value: starting_value,
        reported_value: starting_value,
        shares_outstanding: shares,
        float_shares: shares * float_fraction,
        annual_revenue: revenue,
        net_margin: margin,
        sector_multiple: multiple,
        ipo_price: starting_value,
        is_ipo: true,
        ipo_tick: tick,
        headcount: 0.0,
        average_wage: 0.0,
        labor_expense: 0.0,
        debt_outstanding: 0.0,
        interest_expense: 0.0,
        corporate_tax_paid: 0.0,
        capex: 0.0,
        macro_demand_factor: 1.0,
        is_public: true,
        stage: "Public".to_string(),
        private_valuation: 0.0,
        drift,
        volatility: vol,
        jump_params,
        reporting: reporting_params,
        rng: None,
        reporting_bias_accum: 0.0,
        session_ticks_remaining: 0,
        session_drift: 0.0,
        momentum: 0.0,
        current_tick: 0,
    };
    init_macro_metrics(&mut co, rng);
    co.rng = Some(rng.clone());
    co
}

pub fn generate_private_enterprise(
    sector: Sector,
    seed_revenue: f64,
    tick: usize,
    rng: &mut StdRng,
    used_symbols: &mut HashSet<String>,
) -> Company {
    let mut c = generate_ipo_company(sector, CapTier::SmallCap, seed_revenue, tick, rng, used_symbols);
    c.is_public = false;
    c.is_ipo = false;
    c.stage = if seed_revenue > 25_000_000.0 {
        "Growth".to_string()
    } else {
        "Seed".to_string()
    };
    c.private_valuation = c.annual_revenue * c.sector_multiple;
    c
}

fn init_macro_metrics(c: &mut Company, rng: &mut StdRng) {
    c.macro_demand_factor = 1.0;

    let (mut rev_per_emp, mut avg_wage, mut debt_to_rev, capex_to_rev) = match c.sector {
        Sector::InformationTechnology => (600_000.0, 140_000.0, 0.35, 0.08),
        Sector::Financials => (450_000.0, 130_000.0, 1.20, 0.04),
        Sector::HealthCare => (480_000.0, 115_000.0, 0.50, 0.09),
        Sector::Energy => (850_000.0, 110_000.0, 0.65, 0.13),
        Sector::ConsumerDiscretionary => (240_000.0, 52_000.0, 0.60, 0.05),
        Sector::ConsumerStaples => (290_000.0, 55_000.0, 0.55, 0.04),
        Sector::Industrials => (330_000.0, 80_000.0, 0.60, 0.06),
        Sector::Materials => (380_000.0, 78_000.0, 0.70, 0.08),
        Sector::CommunicationServices => (520_000.0, 120_000.0, 0.85, 0.11),
        Sector::Utilities => (650_000.0, 95_000.0, 1.40, 0.15),
        Sector::RealEstate => (450_000.0, 90_000.0, 1.50, 0.10),
    };

    rev_per_emp *= 0.85 + rng.gen::<f64>() * 0.30;
    avg_wage *= 0.90 + rng.gen::<f64>() * 0.20;
    debt_to_rev *= 0.80 + rng.gen::<f64>() * 0.40;

    c.headcount = (c.annual_revenue / rev_per_emp).round().max(25.0);
    c.average_wage = avg_wage;
    c.labor_expense = c.headcount * c.average_wage;
    c.debt_outstanding = c.annual_revenue * debt_to_rev;
    c.interest_expense = c.debt_outstanding * 0.045;
    c.capex = c.annual_revenue * capex_to_rev;

    let taxable = (c.annual_revenue * c.net_margin * 1.35 - c.interest_expense).max(0.0);
    c.corporate_tax_paid = taxable * 0.21;
}

pub fn generate_universe(n: usize, master_seed: i64, params: &GenerationParams) -> Vec<Company> {
    let mut seeder = StdRng::seed_from_u64(master_seed as u64);
    let mut used_symbols = HashSet::new();
    let mut companies = Vec::with_capacity(n);

    let cap_weights = [
        (CapTier::MegaCap, 0.05),
        (CapTier::LargeCap, 0.15),
        (CapTier::MidCap, 0.35),
        (CapTier::SmallCap, 0.45),
    ];

    for i in 0..n {
        let sector = ALL_SECTORS[i % ALL_SECTORS.len()];

        let r = seeder.gen::<f64>();
        let mut cumulative = 0.0;
        let mut tier = CapTier::SmallCap;
        for (t, w) in cap_weights {
            cumulative += w;
            if r <= cumulative {
                tier = t;
                break;
            }
        }

        let sub_seed = seeder.gen::<u64>();
        let mut company_rng = StdRng::seed_from_u64(sub_seed);

        let c = generate_company(sector, tier, params, &mut company_rng, &mut used_symbols);
        companies.push(c);
    }

    companies
}
