package centralbank

import "math"

// CentralBank models the Federal Reserve system: the dual mandate
// of price stability (2.0% CPI target) and maximum sustainable employment (~4.0% NAIRU).
type CentralBank struct {
	FedFundsRate        float64 // Benchmark policy interest rate (e.g. 0.0475 for 4.75%)
	TenYearYield        float64 // 10-Year U.S. Treasury benchmark yield (e.g. 0.0425 for 4.25%)
	CPIInflationRate    float64 // Headline Consumer Price Index YoY inflation (e.g. 0.024 for 2.4%)
	InflationTarget     float64 // Statutory target inflation (0.020 for 2.0%)
	NaturalUnemployment float64 // Natural unemployment rate NAIRU (0.040 for 4.0%)
	PolicyStance        string  // "Hawkish", "Neutral", "Dovish"
	MeetingTicksRemain  int     // Ticks until next scheduled FOMC policy meeting
	LastDecision        string  // Summary of last rate decision
}

// NewDefaultCentralBank initializes the Federal Reserve at baseline US settings.
func NewDefaultCentralBank() *CentralBank {
	return &CentralBank{
		FedFundsRate:        0.0450, // 4.50%
		TenYearYield:        0.0420, // 4.20%
		CPIInflationRate:    0.0240, // 2.40%
		InflationTarget:     0.0200, // 2.00%
		NaturalUnemployment: 0.0400, // 4.00%
		PolicyStance:        "Neutral",
		MeetingTicksRemain:  150,
		LastDecision:        "FOMC maintained target rate unchanged at 4.50%",
	}
}

// TaylorRuleRate computes the textbook macroeconomic policy target rate.
// r = neutral_rate + inflation + 0.5*(inflation - target) - 0.5*(unemployment - natural_rate)
func (cb *CentralBank) TaylorRuleRate(unemploymentRate float64) float64 {
	neutralRate := 0.025 // 2.5% neutral real + inflation
	inflationGap := cb.CPIInflationRate - cb.InflationTarget
	unempGap := unemploymentRate - cb.NaturalUnemployment

	taylor := neutralRate + cb.CPIInflationRate + 0.5*inflationGap - 0.5*unempGap
	return math.Max(0.0025, math.Min(0.090, taylor)) // 0.25% floor to 9.0% ceiling
}

// Tick advances inflation dynamics and executes FOMC policy meeting decisions:
// Returns true and an announcement string if an FOMC policy rate change occurred.
func (cb *CentralBank) Tick(unemploymentRate, wageGrowth float64, tickFractionOfYear float64) (rateChanged bool, announcement string) {
	// 1. Inflation updates: driven by wage growth pressure, demand, and monetary tightness
	// High interest rates put downward pressure on inflation; wage growth puts upward pressure.
	wagePressure := (wageGrowth - 0.035) * 0.40
	interestDrag := -(cb.FedFundsRate - 0.025) * 0.15
	targetInflation := cb.InflationTarget + wagePressure + interestDrag
	targetInflation = math.Max(0.005, math.Min(0.085, targetInflation))

	cb.CPIInflationRate = 0.98*cb.CPIInflationRate + 0.02*targetInflation

	// 2. 10-Year Treasury Yield tracks Fed Funds with a term premium
	target10Y := cb.FedFundsRate*0.85 + 0.0080
	cb.TenYearYield = 0.95*cb.TenYearYield + 0.05*target10Y

	// 3. FOMC Meeting Cycle
	cb.MeetingTicksRemain--
	if cb.MeetingTicksRemain > 0 {
		return false, ""
	}

	// FOMC Meeting fires (every ~200 ticks = ~1 month of simulated time)
	cb.MeetingTicksRemain = 200
	taylor := cb.TaylorRuleRate(unemploymentRate)
	rateDiff := taylor - cb.FedFundsRate

	// Policy Stance assessment
	if cb.CPIInflationRate > 0.030 {
		cb.PolicyStance = "Hawkish"
	} else if unemploymentRate > 0.055 {
		cb.PolicyStance = "Dovish"
	} else {
		cb.PolicyStance = "Neutral"
	}

	// Move rates in discrete 25 bps (0.0025) steps
	if rateDiff >= 0.0025 {
		// Rate Hike
		hike := 0.0025
		if rateDiff >= 0.0060 {
			hike = 0.0050 // Jumbo 50bps hike
		}
		cb.FedFundsRate += hike
		announcement = "FOMC raised federal funds rate by " + bpsString(hike) + " to " + percentString(cb.FedFundsRate)
		cb.LastDecision = announcement
		return true, announcement
	} else if rateDiff <= -0.0025 {
		// Rate Cut
		cut := 0.0025
		if rateDiff <= -0.0060 {
			cut = 0.0050 // Jumbo 50bps cut
		}
		cb.FedFundsRate -= cut
		announcement = "FOMC lowered federal funds rate by " + bpsString(cut) + " to " + percentString(cb.FedFundsRate)
		cb.LastDecision = announcement
		return true, announcement
	}

	cb.LastDecision = "FOMC held federal funds rate steady at " + percentString(cb.FedFundsRate)
	return false, ""
}

func bpsString(rate float64) string {
	bps := int(math.Round(rate * 10000))
	return string(rune('0'+bps/100)) + "0bps"
}

func percentString(rate float64) string {
	pct := rate * 100
	d1 := int(pct)
	d2 := int((pct - float64(d1)) * 100)
	buf := []byte{
		byte('0' + d1/10),
		byte('0' + d1%10),
		'.',
		byte('0' + d2/10),
		byte('0' + d2%10),
		'%',
	}
	if buf[0] == '0' {
		buf = buf[1:]
	}
	return string(buf)
}
