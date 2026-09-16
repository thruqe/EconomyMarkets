package sentiment

import (
	"math"
)

// Sentiment models citizen happiness, consumer confidence, and their
// willingness to spend vs save in the economy.
type Sentiment struct {
	// Index is on a 0 to 100 scale (modeled on the University of Michigan
	// Consumer Sentiment Index). 85.0 is historical baseline prosperity.
	Index float64

	// Happiness is on a 0 to 100 scale, reflecting overall social well-being.
	Happiness float64

	// SpendingPropensity scales aggregate consumer discretionary spending.
	// > 1.0 means citizens are optimistic and spending freely; < 1.0 means
	// precautionary saving and tightening belts.
	SpendingPropensity float64

	// TrailingMarketReturn tracks the stock market wealth effect.
	TrailingMarketReturn float64
}

// NewDefaultSentiment initializes sentiment at prosperous American baseline levels.
func NewDefaultSentiment() *Sentiment {
	return &Sentiment{
		Index:              85.0,
		Happiness:          82.0,
		SpendingPropensity: 1.0,
	}
}

// Tick updates consumer confidence and citizen happiness based on real-time macro factors.
// - unemploymentRate: current U3 unemployment (e.g. 0.039 for 3.9%)
// - inflationRate: annualized CPI rate (e.g. 0.024 for 2.4%)
// - wageGrowth: annualized wage growth (e.g. 0.041 for 4.1%)
// - marketReturn: trailing stock market return (e.g. +0.08 for +8%)
// - taxBurden: effective personal tax rate (e.g. 0.185)
func (s *Sentiment) Tick(unemploymentRate, inflationRate, wageGrowth, marketReturn, taxBurden float64) {
	s.TrailingMarketReturn = 0.9*s.TrailingMarketReturn + 0.1*marketReturn

	// 1. Real wage growth component (+ if wages outpace inflation, - if inflation eats wages)
	realWageGrowth := wageGrowth - inflationRate
	wageComponent := realWageGrowth * 300.0 // +1% real wage = +3 index points

	// 2. Unemployment penalty (baseline natural unemployment = 4.0%)
	unempGap := unemploymentRate - 0.040
	unempComponent := -unempGap * 400.0 // +1% unemployment = -4 index points

	// 3. Stock market wealth effect
	wealthComponent := s.TrailingMarketReturn * 40.0 // +10% market gain = +4 index points

	// 4. Inflation shock penalty (inflation above 2% hurts sentiment exponentially)
	inflationGap := inflationRate - 0.020
	inflationPenalty := 0.0
	if inflationGap > 0 {
		inflationPenalty = -inflationGap * 250.0
	}

	// 5. Tax burden penalty (baseline 18.5%)
	taxGap := taxBurden - 0.185
	taxPenalty := -taxGap * 100.0

	// Target sentiment index
	targetIndex := 85.0 + wageComponent + unempComponent + wealthComponent + inflationPenalty + taxPenalty
	targetIndex = math.Max(20.0, math.Min(120.0, targetIndex))

	// Smooth adjustment (sentiment is sticky, adjusts toward target with inertia)
	s.Index = 0.95*s.Index + 0.05*targetIndex

	// Citizen Happiness: highly correlated with sentiment, but also weighted by job security and low inflation
	targetHappiness := s.Index*0.80 + (1.0-unemploymentRate)*15.0 - math.Max(0, inflationRate-0.02)*100.0
	targetHappiness = math.Max(15.0, math.Min(98.0, targetHappiness))
	s.Happiness = 0.95*s.Happiness + 0.05*targetHappiness

	// Spending Propensity: index of 85 = 1.0x; 105 = 1.25x; 55 = 0.70x
	s.SpendingPropensity = 0.70 + (s.Index-50.0)*(0.50/50.0)
	s.SpendingPropensity = math.Max(0.60, math.Min(1.40, s.SpendingPropensity))
}
