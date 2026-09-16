package country

import (
	"math"

	"economy/country/centralbank"
	"economy/country/fiscal"
	"economy/country/labor"
	"economy/country/trade"
)

// NationalReport captures comprehensive macroeconomic telemetry of the United States economy.
type NationalReport struct {
	// National Output
	GDP          float64 // Nominal GDP in USD (e.g. ~$28.5 Trillion)
	RealGDPGrowth float64 // Real GDP annualized growth rate (e.g. 0.024 for +2.4%)

	// Monetary & Rates (Federal Reserve)
	FedFundsRate     float64 // Overnight target rate (e.g. 0.045 for 4.50%)
	TenYearYield     float64 // 10-Year Treasury Yield (e.g. 0.042 for 4.20%)
	CPIInflationRate float64 // Consumer Price Index YoY inflation (e.g. 0.024 for 2.40%)
	PolicyStance     string  // "Hawkish", "Neutral", "Dovish"
	FOMCAnnouncement string  // Non-empty when FOMC changes policy rate

	// Employment & Labor Market
	UnemploymentRate   float64 // U3 Civilian Unemployment rate (e.g. 0.039 for 3.9%)
	LaborForce         float64 // Total civilian labor force (e.g. 168.3M)
	EmployedWorkers    float64 // Total employed headcount (e.g. 161.7M)
	AverageHourlyWage  float64 // Average hourly wage in USD (e.g. $34.50)
	AnnualWageGrowth   float64 // Annualized wage growth rate (e.g. 0.040 for 4.0%)
	NetMonthlyPayrolls float64 // Change in non-farm payrolls (e.g. +185,000)

	// International Trade & Currency
	AnnualExports  float64 // Total exports in USD (e.g. $3.1T)
	AnnualImports  float64 // Total imports in USD (e.g. $3.9T)
	TradeBalance   float64 // Net Exports = Exports - Imports (deficit ~$800B)
	DollarIndexDXY float64 // U.S. Dollar Index (baseline 103.5)

	// Federal Fiscal Condition
	FederalRevenue float64 // Federal revenues in USD
	FederalOutlays float64 // Federal expenditures in USD
	BudgetDeficit  float64 // Annual deficit in USD (negative = surplus)
	NationalDebt   float64 // Total national debt in USD

	// Sovereign Treasury & Nation Building
	TreasuryCash          float64 // Liquid funds in sovereign treasury in USD
	CreditRating          string  // "AAA", "AA", "A", "BBB", "BB", "B", "Junk"
	BorrowingYield        float64 // Current bond borrowing rate
	InfrastructureLevel   float64 // 0-100 index (transport, power, logistics)
	HealthcareLevel       float64 // 0-100 index (public health, hospitals)
	EducationLevel        float64 // 0-100 index (skills, universities)
	EnterpriseGrantsLevel float64 // 0-100 index (startup seed funding)
	ExportCapacityLevel   float64 // 0-100 index (export terminals, ports)
	FDIInflow             float64 // Annualized foreign direct investment
	ExportRevenue         float64 // Annualized export earnings
	ExchangeChartered     bool    // Is the National Stock Exchange open?

	// Government Executive Policies
	ActivePolicies []string // IDs of currently active policies
}

// NationalEconomy coordinates the Federal Reserve, Labor Market, Foreign Trade, and Federal Fiscal system.
type NationalEconomy struct {
	Labor       *labor.LaborMarket
	Trade       *trade.TradeEconomy
	Fiscal      *fiscal.FiscalSystem
	CentralBank *centralbank.CentralBank

	NominalGDP    float64
	RealGDPGrowth float64
	prevGDP       float64
}

// NewDefaultNationalEconomy initializes a sovereign macroeconomic system.
func NewDefaultNationalEconomy() *NationalEconomy {
	baseGDP := 28_000_000_000_000.0 // Starting baseline US GDP: $28T

	return &NationalEconomy{
		Labor:         labor.NewDefaultLaborMarket(),
		Trade:         trade.NewDefaultTradeEconomy(),
		Fiscal:        fiscal.NewDefaultFiscalSystem(),
		CentralBank:   centralbank.NewDefaultCentralBank(),
		NominalGDP:    baseGDP,
		RealGDPGrowth: 0.035,
		prevGDP:       baseGDP,
	}
}


// Tick advances the national economy:
// - consumerSpending: annualized PCE consumption C from citizen package
// - corporateInvestment: annualized gross private domestic investment I (corporate CapEx)
// - totalCorporateHeadcount: total enterprise payrolls demand
// - corporateTaxesPaid: annualized corporate income taxes paid
// - citizenTaxesPaid: annualized income and sales taxes collected from citizens
// - tickFractionOfYear: fraction of year per tick
func (n *NationalEconomy) Tick(
	consumerSpending float64,
	corporateInvestment float64,
	totalCorporateHeadcount float64,
	corporateTaxesPaid float64,
	citizenTaxesPaid float64,
	tickFractionOfYear float64,
) NationalReport {
	// 1. Advance Federal Reserve monetary policy and inflation
	rateChanged, fomcAnnouncement := n.CentralBank.Tick(
		n.Labor.UnemploymentRate,
		n.Labor.AnnualWageGrowth,
		tickFractionOfYear,
	)

	// 2. Advance Labor Market with corporate employment demand and inflation
	n.Labor.Tick(
		totalCorporateHeadcount,
		n.CentralBank.CPIInflationRate,
		tickFractionOfYear,
	)

	// 3. Advance Foreign Trade and Dollar Index (DXY) with interest rates and tariffs
	if n.Fiscal.IsPolicyActive(fiscal.PolicyTariffShield) {
		n.Trade.AverageTariffRate = 0.065
	} else {
		n.Trade.AverageTariffRate = 0.028
	}
	n.Trade.Tick(
		n.CentralBank.FedFundsRate,
		tickFractionOfYear,
	)

	// 4. Update Sovereign Credit Rating & Foreign Inflows
	n.Fiscal.UpdateCreditRating(n.NominalGDP, n.CentralBank.CPIInflationRate)

	// Foreign Direct Investment scales with infrastructure, legal/education stability, and GDP
	fdiBase := n.NominalGDP * 0.02 * (n.Fiscal.InfrastructureLevel / 25.0)
	if n.CentralBank.CPIInflationRate < 0.045 {
		fdiBase *= 1.20 // Currency stability bonus
	}
	n.Fiscal.FDIInflow = fdiBase

	// Export revenue scales with export capacity level and GDP
	n.Fiscal.ExportRevenue = n.NominalGDP * 0.08 * (n.Fiscal.ExportCapacityLevel / 20.0)

	// Federal Fiscal updates: tariff revenue = imports * effective tariff
	tariffRev := n.Trade.AnnualImports * n.Trade.AverageTariffRate
	n.Fiscal.Tick(
		citizenTaxesPaid,
		corporateTaxesPaid,
		tariffRev,
		n.Fiscal.BorrowingYield,
		tickFractionOfYear,
	)

	// 5. National Income and Product Accounts: GDP = C + I + G + (X - M)
	// C: Personal Consumption Expenditures
	c := consumerSpending
	if c <= 0 {
		c = n.NominalGDP * 0.65
	}

	// I: Gross Private Domestic Investment
	i := corporateInvestment
	if i <= 0 {
		i = n.NominalGDP * 0.18
	}

	// G: Government Consumption & Gross Investment (mandatory + discretionary outlays excluding interest)
	g := n.Fiscal.MandatorySpending*0.60 + n.Fiscal.DiscretionarySpending

	// X - M: Net Exports
	netExports := n.Trade.TradeBalance

	computedGDP := c + i + g + netExports
	if computedGDP < 20_000_000_000_000.0 && n.NominalGDP > 10_000_000_000_000.0 {
		computedGDP = 20_000_000_000_000.0 // Baseline floor for large economy
	} else if computedGDP < 2_000_000_000.0 {
		computedGDP = 2_000_000_000.0 // Baseline floor for emerging economy
	}

	// Smooth GDP transition
	n.NominalGDP = 0.95*n.NominalGDP + 0.05*computedGDP

	// Real GDP growth = nominal change annualized minus inflation
	if n.prevGDP > 0 && tickFractionOfYear > 0 {
		nominalGrowth := (n.NominalGDP - n.prevGDP) / n.prevGDP / tickFractionOfYear
		realGrowth := nominalGrowth - n.CentralBank.CPIInflationRate
		// Dampen outliers to realistic bounds (-0.05 to +0.08)
		realGrowth = math.Max(-0.05, math.Min(0.08, realGrowth))
		n.RealGDPGrowth = 0.90*n.RealGDPGrowth + 0.10*realGrowth
	}
	n.prevGDP = n.NominalGDP

	var activePolicies []string
	for pID, active := range n.Fiscal.ActivePolicies {
		if active {
			activePolicies = append(activePolicies, pID)
		}
	}

	report := NationalReport{
		GDP:                   n.NominalGDP,
		RealGDPGrowth:         n.RealGDPGrowth,
		FedFundsRate:          n.CentralBank.FedFundsRate,
		TenYearYield:          n.CentralBank.TenYearYield,
		CPIInflationRate:      n.CentralBank.CPIInflationRate,
		PolicyStance:          n.CentralBank.PolicyStance,
		FOMCAnnouncement:      "",
		UnemploymentRate:      n.Labor.UnemploymentRate,
		LaborForce:            n.Labor.LaborForce,
		EmployedWorkers:       n.Labor.EmployedWorkers,
		AverageHourlyWage:     n.Labor.AverageHourlyWage,
		AnnualWageGrowth:      n.Labor.AnnualWageGrowth,
		NetMonthlyPayrolls:    n.Labor.NetMonthlyPayrolls,
		AnnualExports:         n.Trade.AnnualExports,
		AnnualImports:         n.Trade.AnnualImports,
		TradeBalance:          n.Trade.TradeBalance,
		DollarIndexDXY:        n.Trade.DollarIndexDXY,
		FederalRevenue:        n.Fiscal.TotalRevenue,
		FederalOutlays:        n.Fiscal.TotalOutlays,
		BudgetDeficit:         n.Fiscal.BudgetDeficit,
		NationalDebt:          n.Fiscal.NationalDebt,
		TreasuryCash:          n.Fiscal.TreasuryCash,
		CreditRating:          n.Fiscal.CreditRating,
		BorrowingYield:        n.Fiscal.BorrowingYield,
		InfrastructureLevel:   n.Fiscal.InfrastructureLevel,
		HealthcareLevel:       n.Fiscal.HealthcareLevel,
		EducationLevel:        n.Fiscal.EducationLevel,
		EnterpriseGrantsLevel: n.Fiscal.EnterpriseGrantsLevel,
		ExportCapacityLevel:   n.Fiscal.ExportCapacityLevel,
		FDIInflow:             n.Fiscal.FDIInflow,
		ExportRevenue:         n.Fiscal.ExportRevenue,
		ExchangeChartered:     n.Fiscal.ExchangeChartered,
		ActivePolicies:        activePolicies,
	}


	if rateChanged {
		report.FOMCAnnouncement = fomcAnnouncement
	}

	return report
}

// CompositeSectorDemand computes an aggregate demand multiplier for an industry sector
// combining consumer propensity (60%), foreign export demand (20%), and federal procurement (20%).
func (n *NationalEconomy) CompositeSectorDemand(sector string, consumerDemandFactor float64) float64 {
	exportFactor := n.Trade.ExportDemandFactor(sector)
	procurementFactor := n.Fiscal.SectorProcurementDemand(sector)

	composite := 0.60*consumerDemandFactor + 0.20*exportFactor + 0.20*procurementFactor
	return math.Max(0.50, math.Min(2.00, composite))
}
