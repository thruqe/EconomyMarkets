package trade

import "math"

// TradeEconomy models American foreign trade, import/export flows,
// the trade deficit, tariffs, and currency strength (Dollar Index / DXY).
type TradeEconomy struct {
	AnnualExports            float64            // Total goods and services exports in USD (~$3.1T)
	AnnualImports            float64            // Total goods and services imports in USD (~$3.9T)
	TradeBalance             float64            // Net Exports = Exports - Imports (typically negative ~$800B deficit)
	DollarIndexDXY           float64            // U.S. Dollar Index (baseline 103.5)
	AverageTariffRate        float64            // Average effective tariff on imports (e.g. 0.028 for 2.8%)
	SectorExportExposure     map[string]float64 // Sector -> Export demand intensity (e.g. Tech 0.40, Energy 0.35)
	SectorImportSensitivity  map[string]float64 // Sector -> Input import sensitivity
}

// NewDefaultTradeEconomy initializes baseline US global trade dynamics.
func NewDefaultTradeEconomy() *TradeEconomy {
	exp := 3_100_000_000_000.0
	imp := 3_900_000_000_000.0

	return &TradeEconomy{
		AnnualExports:     exp,
		AnnualImports:     imp,
		TradeBalance:      exp - imp,
		DollarIndexDXY:    103.5,
		AverageTariffRate: 0.028,
		SectorExportExposure: map[string]float64{
			"Information Technology": 1.35, // High global software/semiconductor exports
			"Industrials":            1.30, // Aerospace, defense, heavy equipment
			"Energy":                 1.25, // LNG, refined petroleum exports
			"Materials":              1.10, // Chemicals, specialty materials
			"Financials":             1.15, // Global investment banking, asset management
			"Health Care":            1.05, // Pharmaceuticals
			"Consumer Discretionary": 0.90, // Autos, entertainment
			"Consumer Staples":       0.85, // Agriculture, packaged goods
			"Communication Services": 1.10, // Digital media platforms
			"Utilities":              0.50, // Domestic power
			"Real Estate":            0.40, // Domestic land/buildings
		},
		SectorImportSensitivity: map[string]float64{
			"Consumer Discretionary": 1.30, // Foreign manufactured goods, apparel, electronics
			"Information Technology": 1.20, // Foreign chip foundry inputs
			"Industrials":            1.10, // Foreign component parts
			"Materials":              1.00, // Raw mineral imports
		},
	}
}

// Tick advances global trade based on interest rates, DXY currency fluctuations, and tariffs.
func (t *TradeEconomy) Tick(domesticInterestRate float64, tickFractionOfYear float64) {
	// 1. Interest rate parity: Higher US interest rates strengthen the US Dollar (DXY)
	// Baseline Fed rate = 4.5%
	rateGap := domesticInterestRate - 0.045
	targetDXY := 103.5 + rateGap*200.0 // +100bps rate hike = +2.0 DXY points
	targetDXY = math.Max(88.0, math.Min(125.0, targetDXY))
	t.DollarIndexDXY = 0.98*t.DollarIndexDXY + 0.02*targetDXY

	// 2. Currency impact on trade:
	// A stronger dollar (DXY > 103.5) makes US exports slightly more expensive abroad,
	// and makes foreign imports cheaper into the US.
	dxyFactor := t.DollarIndexDXY / 103.5

	baselineExports := 3_100_000_000_000.0
	baselineImports := 3_900_000_000_000.0

	t.AnnualExports = baselineExports * (1.0 / math.Pow(dxyFactor, 0.4))
	t.AnnualImports = baselineImports * math.Pow(dxyFactor, 0.3) * (1.0 - t.AverageTariffRate*0.5)

	t.TradeBalance = t.AnnualExports - t.AnnualImports
}

// ExportDemandFactor returns the export multiplier for a sector.
func (t *TradeEconomy) ExportDemandFactor(sector string) float64 {
	exposure, ok := t.SectorExportExposure[sector]
	if !ok {
		exposure = 1.0
	}
	// Stronger dollar slightly dampens export multiplier, weaker dollar boosts it
	currencyCompetitiveness := 103.5 / t.DollarIndexDXY
	return 1.0 + (exposure-1.0)*0.5 + (currencyCompetitiveness-1.0)*0.3
}
