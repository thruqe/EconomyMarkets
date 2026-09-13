package retail

import (
	"math/rand"
	"testing"

	"economy/company"
	"economy/market"
)

func makeFlatCompany(symbol string, value float64) *company.Company {
	used := make(map[string]bool)
	params := company.DefaultGenerationParams()
	params.BaseJumpParams.LambdaDown = 0
	params.BaseJumpParams.LambdaUp = 0
	params.StartingValueMin = value
	params.StartingValueMax = value

	c := company.GenerateCompany(company.InformationTechnology, company.MidCap, params, rand.New(rand.NewSource(1)), used)
	c.Symbol = symbol
	c.TrueValue = value
	c.ReportedValue = value
	return c
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

// TestStopLossTriggersOnBreach directly constructs a bot with an open
// long position and an active stop, then confirms a price breach
// produces a forced-exit market order — the core sweep-enabling
// mechanic.
func TestStopLossTriggersOnBreach(t *testing.T) {
	comp := makeFlatCompany("SYN", 100)
	acct := market.NewAccount("bot1", 10000, 2, 0.2)
	bot := NewSimulatedRetailTrader("bot1", acct, comp, Beginner, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(1)))

	market.ApplySettledFill(acct, "SYN", market.Buy, 10, 100)
	bot.stop = stopLoss{active: true, triggerAt: 95, isLong: true}

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 94, Quantity: 1000})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 96, Quantity: 1000})

	state := market.NewMarketState("SYN", 1, book, 5, nil, 0)
	// Price below trigger.
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
	if bot.stop.active {
		t.Fatalf("expected the stop to be deactivated after triggering")
	}
}

// TestStopLossDoesNotTriggerAbovePrice confirms a long position's
// stop stays inactive while price remains above the trigger.
func TestStopLossDoesNotTriggerAbovePrice(t *testing.T) {
	comp := makeFlatCompany("SYN", 100)
	acct := market.NewAccount("bot1", 10000, 2, 0.2)
	bot := NewSimulatedRetailTrader("bot1", acct, comp, Beginner, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(1)))

	market.ApplySettledFill(acct, "SYN", market.Buy, 10, 100)
	bot.stop = stopLoss{active: true, triggerAt: 95, isLong: true}

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 97, Quantity: 1000})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 99, Quantity: 1000})

	state := market.NewMarketState("SYN", 1, book, 5, nil, 0)
	state.Mid = 98
	state.HasMid = true

	order := bot.checkStopLoss(98)
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
	for i := range n {
		acct := market.NewAccount("bot"+itoa(i), 10000, 2, 0.2)
		bot := NewSimulatedRetailTrader("bot"+itoa(i), acct, comp, Beginner, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(int64(i))))
		// Force every bot long from the same entry, with a stop placed
		// via the real RoundNumberStop logic — this is what should
		// cluster them at the same trigger level.
		market.ApplySettledFill(acct, "SYN", market.Buy, 10, 100)
		bot.placeStop(100, true, nil)
		crowd = append(crowd, bot)
	}

	// Confirm the stops actually clustered at (or near) the same level
	// before testing the cascade, since the cascade proof depends on
	// that clustering being real.
	triggerLevels := make(map[float64]int)
	for _, b := range crowd {
		triggerLevels[b.stop.triggerAt]++
	}
	t.Logf("distinct stop trigger levels among %d bots: %d", n, len(triggerLevels))
	if len(triggerLevels) > 3 {
		t.Fatalf("expected RoundNumberStop placement to cluster stops into very few distinct levels, got %d distinct levels", len(triggerLevels))
	}

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 90, Quantity: 100000})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 110, Quantity: 100000})

	// Price sweeps down through the clustered stop level.
	state := market.NewMarketState("SYN", 1, book, 5, nil, 0)
	state.Mid = 92
	state.HasMid = true

	triggeredCount := 0
	for _, b := range crowd {
		if order := b.checkStopLoss(92); order != nil {
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
	bot := NewSimulatedRetailTrader("bot1", acct, comp, Pro, FundamentalOnly, MomentumChaser, rand.New(rand.NewSource(1)))

	book := market.NewOrderBook()
	for i := range 10 {
		book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 99 - float64(i), Quantity: 5000})
		book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 101 + float64(i), Quantity: 5000})
	}

	// Warm up the bot's aggregator with enough ticks to get a
	// technicals-window-sized bar history (Pro tier needs
	// AggregatorWindow=50 ticks per bar); state.Mid held flat so any
	// resulting decision is attributable to the fundamental gap, not
	// technical momentum.
	var lastOrders []*market.Order
	for i := range 60 {
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

	beginner := NewSimulatedRetailTrader("b1", beginnerAcct, comp, Beginner, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(1)))
	pro := NewSimulatedRetailTrader("p1", proAcct, comp, Pro, TechnicalOnly, MomentumChaser, rand.New(rand.NewSource(1)))

	if beginner.agg.WindowSize() >= pro.agg.WindowSize() {
		t.Fatalf("expected Beginner's aggregator window (%d) to be smaller than Pro's (%d)",
			beginner.agg.WindowSize(), pro.agg.WindowSize())
	}
	t.Logf("Beginner window=%d ticks/bar, Pro window=%d ticks/bar", beginner.agg.WindowSize(), pro.agg.WindowSize())
}

// TestGenerateCrowdProducesVariedPopulation confirms GenerateCrowd
// actually produces a mix of tiers/styles/archetypes, not a
// degenerate all-one-type population, and is deterministic given a
// fixed seed.
func TestGenerateCrowdProducesVariedPopulation(t *testing.T) {
	comp := makeFlatCompany("SYN", 100)

	crowdA := GenerateCrowd(2000, 42, comp, 500, 5000)
	crowdB := GenerateCrowd(2000, 42, comp, 500, 5000)

	if len(crowdA) != 2000 {
		t.Fatalf("expected 2000 bots, got %d", len(crowdA))
	}

	tierCounts := make(map[SkillTier]int)
	styleCounts := make(map[InformationStyle]int)
	archetypeCounts := make(map[Archetype]int)

	for i, bot := range crowdA {
		tierCounts[bot.tier]++
		styleCounts[bot.style]++
		archetypeCounts[bot.archetype]++

		// Determinism check against crowdB.
		if bot.tier != crowdB[i].tier || bot.style != crowdB[i].style || bot.archetype != crowdB[i].archetype {
			t.Fatalf("bot %d differs between identically-seeded crowds", i)
		}
	}

	t.Logf("tier counts: %v", tierCounts)
	t.Logf("style counts: %v", styleCounts)
	t.Logf("archetype counts: %v", archetypeCounts)

	for _, tier := range []SkillTier{Beginner, Intermediate, Pro} {
		if tierCounts[tier] == 0 {
			t.Fatalf("expected at least one bot of tier %s in a 2000-bot crowd", tier)
		}
	}
	if tierCounts[Beginner] <= tierCounts[Pro] {
		t.Fatalf("expected Beginner to outnumber Pro (retail population skew), got beginner=%d pro=%d", tierCounts[Beginner], tierCounts[Pro])
	}
}
