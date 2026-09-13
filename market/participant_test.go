package market

import "testing"

// dummyParticipant is a minimal stand-in to prove OrderSource works as
// a pure black-box contract. It could equally be a hedge fund from
// package agent or a bot from package retail — market has no way to
// tell, and that's the point.
type dummyParticipant struct {
	id       string
	buysWhen float64 // buys if mid is below this, else does nothing
}

func (d *dummyParticipant) ID() string { return d.id }

func (d *dummyParticipant) NextOrders(state MarketState) []*Order {
	if !state.HasMid {
		return nil
	}
	if state.Mid < d.buysWhen {
		return []*Order{{AgentID: d.id, Side: Buy, Quantity: 10, IsMarket: true}}
	}
	return nil // sitting out is ordinary, not an error
}

func TestOrderSourceContract(t *testing.T) {
	book := NewOrderBook()
	book.AddLimitOrder(&Order{AgentID: "mm", Side: Sell, Price: 99.00, Quantity: 100})
	book.AddLimitOrder(&Order{AgentID: "mm", Side: Buy, Price: 98.50, Quantity: 100})

	// A flat slice of OrderSource — the sim loop's actual shape.
	// Deliberately only one concrete type here since agent/retail
	// don't exist from market's point of view; this just proves the
	// interface is satisfiable and drivable from a generic snapshot.
	participants := []OrderSource{
		&dummyParticipant{id: "trader_a", buysWhen: 100.00}, // will buy, mid is 98.75
		&dummyParticipant{id: "trader_b", buysWhen: 50.00},  // will sit out
	}

	state := NewMarketState("SYN", 1, book, 5, nil, 0)

	var orders []*Order
	for _, p := range participants {
		orders = append(orders, p.NextOrders(state)...)
	}

	if len(orders) != 1 {
		t.Fatalf("expected exactly 1 order (trader_a only), got %d", len(orders))
	}
	if orders[0].AgentID != "trader_a" {
		t.Fatalf("expected order from trader_a, got %s", orders[0].AgentID)
	}

	fills := book.Submit(orders[0])
	if len(fills) == 0 {
		t.Fatalf("expected trader_a's order to fill against resting ask")
	}
}
