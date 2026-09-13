package agent

import (
	"economy/company"
	"economy/market"
)

// Bank models a bank's own trading desk: conviction-driven and
// fundamentals-aware like HedgeFund, but distinguished by two things
// that are genuinely true of real bank desks rather than just "a fund
// with different numbers":
//
//  1. It trades multiple companies from one shared capital pool,
//     rather than one company per fund instance — real desks allocate
//     one balance sheet across many positions simultaneously.
//
//  2. It is subject to account-wide, drawdown-driven de-risking that
//     is independent of any individual position's thesis: as the
//     account's equity falls from its high-water mark, every
//     covered company's target exposure gets scaled down together.
//     This mirrors how a real risk committee or regulatory capital
//     constraint forces a desk to cut exposure across the board
//     during a bad stretch, rather than asking "is this specific
//     trade still a good idea" position by position. The book-wide
//     leverage cap itself (via market.Account.MaxLeverage) plays the
//     role of the hard regulatory capital floor; DrawdownDeRisk*
//     below plays the role of the more dynamic internal risk-limit
//     layer that sits alongside it in a real bank.
type Bank struct {
	id       string
	account  *market.Account
	coverage []*company.Company

	MinTradeThreshold       float64
	FullConvictionThreshold float64
	RebalanceThreshold      float64

	// SkepticismDiscount (0-1) — see convictionSizedOrder's
	// documentation and HedgeFund's field of the same name. Bank
	// defaults to a higher discount than HedgeFund, reflecting a bank
	// desk's typically more conservative, more heavily-scrutinized
	// research process relative to a hedge fund's.
	SkepticismDiscount float64

	// equityHighWaterMark tracks the highest account equity observed
	// so far, updated every tick this Bank is asked for orders.
	// Drawdown is measured relative to this peak, not to starting
	// capital, since a desk that's already given back some of an
	// earlier gain is genuinely in a worse risk position than one
	// sitting flat at its starting capital.
	equityHighWaterMark float64
	hwmInitialized      bool

	// DrawdownDeRiskThreshold is the fraction of drawdown from the
	// high-water mark (0.10 = 10% down from peak) at which risk
	// scaling begins to bite.
	DrawdownDeRiskThreshold float64

	// DrawdownFullCutoff is the drawdown fraction at which risk
	// scaling reaches zero — the desk stops adding any new or
	// incremental risk entirely (existing positions can still be
	// reduced, since a scale of 0 still allows delta to call for
	// unwinding toward a target of zero).
	DrawdownFullCutoff float64

	// externalPrices, when set via SetExternalPrices, is preferred
	// over currentPricesAcrossCoverage's ReportedValue-based fallback
	// for any covered symbol it contains.
	externalPrices map[string]float64
}

// NewBank constructs a Bank covering the given companies from one
// shared account. Threshold defaults are set more conservative than
// HedgeFund's — higher MinTradeThreshold (won't chase small
// mispricings the way an aggressive fund will) and lower per-company
// deployment, reflecting a real bank desk's more conservative mandate
// relative to a hedge fund's.
func NewBank(id string, acct *market.Account, coverage []*company.Company) *Bank {
	return &Bank{
		id:                      id,
		account:                 acct,
		coverage:                coverage,
		MinTradeThreshold:       0.06, // 6% — more conservative than HedgeFund's 3%
		FullConvictionThreshold: 0.30, // slightly higher bar for "fully convinced"
		RebalanceThreshold:      0.05,
		SkepticismDiscount:      0.50, // more conservative than HedgeFund's 0.35
		DrawdownDeRiskThreshold: 0.10, // de-risking begins at 10% drawdown from peak
		DrawdownFullCutoff:      0.30, // no new risk at all by 30% drawdown from peak
	}
}

func (b *Bank) ID() string               { return b.id }
func (b *Bank) Account() *market.Account { return b.account }

// riskScale computes the current account-wide risk-scaling factor
// from drawdown relative to the tracked high-water mark: 1.0 at no
// drawdown, linearly shrinking to 0 between DrawdownDeRiskThreshold
// and DrawdownFullCutoff, and 0 beyond DrawdownFullCutoff.
// currentEquity must reflect mark-to-market equity across everything
// this account holds, not just one company's position, since the
// de-risking response is meant to be account-wide.
func (b *Bank) riskScale(currentEquity float64) float64 {
	if !b.hwmInitialized || currentEquity > b.equityHighWaterMark {
		b.equityHighWaterMark = currentEquity
		b.hwmInitialized = true
		return 1.0
	}
	if b.equityHighWaterMark <= 0 {
		return 1.0 // avoid divide-by-zero in a degenerate account state
	}

	drawdown := (b.equityHighWaterMark - currentEquity) / b.equityHighWaterMark

	if drawdown <= b.DrawdownDeRiskThreshold {
		return 1.0
	}
	if drawdown >= b.DrawdownFullCutoff {
		return 0.0
	}
	span := b.DrawdownFullCutoff - b.DrawdownDeRiskThreshold
	return 1.0 - (drawdown-b.DrawdownDeRiskThreshold)/span
}

// currentPricesAcrossCoverage builds a symbol->price map covering
// every company this Bank trades. For the symbol the current tick's
// MarketState concerns, it uses the real market price (state.Mid).
// For every other covered symbol, it falls back to that company's own
// ReportedValue as a stand-in — consistent with this Bank reading
// ReportedValue elsewhere (see convictionSizedOrder), rather than
// TrueValue, which no participant in this simulation should read
// directly. This fallback is still an approximation: a real desk's
// mark-to-market equity needs the actual *market* price for every
// held symbol, not a fundamentals-based stand-in, for any symbol
// other than the one in the current tick's state. SetExternalPrices
// lets an orchestrator holding every company's real order book supply
// exact prices instead; when set, those take priority over this
// fallback entirely.
func (b *Bank) currentPricesAcrossCoverage(state market.MarketState) map[string]float64 {
	prices := make(map[string]float64, len(b.coverage))
	for _, c := range b.coverage {
		if b.externalPrices != nil {
			if p, ok := b.externalPrices[c.Symbol]; ok {
				prices[c.Symbol] = p
				continue
			}
		}
		if c.Symbol == state.Symbol && state.HasMid {
			prices[c.Symbol] = state.Mid
		} else {
			prices[c.Symbol] = c.ReportedValue
		}
	}
	return prices
}

// SetExternalPrices supplies a real, orchestrator-maintained
// symbol->market-price map that currentPricesAcrossCoverage should
// prefer over its ReportedValue-based fallback. An orchestrator
// holding every company's actual order book should call this once per
// tick before polling this Bank for orders, so cross-company
// mark-to-market equity (and therefore drawdown-driven de-risking) is
// computed from real market prices rather than an approximation. Safe
// to leave unset in isolated tests, where the fallback remains
// available.
func (b *Bank) SetExternalPrices(prices map[string]float64) {
	b.externalPrices = prices
}

// NextOrders implements market.OrderSource for a single covered
// company per call, matching how the sim loop is expected to drive
// any OrderSource: once per company's MarketState each tick. Bank
// identifies which of its covered companies state.Symbol refers to,
// updates its account-wide risk scale from current equity, and
// produces at most one conviction-sized order for that company scaled
// by the current risk factor.
func (b *Bank) NextOrders(state market.MarketState) []*market.Order {
	var target *company.Company
	for _, c := range b.coverage {
		if c.Symbol == state.Symbol {
			target = c
			break
		}
	}
	if target == nil {
		return nil // this Bank doesn't cover the company this tick's state concerns
	}

	equity := b.account.Equity(b.currentPricesAcrossCoverage(state))
	scale := b.riskScale(equity)

	order := convictionSizedOrder(
		b.id, b.account, target, state,
		b.MinTradeThreshold, b.FullConvictionThreshold, b.RebalanceThreshold,
		scale,
		b.SkepticismDiscount,
	)
	if order == nil {
		return nil
	}
	return []*market.Order{order}
}
