package retail

import (
	"math/rand"

	"economy/company"
	"economy/market"
)

// weightedItem pairs a value with a selection weight, used by
// weightedPick below.
type weightedItem[T any] struct {
	item   T
	weight float64
}

func weightedPick[T any](items []weightedItem[T], rng *rand.Rand) T {
	var total float64
	for _, it := range items {
		total += it.weight
	}
	r := rng.Float64() * total
	var cumulative float64
	for _, it := range items {
		cumulative += it.weight
		if r <= cumulative {
			return it.item
		}
	}
	return items[len(items)-1].item
}

var tierPickWeights = []weightedItem[SkillTier]{
	{Beginner, 0.55},
	{Intermediate, 0.35},
	{Pro, 0.10},
}

var stylePickWeights = []weightedItem[InformationStyle]{
	{TechnicalOnly, 0.45},
	{Confused, 0.25},
	{Blended, 0.20},
	{FundamentalOnly, 0.10},
}

var archetypePickWeights = []weightedItem[Archetype]{
	{MomentumChaser, 0.30},
	{Contrarian, 0.20},
	{PanicProne, 0.20},
	{Degenerate, 0.20},
	{Disciplined, 0.10}, // genuinely rarer among retail, as intended
}

func pickTier(rng *rand.Rand) SkillTier         { return weightedPick(tierPickWeights, rng) }
func pickStyle(rng *rand.Rand) InformationStyle { return weightedPick(stylePickWeights, rng) }
func pickArchetype(rng *rand.Rand) Archetype    { return weightedPick(archetypePickWeights, rng) }

// GenerateCrowd deterministically produces n SimulatedRetailTrader
// bots from a single master seed, each trading the given target
// company, distributed across tier/style/archetype according to the
// weightings above. Each bot gets its own independently seeded
// *rand.Rand (derived the same way company.GenerateUniverse derives
// per-company sub-seeds), so any individual bot's behavior can be
// reproduced or inspected in isolation.
//
// startingCashMin/Max bounds retail-scale starting capital — small
// relative to any institutional agent's account, which is what keeps
// any single bot from ever being able to move the market the way a
// HedgeFund/Bank can; only their aggregate, correlated behavior (e.g.
// a stop-loss cluster sweep) produces a visible effect.
func GenerateCrowd(n int, masterSeed int64, target *company.Company, startingCashMin, startingCashMax float64) []*SimulatedRetailTrader {
	seeder := rand.New(rand.NewSource(masterSeed))
	crowd := make([]*SimulatedRetailTrader, 0, n)

	for i := range n {
		tier := pickTier(seeder)
		style := pickStyle(seeder)
		archetype := pickArchetype(seeder)
		cash := startingCashMin + seeder.Float64()*(startingCashMax-startingCashMin)

		subSeed := seeder.Int63()
		botRng := rand.New(rand.NewSource(subSeed))

		id := "retail_" + itoa(i)
		acct := market.NewAccount(id, cash, 2.0, 0.20) // modest retail-scale leverage, generous maintenance margin

		bot := NewSimulatedRetailTrader(id, acct, target, tier, style, archetype, botRng)
		crowd = append(crowd, bot)
	}

	return crowd
}

// itoa is a tiny local integer-to-string helper, mirroring the one in
// package company's naming.go — kept private/duplicated rather than
// shared, consistent with retail importing nothing beyond market,
// company, and technicals.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
