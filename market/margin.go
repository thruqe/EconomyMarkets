package market

// PositionSide mirrors Side but reads more naturally when describing
// a held position rather than an order direction.
type PositionSide int

const (
	Long PositionSide = iota
	Short
)

// Position represents one agent's holding in one instrument.
type Position struct {
	Side      PositionSide
	Quantity  float64
	EntryCost float64 // total cash paid/received at entry (before leverage math)
}

// Account tracks one agent's capital and leverage state. This is
// deliberately separate from "Agent" (the decision-making layer) — an
// Account is just the ledger: cash, positions, margin usage. Any agent
// type (retail, hedge fund, market maker) holds one of these.
type Account struct {
	AgentID string
	Cash    float64

	// Positions keyed by instrument symbol.
	Positions map[string]*Position

	// MaxLeverage is the max (position value / equity) this account is
	// allowed to run. 1.0 = no leverage (fully cash-backed). A retail
	// account might be capped at 2x, a hedge fund at 5-10x, etc.
	MaxLeverage float64

	// MaintenanceMarginRatio is the equity/position-value floor before
	// a margin call is triggered. Typically lower than 1/MaxLeverage —
	// e.g. you might be allowed to *open* at 5x leverage but only get
	// margin-called once equity/position-value drops below 10%.
	MaintenanceMarginRatio float64
}

func NewAccount(agentID string, startingCash float64, maxLeverage, maintenanceMarginRatio float64) *Account {
	return &Account{
		AgentID:                agentID,
		Cash:                   startingCash,
		Positions:              make(map[string]*Position),
		MaxLeverage:            maxLeverage,
		MaintenanceMarginRatio: maintenanceMarginRatio,
	}
}

// PositionValue returns the current mark-to-market value of a position
// at the given current price (always positive — direction is tracked
// separately via Side, this is magnitude of exposure).
func (p *Position) PositionValue(currentPrice float64) float64 {
	return p.Quantity * currentPrice
}

// UnrealizedPnL computes floating profit/loss on a position at the
// given current price.
func (p *Position) UnrealizedPnL(currentPrice float64) float64 {
	marketValue := p.Quantity * currentPrice
	if p.Side == Long {
		return marketValue - p.EntryCost
	}
	// Short: profit when price falls below entry.
	return p.EntryCost - marketValue
}

// Equity computes total account equity (cash + unrealized P&L across
// all held positions) given a map of current prices by symbol.
func (a *Account) Equity(currentPrices map[string]float64) float64 {
	equity := a.Cash
	for symbol, pos := range a.Positions {
		if price, ok := currentPrices[symbol]; ok {
			equity += pos.UnrealizedPnL(price)
		}
	}
	return equity
}

// TotalPositionValue sums the mark-to-market magnitude of all exposure,
// used as the denominator for margin ratio checks.
func (a *Account) TotalPositionValue(currentPrices map[string]float64) float64 {
	var total float64
	for symbol, pos := range a.Positions {
		if price, ok := currentPrices[symbol]; ok {
			total += pos.PositionValue(price)
		}
	}
	return total
}

// MarginRatio returns equity / total position value. A ratio of 1.0
// means no leverage in use; lower ratios mean more leverage/risk.
// Returns (0, false) if the account holds no positions (ratio undefined).
func (a *Account) MarginRatio(currentPrices map[string]float64) (float64, bool) {
	posValue := a.TotalPositionValue(currentPrices)
	if posValue <= 0 {
		return 0, false
	}
	return a.Equity(currentPrices) / posValue, true
}

// IsUnderMaintenanceMargin reports whether this account has fallen
// below its maintenance margin threshold and should be liquidated.
func (a *Account) IsUnderMaintenanceMargin(currentPrices map[string]float64) bool {
	ratio, hasPositions := a.MarginRatio(currentPrices)
	if !hasPositions {
		return false
	}
	return ratio < a.MaintenanceMarginRatio
}

// AvailableBuyingPower returns how much *additional* position value this
// account could still take on given its current equity and max leverage.
func (a *Account) AvailableBuyingPower(currentPrices map[string]float64) float64 {
	equity := a.Equity(currentPrices)
	maxExposure := equity * a.MaxLeverage
	used := a.TotalPositionValue(currentPrices)
	remaining := maxExposure - used
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ApplySettledFill updates acct's position in symbol to reflect one
// settled fill: side is the direction *from this account's
// perspective* (Buy fills add/open a Long, Sell fills add/open a
// Short), quantity and price describe the fill itself. This is the
// one canonical settlement path — the sim loop's tick logic should
// call this for every fill an account is a party to, rather than each
// call site reimplementing position bookkeeping independently.
//
// Handles three cases: no existing position (open a new one),
// same-direction fill (add to the position, weighted into
// EntryCost), and opposing-direction fill (reduce the existing
// position, or fully close and flip to the other side if the fill
// size exceeds what was held).
func ApplySettledFill(acct *Account, symbol string, side Side, quantity, price float64) {
	positionSide := Long
	if side == Sell {
		positionSide = Short
	}

	pos, ok := acct.Positions[symbol]
	if !ok {
		acct.Positions[symbol] = &Position{Side: positionSide, Quantity: quantity, EntryCost: quantity * price}
		return
	}

	if pos.Side == positionSide {
		pos.Quantity += quantity
		pos.EntryCost += quantity * price
		return
	}

	// Opposing fill: reduces the existing position, or fully closes
	// and flips to the other side if this fill's quantity exceeds
	// what was held.
	if quantity < pos.Quantity {
		pos.Quantity -= quantity
		return
	}
	remainder := quantity - pos.Quantity
	pos.Side = positionSide
	pos.Quantity = remainder
	pos.EntryCost = remainder * price
}

// LiquidationEngine scans a set of accounts each tick and generates
// forced-close market orders for any account under maintenance margin.
// This is intentionally decoupled from the OrderBook: it only produces
// orders. The caller (the sim's main tick loop) is responsible for
// submitting those orders to the book, which is what lets a forced
// liquidation itself move price and potentially trigger further
// liquidations elsewhere — the cascade emerges from that loop, not
// from anything hardcoded here.
type LiquidationEngine struct{}

// ForcedOrder pairs a generated liquidation order with the symbol it
// applies to and the account it was generated from, so the caller can
// route it to the right book and update the right account afterward.
type ForcedOrder struct {
	Symbol  string
	Order   Order
	Account *Account
}

// ScanForLiquidations checks every account against current prices and
// returns market orders that fully close out any position(s) held by
// accounts under maintenance margin. currentPrices must cover every
// symbol any account might hold.
func (LiquidationEngine) ScanForLiquidations(accounts []*Account, currentPrices map[string]float64) []ForcedOrder {
	var forced []ForcedOrder

	for _, acct := range accounts {
		if !acct.IsUnderMaintenanceMargin(currentPrices) {
			continue
		}
		for symbol, pos := range acct.Positions {
			if pos.Quantity <= 0 {
				continue
			}
			// Closing a long = sell; closing a short = buy back.
			side := Sell
			if pos.Side == Short {
				side = Buy
			}
			forced = append(forced, ForcedOrder{
				Symbol: symbol,
				Order: Order{
					AgentID:  acct.AgentID,
					Side:     side,
					Quantity: pos.Quantity,
					IsMarket: true,
				},
				Account: acct,
			})
		}
	}

	return forced
}
