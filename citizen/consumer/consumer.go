package consumer

import "math"

// ConsumerEngine translates aggregate disposable income and citizen sentiment
// into sector-level consumption demand factors that drive corporate revenue.
type ConsumerEngine struct {
	TotalConsumerSpending float64            // Total annual personal consumption expenditures (PCE) in USD (~$19 Trillion)
	SectorDemandFactors   map[string]float64 // Sector name -> demand multiplier (1.0 = baseline)
}

// NewDefaultConsumerEngine initializes consumer demand at US macroeconomic baseline.
func NewDefaultConsumerEngine() *ConsumerEngine {
	factors := map[string]float64{
		"Information Technology": 1.0,
		"Health Care":            1.0,
		"Financials":             1.0,
		"Consumer Discretionary": 1.0,
		"Consumer Staples":       1.0,
		"Industrials":            1.0,
		"Energy":                 1.0,
		"Materials":              1.0,
		"Communication Services": 1.0,
		"Utilities":              1.0,
		"Real Estate":            1.0,
	}

	return &ConsumerEngine{
		TotalConsumerSpending: 18_500_000_000_000.0, // ~$18.5T US PCE baseline
		SectorDemandFactors:   factors,
	}
}

// Tick calculates sector demand multipliers based on disposable income and spending propensity.
// - disposableIncome: total aggregate after-tax income
// - spendingPropensity: sentiment multiplier (>1 optimistic, <1 conservative)
// - savingsRate: fraction diverted to personal savings
func (c *ConsumerEngine) Tick(disposableIncome, spendingPropensity, savingsRate float64) {
	effectiveSavings := math.Max(0.02, math.Min(0.15, savingsRate/spendingPropensity))
	c.TotalConsumerSpending = disposableIncome * (1.0 - effectiveSavings) * spendingPropensity

	// Elasticity profiles:
	// Highly elastic sectors boom when propensity > 1.0 and pull back when propensity < 1.0.
	// Inelastic sectors remain stable regardless of sentiment.
	propensityDelta := spendingPropensity - 1.0

	c.SectorDemandFactors["Consumer Staples"] = 1.0 + propensityDelta*0.15      // Inelastic (food, household)
	c.SectorDemandFactors["Health Care"] = 1.0 + propensityDelta*0.10            // Highly inelastic (medicine, care)
	c.SectorDemandFactors["Utilities"] = 1.0 + propensityDelta*0.08              // Inelastic (power, water)
	c.SectorDemandFactors["Energy"] = 1.0 + propensityDelta*0.35                 // Moderate (gasoline, heat)
	c.SectorDemandFactors["Communication Services"] = 1.0 + propensityDelta*0.40 // Telecom, media
	c.SectorDemandFactors["Financials"] = 1.0 + propensityDelta*0.50             // Credit, banking
	c.SectorDemandFactors["Materials"] = 1.0 + propensityDelta*0.60              // Raw goods
	c.SectorDemandFactors["Industrials"] = 1.0 + propensityDelta*0.75            // Capital goods
	c.SectorDemandFactors["Real Estate"] = 1.0 + propensityDelta*0.85            // Housing demand
	c.SectorDemandFactors["Information Technology"] = 1.0 + propensityDelta*1.10 // High elasticity (tech upgrades)
	c.SectorDemandFactors["Consumer Discretionary"] = 1.0 + propensityDelta*1.40 // Extreme elasticity (luxury, travel)

	// Clamp to realistic bounds [0.50, 1.80]
	for k, v := range c.SectorDemandFactors {
		c.SectorDemandFactors[k] = math.Max(0.50, math.Min(1.80, v))
	}
}

// DemandForSector returns the consumption demand multiplier for a sector.
func (c *ConsumerEngine) DemandForSector(sector string) float64 {
	if factor, ok := c.SectorDemandFactors[sector]; ok {
		return factor
	}
	return 1.0
}
