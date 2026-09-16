package agent

import (
	"math"

	"economy/company"
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

	// company is an optional reference to the covered enterprise fundamentals.
	// When present, fairValue is anchored to genuine business performance (ReportedValue).
	company *company.Company

	// LadderLevels specifies how many price depth levels to quote on each side (defaults to 1).
	LadderLevels int

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
		LadderLevels:         1,
		RiskAversion:         0.1,
		BaseHalfSpread:       0.02,
		SpreadVolCoefficient: 50.0,
		MaxInventory:         1000,
		QuoteSize:            100,
	}
}

func (m *MarketMaker) ID() string               { return m.id }
func (m *MarketMaker) Account() *market.Account { return m.account }

// SetCompany associates a company with this market maker so quote ladders track fundamental enterprise performance.
func (m *MarketMaker) SetCompany(c *company.Company) {
	m.company = c
}

// reservationPrice computes the inventory- and volatility-adjusted
// center of the MM's quotes: r = fairValue - inventory*γ*σ².
// A positive inventory (net long) pulls the reservation price down —
// the MM wants to encourage selling and discourage further buying, so
// its whole quote ladder shifts down. A negative inventory (net
// short) pushes it up, for the symmetric reason.
func (m *MarketMaker) reservationPrice(inventory float64, variance float64) float64 {
	effectiveVar := math.Max(variance, 0.0002)
	return m.fairValue.Value() - inventory*m.RiskAversion*effectiveVar
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

	var center float64
	if m.company != nil {
		fund := m.company.ReportedValue
		if fund <= 0 {
			fund = m.company.TrueValue
		}
		if fund > 0 {
			if m.fairValue.Initialized() {
				center = 0.70*fund + 0.30*m.fairValue.Value()
			} else {
				center = fund
			}
		}
	}
	if center <= 0 {
		if !m.fairValue.Initialized() {
			return nil // no fair value estimate yet, nothing to quote around
		}
		center = m.fairValue.Value()
	}

	variance := Variance(state.RecentVolatility)
	inventory := NetInventory(m.account, state.Symbol)

	effectiveVar := math.Max(variance, 0.0002)
	skew := inventory * m.RiskAversion * effectiveVar
	maxSkew := center * 0.04
	if skew > maxSkew {
		skew = maxSkew
	} else if skew < -maxSkew {
		skew = -maxSkew
	}
	r := center - skew

	maxHalfSpread := math.Max(0.10, center*0.02)
	halfSpread := math.Min(m.halfSpread(variance), maxHalfSpread)

	// Dynamic inventory management:
	// Rather than completely pulling one side (which collapses market mid-price and halts trading),
	// dynamically scale quote sizes and skew spreads so the MM continuously maintains a 2-sided book.
	buySize := m.QuoteSize
	sellSize := m.QuoteSize

	if inventory > m.MaxInventory*0.75 {
		// Long inventory high: shrink buy quote to minimum, expand sell quote to shed shares
		excess := (inventory - m.MaxInventory*0.75) / (m.MaxInventory * 0.25)
		buySize = math.Max(1.0, math.Floor(m.QuoteSize*(1.0-0.9*math.Min(1.0, excess))))
		sellSize = math.Floor(m.QuoteSize * (1.0 + 0.5*math.Min(2.0, excess)))
	} else if inventory < -m.MaxInventory*0.75 {
		// Short inventory high: shrink sell quote, expand buy quote to cover shares
		excess := (-inventory - m.MaxInventory*0.75) / (m.MaxInventory * 0.25)
		sellSize = math.Max(1.0, math.Floor(m.QuoteSize*(1.0-0.9*math.Min(1.0, excess))))
		buySize = math.Floor(m.QuoteSize * (1.0 + 0.5*math.Min(2.0, excess)))
	}

	canBuy := inventory < m.MaxInventory
	canSell := inventory > -m.MaxInventory

	if m.LadderLevels <= 1 {
		bidPrice := math.Round((r-halfSpread)*100) / 100
		askPrice := math.Round((r+halfSpread)*100) / 100
		if bidPrice < 0.01 {
			bidPrice = 0.01
		}
		if askPrice <= bidPrice {
			askPrice = bidPrice + 0.01
		}

		var orders []*market.Order
		if canBuy {
			orders = append(orders, &market.Order{
				AgentID:  m.id,
				Side:     market.Buy,
				Price:    bidPrice,
				Quantity: buySize,
				IsMarket: false,
			})
		}
		if canSell {
			orders = append(orders, &market.Order{
				AgentID:  m.id,
				Side:     market.Sell,
				Price:    askPrice,
				Quantity: sellSize,
				IsMarket: false,
			})
		}
		return orders
	}

	// Multi-level depth ladder: provides deep, mobile liquidity that travels with the market
	ladderSteps := []struct {
		offsetMult float64
		sizeMult   float64
	}{
		{1.0, 1.0},
		{2.5, 1.5},
		{5.0, 2.5},
		{9.0, 4.0},
		{15.0, 6.0},
	}
	if m.LadderLevels < len(ladderSteps) {
		ladderSteps = ladderSteps[:m.LadderLevels]
	}

	var orders []*market.Order
	for _, step := range ladderSteps {
		stepOffset := halfSpread * step.offsetMult
		bp := math.Round((r-stepOffset)*100) / 100
		ap := math.Round((r+stepOffset)*100) / 100
		if bp < 0.01 {
			bp = 0.01
		}
		if ap <= bp {
			ap = bp + 0.01
		}

		if canBuy {
			orders = append(orders, &market.Order{
				AgentID:  m.id,
				Side:     market.Buy,
				Price:    bp,
				Quantity: math.Round(buySize * step.sizeMult),
				IsMarket: false,
			})
		}
		if canSell {
			orders = append(orders, &market.Order{
				AgentID:  m.id,
				Side:     market.Sell,
				Price:    ap,
				Quantity: math.Round(sellSize * step.sizeMult),
				IsMarket: false,
			})
		}
	}
	return orders
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
