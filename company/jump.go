package company

import "math/rand"

// JumpParams defines the rare-event ("jump") side of a company's value
// process, layered on top of ordinary GBM drift. This is what produces
// crisis-scale drops and, much more rarely, mania-scale revaluations —
// discrete, fat-tailed shocks that a continuous GBM process cannot
// produce on its own by construction.
//
// Calibrated so that, per company over roughly 1000 ticks: a crisis-
// scale jump occurs on average once every ~800-1200 ticks (most
// companies see zero or one across a run), and a mania-scale jump
// occurs on average once every ~5000-8000 ticks (meaning, across a
// full 500-company universe over 1000 ticks, only a small handful of
// manias happen across the *entire* universe — a singular, memorable
// event rather than routine behavior, deliberately mirroring how rare
// something like a SpaceX-style revaluation actually is).
type JumpParams struct {
	// LambdaDown/LambdaUp are per-tick probabilities of a downside or
	// upside jump occurring at all. LambdaDown is set higher than
	// LambdaUp by design — the "leverage effect": bad news hits more
	// often (and, via magnitude below, harder) than equivalent good
	// news.
	LambdaDown float64
	LambdaUp   float64

	// Downside jump magnitude, drawn uniformly from this range when a
	// downside jump triggers. Expressed as a fractional drop (0.35 =
	// -35%).
	DownJumpMin, DownJumpMax float64

	// Upside jump magnitude for ordinary good news (not mania-scale),
	// drawn uniformly from this range when an upside jump triggers and
	// the mania sub-check (below) does not fire. Expressed as a
	// fractional gain (0.25 = +25%).
	UpJumpMin, UpJumpMax float64

	// ManiaProbability is the conditional probability that, given an
	// upside jump has already been triggered, it escalates to a
	// mania-scale event instead of ordinary good news.
	ManiaProbability float64

	// ManiaMultiplierMin/Max bound the multiplier applied to value on
	// a mania event (2.0 = value doubles). Deliberately kept in the
	// low single digits (not 10x+) — large enough to be a genuine,
	// talked-about event, without producing values so extreme they
	// dwarf everything else in the simulation and read as a bug.
	ManiaMultiplierMin, ManiaMultiplierMax float64
}

// DefaultJumpParams returns baseline jump parameters before any
// sector/cap-tier scaling is applied. Per-tick probabilities are
// deliberately small; see the type doc for the resulting expected
// frequency over a ~1000-tick run.
func DefaultJumpParams() JumpParams {
	return JumpParams{
		LambdaDown:         1.0 / 2000, // downside jump frequency with leverage effect
		LambdaUp:           1.0 / 2500, // positive growth catalyst frequency
		DownJumpMin:        0.06,
		DownJumpMax:        0.25,
		UpJumpMin:          0.06,
		UpJumpMax:          0.25,
		ManiaProbability:   0.03, // 3% of upside jumps escalate to mania scale
		ManiaMultiplierMin: 1.5,
		ManiaMultiplierMax: 3.5,
	}
}

// Scaled returns a copy of p with LambdaDown/LambdaUp adjusted by the
// given sector and cap-tier jump-risk multipliers. Magnitude ranges
// and ManiaProbability are left unscaled deliberately — the intent is
// that riskier sectors/smaller caps have jumps more *often*, not that
// their jumps are individually larger or more mania-prone once
// triggered; keeping magnitude distributions shared makes rare events
// comparable in kind across the whole universe, varying only in how
// often each company is exposed to them.
func (p JumpParams) Scaled(sectorMultiplier, capMultiplier float64) JumpParams {
	scaled := p
	scaled.LambdaDown = p.LambdaDown * sectorMultiplier * capMultiplier
	scaled.LambdaUp = p.LambdaUp * sectorMultiplier * capMultiplier
	return scaled
}

// JumpKind categorizes what happened on a tick where a jump fired, for
// logging/analysis — e.g. so a research notebook can count how many
// crises vs. manias occurred across the universe over a run.
type JumpKind int

const (
	NoJump JumpKind = iota
	CrisisJump
	GoodNewsJump
	ManiaJump
)

func (k JumpKind) String() string {
	switch k {
	case CrisisJump:
		return "crisis"
	case GoodNewsJump:
		return "good_news"
	case ManiaJump:
		return "mania"
	default:
		return "none"
	}
}

// JumpResult records what a single tick's jump check produced, so
// callers can log/react to jumps distinctly from ordinary drift.
type JumpResult struct {
	Kind       JumpKind
	Multiplier float64 // applied multiplicatively to value; 1.0 if Kind == NoJump
}

// rollJump performs one tick's Poisson-style jump check against rng,
// returning the outcome. At most one jump (down or up) can fire per
// tick — a company experiencing both a crisis and a mania in the same
// instant isn't a case worth modeling, so downside is checked first
// and, if it doesn't fire, upside is checked.
func rollJump(p JumpParams, rng *rand.Rand) JumpResult {
	if rng == nil {
		return JumpResult{Kind: NoJump, Multiplier: 1.0}
	}
	if rng.Float64() < p.LambdaDown {
		frac := p.DownJumpMin + rng.Float64()*(p.DownJumpMax-p.DownJumpMin)
		return JumpResult{Kind: CrisisJump, Multiplier: 1.0 - frac}
	}
	if rng.Float64() < p.LambdaUp {
		if rng.Float64() < p.ManiaProbability {
			mult := p.ManiaMultiplierMin + rng.Float64()*(p.ManiaMultiplierMax-p.ManiaMultiplierMin)
			return JumpResult{Kind: ManiaJump, Multiplier: mult}
		}
		frac := p.UpJumpMin + rng.Float64()*(p.UpJumpMax-p.UpJumpMin)
		return JumpResult{Kind: GoodNewsJump, Multiplier: 1.0 + frac}
	}
	return JumpResult{Kind: NoJump, Multiplier: 1.0}
}
