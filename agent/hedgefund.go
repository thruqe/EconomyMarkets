package agent

import (
	"economy/company"
	"economy/market"
)

// convictionSizedOrder computes a single conviction-sized order for
// one company against one account, shared by any agent type that
// trades on a fundamentals-vs-price gap (HedgeFund uses this directly
// with riskScale always 1.0; Bank uses it per covered company with a
// riskScale derived from its own drawdown state). Centralizing this
// here means the core mispricing -> conviction -> target-exposure ->
// order mechanism is defined once, not reimplemented per agent type.
//
// This reads target.ReportedValue, not target.TrueValue — no
// participant in this simulation, institutional or retail, has access
// to true intrinsic value, matching how no real analyst or fund
// actually observes a company's true value directly either (see
// package company's documentation on ReportingProfile). What
// distinguishes a sophisticated institutional agent from a retail
// trader reading the same report is not omniscience but skepticism:
// skepticismDiscount (0-1) partially haircuts ReportedValue back
// toward the current market price before computing the perceived
// mispricing, modeling how a real analyst discounts a report that
// looks too good/bad to be true rather than taking it at face value.
// 0 means no discount at all (equivalent to trusting the report
// completely, as retail.FundamentalOnly currently does); values closer
// to 1 mean the agent barely moves off the market's own prior. This
// does not let the agent detect fraud — a sufficiently large,
// well-hidden gap still fools a discounted estimate, exactly as real
// fraud fools real analysts for a time — it only means a sophisticated
// agent's perceived gap, and therefore its conviction and position
// size, is more conservative than retail's for the same report.
//
// riskScale multiplies the resulting target exposure before computing
// the order delta: 1.0 means full conviction-based sizing as designed,
// values below 1.0 shrink the target (used for account-wide
// de-risking), and 0 suppresses new/added risk entirely while still
// allowing existing positions to be reduced toward zero (since a
// riskScale of 0 makes targetValue 0, and delta will call for fully
// unwinding whatever's currently held).
func convictionSizedOrder(
	agentID string,
	acct *market.Account,
	target *company.Company,
	state market.MarketState,
	minTradeThreshold, fullConvictionThreshold, rebalanceThreshold float64,
	riskScale float64,
	skepticismDiscount float64,
) *market.Order {
	if !state.HasMid || state.Mid <= 0 {
		return nil
	}

	marketPrice := state.Mid

	// Haircut ReportedValue partway back toward the market's own
	// price before treating it as this agent's fair-value estimate —
	// the skepticism adjustment. At skepticismDiscount=0 this reduces
	// to trusting ReportedValue completely.
	perceivedValue := target.ReportedValue - skepticismDiscount*(target.ReportedValue-marketPrice)

	mispricing := (perceivedValue - marketPrice) / marketPrice
	absMispricing := absValue(mispricing)

	conv := conviction(absMispricing, minTradeThreshold, fullConvictionThreshold)

	currentPrices := map[string]float64{state.Symbol: marketPrice}
	availableBuyingPower := acct.AvailableBuyingPower(currentPrices)
	currentPositionValue := signedPositionValue(acct, state.Symbol, marketPrice)

	maxTotalExposure := availableBuyingPower + absValue(currentPositionValue)
	targetValue := conv * riskScale * maxTotalExposure
	if mispricing < 0 {
		targetValue = -targetValue
	}

	delta := targetValue - currentPositionValue

	rebalanceFloor := rebalanceThreshold * (availableBuyingPower + absValue(currentPositionValue) + 1)
	if absValue(delta) < rebalanceFloor {
		return nil
	}

	side := market.Buy
	if delta < 0 {
		side = market.Sell
	}
	quantity := absValue(delta) / marketPrice
	if quantity <= 0 {
		return nil
	}

	return &market.Order{
		AgentID:  agentID,
		Side:     side,
		Quantity: quantity,
		IsMarket: true,
	}
}

// conviction maps an absolute mispricing fraction to a 0-1 conviction
// level: 0 below minTradeThreshold, linearly scaling to 1 at
// fullConvictionThreshold and beyond. Standalone function (rather than
// a HedgeFund method) so Bank can reuse the identical mapping with its
// own threshold values.
func conviction(absMispricing, minTradeThreshold, fullConvictionThreshold float64) float64 {
	if absMispricing < minTradeThreshold {
		return 0
	}
	if absMispricing >= fullConvictionThreshold {
		return 1
	}
	span := fullConvictionThreshold - minTradeThreshold
	return (absMispricing - minTradeThreshold) / span
}

// HedgeFund is a conviction-sized value trader: it compares a
// company's reported fundamental value against the market's current
// price (with a skepticism discount — see convictionSizedOrder) and
// trades toward closing that perceived gap. Unlike MarketMaker, which
// wants to stay flat at all times, a HedgeFund deliberately wants
// directional exposure while a mispricing persists, and unwinds as
// price converges back toward its perceived fair value — its target
// position shrinks toward zero as the gap closes, which is what
// drives both entry and exit through one mechanism rather than
// separate code paths.
//
// This is also the first type in this codebase that reads
// company.Company directly — it is the reason the company package
// exists: without an independent fundamental-value signal, there is
// nothing for a value trader to have real conviction about.
type HedgeFund struct {
	id      string
	account *market.Account
	company *company.Company // the single company this fund specializes in

	// MinTradeThreshold is the smallest absolute mispricing (as a
	// fraction of market price) worth acting on at all. Below this,
	// the gap is treated as noise — real funds don't trade on
	// rounding-error-sized discrepancies once transaction costs and
	// risk are considered.
	MinTradeThreshold float64

	// FullConvictionThreshold is the mispricing size (as a fraction of
	// market price) at which the fund considers itself maximally
	// confident and is willing to deploy all of its available buying
	// power toward the position. Mispricings between MinTradeThreshold
	// and FullConvictionThreshold scale conviction linearly between 0
	// and 1.
	FullConvictionThreshold float64

	// RebalanceThreshold is the minimum change (as a fraction of
	// current available buying power) between the fund's existing
	// position and its newly computed target before it bothers
	// submitting an adjusting order. Without this, tiny tick-to-tick
	// noise in ReportedValue or price would cause the fund to
	// constantly submit negligible adjusting orders — real funds
	// don't continuously micro-adjust a position over noise-level
	// changes.
	RebalanceThreshold float64

	// SkepticismDiscount (0-1) partially haircuts ReportedValue back
	// toward market price before this fund treats it as its own
	// fair-value estimate — see convictionSizedOrder's documentation.
	// A HedgeFund defaults to a meaningful, nonzero discount: real
	// funds are professionally skeptical of headline numbers, even
	// though (matching reality) a large enough hidden gap can still
	// mislead a discounted estimate, not just an undiscounted one.
	SkepticismDiscount float64
}

// NewHedgeFund constructs a HedgeFund specialized on a single company.
// A fund covering multiple companies is naturally modeled as multiple
// HedgeFund instances (one per covered company) sharing... nothing —
// deliberately: giving each its own independent market.Account would
// model separately-capitalized desks; sharing one market.Account
// across several HedgeFund instances would model one desk allocating
// shared capital across multiple theses. Both are legitimate setups;
// this constructor supports either by taking the account as a
// parameter rather than always creating a fresh one.
func NewHedgeFund(id string, acct *market.Account, target *company.Company) *HedgeFund {
	return &HedgeFund{
		id:                      id,
		account:                 acct,
		company:                 target,
		MinTradeThreshold:       0.03, // 3% — below this, treated as noise
		FullConvictionThreshold: 0.25, // 25% — at or beyond this, full conviction
		RebalanceThreshold:      0.05, // 5% of available buying power
		SkepticismDiscount:      0.35, // moderate professional skepticism
	}
}

func (h *HedgeFund) ID() string               { return h.id }
func (h *HedgeFund) Account() *market.Account { return h.account }

// NextOrders implements market.OrderSource. Each tick, the fund
// computes its ideal target position value from current conviction,
// compares it against its actual current position value, and submits
// a single market order for the difference if that difference is
// large enough to bother acting on.
func (h *HedgeFund) NextOrders(state market.MarketState) []*market.Order {
	order := convictionSizedOrder(
		h.id, h.account, h.company, state,
		h.MinTradeThreshold, h.FullConvictionThreshold, h.RebalanceThreshold,
		1.0, // full risk scale: a standalone fund isn't subject to Bank-style account-wide de-risking
		h.SkepticismDiscount,
	)
	if order == nil {
		return nil
	}
	return []*market.Order{order}
}

// signedPositionValue returns the current position's value, signed by
// direction: positive for a long, negative for a short, zero if flat.
// Distinct from NetInventory (which returns signed *quantity*) since
// callers here need signed *value* to compare directly against a
// target exposure expressed in value terms.
func signedPositionValue(acct *market.Account, symbol string, price float64) float64 {
	pos, ok := acct.Positions[symbol]
	if !ok {
		return 0
	}
	value := pos.Quantity * price
	if pos.Side == market.Short {
		return -value
	}
	return value
}

func absValue(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
