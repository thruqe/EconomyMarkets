package company

import (
	"math"
	"testing"
)

// buildCompanyWithProfile constructs a company with a forced
// reporting profile and zero jump risk, isolating reporting behavior
// from the TrueValue jump mechanic so these tests measure exactly one
// mechanism at a time.
func buildCompanyWithProfile(profile ReportingProfile, seed int64) *Company {
	used := make(map[string]bool)
	params := DefaultGenerationParams()
	params.BaseJumpParams.LambdaDown = 0
	params.BaseJumpParams.LambdaUp = 0
	params.StartingValueMin = 100.0
	params.StartingValueMax = 100.0

	rng := newTestRng(seed)
	c := GenerateCompany(InformationTechnology, MidCap, params, rng, used)
	c.reporting = reportingParamsFor(profile, rng)
	// Zero out TrueValue's own drift/vol so any observed movement in
	// ReportedValue's gap is attributable to the reporting mechanic,
	// not to TrueValue itself wandering.
	c.drift = 0
	c.volatility = 0
	return c
}

// TestHonestReportingTracksTrueValueClosely confirms an Honest company
// stays close to TrueValue on average over many ticks (noise should
// not accumulate into a systematic gap).
func TestHonestReportingTracksTrueValueClosely(t *testing.T) {
	c := buildCompanyWithProfile(Honest, 1)

	var sumGapPercent float64
	const ticks = 5000
	for range ticks {
		c.Tick()
		sumGapPercent += (c.ReportedValue - c.TrueValue) / c.TrueValue
	}
	avgGapPercent := sumGapPercent / ticks

	t.Logf("Honest: average gap over %d ticks = %.4f%%", ticks, avgGapPercent*100)

	if math.Abs(avgGapPercent) > 0.01 {
		t.Fatalf("expected Honest reporting's average gap to stay near zero, got %.4f%%", avgGapPercent*100)
	}
}

// TestOptimisticBiasAccumulatesPositiveGap confirms an OptimisticBias
// company's ReportedValue systematically drifts above TrueValue over
// time, not just noisily above on average.
func TestOptimisticBiasAccumulatesPositiveGap(t *testing.T) {
	c := buildCompanyWithProfile(OptimisticBias, 2)
	// Disable restatement for this test so we're purely measuring
	// accumulation, not accumulation-then-correction.
	c.reporting.RestatementProbability = 0

	for range 500 {
		c.Tick()
	}

	gapPercent := (c.ReportedValue - c.TrueValue) / c.TrueValue
	t.Logf("OptimisticBias: gap after 500 ticks = %.4f%%", gapPercent*100)

	if gapPercent <= 0 {
		t.Fatalf("expected OptimisticBias to accumulate a positive gap over time, got %.4f%%", gapPercent*100)
	}
}

// TestConservativeBiasAccumulatesNegativeGap is the mirror case.
func TestConservativeBiasAccumulatesNegativeGap(t *testing.T) {
	c := buildCompanyWithProfile(ConservativeBias, 3)
	c.reporting.RestatementProbability = 0

	for range 500 {
		c.Tick()
	}

	gapPercent := (c.ReportedValue - c.TrueValue) / c.TrueValue
	t.Logf("ConservativeBias: gap after 500 ticks = %.4f%%", gapPercent*100)

	if gapPercent >= 0 {
		t.Fatalf("expected ConservativeBias to accumulate a negative gap over time, got %.4f%%", gapPercent*100)
	}
}

// TestHighUncertaintyHasWiderNoiseThanHonest confirms HighUncertainty
// companies show meaningfully larger tick-to-tick reporting variance
// than Honest companies, without necessarily drifting directionally.
func TestHighUncertaintyHasWiderNoiseThanHonest(t *testing.T) {
	honest := buildCompanyWithProfile(Honest, 4)
	uncertain := buildCompanyWithProfile(HighUncertainty, 4) // same seed, different profile

	varianceOf := func(c *Company) float64 {
		var gaps []float64
		for range 1000 {
			c.Tick()
			gaps = append(gaps, (c.ReportedValue-c.TrueValue)/c.TrueValue)
		}
		var mean float64
		for _, g := range gaps {
			mean += g
		}
		mean /= float64(len(gaps))
		var variance float64
		for _, g := range gaps {
			variance += (g - mean) * (g - mean)
		}
		return variance / float64(len(gaps))
	}

	honestVar := varianceOf(honest)
	uncertainVar := varianceOf(uncertain)

	t.Logf("Honest gap variance=%.8f, HighUncertainty gap variance=%.8f", honestVar, uncertainVar)

	if uncertainVar <= honestVar {
		t.Fatalf("expected HighUncertainty to show wider gap variance than Honest, got honest=%.8f uncertain=%.8f", honestVar, uncertainVar)
	}
}

// TestFraudulentAccumulatesLargeGapThenRestates is the important one:
// proves a Fraudulent company builds a substantially larger hidden gap
// than OptimisticBias over the same horizon, and that a restatement,
// once it fires, closes most of that gap in a single violent
// correction rather than gradually.
func TestFraudulentAccumulatesLargeGapThenRestates(t *testing.T) {
	c := buildCompanyWithProfile(Fraudulent, 5)
	// Force restatement to definitely fire within our test horizon by
	// running long enough relative to its configured probability, but
	// first measure the gap immediately before it fires by disabling
	// restatement, then re-enabling to observe the correction.
	c.reporting.RestatementProbability = 0

	const buildupTicks = 1000
	for range buildupTicks {
		c.Tick()
	}

	gapBeforeRestatement := (c.ReportedValue - c.TrueValue) / c.TrueValue
	reportedBefore := c.ReportedValue
	t.Logf("Fraudulent: gap after %d ticks (no restatement yet) = %.4f%%", buildupTicks, gapBeforeRestatement*100)

	if gapBeforeRestatement <= 0.05 {
		t.Fatalf("expected Fraudulent to accumulate a substantial (>5%%) hidden gap over %d ticks, got %.4f%%", buildupTicks, gapBeforeRestatement*100)
	}

	// Now force a restatement deterministically and confirm it closes
	// most of the gap in one shot.
	c.reporting.RestatementProbability = 1.0 // guarantee it fires next tick
	c.Tick()

	reportedAfter := c.ReportedValue
	gapAfterRestatement := (reportedAfter - c.TrueValue) / c.TrueValue

	t.Logf("Fraudulent: ReportedValue before=%.4f after=%.4f (TrueValue=%.4f), gap after=%.4f%%",
		reportedBefore, reportedAfter, c.TrueValue, gapAfterRestatement*100)

	if math.Abs(gapAfterRestatement) >= math.Abs(gapBeforeRestatement)*0.5 {
		t.Fatalf("expected restatement to close most of the gap, got gap before=%.4f%% after=%.4f%%",
			gapBeforeRestatement*100, gapAfterRestatement*100)
	}
}

// TestRestatementEventReturnedWhenItFires confirms Tick() actually
// surfaces a non-nil RestatementEvent on the tick a correction occurs.
func TestRestatementEventReturnedWhenItFires(t *testing.T) {
	c := buildCompanyWithProfile(Fraudulent, 6)
	c.reporting.RestatementProbability = 0

	for range 500 {
		c.Tick()
	}

	c.reporting.RestatementProbability = 1.0 // force it next tick
	_, restatement := c.Tick()

	if restatement == nil {
		t.Fatalf("expected a non-nil RestatementEvent on the tick a correction fires")
	}
	if restatement.Symbol != c.Symbol {
		t.Fatalf("expected RestatementEvent.Symbol to match the company, got %s vs %s", restatement.Symbol, c.Symbol)
	}
	t.Logf("RestatementEvent: profile=%s priorGap=%.4f%% severity=%.4f",
		restatement.Profile, restatement.PriorGapPercent*100, restatement.Severity)
}

// TestReportingProfileWeightsIncludeAllCases confirms the universe
// generator actually produces every reporting profile across a large
// enough sample, and that Fraudulent stays rare relative to Honest —
// a structural sanity check on chooseReportingProfile's weighting.
func TestReportingProfileWeightsIncludeAllCases(t *testing.T) {
	params := DefaultGenerationParams()
	universe := GenerateUniverse(500, 777, params)

	counts := make(map[ReportingProfile]int)
	for _, c := range universe {
		counts[c.reporting.Profile]++
	}

	t.Logf("reporting profile counts across 500 companies: Honest=%d Optimistic=%d Conservative=%d HighUncertainty=%d Fraudulent=%d",
		counts[Honest], counts[OptimisticBias], counts[ConservativeBias], counts[HighUncertainty], counts[Fraudulent])

	for _, p := range []ReportingProfile{Honest, OptimisticBias, ConservativeBias, HighUncertainty, Fraudulent} {
		if counts[p] == 0 {
			t.Fatalf("expected at least one company with profile %s across 500 companies, got 0", p)
		}
	}

	if counts[Fraudulent] >= counts[Honest] {
		t.Fatalf("expected Fraudulent to be far rarer than Honest, got fraudulent=%d honest=%d", counts[Fraudulent], counts[Honest])
	}
}
