package demographics

import (
	"math"
)

// Demographics models the population structure, workforce, and income
// of an American-style consumer base (~340 million citizens).
type Demographics struct {
	TotalPopulation           float64 // Total headcount (e.g. 340,000,000)
	WorkingAgePopulation      float64 // Age 18-64 (e.g. 212,000,000)
	LaborForceParticipation   float64 // Participation rate (e.g. 0.628 for 62.8%)
	EmployedCount             float64 // Currently working citizens
	MedianHouseholdIncome     float64 // Median annual household income (USD, e.g. $78,500)
	AverageHourlyEarnings     float64 // Average wage in USD/hour (e.g. $34.50)
	AggregateDisposableIncome float64 // Total after-tax disposable personal income in USD
	PersonalSavingsRate       float64 // Fraction of disposable income saved (e.g. 0.045 for 4.5%)
	PopulationGrowthRate      float64 // Annualized growth rate (e.g. 0.005 for 0.5%)
}

// NewDefaultDemographics returns calibrated baseline US demographics.
func NewDefaultDemographics() *Demographics {
	pop := 340_000_000.0
	workingAge := 268_000_000.0
	partRate := 0.628
	laborForce := workingAge * partRate // ~168.3M
	employed := laborForce * 0.961      // 3.9% natural initial unemployment (~161.7M)

	hourlyWage := 34.50
	annualWage := hourlyWage * 2000.0           // 2000 work hours/year
	disposable := employed * annualWage * 0.815 // ~18.5% effective tax rate

	return &Demographics{
		TotalPopulation:           pop,
		WorkingAgePopulation:      workingAge,
		LaborForceParticipation:   partRate,
		EmployedCount:             employed,
		MedianHouseholdIncome:     78_500.0,
		AverageHourlyEarnings:     hourlyWage,
		AggregateDisposableIncome: disposable,
		PersonalSavingsRate:       0.048,
		PopulationGrowthRate:      0.005,
	}
}

// LaborForce returns the active civilian labor force count.
func (d *Demographics) LaborForce() float64 {
	return d.WorkingAgePopulation * d.LaborForceParticipation
}

// UnemploymentRate returns the current civilian unemployment rate (U3).
func (d *Demographics) UnemploymentRate() float64 {
	lf := d.LaborForce()
	if lf <= 0 {
		return 0.0
	}
	unemployed := lf - d.EmployedCount
	if unemployed < 0 {
		unemployed = 0
	}
	return unemployed / lf
}

// Tick advances population growth, adjusts workforce to employment changes,
// and updates aggregate personal income.
func (d *Demographics) Tick(employmentCount float64, wageRate float64, effectiveTaxRate float64, tickFractionOfYear float64) {
	// 1. Natural population growth
	popGrowth := d.TotalPopulation * d.PopulationGrowthRate * tickFractionOfYear
	d.TotalPopulation += popGrowth
	d.WorkingAgePopulation += popGrowth * 0.62

	// 2. Sync employment count from labor market
	if employmentCount > 0 {
		d.EmployedCount = math.Min(employmentCount, d.LaborForce())
	}

	// 3. Update wage rates
	if wageRate > 0 {
		d.AverageHourlyEarnings = wageRate
	}

	// 4. Recompute aggregate gross and disposable income
	annualHours := 2000.0
	annualWagePerWorker := d.AverageHourlyEarnings * annualHours
	grossIncome := d.EmployedCount * annualWagePerWorker
	taxRate := math.Max(0.05, math.Min(0.40, effectiveTaxRate))
	d.AggregateDisposableIncome = grossIncome * (1.0 - taxRate)
}

// SetPublicFactors adjusts population growth rate and minimum wages based on national healthcare and education investments.
func (d *Demographics) SetPublicFactors(healthcareLevel, educationLevel float64) {
	// Baseline growth 0.5%/yr, surges up to 2.8%/yr with high healthcare and living standards
	d.PopulationGrowthRate = 0.005 + (healthcareLevel/100.0)*0.023
	// High education level raises average worker skill and wage baseline (from $15 up to $50/hr)
	targetWage := 15.0 + (educationLevel/100.0)*35.0
	if d.AverageHourlyEarnings < targetWage {
		d.AverageHourlyEarnings = targetWage
	}
}
