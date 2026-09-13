// Package retail contains retail-style market participants: both
// human-driven traders (a thin bridge from UI actions to orders) and
// simulated retail bots. This package deliberately shares nothing with
// package agent beyond market.OrderSource — no common interface, no
// shared status/lifecycle vocabulary — mirroring how, in reality, a
// hedge fund and a retail trader have no relationship to each other
// beyond both observing and acting on the same public order book.
//
// Simulated retail bots vary along three independent axes, generated
// the same way package company generates its universe: many bots from
// a master seed, each combining a SkillTier, an InformationStyle, and
// an Archetype rather than being hand-authored individually.
package retail

// SkillTier governs how sophisticated a bot's process is: which
// technicals window it watches, how disciplined its position sizing
// and stop placement are, and whether it pays any attention to
// fundamentals at all.
type SkillTier int

const (
	Beginner SkillTier = iota
	Intermediate
	Pro
)

func (s SkillTier) String() string {
	switch s {
	case Beginner:
		return "Beginner"
	case Intermediate:
		return "Intermediate"
	case Pro:
		return "Pro"
	default:
		return "Unknown Tier"
	}
}

// InformationStyle governs what kind of signal a bot actually trusts.
// This is independent of SkillTier and Archetype: a Beginner and a Pro
// can both be TechnicalOnly, just reacting to different window sizes
// and with different discipline.
type InformationStyle int

const (
	// TechnicalOnly bots react purely to price action via their own
	// technicals.Aggregator — momentum, and for higher tiers, RSI/MACD.
	// They never look at company fundamentals at all.
	TechnicalOnly InformationStyle = iota

	// FundamentalOnly bots compare company.ReportedValue (not
	// TrueValue — see package company's documentation on why nobody,
	// retail or institutional, should read TrueValue directly in a
	// realistic setup) against market price. This is the style that
	// can be genuinely, blamelessly wrong on a company with a
	// misleading ReportingProfile: the bot is doing exactly what a
	// fundamentals-driven trader should do with the information
	// actually available to it.
	FundamentalOnly

	// Blended bots weigh both a technical signal and a fundamental
	// signal together.
	Blended

	// Confused bots apply their rule inconsistently tick to tick —
	// sometimes reacting to technicals, sometimes to fundamentals,
	// sometimes to neither in any stable way. This models a real and
	// common retail pattern: no consistent process, not a weaker
	// version of a consistent one.
	Confused
)

func (i InformationStyle) String() string {
	switch i {
	case TechnicalOnly:
		return "TechnicalOnly"
	case FundamentalOnly:
		return "FundamentalOnly"
	case Blended:
		return "Blended"
	case Confused:
		return "Confused"
	default:
		return "Unknown Style"
	}
}

// Archetype governs the shape of a bot's reaction function: what
// signal direction it responds to, and how.
type Archetype int

const (
	// MomentumChaser buys on positive momentum, sells on negative —
	// buys strength, sells weakness, the classic FOMO pattern.
	MomentumChaser Archetype = iota

	// Contrarian does the reverse: buys weakness (dips), sells
	// strength (rips) — the "buy the dip" retail pattern.
	Contrarian

	// PanicProne has an amplified, asymmetric reaction specifically to
	// negative momentum: small positive moves barely register, sharp
	// drops trigger a disproportionately large sell probability.
	PanicProne

	// Disciplined has small, dampened reactions across the board and
	// longer cooldowns — the rarer, more patient retail trader.
	Disciplined

	// Degenerate reacts most strongly to volatility itself (via
	// Bollinger band width / ATR), regardless of direction — attracted
	// to chaos, not trend.
	Degenerate
)

func (a Archetype) String() string {
	switch a {
	case MomentumChaser:
		return "MomentumChaser"
	case Contrarian:
		return "Contrarian"
	case PanicProne:
		return "PanicProne"
	case Disciplined:
		return "Disciplined"
	case Degenerate:
		return "Degenerate"
	default:
		return "Unknown Archetype"
	}
}

// TierProfile bundles the parameters SkillTier controls.
type TierProfile struct {
	// AggregatorWindow is how many raw ticks this tier rolls into one
	// technicals bar — small for a Beginner (reacts to short-term
	// noise), larger for a Pro (reacts to a smoother, longer view).
	AggregatorWindow int

	// MomentumWindow/RSIWindow/etc. are expressed in *bars* (not raw
	// ticks), consistent with how package technicals' indicators take
	// a bar window.
	MomentumWindow int

	// UsesAdvancedIndicators gates whether a bot of this tier computes
	// RSI/MACD at all, versus only simple momentum — a Beginner
	// realistically doesn't think in RSI/MACD terms.
	UsesAdvancedIndicators bool

	// ReactsToFundamentalEvents gates whether this tier pays any
	// attention to company.FundamentalEvent/RestatementEvent news at
	// all — real casual retail mostly doesn't read past headlines;
	// sophisticated retail (Pro) is the tier most likely to actually
	// engage with reported fundamentals.
	ReactsToFundamentalEvents bool

	// PositionSizeFractionMin/Max bound how much of available buying
	// power a single trade commits, as a fraction — Beginners/
	// Degenerates size large and undisciplined, Pros size small and
	// consistent.
	PositionSizeFractionMin, PositionSizeFractionMax float64

	// StopLossPlacement governs how a fresh stop is placed relative to
	// entry/recent swing levels — see StopPlacementStyle.
	StopPlacement StopPlacementStyle

	// CooldownTicksMin/Max bound how long a bot waits after acting
	// before it will consider trading again.
	CooldownTicksMin, CooldownTicksMax int
}

// StopPlacementStyle governs how predictable/exploitable a bot's stop
// placement is — the mechanism that determines whether a bot's stop
// contributes to a genuine, sweepable cluster of resting liquidity at
// obvious levels, or sits somewhere less predictable.
type StopPlacementStyle int

const (
	// RoundNumberStop places stops at the nearest obvious round
	// number below/above entry — exactly where real liquidity sweeps
	// target, since many traders independently place stops at the
	// same predictable levels.
	RoundNumberStop StopPlacementStyle = iota

	// SwingLevelStop places stops just past the most recent swing
	// high/low — also a common, somewhat predictable real pattern.
	SwingLevelStop

	// BufferedStop places stops at a wider, less predictable
	// ATR-scaled distance from entry — real, more sophisticated stop
	// placement that isn't sitting exactly at an obvious level.
	BufferedStop
)

var tierProfiles = map[SkillTier]TierProfile{
	Beginner: {
		AggregatorWindow:          5,
		MomentumWindow:            3,
		UsesAdvancedIndicators:    false,
		ReactsToFundamentalEvents: false,
		PositionSizeFractionMin:   0.15,
		PositionSizeFractionMax:   0.60, // undisciplined, sometimes near all-in
		StopPlacement:             RoundNumberStop,
		CooldownTicksMin:          2,
		CooldownTicksMax:          6,
	},
	Intermediate: {
		AggregatorWindow:          15,
		MomentumWindow:            8,
		UsesAdvancedIndicators:    false,
		ReactsToFundamentalEvents: false,
		PositionSizeFractionMin:   0.05,
		PositionSizeFractionMax:   0.25,
		StopPlacement:             SwingLevelStop,
		CooldownTicksMin:          4,
		CooldownTicksMax:          12,
	},
	Pro: {
		AggregatorWindow:          50,
		MomentumWindow:            20,
		UsesAdvancedIndicators:    true,
		ReactsToFundamentalEvents: true,
		PositionSizeFractionMin:   0.02,
		PositionSizeFractionMax:   0.08, // small, consistent
		StopPlacement:             BufferedStop,
		CooldownTicksMin:          8,
		CooldownTicksMax:          20,
	},
}
