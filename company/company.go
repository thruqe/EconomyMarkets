package company

import "math/rand"

// FundamentalEvent is the public record of a single tick's jump
// outcome — the moment "the market learns something" about a
// company's fundamentals, as opposed to the ordinary, silent GBM drift
// that happens every tick regardless. This is what lets any agent
// (not just institutional agents reading TrueValue directly) react to
// news as a discrete, legible signal — much closer to how real retail
// participants actually experience fundamentals: as headline-style
// events, not a continuously-observed intrinsic-value number.
type FundamentalEvent struct {
	Symbol     string
	Tick       int
	Kind       JumpKind
	Multiplier float64 // the multiplier applied to TrueValue this tick
}

// RestatementEvent is the public record of a correction to
// ReportedValue — the moment previously reported information is
// revised, whether through a gentle "better information now
// available" correction (Honest/HighUncertainty profiles) or a
// violent fraud discovery (Fraudulent profile). This is distinct from
// FundamentalEvent: a restatement changes what was *reported*, not
// what was *true* — TrueValue is untouched by a restatement, only the
// gap between ReportedValue and TrueValue closes (partially or
// mostly, depending on severity).
type RestatementEvent struct {
	Symbol          string
	Tick            int
	Profile         ReportingProfile
	PriorGapPercent float64 // (ReportedValue-TrueValue)/TrueValue immediately before correction
	Severity        float64 // fraction of the gap closed by this correction
}

// Company is a single simulated company's fundamental data. It has no
// notion of market price, trading, or order books — those live
// entirely in package market. Company's only job is to generate an
// independent "true value" signal and evolve it tick by tick, which is
// what gives hedge-fund-style agents an actual reason to trade: they
// compare this value against the market's price, not against each
// other.
type Company struct {
	Symbol  string
	Name    string
	Sector  Sector
	CapTier CapTier

	// TrueValue is the current fundamental value per share. Evolves
	// each tick via Tick() — GBM drift plus the (rare) chance of a
	// jump event. Nothing in the market ever observes this directly
	// in a fully realistic setup — see ReportedValue.
	TrueValue float64

	// ReportedValue is what the company actually discloses — what a
	// fundamentals-driven trader (institutional or retail) would
	// realistically read. Usually close to TrueValue, but can diverge
	// according to ReportingProfile: a small persistent bias, wide
	// honest uncertainty, or (rarely) a large hidden gap that later
	// corrects violently. Existing institutional agents (HedgeFund,
	// Bank) currently still read TrueValue directly for simplicity —
	// a known simplification, not a claim that this is realistic;
	// switching them to ReportedValue is a natural, separate follow-up.
	ReportedValue float64

	SharesOutstanding float64
	Float             float64

	drift      float64
	volatility float64
	jumpParams JumpParams
	rng        *rand.Rand

	reporting ReportingParams

	// reportingBiasAccum tracks the persistent, compounding component
	// of the gap between ReportedValue and TrueValue, as a fraction of
	// TrueValue — separate from ReportedValue itself, which also
	// carries fresh per-tick noise on top. This is what lets
	// Optimistic/Conservative/Fraudulent bias genuinely accumulate
	// tick over tick rather than being overwritten by the next tick's
	// noise draw (ReportedValue is recomputed from TrueValue plus
	// noise plus this accumulator every tick, not carried forward
	// directly).
	reportingBiasAccum float64

	// currentTick tracks how many times Tick has been called, purely
	// so emitted events can carry a tick number without the caller
	// having to pass one in — Company already knows its own position
	// in time.
	currentTick int
}

// Tick advances TrueValue and ReportedValue by exactly one time step.
// TrueValue evolves via ordinary GBM drift, or a jump event (crisis,
// good-news, or mania) that overrides the ordinary step — at most one
// jump per tick. Independently, ReportedValue evolves according to the
// company's ReportingProfile: ordinary noise plus any persistent bias
// each tick, and a (usually rare) chance of a restatement event that
// corrects some fraction of the accumulated gap back toward TrueValue.
//
// Returns the tick's FundamentalEvent (Kind is NoJump on ordinary
// ticks) and, if one fired, a non-nil *RestatementEvent — the two are
// independent and can both occur, both occur separately, or neither
// occur on any given tick, since they represent orthogonal mechanisms
// (real operational shocks vs. reporting-integrity corrections).
func (c *Company) Tick() (FundamentalEvent, *RestatementEvent) {
	c.currentTick++

	jump := rollJump(c.jumpParams, c.rng)
	event := FundamentalEvent{Symbol: c.Symbol, Tick: c.currentTick, Kind: jump.Kind, Multiplier: jump.Multiplier}

	if jump.Kind != NoJump {
		// A jump tick applies its multiplier directly rather than
		// blending with the ordinary GBM step — a discrete shock is
		// meant to dominate the tick it occurs on, not be diluted by
		// simultaneous ordinary drift.
		c.TrueValue *= jump.Multiplier
		if c.TrueValue < 0.01 {
			c.TrueValue = 0.01 // floor: a company's fundamental value
			// shouldn't hit exactly zero or go negative in this model;
			// a near-zero floor represents effective insolvency
			// without dividing-by-zero or sign-flip issues downstream.
		}
	} else {
		// Ordinary GBM step: dV/V = drift*dt + vol*dW, with dt
		// implicitly 1 tick. Using a standard normal draw scaled by
		// volatility for the Wiener increment.
		z := c.rng.NormFloat64()
		c.TrueValue *= 1 + c.drift + c.volatility*z
		if c.TrueValue < 0.01 {
			c.TrueValue = 0.01
		}
	}

	restatement := c.tickReporting()

	return event, restatement
}

// tickReporting advances ReportedValue by one step: the persistent
// bias accumulator (see reportingBiasAccum) compounds first, then
// fresh per-tick noise is layered on top to produce this tick's
// ReportedValue — so bias genuinely persists tick over tick instead
// of being reconstructed from (and overwritten by) ReportedValue
// itself. A restatement check may then correct some fraction of the
// accumulated bias back toward zero. Returns a non-nil
// *RestatementEvent if a correction fired this tick.
func (c *Company) tickReporting() *RestatementEvent {
	// Compound the persistent bias into the accumulator first — this
	// is the actual state that survives across ticks. Honest/
	// HighUncertainty have PersistentBias == 0, so their accumulator
	// never moves and their gap stays purely noise-driven.
	c.reportingBiasAccum += c.reporting.PersistentBias

	// This tick's ReportedValue: TrueValue adjusted by the
	// accumulated bias plus fresh, non-persistent noise.
	noise := c.rng.NormFloat64() * c.reporting.NoiseStdDev
	c.ReportedValue = c.TrueValue * (1 + c.reportingBiasAccum + noise)

	if c.ReportedValue < 0.01 {
		c.ReportedValue = 0.01
	}

	if c.rng.Float64() >= c.reporting.RestatementProbability {
		return nil
	}

	// Restatement fires: close some fraction of the *accumulated
	// bias* (not the noisy instantaneous gap, which would make the
	// correction size dependent on this tick's random noise draw
	// rather than the genuine underlying misreporting being
	// corrected).
	if c.reportingBiasAccum == 0 {
		return nil // nothing systematic to correct
	}
	priorGapPercent := c.reportingBiasAccum

	severity := c.reporting.RestatementSeverityMin +
		c.rng.Float64()*(c.reporting.RestatementSeverityMax-c.reporting.RestatementSeverityMin)

	c.reportingBiasAccum -= c.reportingBiasAccum * severity

	// Recompute ReportedValue immediately under the corrected
	// accumulator (keeping this tick's already-drawn noise), so the
	// correction is reflected the same tick it fires rather than
	// waiting a tick to show up.
	c.ReportedValue = c.TrueValue * (1 + c.reportingBiasAccum + noise)
	if c.ReportedValue < 0.01 {
		c.ReportedValue = 0.01
	}

	return &RestatementEvent{
		Symbol:          c.Symbol,
		Tick:            c.currentTick,
		Profile:         c.reporting.Profile,
		PriorGapPercent: priorGapPercent,
		Severity:        severity,
	}
}

// GenerationParams bundles the tunable ranges used when generating a
// single company, separated from GenerateUniverse's signature so
// individual companies can also be generated/inspected in isolation
// with the same knobs.
type GenerationParams struct {
	StartingValueMin, StartingValueMax float64
	BaseJumpParams                     JumpParams
}

// DefaultGenerationParams returns reasonable starting-value bounds and
// the default jump parameters described in JumpParams' documentation.
func DefaultGenerationParams() GenerationParams {
	return GenerationParams{
		StartingValueMin: 5.0,
		StartingValueMax: 500.0,
		BaseJumpParams:   DefaultJumpParams(),
	}
}

// GenerateCompany produces a single company from the given sector,
// cap tier, and rng. rng should be a dedicated sub-generator (see
// GenerateUniverse) rather than a shared/global source, so individual
// companies can be regenerated or inspected deterministically in
// isolation.
func GenerateCompany(sector Sector, tier CapTier, params GenerationParams, rng *rand.Rand, usedSymbols map[string]bool) *Company {
	sp := sectorProfiles[sector]
	cp := capTierProfiles[tier]

	drift := sp.DriftMin + rng.Float64()*(sp.DriftMax-sp.DriftMin)
	vol := (sp.VolMin + rng.Float64()*(sp.VolMax-sp.VolMin)) * cp.VolMultiplier

	jumpParams := params.BaseJumpParams.Scaled(sp.JumpRiskMultiplier, cp.JumpMultiplier)

	startingValue := params.StartingValueMin + rng.Float64()*(params.StartingValueMax-params.StartingValueMin)

	shares := cp.SharesOutstandingMin + rng.Float64()*(cp.SharesOutstandingMax-cp.SharesOutstandingMin)
	floatFraction := cp.FloatFractionMin + rng.Float64()*(cp.FloatFractionMax-cp.FloatFractionMin)

	name := GenerateName(sector, rng)
	symbol := GenerateSymbol(name, usedSymbols)

	reportingProfile := chooseReportingProfile(sector, rng)
	reportingParams := reportingParamsFor(reportingProfile, rng)

	return &Company{
		Symbol:            symbol,
		Name:              name,
		Sector:            sector,
		CapTier:           tier,
		TrueValue:         startingValue,
		ReportedValue:     startingValue, // starts in sync; divergence accumulates via Tick
		SharesOutstanding: shares,
		Float:             shares * floatFraction,
		drift:             drift,
		volatility:        vol,
		jumpParams:        jumpParams,
		rng:               rng,
		reporting:         reportingParams,
	}
}

// GenerateUniverse deterministically produces n companies from a
// single master seed. Each company gets its own independently seeded
// *rand.Rand (derived from the master seed via a simple splitting
// generator), so any individual company's path can be reproduced or
// inspected without needing to replay the entire universe's generation
// in lockstep — useful for debugging a single company's behavior.
//
// Sector and cap tier are assigned by cycling/sampling so the universe
// has broad coverage across the taxonomy rather than being skewed
// toward whatever the RNG happens to favor early on; cap tier is
// weighted toward more mid/small caps and fewer mega caps, mirroring
// how real markets have far more small companies than giant ones.
func GenerateUniverse(n int, masterSeed int64, params GenerationParams) []*Company {
	seeder := rand.New(rand.NewSource(masterSeed))
	usedSymbols := make(map[string]bool)
	companies := make([]*Company, 0, n)

	capWeights := []struct {
		tier   CapTier
		weight float64
	}{
		{MegaCap, 0.05},
		{LargeCap, 0.15},
		{MidCap, 0.35},
		{SmallCap, 0.45},
	}

	for i := range n {
		sector := AllSectors[i%len(AllSectors)] // even coverage across sectors, in order

		tier := weightedCapTier(capWeights, seeder)

		// Derive an independent sub-seed per company from the master
		// seeder, then build that company's own *rand.Rand from it —
		// this is what makes each company's ongoing Tick() sequence
		// reproducible in isolation given just its own seed.
		subSeed := seeder.Int63()
		companyRng := rand.New(rand.NewSource(subSeed))

		c := GenerateCompany(sector, tier, params, companyRng, usedSymbols)
		companies = append(companies, c)
	}

	return companies
}

func weightedCapTier(weights []struct {
	tier   CapTier
	weight float64
}, rng *rand.Rand) CapTier {
	r := rng.Float64()
	var cumulative float64
	for _, w := range weights {
		cumulative += w.weight
		if r <= cumulative {
			return w.tier
		}
	}
	return weights[len(weights)-1].tier
}
