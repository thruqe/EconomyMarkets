use std::collections::HashMap;
use parking_lot::RwLock;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct FiscalSystem {
    #[serde(skip)]
    #[allow(dead_code)]
    lock: RwLock<()>,
    pub total_revenue: f64,
    pub total_outlays: f64,
    pub budget_deficit: f64,
    pub national_debt: f64,
    pub corporate_tax_rate: f64,
    pub mandatory_spending: f64,
    pub discretionary_spending: f64,
    pub net_interest_expense: f64,
    pub procurement_spending_rate: f64,
    pub active_policies: HashMap<String, bool>,

    pub treasury_cash: f64,
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
    pub reserve_currency_reserves: f64,
    pub personal_income_tax_rate: f64,
    pub sales_tax_rate: f64,
    pub policy_rollouts: HashMap<String, u32>,
    pub foreign_debt_custody_enabled: bool,
    pub foreign_debt_custody_nation: String,
    pub foreign_debt_custody_amount: f64,
}

impl Clone for FiscalSystem {
    fn clone(&self) -> Self {
        Self {
            lock: RwLock::new(()),
            total_revenue: self.total_revenue,
            total_outlays: self.total_outlays,
            budget_deficit: self.budget_deficit,
            national_debt: self.national_debt,
            corporate_tax_rate: self.corporate_tax_rate,
            mandatory_spending: self.mandatory_spending,
            discretionary_spending: self.discretionary_spending,
            net_interest_expense: self.net_interest_expense,
            procurement_spending_rate: self.procurement_spending_rate,
            active_policies: self.active_policies.clone(),
            treasury_cash: self.treasury_cash,
            credit_rating: self.credit_rating.clone(),
            borrowing_yield: self.borrowing_yield,
            infrastructure_level: self.infrastructure_level,
            healthcare_level: self.healthcare_level,
            education_level: self.education_level,
            enterprise_grants_level: self.enterprise_grants_level,
            export_capacity_level: self.export_capacity_level,
            fdi_inflow: self.fdi_inflow,
            export_revenue: self.export_revenue,
            exchange_chartered: self.exchange_chartered,
            maintenance_spending_rate: self.maintenance_spending_rate,
            depreciation_rate: self.depreciation_rate,
            tourism_revenue: self.tourism_revenue,
            productivity_index: self.productivity_index,
            reserve_currency_reserves: self.reserve_currency_reserves,
            personal_income_tax_rate: self.personal_income_tax_rate,
            sales_tax_rate: self.sales_tax_rate,
            policy_rollouts: self.policy_rollouts.clone(),
            foreign_debt_custody_enabled: self.foreign_debt_custody_enabled,
            foreign_debt_custody_nation: self.foreign_debt_custody_nation.clone(),
            foreign_debt_custody_amount: self.foreign_debt_custody_amount,
        }
    }
}

// Category 1: Fiscal & Infrastructure
pub const POLICY_RURAL_ELECTRIFICATION: &str = "rural_electrification";
pub const POLICY_PORT_MODERNIZATION: &str = "port_modernization";
pub const POLICY_TARIFF_SHIELD: &str = "tariff_shield";
pub const POLICY_HIGH_SPEED_RAIL: &str = "high_speed_rail";
pub const POLICY_SOVEREIGN_WEALTH_FUND: &str = "sovereign_wealth_fund";
pub const POLICY_DEEP_SPACE_PORT: &str = "deep_space_port";

// Category 2: Monetary & Banking
pub const POLICY_MICROFINANCE_ACT: &str = "microfinance_act";
pub const POLICY_RATE_CUT_50BP: &str = "rate_cut_50bp";
pub const POLICY_RATE_HIKE_50BP: &str = "rate_hike_50bp";
pub const POLICY_DISCOUNT_WINDOW_REFORM: &str = "discount_window_reform";
pub const POLICY_QUANTITATIVE_EASING: &str = "quantitative_easing";
pub const POLICY_RESERVE_CURRENCY_OPS: &str = "reserve_currency_ops";

// Category 3: Trade & Tariffs
pub const POLICY_AGRI_EXPORT_COMPACT: &str = "agri_export_compact";
pub const POLICY_MINERAL_ROYALTY: &str = "mineral_royalty";
pub const POLICY_REGIONAL_FREE_TRADE: &str = "regional_free_trade";
pub const POLICY_ENERGY_EXPORT_TERMINAL: &str = "energy_export_terminal";
pub const POLICY_GLOBAL_TRADE_HEGEMONY: &str = "global_trade_hegemony";

// Category 4: Industry & Technology
pub const POLICY_SMALL_BUSINESS_INCUBATOR: &str = "small_biz_incubator";
pub const POLICY_LIGHT_MANUFACTURING_GRANTS: &str = "light_mfg_grants";
pub const POLICY_SEMICONDUCTOR_FAB_INITIATIVE: &str = "semiconductor_fab";
pub const POLICY_CLEAN_ENERGY_GRID: &str = "clean_energy_grid";
pub const POLICY_AI_QUANTUM_MEGACENTER: &str = "ai_quantum_center";

// Category 5: Labor & Welfare
pub const POLICY_PRIMARY_HEALTH_CLINICS: &str = "primary_health_clinics";
pub const POLICY_VOCATIONAL_APPRENTICESHIPS: &str = "vocational_apprenticeships";
pub const POLICY_LIVING_WAGE_GUARANTEE: &str = "living_wage_guarantee";
pub const POLICY_HIGH_SKILL_VISA: &str = "high_skill_visa";
pub const POLICY_CITIZEN_STIMULUS: &str = "citizen_stimulus";

// Category 6: Geopolitics & Foreign Relations
pub const POLICY_GOOD_NEIGHBOR_MISSION: &str = "good_neighbor_mission";
pub const POLICY_REGIONAL_SECURITY_PACT: &str = "regional_security_pact";
pub const POLICY_SUPERPOWER_ALLIANCE: &str = "superpower_alliance";

// Citizen Welfare & Labor Policies
pub const POLICY_UNIVERSAL_CHILDCARE: &str = "universal_childcare";
pub const POLICY_SALES_VAT: &str = "sales_vat";
pub const POLICY_LABOR_DEREGULATION: &str = "labor_deregulation";
pub const POLICY_APPRENTICE_SUBSIDY: &str = "apprentice_subsidy";

// High-Impact Strategic Directives (Unmarked Dangerous Policies)
pub const POLICY_UNLIMITED_LIQUIDITY: &str = "unlimited_liquidity";
pub const POLICY_PRICE_FREEZE: &str = "price_freeze";
pub const POLICY_WINDFALL_SEIZURE: &str = "windfall_seizure";
pub const POLICY_DEBT_RELIEF: &str = "debt_relief";
pub const POLICY_AUTARKY_TARIFF: &str = "autarky_tariff";

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PolicyInfo {
    pub id: String,
    pub category: String,
    pub name: String,
    pub description: String,
    pub favored_sector: String,
    pub annual_cost: f64,
    pub min_tier: String, // "small", "mid", "powerhouse"
    pub rollout_days: u32,
    pub direct_positive: String,
    pub direct_negative: String,
    pub indirect_positive: String,
    pub indirect_negative: String,
}

pub fn all_known_policies() -> Vec<PolicyInfo> {
    vec![
        // === 1. Fiscal & Infrastructure ===
        PolicyInfo {
            id: POLICY_RURAL_ELECTRIFICATION.to_string(),
            category: "fiscal".to_string(),
            name: "Rural Grid & Electrification".to_string(),
            description: "+15% demand to Utilities via foundational provincial grid connection".to_string(),
            favored_sector: "Utilities".to_string(),
            annual_cost: 250_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 90,
            direct_positive: "+15% demand to Utilities via foundational provincial grid connection".to_string(),
            direct_negative: "-$250M annual budget outlay; requires treasury cash or bond issuance".to_string(),
            indirect_positive: "+0.6% boost to rural agricultural & industrial labor productivity".to_string(),
            indirect_negative: "Temporary local material shortages during grid installation".to_string(),
        },
        PolicyInfo {
            id: POLICY_PORT_MODERNIZATION.to_string(),
            category: "fiscal".to_string(),
            name: "National Port Modernization".to_string(),
            description: "+20% demand to Industrials via automated container shipping docks".to_string(),
            favored_sector: "Industrials".to_string(),
            annual_cost: 400_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 120,
            direct_positive: "+20% shipping container handling capacity & export throughput".to_string(),
            direct_negative: "-$400M annual fiscal construction cost expanding sovereign debt".to_string(),
            indirect_positive: "+$850M foreign merchandise export trade volume".to_string(),
            indirect_negative: "Harbor congestion and logistics reallocation during harbor dredging".to_string(),
        },
        PolicyInfo {
            id: POLICY_TARIFF_SHIELD.to_string(),
            category: "fiscal".to_string(),
            name: "Strategic Infant-Industry Tariff".to_string(),
            description: "Protects domestic producers with a 6.5% tariff, generating treasury revenue".to_string(),
            favored_sector: "Consumer Discretionary".to_string(),
            annual_cost: -300_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 45,
            direct_positive: "+$300M in immediate treasury customs tariff revenues".to_string(),
            direct_negative: "Retaliatory trade barriers threatened by foreign partner states".to_string(),
            indirect_positive: "+12% profit margins for domestic infant manufacturing firms".to_string(),
            indirect_negative: "+0.7% upward pressure on domestic consumer shelf prices".to_string(),
        },
        PolicyInfo {
            id: POLICY_HIGH_SPEED_RAIL.to_string(),
            category: "fiscal".to_string(),
            name: "National High-Speed Transit Corridor".to_string(),
            description: "+30% demand to Industrials & +20% Materials for national connectivity".to_string(),
            favored_sector: "Industrials".to_string(),
            annual_cost: 1_800_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 180,
            direct_positive: "+30% demand to Industrials & +20% Materials for transit corridor".to_string(),
            direct_negative: "-$1.8B annual fiscal outlay substantially draining treasury liquidity".to_string(),
            indirect_positive: "+1.2% national GDP growth through seamless regional commerce".to_string(),
            indirect_negative: "Significant debt-service burden if tax revenues decline".to_string(),
        },
        PolicyInfo {
            id: POLICY_SOVEREIGN_WEALTH_FUND.to_string(),
            category: "fiscal".to_string(),
            name: "Sovereign Wealth Capital Fund".to_string(),
            description: "Deploys treasury surplus directly into equity stakes of domestic innovators".to_string(),
            favored_sector: "Financials".to_string(),
            annual_cost: 3_000_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 90,
            direct_positive: "+15% long-term investment dividend returns for public coffers".to_string(),
            direct_negative: "-$3.0B initial capital allocation directly reducing available cash".to_string(),
            indirect_positive: "+25% venture equity funding for domestic technology pioneers".to_string(),
            indirect_negative: "Market volatility risk if global equity valuations decline".to_string(),
        },
        PolicyInfo {
            id: POLICY_DEEP_SPACE_PORT.to_string(),
            category: "fiscal".to_string(),
            name: "Deep Space & Fusion Mega-Project".to_string(),
            description: "+40% demand to Industrials & +30% Tech for planetary aerospace dominance".to_string(),
            favored_sector: "Industrials".to_string(),
            annual_cost: 15_000_000_000.0,
            min_tier: "powerhouse".to_string(),
            rollout_days: 365,
            direct_positive: "+40% demand to Industrials & +30% Tech for aerospace dominance".to_string(),
            direct_negative: "-$15.0B colossal multi-year fiscal expenditure forcing debt issuance".to_string(),
            indirect_positive: "+$3.2B international satellite launch contract revenues".to_string(),
            indirect_negative: "Heavily widens national budget deficit during development phase".to_string(),
        },

        // === 2. Monetary & Banking ===
        PolicyInfo {
            id: POLICY_MICROFINANCE_ACT.to_string(),
            category: "monetary".to_string(),
            name: "Microfinance & Cooperative Banking".to_string(),
            description: "Empowers community credit unions to provide micro-loans to small enterprises".to_string(),
            favored_sector: "Financials".to_string(),
            annual_cost: 100_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 60,
            direct_positive: "+18% credit availability for provincial sole-proprietor ventures".to_string(),
            direct_negative: "-$100M annual central banking seed subsidy".to_string(),
            indirect_positive: "+35,000 net new private-sector micro-enterprise jobs".to_string(),
            indirect_negative: "Marginal uptick in non-performing retail micro-credit defaults".to_string(),
        },
        PolicyInfo {
            id: POLICY_RATE_CUT_50BP.to_string(),
            category: "monetary".to_string(),
            name: "Emergency Rate Cut (-0.50%)".to_string(),
            description: "Lowers benchmark interest rate to stimulate commercial borrowing & equity markets".to_string(),
            favored_sector: "Real Estate".to_string(),
            annual_cost: 0.0,
            min_tier: "small".to_string(),
            rollout_days: 30,
            direct_positive: "-0.50% benchmark borrowing costs stimulating commercial loans".to_string(),
            direct_negative: "Lower yield returns for fixed-income sovereign bondholders".to_string(),
            indirect_positive: "+4.5% equity valuation appreciation and commercial real estate boost".to_string(),
            indirect_negative: "Downward exchange rate depreciation pressure on national currency".to_string(),
        },
        PolicyInfo {
            id: POLICY_RATE_HIKE_50BP.to_string(),
            category: "monetary".to_string(),
            name: "Counter-Inflation Rate Hike (+0.50%)".to_string(),
            description: "Raises borrowing costs to defend sovereign currency strength & tame inflation".to_string(),
            favored_sector: "Financials".to_string(),
            annual_cost: 0.0,
            min_tier: "small".to_string(),
            rollout_days: 30,
            direct_positive: "+0.50% benchmark return strengthening domestic currency exchange rate".to_string(),
            direct_negative: "Increased corporate borrowing expenses and slower debt-financed hiring".to_string(),
            indirect_positive: "Cooling headline CPI inflation and suppressing import costs".to_string(),
            indirect_negative: "Temporary drag on equity indices and capital expenditure".to_string(),
        },
        PolicyInfo {
            id: POLICY_DISCOUNT_WINDOW_REFORM.to_string(),
            category: "monetary".to_string(),
            name: "Central Bank Discount Window Facility".to_string(),
            description: "Guarantees lender-of-last-resort liquidity backstops for commercial banks".to_string(),
            favored_sector: "Financials".to_string(),
            annual_cost: 500_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 60,
            direct_positive: "Eliminates liquidity insolvency risk for national banking institutions".to_string(),
            direct_negative: "-$500M annual standing credit liquidity commitment".to_string(),
            indirect_positive: "Deepens interbank lending confidence and corporate line availability".to_string(),
            indirect_negative: "Potential moral hazard encouraging aggressive commercial lending".to_string(),
        },
        PolicyInfo {
            id: POLICY_QUANTITATIVE_EASING.to_string(),
            category: "monetary".to_string(),
            name: "Large-Scale Asset Purchases (QE)".to_string(),
            description: "Direct secondary market bond purchases to lower long-term borrowing yields".to_string(),
            favored_sector: "All Sectors".to_string(),
            annual_cost: 4_000_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 90,
            direct_positive: "Drives down 10-year sovereign bond yields to historic lows".to_string(),
            direct_negative: "-$4.0B balance-sheet asset purchase commitment".to_string(),
            indirect_positive: "+15% asset price inflation and abundant corporate refinancing liquidity".to_string(),
            indirect_negative: "Risk of long-run currency debasement and monetary overextension".to_string(),
        },
        PolicyInfo {
            id: POLICY_RESERVE_CURRENCY_OPS.to_string(),
            category: "monetary".to_string(),
            name: "Global Reserve Currency Hegemony".to_string(),
            description: "Establishes sovereign currency as bilateral settlement unit worldwide".to_string(),
            favored_sector: "Financials".to_string(),
            annual_cost: 0.0,
            min_tier: "powerhouse".to_string(),
            rollout_days: 180,
            direct_positive: "Global reserve currency status allowing debt issuance at minimal yields".to_string(),
            direct_negative: "Subordination of domestic monetary independence to global dollar pool".to_string(),
            indirect_positive: "+$25B bilateral foreign reserve custody inflows backing currency".to_string(),
            indirect_negative: "Persistent trade deficit bias typical of global reserve anchors".to_string(),
        },

        // === 3. Trade & Tariffs ===
        PolicyInfo {
            id: POLICY_AGRI_EXPORT_COMPACT.to_string(),
            category: "trade".to_string(),
            name: "Bilateral Agricultural Export Compact".to_string(),
            description: "Eliminates quarantine hurdles and tariffs with neighbors for domestic food exports".to_string(),
            favored_sector: "Consumer Staples".to_string(),
            annual_cost: -150_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 45,
            direct_positive: "+$150M annual tariff exemption gains for domestic agricultural producers".to_string(),
            direct_negative: "Quarantine protocol enforcement administrative overhead".to_string(),
            indirect_positive: "+8.5% rural household income growth across farming communities".to_string(),
            indirect_negative: "Domestic market food price fluctuations linked to global export bids".to_string(),
        },
        PolicyInfo {
            id: POLICY_MINERAL_ROYALTY.to_string(),
            category: "trade".to_string(),
            name: "Strategic Mineral Extraction Royalty".to_string(),
            description: "Levies foreign extraction royalty on domestic lithium, copper, and iron deposits".to_string(),
            favored_sector: "Materials".to_string(),
            annual_cost: -350_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 60,
            direct_positive: "+$350M annual direct royalty revenues deposited into treasury cash".to_string(),
            direct_negative: "Disgruntled foreign multinational mining concessionaires".to_string(),
            indirect_positive: "Preserves sovereign natural resource wealth for public benefit".to_string(),
            indirect_negative: "Slight moderation in foreign speculative mineral exploration capex".to_string(),
        },
        PolicyInfo {
            id: POLICY_REGIONAL_FREE_TRADE.to_string(),
            category: "trade".to_string(),
            name: "Regional Free Trade Accord".to_string(),
            description: "Lowers barriers with regional trading partners to surge cross-border commerce".to_string(),
            favored_sector: "Industrials".to_string(),
            annual_cost: -800_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 90,
            direct_positive: "+$800M cross-border manufacturing commerce and tariff elimination".to_string(),
            direct_negative: "Uncompetitive domestic legacy producers exposed to foreign rivals".to_string(),
            indirect_positive: "+1.1% national industrial productivity and component cost reductions".to_string(),
            indirect_negative: "Short-term customs tariff collection loss offset by trade volume".to_string(),
        },
        PolicyInfo {
            id: POLICY_ENERGY_EXPORT_TERMINAL.to_string(),
            category: "trade".to_string(),
            name: "Deepwater LNG & Energy Export Terminal".to_string(),
            description: "State-of-the-art export facility for domestic clean gas and energy supplies".to_string(),
            favored_sector: "Energy".to_string(),
            annual_cost: -1_500_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 150,
            direct_positive: "+$1.5B annual foreign exchange revenues from natural gas and energy exports".to_string(),
            direct_negative: "-$800M initial terminal pipeline construction co-funding".to_string(),
            indirect_positive: "+45% sovereign foreign currency reserves accumulation".to_string(),
            indirect_negative: "Domestic industrial utility tariffs exposed to world market energy rates".to_string(),
        },
        PolicyInfo {
            id: POLICY_GLOBAL_TRADE_HEGEMONY.to_string(),
            category: "trade".to_string(),
            name: "Global Multilateral Trade Nexus".to_string(),
            description: "Positions nation as the primary clearing hub for global container shipping".to_string(),
            favored_sector: "Industrials".to_string(),
            annual_cost: -5_000_000_000.0,
            min_tier: "powerhouse".to_string(),
            rollout_days: 240,
            direct_positive: "+$5.0B annual maritime clearing fees and transshipment tariffs".to_string(),
            direct_negative: "Heavy ongoing naval and maritime patrol maintenance expenses".to_string(),
            indirect_positive: "Sovereign control over critical international trade shipping lanes".to_string(),
            indirect_negative: "Geopolitical scrutiny and rival maritime coalition tensions".to_string(),
        },

        // === 4. Industry & Technology ===
        PolicyInfo {
            id: POLICY_SMALL_BUSINESS_INCUBATOR.to_string(),
            category: "industry".to_string(),
            name: "Early-Stage Commercial Seed Grants".to_string(),
            description: "Direct non-dilutive matching grants for new provincial business formation".to_string(),
            favored_sector: "Consumer Discretionary".to_string(),
            annual_cost: 180_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 60,
            direct_positive: "+$180M matching seed grants distributed to 500+ commercial startups".to_string(),
            direct_negative: "-$180M annual fiscal cost deducted from treasury cash".to_string(),
            indirect_positive: "+22,000 jobs created in retail, hospitality, and local craft trades".to_string(),
            indirect_negative: "Standard early-stage business failure rates (~15% default rate)".to_string(),
        },
        PolicyInfo {
            id: POLICY_LIGHT_MANUFACTURING_GRANTS.to_string(),
            category: "industry".to_string(),
            name: "Domestic Light Manufacturing Subsidies".to_string(),
            description: "+20% demand to Industrials via factory machinery subsidies & domestic supply lines".to_string(),
            favored_sector: "Industrials".to_string(),
            annual_cost: 300_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 90,
            direct_positive: "+20% output growth for domestic component and textile factories".to_string(),
            direct_negative: "-$300M machinery subsidy program cost expanding budget deficit".to_string(),
            indirect_positive: "Reduces dependency on foreign manufacturing import supply chains".to_string(),
            indirect_negative: "Requires ongoing technical workforce skill retooling".to_string(),
        },
        PolicyInfo {
            id: POLICY_SEMICONDUCTOR_FAB_INITIATIVE.to_string(),
            category: "industry".to_string(),
            name: "Semiconductor Cleanroom Fab Initiative".to_string(),
            description: "+35% demand to Tech via state-subsidized advanced microchip fabrication plants".to_string(),
            favored_sector: "Information Technology".to_string(),
            annual_cost: 2_500_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 270,
            direct_positive: "+35% technology sector valuation and domestic silicon sovereignty".to_string(),
            direct_negative: "-$2.5B heavy capital outlay requiring sovereign bond financing".to_string(),
            indirect_positive: "+$1.8B high-value semiconductor exports to allied world nations".to_string(),
            indirect_negative: "Vulnerability to global silicon cycle downturns and supply gluts".to_string(),
        },
        PolicyInfo {
            id: POLICY_CLEAN_ENERGY_GRID.to_string(),
            category: "industry".to_string(),
            name: "Modular Nuclear & Solar Grid Buildout".to_string(),
            description: "+25% Utilities & +20% Energy via next-gen baseload clean generation facilities".to_string(),
            favored_sector: "Utilities".to_string(),
            annual_cost: 1_200_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 180,
            direct_positive: "+25% baseload grid generation capacity from modular nuclear & solar".to_string(),
            direct_negative: "-$1.2B annual infrastructure buildout expense adding to national debt".to_string(),
            indirect_positive: "Shields domestic industry from global fossil fuel price volatility".to_string(),
            indirect_negative: "Decade-long capital amortization cycle before full operational parity".to_string(),
        },
        PolicyInfo {
            id: POLICY_AI_QUANTUM_MEGACENTER.to_string(),
            category: "industry".to_string(),
            name: "Sovereign AI & Quantum Compute Center".to_string(),
            description: "+40% demand to Tech via exascale sovereign supercomputing installations".to_string(),
            favored_sector: "Information Technology".to_string(),
            annual_cost: 8_000_000_000.0,
            min_tier: "powerhouse".to_string(),
            rollout_days: 300,
            direct_positive: "+40% compute capacity for national defense, science, and enterprises".to_string(),
            direct_negative: "-$8.0B multi-gigawatt sovereign compute procurement expenditure".to_string(),
            indirect_positive: "Attracts premier multinational AI enterprise research labs".to_string(),
            indirect_negative: "Massive electrical grid draw requiring prioritized power allocation".to_string(),
        },

        // === 5. Labor & Welfare ===
        PolicyInfo {
            id: POLICY_PRIMARY_HEALTH_CLINICS.to_string(),
            category: "labor".to_string(),
            name: "Provincial Primary Health Clinics".to_string(),
            description: "Free outpatient clinics across all districts, increasing labor participation by 1.8%".to_string(),
            favored_sector: "Health Care".to_string(),
            annual_cost: 200_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 75,
            direct_positive: "Extends free district outpatient care, lifting labor participation +1.8%".to_string(),
            direct_negative: "-$200M recurring annual public health operating outlays".to_string(),
            indirect_positive: "-12% lost workforce hours from preventable illness and absenteeism".to_string(),
            indirect_negative: "Staffing competition for medical personnel across urban hospitals".to_string(),
        },
        PolicyInfo {
            id: POLICY_VOCATIONAL_APPRENTICESHIPS.to_string(),
            category: "labor".to_string(),
            name: "National Technical Apprenticeship Act".to_string(),
            description: "Paid industrial apprenticeships in metallurgy, construction, and electronics".to_string(),
            favored_sector: "Industrials".to_string(),
            annual_cost: 150_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 90,
            direct_positive: "+15% influx of certified machinists, electricians, and technicians".to_string(),
            direct_negative: "-$150M annual technical college training subsidies".to_string(),
            indirect_positive: "Drives down youth unemployment and improves industrial safety standards".to_string(),
            indirect_negative: "Slight 1-year training lead time before trainees reach peak productivity".to_string(),
        },
        PolicyInfo {
            id: POLICY_LIVING_WAGE_GUARANTEE.to_string(),
            category: "labor".to_string(),
            name: "National Living Wage Benchmark".to_string(),
            description: "Establishes guaranteed statutory living wage, boosting retail consumer spending".to_string(),
            favored_sector: "Consumer Staples".to_string(),
            annual_cost: 1_000_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 45,
            direct_positive: "+$1.0B disposable income surge for low-wage households boosting retail sales".to_string(),
            direct_negative: "Increases statutory labor expense for labor-intensive hospitality firms".to_string(),
            indirect_positive: "+8% consumer confidence and rapid poverty reduction".to_string(),
            indirect_negative: "Marginal incentive for marginal businesses to automate entry-level roles".to_string(),
        },
        PolicyInfo {
            id: POLICY_HIGH_SKILL_VISA.to_string(),
            category: "labor".to_string(),
            name: "Elite STEM & Research Fast-Track Visa".to_string(),
            description: "Attracts thousands of premier foreign engineers, researchers, and biochemists".to_string(),
            favored_sector: "Information Technology".to_string(),
            annual_cost: 80_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 45,
            direct_positive: "+12,000 elite foreign PhD researchers, engineers, and bioscientists".to_string(),
            direct_negative: "-$80M expedited immigration administration and consular logistics".to_string(),
            indirect_positive: "+$600M corporate R&D investment and patent generation acceleration".to_string(),
            indirect_negative: "Slightly tighter premium housing availability in metropolitan tech hubs".to_string(),
        },
        PolicyInfo {
            id: POLICY_CITIZEN_STIMULUS.to_string(),
            category: "labor".to_string(),
            name: "Universal Sovereign Citizen Dividend".to_string(),
            description: "Direct annual cash distributions to all sovereign citizens from treasury surplus".to_string(),
            favored_sector: "All Sectors".to_string(),
            annual_cost: 12_000_000_000.0,
            min_tier: "powerhouse".to_string(),
            rollout_days: 30,
            direct_positive: "+$12.0B direct consumer purchasing power injected into all households".to_string(),
            direct_negative: "-$12.0B immediate massive cash withdrawal requiring debt issuance".to_string(),
            indirect_positive: "+25% immediate retail transaction surge across all domestic sectors".to_string(),
            indirect_negative: "+1.8% CPI inflation spike and credit rating pressure from debt expansion".to_string(),
        },

        // === 6. Geopolitics & Foreign Relations ===
        PolicyInfo {
            id: POLICY_GOOD_NEIGHBOR_MISSION.to_string(),
            category: "foreign".to_string(),
            name: "Good Neighbor Diplomatic Envoy".to_string(),
            description: "Diplomatic consulates to foster trade corridors and defuse regional tensions".to_string(),
            favored_sector: "Consumer Staples".to_string(),
            annual_cost: 50_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 30,
            direct_positive: "Improves diplomatic relations to Friendly/Allied with regional neighbors".to_string(),
            direct_negative: "-$50M diplomatic delegation and embassy maintenance outlays".to_string(),
            indirect_positive: "+$240M cross-border bilateral trade volume agreements".to_string(),
            indirect_negative: "Diplomatic exposure to neighboring internal geopolitical disputes".to_string(),
        },
        PolicyInfo {
            id: POLICY_REGIONAL_SECURITY_PACT.to_string(),
            category: "foreign".to_string(),
            name: "Mutual Regional Defense Treaty".to_string(),
            description: "Defense pact stabilizing sovereign credit rating and regional trade routes".to_string(),
            favored_sector: "Industrials".to_string(),
            annual_cost: 900_000_000.0,
            min_tier: "mid".to_string(),
            rollout_days: 120,
            direct_positive: "Guaranteed mutual defense backstop stabilizing sovereign credit ratings".to_string(),
            direct_negative: "-$900M defense modernization and joint military exercise commitment".to_string(),
            indirect_positive: "+$450M defense manufacturing equipment export opportunities".to_string(),
            indirect_negative: "Treaty obligation to assist allies during external security crises".to_string(),
        },
        PolicyInfo {
            id: POLICY_SUPERPOWER_ALLIANCE.to_string(),
            category: "foreign".to_string(),
            name: "Global Superpower Strategic Alliance".to_string(),
            description: "Bilateral strategic pact with world superpowers guaranteeing trade access".to_string(),
            favored_sector: "All Sectors".to_string(),
            annual_cost: 4_000_000_000.0,
            min_tier: "powerhouse".to_string(),
            rollout_days: 180,
            direct_positive: "Formal security guarantee & preferred trade access with world superpowers".to_string(),
            direct_negative: "-$4.0B alliance strategic contribution and joint facility hosting cost".to_string(),
            indirect_positive: "+18% foreign direct investment (FDI) inflows from superpower multinationals".to_string(),
            indirect_negative: "Rival geopolitical blocs may cool trade ties and impose targeted limits".to_string(),
        },

        // === 7. Citizen Welfare & Labor Dynamics ===
        PolicyInfo {
            id: POLICY_UNIVERSAL_CHILDCARE.to_string(),
            category: "labor".to_string(),
            name: "Universal Early Childcare & Family Subsidies".to_string(),
            description: "Provides free municipal childcare centers, expanding labor force participation by +2.2%".to_string(),
            favored_sector: "Health Care".to_string(),
            annual_cost: 220_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 90,
            direct_positive: "+2.2% increase in female and caregiver workforce participation".to_string(),
            direct_negative: "-$220M annual municipal childcare center operational expenditures".to_string(),
            indirect_positive: "+$620M annual gross taxable wage earnings from re-entering workers".to_string(),
            indirect_negative: "Initial waitlists during municipal facility construction phase".to_string(),
        },
        PolicyInfo {
            id: POLICY_SALES_VAT.to_string(),
            category: "fiscal".to_string(),
            name: "Broad-Based Federal Sales Consumption VAT (7.5%)".to_string(),
            description: "Levies a 7.5% transaction levy on consumer goods, generating steady non-debt treasury revenue".to_string(),
            favored_sector: "Consumer Staples".to_string(),
            annual_cost: -750_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 30,
            direct_positive: "+$750M steady non-debt treasury revenue collected from consumption".to_string(),
            direct_negative: "Mild -1.5% dampening effect on discretionary retail consumer spending".to_string(),
            indirect_positive: "Significantly narrows federal budget deficit without raising wage taxes".to_string(),
            indirect_negative: "Disproportionately felt by low-income households without exemptions".to_string(),
        },
        PolicyInfo {
            id: POLICY_LABOR_DEREGULATION.to_string(),
            category: "labor".to_string(),
            name: "Commercial Labor Flexibility & At-Will Reform".to_string(),
            description: "Relaxes statutory severance mandates to boost corporate operating margins & hiring speed".to_string(),
            favored_sector: "Industrials".to_string(),
            annual_cost: 0.0,
            min_tier: "small".to_string(),
            rollout_days: 30,
            direct_positive: "+4.2% expansion in corporate operating profit margins and hiring velocity".to_string(),
            direct_negative: "Erosion of statutory severance safety net for discharged workers".to_string(),
            indirect_positive: "+18,000 private-sector job openings as hiring hesitation vanishes".to_string(),
            indirect_negative: "-4% initial drop in labor union consumer sentiment index".to_string(),
        },
        PolicyInfo {
            id: POLICY_APPRENTICE_SUBSIDY.to_string(),
            category: "industry".to_string(),
            name: "Advanced Manufacturing Apprenticeship Matching".to_string(),
            description: "Direct co-funding for private sector industrial trainees, lifting workforce productivity".to_string(),
            favored_sector: "Industrials".to_string(),
            annual_cost: 140_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 60,
            direct_positive: "+15% corporate industrial apprentice matching in high-tech manufacturing".to_string(),
            direct_negative: "-$140M state co-funding grant outlays from treasury cash".to_string(),
            indirect_positive: "Accelerates modern robotics and precision CNC tooling adoption".to_string(),
            indirect_negative: "Requires rigorous audit oversight to prevent corporate grant misallocation".to_string(),
        },

        // === 8. Strategic Directives & Crisis Levers ===
        PolicyInfo {
            id: POLICY_UNLIMITED_LIQUIDITY.to_string(),
            category: "monetary".to_string(),
            name: "Universal Sovereign Liquidity Facility".to_string(),
            description: "Guarantees zero-cost standing credit lines to all commercial institutions and domestic households".to_string(),
            favored_sector: "Financials".to_string(),
            annual_cost: 6_000_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 15,
            direct_positive: "Instantly prevents any domestic banking insolvency or margin default".to_string(),
            direct_negative: "-$6.0B sovereign cash drain; degrades credit rating directly to Junk (CCC)".to_string(),
            indirect_positive: "Suppresses all banking failure contagion across the commercial sector".to_string(),
            indirect_negative: "+4.5% hyper-inflation hazard and massive sovereign borrowing yield spike".to_string(),
        },
        PolicyInfo {
            id: POLICY_PRICE_FREEZE.to_string(),
            category: "trade".to_string(),
            name: "Emergency Essential Commodity Price Cap Act".to_string(),
            description: "Establishes statutory price ceilings on consumer groceries, fuel, and utilities to protect household purchasing power".to_string(),
            favored_sector: "Consumer Staples".to_string(),
            annual_cost: 0.0,
            min_tier: "small".to_string(),
            rollout_days: 15,
            direct_positive: "Haults consumer price index inflation on staple groceries and energy".to_string(),
            direct_negative: "Cuts corporate revenues by -45% and causes severe industrial shortages".to_string(),
            indirect_positive: "Short-term relief for household budgets facing price spikes".to_string(),
            indirect_negative: "Productivity index collapses (-8 pts) as black markets and rationing emerge".to_string(),
        },
        PolicyInfo {
            id: POLICY_WINDFALL_SEIZURE.to_string(),
            category: "fiscal".to_string(),
            name: "100% Windfall Sovereign Asset Recovery Levy".to_string(),
            description: "Enacts a mandatory one-time capital levy on foreign deposits and multinational corporate balances to expand reserves".to_string(),
            favored_sector: "Financials".to_string(),
            annual_cost: -10_000_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 15,
            direct_positive: "+$10.0B massive emergency cash injection into treasury cash reserves".to_string(),
            direct_negative: "FDI collapses to 0; international investors completely flee sovereign market".to_string(),
            indirect_positive: "Provides immense immediate liquidity to avoid sovereign debt default".to_string(),
            indirect_negative: "Severe diplomatic backlash, trade embargoes, and capital flight".to_string(),
        },
        PolicyInfo {
            id: POLICY_DEBT_RELIEF.to_string(),
            category: "fiscal".to_string(),
            name: "Emergency Sovereign Debt Relief Decree".to_string(),
            description: "Unilaterally restructures and clears 50% of outstanding national debt obligations to unburden future generations".to_string(),
            favored_sector: "Real Estate".to_string(),
            annual_cost: 0.0,
            min_tier: "small".to_string(),
            rollout_days: 15,
            direct_positive: "Clears 50% of outstanding national debt obligations in one decree".to_string(),
            direct_negative: "Credit rating collapses to Default (D); borrowing yield spikes to 25.0%".to_string(),
            indirect_positive: "Dramatically reduces ongoing annual net debt interest expense".to_string(),
            indirect_negative: "Sovereign bond market completely freezes; international capital locked out".to_string(),
        },
        PolicyInfo {
            id: POLICY_AUTARKY_TARIFF.to_string(),
            category: "trade".to_string(),
            name: "Comprehensive Strategic Autarky Protective Wall".to_string(),
            description: "Imposes an across-the-board 50% tariff on all incoming foreign trade to compel domestic industrial self-reliance".to_string(),
            favored_sector: "Materials".to_string(),
            annual_cost: -4_500_000_000.0,
            min_tier: "small".to_string(),
            rollout_days: 30,
            direct_positive: "Imposes 50% tariff generating immediate customs revenues for treasury".to_string(),
            direct_negative: "Export revenues collapse by -85% as world retaliates with total bans".to_string(),
            indirect_positive: "Guarantees 100% domestic market capture for surviving local producers".to_string(),
            indirect_negative: "Severe domestic shortages of critical foreign tech, medicines, and parts".to_string(),
        },
    ]
}

impl Default for FiscalSystem {
    fn default() -> Self {
        // Starting metrics for a developing sovereign nation ($45B GDP):
        let debt = 24_000_000_000.0;    // $24B National Debt (~53% Debt-to-GDP)
        let mand = 6_500_000_000.0;     // $6.5B Mandatory spending
        let disc = 3_200_000_000.0;     // $3.2B Discretionary spending
        let yield_val = 0.042;          // 4.2% Borrowing yield
        let interest = debt * yield_val;
        let outlays = mand + disc + interest;
        let rev = 9_500_000_000.0;      // $9.5B Annual revenue

        Self {
            lock: RwLock::new(()),
            total_revenue: rev,
            total_outlays: outlays,
            budget_deficit: outlays - rev,
            national_debt: debt,
            corporate_tax_rate: 0.21,
            mandatory_spending: mand,
            discretionary_spending: disc,
            net_interest_expense: interest,
            procurement_spending_rate: 600_000_000.0,
            active_policies: HashMap::new(),
            treasury_cash: 4_500_000_000.0, // $4.5B starting treasury cash
            credit_rating: "A+".to_string(),
            borrowing_yield: yield_val,
            infrastructure_level: 45.0,
            healthcare_level: 42.0,
            education_level: 44.0,
            enterprise_grants_level: 30.0,
            export_capacity_level: 35.0,
            fdi_inflow: 800_000_000.0,
            export_revenue: 4_200_000_000.0,
            exchange_chartered: true,
            maintenance_spending_rate: 450_000_000.0,
            depreciation_rate: 0.025,
            tourism_revenue: 350_000_000.0,
            productivity_index: 100.0,
            reserve_currency_reserves: 650_000_000.0, // 650M ZTH starting reserve currency holdings
            personal_income_tax_rate: 0.185,
            sales_tax_rate: 0.085,
            policy_rollouts: HashMap::new(),
            foreign_debt_custody_enabled: false,
            foreign_debt_custody_nation: String::new(),
            foreign_debt_custody_amount: 0.0,
        }
    }
}

impl FiscalSystem {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn update_credit_rating(&mut self, gdp: f64, inflation_rate: f64) {
        let effective_gdp = if gdp <= 0.0 { 10_000_000_000.0 } else { gdp };
        // Net debt considers foreign exchange reserves in Global Reserve Currency (collateral buffer)
        let fx_buffer = (self.reserve_currency_reserves * 2.45 * 0.70).min(self.national_debt * 0.80);
        let net_debt = (self.national_debt - fx_buffer).max(0.0);
        let debt_ratio = net_debt / effective_gdp;

        if self.is_policy_active(POLICY_DEBT_RELIEF) {
            self.credit_rating = "Default (D)".to_string();
            self.borrowing_yield = 0.25;
            return;
        }
        if self.is_policy_active(POLICY_UNLIMITED_LIQUIDITY) {
            self.credit_rating = "Junk (CCC)".to_string();
            self.borrowing_yield = 0.145;
            return;
        }

        if debt_ratio < 0.25 {
            self.credit_rating = "AAA".to_string();
            self.borrowing_yield = 0.025;
        } else if debt_ratio < 0.45 {
            self.credit_rating = "AA".to_string();
            self.borrowing_yield = 0.032;
        } else if debt_ratio < 0.65 {
            self.credit_rating = "A+".to_string();
            self.borrowing_yield = 0.038;
        } else if debt_ratio < 0.85 {
            self.credit_rating = "BBB".to_string();
            self.borrowing_yield = 0.048;
        } else if debt_ratio < 1.10 {
            self.credit_rating = "BB".to_string();
            self.borrowing_yield = 0.065;
        } else if debt_ratio < 1.40 {
            self.credit_rating = "B".to_string();
            self.borrowing_yield = 0.088;
        } else {
            self.credit_rating = "Junk".to_string();
            self.borrowing_yield = 0.125;
        }

        if inflation_rate > 0.04 {
            self.borrowing_yield += (inflation_rate - 0.04) * 0.8;
        }
    }

    pub fn borrow_debt(&mut self, amount: f64) -> Result<(), &'static str> {
        if amount <= 0.0 {
            return Err("invalid borrowing amount");
        }
        self.national_debt += amount;
        self.treasury_cash += amount;
        Ok(())
    }

    pub fn repay_debt(&mut self, mut amount: f64) -> Result<(), &'static str> {
        if amount <= 0.0 {
            return Err("invalid repayment amount");
        }
        if self.treasury_cash < amount {
            return Err("insufficient treasury cash");
        }
        if self.national_debt < amount {
            amount = self.national_debt;
        }
        self.treasury_cash -= amount;
        self.national_debt -= amount;
        Ok(())
    }

    pub fn buy_reserve_currency(&mut self, crn_amount: f64, rate: f64) -> Result<f64, &'static str> {
        if crn_amount <= 0.0 {
            return Err("invalid purchase amount");
        }
        if self.treasury_cash < crn_amount {
            return Err("insufficient treasury cash");
        }
        if rate <= 0.0 {
            return Err("invalid exchange rate");
        }
        let zth_purchased = crn_amount / rate;
        self.treasury_cash -= crn_amount;
        self.reserve_currency_reserves += zth_purchased;
        Ok(zth_purchased)
    }

    pub fn sell_reserve_currency(&mut self, zth_amount: f64, rate: f64) -> Result<f64, &'static str> {
        if zth_amount <= 0.0 {
            return Err("invalid sale amount");
        }
        if self.reserve_currency_reserves < zth_amount {
            return Err("insufficient reserve currency balance");
        }
        if rate <= 0.0 {
            return Err("invalid exchange rate");
        }
        let crn_proceeds = zth_amount * rate;
        self.reserve_currency_reserves -= zth_amount;
        self.treasury_cash += crn_proceeds;
        Ok(crn_proceeds)
    }

    pub fn invest_infrastructure(&mut self, amount: f64) -> Result<(), &'static str> {
        if amount <= 0.0 {
            return Err("invalid investment amount");
        }
        if self.treasury_cash < amount {
            return Err("insufficient treasury cash");
        }
        self.treasury_cash -= amount;
        let gain = (amount / 500_000_000.0) * 4.5;
        self.infrastructure_level = (self.infrastructure_level + gain).min(100.0);
        self.export_capacity_level = (self.export_capacity_level + gain * 0.7).min(100.0);
        Ok(())
    }

    pub fn invest_population(&mut self, amount: f64, pillar: &str) -> Result<(), &'static str> {
        if amount <= 0.0 {
            return Err("invalid investment amount");
        }
        if self.treasury_cash < amount {
            return Err("insufficient treasury cash");
        }
        self.treasury_cash -= amount;
        let gain = (amount / 500_000_000.0) * 4.5;
        match pillar {
            "healthcare" => {
                self.healthcare_level = (self.healthcare_level + gain).min(100.0);
            }
            "education" => {
                self.education_level = (self.education_level + gain).min(100.0);
            }
            _ => {
                self.healthcare_level = (self.healthcare_level + gain * 0.5).min(100.0);
                self.education_level = (self.education_level + gain * 0.5).min(100.0);
            }
        }
        Ok(())
    }

    pub fn invest_enterprise_grants(&mut self, amount: f64) -> Result<(), &'static str> {
        if amount <= 0.0 {
            return Err("invalid investment amount");
        }
        if self.treasury_cash < amount {
            return Err("insufficient treasury cash");
        }
        self.treasury_cash -= amount;
        let gain = (amount / 500_000_000.0) * 6.0;
        self.enterprise_grants_level = (self.enterprise_grants_level + gain).min(100.0);
        Ok(())
    }

    pub fn charter_exchange(&mut self) -> Result<(), &'static str> {
        if self.infrastructure_level < 20.0 || self.education_level < 20.0 {
            return Err("cannot charter exchange: requires Infrastructure >= 20% and Education >= 20%");
        }
        self.exchange_chartered = true;
        Ok(())
    }

    pub fn set_maintenance_rate(&mut self, rate: f64) -> Result<(), &'static str> {
        if rate < 0.0 {
            return Err("maintenance rate cannot be negative");
        }
        self.maintenance_spending_rate = rate;
        Ok(())
    }

    pub fn tick(
        &mut self,
        citizen_taxes_paid: f64,
        corporate_taxes_paid: f64,
        tariff_revenue: f64,
        _interest_rate: f64,
        tick_fraction_of_year: f64,
    ) {
        // Continuous capital wear-and-tear (depreciation)
        let wear = (self.depreciation_rate * tick_fraction_of_year).clamp(0.0, 0.05);
        self.infrastructure_level = (self.infrastructure_level * (1.0 - wear)).max(5.0);
        self.healthcare_level = (self.healthcare_level * (1.0 - wear)).max(5.0);
        self.education_level = (self.education_level * (1.0 - wear)).max(5.0);
        self.enterprise_grants_level = (self.enterprise_grants_level * (1.0 - wear)).max(5.0);
        self.export_capacity_level = (self.export_capacity_level * (1.0 - wear)).max(5.0);

        // Statutory maintenance offset & ongoing asset enhancement
        // $120M maintains parity against depreciation; higher amounts actively upgrade capital quality
        let maintenance_gain = (self.maintenance_spending_rate / 200_000_000.0) * 3.5 * tick_fraction_of_year;
        self.infrastructure_level = (self.infrastructure_level + maintenance_gain * 1.1).min(100.0);
        self.healthcare_level = (self.healthcare_level + maintenance_gain * 1.0).min(100.0);
        self.education_level = (self.education_level + maintenance_gain * 0.9).min(100.0);
        self.export_capacity_level = (self.export_capacity_level + maintenance_gain * 0.8).min(100.0);

        // Dynamic Tourism receipts: scales with safety/healthcare, infrastructure connectivity, and credit rating
        let infra_factor = (self.infrastructure_level / 25.0).powf(0.85);
        let health_factor = (self.healthcare_level / 25.0).powf(0.50);
        let rating_mult = match self.credit_rating.as_str() {
            "AAA" => 1.35,
            "AA" => 1.20,
            "A+" => 1.10,
            "BBB" => 1.00,
            "BB" => 0.85,
            "B" => 0.70,
            _ => 0.50,
        };
        self.tourism_revenue = (70_000_000.0 * infra_factor * health_factor * rating_mult).max(10_000_000.0);

        // Productivity (TFP) Index: derived from physical infra, STEM education, enterprise grants
        self.productivity_index = 75.0
            + (self.infrastructure_level * 0.35)
            + (self.education_level * 0.45)
            + (self.enterprise_grants_level * 0.25);

        self.total_revenue =
            citizen_taxes_paid + corporate_taxes_paid + tariff_revenue + self.export_revenue + self.tourism_revenue;

        let effective_debt_rate = self.borrowing_yield.clamp(0.015, 0.15);
        self.net_interest_expense = self.national_debt * effective_debt_rate;

        let public_service_base =
            (self.infrastructure_level + self.healthcare_level + self.education_level) / 150.0;
        
        // Base mandatory spending dynamically scaled for ~$185B economy
        self.mandatory_spending = (24_000_000_000.0 * public_service_base).max(18_000_000_000.0);

        // Account for active policy recurring annual costs / revenues factoring in rollout efficacy
        let mut policy_net_cost = 0.0;
        for p in all_known_policies() {
            if self.is_policy_active(&p.id) {
                let days = *self.policy_rollouts.get(&p.id).unwrap_or(&p.rollout_days);
                let efficacy = (days as f64 / p.rollout_days.max(1) as f64).clamp(0.15, 1.0);
                policy_net_cost += p.annual_cost * efficacy;
            }
        }

        self.total_outlays =
            self.mandatory_spending + self.discretionary_spending + self.net_interest_expense + self.maintenance_spending_rate + policy_net_cost.max(0.0);
        
        if policy_net_cost < 0.0 {
            self.total_revenue += -policy_net_cost; // revenue-generating policy (e.g. tariffs, royalties)
        }

        self.budget_deficit = self.total_outlays - self.total_revenue;

        // INEVITABLE DEBT & REALISTIC CASH DYNAMICS:
        // Cash flow: actual tax collections and sovereign revenues add to treasury cash;
        // government outlays, public services, and interest expenses are paid directly from treasury cash.
        let net_cashflow = (self.total_revenue - self.total_outlays) * tick_fraction_of_year;
        self.treasury_cash += net_cashflow + (self.fdi_inflow * 0.04) * tick_fraction_of_year;

        // Inevitable debt trigger: If treasury cash falls below the $500M statutory emergency liquidity reserve,
        // the state is forced into emergency sovereign debt issuance to maintain government operations and payroll!
        if self.treasury_cash < 500_000_000.0 {
            let emergency_debt = (2_500_000_000.0 - self.treasury_cash).max(1_000_000_000.0);
            self.national_debt += emergency_debt;
            self.treasury_cash += emergency_debt;
            // Sovereign debt strain raises borrowing yields
            self.borrowing_yield = (self.borrowing_yield + 0.0035).min(0.25);
        }

        // Advance active policy rollout days (1 tick = 1 calendar day)
        for (pol_id, active) in &self.active_policies {
            if *active {
                *self.policy_rollouts.entry(pol_id.clone()).or_insert(0) += 1;
            }
        }

        // Dangerous policy consequences
        if self.is_policy_active(POLICY_PRICE_FREEZE) {
            self.total_revenue *= 0.55; // Corporate profit collapse
            self.productivity_index = (self.productivity_index - 8.0).max(40.0);
        }
        if self.is_policy_active(POLICY_AUTARKY_TARIFF) {
            self.export_revenue *= 0.15;
            self.tourism_revenue *= 0.10;
        }
        if self.is_policy_active(POLICY_WINDFALL_SEIZURE) {
            self.fdi_inflow = 0.0;
            self.export_revenue *= 0.30;
        }

        // Citizen policy benefits
        if self.is_policy_active(POLICY_UNIVERSAL_CHILDCARE) {
            self.productivity_index += 1.5;
        }
        if self.is_policy_active(POLICY_APPRENTICE_SUBSIDY) {
            self.productivity_index += 2.0;
        }

        self.procurement_spending_rate = self.discretionary_spending * 0.42;
    }

    pub fn sector_procurement_demand(&self, sector: &str) -> f64 {
        let mut base = match sector {
            "Industrials" => 1.35,
            "Information Technology" => 1.20,
            "Health Care" => 1.15,
            "Materials" => 1.10,
            _ => 1.00,
        };

        let is_active = |id: &str| self.active_policies.get(id).copied().unwrap_or(false);

        // Fiscal & Infra
        if is_active(POLICY_RURAL_ELECTRIFICATION) && sector == "Utilities" {
            base += 0.20;
        }
        if is_active(POLICY_PORT_MODERNIZATION) && sector == "Industrials" {
            base += 0.25;
        }
        if is_active(POLICY_TARIFF_SHIELD) && sector == "Consumer Discretionary" {
            base += 0.25;
        }
        if is_active(POLICY_HIGH_SPEED_RAIL) {
            if sector == "Industrials" {
                base += 0.30;
            } else if sector == "Materials" {
                base += 0.20;
            }
        }
        if is_active(POLICY_SOVEREIGN_WEALTH_FUND) && sector == "Financials" {
            base += 0.25;
        }
        if is_active(POLICY_DEEP_SPACE_PORT) {
            if sector == "Industrials" {
                base += 0.40;
            } else if sector == "Information Technology" {
                base += 0.30;
            }
        }

        // Monetary & Banking
        if is_active(POLICY_MICROFINANCE_ACT) && sector == "Financials" {
            base += 0.15;
        }
        if is_active(POLICY_RATE_CUT_50BP) {
            if sector == "Real Estate" {
                base += 0.25;
            } else {
                base += 0.05;
            }
        }
        if is_active(POLICY_RATE_HIKE_50BP) {
            if sector == "Financials" {
                base += 0.15;
            } else if sector == "Real Estate" {
                base -= 0.15;
            }
        }
        if is_active(POLICY_DISCOUNT_WINDOW_REFORM) && sector == "Financials" {
            base += 0.20;
        }
        if is_active(POLICY_QUANTITATIVE_EASING) {
            base += 0.10;
        }
        if is_active(POLICY_RESERVE_CURRENCY_OPS) && sector == "Financials" {
            base += 0.35;
        }

        // Trade
        if is_active(POLICY_AGRI_EXPORT_COMPACT) && sector == "Consumer Staples" {
            base += 0.25;
        }
        if is_active(POLICY_MINERAL_ROYALTY) && sector == "Materials" {
            base += 0.20;
        }
        if is_active(POLICY_REGIONAL_FREE_TRADE) {
            if sector == "Industrials" {
                base += 0.25;
            } else if sector == "Consumer Discretionary" {
                base += 0.15;
            }
        }
        if is_active(POLICY_ENERGY_EXPORT_TERMINAL) && sector == "Energy" {
            base += 0.35;
        }
        if is_active(POLICY_GLOBAL_TRADE_HEGEMONY) {
            base += 0.15;
        }

        // Industry & Tech
        if is_active(POLICY_SMALL_BUSINESS_INCUBATOR) && sector == "Consumer Discretionary" {
            base += 0.20;
        }
        if is_active(POLICY_LIGHT_MANUFACTURING_GRANTS) && sector == "Industrials" {
            base += 0.25;
        }
        if is_active(POLICY_SEMICONDUCTOR_FAB_INITIATIVE) && sector == "Information Technology" {
            base += 0.40;
        }
        if is_active(POLICY_CLEAN_ENERGY_GRID) {
            if sector == "Utilities" {
                base += 0.30;
            } else if sector == "Energy" {
                base += 0.20;
            }
        }
        if is_active(POLICY_AI_QUANTUM_MEGACENTER) && sector == "Information Technology" {
            base += 0.45;
        }

        // Labor & Welfare
        if is_active(POLICY_PRIMARY_HEALTH_CLINICS) && sector == "Health Care" {
            base += 0.20;
        }
        if is_active(POLICY_VOCATIONAL_APPRENTICESHIPS) && sector == "Industrials" {
            base += 0.20;
        }
        if is_active(POLICY_LIVING_WAGE_GUARANTEE) && sector == "Consumer Staples" {
            base += 0.20;
        }
        if is_active(POLICY_HIGH_SKILL_VISA) && sector == "Information Technology" {
            base += 0.25;
        }
        if is_active(POLICY_CITIZEN_STIMULUS) {
            base += 0.15;
        }

        // Foreign
        if is_active(POLICY_GOOD_NEIGHBOR_MISSION) && sector == "Consumer Staples" {
            base += 0.10;
        }
        if is_active(POLICY_REGIONAL_SECURITY_PACT) && sector == "Industrials" {
            base += 0.20;
        }
        if is_active(POLICY_SUPERPOWER_ALLIANCE) {
            base += 0.15;
        }

        // Citizen & Strategic Directives
        if is_active(POLICY_UNIVERSAL_CHILDCARE) && sector == "Health Care" {
            base += 0.20;
        }
        if is_active(POLICY_APPRENTICE_SUBSIDY) && sector == "Industrials" {
            base += 0.25;
        }
        if is_active(POLICY_SALES_VAT) && sector == "Consumer Staples" {
            base -= 0.15;
        }
        if is_active(POLICY_LABOR_DEREGULATION) && sector == "Industrials" {
            base += 0.20;
        }
        if is_active(POLICY_PRICE_FREEZE) {
            base -= 0.30;
        }
        if is_active(POLICY_AUTARKY_TARIFF) && sector == "Materials" {
            base += 0.30;
        }

        base
    }

    pub fn toggle_policy_checked(&mut self, policy_id: &str, current_tier: &str) -> Result<(bool, String), String> {
        let policy_info = all_known_policies()
            .into_iter()
            .find(|p| p.id == policy_id)
            .ok_or_else(|| format!("Unknown policy: {}", policy_id))?;

        // Check tier prerequisite
        let tier_rank = |t: &str| match t {
            "powerhouse" => 3,
            "mid" => 2,
            _ => 1,
        };

        if tier_rank(current_tier) < tier_rank(&policy_info.min_tier) {
            let required_name = match policy_info.min_tier.as_str() {
                "mid" => "Emerging State (GDP ≥ $60B)",
                "powerhouse" => "Global Powerhouse (GDP ≥ $300B)",
                _ => "Small Nation",
            };
            return Err(format!("Policy '{}' is locked. Requires {}.", policy_info.name, required_name));
        }

        let current = self.active_policies.get(policy_id).copied().unwrap_or(false);
        let new_state = !current;
        self.active_policies.insert(policy_id.to_string(), new_state);

        if new_state {
            self.policy_rollouts.insert(policy_id.to_string(), 1);
            if policy_id == POLICY_WINDFALL_SEIZURE {
                self.treasury_cash += 10_000_000_000.0;
            } else if policy_id == POLICY_DEBT_RELIEF {
                self.national_debt *= 0.50;
            }
        } else {
            self.policy_rollouts.remove(policy_id);
        }

        Ok((new_state, policy_info.name))
    }

    pub fn set_tax_rates(&mut self, personal_income: Option<f64>, corporate: Option<f64>, sales: Option<f64>) {
        if let Some(r) = personal_income {
            self.personal_income_tax_rate = r.clamp(0.05, 0.50);
        }
        if let Some(r) = corporate {
            self.corporate_tax_rate = r.clamp(0.05, 0.45);
        }
        if let Some(r) = sales {
            self.sales_tax_rate = r.clamp(0.0, 0.30);
        }
    }

    pub fn borrow_bilateral(&mut self, lender_country_id: &str, amount: f64, offered_interest_rate: f64) -> Result<String, &'static str> {
        if amount <= 0.0 {
            return Err("Borrow amount must be greater than zero");
        }
        self.national_debt += amount;
        self.treasury_cash += amount;
        self.borrowing_yield = (self.borrowing_yield * 0.80 + offered_interest_rate * 0.20).clamp(0.01, 0.25);
        Ok(format!("Successfully borrowed ${:.2}M from sovereign partner {}", amount / 1e6, lender_country_id))
    }

    pub fn toggle_foreign_debt_custody(&mut self, nation_id: &str, enable: bool) -> (bool, f64) {
        self.foreign_debt_custody_enabled = enable;
        if enable {
            self.foreign_debt_custody_nation = nation_id.to_string();
            self.foreign_debt_custody_amount = 15_000_000_000.0; // $15B in foreign sovereign custody assets
        } else {
            self.foreign_debt_custody_nation = String::new();
            self.foreign_debt_custody_amount = 0.0;
        }
        (self.foreign_debt_custody_enabled, self.foreign_debt_custody_amount)
    }

    pub fn is_policy_active(&self, policy_id: &str) -> bool {
        self.active_policies.get(policy_id).copied().unwrap_or(false)
    }

    pub fn get_active_policies(&self) -> Vec<String> {
        self.active_policies
            .iter()
            .filter(|(_, &active)| active)
            .map(|(k, _)| k.clone())
            .collect()
    }
}
