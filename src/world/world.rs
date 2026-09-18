use chrono::{DateTime, Datelike, TimeZone, Utc};
use rand::rngs::StdRng;
use rand::{Rng, SeedableRng};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// SimDate represents a simulated calendar date (Y/M/D).
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub struct SimDate {
    pub year: i32,
    pub month: u32,
    pub day: u32,
}

impl std::fmt::Display for SimDate {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let month_name = match self.month {
            1 => "Jan",
            2 => "Feb",
            3 => "Mar",
            4 => "Apr",
            5 => "May",
            6 => "Jun",
            7 => "Jul",
            8 => "Aug",
            9 => "Sep",
            10 => "Oct",
            11 => "Nov",
            _ => "Dec",
        };
        write!(f, "{} {:02}, {}", month_name, self.day, self.year)
    }
}

/// Simulation calendar epoch: Jan 1, 2020 00:00:00 UTC.
pub fn sim_epoch() -> DateTime<Utc> {
    Utc.with_ymd_and_hms(2020, 1, 1, 0, 0, 0).unwrap()
}

/// Convert simulation-time milliseconds into a SimDate.
pub fn date_from_millis(ms: i64) -> SimDate {
    let t = sim_epoch() + chrono::Duration::milliseconds(ms);
    SimDate {
        year: t.year(),
        month: t.month(),
        day: t.day(),
    }
}

/// Returns a human-readable time-per-real-second label for a speed multiplier.
pub fn speed_label(speed: f64) -> String {
    if speed <= 0.0 {
        return "PAUSED".to_string();
    }
    let days_per_sec = speed; // 1.0x = 1 day/s
    if days_per_sec >= 365.0 {
        format!("~{:.1} yr/s", days_per_sec / 365.0)
    } else if days_per_sec >= 30.0 {
        format!("~{:.1} mo/s", days_per_sec / 30.0)
    } else if days_per_sec >= 7.0 {
        format!("~{:.1} wk/s", days_per_sec / 7.0)
    } else if days_per_sec >= 1.0 {
        format!("~{:.1} day/s", days_per_sec)
    } else {
        let hours_per_sec = days_per_sec * 24.0;
        format!("~{:.0} hr/s", hours_per_sec)
    }
}

/// DiplomaticStance describes the political relationship between two countries.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum DiplomaticStance {
    Neutral = 0,
    Friendly = 1,
    Allied = 2,
    Cold = 3,
    Hostile = 4,
}

impl std::fmt::Display for DiplomaticStance {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let s = match self {
            Self::Neutral => "Neutral",
            Self::Friendly => "Friendly",
            Self::Allied => "Allied",
            Self::Cold => "Cold",
            Self::Hostile => "Hostile",
        };
        write!(f, "{}", s)
    }
}

impl DiplomaticStance {
    pub fn step_closer_to_allied(&self) -> Self {
        match self {
            Self::Hostile => Self::Cold,
            Self::Cold => Self::Neutral,
            Self::Neutral => Self::Friendly,
            Self::Friendly => Self::Allied,
            Self::Allied => Self::Allied,
        }
    }
}

/// Bilateral foreign policy relation between home and a foreign country.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Relation {
    pub tariff_rate: f64,
    pub stance: DiplomaticStance,
    pub trade_volume: f64,
    pub aid_flow: f64,
    pub sanctions_level: i32,
}

/// Region represents a sub-national economic zone (province/state).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Region {
    pub name: String,
    pub gdp: f64,
    pub tax_revenue: f64,
    pub unemployment_rate: f64,
    pub population: f64,
    pub sector_strengths: HashMap<String, f64>,
    pub budget: f64,
    pub debt: f64,
}

/// ForeignCountry is a procedurally generated sovereign state.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ForeignCountry {
    pub id: String,
    pub name: String,
    pub currency: String,
    pub flag_code: String,
    pub gdp: f64,
    pub gdp_growth: f64,
    pub inflation: f64,
    pub interest_rate: f64,
    pub unemployment: f64,
    pub population: f64,
    pub trade_balance: f64,
    pub currency_strength: f64,
    pub national_debt: f64,
    pub debt_to_gdp: f64,
    pub relation: Relation,

    // Detailed Demographic, Labor & Geopolitical Properties
    pub labor_force: f64,
    pub employed_workers: f64,
    pub gdp_per_capita: f64,
    pub average_hourly_wage: f64,
    pub job_openings: f64,
    pub productivity_index: f64,
    pub geopolitical_power: f64,
    pub sanctioned_by_us: bool,
    pub sanctioned_us: bool,
    pub sanctions_summary: String,
    pub is_reserve_currency: bool,

    #[serde(skip)]
    pub rng: Option<StdRng>,
}

impl ForeignCountry {
    pub fn tick(&mut self, tick_frac_of_year: f64) {
        let rng = self.rng.get_or_insert_with(|| StdRng::seed_from_u64(42));
        let r1: f64 = rng.gen();
        let growth_shock = (r1 - 0.495) * 0.006;
        self.gdp_growth = (self.gdp_growth + growth_shock).clamp(-0.06, 0.10);
        self.gdp *= 1.0 + self.gdp_growth * tick_frac_of_year;

        let r2: f64 = rng.gen();
        let infl_shock = (r2 - 0.5) * 0.001;
        self.inflation = (self.inflation + infl_shock).clamp(0.005, 0.12);

        self.national_debt *= 1.0 + 0.02 * tick_frac_of_year;
        if self.gdp > 0.0 {
            self.debt_to_gdp = self.national_debt / self.gdp;
        }

        // Maintain and drift demographic & labor metrics
        self.labor_force = self.population * 0.635;
        self.employed_workers = self.labor_force * (1.0 - self.unemployment);
        if self.population > 0.0 {
            self.gdp_per_capita = self.gdp / self.population;
            self.average_hourly_wage = (self.gdp_per_capita / 2080.0).clamp(6.5, 120.0);
        }
        self.job_openings = (self.labor_force * 0.041 * (1.0 + self.gdp_growth)).max(10_000.0);

        // Geopolitical power score (10 - 99)
        let gdp_power = (self.gdp / 500_000_000_000.0 * 50.0).clamp(10.0, 55.0);
        let debt_penalty = (self.debt_to_gdp * 12.0).clamp(0.0, 18.0);
        let prod_boost = (self.productivity_index - 100.0) * 0.35;
        self.geopolitical_power = (gdp_power + self.currency_strength * 12.0 - debt_penalty + prod_boost).clamp(12.0, 96.0);

        // Sanctions impact
        if self.sanctioned_by_us {
            self.gdp_growth = (self.gdp_growth - 0.008).clamp(-0.12, 0.08);
            self.relation.trade_volume *= 0.998;
            self.currency_strength = (self.currency_strength * 0.9995).max(0.20);
        }
        if self.sanctioned_us {
            self.relation.trade_volume *= 0.997;
        }

        if self.relation.stance >= DiplomaticStance::Cold {
            self.relation.trade_volume *= 0.9999;
        } else if self.relation.stance == DiplomaticStance::Allied {
            self.relation.trade_volume *= 1.00005;
        }
    }
}

/// Bilateral Financial Loan Request from a foreign sovereign state.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ForeignLoanRequest {
    pub id: String,
    pub country_id: String,
    pub country_name: String,
    pub currency: String,
    pub amount: f64,          // In home currency (e.g. $80,000,000)
    pub interest_rate: f64,   // e.g. 0.068 (6.8%)
    pub term_ticks: i64,      // duration of loan
    pub purpose: String,      // narrative reason
    pub status: String,       // "pending", "accepted", "declined", "matured_repaid"
    pub ticks_remaining: i64,
    pub interest_earned: f64,
    pub outcome_message: String,
}

/// Bilateral Diplomatic, Trade and Defense Contract proposed by a foreign nation.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BilateralContract {
    pub id: String,
    pub country_id: String,
    pub country_name: String,
    pub contract_type: String, // "Trade Agreement", "Strategic Energy Pact", "Infrastructure Concession", "Commodity Export Accord"
    pub title: String,
    pub terms: String,
    pub annual_revenue_gain: f64,
    pub export_capacity_boost: f64,
    pub duration_ticks: i64,
    pub status: String, // "pending", "active", "declined", "expired"
    pub outcome: String,
}

/// Live Forex Pair between home currency and a foreign sovereign currency.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ForexPair {
    pub symbol: String,        // e.g. "CRN/VAL"
    pub base_currency: String, // "CRN"
    pub quote_currency: String,// "VAL"
    pub rate: f64,             // exchange rate
    pub change_24h: f64,       // % change
    pub high_24h: f64,
    pub low_24h: f64,
    pub base_rate: f64,
}

/// WorldStock represents an equity listed in an international sovereign exchange.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorldStock {
    pub ticker: String,
    pub name: String,
    pub country_id: String,
    pub country_name: String,
    pub sector: String,
    pub price: f64,
    pub change_pct: f64,
    pub market_cap: f64,
    pub pe_ratio: f64,
    pub volume: f64,
}

/// CountryIndex represents a sovereign benchmark equity index for a nation.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CountryIndex {
    pub symbol: String,         // e.g. "CRN-50", "VAL-30", "SOL-20", "ZTH-40", etc.
    pub name: String,           // e.g. "Sovereign 50 Composite"
    pub country_id: String,     // "HOME" or "VAL", "SOL", "ZTH", etc.
    pub country_name: String,   // "Republic of Eldoria"
    pub price: f64,             // e.g. 5240.50
    pub change_24h: f64,        // e.g. +1.24%
    pub high_24h: f64,
    pub low_24h: f64,
    pub volume: f64,
    pub constituents_count: usize,
    pub base_price: f64,
}

/// HomeCountry is the user's fledgling sovereign country.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HomeCountry {
    pub name: String,
    pub currency: String,
    pub flag_code: String,
    pub founded: i32,
    pub regions: Vec<Region>,
    pub configured: bool,
}

impl HomeCountry {
    pub fn is_configured(&self) -> bool {
        self.configured && !self.name.is_empty()
    }

    pub fn configure(&mut self, name: &str, currency: &str, flag_code: &str, founded: i32) {
        self.name = name.trim().to_string();
        self.currency = currency.trim().to_uppercase();
        self.flag_code = flag_code.trim().to_uppercase();
        self.founded = founded;
        self.configured = true;
    }
}

/// ForeignPolicyAction records a diplomatic action taken by the player.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ForeignPolicyAction {
    pub tick: i64,
    pub country_id: String,
    pub action: String,
    pub detail: String,
}

/// World holds the entire multi-country simulation world.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct World {
    pub home: HomeCountry,
    pub foreign: HashMap<String, ForeignCountry>,
    pub loan_requests: Vec<ForeignLoanRequest>,
    pub bilateral_contracts: Vec<BilateralContract>,
    pub forex_pairs: Vec<ForexPair>,
    pub world_stocks: Vec<WorldStock>,
    pub country_indices: Vec<CountryIndex>,
    pub policy_log: Vec<ForeignPolicyAction>,
    pub reserve_currency: String,
    pub reserve_country_name: String,
    pub tick_counter: i64,
    #[serde(skip)]
    pub rng: Option<StdRng>,
}

fn format_money(v: f64) -> String {
    if v >= 1e12 {
        format!("${:.2}T", v / 1e12)
    } else if v >= 1e9 {
        format!("${:.2}B", v / 1e9)
    } else if v >= 1e6 {
        format!("${:.2}M", v / 1e6)
    } else {
        format!("${:.0}", v)
    }
}

pub fn default_regions() -> Vec<Region> {
    vec![
        Region {
            name: "Capital District".into(),
            gdp: 5_500_000_000.0,
            tax_revenue: 250_000_000.0,
            unemployment_rate: 0.038,
            population: 2_500_000.0,
            budget: 180_000_000.0,
            debt: 0.0,
            sector_strengths: [
                ("Financials".into(), 1.4),
                ("Information Technology".into(), 1.3),
                ("Health Care".into(), 1.2),
            ]
            .into_iter()
            .collect(),
        },
        Region {
            name: "Highland Industrial Zone".into(),
            gdp: 4_200_000_000.0,
            tax_revenue: 190_000_000.0,
            unemployment_rate: 0.042,
            population: 2_200_000.0,
            budget: 140_000_000.0,
            debt: 0.0,
            sector_strengths: [
                ("Industrials".into(), 1.4),
                ("Materials".into(), 1.3),
                ("Consumer Staples".into(), 1.1),
            ]
            .into_iter()
            .collect(),
        },
        Region {
            name: "Maritime Coastal Hub".into(),
            gdp: 3_800_000_000.0,
            tax_revenue: 160_000_000.0,
            unemployment_rate: 0.039,
            population: 1_800_000.0,
            budget: 120_000_000.0,
            debt: 0.0,
            sector_strengths: [
                ("Energy".into(), 1.4),
                ("Industrials".into(), 1.2),
                ("Consumer Discretionary".into(), 1.2),
            ]
            .into_iter()
            .collect(),
        },
        Region {
            name: "Verdant River Basin".into(),
            gdp: 2_500_000_000.0,
            tax_revenue: 100_000_000.0,
            unemployment_rate: 0.041,
            population: 1_500_000.0,
            budget: 80_000_000.0,
            debt: 0.0,
            sector_strengths: [
                ("Consumer Staples".into(), 1.5),
                ("Utilities".into(), 1.3),
                ("Real Estate".into(), 1.1),
            ]
            .into_iter()
            .collect(),
        },
    ]
}

fn create_foreign_nation(
    id: &str,
    name: &str,
    currency: &str,
    flag_code: &str,
    gdp: f64,
    gdp_growth: f64,
    inflation: f64,
    interest_rate: f64,
    unemployment: f64,
    population: f64,
    trade_balance: f64,
    currency_strength: f64,
    national_debt: f64,
    productivity_index: f64,
    tariff_rate: f64,
    stance: DiplomaticStance,
    trade_volume: f64,
) -> ForeignCountry {
    let labor_force = population * 0.635;
    let employed_workers = labor_force * (1.0 - unemployment);
    let gdp_per_capita = if population > 0.0 { gdp / population } else { 0.0 };
    let average_hourly_wage = (gdp_per_capita / 2080.0).clamp(6.5, 120.0);
    let job_openings = (labor_force * 0.041).max(10_000.0);
    let debt_to_gdp = if gdp > 0.0 { national_debt / gdp } else { 0.0 };

    let gdp_power = (gdp / 500_000_000_000.0 * 50.0).clamp(10.0, 55.0);
    let debt_penalty = (debt_to_gdp * 12.0).clamp(0.0, 18.0);
    let prod_boost = (productivity_index - 100.0) * 0.35;
    let geopolitical_power = (gdp_power + currency_strength * 12.0 - debt_penalty + prod_boost).clamp(12.0, 96.0);

    ForeignCountry {
        id: id.into(),
        name: name.into(),
        currency: currency.into(),
        flag_code: flag_code.into(),
        gdp,
        gdp_growth,
        inflation,
        interest_rate,
        unemployment,
        population,
        trade_balance,
        currency_strength,
        national_debt,
        debt_to_gdp,
        relation: Relation {
            tariff_rate,
            stance,
            trade_volume,
            aid_flow: 0.0,
            sanctions_level: 0,
        },
        labor_force,
        employed_workers,
        gdp_per_capita,
        average_hourly_wage,
        job_openings,
        productivity_index,
        geopolitical_power,
        sanctioned_by_us: false,
        sanctioned_us: false,
        sanctions_summary: "No active sanctions.".into(),
        is_reserve_currency: id == "ZTH",
        rng: None,
    }
}

impl World {
    pub fn new(seed: u64) -> Self {
        let mut rng = StdRng::seed_from_u64(seed);
        let mut foreign = HashMap::new();

        // 8 procedurally generated sovereign states with custom currencies & identities
        let foreign_nations = vec![
            create_foreign_nation(
                "VAL", "Republic of Valoria", "VAL", "VL",
                340_000_000_000.0, 0.028, 0.024, 0.035, 0.045, 32_000_000.0,
                18_000_000_000.0, 1.12, 190_000_000_000.0, 115.0,
                0.025, DiplomaticStance::Allied, 450_000_000.0,
            ),
            create_foreign_nation(
                "SOL", "Solaris Commonwealth", "SOL", "SL",
                190_000_000_000.0, 0.035, 0.029, 0.040, 0.051, 24_000_000.0,
                8_000_000_000.0, 0.85, 95_000_000_000.0, 108.0,
                0.030, DiplomaticStance::Friendly, 320_000_000.0,
            ),
            create_foreign_nation(
                "ZTH", "Zenthian Federation", "ZTH", "ZN",
                460_000_000_000.0, 0.019, 0.038, 0.052, 0.058, 55_000_000.0,
                24_000_000_000.0, 2.45, 310_000_000_000.0, 122.0,
                0.065, DiplomaticStance::Cold, 210_000_000.0,
            ),
            create_foreign_nation(
                "AUR", "Aurelia Grand Duchy", "AUR", "AU",
                110_000_000_000.0, 0.022, 0.018, 0.025, 0.032, 8_000_000.0,
                14_000_000_000.0, 0.68, 42_000_000_000.0, 130.0,
                0.020, DiplomaticStance::Friendly, 280_000_000.0,
            ),
            create_foreign_nation(
                "KLU", "Kaelen Maritime Union", "KLU", "KL",
                275_000_000_000.0, 0.031, 0.032, 0.045, 0.048, 28_000_000.0,
                -5_000_000_000.0, 1.42, 155_000_000_000.0, 104.0,
                0.035, DiplomaticStance::Neutral, 190_000_000.0,
            ),
            create_foreign_nation(
                "OKH", "Oakhaven Free State", "OKH", "OK",
                85_000_000_000.0, 0.042, 0.041, 0.055, 0.039, 12_000_000.0,
                6_000_000_000.0, 3.15, 38_000_000_000.0, 94.0,
                0.040, DiplomaticStance::Neutral, 140_000_000.0,
            ),
            create_foreign_nation(
                "VDT", "Verdant Republic", "VDT", "VR",
                135_000_000_000.0, 0.039, 0.026, 0.038, 0.042, 16_000_000.0,
                9_000_000_000.0, 1.78, 68_000_000_000.0, 112.0,
                0.025, DiplomaticStance::Friendly, 240_000_000.0,
            ),
            create_foreign_nation(
                "BRL", "Boreal Protectorate", "BRL", "BP",
                165_000_000_000.0, 0.025, 0.022, 0.032, 0.040, 19_000_000.0,
                11_000_000_000.0, 0.94, 82_000_000_000.0, 102.0,
                0.030, DiplomaticStance::Neutral, 160_000_000.0,
            ),
            create_foreign_nation(
                "YMT", "Yamato Archipelago", "YMT", "YM",
                520_000_000_000.0, 0.021, 0.016, 0.012, 0.028, 62_000_000.0,
                26_000_000_000.0, 1.95, 280_000_000_000.0, 128.0,
                0.020, DiplomaticStance::Friendly, 510_000_000.0,
            ),
            create_foreign_nation(
                "SGH", "Singha Federation", "SGH", "SG",
                480_000_000_000.0, 0.042, 0.028, 0.038, 0.034, 45_000_000.0,
                32_000_000_000.0, 1.35, 210_000_000_000.0, 134.0,
                0.015, DiplomaticStance::Friendly, 480_000_000.0,
            ),
            create_foreign_nation(
                "OAS", "Oasis Emirates", "OAS", "OE",
                390_000_000_000.0, 0.036, 0.031, 0.042, 0.031, 14_000_000.0,
                45_000_000_000.0, 2.80, 110_000_000_000.0, 110.0,
                0.025, DiplomaticStance::Neutral, 390_000_000.0,
            ),
            create_foreign_nation(
                "KEM", "Kemerov Union", "KEM", "KM",
                310_000_000_000.0, 0.014, 0.052, 0.075, 0.062, 42_000_000.0,
                -12_000_000_000.0, 0.72, 260_000_000_000.0, 98.0,
                0.060, DiplomaticStance::Cold, 175_000_000.0,
            ),
            create_foreign_nation(
                "CAS", "Cascadia Free State", "CAS", "CS",
                220_000_000_000.0, 0.038, 0.024, 0.034, 0.039, 21_000_000.0,
                14_000_000_000.0, 1.15, 95_000_000_000.0, 122.0,
                0.020, DiplomaticStance::Allied, 320_000_000.0,
            ),
            create_foreign_nation(
                "AND", "Andesia Republic", "AND", "AR",
                140_000_000_000.0, 0.032, 0.045, 0.058, 0.055, 18_000_000.0,
                7_000_000_000.0, 0.65, 78_000_000_000.0, 101.0,
                0.035, DiplomaticStance::Friendly, 190_000_000.0,
            ),
            create_foreign_nation(
                "NDL", "Nile Delta League", "NDL", "NL",
                125_000_000_000.0, 0.039, 0.048, 0.062, 0.059, 29_000_000.0,
                4_000_000_000.0, 0.55, 65_000_000_000.0, 96.0,
                0.030, DiplomaticStance::Neutral, 160_000_000.0,
            ),
            create_foreign_nation(
                "SAH", "Saharan Alliance", "SAH", "SH",
                95_000_000_000.0, 0.045, 0.039, 0.049, 0.065, 15_000_000.0,
                5_000_000_000.0, 0.45, 42_000_000_000.0, 93.0,
                0.040, DiplomaticStance::Friendly, 130_000_000.0,
            ),
        ];

        for mut fc in foreign_nations {
            fc.rng = Some(StdRng::seed_from_u64(rng.gen()));
            if fc.gdp > 0.0 {
                fc.debt_to_gdp = fc.national_debt / fc.gdp;
            }
            foreign.insert(fc.id.clone(), fc);
        }

        // Live Forex pairs against home currency CRN
        let forex_pairs = vec![
            ForexPair {
                symbol: "CRN/VAL".into(),
                base_currency: "CRN".into(),
                quote_currency: "VAL".into(),
                rate: 1.1240,
                change_24h: 0.35,
                high_24h: 1.1310,
                low_24h: 1.1190,
                base_rate: 1.1240,
            },
            ForexPair {
                symbol: "CRN/SOL".into(),
                base_currency: "CRN".into(),
                quote_currency: "SOL".into(),
                rate: 0.8520,
                change_24h: -0.22,
                high_24h: 0.8590,
                low_24h: 0.8490,
                base_rate: 0.8520,
            },
            ForexPair {
                symbol: "CRN/ZTH".into(),
                base_currency: "CRN".into(),
                quote_currency: "ZTH".into(),
                rate: 2.4580,
                change_24h: 0.78,
                high_24h: 2.4720,
                low_24h: 2.4410,
                base_rate: 2.4580,
            },
            ForexPair {
                symbol: "CRN/AUR".into(),
                base_currency: "CRN".into(),
                quote_currency: "AUR".into(),
                rate: 0.6840,
                change_24h: -0.15,
                high_24h: 0.6890,
                low_24h: 0.6810,
                base_rate: 0.6840,
            },
            ForexPair {
                symbol: "CRN/KLU".into(),
                base_currency: "CRN".into(),
                quote_currency: "KLU".into(),
                rate: 1.4230,
                change_24h: 0.12,
                high_24h: 1.4310,
                low_24h: 1.4180,
                base_rate: 1.4230,
            },
            ForexPair {
                symbol: "CRN/OKH".into(),
                base_currency: "CRN".into(),
                quote_currency: "OKH".into(),
                rate: 3.1550,
                change_24h: 0.45,
                high_24h: 3.1750,
                low_24h: 3.1420,
                base_rate: 3.1550,
            },
            ForexPair {
                symbol: "CRN/VDT".into(),
                base_currency: "CRN".into(),
                quote_currency: "VDT".into(),
                rate: 1.7820,
                change_24h: -0.40,
                high_24h: 1.7940,
                low_24h: 1.7750,
                base_rate: 1.7820,
            },
            ForexPair {
                symbol: "CRN/BRL".into(),
                base_currency: "CRN".into(),
                quote_currency: "BRL".into(),
                rate: 0.9410,
                change_24h: 0.18,
                high_24h: 0.9460,
                low_24h: 0.9380,
                base_rate: 0.9410,
            },
            ForexPair {
                symbol: "CRN/YMT".into(),
                base_currency: "CRN".into(),
                quote_currency: "YMT".into(),
                rate: 1.9540,
                change_24h: 0.42,
                high_24h: 1.9650,
                low_24h: 1.9420,
                base_rate: 1.9540,
            },
            ForexPair {
                symbol: "CRN/SGH".into(),
                base_currency: "CRN".into(),
                quote_currency: "SGH".into(),
                rate: 1.3520,
                change_24h: 0.25,
                high_24h: 1.3590,
                low_24h: 1.3450,
                base_rate: 1.3520,
            },
            ForexPair {
                symbol: "CRN/OAS".into(),
                base_currency: "CRN".into(),
                quote_currency: "OAS".into(),
                rate: 2.8050,
                change_24h: -0.32,
                high_24h: 2.8210,
                low_24h: 2.7910,
                base_rate: 2.8050,
            },
            ForexPair {
                symbol: "CRN/KEM".into(),
                base_currency: "CRN".into(),
                quote_currency: "KEM".into(),
                rate: 0.7240,
                change_24h: -0.65,
                high_24h: 0.7320,
                low_24h: 0.7180,
                base_rate: 0.7240,
            },
            ForexPair {
                symbol: "CRN/CAS".into(),
                base_currency: "CRN".into(),
                quote_currency: "CAS".into(),
                rate: 1.1520,
                change_24h: 0.15,
                high_24h: 1.1590,
                low_24h: 1.1460,
                base_rate: 1.1520,
            },
            ForexPair {
                symbol: "CRN/AND".into(),
                base_currency: "CRN".into(),
                quote_currency: "AND".into(),
                rate: 0.6540,
                change_24h: 0.38,
                high_24h: 0.6610,
                low_24h: 0.6480,
                base_rate: 0.6540,
            },
            ForexPair {
                symbol: "CRN/NDL".into(),
                base_currency: "CRN".into(),
                quote_currency: "NDL".into(),
                rate: 0.5520,
                change_24h: -0.12,
                high_24h: 0.5580,
                low_24h: 0.5470,
                base_rate: 0.5520,
            },
            ForexPair {
                symbol: "CRN/SAH".into(),
                base_currency: "CRN".into(),
                quote_currency: "SAH".into(),
                rate: 0.4510,
                change_24h: 0.22,
                high_24h: 0.4570,
                low_24h: 0.4460,
                base_rate: 0.4510,
            },
        ];

        // International World Stocks listed across foreign sovereign states
        let world_stocks = vec![
            WorldStock {
                ticker: "VAL_AERO".into(),
                name: "Valoria Aerospace Systems".into(),
                country_id: "VAL".into(),
                country_name: "Republic of Valoria".into(),
                sector: "Industrials".into(),
                price: 142.50,
                change_pct: 1.25,
                market_cap: 45_600_000_000.0,
                pe_ratio: 24.2,
                volume: 850_000.0,
            },
            WorldStock {
                ticker: "VAL_CHIP".into(),
                name: "Valoria Silico-Nanotech".into(),
                country_id: "VAL".into(),
                country_name: "Republic of Valoria".into(),
                sector: "Information Technology".into(),
                price: 210.80,
                change_pct: 2.80,
                market_cap: 62_400_000_000.0,
                pe_ratio: 32.5,
                volume: 1_200_000.0,
            },
            WorldStock {
                ticker: "SOL_SOLR".into(),
                name: "Solaris Helios Power Grid".into(),
                country_id: "SOL".into(),
                country_name: "Solaris Commonwealth".into(),
                sector: "Utilities".into(),
                price: 88.40,
                change_pct: -0.45,
                market_cap: 28_200_000_000.0,
                pe_ratio: 18.4,
                volume: 620_000.0,
            },
            WorldStock {
                ticker: "SOL_AGRI".into(),
                name: "Commonwealth Agri-Synth".into(),
                country_id: "SOL".into(),
                country_name: "Solaris Commonwealth".into(),
                sector: "Consumer Staples".into(),
                price: 54.20,
                change_pct: 0.85,
                market_cap: 19_500_000_000.0,
                pe_ratio: 15.6,
                volume: 410_000.0,
            },
            WorldStock {
                ticker: "ZTH_STEL".into(),
                name: "Zenthia Heavy Metallurgy".into(),
                country_id: "ZTH".into(),
                country_name: "Zenthian Federation".into(),
                sector: "Materials".into(),
                price: 115.00,
                change_pct: -1.10,
                market_cap: 58_000_000_000.0,
                pe_ratio: 12.8,
                volume: 980_000.0,
            },
            WorldStock {
                ticker: "ZTH_RAIL".into(),
                name: "Federated Trans-Continental Rail".into(),
                country_id: "ZTH".into(),
                country_name: "Zenthian Federation".into(),
                sector: "Industrials".into(),
                price: 72.30,
                change_pct: 0.35,
                market_cap: 34_100_000_000.0,
                pe_ratio: 14.1,
                volume: 530_000.0,
            },
            WorldStock {
                ticker: "AUR_BANK".into(),
                name: "Aurelia Private Merchant Bancorp".into(),
                country_id: "AUR".into(),
                country_name: "Aurelia Grand Duchy".into(),
                sector: "Financials".into(),
                price: 320.00,
                change_pct: 0.95,
                market_cap: 52_000_000_000.0,
                pe_ratio: 16.9,
                volume: 380_000.0,
            },
            WorldStock {
                ticker: "AUR_LUXE".into(),
                name: "Grand Duchy Haute Horlogerie".into(),
                country_id: "AUR".into(),
                country_name: "Aurelia Grand Duchy".into(),
                sector: "Consumer Discretionary".into(),
                price: 185.50,
                change_pct: 1.65,
                market_cap: 22_400_000_000.0,
                pe_ratio: 28.3,
                volume: 240_000.0,
            },
            WorldStock {
                ticker: "KLU_PORT".into(),
                name: "Kaelen Deepwater Terminal Ports".into(),
                country_id: "KLU".into(),
                country_name: "Kaelen Maritime Union".into(),
                sector: "Industrials".into(),
                price: 94.60,
                change_pct: 0.70,
                market_cap: 36_800_000_000.0,
                pe_ratio: 17.5,
                volume: 710_000.0,
            },
            WorldStock {
                ticker: "KLU_ENRG".into(),
                name: "Kaelen Oceanic LNG Consortium".into(),
                country_id: "KLU".into(),
                country_name: "Kaelen Maritime Union".into(),
                sector: "Energy".into(),
                price: 135.20,
                change_pct: -0.80,
                market_cap: 48_100_000_000.0,
                pe_ratio: 11.4,
                volume: 890_000.0,
            },
            WorldStock {
                ticker: "OKH_MINE".into(),
                name: "Oakhaven Sovereign Rare Earths".into(),
                country_id: "OKH".into(),
                country_name: "Oakhaven Free State".into(),
                sector: "Materials".into(),
                price: 64.80,
                change_pct: 3.40,
                market_cap: 18_200_000_000.0,
                pe_ratio: 13.2,
                volume: 950_000.0,
            },
            WorldStock {
                ticker: "VDT_PHAR".into(),
                name: "Verdant Bio-Genetic Therapeutics".into(),
                country_id: "VDT".into(),
                country_name: "Verdant Republic".into(),
                sector: "Health Care".into(),
                price: 168.00,
                change_pct: 1.85,
                market_cap: 38_500_000_000.0,
                pe_ratio: 31.0,
                volume: 670_000.0,
            },
            WorldStock {
                ticker: "BRL_CRYO".into(),
                name: "Boreal Cryogenic Defense Systems".into(),
                country_id: "BRL".into(),
                country_name: "Boreal Protectorate".into(),
                sector: "Industrials".into(),
                price: 124.50,
                change_pct: 0.45,
                market_cap: 29_300_000_000.0,
                pe_ratio: 21.0,
                volume: 480_000.0,
            },
        ];

        // Initial Bilateral Loan Requests from foreign states
        let loan_requests = vec![
            ForeignLoanRequest {
                id: "REQ-701".into(),
                country_id: "SOL".into(),
                country_name: "Solaris Commonwealth".into(),
                currency: "SOL".into(),
                amount: 80_000_000.0,
                interest_rate: 0.068,
                term_ticks: 120,
                purpose: "Cross-Border Solar Array Interconnect Grid".into(),
                status: "pending".into(),
                ticks_remaining: 120,
                interest_earned: 0.0,
                outcome_message: "Pending sovereign directorate review".into(),
            },
            ForeignLoanRequest {
                id: "REQ-702".into(),
                country_id: "OKH".into(),
                country_name: "Oakhaven Free State".into(),
                currency: "OKH".into(),
                amount: 120_000_000.0,
                interest_rate: 0.075,
                term_ticks: 150,
                purpose: "Deep-Core Rare Earth Extraction & Refinery Facility".into(),
                status: "pending".into(),
                ticks_remaining: 150,
                interest_earned: 0.0,
                outcome_message: "Pending sovereign directorate review".into(),
            },
            ForeignLoanRequest {
                id: "REQ-703".into(),
                country_id: "VDT".into(),
                country_name: "Verdant Republic".into(),
                currency: "VDT".into(),
                amount: 65_000_000.0,
                interest_rate: 0.062,
                term_ticks: 90,
                purpose: "Bio-Pharmaceutical Cleanroom Development Initiative".into(),
                status: "pending".into(),
                ticks_remaining: 90,
                interest_earned: 0.0,
                outcome_message: "Pending sovereign directorate review".into(),
            },
        ];

        // Initial Bilateral Pacts, Trade Accords & Foreign Delegations
        let bilateral_contracts = vec![
            BilateralContract {
                id: "CON-401".into(),
                country_id: "AET".into(),
                country_name: "Aethelgard Dominion".into(),
                contract_type: "Preferential Trade Accord".into(),
                title: "Customs Duty Waiver on Heavy Machinery & Industrial Tools".into(),
                terms: "Eliminates tariffs on bilateral industrial shipments, expanding export revenue capacity.".into(),
                annual_revenue_gain: 420_000_000.0,
                export_capacity_boost: 0.12,
                duration_ticks: 350,
                status: "pending".into(),
                outcome: "Awaiting sovereign ratification".into(),
            },
            BilateralContract {
                id: "CON-402".into(),
                country_id: "VAL".into(),
                country_name: "Valoria Republic".into(),
                contract_type: "Strategic Energy Pact".into(),
                title: "Guaranteed Long-Term LNG & Hydrocarbon Import Corridor".into(),
                terms: "Secures favorable wholesale fuel pricing and guaranteed delivery quotas for domestic industry.".into(),
                annual_revenue_gain: 580_000_000.0,
                export_capacity_boost: 0.15,
                duration_ticks: 400,
                status: "pending".into(),
                outcome: "Awaiting sovereign ratification".into(),
            },
            BilateralContract {
                id: "CON-403".into(),
                country_id: "ZEP".into(),
                country_name: "Zephyria Isles".into(),
                contract_type: "Maritime Freight Protocol".into(),
                title: "Expedited Container Transshipment & Deepwater Port Access".into(),
                terms: "Provides automated customs clearance and priority berthing at maritime logistics hubs.".into(),
                annual_revenue_gain: 310_000_000.0,
                export_capacity_boost: 0.09,
                duration_ticks: 280,
                status: "pending".into(),
                outcome: "Awaiting sovereign ratification".into(),
            },
        ];

        // Sovereign Benchmark Country Indices
        let country_indices = vec![
            CountryIndex {
                symbol: "CRN-50".into(),
                name: "Sovereign 50 Composite".into(),
                country_id: "HOME".into(),
                country_name: "Republic of Eldoria".into(),
                price: 4850.20,
                change_24h: 1.25,
                high_24h: 4890.00,
                low_24h: 4820.50,
                volume: 45_200_000.0,
                constituents_count: 50,
                base_price: 4850.20,
            },
            CountryIndex {
                symbol: "VAL-30".into(),
                name: "Valoria Industrial 30".into(),
                country_id: "VAL".into(),
                country_name: "Republic of Valoria".into(),
                price: 3920.80,
                change_24h: 0.85,
                high_24h: 3945.00,
                low_24h: 3895.00,
                volume: 82_400_000.0,
                constituents_count: 30,
                base_price: 3920.80,
            },
            CountryIndex {
                symbol: "SOL-20".into(),
                name: "Solaris Clean Energy & Tech".into(),
                country_id: "SOL".into(),
                country_name: "Solaris Commonwealth".into(),
                price: 2480.40,
                change_24h: 1.45,
                high_24h: 2510.00,
                low_24h: 2465.00,
                volume: 54_100_000.0,
                constituents_count: 20,
                base_price: 2480.40,
            },
            CountryIndex {
                symbol: "ZTH-40".into(),
                name: "Zenthian Heavy 40 Index".into(),
                country_id: "ZTH".into(),
                country_name: "Zenthian Federation".into(),
                price: 6120.90,
                change_24h: -0.65,
                high_24h: 6180.00,
                low_24h: 6090.00,
                volume: 115_000_000.0,
                constituents_count: 40,
                base_price: 6120.90,
            },
            CountryIndex {
                symbol: "AUR-15".into(),
                name: "Aurelia Private Banking 15".into(),
                country_id: "AUR".into(),
                country_name: "Aurelia Grand Duchy".into(),
                price: 1940.50,
                change_24h: 0.35,
                high_24h: 1955.00,
                low_24h: 1930.00,
                volume: 28_600_000.0,
                constituents_count: 15,
                base_price: 1940.50,
            },
            CountryIndex {
                symbol: "KLU-25".into(),
                name: "Kaelen Maritime & Logistics".into(),
                country_id: "KLU".into(),
                country_name: "Kaelen Maritime Union".into(),
                price: 2860.10,
                change_24h: 0.20,
                high_24h: 2885.00,
                low_24h: 2840.00,
                volume: 62_300_000.0,
                constituents_count: 25,
                base_price: 2860.10,
            },
            CountryIndex {
                symbol: "OKH-20".into(),
                name: "Oakhaven Mineral & Resources".into(),
                country_id: "OKH".into(),
                country_name: "Oakhaven Free State".into(),
                price: 1420.30,
                change_24h: 2.10,
                high_24h: 1445.00,
                low_24h: 1405.00,
                volume: 19_500_000.0,
                constituents_count: 20,
                base_price: 1420.30,
            },
            CountryIndex {
                symbol: "VDT-15".into(),
                name: "Verdant Biotech & Agriculture".into(),
                country_id: "VDT".into(),
                country_name: "Verdant Republic".into(),
                price: 2180.70,
                change_24h: 1.15,
                high_24h: 2205.00,
                low_24h: 2160.00,
                volume: 32_800_000.0,
                constituents_count: 15,
                base_price: 2180.70,
            },
            CountryIndex {
                symbol: "BRL-20".into(),
                name: "Boreal Cryo & Defense Index".into(),
                country_id: "BRL".into(),
                country_name: "Boreal Protectorate".into(),
                price: 2740.00,
                change_24h: -0.45,
                high_24h: 2760.00,
                low_24h: 2725.00,
                volume: 41_200_000.0,
                constituents_count: 20,
                base_price: 2740.00,
            },
            CountryIndex {
                symbol: "YMT-225".into(),
                name: "Yamato Nikkei Composite".into(),
                country_id: "YMT".into(),
                country_name: "Yamato Archipelago".into(),
                price: 5480.20,
                change_24h: 0.75,
                high_24h: 5510.00,
                low_24h: 5440.00,
                volume: 98_500_000.0,
                constituents_count: 225,
                base_price: 5480.20,
            },
            CountryIndex {
                symbol: "SGH-50".into(),
                name: "Singha Semiconductor 50".into(),
                country_id: "SGH".into(),
                country_name: "Singha Federation".into(),
                price: 4320.60,
                change_24h: 1.65,
                high_24h: 4365.00,
                low_24h: 4280.00,
                volume: 85_200_000.0,
                constituents_count: 50,
                base_price: 4320.60,
            },
            CountryIndex {
                symbol: "OAS-30".into(),
                name: "Oasis Petrochemical Benchmark".into(),
                country_id: "OAS".into(),
                country_name: "Oasis Emirates".into(),
                price: 3680.40,
                change_24h: -0.25,
                high_24h: 3710.00,
                low_24h: 3660.00,
                volume: 62_000_000.0,
                constituents_count: 30,
                base_price: 3680.40,
            },
            CountryIndex {
                symbol: "KEM-40".into(),
                name: "Kemerov Industrial Heavy 40".into(),
                country_id: "KEM".into(),
                country_name: "Kemerov Union".into(),
                price: 1820.50,
                change_24h: -1.15,
                high_24h: 1845.00,
                low_24h: 1805.00,
                volume: 48_000_000.0,
                constituents_count: 40,
                base_price: 1820.50,
            },
            CountryIndex {
                symbol: "CAS-25".into(),
                name: "Cascadia Green Innovation 25".into(),
                country_id: "CAS".into(),
                country_name: "Cascadia Free State".into(),
                price: 2950.00,
                change_24h: 1.10,
                high_24h: 2980.00,
                low_24h: 2925.00,
                volume: 38_500_000.0,
                constituents_count: 25,
                base_price: 2950.00,
            },
        ];

        Self {
            home: HomeCountry {
                name: "Republic of Eldoria".into(),
                currency: "CRN".into(),
                flag_code: "EL".into(),
                founded: 2024,
                regions: default_regions(),
                configured: true,
            },
            foreign,
            loan_requests,
            bilateral_contracts,
            forex_pairs,
            world_stocks,
            country_indices,
            policy_log: Vec::new(),
            reserve_currency: "ZTH".into(),
            reserve_country_name: "Zenthian Federation".into(),
            tick_counter: 0,
            rng: Some(rng),
        }
    }

    pub fn tick(&mut self, tick_frac_of_year: f64) {
        self.tick_with_macro(tick_frac_of_year, 0.04, &[], 0.50, 45.0);
    }

    pub fn tick_with_macro(
        &mut self,
        tick_frac_of_year: f64,
        home_interest_rate: f64,
        active_policies: &[String],
        home_debt_to_gdp: f64,
        home_power: f64,
    ) {
        self.tick_counter += 1;
        for fc in self.foreign.values_mut() {
            fc.tick(tick_frac_of_year);
        }

        // Compute macro policy bias on home sovereign currency
        let mut home_currency_bias = 0.0;

        for pol in active_policies {
            match pol.as_str() {
                "rate_hike_50bp" => home_currency_bias += 0.0035,
                "rate_cut_50bp" => home_currency_bias -= 0.0030,
                "reserve_currency_ops" => home_currency_bias += 0.0060,
                "quantitative_easing" => home_currency_bias -= 0.0035,
                "unlimited_liquidity" => home_currency_bias -= 0.0120,
                "tariff_shield" => home_currency_bias += 0.0020,
                "autarky_tariff" => home_currency_bias -= 0.0025,
                "price_freeze" => home_currency_bias -= 0.0040,
                "debt_relief" => home_currency_bias -= 0.0050,
                "microfinance_act" => home_currency_bias += 0.0010,
                "regional_free_trade" => home_currency_bias += 0.0025,
                "global_trade_hegemony" => home_currency_bias += 0.0045,
                _ => {}
            }
        }

        let rate_differential = (home_interest_rate - 0.040).clamp(-0.06, 0.06) * 0.05;
        home_currency_bias += rate_differential;

        // Realistic debt impact on sovereign currency value
        // High debt burdens directly devalue the currency, while low debt strengthens it
        if home_debt_to_gdp > 0.45 {
            home_currency_bias -= (home_debt_to_gdp - 0.45) * 0.0075;
        } else {
            home_currency_bias += (0.45 - home_debt_to_gdp) * 0.0030;
        }

        if home_power > 50.0 {
            home_currency_bias += (home_power - 50.0) * 0.00008;
        }

        if let Some(rng) = self.rng.as_mut() {
            // Live Forex fluctuations reacting dynamically to sovereign policy
            for fx in &mut self.forex_pairs {
                let foreign_country = self.foreign.get(&fx.quote_currency);
                let (fc_interest, sanctioned_by_us, sanctioned_us) = if let Some(fc) = foreign_country {
                    (fc.interest_rate, fc.sanctioned_by_us, fc.sanctioned_us)
                } else {
                    (0.04, false, false)
                };

                let bilateral_interest_diff = (home_interest_rate - fc_interest) * 0.01;
                let sanction_effect = if sanctioned_by_us { 0.0005 } else if sanctioned_us { -0.0005 } else { 0.0 };

                let raw_noise = (rng.gen::<f64>() - 0.50) * 0.0012;
                let net_drift = raw_noise + (home_currency_bias * 0.02) + bilateral_interest_diff + sanction_effect;

                fx.rate = (fx.rate * (1.0 + net_drift)).max(0.01);
                fx.change_24h = ((fx.rate - fx.base_rate) / fx.base_rate) * 100.0;
                if fx.rate > fx.high_24h {
                    fx.high_24h = fx.rate;
                }
                if fx.rate < fx.low_24h {
                    fx.low_24h = fx.rate;
                }
            }

            // World Stocks fluctuations
            for ws in &mut self.world_stocks {
                let s_drift: f64 = (rng.gen::<f64>() - 0.498) * 0.002;
                ws.price = (ws.price * (1.0 + s_drift)).max(1.0);
                ws.change_pct = (ws.change_pct * 0.95) + (s_drift * 100.0 * 0.05);
                let w_vol_shock = 1.0 + (s_drift.abs() * 20.0) + (rng.gen::<f64>() - 0.5) * 0.05;
                ws.volume = (ws.volume * 0.97 + (ws.market_cap / 80_000.0) * 0.03 * w_vol_shock).clamp(100_000.0, 50_000_000.0);
            }

            // Update Sovereign Country Benchmark Indices
            for idx in &mut self.country_indices {
                let (growth_factor, prod_factor, policy_boost) = if idx.country_id == "HOME" {
                    let mut pol_boost = 0.0;
                    if active_policies.contains(&"semiconductor_fab".to_string()) { pol_boost += 0.0002; }
                    if active_policies.contains(&"ai_quantum_center".to_string()) { pol_boost += 0.0003; }
                    if active_policies.contains(&"high_speed_rail".to_string()) { pol_boost += 0.0002; }
                    if active_policies.contains(&"light_mfg_grants".to_string()) { pol_boost += 0.0002; }
                    if active_policies.contains(&"price_freeze".to_string()) { pol_boost -= 0.0004; }
                    if active_policies.contains(&"windfall_seizure".to_string()) { pol_boost -= 0.0005; }
                    (0.038, 1.05, pol_boost)
                } else if let Some(fc) = self.foreign.get(&idx.country_id) {
                    (fc.gdp_growth, fc.productivity_index / 100.0, 0.0)
                } else {
                    (0.02, 1.0, 0.0)
                };

                let annual_drift = (growth_factor * prod_factor * 0.08) + policy_boost;
                let step_drift = annual_drift / 500.0;
                let idx_drift = (rng.gen::<f64>() - 0.495) * 0.0018 + step_drift;
                idx.price = (idx.price * (1.0 + idx_drift)).max(10.0);
                idx.change_24h = ((idx.price - idx.base_price) / idx.base_price) * 100.0;
                if idx.price > idx.high_24h {
                    idx.high_24h = idx.price;
                }
                if idx.price < idx.low_24h {
                    idx.low_24h = idx.price;
                }

                // Dynamic 24h trading volume for benchmark sovereign indices
                // Scales with constituents count, market volatility, and liquidity shocks
                let vol_mult = 1.0 + (idx_drift.abs() * 25.0) + (rng.gen::<f64>() - 0.5) * 0.04;
                let base_vol = (idx.constituents_count as f64) * 1_250_000.0;
                idx.volume = (idx.volume * 0.96 + base_vol * 0.04 * vol_mult).clamp(15_000_000.0, 650_000_000.0);
            }

            // Periodically refresh or introduce new loan requests if count drops below 2
            if self.loan_requests.iter().filter(|r| r.status == "pending").count() < 2
                && rng.gen::<f64>() < 0.05
            {
                let foreign_ids: Vec<String> = self.foreign.keys().cloned().collect();
                if !foreign_ids.is_empty() {
                    let idx = rng.gen_range(0..foreign_ids.len());
                    let country_id = &foreign_ids[idx];
                    if let Some(fc) = self.foreign.get(country_id) {
                        let req_id = format!("REQ-{}", rng.gen_range(710..999));
                        let purposes = [
                            "Deepwater Port Channel Dredging Facility",
                            "Emergency Agricultural Grain Reserve Expansion",
                            "Regional Hydro-Electric Substation Overhaul",
                            "High-Speed Fiber-Optic Telecomm Backbone",
                        ];
                        let purpose = purposes[rng.gen_range(0..purposes.len())].to_string();
                        let amount = (rng.gen_range(40..150) as f64) * 1_000_000.0;
                        let rate = 0.055 + rng.gen::<f64>() * 0.035;

                        let term = rng.gen_range(80..200);
                        self.loan_requests.push(ForeignLoanRequest {
                            id: req_id,
                            country_id: fc.id.clone(),
                            country_name: fc.name.clone(),
                            currency: fc.currency.clone(),
                            amount,
                            interest_rate: rate,
                            term_ticks: term,
                            purpose,
                            status: "pending".into(),
                            ticks_remaining: term,
                            interest_earned: 0.0,
                            outcome_message: "Pending sovereign directorate review".into(),
                        });
                    }
                }
            }
        }
    }

    pub fn accept_loan_request(&mut self, request_id: &str) -> Result<(f64, String, String), &'static str> {
        let req = self
            .loan_requests
            .iter_mut()
            .find(|r| r.id == request_id)
            .ok_or("loan request not found")?;

        if req.status != "pending" {
            return Err("loan request is no longer pending");
        }

        req.status = "accepted".to_string();
        req.ticks_remaining = req.term_ticks;
        req.interest_earned = 0.0;
        req.outcome_message = format!("Active: Sovereign coupon yield servicing (~{:.1}% APR)", req.interest_rate * 100.0);
        let amount = req.amount;
        let c_id = req.country_id.clone();
        let c_name = req.country_name.clone();

        if let Some(fc) = self.foreign.get_mut(&c_id) {
            fc.relation.stance = fc.relation.stance.step_closer_to_allied();
            fc.relation.aid_flow += amount;
        }

        let action = ForeignPolicyAction {
            tick: self.tick_counter,
            country_id: c_id.clone(),
            action: "Bilateral Loan Approved".to_string(),
            detail: format!(
                "Approved {} bilateral loan to {} @ {:.2}% interest",
                format_money(amount),
                c_name,
                req.interest_rate * 100.0
            ),
        };
        self.policy_log.push(action);

        Ok((amount, c_id, c_name))
    }

    pub fn reject_loan_request(&mut self, request_id: &str) -> Result<String, &'static str> {
        let req = self
            .loan_requests
            .iter_mut()
            .find(|r| r.id == request_id)
            .ok_or("loan request not found")?;

        if req.status != "pending" {
            return Err("loan request is no longer pending");
        }

        req.status = "declined".to_string();
        req.outcome_message = "Declined by sovereign directorate".to_string();
        let c_name = req.country_name.clone();

        let action = ForeignPolicyAction {
            tick: self.tick_counter,
            country_id: req.country_id.clone(),
            action: "Bilateral Loan Declined".to_string(),
            detail: format!("Declined loan request from {}", c_name),
        };
        self.policy_log.push(action);

        Ok(c_name)
    }

    pub fn respond_bilateral_contract(&mut self, contract_id: &str, action: &str) -> Result<(String, String, f64, f64), &'static str> {
        let con = self
            .bilateral_contracts
            .iter_mut()
            .find(|c| c.id == contract_id)
            .ok_or("contract not found")?;

        if con.status != "pending" {
            return Err("contract proposal is no longer pending");
        }

        if action == "accept" {
            con.status = "active".to_string();
            con.outcome = format!("Ratified: Annual yield +{}, +{:.0}% export boost", format_money(con.annual_revenue_gain), con.export_capacity_boost * 100.0);
            let title = con.title.clone();
            let c_name = con.country_name.clone();
            let rev = con.annual_revenue_gain;
            let boost = con.export_capacity_boost;

            if let Some(fc) = self.foreign.get_mut(&con.country_id) {
                fc.relation.stance = fc.relation.stance.step_closer_to_allied();
                fc.relation.trade_volume *= 1.0 + boost;
            }

            let log_action = ForeignPolicyAction {
                tick: self.tick_counter,
                country_id: con.country_id.clone(),
                action: format!("Accord Ratified: {}", con.contract_type),
                detail: format!("Ratified '{}' with {}. Direct revenue: {}", title, c_name, format_money(rev)),
            };
            self.policy_log.push(log_action);

            Ok((con.country_id.clone(), c_name, rev, boost))
        } else {
            con.status = "declined".to_string();
            con.outcome = "Declined during sovereign diplomatic review".to_string();
            let c_name = con.country_name.clone();

            let log_action = ForeignPolicyAction {
                tick: self.tick_counter,
                country_id: con.country_id.clone(),
                action: format!("Accord Declined: {}", con.contract_type),
                detail: format!("Declined proposed '{}' from {}", con.title, c_name),
            };
            self.policy_log.push(log_action);

            Ok((con.country_id.clone(), c_name, 0.0, 0.0))
        }
    }

    pub fn impose_sanctions(&mut self, country_id: &str, level: i32, tick: i64) -> bool {
        if let Some(fc) = self.foreign.get_mut(country_id) {
            fc.relation.sanctions_level = level;
            if level >= 2 {
                fc.relation.stance = DiplomaticStance::Hostile;
            } else if level >= 1 && fc.relation.stance < DiplomaticStance::Cold {
                fc.relation.stance = DiplomaticStance::Cold;
            }
            let action = ForeignPolicyAction {
                tick,
                country_id: country_id.to_string(),
                action: "Sanctions".to_string(),
                detail: format!("Level {} sanctions imposed on {}", level, fc.name),
            };
            self.policy_log.push(action);
            true
        } else {
            false
        }
    }

    pub fn negotiate_deal(&mut self, country_id: &str, new_tariff: f64, tick: i64) -> bool {
        if let Some(fc) = self.foreign.get_mut(country_id) {
            let old_tariff = fc.relation.tariff_rate;
            fc.relation.tariff_rate = new_tariff;
            if new_tariff < old_tariff && fc.relation.stance < DiplomaticStance::Allied {
                fc.relation.stance = fc.relation.stance.step_closer_to_allied();
            }
            let action = ForeignPolicyAction {
                tick,
                country_id: country_id.to_string(),
                action: "Trade Accord".to_string(),
                detail: format!(
                    "Tariff with {}: {:.1}% → {:.1}%",
                    fc.name,
                    old_tariff * 100.0,
                    new_tariff * 100.0
                ),
            };
            self.policy_log.push(action);
            true
        } else {
            false
        }
    }

    pub fn foreign_countries_sorted(&self) -> Vec<&ForeignCountry> {
        let mut list: Vec<&ForeignCountry> = self.foreign.values().collect();
        list.sort_by(|a, b| a.name.cmp(&b.name));
        list
    }

    pub fn sanction_country(&mut self, country_id: &str, home_power: f64) -> Result<(String, String), &'static str> {
        let fc = self.foreign.get_mut(country_id).ok_or("Foreign country not found")?;
        if fc.sanctioned_by_us {
            return Err("Country is already under sanctions");
        }

        fc.sanctioned_by_us = true;
        fc.relation.stance = DiplomaticStance::Hostile;
        fc.relation.sanctions_level = 3;

        let action_name = format!("Imposed Sanctions on {}", fc.name);
        let detail: String;

        // Check power disparity: If we are a small nation and target a large superpower, they retaliate!
        if home_power < 35.0 && fc.geopolitical_power > 60.0 {
            fc.sanctioned_us = true;
            fc.sanctions_summary = format!("Active: Domestic sanctions enforced. {} retaliated with crushing counter-tariffs!", fc.name);
            detail = format!("Sovereign sanctions enacted against {}. Warning: As a developing nation, {} retaliated with reciprocal import embargoes!", fc.name, fc.name);
        } else if home_power >= 70.0 {
            fc.sanctions_summary = format!("Hegemonic trade and financial embargo enforced on {}.", fc.name);
            detail = format!("Global Powerhouse hegemony exerted: Full financial asset freeze & trade blockade clamped onto {}.", fc.name);
        } else {
            fc.sanctions_summary = format!("Bilateral trade restrictions and capital limits placed on {}.", fc.name);
            detail = format!("Enacted targeted trade sanctions and strategic mineral export bans on {}.", fc.name);
        }

        let action = ForeignPolicyAction {
            tick: self.tick_counter,
            country_id: country_id.to_string(),
            action: action_name.clone(),
            detail: detail.clone(),
        };
        self.policy_log.push(action);

        Ok((action_name, detail))
    }

    pub fn lift_sanction(&mut self, country_id: &str) -> Result<(String, String), &'static str> {
        let fc = self.foreign.get_mut(country_id).ok_or("Foreign country not found")?;
        if !fc.sanctioned_by_us {
            return Err("Country is not currently sanctioned");
        }

        fc.sanctioned_by_us = false;
        fc.sanctioned_us = false;
        fc.relation.stance = DiplomaticStance::Cold;
        fc.relation.sanctions_level = 0;
        fc.sanctions_summary = "No active sanctions.".to_string();

        let action_name = format!("Sanctions Lifted: {}", fc.name);
        let detail = format!("Diplomatic accord achieved: Sanctions and counter-sanctions lifted with {}.", fc.name);

        let action = ForeignPolicyAction {
            tick: self.tick_counter,
            country_id: country_id.to_string(),
            action: action_name.clone(),
            detail: detail.clone(),
        };
        self.policy_log.push(action);

        Ok((action_name, detail))
    }
}
