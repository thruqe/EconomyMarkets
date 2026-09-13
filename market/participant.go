package market

// MarketState is the public, observable snapshot of a single instrument's
// market — everything any participant could plausibly see by watching
// the market itself. It is the *only* channel through which participants
// influence each other: a hedge fund's trade changes the book, which
// changes the next MarketState, which a retail trader's strategy reacts
// to. No participant type ever sees another participant directly.
type MarketState struct {
	Symbol string
	Tick   int

	BestBid float64
	HasBid  bool
	BestAsk float64
	HasAsk  bool

	Mid    float64
	HasMid bool

	// BidDepth/AskDepth are cumulative resting quantity within some
	// number of top price levels (see OrderBook.DepthAtLevels) — a
	// proxy for how much size the market can currently absorb before
	// a taker starts walking further into the book.
	BidDepth float64
	AskDepth float64

	// PriceHistory is recent mid-price observations, oldest first,
	// most recent last. Used by any strategy that reacts to trend or
	// realized volatility rather than just the current tick's price.
	PriceHistory []float64

	// RecentVolatility is a precomputed measure (e.g. stddev of
	// recent returns) over PriceHistory's window, so every
	// participant works from the same volatility figure rather than
	// each recomputing it slightly differently.
	RecentVolatility float64
}

// NewMarketState builds a MarketState snapshot directly from an
// OrderBook plus whatever price history/volatility the sim loop is
// tracking externally (the book itself has no memory of past ticks).
func NewMarketState(symbol string, tick int, book *OrderBook, depthLevels int, priceHistory []float64, recentVolatility float64) MarketState {
	bid, hasBid := book.BestBid()
	ask, hasAsk := book.BestAsk()
	mid, hasMid := book.MidPrice()
	bidDepth, askDepth := book.DepthAtLevels(depthLevels)

	return MarketState{
		Symbol:           symbol,
		Tick:             tick,
		BestBid:          bid,
		HasBid:           hasBid,
		BestAsk:          ask,
		HasAsk:           hasAsk,
		Mid:              mid,
		HasMid:           hasMid,
		BidDepth:         bidDepth,
		AskDepth:         askDepth,
		PriceHistory:     priceHistory,
		RecentVolatility: recentVolatility,
	}
}

// OrderSource is the only thing market itself requires of a
// participant. It says nothing about what kind of participant this is,
// how it decides, or why it might choose not to trade this tick — that
// is entirely private to whichever package implements it (agent,
// retail, or anything else). market never imports agent or retail, and
// agent and retail never import each other; MarketState is the only
// shared surface, exactly mirroring how real participants only ever
// observe the market itself, never each other directly.
//
// NextOrders returns every order this participant wants to submit this
// tick — zero, one, or several. Returning an empty or nil slice means
// sitting out entirely, a valid and ordinary outcome, not an error.
// Plural orders are not a special case for market makers alone: any
// participant may legitimately want to act on more than one order in
// a single tick (a market maker quoting both sides, a fund placing an
// entry alongside a protective order, and so on), so the contract
// supports it uniformly rather than special-casing a single type.
type OrderSource interface {
	ID() string
	NextOrders(state MarketState) []*Order
}
