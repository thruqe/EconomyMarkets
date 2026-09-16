package fiscal

import (
	"fmt"
	"math"
	"sync"
)

type FiscalSystem struct {
	mu                      sync.RWMutex
	TotalRevenue            float64         `json:"total_revenue"`
	TotalOutlays            float64         `json:"total_outlays"`
	BudgetDeficit           float64         `json:"budget_deficit"`
	NationalDebt            float64         `json:"national_debt"`
	CorporateTaxRate        float64         `json:"corporate_tax_rate"`
	MandatorySpending       float64         `json:"mandatory_spending"`
	DiscretionarySpending   float64         `json:"discretionary_spending"`
	NetInterestExpense      float64         `json:"net_interest_expense"`
	ProcurementSpendingRate float64         `json:"procurement_spending_rate"`
	ActivePolicies          map[string]bool `json:"active_policies"`

	// Sovereign Treasury & Nation Building
	TreasuryCash          float64 `json:"treasury_cash"`           // Liquid sovereign funds in treasury
	CreditRating          string  `json:"credit_rating"`           // "AAA", "AA", "A", "BBB", "BB", "B", "Junk"
	BorrowingYield        float64 `json:"borrowing_yield"`         // Annual yield required by lenders (e.g. 0.038)
	InfrastructureLevel   float64 `json:"infrastructure_level"`    // 0-100 index (Transport, Power, Ports)
	HealthcareLevel       float64 `json:"healthcare_level"`        // 0-100 index (Hospitals, public health)
	EducationLevel        float64 `json:"education_level"`         // 0-100 index (Universities, workforce skills)
	EnterpriseGrantsLevel float64 `json:"enterprise_grants_level"`  // 0-100 index (Startup incubators, seed funds)
	ExportCapacityLevel   float64 `json:"export_capacity_level"`    // 0-100 index (Export facilities, trade logistics)
	FDIInflow             float64 `json:"fdi_inflow"`              // Annualized Foreign Direct Investment in USD
	ExportRevenue         float64 `json:"export_revenue"`          // Annualized sovereign export earnings in USD
	ExchangeChartered     bool    `json:"exchange_chartered"`      // Is the National Stock Exchange inaugurated?
}

// NewDefaultFiscalSystem initializes baseline sovereign fiscal statistics.
func NewDefaultFiscalSystem() *FiscalSystem {
	debt := 34_500_000_000_000.0
	mand := 4_100_000_000_000.0
	disc := 1_800_000_000_000.0
	yield := 0.042
	interest := debt * yield
	outlays := mand + disc + interest
	rev := 4_900_000_000_000.0

	return &FiscalSystem{
		TotalRevenue:            rev,
		TotalOutlays:            outlays,
		BudgetDeficit:           outlays - rev,
		NationalDebt:            debt,
		CorporateTaxRate:        0.21,
		MandatorySpending:       mand,
		DiscretionarySpending:   disc,
		NetInterestExpense:      interest,
		ProcurementSpendingRate: 800_000_000_000.0,
		ActivePolicies:          make(map[string]bool),

		// Sovereign Treasury & Nation Building initial values
		TreasuryCash:          500_000_000_000.0,
		CreditRating:          "AA+",
		BorrowingYield:        yield,
		InfrastructureLevel:   70.0,
		HealthcareLevel:       70.0,
		EducationLevel:        75.0,
		EnterpriseGrantsLevel: 60.0,
		ExportCapacityLevel:   75.0,
		ExchangeChartered:     true,
	}
}

// UpdateCreditRating calculates sovereign credit rating and borrowing rate based on Debt/GDP and inflation.
func (f *FiscalSystem) UpdateCreditRating(gdp, inflationRate float64) {
	if gdp <= 0 {
		gdp = 1_000_000_000.0
	}
	debtRatio := f.NationalDebt / gdp

	switch {
	case debtRatio < 0.40:
		f.CreditRating = "AAA"
		f.BorrowingYield = 0.028
	case debtRatio < 0.65:
		f.CreditRating = "AA"
		f.BorrowingYield = 0.034
	case debtRatio < 0.90:
		f.CreditRating = "A"
		f.BorrowingYield = 0.041
	case debtRatio < 1.15:
		f.CreditRating = "BBB"
		f.BorrowingYield = 0.053
	case debtRatio < 1.45:
		f.CreditRating = "BB"
		f.BorrowingYield = 0.072
	case debtRatio < 1.85:
		f.CreditRating = "B"
		f.BorrowingYield = 0.098
	default:
		f.CreditRating = "Junk"
		f.BorrowingYield = 0.138
	}

	// High inflation pushes up bond yields
	if inflationRate > 0.04 {
		f.BorrowingYield += (inflationRate - 0.04) * 0.8
	}
}

// BorrowDebt issues sovereign bonds, adding liquid cash to TreasuryCash and expanding NationalDebt.
func (f *FiscalSystem) BorrowDebt(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("invalid borrowing amount")
	}
	f.NationalDebt += amount
	f.TreasuryCash += amount
	return nil
}

// RepayDebt pays down sovereign national debt using available TreasuryCash.
func (f *FiscalSystem) RepayDebt(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("invalid repayment amount")
	}
	if f.TreasuryCash < amount {
		return fmt.Errorf("insufficient treasury cash (available: $%.2fB)", f.TreasuryCash/1e9)
	}
	if f.NationalDebt < amount {
		amount = f.NationalDebt
	}
	f.TreasuryCash -= amount
	f.NationalDebt -= amount
	return nil
}

// InvestInfrastructure allocates treasury funds to power grids, ports, transport, and utilities.
func (f *FiscalSystem) InvestInfrastructure(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("invalid investment amount")
	}
	if f.TreasuryCash < amount {
		return fmt.Errorf("insufficient treasury cash (available: $%.2fB)", f.TreasuryCash/1e9)
	}
	f.TreasuryCash -= amount
	gain := (amount / 2_000_000_000.0) * 4.5
	f.InfrastructureLevel = math.Min(100.0, f.InfrastructureLevel+gain)
	f.ExportCapacityLevel = math.Min(100.0, f.ExportCapacityLevel+gain*0.7)
	return nil
}

// InvestPopulation allocates treasury funds to healthcare, education, or family grants.
func (f *FiscalSystem) InvestPopulation(amount float64, pillar string) error {
	if amount <= 0 {
		return fmt.Errorf("invalid investment amount")
	}
	if f.TreasuryCash < amount {
		return fmt.Errorf("insufficient treasury cash (available: $%.2fB)", f.TreasuryCash/1e9)
	}
	f.TreasuryCash -= amount
	gain := (amount / 2_000_000_000.0) * 4.5

	switch pillar {
	case "healthcare":
		f.HealthcareLevel = math.Min(100.0, f.HealthcareLevel+gain)
	case "education":
		f.EducationLevel = math.Min(100.0, f.EducationLevel+gain)
	default: // family & housing
		f.HealthcareLevel = math.Min(100.0, f.HealthcareLevel+gain*0.5)
		f.EducationLevel = math.Min(100.0, f.EducationLevel+gain*0.5)
	}
	return nil
}

// InvestEnterpriseGrants provides incubation seed capital to spur emerging domestic companies.
func (f *FiscalSystem) InvestEnterpriseGrants(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("invalid investment amount")
	}
	if f.TreasuryCash < amount {
		return fmt.Errorf("insufficient treasury cash (available: $%.2fB)", f.TreasuryCash/1e9)
	}
	f.TreasuryCash -= amount
	gain := (amount / 2_000_000_000.0) * 6.0
	f.EnterpriseGrantsLevel = math.Min(100.0, f.EnterpriseGrantsLevel+gain)
	return nil
}

// CharterExchange opens the National Stock Exchange for public IPO listings and trading.
func (f *FiscalSystem) CharterExchange() error {
	if f.InfrastructureLevel < 20.0 || f.EducationLevel < 20.0 {
		return fmt.Errorf("cannot charter exchange: requires Infrastructure >= 20%% and Education >= 20%%")
	}
	f.ExchangeChartered = true
	return nil
}

// Tick updates federal revenue, outlays, interest payments, compounds debt, and updates TreasuryCash balance.
func (f *FiscalSystem) Tick(citizenTaxesPaid, corporateTaxesPaid, tariffRevenue, interestRate, tickFractionOfYear float64) {
	// 1. Total revenue including export earnings
	f.TotalRevenue = citizenTaxesPaid + corporateTaxesPaid + tariffRevenue + f.ExportRevenue

	// 2. Net interest expense on sovereign debt
	effectiveDebtRate := math.Max(0.015, math.Min(0.15, f.BorrowingYield))
	f.NetInterestExpense = f.NationalDebt * effectiveDebtRate

	// 3. Outlays: public services + administration + debt interest
	// Base public services scale with infrastructure, healthcare, and education levels
	publicServiceBase := (f.InfrastructureLevel + f.HealthcareLevel + f.EducationLevel) / 215.0
	if f.NationalDebt < 100_000_000_000.0 {
		// Emerging economy: public services scale around $800M - $2B
		f.MandatorySpending = math.Max(600_000_000.0, 1_200_000_000.0*publicServiceBase)
	} else {
		// Large established economy: scales around $3.5T - $4.5T
		f.MandatorySpending = math.Max(3_500_000_000_000.0, 4_100_000_000_000.0*publicServiceBase)
	}

	f.TotalOutlays = f.MandatorySpending + f.DiscretionarySpending + f.NetInterestExpense

	// 4. Deficit/Surplus and Treasury Cashflow
	f.BudgetDeficit = f.TotalOutlays - f.TotalRevenue

	// Net cash flow to the sovereign treasury:
	// Surplus (TotalRevenue > TotalOutlays) deposits into TreasuryCash!
	// Deficit draws down TreasuryCash. If TreasuryCash runs dry, it automatically borrows into NationalDebt.
	netCashflow := (f.TotalRevenue - f.TotalOutlays + f.FDIInflow*0.12) * tickFractionOfYear
	f.TreasuryCash += netCashflow

	if f.TreasuryCash < 0 {
		// Treasury liquidity shortfall is absorbed by debt issuance
		f.NationalDebt += -f.TreasuryCash
		f.TreasuryCash = 0
	}

	// Procurement spending tracks discretionary capital budget
	f.ProcurementSpendingRate = f.DiscretionarySpending * 0.42
}

const (
	PolicyTechSubsidies         = "tech_subsidies"
	PolicyDefenseInfrastructure = "defense_infra"
	PolicyCleanEnergy           = "clean_energy"
	PolicyTariffShield          = "tariff_shield"
	PolicyCitizenStimulus       = "citizen_stimulus"

	// Monetary
	PolicyRateHike50bp          = "rate_hike_50bp"
	PolicyRateCut50bp           = "rate_cut_50bp"
	PolicyQuantitativeEasing    = "quantitative_easing"
	PolicyQuantitativeTightening = "quantitative_tightening"

	// Trade
	PolicyFreeTradesPact         = "free_trade_pact"
	PolicyChinaTechBan           = "china_tech_ban"
	PolicyDomesticContentMandate = "domestic_content_mandate"
	PolicyEnergyExportExpansion  = "energy_export_expansion"

	// Regulatory
	PolicyTechAntitrust          = "tech_antitrust"
	PolicyFinancialDeregulation  = "financial_deregulation"
	PolicyHealthcarePriceControls = "healthcare_price_controls"
	PolicyHousingDeregulation    = "housing_deregulation"

	// Labor
	PolicyMinWageHike         = "min_wage_hike"
	PolicyImmigrationExpansion = "immigration_expansion"
	PolicyUnionProtectionAct  = "union_protection_act"
)

// PolicyInfo defines an executive government economic policy.
type PolicyInfo struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Category      string  `json:"category"` // "fiscal", "monetary", "trade", "regulatory", "labor"
	Description   string  `json:"description"`
	FavoredSector string  `json:"favored_sector"`
	AnnualCost    float64 `json:"annual_cost"`
}

// KnownPolicies holds all fiscal (executive spending) policies.
var KnownPolicies = []PolicyInfo{
	{
		ID: PolicyTechSubsidies, Category: "fiscal",
		Name:          "CHIPS & Tech Subsidies",
		Description:   "+25% demand to Tech via domestic semiconductor & cloud grants",
		FavoredSector: "Information Technology",
		AnnualCost:    50_000_000_000.0,
	},
	{
		ID: PolicyDefenseInfrastructure, Category: "fiscal",
		Name:          "Defense & Infrastructure Act",
		Description:   "+35% demand to Industrials & +25% Materials for defense & rebuilding",
		FavoredSector: "Industrials",
		AnnualCost:    80_000_000_000.0,
	},
	{
		ID: PolicyCleanEnergy, Category: "fiscal",
		Name:          "Clean Power & Grid Credits",
		Description:   "+25% demand to Utilities & +20% Energy for next-gen power expansion",
		FavoredSector: "Utilities",
		AnnualCost:    40_000_000_000.0,
	},
	{
		ID: PolicyTariffShield, Category: "fiscal",
		Name:          "Strategic Import Tariffs",
		Description:   "Raises tariffs to 6.5%, protecting domestic Consumer Discretionary",
		FavoredSector: "Consumer Discretionary",
		AnnualCost:    -45_000_000_000.0,
	},
	{
		ID: PolicyCitizenStimulus, Category: "fiscal",
		Name:          "Emergency Citizen Stimulus",
		Description:   "$2,000 direct checks to 160M citizens, surging sentiment & spending",
		FavoredSector: "All Sectors",
		AnnualCost:    320_000_000_000.0,
	},
}

// KnownMonetaryPolicies holds central bank rate & balance sheet policies.
var KnownMonetaryPolicies = []PolicyInfo{
	{
		ID: PolicyRateHike50bp, Category: "monetary",
		Name:          "Rate Hike +0.50%",
		Description:   "Central bank raises key rate 50bp — fights inflation, pressures growth stocks",
		FavoredSector: "Financials",
		AnnualCost:    0,
	},
	{
		ID: PolicyRateCut50bp, Category: "monetary",
		Name:          "Rate Cut -0.50%",
		Description:   "Central bank cuts 50bp — stimulates borrowing & investment, boosts equities",
		FavoredSector: "Real Estate",
		AnnualCost:    0,
	},
	{
		ID: PolicyQuantitativeEasing, Category: "monetary",
		Name:          "Quantitative Easing (QE)",
		Description:   "Buys $120B/mo in bonds — expands money supply, lowers long-term yields",
		FavoredSector: "All Sectors",
		AnnualCost:    1_440_000_000_000.0,
	},
	{
		ID: PolicyQuantitativeTightening, Category: "monetary",
		Name:          "Quantitative Tightening (QT)",
		Description:   "Shrinks balance sheet $90B/mo — reduces liquidity, tightens conditions",
		FavoredSector: "Financials",
		AnnualCost:    -1_080_000_000_000.0,
	},
}

// KnownTradePolicies holds trade agreements and export/import policies.
var KnownTradePolicies = []PolicyInfo{
	{
		ID: PolicyFreeTradesPact, Category: "trade",
		Name:          "Global Free Trade Pact",
		Description:   "Eliminates tariffs with 30+ nations — boosts exports, lowers consumer prices",
		FavoredSector: "Consumer Staples",
		AnnualCost:    -80_000_000_000.0,
	},
	{
		ID: PolicyChinaTechBan, Category: "trade",
		Name:          "China Tech Decoupling",
		Description:   "Bans Chinese tech imports & restricts exports — reshores semiconductor manufacturing",
		FavoredSector: "Information Technology",
		AnnualCost:    15_000_000_000.0,
	},
	{
		ID: PolicyDomesticContentMandate, Category: "trade",
		Name:          "Domestic Content Mandate",
		Description:   "60% domestic content for gov contracts — boosts Industrials & Materials",
		FavoredSector: "Industrials",
		AnnualCost:    25_000_000_000.0,
	},
	{
		ID: PolicyEnergyExportExpansion, Category: "trade",
		Name:          "LNG & Oil Export Expansion",
		Description:   "Opens new export terminals & pipelines — expands Energy sector revenue",
		FavoredSector: "Energy",
		AnnualCost:    -30_000_000_000.0,
	},
}

// KnownRegulatoryPolicies holds antitrust, deregulation, and sector regulation policies.
var KnownRegulatoryPolicies = []PolicyInfo{
	{
		ID: PolicyTechAntitrust, Category: "regulatory",
		Name:          "Big Tech Antitrust",
		Description:   "DOJ/FTC breaks up dominant tech platforms — disrupts megacaps, opens markets",
		FavoredSector: "Communication Services",
		AnnualCost:    5_000_000_000.0,
	},
	{
		ID: PolicyFinancialDeregulation, Category: "regulatory",
		Name:          "Financial Deregulation",
		Description:   "Loosens bank capital requirements — expands credit & M&A activity",
		FavoredSector: "Financials",
		AnnualCost:    -12_000_000_000.0,
	},
	{
		ID: PolicyHealthcarePriceControls, Category: "regulatory",
		Name:          "Drug Price Controls",
		Description:   "Medicare negotiates drug prices — reduces pharma margins, boosts citizen health",
		FavoredSector: "Health Care",
		AnnualCost:    -65_000_000_000.0,
	},
	{
		ID: PolicyHousingDeregulation, Category: "regulatory",
		Name:          "Zoning & Housing Deregulation",
		Description:   "Federal zoning override — unlocks real estate development, lowers home prices",
		FavoredSector: "Real Estate",
		AnnualCost:    8_000_000_000.0,
	},
}

// KnownLaborPolicies holds minimum wage, immigration, and labor rights policies.
var KnownLaborPolicies = []PolicyInfo{
	{
		ID: PolicyMinWageHike, Category: "labor",
		Name:          "Federal Minimum Wage Hike",
		Description:   "$18/hr federal minimum — raises consumer spending, increases employer costs",
		FavoredSector: "Consumer Staples",
		AnnualCost:    40_000_000_000.0,
	},
	{
		ID: PolicyImmigrationExpansion, Category: "labor",
		Name:          "High-Skilled Immigration Expansion",
		Description:   "+500K H-1B/yr — fills tech & engineering gaps, suppresses wage inflation",
		FavoredSector: "Information Technology",
		AnnualCost:    2_000_000_000.0,
	},
	{
		ID: PolicyUnionProtectionAct, Category: "labor",
		Name:          "Union Protection Act",
		Description:   "Expands collective bargaining — raises Industrials wages, increases cost pressure",
		FavoredSector: "Industrials",
		AnnualCost:    15_000_000_000.0,
	},
}

// AllKnownPolicies returns every policy across all categories for toggle/lookup.
func AllKnownPolicies() []PolicyInfo {
	var all []PolicyInfo
	all = append(all, KnownPolicies...)
	all = append(all, KnownMonetaryPolicies...)
	all = append(all, KnownTradePolicies...)
	all = append(all, KnownRegulatoryPolicies...)
	all = append(all, KnownLaborPolicies...)
	return all
}

// SectorProcurementDemand returns the government contract multiplier for a sector,
// factoring in any enacted executive government policies.
func (f *FiscalSystem) SectorProcurementDemand(sector string) float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	base := 1.00
	switch sector {
	case "Industrials":
		base = 1.45 // Defense, aerospace, shipbuilding contractors
	case "Information Technology":
		base = 1.25 // Cyber defense, government cloud & software
	case "Health Care":
		base = 1.20 // Medicare, Medicaid, VA procurement
	case "Materials":
		base = 1.15 // Infrastructure steel, rare minerals
	default:
		base = 1.00
	}

	// --- Fiscal Policy Boosts ---
	if f.ActivePolicies[PolicyTechSubsidies] && sector == "Information Technology" {
		base += 0.25
	}
	if f.ActivePolicies[PolicyDefenseInfrastructure] {
		if sector == "Industrials" {
			base += 0.35
		} else if sector == "Materials" {
			base += 0.25
		}
	}
	if f.ActivePolicies[PolicyCleanEnergy] {
		if sector == "Utilities" {
			base += 0.25
		} else if sector == "Energy" {
			base += 0.20
		}
	}
	if f.ActivePolicies[PolicyTariffShield] && sector == "Consumer Discretionary" {
		base += 0.20
	}
	if f.ActivePolicies[PolicyCitizenStimulus] {
		base += 0.10 // All sectors benefit from stimulus spending
	}

	// --- Monetary Policy Effects ---
	if f.ActivePolicies[PolicyRateHike50bp] {
		if sector == "Financials" {
			base += 0.15 // Banks earn more on spread
		} else if sector == "Real Estate" {
			base -= 0.20 // Higher rates crush real estate
		} else if sector == "Information Technology" {
			base -= 0.10 // Growth stocks hurt by higher rates
		}
	}
	if f.ActivePolicies[PolicyRateCut50bp] {
		if sector == "Real Estate" {
			base += 0.25 // Cheap mortgages fuel real estate
		} else if sector == "Financials" {
			base -= 0.10 // Compressed margins
		} else {
			base += 0.05 // Broad mild positive
		}
	}
	if f.ActivePolicies[PolicyQuantitativeEasing] {
		base += 0.08 // Broad liquidity boost to all sectors
	}
	if f.ActivePolicies[PolicyQuantitativeTightening] {
		base -= 0.05 // Broad mild tightening drag
		if sector == "Financials" {
			base += 0.10 // Banks benefit from tighter conditions
		}
	}

	// --- Trade Policy Effects ---
	if f.ActivePolicies[PolicyFreeTradesPact] {
		if sector == "Consumer Staples" {
			base += 0.20
		} else if sector == "Energy" {
			base += 0.15
		} else {
			base += 0.05
		}
	}
	if f.ActivePolicies[PolicyChinaTechBan] {
		if sector == "Information Technology" {
			base += 0.30 // Reshoring windfall for domestic tech
		}
	}
	if f.ActivePolicies[PolicyDomesticContentMandate] {
		if sector == "Industrials" {
			base += 0.20
		} else if sector == "Materials" {
			base += 0.15
		}
	}
	if f.ActivePolicies[PolicyEnergyExportExpansion] && sector == "Energy" {
		base += 0.35
	}

	// --- Regulatory Policy Effects ---
	if f.ActivePolicies[PolicyTechAntitrust] {
		if sector == "Information Technology" {
			base -= 0.15 // Penalties and breakup costs
		} else if sector == "Communication Services" {
			base += 0.10 // Competition unlocked
		}
	}
	if f.ActivePolicies[PolicyFinancialDeregulation] && sector == "Financials" {
		base += 0.25
	}
	if f.ActivePolicies[PolicyHealthcarePriceControls] && sector == "Health Care" {
		base -= 0.15 // Margin compression on pharma
	}
	if f.ActivePolicies[PolicyHousingDeregulation] && sector == "Real Estate" {
		base += 0.20
	}

	// --- Labor Policy Effects ---
	if f.ActivePolicies[PolicyMinWageHike] {
		if sector == "Consumer Staples" {
			base += 0.15 // More spending power → more consumption
		} else if sector == "Consumer Discretionary" {
			base += 0.10
		} else if sector == "Industrials" {
			base -= 0.10 // Higher labor costs
		}
	}
	if f.ActivePolicies[PolicyImmigrationExpansion] && sector == "Information Technology" {
		base += 0.20 // Talent pool expansion
	}
	if f.ActivePolicies[PolicyUnionProtectionAct] {
		if sector == "Industrials" {
			base -= 0.10 // Cost pressure
		} else if sector == "Consumer Staples" {
			base += 0.08 // Better paid workers spend more
		}
	}

	return base
}

// TogglePolicy flips an executive policy on or off, returning the new state and policy name.
func (f *FiscalSystem) TogglePolicy(policyID string) (active bool, name string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.ActivePolicies == nil {
		f.ActivePolicies = make(map[string]bool)
	}
	f.ActivePolicies[policyID] = !f.ActivePolicies[policyID]
	active = f.ActivePolicies[policyID]

	for _, p := range AllKnownPolicies() {
		if p.ID == policyID {
			name = p.Name
			break
		}
	}
	if name == "" {
		name = policyID
	}
	return active, name
}

// IsPolicyActive checks if a policy is currently enacted.
func (f *FiscalSystem) IsPolicyActive(policyID string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.ActivePolicies == nil {
		return false
	}
	return f.ActivePolicies[policyID]
}

// GetActivePolicies returns a safe, copied list of currently active policy IDs.
func (f *FiscalSystem) GetActivePolicies() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var list []string
	for p, active := range f.ActivePolicies {
		if active {
			list = append(list, p)
		}
	}
	return list
}

// GetActivePolicyMap returns a safe copy of the active policies map.
func (f *FiscalSystem) GetActivePolicyMap() map[string]bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	res := make(map[string]bool, len(f.ActivePolicies))
	for k, v := range f.ActivePolicies {
		res[k] = v
	}
	return res
}


