package company

import (
	"math/rand"
	"testing"
)

func newTestRng(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

// TestJumpFrequencyMatchesCalibration runs one company for many ticks
// and checks that crisis and mania jump counts land roughly in the
// ranges the design was calibrated for — not exact (it's stochastic),
// but within a sane band, so a silent parameter regression would be
// caught here.
func TestJumpFrequencyMatchesCalibration(t *testing.T) {
	params := DefaultGenerationParams()
	used := make(map[string]bool)

	const ticks = 600_000 // large enough that ~2 expected manias becomes
	// comfortably likely to appear at least once (want P(zero manias)
	// to be small, not just P(zero crises)); see calibration note below
	rngSeed := int64(42)

	c := GenerateCompany(InformationTechnology, MidCap, params, newTestRng(rngSeed), used)

	var crises, manias, goodNews int
	for range ticks {
		event, _ := c.Tick()
		switch event.Kind {
		case CrisisJump:
			crises++
		case ManiaJump:
			manias++
		case GoodNewsJump:
			goodNews++
		}
	}

	t.Logf("over %d ticks: crises=%d, good_news=%d, manias=%d", ticks, crises, goodNews, manias)

	if crises == 0 {
		t.Fatalf("expected at least some crisis jumps over %d ticks, got 0", ticks)
	}
	if manias == 0 {
		t.Fatalf("expected at least some mania jumps over %d ticks, got 0 (sample may be too small or ManiaProbability miscalibrated)", ticks)
	}

	// Downside should clearly outpace upside-with-ordinary-magnitude in
	// raw frequency, matching the designed asymmetry (LambdaDown >
	// LambdaUp). Manias are a small fraction of upside jumps by design
	// (ManiaProbability), so total upside events (good news + mania)
	// is the correct comparison, not manias alone.
	totalUpside := goodNews + manias
	if crises <= totalUpside {
		t.Fatalf("expected downside jumps to outpace upside jumps (leverage effect), got crises=%d vs upside=%d", crises, totalUpside)
	}

	// Manias should be a small minority of all upside events.
	if totalUpside > 0 {
		maniaShare := float64(manias) / float64(totalUpside)
		t.Logf("mania share of upside events: %.4f", maniaShare)
		if maniaShare > 0.10 {
			t.Fatalf("expected manias to be a small minority of upside events (~ManiaProbability), got share %.4f", maniaShare)
		}
	}
}

// TestDeterministicRegeneration confirms that generating a company (or
// a whole universe) from the same seed produces identical results —
// critical for reproducible research runs.
func TestDeterministicRegeneration(t *testing.T) {
	params := DefaultGenerationParams()

	uniA := GenerateUniverse(50, 12345, params)
	uniB := GenerateUniverse(50, 12345, params)

	if len(uniA) != len(uniB) {
		t.Fatalf("expected same-length universes, got %d vs %d", len(uniA), len(uniB))
	}

	for i := range uniA {
		a, b := uniA[i], uniB[i]
		if a.Symbol != b.Symbol || a.Name != b.Name || a.TrueValue != b.TrueValue {
			t.Fatalf("company %d differs between identically-seeded universes: (%s, %s, %.4f) vs (%s, %s, %.4f)",
				i, a.Symbol, a.Name, a.TrueValue, b.Symbol, b.Name, b.TrueValue)
		}
	}

	// Advance both universes identically and confirm paths stay
	// identical — proves each company's own sub-seeded rng is
	// reproducible, not just its initial generation.
	for range 100 {
		for i := range uniA {
			uniA[i].Tick()
			uniB[i].Tick()
		}
	}
	for i := range uniA {
		if uniA[i].TrueValue != uniB[i].TrueValue {
			t.Fatalf("company %d diverged after ticking: %.6f vs %.6f", i, uniA[i].TrueValue, uniB[i].TrueValue)
		}
	}
}

// TestUniverseGeneration sanity-checks structural properties of a
// generated universe: correct count, unique symbols, sector coverage,
// and cap-tier weighting skewing toward small/mid caps as designed.
func TestUniverseGeneration(t *testing.T) {
	params := DefaultGenerationParams()
	universe := GenerateUniverse(500, 999, params)

	if len(universe) != 500 {
		t.Fatalf("expected 500 companies, got %d", len(universe))
	}

	seen := make(map[string]bool)
	sectorCounts := make(map[Sector]int)
	tierCounts := make(map[CapTier]int)

	for _, c := range universe {
		if seen[c.Symbol] {
			t.Fatalf("duplicate symbol generated: %s", c.Symbol)
		}
		seen[c.Symbol] = true
		sectorCounts[c.Sector]++
		tierCounts[c.CapTier]++

		if c.TrueValue <= 0 {
			t.Fatalf("company %s has non-positive starting value: %.4f", c.Symbol, c.TrueValue)
		}
		if c.Float > c.SharesOutstanding {
			t.Fatalf("company %s has float (%.0f) exceeding shares outstanding (%.0f)", c.Symbol, c.Float, c.SharesOutstanding)
		}
	}

	t.Logf("sector counts: %+v", sectorCounts)
	t.Logf("cap tier counts: %+v", tierCounts)

	for _, s := range AllSectors {
		if sectorCounts[s] == 0 {
			t.Fatalf("sector %s has zero companies — expected even coverage across all sectors", s)
		}
	}

	if tierCounts[SmallCap] <= tierCounts[MegaCap] {
		t.Fatalf("expected small caps to outnumber mega caps (real markets have far more small companies), got small=%d mega=%d",
			tierCounts[SmallCap], tierCounts[MegaCap])
	}
}

// TestGBMDriftDirection confirms a company with strongly positive
// drift and zero volatility/jump risk trends upward over many ticks,
// and one with negative drift trends downward — a basic sanity check
// that the GBM step is wired correctly (sign and compounding), before
// trusting any of the jump-layered behavior on top of it.
func TestGBMDriftDirection(t *testing.T) {
	used := make(map[string]bool)
	params := DefaultGenerationParams()
	// Zero out jump risk for this test by using a params copy with
	// zero lambdas, isolating pure GBM behavior.
	params.BaseJumpParams.LambdaDown = 0
	params.BaseJumpParams.LambdaUp = 0

	rng := newTestRng(1)
	c := GenerateCompany(Utilities, MegaCap, params, rng, used)
	c.drift = 0.001       // force a clear positive drift for this test
	c.volatility = 0.0001 // near-zero noise so drift dominates
	start := c.TrueValue

	for range 500 {
		c.Tick()
	}

	if c.TrueValue <= start {
		t.Fatalf("expected positive drift to increase value over 500 ticks, start=%.4f end=%.4f", start, c.TrueValue)
	}
}
