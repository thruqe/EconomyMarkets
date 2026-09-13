package market

import (
	"errors"
	"sort"
)

// Side represents the direction of an order.
type Side int

const (
	Buy Side = iota
	Sell
)

// Order is a resting or incoming order in the book.
// AgentID links back to whichever agent (retail, fund, market maker, etc.)
// placed it — the book itself doesn't care what kind of agent that is.
type Order struct {
	ID       uint64
	AgentID  string
	Side     Side
	Price    float64 // ignored for market orders
	Quantity float64
	IsMarket bool
}

// Fill represents one match between an incoming order and a resting order.
type Fill struct {
	TakerAgentID string
	MakerAgentID string
	Price        float64
	Quantity     float64
}

// priceLevel holds all resting quantity at a single price, FIFO by arrival.
type priceLevel struct {
	price  float64
	orders []*Order // queue, front = oldest = filled first
}

// OrderBook is a lightweight limit order book: discrete price levels,
// each holding a FIFO queue of resting orders. This is intentionally
// simple (no full L3 depth) — enough to get realistic slippage and
// price-impact behavior without the complexity of a production matching engine.
type OrderBook struct {
	bids        []*priceLevel // sorted descending by price (best bid first)
	asks        []*priceLevel // sorted ascending by price (best ask first)
	nextOrderID uint64
}

func NewOrderBook() *OrderBook {
	return &OrderBook{}
}

var ErrEmptyBook = errors.New("no liquidity available on the requested side")

// BestBid returns the highest resting buy price, or (0, false) if none.
func (b *OrderBook) BestBid() (float64, bool) {
	if len(b.bids) == 0 {
		return 0, false
	}
	return b.bids[0].price, true
}

// BestAsk returns the lowest resting sell price, or (0, false) if none.
func (b *OrderBook) BestAsk() (float64, bool) {
	if len(b.asks) == 0 {
		return 0, false
	}
	return b.asks[0].price, true
}

// Spread returns ask-bid, or (0, false) if either side is empty.
func (b *OrderBook) Spread() (float64, bool) {
	bid, ok1 := b.BestBid()
	ask, ok2 := b.BestAsk()
	if !ok1 || !ok2 {
		return 0, false
	}
	return ask - bid, true
}

// MidPrice returns (bid+ask)/2, or (0, false) if either side is empty.
func (b *OrderBook) MidPrice() (float64, bool) {
	bid, ok1 := b.BestBid()
	ask, ok2 := b.BestAsk()
	if !ok1 || !ok2 {
		return 0, false
	}
	return (bid + ask) / 2, true
}

// AddLimitOrder places a resting limit order on the book without matching.
// Use Submit() for orders that should attempt to match immediately —
// this is exposed separately mainly for seeding initial liquidity
// (e.g. a market maker populating both sides at the start of a tick).
func (b *OrderBook) AddLimitOrder(o *Order) {
	b.nextOrderID++
	o.ID = b.nextOrderID
	if o.Side == Buy {
		b.bids = insertLevel(b.bids, o, true)
	} else {
		b.asks = insertLevel(b.asks, o, false)
	}
}

func insertLevel(levels []*priceLevel, o *Order, descending bool) []*priceLevel {
	for _, lvl := range levels {
		if lvl.price == o.Price {
			lvl.orders = append(lvl.orders, o)
			return levels
		}
	}
	levels = append(levels, &priceLevel{price: o.Price, orders: []*Order{o}})
	sort.Slice(levels, func(i, j int) bool {
		if descending {
			return levels[i].price > levels[j].price
		}
		return levels[i].price < levels[j].price
	})
	return levels
}

// Submit processes an incoming order (market or aggressive limit),
// walking the opposite side of the book and returning every fill
// generated. This is where slippage and price impact come from directly:
// a large order consumes multiple price levels, so the average fill
// price gets worse the deeper it has to walk — no separate formula needed.
//
// Any unfilled remainder of a limit order is added to the book as
// resting liquidity. Unfilled remainder of a market order is dropped
// (in a real market this is roughly "fill or kill" behavior for the
// portion that couldn't be matched; adjust here if you want your sim
// to model failed/partial market orders differently).
func (b *OrderBook) Submit(o *Order) []Fill {
	var fills []Fill
	remaining := o.Quantity

	opposite := &b.asks
	if o.Side == Sell {
		opposite = &b.bids
	}

	for remaining > 0 && len(*opposite) > 0 {
		lvl := (*opposite)[0]

		// price limit check: for non-market orders, stop if the level
		// no longer satisfies the limit price.
		if !o.IsMarket {
			if o.Side == Buy && lvl.price > o.Price {
				break
			}
			if o.Side == Sell && lvl.price < o.Price {
				break
			}
		}

		for len(lvl.orders) > 0 && remaining > 0 {
			resting := lvl.orders[0]
			qty := min(remaining, resting.Quantity)

			fills = append(fills, Fill{
				TakerAgentID: o.AgentID,
				MakerAgentID: resting.AgentID,
				Price:        lvl.price,
				Quantity:     qty,
			})

			resting.Quantity -= qty
			remaining -= qty

			if resting.Quantity <= 0 {
				lvl.orders = lvl.orders[1:]
			}
		}

		if len(lvl.orders) == 0 {
			*opposite = (*opposite)[1:]
		}
	}

	// resting remainder for limit orders only
	if remaining > 0 && !o.IsMarket {
		o.Quantity = remaining
		b.AddLimitOrder(o)
	}

	return fills
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// DepthAtLevels returns cumulative quantity available within the given
// number of price levels on each side — useful for agents that want to
// gauge liquidity before sizing an order (e.g. a hedge fund checking
// whether the book can absorb its intended trade without excessive slippage).
func (b *OrderBook) DepthAtLevels(n int) (bidQty, askQty float64) {
	for i := 0; i < n && i < len(b.bids); i++ {
		for _, o := range b.bids[i].orders {
			bidQty += o.Quantity
		}
	}
	for i := 0; i < n && i < len(b.asks); i++ {
		for _, o := range b.asks[i].orders {
			askQty += o.Quantity
		}
	}
	return bidQty, askQty
}
