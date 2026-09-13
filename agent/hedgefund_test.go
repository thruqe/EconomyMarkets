package agent

import (
	"math/rand"
	"testing"

	"economy/company"
	"economy/market"
)

// makeTestCompany builds a company with a fixed, known starting value
// (both TrueValue and ReportedValue start in sync at this number — see
// company.GenerateCompany) and no jump/drift/reporting-noise risk, so
// tests can control the reported figure agents actually read
// precisely, rather than fighting stochastic noise.
func makeTestCompany(symbol string, reportedValue float64) *company.Company {
	used := make(map[string]bool)
	params := company.DefaultGenerationParams()
	params.BaseJumpParams.LambdaDown = 0
	params.BaseJumpParams.LambdaUp = 0
	params.StartingValueMin = reportedValue
	params.StartingValueMax = reportedValue

	c := company.GenerateCompany(company.InformationTechnology, company.MidCap, params, rand.New(rand.NewSource(1)), used)
	c.Symbol = symbol
	return c
}

// TestHedgeFundIgnoresNoiseLevelMispricing confirms the fund does
// nothing when the gap between fundamentals and price is below
// MinTradeThreshold.
func TestHedgeFundIgnoresNoiseLevelMispricing(t *testing.T) {
	comp := makeTestCompany("SYN", 100.0)
	acct := market.NewAccount("fund1", 1_000_000, 5.0, 0.10)
	fund := NewHedgeFund("fund1", acct, comp)

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 100.90, Quantity: 1000})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 101.10, Quantity: 1000})

	state := market.NewMarketState("SYN", 1, book, 5, nil, 0)

	orders := fund.NextOrders(state)
	if len(orders) != 0 {
		t.Fatalf("expected no orders for a ~1%% mispricing (below MinTradeThreshold), got %d", len(orders))
	}
}

// TestHedgeFundBuysWhenUndervalued confirms the fund goes long (buys)
// when TrueValue is meaningfully above market price.
func TestHedgeFundBuysWhenUndervalued(t *testing.T) {
	comp := makeTestCompany("SYN", 150.0) // true value well above market
	acct := market.NewAccount("fund1", 1_000_000, 5.0, 0.10)
	fund := NewHedgeFund("fund1", acct, comp)

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 99.90, Quantity: 5000})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 100.10, Quantity: 5000})

	state := market.NewMarketState("SYN", 1, book, 5, nil, 0)

	orders := fund.NextOrders(state)
	if len(orders) != 1 {
		t.Fatalf("expected exactly one order for a large undervaluation, got %d", len(orders))
	}
	if orders[0].Side != market.Buy {
		t.Fatalf("expected a buy order when undervalued, got side=%v", orders[0].Side)
	}
	if !orders[0].IsMarket {
		t.Fatalf("expected a market order for decisive conviction-driven execution")
	}
}

// TestHedgeFundSellsWhenOvervalued confirms the symmetric case: the
// fund shorts (sells) when TrueValue is meaningfully below price.
func TestHedgeFundSellsWhenOvervalued(t *testing.T) {
	comp := makeTestCompany("SYN", 70.0) // true value well below market
	acct := market.NewAccount("fund1", 1_000_000, 5.0, 0.10)
	fund := NewHedgeFund("fund1", acct, comp)

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 99.90, Quantity: 5000})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 100.10, Quantity: 5000})

	state := market.NewMarketState("SYN", 1, book, 5, nil, 0)

	orders := fund.NextOrders(state)
	if len(orders) != 1 {
		t.Fatalf("expected exactly one order for a large overvaluation, got %d", len(orders))
	}
	if orders[0].Side != market.Sell {
		t.Fatalf("expected a sell order when overvalued, got side=%v", orders[0].Side)
	}
}

// TestHedgeFundConvictionScalesWithMispricing confirms that a larger
// mispricing produces a larger order size than a smaller one, all
// else equal — the core conviction-sizing behavior.
func TestHedgeFundConvictionScalesWithMispricing(t *testing.T) {
	book := func() *market.OrderBook {
		b := market.NewOrderBook()
		b.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 99.90, Quantity: 100000})
		b.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 100.10, Quantity: 100000})
		return b
	}

	// Moderate mispricing (~10%).
	compModerate := makeTestCompany("SYN", 110.0)
	acctModerate := market.NewAccount("fund1", 1_000_000, 5.0, 0.10)
	fundModerate := NewHedgeFund("fund1", acctModerate, compModerate)
	stateModerate := market.NewMarketState("SYN", 1, book(), 5, nil, 0)
	ordersModerate := fundModerate.NextOrders(stateModerate)

	// Large mispricing (~40%, beyond full conviction threshold).
	compLarge := makeTestCompany("SYN", 140.0)
	acctLarge := market.NewAccount("fund1", 1_000_000, 5.0, 0.10)
	fundLarge := NewHedgeFund("fund1", acctLarge, compLarge)
	stateLarge := market.NewMarketState("SYN", 1, book(), 5, nil, 0)
	ordersLarge := fundLarge.NextOrders(stateLarge)

	if len(ordersModerate) != 1 || len(ordersLarge) != 1 {
		t.Fatalf("expected exactly one order in each case, got moderate=%d large=%d", len(ordersModerate), len(ordersLarge))
	}

	t.Logf("moderate mispricing order qty: %.2f, large mispricing order qty: %.2f",
		ordersModerate[0].Quantity, ordersLarge[0].Quantity)

	if ordersLarge[0].Quantity <= ordersModerate[0].Quantity {
		t.Fatalf("expected larger mispricing to produce larger order size, got moderate=%.2f large=%.2f",
			ordersModerate[0].Quantity, ordersLarge[0].Quantity)
	}
}

// TestHedgeFundConvergesPriceTowardFairValue is the real end-to-end
// proof: run a fund against a live order book with resting liquidity
// on both sides, let it trade for several ticks, and confirm the
// book's price actually moves toward the company's TrueValue as a
// result — not just that an order gets generated, but that the
// mechanism it was built for (closing a real mispricing by walking the
// book) actually happens.
func TestHedgeFundConvergesPriceTowardFairValue(t *testing.T) {
	comp := makeTestCompany("SYN", 150.0) // reported value meaningfully above starting market price; the fund's own perceived value (after its SkepticismDiscount haircut) sits lower than 150, so price is expected to converge toward that discounted level, not all the way to 150
	acct := market.NewAccount("fund1", 10_000_000, 5.0, 0.10)
	fund := NewHedgeFund("fund1", acct, comp)

	book := market.NewOrderBook()
	// Seed enough resting depth that even a single max-conviction,
	// fully-leveraged order from this fund cannot exhaust one side of
	// the book in a single tick. With $10M cash and 5x leverage, this
	// fund's ceiling is ~$50M of buying power (conviction-scaled down
	// from there); at ~100/share that's up to ~500,000 units of
	// demand, so each side needs comfortably more resting quantity
	// than that for the book to still have a valid price after each
	// tick's fill.
	for i := 0; i < 40; i++ {
		bidPrice := 99.90 - float64(i)*0.10
		askPrice := 100.10 + float64(i)*0.10
		book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: bidPrice, Quantity: 20000})
		book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: askPrice, Quantity: 20000})
	}

	startMid, ok := book.MidPrice()
	if !ok {
		t.Fatalf("expected a valid starting mid price from seeded book")
	}

	for tick := 0; tick < 5; tick++ {
		state := market.NewMarketState("SYN", tick, book, 5, nil, 0)
		orders := fund.NextOrders(state)
		for _, o := range orders {
			fills := book.Submit(o)
			// Settle fills into the fund's account so subsequent
			// ticks see its updated position (mirrors what a real sim
			// loop's settlement step would do).
			for _, f := range fills {
				if f.TakerAgentID != fund.ID() {
					continue
				}
				market.ApplySettledFill(acct, "SYN", o.Side, f.Quantity, f.Price)
			}
		}

		// Guard against silently treating an exhausted book as a
		// valid price: if the fund's order pace outstrips seeded
		// depth, that's a scenario worth failing loudly on rather
		// than comparing against a meaningless zero mid.
		if _, ok := book.MidPrice(); !ok {
			t.Fatalf("book has no valid mid price after tick %d — one side was fully exhausted; seed more depth or reduce fund conviction for this test", tick)
		}
	}

	endMid, ok := book.MidPrice()
	if !ok {
		t.Fatalf("expected a valid ending mid price")
	}
	t.Logf("mid price: start=%.4f end=%.4f, true value=%.4f", startMid, endMid, comp.TrueValue)

	if endMid <= startMid {
		t.Fatalf("expected fund buying to push price up toward true value, start=%.4f end=%.4f", startMid, endMid)
	}
	if endMid > comp.TrueValue {
		t.Fatalf("expected price to move toward but not overshoot true value in this scenario, end=%.4f true=%.4f", endMid, comp.TrueValue)
	}
}
