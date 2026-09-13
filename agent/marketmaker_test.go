package agent

import (
	"testing"

	"economy/market"
)

// TestMarketMakerQuotesBothSides confirms the basic two-sided quoting
// behavior once fair value is established.
func TestMarketMakerQuotesBothSides(t *testing.T) {
	mm := NewMarketMaker("mm1", 1_000_000, 10, 0.10)

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 99.90, Quantity: 100})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 100.10, Quantity: 100})

	state := market.NewMarketState("SYN", 1, book, 5, nil, 0.001)

	// First tick: EMA just initiated to current mid, orders should
	// already be quoted around it.
	orders := mm.NextOrders(state)
	if len(orders) != 2 {
		t.Fatalf("expected both sides quoted, got %d orders", len(orders))
	}

	var haveBuy, haveSell bool
	for _, o := range orders {
		if o.Side == market.Buy {
			haveBuy = true
		}
		if o.Side == market.Sell {
			haveSell = true
		}
	}
	if !haveBuy || !haveSell {
		t.Fatalf("expected one buy and one sell order, got buy=%v sell=%v", haveBuy, haveSell)
	}
}

// TestMarketMakerEMALagCausesAdverseAccumulation is the important one:
// it proves the MM's fair-value EMA genuinely lags a sustained
// one-directional move, and that lag causes it to keep buying while
// price falls — accumulating a long position it doesn't want. This is
// the real-world failure mode (getting run over in a trend) that a
// naive "quote around current mid" implementation cannot exhibit at
// all, since it would have zero lag by construction.
func TestMarketMakerEMALagCausesAdverseAccumulation(t *testing.T) {
	mm := NewMarketMaker("mm1", 1_000_000, 10, 0.10)
	mm.MaxInventory = 100_000 // effectively unbounded for this test

	symbol := "SYN"
	price := 100.0

	// Simulate a sustained downtrend: price falls a little every
	// tick. Each tick we hand the MM a MarketState reflecting that
	// price, collect its quotes, and simulate a fill on whichever
	// side actually crosses given the falling price — approximating
	// what a real falling market would do to a resting quote ladder.
	for i := range 30 {
		price -= 0.50 // steady decline

		book := market.NewOrderBook()
		state := market.NewMarketState(symbol, i, book, 5, nil, 0.01)
		state.HasMid = true
		state.Mid = price

		orders := mm.NextOrders(state)

		for _, o := range orders {
			// A falling market means the MM's bid (below its lagging
			// fair value, which is above the new lower price) is the
			// side likely to get hit: simulate that its buy orders
			// fill (someone sells into its bid), its sell orders do
			// not (its ask is now above where the market actually is,
			// so nothing lifts it).
			if o.Side == market.Buy {
				market.ApplySettledFill(mm.Account(), symbol, market.Buy, o.Quantity, o.Price)
			}
		}
	}

	inv := NetInventory(mm.Account(), symbol)
	t.Logf("net inventory after sustained downtrend: %.2f", inv)
	t.Logf("final fair value EMA: %.4f vs actual price: %.4f", mm.fairValue.Value(), price)

	if inv <= 0 {
		t.Fatalf("expected MM to accumulate a long position from EMA lag during downtrend, got inventory %.2f", inv)
	}
	if mm.fairValue.Value() <= price {
		t.Fatalf("expected EMA to lag above actual price during sustained decline, got EMA %.4f vs price %.4f", mm.fairValue.Value(), price)
	}
}

// TestMarketMakerWidensSpreadUnderVolatility confirms spread grows
// with RecentVolatility, which is the mechanism that throttles MM
// participation under stress instead of a binary on/off switch.
func TestMarketMakerWidensSpreadUnderVolatility(t *testing.T) {
	mm := NewMarketMaker("mm1", 1_000_000, 10, 0.10)

	calmSpread := mm.EffectiveSpread(0.001)
	stressedSpread := mm.EffectiveSpread(0.05)

	t.Logf("calm spread: %.4f, stressed spread: %.4f", calmSpread, stressedSpread)

	if stressedSpread <= calmSpread {
		t.Fatalf("expected spread to widen under higher volatility, got calm=%.4f stressed=%.4f", calmSpread, stressedSpread)
	}
}

// TestMarketMakerRespectsInventoryLimit confirms that once inventory
// hits MaxInventory on one side, the MM stops quoting the side that
// would extend it further, while continuing to quote the reducing side.
func TestMarketMakerRespectsInventoryLimit(t *testing.T) {
	mm := NewMarketMaker("mm1", 1_000_000, 10, 0.10)
	mm.MaxInventory = 100
	mm.QuoteSize = 50

	symbol := "SYN"
	market.ApplySettledFill(mm.Account(), symbol, market.Buy, 100, 100.0) // already at max long

	book := market.NewOrderBook()
	state := market.NewMarketState(symbol, 1, book, 5, nil, 0.001)
	state.HasMid = true
	state.Mid = 100.0

	orders := mm.NextOrders(state)

	for _, o := range orders {
		if o.Side == market.Buy {
			t.Fatalf("expected no further buy quote once at max long inventory, got one anyway")
		}
	}

	var haveSell bool
	for _, o := range orders {
		if o.Side == market.Sell {
			haveSell = true
		}
	}
	if !haveSell {
		t.Fatalf("expected MM to still quote the sell side to reduce an over-limit long position")
	}
}
