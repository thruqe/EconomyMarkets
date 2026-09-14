// Package sim contains the orchestrator: the tick loop that ties
// every other package together into a runnable simulation. It is the
// only package that imports market, company, technicals, agent, and
// retail all at once — by design, since orchestrating them is
// precisely its job. Every other package remains a sealed box; sim is
// where the seams actually get wired.
package sim

import (
	"math"

	"economy/company"
	"economy/market"
)

// companyMarket bundles one company with everything needed to run its
// market: its own order book, rolling mid-price history (the book
// itself has no memory of past ticks), and the list of participants
// that are allowed to trade it this run.
type companyMarket struct {
	co   *company.Company
	book *market.OrderBook

	// priceHistory is rolling mid-price observations, oldest first,
	// capped at maxHistory — this is what feeds MarketState's
	// PriceHistory/RecentVolatility fields each tick.
	priceHistory []float64
	maxHistory   int

	// participants is every market.OrderSource polled for orders on
	// this company each tick — a company's market maker, any
	// hedge funds/banks covering it, and every retail bot that
	// watches it (retail bots watching multiple companies appear once
	// per company they watch, which is correct: NextOrders is called
	// once per company per tick for any participant, exactly matching
	// how agent.Bank and retail.SimulatedRetailTrader are already
	// designed to be driven).
	participants []market.OrderSource
}

func newCompanyMarket(co *company.Company, maxHistory int) *companyMarket {
	return &companyMarket{
		co:           co,
		book:         market.NewOrderBook(),
		maxHistory:   maxHistory,
		priceHistory: make([]float64, 0, maxHistory),
	}
}

// recordPrice appends the current mid-price (if the book has one) to
// rolling history, dropping the oldest entry once maxHistory is
// exceeded. Called once per tick after all of this tick's trading has
// settled, so the next tick's MarketState reflects what just happened.
func (cm *companyMarket) recordPrice() {
	mid, ok := cm.book.MidPrice()
	if !ok {
		return // book has no valid price yet (e.g. one side empty) — nothing to record
	}
	cm.priceHistory = append(cm.priceHistory, mid)
	if len(cm.priceHistory) > cm.maxHistory {
		cm.priceHistory = cm.priceHistory[len(cm.priceHistory)-cm.maxHistory:]
	}
}

// recentVolatility computes the standard deviation of simple returns
// over the current rolling price history — the measure
// MarketState.RecentVolatility carries, computed once per tick here
// so every participant polled this tick works from the same figure
// rather than each recomputing it slightly differently (the same
// reasoning documented on MarketState.RecentVolatility itself).
// Returns 0 if fewer than 2 price points exist (not enough to compute
// a return, let alone a standard deviation).
func (cm *companyMarket) recentVolatility() float64 {
	if len(cm.priceHistory) < 2 {
		return 0
	}

	returns := make([]float64, 0, len(cm.priceHistory)-1)
	for i := 1; i < len(cm.priceHistory); i++ {
		prev := cm.priceHistory[i-1]
		if prev == 0 {
			continue
		}
		returns = append(returns, (cm.priceHistory[i]-prev)/prev)
	}
	if len(returns) < 2 {
		return 0
	}

	var mean float64
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))

	var variance float64
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	variance /= float64(len(returns))

	return math.Sqrt(variance)
}

// state builds this tick's MarketState from current book state and
// rolling history — the one snapshot every polled participant for
// this company sees this tick.
func (cm *companyMarket) state(tick int, depthLevels int) market.MarketState {
	historyCopy := make([]float64, len(cm.priceHistory))
	copy(historyCopy, cm.priceHistory)
	return market.NewMarketState(cm.co.Symbol, tick, cm.book, depthLevels, historyCopy, cm.recentVolatility())
}
