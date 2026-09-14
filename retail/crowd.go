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

// attentionScore estimates how much retail attention a company would
// realistically draw, computed from static properties available at
// generation time (sector/cap-tier-driven volatility, and smaller cap
// tiers generally drawing more speculative retail interest than
// mega-caps). This is a genuine simplification worth being explicit
// about: real retail attention is highly time-varying — a company can
// become "the" story stock this week and be ignored next month, often
// driven by news/social attention this simulation doesn't model as an
// attention-specific signal. A static, generation-time score is a
// reasonable first approximation (it does correctly bias toward
// smaller, more volatile names the way real retail attention tends
// to), but a fuller version would let attention shift dynamically
// during a run — e.g. spiking after a FundamentalEvent or a large
// recent price move — which is a natural future refinement, not
// something quietly claimed as already present here.
func attentionScore(c *company.Company) float64 {
	capWeight := map[company.CapTier]float64{
		company.MegaCap:  0.3,
		company.LargeCap: 0.6,
		company.MidCap:   1.0,
		company.SmallCap: 1.6, // smaller caps draw disproportionate speculative retail interest
	}[c.CapTier]
	return capWeight
}

// GenerateWatchlistPool deterministically produces n
// SimulatedRetailTrader bots from a single master seed, sharing one
// pool across the entire universe rather than one crowd per company —
// each bot is assigned a small watchlist (between minWatchlist and
// maxWatchlist companies, inclusive) drawn from universe, sampled with
// probability proportional to attentionScore. This concentrates most
// bots onto a comparatively small set of "attention" companies (real
// retail's well-documented concentration in story/momentum/meme
// names) while most companies in the universe end up with
// comparatively little retail coverage — the intended contrast with a
// uniform per-company crowd, and a closer match to how real retail
// attention is actually distributed across a market.
//
// Each bot gets its own independently seeded *rand.Rand (derived the
// same way company.GenerateUniverse derives per-company sub-seeds),
// so any individual bot's behavior can be reproduced or inspected in
// isolation. startingCashMin/Max bounds retail-scale starting capital
// — small relative to any institutional agent's account, which is
// what keeps any single bot from ever being able to move the market
// the way a HedgeFund/Bank can; only their aggregate, correlated
// behavior (e.g. a stop-loss cluster sweep) produces a visible effect.
func GenerateWatchlistPool(
	n int,
	masterSeed int64,
	universe []*company.Company,
	minWatchlist, maxWatchlist int,
	startingCashMin, startingCashMax float64,
) []*SimulatedRetailTrader {
	seeder := rand.New(rand.NewSource(masterSeed))
	pool := make([]*SimulatedRetailTrader, 0, n)

	scored := make([]weightedItem[*company.Company], len(universe))
	for i, c := range universe {
		scored[i] = weightedItem[*company.Company]{item: c, weight: attentionScore(c)}
	}

	for i := range n {
		tier := pickTier(seeder)
		style := pickStyle(seeder)
		archetype := pickArchetype(seeder)
		cash := startingCashMin + seeder.Float64()*(startingCashMax-startingCashMin)

		watchlistSize := minWatchlist
		if maxWatchlist > minWatchlist {
			watchlistSize = minWatchlist + seeder.Intn(maxWatchlist-minWatchlist+1)
		}
		watchlist := sampleWatchlist(scored, watchlistSize, seeder)

		subSeed := seeder.Int63()
		botRng := rand.New(rand.NewSource(subSeed))

		id := "retail_" + itoa(i)
		acct := market.NewAccount(id, cash, 2.0, 0.20) // modest retail-scale leverage, generous maintenance margin

		bot := NewSimulatedRetailTrader(id, acct, watchlist, tier, style, archetype, botRng)
		pool = append(pool, bot)
	}

	return pool
}

// sampleWatchlist draws size distinct companies from scored, with
// selection probability proportional to each company's weight,
// without replacement (a bot's watchlist shouldn't contain the same
// company twice). Uses a simple weighted-without-replacement approach
// (draw, remove, renormalize) rather than a more elaborate algorithm,
// since watchlist sizes here are always small (a handful of
// companies), making the O(size * len(scored)) cost negligible.
func sampleWatchlist(scored []weightedItem[*company.Company], size int, rng *rand.Rand) []*company.Company {
	if size > len(scored) {
		size = len(scored)
	}

	remaining := make([]weightedItem[*company.Company], len(scored))
	copy(remaining, scored)

	result := make([]*company.Company, 0, size)
	for len(result) < size && len(remaining) > 0 {
		picked := weightedPick(remaining, rng)
		result = append(result, picked)

		for i, it := range remaining {
			if it.item == picked {
				remaining = append(remaining[:i], remaining[i+1:]...)
				break
			}
		}
	}

	return result
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
