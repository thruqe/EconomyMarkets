package agent

import (
	"economy/market"
)

// MarketMaker provides continuous two-sided liquidity, deriving its
// quotes from an inventory-based model (Avellaneda-Stoikov), adapted
// for a discrete-tick simulation rather than continuous time.
//
// Three things distinguish this from a naive "quote around the mid"
// implementation, and all three are necessary for realistic behavior,
// not independent nice-to-haves:
//
//  1. Fair value is the MM's own EMA of past mid-prices, not the
//     current mid itself. This gives it genuine informational lag —
//     during a fast one-directional move, its quotes trail the market,
//     so it keeps buying while price falls (or selling while it
//     rises), accumulating inventory it doesn't want. Real market
//     makers get run over this way; a MM with no independent fair
//     value estimate cannot exhibit that failure mode at all.
//
//  2. The reservation price (the true center of its quotes) is shifted
//     away from fair value by its current inventory, scaled by both
//     risk aversion AND volatility — holding a large position is more
//     dangerous when volatility is high, so the skew has to respond to
//     both, not inventory alone.
//
//  3. Spread width itself grows with volatility (and with risk
//     aversion), which is what naturally throttles the MM's fill rate
//     and presence as conditions worsen — continuous widening rather
//     than a binary on/off pull-out, matching how real liquidity
//     provision degrades under stress.
type MarketMaker struct {
	id      string
	account *market.Account

	// fairValue is the MM's own running estimate of fair value,
	// updated once per tick from the observed mid-price. Distinct
	// from MarketState.Mid on purpose — see type doc.
	fairValue *EMA

	// RiskAversion (γ) controls how strongly inventory and volatility
	// widen the spread and shift the reservation price. Higher =
	// more risk-averse: wider spreads, more aggressive inventory
	// unwinding pressure.
	RiskAversion float64

	// BaseHalfSpread is the minimum half-spread quoted even at zero
	// volatility and zero inventory — covers a MM's baseline cost of
	// operating (adverse selection risk that exists even in calm
	// markets).
	BaseHalfSpread float64

	// SpreadVolCoefficient (k) scales how much variance contributes
	// to spread widening: halfSpread = BaseHalfSpread + k*variance.
	SpreadVolCoefficient float64

	// MaxInventory is the position size (in units of the instrument,
	// signed) beyond which the MM stops quoting on the side that
	// would increase exposure further, but will still quote (often
	// more aggressively) on the side that reduces it.
	MaxInventory float64

	// QuoteSize is the size posted at best bid/ask each tick. Kept
	// fixed for now — a natural later refinement is scaling this down
	// as inventory approaches MaxInventory.
	QuoteSize float64
}

// NewMarketMaker constructs a MarketMaker with an initial cash
// account. maxLeverage/maintenanceMarginRatio follow the same meaning
// as any other market.Account — a MM typically runs high leverage
// relative to a directional trader, since its risk is managed via
// inventory limits and spread, not position sizing alone.
func NewMarketMaker(id string, startingCash, maxLeverage, maintenanceMarginRatio float64) *MarketMaker {
	return &MarketMaker{
		id:                   id,
		account:              market.NewAccount(id, startingCash, maxLeverage, maintenanceMarginRatio),
		fairValue:            NewEMA(0.1),
		RiskAversion:         0.1,
		BaseHalfSpread:       0.02,
		SpreadVolCoefficient: 50.0,
		MaxInventory:         1000,
		QuoteSize:            100,
	}
}

func (m *MarketMaker) ID() string               { return m.id }
func (m *MarketMaker) Account() *market.Account { return m.account }

// reservationPrice computes the inventory- and volatility-adjusted
// center of the MM's quotes: r = fairValue - inventory*γ*σ².
// A positive inventory (net long) pulls the reservation price down —
// the MM wants to encourage selling and discourage further buying, so
// its whole quote ladder shifts down. A negative inventory (net
// short) pushes it up, for the symmetric reason.
func (m *MarketMaker) reservationPrice(inventory float64, variance float64) float64 {
	return m.fairValue.Value() - inventory*m.RiskAversion*variance
}

// halfSpread computes quote half-width: widens with both configured
// risk aversion and current variance, so degraded conditions
// (high RecentVolatility) produce wider quotes and therefore lower
// effective participation, without a hard on/off switch.
func (m *MarketMaker) halfSpread(variance float64) float64 {
	return m.BaseHalfSpread + m.SpreadVolCoefficient*m.RiskAversion*variance
}

// NextOrders implements market.OrderSource. A market maker with a
// valid fair-value estimate quotes both sides in the same tick,
// subject to inventory limits: it will not add to a side that would
// push its position further past MaxInventory in the direction it is
// already extended, but continues to quote the opposite (risk-
// reducing) side even when the risk-increasing side is blocked. This
// mirrors a real desk's behavior under an inventory limit breach —
// it doesn't stop trading altogether, it stops trading in the
// direction that makes things worse.
func (m *MarketMaker) NextOrders(state market.MarketState) []*market.Order {
	if state.HasMid {
		m.fairValue.Update(state.Mid)
	}
	if !m.fairValue.Initialized() {
		return nil // no fair value estimate yet, nothing to quote around
	}

	variance := Variance(state.RecentVolatility)
	inventory := NetInventory(m.account, state.Symbol)

	r := m.reservationPrice(inventory, variance)
	halfSpread := m.halfSpread(variance)

	bidPrice := r - halfSpread
	askPrice := r + halfSpread

	canBuy := inventory+m.QuoteSize <= m.MaxInventory
	canSell := inventory-m.QuoteSize >= -m.MaxInventory

	var orders []*market.Order

	if canBuy {
		orders = append(orders, &market.Order{
			AgentID:  m.id,
			Side:     market.Buy,
			Price:    bidPrice,
			Quantity: m.QuoteSize,
			IsMarket: false,
		})
	}
	if canSell {
		orders = append(orders, &market.Order{
			AgentID:  m.id,
			Side:     market.Sell,
			Price:    askPrice,
			Quantity: m.QuoteSize,
			IsMarket: false,
		})
	}

	return orders // nil if both sides are blocked by inventory limits
}

// EffectiveSpread is exposed for logging/analysis: lets the sim loop
// or a research notebook record how wide this MM was actually quoting
// at a given tick's volatility, without duplicating the formula.
func (m *MarketMaker) EffectiveSpread(recentVolatility float64) float64 {
	return 2 * m.halfSpread(Variance(recentVolatility))
}

// EffectiveReservationPrice mirrors EffectiveSpread for the
// reservation price, using current inventory for whichever symbol is
// passed in.
func (m *MarketMaker) EffectiveReservationPrice(symbol string, recentVolatility float64) float64 {
	return m.reservationPrice(NetInventory(m.account, symbol), Variance(recentVolatility))
}
