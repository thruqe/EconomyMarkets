package labor

import "math"

// LaborMarket models the American employment landscape, payrolls, and wages.
type LaborMarket struct {
	LaborForce              float64 // Total civilian labor force (e.g. 168.3M)
	EmployedWorkers         float64 // Total employed headcount (e.g. 161.7M)
	UnemploymentRate        float64 // Civilian unemployment rate U3 (e.g. 0.039)
	NaturalUnemploymentRate float64 // Non-Accelerating Inflation Rate of Unemployment (NAIRU, e.g. 0.040)
	AverageHourlyWage       float64 // USD / hour (e.g. $34.50)
	AnnualWageGrowth        float64 // Annualized wage growth rate (e.g. 0.041 for 4.1%)
	JobOpenings             float64 // JOLTS total openings (e.g. 8,500,000)
	NetMonthlyPayrolls      float64 // Last monthly non-farm payrolls change (e.g. +195,000)
}

// NewDefaultLaborMarket initializes standard US employment conditions.
func NewDefaultLaborMarket() *LaborMarket {
	lf := 168_300_000.0
	unemp := 0.039
	employed := lf * (1.0 - unemp)

	return &LaborMarket{
		LaborForce:              lf,
		EmployedWorkers:         employed,
		UnemploymentRate:        unemp,
		NaturalUnemploymentRate: 0.040,
		AverageHourlyWage:       34.50,
		AnnualWageGrowth:        0.040,
		JobOpenings:             8_500_000.0,
		NetMonthlyPayrolls:      185_000.0,
	}
}

// Tick advances the labor market:
// - totalCorporateHeadcount: total employees hired across enterprise universe
// - inflationRate: current CPI inflation rate
// - tickFractionOfYear: fraction of a calendar year represented by this tick
func (l *LaborMarket) Tick(totalCorporateHeadcount float64, inflationRate float64, tickFractionOfYear float64) {
	// 1. Blend corporate headcount demand into national employment
	if totalCorporateHeadcount > 0 {
		// Scale corporate sample universe to macro national scale
		targetEmployed := math.Min(l.LaborForce*0.975, math.Max(l.LaborForce*0.88, totalCorporateHeadcount))
		delta := (targetEmployed - l.EmployedWorkers) * 0.05
		l.EmployedWorkers += delta
		l.NetMonthlyPayrolls = delta * (1.0 / math.Max(0.001, tickFractionOfYear*12.0))
	}

	// 2. Compute Unemployment Rate
	unemployed := l.LaborForce - l.EmployedWorkers
	if unemployed < 0 {
		unemployed = 0
	}
	l.UnemploymentRate = unemployed / l.LaborForce

	// 3. Phillips Curve Wage Dynamics:
	// When unemployment is below the natural rate (NAIRU=4.0%), labor scarcity accelerates wage growth.
	// When unemployment is high, wage growth slows down.
	unempGap := l.NaturalUnemploymentRate - l.UnemploymentRate
	targetWageGrowth := 0.035 + unempGap*0.75 + inflationRate*0.40
	targetWageGrowth = math.Max(0.01, math.Min(0.09, targetWageGrowth))
	l.AnnualWageGrowth = 0.95*l.AnnualWageGrowth + 0.05*targetWageGrowth

	// 4. Update wage rate
	l.AverageHourlyWage *= (1.0 + l.AnnualWageGrowth*tickFractionOfYear)

	// 5. Update Job Openings (JOLTS)
	l.JobOpenings = 5_000_000.0 + math.Max(0, (0.07-l.UnemploymentRate))*120_000_000.0
}
