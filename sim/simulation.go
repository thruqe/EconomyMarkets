package sim

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"economy/agent"
	"economy/citizen"
	"economy/company"
	"economy/country"
	"economy/country/fiscal"
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
	mu sync.RWMutex

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
	rng      *rand.Rand

	// Macroeconomic Ecosystem
	Citizen        *citizen.CitizenEconomy
	National       *country.NationalEconomy
	CitizenReport  citizen.CitizenReport
	NationalReport country.NationalReport
}

// NewSimulation constructs an empty orchestrator ready to have
// companies and participants registered via AddCompany/AddParticipant
// before Step is called. depthLevels controls how many order book
// price levels MarketState.BidDepth/AskDepth sum over (see
// market.OrderBook.DepthAtLevels); maxHistory bounds how many past
// mid-prices each company's rolling history retains.
func NewSimulation(depthLevels, maxHistory int) *Simulation {
	cit := citizen.NewDefaultCitizenEconomy()
	nat := country.NewDefaultNationalEconomy()

	// Initial reports
	citRep := cit.Tick(nat.Labor.EmployedWorkers, nat.Labor.AverageHourlyWage, nat.CentralBank.CPIInflationRate, nat.Labor.AnnualWageGrowth, 0.0, 1.0/252.0)
	natRep := nat.Tick(citRep.ConsumerSpending, 5_000_000_000_000.0, 160_000_000.0, 500_000_000_000.0, 4_000_000_000_000.0, 1.0/252.0)

	return &Simulation{
		markets:        make(map[string]*companyMarket),
		depthLevels:    depthLevels,
		maxHistory:     maxHistory,
		rng:            rand.New(rand.NewSource(1337)),
		Citizen:        cit,
		National:       nat,
		CitizenReport:  citRep,
		NationalReport: natRep,
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

// ListIPO admits a newly generated company into the live simulation as an IPO.
// It registers the company, provisions an underwriter MarketMaker, seeds the opening book
// with institutional depth around the IPO offering price, and records an IPO event.
func (s *Simulation) ListIPO(co *company.Company, underwriterCapital float64) {
	if _, exists := s.markets[co.Symbol]; exists {
		return
	}
	s.AddCompany(co)

	if underwriterCapital <= 0 {
		underwriterCapital = 25_000_000
	}

	// 1. Provision Underwriting Market Maker with dynamic depth
	mm := agent.NewMarketMaker("underwriter_"+co.Symbol, underwriterCapital, 10, 0.08)
	mm.MaxInventory = 500_000
	mm.QuoteSize = 250
	mm.LadderLevels = 5
	mm.SetCompany(co)
	s.AddParticipant(mm, co.Symbol)

	// 2. Provision Institutional Hedge Fund allocating capital
	hfAcct := market.NewAccount("hf_ipo_"+co.Symbol, 15_000_000, 5, 0.10)
	hf := agent.NewHedgeFund("hf_ipo_"+co.Symbol, hfAcct, co)
	hf.MinTradeThreshold = 0.003
	hf.FullConvictionThreshold = 0.025
	hf.RebalanceThreshold = 0.003
	hf.ExecutionRate = 0.06
	hf.MaxOrderShares = 200
	s.AddParticipant(hf, co.Symbol)

	// 3. Provision Institutional Investment Bank covering the company
	bankAcct := market.NewAccount("bank_ipo_"+co.Symbol, 25_000_000, 3, 0.15)
	bank := agent.NewBank("bank_ipo_"+co.Symbol, bankAcct, []*company.Company{co})
	bank.MinTradeThreshold = 0.006
	bank.FullConvictionThreshold = 0.040
	bank.RebalanceThreshold = 0.006
	bank.ExecutionRate = 0.05
	bank.MaxOrderShares = 250
	s.AddParticipant(bank, co.Symbol)

	offeringPrice := co.TrueValue
	if offeringPrice <= 0 {
		offeringPrice = 100.0
	}

	// Seed opening quotes from the market maker so newly listed IPO has immediate valid mid price
	book := s.Book(co.Symbol)
	for i := range 5 {
		offset := 0.05 * float64(i+1)
		book.AddLimitOrder(&market.Order{AgentID: mm.ID(), Side: market.Buy, Price: math.Round((offeringPrice-offset)*100) / 100, Quantity: 250})
		book.AddLimitOrder(&market.Order{AgentID: mm.ID(), Side: market.Sell, Price: math.Round((offeringPrice+offset)*100) / 100, Quantity: 250})
	}

	// 4. Record IPO event
	s.EventLog = append(s.EventLog, Event{
		Tick:       s.tick,
		Kind:       EventIPO,
		Symbol:     co.Symbol,
		IPORevenue: co.AnnualRevenue,
		IPOPrice:   offeringPrice,
		IPOShares:  co.SharesOutstanding,
	})
}

// SetTick sets the current tick (used when restoring saved state).
func (s *Simulation) SetTick(t int) {
	s.tick = t
}

// BailoutCompany executes a government rescue facility for a distressed corporation:
// injects capital, slashes debt by 50%, cuts interest burden, and lifts valuation by +50%.
func (s *Simulation) BailoutCompany(symbol string, capitalInjection float64) error {
	cm, ok := s.markets[symbol]
	if !ok {
		return fmt.Errorf("company %s not registered in simulation", symbol)
	}
	co := cm.co
	if capitalInjection <= 0 {
		capitalInjection = 500_000_000.0
	}

	debtRelief := co.DebtOutstanding * 0.50
	co.DebtOutstanding -= debtRelief
	co.InterestExpense *= 0.50
	co.TrueValue *= 1.50
	co.ReportedValue *= 1.50

	if s.National != nil {
		s.National.Fiscal.NationalDebt += capitalInjection
	}

	headline := fmt.Sprintf("Federal Government injected $%.0fM emergency facility (cut debt by $%.0fM, +50%% intrinsic value)", capitalInjection/1e6, debtRelief/1e6)
	s.EventLog = append(s.EventLog, Event{
		Tick:            s.tick,
		Kind:            EventBailout,
		Symbol:          symbol,
		DistressDetails: headline,
	})

	return nil
}

// TogglePolicy flips an executive policy on or off, emits an announcement event,
// and applies direct stimulus effects if the citizen stimulus policy is enacted.
func (s *Simulation) TogglePolicy(policyID string) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.National == nil {
		return false, "National economy not initialized"
	}
	active, name := s.National.Fiscal.TogglePolicy(policyID)

	// If citizen stimulus was triggered, inject directly into Citizen demographics and sentiment!
	if policyID == fiscal.PolicyCitizenStimulus && active && s.Citizen != nil {
		s.Citizen.Sentiment.Index = math.Min(100, s.Citizen.Sentiment.Index+15.0)
		s.Citizen.Sentiment.Happiness = math.Min(100, s.Citizen.Sentiment.Happiness+12.0)
		s.Citizen.Demographics.AggregateDisposableIncome += 320_000_000_000.0
		s.National.Fiscal.NationalDebt += 320_000_000_000.0
	}

	actionStr := "ENACTED"
	if !active {
		actionStr = "REPEALED"
	}
	headline := fmt.Sprintf("Executive Policy %s: %s", actionStr, name)
	s.EventLog = append(s.EventLog, Event{
		Tick:          s.tick,
		Kind:          EventMacro,
		Symbol:        "US_GOV",
		MacroHeadline: headline,
	})

	if s.National != nil {
		s.NationalReport.ActivePolicies = s.National.Fiscal.GetActivePolicies()
	}

	return active, name
}

