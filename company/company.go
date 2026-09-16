package company

import (
	"math"
	"math/rand"
	"time"
)

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

	// Real-world corporate financial statements & valuation metrics
	AnnualRevenue  float64 // Annual enterprise revenue in USD (e.g. $12.5B)
	NetMargin      float64 // Net profit margin (e.g. 0.22 for 22%)
	SectorMultiple float64 // Sector Price-to-Sales multiple (e.g. 5.5x)
	IPOPrice       float64 // Offering price at listing
	IsIPO          bool    // True if admitted dynamically as a new listing
	IPOTick        int     // Tick number at which the company listed

	// Macroeconomic & Corporate Balance Sheet Footprint
	Headcount         float64 // Total employee headcount (e.g. 45,000)
	AverageWage       float64 // Average annual compensation per employee (e.g. $85,000)
	LaborExpense      float64 // Annual labor payroll costs (Headcount * AverageWage)
	DebtOutstanding   float64 // Total corporate debt on balance sheet (USD)
	InterestExpense   float64 // Annual interest expense (DebtOutstanding * borrowingRate)
	CorporateTaxPaid  float64 // Annual income taxes paid to federal government
	CapEx             float64 // Annual capital investment (contributes to GDP)
	MacroDemandFactor float64 // Composite sector demand multiplier from consumer & trade

	// Private vs Public Enterprise Lifecycle
	IsPublic         bool    // True if listed on chartered stock exchange; false if private venture
	Stage            string  // "Seed", "Growth", "Pre-IPO", "Public"
	PrivateValuation float64 // Estimated valuation for private enterprises

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

	// sessionTicksRemaining and sessionDrift model persistent business/market session trends
	// ensuring markets have decisive directional momentum rather than stationary white-noise barcode wicks.
	sessionTicksRemaining int
	sessionDrift          float64

	// currentTick tracks how many times Tick has been called, purely
	// so emitted events can carry a tick number without the caller
	// having to pass one in — Company already knows its own position
	// in time.
	currentTick int
}

// MarketCap returns the true fundamental enterprise market capitalization.
func (c *Company) MarketCap() float64 {
	return c.TrueValue * c.SharesOutstanding
}

// ReportedMarketCap returns the disclosed/reported enterprise market capitalization.
func (c *Company) ReportedMarketCap() float64 {
	return c.ReportedValue * c.SharesOutstanding
}

// NetIncome returns the annual net profit/earnings of the company.
func (c *Company) NetIncome() float64 {
	return c.AnnualRevenue * c.NetMargin
}

// EarningsPerShare returns the fundamental earnings per share (EPS).
func (c *Company) EarningsPerShare() float64 {
	if c.SharesOutstanding <= 0 {
		return 0
	}
	return c.NetIncome() / c.SharesOutstanding
}

// PriceToEarnings computes the current P/E multiple relative to a market price.
func (c *Company) PriceToEarnings(marketPrice float64) float64 {
	eps := c.EarningsPerShare()
	if eps <= 0 {
		return 0
	}
	return marketPrice / eps
}

// PriceToSales computes the current P/S multiple relative to a market price.
func (c *Company) PriceToSales(marketPrice float64) float64 {
	if c.AnnualRevenue <= 0 {
		return 0
	}
	return (marketPrice * c.SharesOutstanding) / c.AnnualRevenue
}

// OperatingIncome returns enterprise operating profit (EBIT) before interest and taxes.
func (c *Company) OperatingIncome() float64 {
	opMargin := c.NetMargin * 1.35
	return c.AnnualRevenue * opMargin
}

// UpdateMacro applies national macroeconomic conditions to the enterprise:
// - hourlyWage: prevailing national average wage
// - corporateTaxRate: statutory corporate tax rate (e.g. 0.21)
// - borrowingRate: benchmark treasury / commercial debt yield
// - demandMultiplier: sector composite demand factor from consumer spending, foreign trade, and government procurement
// - tickFractionOfYear: fraction of calendar year per tick
func (c *Company) UpdateMacro(hourlyWage float64, corporateTaxRate float64, borrowingRate float64, demandMultiplier float64, tickFractionOfYear float64) {
	if demandMultiplier > 0 {
		c.MacroDemandFactor = demandMultiplier
	} else {
		c.MacroDemandFactor = 1.0
	}

	// 1. Update labor costs based on national wage level and sector wage premium
	if hourlyWage > 0 && c.Headcount > 0 {
		sectorPremium := c.AverageWage / 70000.0
		if sectorPremium < 0.6 {
			sectorPremium = 0.6
		} else if sectorPremium > 2.5 {
			sectorPremium = 2.5
		}
		c.AverageWage = hourlyWage * 2000.0 * sectorPremium
		c.LaborExpense = c.Headcount * c.AverageWage
	}

	// 2. Update interest expense based on national debt yield + credit risk spread
	creditSpread := 0.015
	if c.CapTier == SmallCap {
		creditSpread = 0.035
	} else if c.CapTier == MidCap {
		creditSpread = 0.022
	}
	c.InterestExpense = c.DebtOutstanding * (borrowingRate + creditSpread)

	// 3. Update tax obligations: taxable income = EBIT - Interest
	ebit := c.OperatingIncome()
	taxableIncome := ebit - c.InterestExpense
	if taxableIncome > 0 {
		c.CorporateTaxPaid = taxableIncome * corporateTaxRate
	} else {
		c.CorporateTaxPaid = 0
	}

	// 4. Macro demand influences annual revenue drift:
	revenueGrowth := (c.MacroDemandFactor - 1.0) * 0.05
	c.AnnualRevenue *= (1.0 + revenueGrowth*tickFractionOfYear)
	if c.AnnualRevenue < 1_000_000.0 {
		c.AnnualRevenue = 1_000_000.0
	}

	// 5. Headcount adapts gradually to revenue scale
	if c.Headcount > 0 {
		revPerEmp := c.AnnualRevenue / c.Headcount
		if revPerEmp > 0 {
			targetHeadcount := c.AnnualRevenue / revPerEmp
			c.Headcount = 0.99*c.Headcount + 0.01*targetHeadcount
		}
	}
}

// Volatility returns the annual volatility parameter of the company.
func (c *Company) Volatility() float64 {
	return c.volatility
}

// InitRuntime initializes unexported runtime fields (RNG, jump parameters, drift,
// volatility, reporting profile) needed for simulation execution.
// This is idempotent and safely ensures deserialized companies or newly created
// companies are fully ready for Tick() execution.
func (c *Company) InitRuntime(rng *rand.Rand) {
	if rng != nil {
		c.rng = rng
	} else if c.rng == nil {
		var seed int64
		for _, b := range []byte(c.Symbol) {
			seed = seed*31 + int64(b)
		}
		if seed == 0 {
			seed = time.Now().UnixNano()
		}
		c.rng = rand.New(rand.NewSource(seed))
	}

	sp, ok := sectorProfiles[c.Sector]
	if !ok {
		sp = sectorProfiles[InformationTechnology]
	}
	cp, ok := capTierProfiles[c.CapTier]
	if !ok {
		cp = capTierProfiles[MidCap]
	}

	if c.volatility == 0 {
		volMult := cp.VolMultiplier
		if volMult <= 0 {
			volMult = 1.0
		}
		volSpan := sp.VolMax - sp.VolMin
		if volSpan <= 0 {
			c.volatility = 0.0012 * volMult
		} else {
			c.volatility = (sp.VolMin + c.rng.Float64()*volSpan) * volMult
		}
	}

	if c.drift == 0 {
		driftSpan := sp.DriftMax - sp.DriftMin
		if driftSpan <= 0 {
			c.drift = 0.00003
		} else {
			c.drift = sp.DriftMin + c.rng.Float64()*driftSpan
		}
	}

	if c.jumpParams.LambdaDown == 0 && c.jumpParams.LambdaUp == 0 {
		jumpRisk := sp.JumpRiskMultiplier
		if jumpRisk <= 0 {
			jumpRisk = 1.0
		}
		tierJump := cp.JumpMultiplier
		if tierJump <= 0 {
			tierJump = 1.0
		}
		c.jumpParams = DefaultJumpParams().Scaled(jumpRisk, tierJump)
	}

	if c.reporting.NoiseStdDev == 0 && c.reporting.PersistentBias == 0 && c.reporting.RestatementProbability == 0 {
		profile := chooseReportingProfile(c.Sector, c.rng)
		c.reporting = reportingParamsFor(profile, c.rng)
	}

	if c.MacroDemandFactor <= 0 {
		c.MacroDemandFactor = 1.0
	}
	if c.TrueValue <= 0 {
		c.TrueValue = 10.0
	}
	if c.ReportedValue <= 0 {
		c.ReportedValue = c.TrueValue
	}
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
	if c.rng == nil {
		c.InitRuntime(nil)
	}
	c.currentTick++

	// Decisive session trend momentum:
	// A company experiences distinct market sessions (e.g. 600 to 1800 ticks = ~2.5 to 7.5 minutes)
	// where business performance, product demand, and sector dynamics drive clear directional movement.
	if c.sessionTicksRemaining <= 0 {
		if c.rng != nil {
			c.sessionTicksRemaining = 600 + c.rng.Intn(1200)
			roll := c.rng.Float64()
			if roll < 0.52 {
				// Bullish session: positive growth drift
				c.sessionDrift = 0.00003 + c.rng.Float64()*0.00006
			} else if roll < 0.85 {
				// Bearish / Pullback session: mild negative drift
				c.sessionDrift = -0.00002 - c.rng.Float64()*0.00004
			} else {
				// Neutral / Consolidation session
				c.sessionDrift = (c.rng.Float64() - 0.45) * 0.00002
			}
		}
	} else {
		c.sessionTicksRemaining--
	}

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
		// Ordinary GBM step with session momentum
		z := 0.0
		if c.rng != nil {
			z = c.rng.NormFloat64()
		}
		macroDrift := (c.MacroDemandFactor - 1.0) * 0.00003
		interestDrag := 0.0
		if c.AnnualRevenue > 0 {
			interestDrag = (c.InterestExpense / c.AnnualRevenue) * 0.00001
		}
		effectiveDrift := c.drift + c.sessionDrift + macroDrift - interestDrag
		c.TrueValue *= 1 + effectiveDrift + c.volatility*z
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
	// Bound accumulated bias so it does not drift unbounded to negative or positive infinity
	// over long-running simulations (e.g. conservative bias driving value to zero).
	if c.reportingBiasAccum > 0.45 {
		c.reportingBiasAccum = 0.45
	} else if c.reportingBiasAccum < -0.25 {
		c.reportingBiasAccum = -0.25
	}

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

	shares := cp.SharesOutstandingMin + rng.Float64()*(cp.SharesOutstandingMax-cp.SharesOutstandingMin)
	floatFraction := cp.FloatFractionMin + rng.Float64()*(cp.FloatFractionMax-cp.FloatFractionMin)

	startingValue := params.StartingValueMin + rng.Float64()*(params.StartingValueMax-params.StartingValueMin)

	name := GenerateName(sector, rng)
	symbol := GenerateSymbol(name, usedSymbols)

	reportingProfile := chooseReportingProfile(sector, rng)
	reportingParams := reportingParamsFor(reportingProfile, rng)

	margin := sp.NetMarginMin + rng.Float64()*(sp.NetMarginMax-sp.NetMarginMin)
	multiple := sp.PSMultipleMin + rng.Float64()*(sp.PSMultipleMax-sp.PSMultipleMin)
	revenue := (startingValue * shares) / multiple

	c := &Company{
		Symbol:            symbol,
		Name:              name,
		Sector:            sector,
		CapTier:           tier,
		TrueValue:         startingValue,
		ReportedValue:     startingValue, // starts in sync; divergence accumulates via Tick
		SharesOutstanding: shares,
		Float:             shares * floatFraction,
		AnnualRevenue:     revenue,
		NetMargin:         margin,
		SectorMultiple:    multiple,
		IPOPrice:          startingValue,
		IsIPO:             false,
		IPOTick:           0,
		IsPublic:          true,
		Stage:             "Public",
		drift:             drift,
		volatility:        vol,
		jumpParams:        jumpParams,
		rng:               rng,
		reporting:         reportingParams,
	}
	initMacroMetrics(c, rng)
	return c
}

// GenerateIPOCompany produces an enterprise IPO with specific or procedurally sampled multi-billion revenue.
func GenerateIPOCompany(sector Sector, tier CapTier, customRevenue float64, tick int, rng *rand.Rand, usedSymbols map[string]bool) *Company {
	sp := sectorProfiles[sector]
	cp := capTierProfiles[tier]

	drift := sp.DriftMin + rng.Float64()*(sp.DriftMax-sp.DriftMin)
	vol := (sp.VolMin + rng.Float64()*(sp.VolMax-sp.VolMin)) * cp.VolMultiplier

	baseJump := DefaultJumpParams()
	jumpParams := baseJump.Scaled(sp.JumpRiskMultiplier, cp.JumpMultiplier)

	shares := cp.SharesOutstandingMin + rng.Float64()*(cp.SharesOutstandingMax-cp.SharesOutstandingMin)
	floatFraction := cp.FloatFractionMin + rng.Float64()*(cp.FloatFractionMax-cp.FloatFractionMin)

	revenue := customRevenue
	if revenue <= 0 {
		revenue = cp.RevenueMin + rng.Float64()*(cp.RevenueMax-cp.RevenueMin)
	}

	margin := sp.NetMarginMin + rng.Float64()*(sp.NetMarginMax-sp.NetMarginMin)
	multiple := sp.PSMultipleMin + rng.Float64()*(sp.PSMultipleMax-sp.PSMultipleMin)

	impliedMarketCap := revenue * multiple
	startingValue := impliedMarketCap / shares
	if startingValue < 5.0 {
		startingValue = 5.0
	} else if startingValue > 1500.0 {
		startingValue = 1500.0
	}

	name := GenerateName(sector, rng)
	symbol := GenerateSymbol(name, usedSymbols)

	reportingProfile := chooseReportingProfile(sector, rng)
	reportingParams := reportingParamsFor(reportingProfile, rng)

	c := &Company{
		Symbol:            symbol,
		Name:              name,
		Sector:            sector,
		CapTier:           tier,
		TrueValue:         startingValue,
		ReportedValue:     startingValue,
		SharesOutstanding: shares,
		Float:             shares * floatFraction,
		AnnualRevenue:     revenue,
		NetMargin:         margin,
		SectorMultiple:    multiple,
		IPOPrice:          startingValue,
		IsIPO:             true,
		IPOTick:           tick,
		IsPublic:          true,
		Stage:             "Public",
		drift:             drift,
		volatility:        vol,
		jumpParams:        jumpParams,
		rng:               rng,
		reporting:         reportingParams,
	}
	initMacroMetrics(c, rng)
	return c
}

// GeneratePrivateEnterprise produces an emerging local business (pre-market, seed or growth stage).
func GeneratePrivateEnterprise(sector Sector, seedRevenue float64, tick int, rng *rand.Rand, usedSymbols map[string]bool) *Company {
	c := GenerateIPOCompany(sector, SmallCap, seedRevenue, tick, rng, usedSymbols)
	c.IsPublic = false
	c.IsIPO = false
	c.Stage = "Seed"
	if seedRevenue > 25_000_000 {
		c.Stage = "Growth"
	}
	c.PrivateValuation = c.AnnualRevenue * c.SectorMultiple
	return c
}


func initMacroMetrics(c *Company, rng *rand.Rand) {
	c.MacroDemandFactor = 1.0

	var revPerEmp float64
	var avgWage float64
	var debtToRev float64
	var capexToRev float64

	switch c.Sector {
	case InformationTechnology:
		revPerEmp = 600_000
		avgWage = 140_000
		debtToRev = 0.35
		capexToRev = 0.08
	case Financials:
		revPerEmp = 450_000
		avgWage = 130_000
		debtToRev = 1.20
		capexToRev = 0.04
	case HealthCare:
		revPerEmp = 480_000
		avgWage = 115_000
		debtToRev = 0.50
		capexToRev = 0.09
	case Energy:
		revPerEmp = 850_000
		avgWage = 110_000
		debtToRev = 0.65
		capexToRev = 0.13
	case ConsumerDiscretionary:
		revPerEmp = 240_000
		avgWage = 52_000
		debtToRev = 0.60
		capexToRev = 0.05
	case ConsumerStaples:
		revPerEmp = 290_000
		avgWage = 55_000
		debtToRev = 0.55
		capexToRev = 0.04
	case Industrials:
		revPerEmp = 330_000
		avgWage = 80_000
		debtToRev = 0.60
		capexToRev = 0.06
	case Materials:
		revPerEmp = 380_000
		avgWage = 78_000
		debtToRev = 0.70
		capexToRev = 0.08
	case CommunicationServices:
		revPerEmp = 520_000
		avgWage = 120_000
		debtToRev = 0.85
		capexToRev = 0.11
	case Utilities:
		revPerEmp = 650_000
		avgWage = 95_000
		debtToRev = 1.40
		capexToRev = 0.15
	case RealEstate:
		revPerEmp = 450_000
		avgWage = 90_000
		debtToRev = 1.50
		capexToRev = 0.10
	default:
		revPerEmp = 350_000
		avgWage = 75_000
		debtToRev = 0.60
		capexToRev = 0.06
	}

	if rng != nil {
		revPerEmp *= (0.85 + rng.Float64()*0.30)
		avgWage *= (0.90 + rng.Float64()*0.20)
		debtToRev *= (0.80 + rng.Float64()*0.40)
	}

	c.Headcount = math.Max(25.0, math.Round(c.AnnualRevenue/revPerEmp))
	c.AverageWage = avgWage
	c.LaborExpense = c.Headcount * c.AverageWage
	c.DebtOutstanding = c.AnnualRevenue * debtToRev
	c.InterestExpense = c.DebtOutstanding * 0.045
	c.CapEx = c.AnnualRevenue * capexToRev

	taxable := math.Max(0, c.AnnualRevenue*c.NetMargin*1.35-c.InterestExpense)
	c.CorporateTaxPaid = taxable * 0.21
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
