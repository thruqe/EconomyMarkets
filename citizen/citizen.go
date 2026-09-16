package citizen

import (
	"economy/citizen/consumer"
	"economy/citizen/demographics"
	"economy/citizen/sentiment"
	"economy/citizen/taxation"
)

// CitizenReport captures the public economic telemetry of the citizen population.
type CitizenReport struct {
	Population            float64            // Total headcount (e.g. 340M)
	LaborForce            float64            // Civilian labor force
	EmployedCount         float64            // Working headcount
	UnemploymentRate      float64            // U3 rate
	AverageHourlyEarnings float64            // Average hourly wage (USD)
	DisposableIncome      float64            // Aggregate after-tax income (USD)
	ConsumerConfidence    float64            // 0-100 scale
	Happiness             float64            // 0-100 scale
	ConsumerSpending      float64            // Annualized PCE spending (USD)
	TaxesPaidThisTick     float64            // Federal/State taxes paid this tick
	SectorDemand          map[string]float64 // Sector -> demand multiplier
}

// CitizenEconomy coordinates all citizen demographic, psychological, and fiscal behavior.
type CitizenEconomy struct {
	Demographics *demographics.Demographics
	Sentiment    *sentiment.Sentiment
	Taxation     *taxation.Taxation
	Consumer     *consumer.ConsumerEngine
}

// NewDefaultCitizenEconomy constructs an American-style citizen ecosystem.
func NewDefaultCitizenEconomy() *CitizenEconomy {
	return &CitizenEconomy{
		Demographics: demographics.NewDefaultDemographics(),
		Sentiment:    sentiment.NewDefaultSentiment(),
		Taxation:     taxation.NewDefaultTaxation(),
		Consumer:     consumer.NewDefaultConsumerEngine(),
	}
}

// Tick advances the citizen population, sentiment, taxes, and consumer spending.
func (c *CitizenEconomy) Tick(employmentCount, wageRate, inflationRate, wageGrowth, marketReturn, tickFractionOfYear float64) CitizenReport {
	// 1. Advance demographics & calculate disposable income
	effectiveTax := c.Taxation.EffectiveIncomeTaxRate
	c.Demographics.Tick(employmentCount, wageRate, effectiveTax, tickFractionOfYear)

	// 2. Advance sentiment / happiness
	unemp := c.Demographics.UnemploymentRate()
	c.Sentiment.Tick(unemp, inflationRate, wageGrowth, marketReturn, effectiveTax)

	// 3. Compute consumer demand across sectors
	c.Consumer.Tick(
		c.Demographics.AggregateDisposableIncome,
		c.Sentiment.SpendingPropensity,
		c.Demographics.PersonalSavingsRate,
	)

	// 4. Compute taxes paid on earnings and consumption
	grossIncome := c.Demographics.EmployedCount * c.Demographics.AverageHourlyEarnings * 2000.0 * tickFractionOfYear
	incomeTax := c.Taxation.ComputeIncomeTax(grossIncome)
	salesTax := c.Taxation.ComputeSalesTax(c.Consumer.TotalConsumerSpending * tickFractionOfYear)
	c.Taxation.RecordTaxes(incomeTax, 0, salesTax)

	// Clone sector demand map for the report
	sectorDemand := make(map[string]float64, len(c.Consumer.SectorDemandFactors))
	for k, v := range c.Consumer.SectorDemandFactors {
		sectorDemand[k] = v
	}

	return CitizenReport{
		Population:            c.Demographics.TotalPopulation,
		LaborForce:            c.Demographics.LaborForce(),
		EmployedCount:         c.Demographics.EmployedCount,
		UnemploymentRate:      unemp,
		AverageHourlyEarnings: c.Demographics.AverageHourlyEarnings,
		DisposableIncome:      c.Demographics.AggregateDisposableIncome,
		ConsumerConfidence:    c.Sentiment.Index,
		Happiness:             c.Sentiment.Happiness,
		ConsumerSpending:      c.Consumer.TotalConsumerSpending,
		TaxesPaidThisTick:     incomeTax + salesTax,
		SectorDemand:          sectorDemand,
	}
}

// SetPublicFactors routes national health and education levels to citizen demographics.
func (c *CitizenEconomy) SetPublicFactors(healthcareLevel, educationLevel float64) {
	if c.Demographics != nil {
		c.Demographics.SetPublicFactors(healthcareLevel, educationLevel)
	}
}
