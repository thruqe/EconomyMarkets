package retail

import "math"

// Signal bundles the inputs an archetype's reaction function can use.
// Not every archetype uses every field — e.g. Degenerate only cares
// about Volatility, MomentumChaser/Contrarian/PanicProne only care
// about Momentum.
type Signal struct {
	// Momentum is a simple rate-of-change measure (see
	// technicals.Momentum), positive for a rising market, negative for
	// falling.
	Momentum float64

	// Volatility is a normalized measure of how wide recent price
	// action has been (e.g. derived from Bollinger band width or ATR
	// relative to price) — what Degenerate reacts to regardless of
	// direction.
	Volatility float64

	// FundamentalGap is (ReportedValue-price)/price, the same shape of
	// signal HedgeFund uses, but read from ReportedValue rather than
	// TrueValue. Zero/unused for TechnicalOnly bots.
	FundamentalGap float64
}

// ActionProbabilities is a buy/sell/hold probability triple that sums
// to 1.0, used to weight a single random draw into a decision.
type ActionProbabilities struct {
	Buy  float64
	Sell float64
	Hold float64
}

// normalize rescales three non-negative weights so they sum to 1,
// guarding against a degenerate all-zero input by returning a pure
// hold.
func normalize(buy, sell, hold float64) ActionProbabilities {
	total := buy + sell + hold
	if total <= 0 {
		return ActionProbabilities{Hold: 1}
	}
	return ActionProbabilities{Buy: buy / total, Sell: sell / total, Hold: hold / total}
}

// sigmoid maps any real value into (0, 1), used to turn an unbounded
// signal strength into a bounded probability weight.
func sigmoid(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}

// ReactionFunc computes buy/sell/hold probabilities for a given
// archetype and signal. baseActivity scales overall willingness to
// act at all (vs. hold) — lower for Disciplined, higher for
// Degenerate/Beginner-flavored archetypes — kept as a parameter here
// rather than hardcoded per archetype, since SkillTier can also
// influence overall activity level independent of Archetype.
type ReactionFunc func(s Signal, baseActivity float64) ActionProbabilities

// reactionFuncs maps each Archetype to its ReactionFunc. Kept as a
// lookup table rather than a switch inside a single function so each
// archetype's logic is independently readable and testable.
var reactionFuncs = map[Archetype]ReactionFunc{
	MomentumChaser: func(s Signal, baseActivity float64) ActionProbabilities {
		// Buys strength, sells weakness: momentum directly drives the
		// buy/sell weight split, symmetric in both directions.
		strength := sigmoid(s.Momentum * 20) // scale momentum (a small fraction) into a meaningful sigmoid input
		buy := baseActivity * strength
		sell := baseActivity * (1 - strength)
		hold := 1 - baseActivity
		return normalize(buy, sell, hold)
	},

	Contrarian: func(s Signal, baseActivity float64) ActionProbabilities {
		// Mirror of MomentumChaser: buys weakness, sells strength.
		strength := sigmoid(s.Momentum * 20)
		buy := baseActivity * (1 - strength)
		sell := baseActivity * strength
		hold := 1 - baseActivity
		return normalize(buy, sell, hold)
	},

	PanicProne: func(s Signal, baseActivity float64) ActionProbabilities {
		// Asymmetric: small positive momentum registers mild buy response,
		// but negative momentum is amplified sharply into a sell response.
		if s.Momentum >= 0 {
			buy := baseActivity * (0.10 + 0.30*sigmoid(s.Momentum*10))
			sell := baseActivity * 0.05
			hold := 1 - buy - sell
			return normalize(buy, sell, hold)
		}
		amplified := math.Min(1, math.Abs(s.Momentum)*40) // sharp amplification of drops
		sell := baseActivity * (0.20 + 0.80*amplified)
		hold := 1 - sell
		return normalize(0, sell, hold)
	},

	Disciplined: func(s Signal, baseActivity float64) ActionProbabilities {
		// Small, dampened reactions across the board with steady baseline engagement.
		dampened := baseActivity * 0.35
		strength := sigmoid(s.Momentum * 20)
		buy := dampened * strength
		sell := dampened * (1 - strength)
		hold := 1 - dampened
		return normalize(buy, sell, hold)
	},

	Degenerate: func(s Signal, baseActivity float64) ActionProbabilities {
		// Reacts to volatility itself with a reliable baseline activity floor (0.35)
		// ensuring ongoing speculative retail flow so the market never permanently dies.
		volDriven := baseActivity * (0.35 + 0.65*math.Min(1, s.Volatility*8))
		directionLean := sigmoid(s.Momentum * 10) // mild lean, not the main driver
		buy := volDriven * directionLean
		sell := volDriven * (1 - directionLean)
		hold := 1 - volDriven
		return normalize(buy, sell, hold)
	},
}

// ReactionFor returns the ReactionFunc for a given archetype.
func ReactionFor(a Archetype) ReactionFunc {
	return reactionFuncs[a]
}
