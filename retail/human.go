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
	id      string
	account *market.Account
	pending []*market.Order
}

// NewHumanTrader constructs a HumanTrader with the given starting
// account. A real deployment would likely tie this account's starting
// cash to whatever the product's onboarding/virtual-currency grant is
// — that policy lives outside this package.
func NewHumanTrader(id string, acct *market.Account) *HumanTrader {
	return &HumanTrader{id: id, account: acct}
}

func (h *HumanTrader) ID() string               { return h.id }
func (h *HumanTrader) Account() *market.Account { return h.account }

// SubmitOrder queues an order to be included in this trader's next
// NextOrders call. Intended to be called from whatever handles the
// UI action (a button press, a form submission) between ticks; not
// safe for concurrent use without external synchronization, since a
// real deployment's UI layer is expected to serialize actions per
// user session.
func (h *HumanTrader) SubmitOrder(o *market.Order) {
	o.AgentID = h.id // ensure the queued order is always correctly attributed, regardless of what the caller set
	h.pending = append(h.pending, o)
}

// NextOrders implements market.OrderSource: drains whatever orders
// were queued since the last call. A human who submitted nothing this
// tick simply returns nil — sitting out is the ordinary case for a
// human trader most ticks, exactly as for any other participant.
func (h *HumanTrader) NextOrders(state market.MarketState) []*market.Order {
	if len(h.pending) == 0 {
		return nil
	}
	orders := h.pending
	h.pending = nil
	return orders
}
