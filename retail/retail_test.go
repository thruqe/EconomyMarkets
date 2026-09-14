package retail

import (
	"math/rand"
	"testing"

	"economy/company"
	"economy/market"
)

func makeFlatCompany(symbol string, value float64) *company.Company {
	return makeFlatCompanyWithTier(symbol, value, company.MidCap)
}

// makeFlatCompanyWithTier is like makeFlatCompany but lets tests
// control CapTier explicitly — needed anywhere attentionScore's
// cap-tier-driven weighting is under test, since attentionScore reads
// c.CapTier directly and GenerateCompany's own tier argument only
// influences generated *ranges* (share count, float, vol multiplier),
// not a field tests can rely on matching a symbol's suggestive name
// unless set explicitly.
func makeFlatCompanyWithTier(symbol string, value float64, tier company.CapTier) *company.Company {
	used := make(map[string]bool)
	params := company.DefaultGenerationParams()
	params.BaseJumpParams.LambdaDown = 0
	params.BaseJumpParams.LambdaUp = 0
	params.StartingValueMin = value
	params.StartingValueMax = value

	c := company.GenerateCompany(company.InformationTechnology, tier, params, rand.New(rand.NewSource(1)), used)
	c.Symbol = symbol
	c.CapTier = tier
	c.TrueValue = value
	c.ReportedValue = value
	return c
}

// makeTestUniverse builds a small, deterministic set of flat-valued
// companies across a genuine spread of cap tiers (see
// makeFlatCompanyWithTier), used by tests that need
// GenerateWatchlistPool's attention-weighted sampling to have
// something real to sample from.
func makeTestUniverse() []*company.Company {
	return []*company.Company{
		makeFlatCompanyWithTier("MEGA", 500, company.MegaCap),
		makeFlatCompanyWithTier("LARG", 200, company.LargeCap),
		makeFlatCompanyWithTier("MIDA", 100, company.MidCap),
		makeFlatCompanyWithTier("MIDB", 90, company.MidCap),
		makeFlatCompanyWithTier("SMLA", 20, company.SmallCap),
		makeFlatCompanyWithTier("SMLB", 15, company.SmallCap),
		makeFlatCompanyWithTier("SMLC", 10, company.SmallCap),
	}
}

// TestHumanTraderDrainsQueueOnce confirms HumanTrader returns queued
// orders exactly once, then nil on subsequent calls until more are
// queued.
func TestHumanTraderDrainsQueueOnce(t *testing.T) {
	acct := market.NewAccount("human1", 10000, 2, 0.2)
	h := NewHumanTrader("human1", acct)

	state := market.MarketState{Symbol: "SYN", HasMid: true, Mid: 100}

	if orders := h.NextOrders(state); orders != nil {
		t.Fatalf("expected nil with nothing queued, got %d orders", len(orders))
	}

	h.SubmitOrder(&market.Order{Side: market.Buy, Quantity: 10, IsMarket: true})
	orders := h.NextOrders(state)
	if len(orders) != 1 {
		t.Fatalf("expected exactly 1 queued order, got %d", len(orders))
	}
	if orders[0].AgentID != "human1" {
		t.Fatalf("expected SubmitOrder to force correct AgentID, got %s", orders[0].AgentID)
	}

	if orders := h.NextOrders(state); orders != nil {
		t.Fatalf("expected nil after queue drained, got %d orders", len(orders))
	}
}

// TestBotIgnoresUnwatchedSymbol confirms a bot produces no orders for
// a company outside its watchlist — the mechanism that lets a shared
// pool serve many companies without every bot reacting to everything.
func TestBotIgnoresUnwatchedSymbol(t *testing.T) {
	watched := makeFlatCompany("WTCH", 100)
	acct := market.NewAccount("bot1", 10000, 2, 0.2)
	bot := NewSimulatedRetailTrader("bot1", acct, []*company.Company{watched}, Beginner, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(1)))

	if bot.Watches("OTHER") {
		t.Fatalf("expected bot to not watch a company outside its watchlist")
	}

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 99, Quantity: 1000})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 101, Quantity: 1000})
	state := market.NewMarketState("OTHER", 1, book, 5, nil, 0)
	state.HasMid = true
	state.Mid = 100

	orders := bot.NextOrders(state)
	if orders != nil {
		t.Fatalf("expected nil for an unwatched symbol, got %d orders", len(orders))
	}
}

// TestStopLossTriggersOnBreach directly constructs a bot with an open
// long position and an active stop, then confirms a price breach
// produces a forced-exit market order — the core sweep-enabling
// mechanic.
func TestStopLossTriggersOnBreach(t *testing.T) {
	comp := makeFlatCompany("SYN", 100)
	acct := market.NewAccount("bot1", 10000, 2, 0.2)
	bot := NewSimulatedRetailTrader("bot1", acct, []*company.Company{comp}, Beginner, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(1)))

	market.ApplySettledFill(acct, "SYN", market.Buy, 10, 100)
	bot.state["SYN"].stop = stopLoss{active: true, triggerAt: 95, isLong: true}

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 94, Quantity: 1000})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 96, Quantity: 1000})

	state := market.NewMarketState("SYN", 1, book, 5, nil, 0)
	state.Mid = 94
	state.HasMid = true

	orders := bot.NextOrders(state)
	if len(orders) != 1 {
		t.Fatalf("expected exactly one forced exit order on stop breach, got %d", len(orders))
	}
	if orders[0].Side != market.Sell {
		t.Fatalf("expected a sell order to close the long, got side=%v", orders[0].Side)
	}
	if orders[0].Quantity != 10 {
		t.Fatalf("expected the exit order to close the full 10-unit position, got qty=%.2f", orders[0].Quantity)
	}
	if bot.state["SYN"].stop.active {
		t.Fatalf("expected the stop to be deactivated after triggering")
	}
}

// TestStopLossDoesNotTriggerAbovePrice confirms a long position's
// stop stays inactive while price remains above the trigger.
func TestStopLossDoesNotTriggerAbovePrice(t *testing.T) {
	comp := makeFlatCompany("SYN", 100)
	acct := market.NewAccount("bot1", 10000, 2, 0.2)
	bot := NewSimulatedRetailTrader("bot1", acct, []*company.Company{comp}, Beginner, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(1)))

	market.ApplySettledFill(acct, "SYN", market.Buy, 10, 100)
	bot.state["SYN"].stop = stopLoss{active: true, triggerAt: 95, isLong: true}

	order := bot.checkStopLoss("SYN", 98)
	if order != nil {
		t.Fatalf("expected no stop trigger while price (98) stays above trigger (95)")
	}
}

// TestLiquiditySweepCascade is the important integration proof: many
// Beginner bots (RoundNumberStop — the most predictable, most
// sweepable placement) open long positions clustered near the same
// entry price, so their stops cluster at the same round-number level.
// Once a large sell pushes price through that level, this test
// confirms a meaningful fraction of the crowd's stops fire in the same
// or immediately following tick — the cascade a real liquidity sweep
// produces, rather than stops triggering independently and randomly
// spread out.
func TestLiquiditySweepCascade(t *testing.T) {
	comp := makeFlatCompany("SYN", 100)

	const n = 200
	crowd := make([]*SimulatedRetailTrader, 0, n)
	for i := 0; i < n; i++ {
		acct := market.NewAccount("bot"+itoa(i), 10000, 2, 0.2)
		bot := NewSimulatedRetailTrader("bot"+itoa(i), acct, []*company.Company{comp}, Beginner, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(int64(i))))
		// Force every bot long from the same entry, with a stop placed
		// via the real RoundNumberStop logic — this is what should
		// cluster them at the same trigger level.
		market.ApplySettledFill(acct, "SYN", market.Buy, 10, 100)
		bot.placeStop(bot.state["SYN"], 100, true, nil)
		crowd = append(crowd, bot)
	}

	// Confirm the stops actually clustered at (or near) the same level
	// before testing the cascade, since the cascade proof depends on
	// that clustering being real.
	triggerLevels := make(map[float64]int)
	for _, b := range crowd {
		triggerLevels[b.state["SYN"].stop.triggerAt]++
	}
	t.Logf("distinct stop trigger levels among %d bots: %d", n, len(triggerLevels))
	if len(triggerLevels) > 3 {
		t.Fatalf("expected RoundNumberStop placement to cluster stops into very few distinct levels, got %d distinct levels", len(triggerLevels))
	}

	triggeredCount := 0
	for _, b := range crowd {
		if order := b.checkStopLoss("SYN", 92); order != nil {
			triggeredCount++
		}
	}

	t.Logf("bots triggered on the sweep: %d / %d", triggeredCount, n)

	if triggeredCount < n/2 {
		t.Fatalf("expected a clear majority of clustered-stop bots to trigger together on the sweep, got %d/%d", triggeredCount, n)
	}
}

// TestFundamentalOnlyBotMisledByFraudulentCompany is the scenario the
// whole reporting-integrity layer was built for: a FundamentalOnly bot
// reads ReportedValue (not TrueValue), so on a company whose
// ReportedValue has drifted well above TrueValue (simulating a
// Fraudulent profile mid-buildup, before discovery), the bot judges
// the market UNDERVALUED and buys — a completely reasonable decision
// given the information it had access to, even though the position is
// actually a bad one relative to true fundamentals. This test proves
// the bot's decision tracks ReportedValue, not TrueValue.
func TestFundamentalOnlyBotMisledByFraudulentCompany(t *testing.T) {
	comp := makeFlatCompany("SYN", 100)
	comp.TrueValue = 60      // real value has quietly deteriorated
	comp.ReportedValue = 140 // but reports still show strength - the fraud gap

	acct := market.NewAccount("bot1", 1_000_000, 2, 0.2)
	bot := NewSimulatedRetailTrader("bot1", acct, []*company.Company{comp}, Pro, FundamentalOnly, MomentumChaser, rand.New(rand.NewSource(1)))

	book := market.NewOrderBook()
	for i := 0; i < 10; i++ {
		book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 99 - float64(i), Quantity: 5000})
		book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 101 + float64(i), Quantity: 5000})
	}

	// Warm up the bot's aggregator with enough ticks to get a
	// technicals-window-sized bar history (Pro tier needs
	// AggregatorWindow=50 ticks per bar); state.Mid held flat so any
	// resulting decision is attributable to the fundamental gap, not
	// technical momentum.
	var lastOrders []*market.Order
	for i := 0; i < 60; i++ {
		state := market.NewMarketState("SYN", i, book, 5, nil, 0)
		state.HasMid = true
		state.Mid = 100
		lastOrders = bot.NextOrders(state)
		if len(lastOrders) > 0 {
			break
		}
	}

	if len(lastOrders) == 0 {
		t.Fatalf("expected the bot to eventually act on a large (40%%) fundamental gap once it had enough bar history")
	}
	if lastOrders[0].Side != market.Buy {
		t.Fatalf("expected the bot to buy, since ReportedValue (140) is well above market price (100) — this is the misled-but-reasonable decision the fraud scenario produces, got side=%v", lastOrders[0].Side)
	}

	t.Logf("FundamentalOnly bot bought based on ReportedValue=%.2f while TrueValue=%.2f — misled by the reporting gap exactly as designed",
		comp.ReportedValue, comp.TrueValue)
}

// TestArchetypeDirectionsDiffer confirms MomentumChaser and Contrarian
// produce opposite-leaning buy probabilities for the same positive
// momentum signal — the core defining property distinguishing them.
func TestArchetypeDirectionsDiffer(t *testing.T) {
	signal := Signal{Momentum: 0.05} // clearly positive momentum

	momentumProbs := ReactionFor(MomentumChaser)(signal, 0.5)
	contrarianProbs := ReactionFor(Contrarian)(signal, 0.5)

	t.Logf("MomentumChaser: buy=%.4f sell=%.4f | Contrarian: buy=%.4f sell=%.4f",
		momentumProbs.Buy, momentumProbs.Sell, contrarianProbs.Buy, contrarianProbs.Sell)

	if momentumProbs.Buy <= momentumProbs.Sell {
		t.Fatalf("expected MomentumChaser to lean buy on positive momentum, got buy=%.4f sell=%.4f", momentumProbs.Buy, momentumProbs.Sell)
	}
	if contrarianProbs.Sell <= contrarianProbs.Buy {
		t.Fatalf("expected Contrarian to lean sell on positive momentum, got buy=%.4f sell=%.4f", contrarianProbs.Buy, contrarianProbs.Sell)
	}
}

// TestPanicProneAsymmetry confirms PanicProne reacts far more strongly
// to a sharp drop than to an equally-sized rise — the defining
// asymmetric property.
func TestPanicProneAsymmetry(t *testing.T) {
	drop := Signal{Momentum: -0.05}
	rise := Signal{Momentum: 0.05}

	dropProbs := ReactionFor(PanicProne)(drop, 0.5)
	riseProbs := ReactionFor(PanicProne)(rise, 0.5)

	t.Logf("PanicProne on drop: sell=%.4f | on equal rise: buy=%.4f", dropProbs.Sell, riseProbs.Buy)

	if dropProbs.Sell <= riseProbs.Buy {
		t.Fatalf("expected PanicProne's sell reaction to a drop to exceed its buy reaction to an equal rise, got sell=%.4f buy=%.4f", dropProbs.Sell, riseProbs.Buy)
	}
}

// TestTierWindowsDiffer confirms Beginner and Pro bots use genuinely
// different aggregator window sizes, the mechanism underlying
// multi-timeframe behavior.
func TestTierWindowsDiffer(t *testing.T) {
	comp := makeFlatCompany("SYN", 100)
	beginnerAcct := market.NewAccount("b1", 10000, 2, 0.2)
	proAcct := market.NewAccount("p1", 10000, 2, 0.2)

	beginner := NewSimulatedRetailTrader("b1", beginnerAcct, []*company.Company{comp}, Beginner, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(1)))
	pro := NewSimulatedRetailTrader("p1", proAcct, []*company.Company{comp}, Pro, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(1)))

	beginnerWindow := beginner.state["SYN"].agg.WindowSize()
	proWindow := pro.state["SYN"].agg.WindowSize()

	if beginnerWindow >= proWindow {
		t.Fatalf("expected Beginner's aggregator window (%d) to be smaller than Pro's (%d)", beginnerWindow, proWindow)
	}
	t.Logf("Beginner window=%d ticks/bar, Pro window=%d ticks/bar", beginnerWindow, proWindow)
}

// TestWatchlistSizeWithinBounds confirms every bot's watchlist size
// falls within the requested [min, max] range.
func TestWatchlistSizeWithinBounds(t *testing.T) {
	universe := makeTestUniverse()
	pool := GenerateWatchlistPool(500, 42, universe, 2, 4, 500, 5000)

	for _, bot := range pool {
		size := len(bot.watchlist)
		if size < 2 || size > 4 {
			t.Fatalf("expected watchlist size in [2,4], got %d for bot %s", size, bot.ID())
		}
	}
}

// TestAttentionSkewConcentratesOnSmallCaps confirms GenerateWatchlistPool's
// attention weighting actually produces the intended concentration:
// small-cap companies (higher attentionScore) should appear on
// meaningfully more watchlists than the mega-cap company, across a
// large enough pool for the skew to be statistically clear.
func TestAttentionSkewConcentratesOnSmallCaps(t *testing.T) {
	universe := makeTestUniverse()
	pool := GenerateWatchlistPool(3000, 7, universe, 2, 3, 500, 5000)

	coverage := make(map[string]int)
	for _, bot := range pool {
		for symbol := range bot.watchlist {
			coverage[symbol]++
		}
	}

	t.Logf("watchlist coverage counts: %v", coverage)

	if coverage["SMLA"] <= coverage["MEGA"] {
		t.Fatalf("expected small-cap SMLA to appear on more watchlists than mega-cap MEGA, got SMLA=%d MEGA=%d",
			coverage["SMLA"], coverage["MEGA"])
	}
}

// TestGenerateWatchlistPoolDeterministic confirms two pools generated
// from the same seed are identical in composition (tier/style/
// archetype/watchlist), and that the pool is genuinely varied, not a
// degenerate all-one-type population.
func TestGenerateWatchlistPoolDeterministic(t *testing.T) {
	universe := makeTestUniverse()

	poolA := GenerateWatchlistPool(1000, 42, universe, 1, 3, 500, 5000)
	poolB := GenerateWatchlistPool(1000, 42, universe, 1, 3, 500, 5000)

	if len(poolA) != 1000 {
		t.Fatalf("expected 1000 bots, got %d", len(poolA))
	}

	tierCounts := make(map[SkillTier]int)
	styleCounts := make(map[InformationStyle]int)
	archetypeCounts := make(map[Archetype]int)

	for i, bot := range poolA {
		tierCounts[bot.tier]++
		styleCounts[bot.style]++
		archetypeCounts[bot.archetype]++

		other := poolB[i]
		if bot.tier != other.tier || bot.style != other.style || bot.archetype != other.archetype {
			t.Fatalf("bot %d differs between identically-seeded pools", i)
		}
		if len(bot.watchlist) != len(other.watchlist) {
			t.Fatalf("bot %d watchlist size differs between identically-seeded pools: %d vs %d", i, len(bot.watchlist), len(other.watchlist))
		}
		for symbol := range bot.watchlist {
			if !other.Watches(symbol) {
				t.Fatalf("bot %d watches %s in pool A but not in identically-seeded pool B", i, symbol)
			}
		}
	}

	t.Logf("tier counts: %v", tierCounts)
	t.Logf("style counts: %v", styleCounts)
	t.Logf("archetype counts: %v", archetypeCounts)

	for _, tier := range []SkillTier{Beginner, Intermediate, Pro} {
		if tierCounts[tier] == 0 {
			t.Fatalf("expected at least one bot of tier %s in a 1000-bot pool", tier)
		}
	}
	if tierCounts[Beginner] <= tierCounts[Pro] {
		t.Fatalf("expected Beginner to outnumber Pro (retail population skew), got beginner=%d pro=%d", tierCounts[Beginner], tierCounts[Pro])
	}
}
