package retail

import "economy/market"

// HumanTrader is the bridge between a real person's UI actions (in
// whatever frontend this simulation eventually has) and the sim's
// tick loop. It deliberately makes no decisions of its own — all the
// real logic (what to buy, at what price, how much) is a UI/frontend
// concern outside this package. HumanTrader's only job is to hold
// orders queued via SubmitOrder until the next tick, then hand them
// to the sim exactly like any other market.OrderSource.
type HumanTrader struct {
	id              string
	account         *market.Account
	pending         []*market.Order
	pendingBySymbol map[string][]*market.Order
}

// NewHumanTrader constructs a HumanTrader with the given starting
// account. A real deployment would likely tie this account's starting
// cash to whatever the product's onboarding/virtual-currency grant is
// — that policy lives outside this package.
func NewHumanTrader(id string, acct *market.Account) *HumanTrader {
	return &HumanTrader{
		id:              id,
		account:         acct,
		pendingBySymbol: make(map[string][]*market.Order),
	}
}

func (h *HumanTrader) ID() string               { return h.id }
func (h *HumanTrader) Account() *market.Account { return h.account }

// SubmitOrder queues an order to be included in this trader's next
// NextOrders call. Intended to be called from whatever handles the
// UI action (a button press, a form submission) between ticks.
func (h *HumanTrader) SubmitOrder(o *market.Order) {
	o.AgentID = h.id
	h.pending = append(h.pending, o)
}

// SubmitOrderForSymbol queues an order targeted specifically to the given
// market symbol.
func (h *HumanTrader) SubmitOrderForSymbol(symbol string, o *market.Order) {
	o.AgentID = h.id
	if h.pendingBySymbol == nil {
		h.pendingBySymbol = make(map[string][]*market.Order)
	}
	h.pendingBySymbol[symbol] = append(h.pendingBySymbol[symbol], o)
}

// NextOrders implements market.OrderSource: drains whatever orders
// were queued for this market's symbol since the last call.
func (h *HumanTrader) NextOrders(state market.MarketState) []*market.Order {
	var result []*market.Order

	if h.pendingBySymbol != nil && len(h.pendingBySymbol[state.Symbol]) > 0 {
		result = append(result, h.pendingBySymbol[state.Symbol]...)
		delete(h.pendingBySymbol, state.Symbol)
	}

	if len(h.pending) > 0 {
		result = append(result, h.pending...)
		h.pending = nil
	}

	if len(result) == 0 {
		return nil
	}
	return result
}
