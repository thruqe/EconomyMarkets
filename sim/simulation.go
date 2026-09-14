package sim

import (
	"economy/agent"
	"economy/company"
	"economy/market"
)

// AccountHolder is satisfied by any participant that exposes its
// underlying market.Account — every concrete participant type built
// so far (agent.MarketMaker, agent.HedgeFund, agent.Bank,
// retail.SimulatedRetailTrader, retail.HumanTrader) already has an
// Account() method, so this costs nothing to require and is what lets
// the orchestrator run a single, global liquidation scan across every
// account in the simulation regardless of participant type. Defined
// locally in sim rather than in market, since market intentionally
// requires nothing beyond OrderSource from participants — this is an
// orchestration-level need, not a market-mechanics one.
type AccountHolder interface {
	market.OrderSource
	Account() *market.Account
}

// Simulation is the orchestrator: it owns every company's market,
// every participant, and drives the tick loop that is the only place
// in this codebase where package market, company, technicals, agent,
// and retail are all used together.
type Simulation struct {
	markets map[string]*companyMarket // symbol -> that company's market
	order   []string                  // symbols in a stable, deterministic iteration order

	holders []AccountHolder // every account-bearing participant, for the global liquidation scan
	banks   []*agent.Bank   // tracked separately so SetExternalPrices can be called each tick — see step 6 in Step's doc

	depthLevels int
	maxHistory  int

	tick int

	// EventLog collects notable events (fundamental jumps,
	// restatements, liquidations) across the run for later
	// inspection — see events.go.
	EventLog []Event
}

// NewSimulation constructs an empty orchestrator ready to have
// companies and participants registered via AddCompany/AddParticipant
// before Step is called. depthLevels controls how many order book
// price levels MarketState.BidDepth/AskDepth sum over (see
// market.OrderBook.DepthAtLevels); maxHistory bounds how many past
// mid-prices each company's rolling history retains.
func NewSimulation(depthLevels, maxHistory int) *Simulation {
	return &Simulation{
		markets:     make(map[string]*companyMarket),
		depthLevels: depthLevels,
		maxHistory:  maxHistory,
	}
}

// AddCompany registers a company with the simulation, giving it a
// fresh, empty order book. Must be called before any participant that
// trades this company is added, and before the first Step.
func (s *Simulation) AddCompany(co *company.Company) {
	if _, exists := s.markets[co.Symbol]; exists {
		return // idempotent: re-adding an already-registered company is a no-op, not an error
	}
	s.markets[co.Symbol] = newCompanyMarket(co, s.maxHistory)
	s.order = append(s.order, co.Symbol)
}

// AddParticipant registers a participant to trade the given symbols.
// The participant must implement AccountHolder — every concrete
// participant type in this codebase already does. For symbols not yet
// registered via AddCompany, AddParticipant panics rather than
// silently dropping the association, since a participant that
// believes it's trading a company the orchestrator has never heard of
// is a wiring bug worth catching immediately, not a scenario to
// tolerate quietly.
//
// If p is an *agent.Bank, it is additionally tracked so Step can call
// SetExternalPrices on it each tick with real cross-company market
// prices — see Step's documentation, and agent.Bank's own
// documentation on why that matters for accurate drawdown tracking.
func (s *Simulation) AddParticipant(p AccountHolder, symbols ...string) {
	for _, symbol := range symbols {
		cm, ok := s.markets[symbol]
		if !ok {
			panic("sim: AddParticipant called with unregistered symbol " + symbol + " — call AddCompany first")
		}
		cm.participants = append(cm.participants, p)
	}
	s.holders = append(s.holders, p)

	if bank, ok := p.(*agent.Bank); ok {
		s.banks = append(s.banks, bank)
	}
}

// Tick reports the number of completed Step calls so far.
func (s *Simulation) Tick() int {
	return s.tick
}

// Company returns the registered company for symbol, or nil if it
// isn't registered — exposed mainly for tests and result inspection
// (e.g. reading a company's final TrueValue/ReportedValue after a run).
func (s *Simulation) Company(symbol string) *company.Company {
	cm, ok := s.markets[symbol]
	if !ok {
		return nil
	}
	return cm.co
}

// Book returns the registered order book for symbol, or nil if it
// isn't registered — exposed mainly for tests and result inspection.
func (s *Simulation) Book(symbol string) *market.OrderBook {
	cm, ok := s.markets[symbol]
	if !ok {
		return nil
	}
	return cm.book
}

// Symbols returns every registered company's symbol, in the stable
// order they were added.
func (s *Simulation) Symbols() []string {
	out := make([]string, len(s.order))
	copy(out, s.order)
	return out
}
