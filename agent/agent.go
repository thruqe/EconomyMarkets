// Package agent contains institutional-style market participants
// (market makers, hedge funds, banks, large investors). Each concrete
// type is a fully independent implementation of market.OrderSource,
// with its own private decision logic, its own reasons to trade or
// sit out, and no shared status/lifecycle vocabulary between types —
// that mirrors real market structure, where a hedge fund and a market
// maker have no relationship to each other beyond both observing and
// acting on the same public order book.
//
// This file holds only genuinely reusable math/helpers that multiple
// agent types independently need — not a shared behavioral contract.
// Nothing here is required by market.OrderSource, and nothing here
// implies any agent type must use it.
package agent

import "economy/market"

// NetInventory reads an account's current signed position size for a
// given symbol directly from its market.Account: positive for a net
// long, negative for a net short, zero if flat or the symbol isn't
// held. This is the same computation MarketMaker needs privately, and
// every other institutional agent type (hedge funds, banks, large
// investors) will need identically, so it lives here once rather than
// being reimplemented per type.
//
// This deliberately reads from market.Account rather than from
// MarketState — an agent's own inventory is private information about
// that agent, never part of the public market snapshot, exactly as no
// real participant can see another's book.
func NetInventory(acct *market.Account, symbol string) float64 {
	pos, ok := acct.Positions[symbol]
	if !ok {
		return 0
	}
	if pos.Side == market.Short {
		return -pos.Quantity
	}
	return pos.Quantity
}

// EMA is a minimal exponential moving average, reusable by any agent
// type that wants a lagging estimate of some observed quantity
// (fair value, a slower volatility estimate, etc.) rather than
// reacting to the raw instantaneous value each tick. Kept as a small
// standalone type rather than baked into any one agent's struct, so
// e.g. a HedgeFund can track its own fair-value EMA independently
// from a MarketMaker's, with different smoothing.
type EMA struct {
	Alpha       float64 // smoothing factor: higher = reacts faster to new observations
	value       float64
	initialized bool
}

// NewEMA constructs an EMA with the given smoothing factor. Alpha
// should be in (0, 1]; values closer to 1 track the input more
// closely (less lag, less smoothing), values closer to 0 lag more
// but smooth out noise more.
func NewEMA(alpha float64) *EMA {
	return &EMA{Alpha: alpha}
}

// Update feeds in a new observation and returns the updated EMA
// value. The first call initializes the EMA directly to the observed
// value (no prior estimate to blend with).
func (e *EMA) Update(observation float64) float64 {
	if !e.initialized {
		e.value = observation
		e.initialized = true
		return e.value
	}
	e.value = e.Alpha*observation + (1-e.Alpha)*e.value
	return e.value
}

// Value returns the current EMA estimate without updating it.
func (e *EMA) Value() float64 {
	return e.value
}

// Initialized reports whether Update has been called at least once —
// useful for an agent that shouldn't quote/trade until it has a real
// fair-value estimate rather than a zero-value placeholder.
func (e *EMA) Initialized() bool {
	return e.initialized
}

// Variance converts a volatility figure (as carried on
// market.MarketState.RecentVolatility, typically a standard deviation
// of returns) into variance, since several pricing/risk formulas
// (inventory skew, spread widening) are expressed in terms of
// variance rather than volatility directly. Trivial, but centralizing
// it means every agent type applies the same convention rather than
// each squaring RecentVolatility inline and risking inconsistency if
// the convention ever changes (e.g. to annualized vol).
func Variance(recentVolatility float64) float64 {
	return recentVolatility * recentVolatility
}
