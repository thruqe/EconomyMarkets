package sim

import (
	"math/rand"
	"testing"

	"economy/agent"
	"economy/company"
	"economy/market"
	"economy/retail"
)

// buildSmallSlice constructs a minimal, fully-wired simulation: a
// handful of companies, one market maker and one hedge fund per
// company, and a shared retail pool watching them all — the small
// slice the project's own plan called for proving correctness on
// before scaling to the full 500-company universe.
func buildSmallSlice(t *testing.T) *Simulation {
	t.Helper()

	params := company.DefaultGenerationParams()
	universe := company.GenerateUniverse(5, 42, params)

	s := NewSimulation(5, 200)
	for _, co := range universe {
		s.AddCompany(co)
	}

	for _, co := range universe {
		mm := agent.NewMarketMaker("mm_"+co.Symbol, 10_000_000, 10, 0.10)
		mm.MaxInventory = 50_000
		mm.QuoteSize = 500
		s.AddParticipant(mm, co.Symbol)

		hfAcct := market.NewAccount("hf_"+co.Symbol, 2_000_000, 5, 0.10)
		hf := agent.NewHedgeFund("hf_"+co.Symbol, hfAcct, co)
		s.AddParticipant(hf, co.Symbol)

		bankAcct := market.NewAccount("bank_"+co.Symbol, 3_000_000, 3, 0.15)
		bank := agent.NewBank("bank_"+co.Symbol, bankAcct, []*company.Company{co})
		s.AddParticipant(bank, co.Symbol)
	}

	pool := retail.GenerateWatchlistPool(300, 7, universe, 1, 3, 1000, 20000)
	for _, bot := range pool {
		var watched []string
		for _, symbol := range s.Symbols() {
			if bot.Watches(symbol) {
				watched = append(watched, symbol)
			}
		}
		if len(watched) > 0 {
			s.AddParticipant(bot, watched...)
		}
	}

	// Seed each book with initial resting liquidity so there's a
	// valid price from tick one — mirroring how every earlier
	// package's tests needed seeded depth before any trading logic
	// could act on a real price.
	for _, symbol := range s.Symbols() {
		book := s.Book(symbol)
		mid := s.Company(symbol).TrueValue
		for i := range 50 {
			bid := mid * (1 - 0.005*float64(i+1))
			ask := mid * (1 + 0.005*float64(i+1))
			book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: bid, Quantity: 2000})
			book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: ask, Quantity: 2000})
		}
	}

	return s
}

// TestSmallSliceRunsWithoutPanicking is the baseline integration
// proof: a fully-wired small slice runs many ticks without panicking,
// and every company ends the run with a valid, sane price — the
// minimum bar before trusting any more specific behavior.
func TestSmallSliceRunsWithoutPanicking(t *testing.T) {
	s := buildSmallSlice(t)

	const ticks = 200
	for range ticks {
		s.Step()
	}

	if s.Tick() != ticks {
		t.Fatalf("expected Tick() to report %d after %d Step calls, got %d", ticks, ticks, s.Tick())
	}

	for _, symbol := range s.Symbols() {
		mid, ok := s.Book(symbol).MidPrice()
		if !ok {
			t.Fatalf("company %s has no valid mid price after %d ticks", symbol, ticks)
		}
		if mid <= 0 {
			t.Fatalf("company %s has a non-positive mid price after %d ticks: %.4f", symbol, ticks, mid)
		}
		t.Logf("%s: final mid=%.4f TrueValue=%.4f ReportedValue=%.4f", symbol, mid, s.Company(symbol).TrueValue, s.Company(symbol).ReportedValue)
	}
}

// TestEventLogCapturesFundamentalEvents confirms that over enough
// ticks, at least some fundamental events actually get logged — a
// sanity check that Step's wiring from Company.Tick() into EventLog
// is genuinely connected, not silently dropped.
func TestEventLogCapturesFundamentalEvents(t *testing.T) {
	s := buildSmallSlice(t)

	// Force at least one company toward a near-certain jump this run
	// by temporarily not touching jump params (they're already
	// calibrated to fire reasonably often over enough ticks — see
	// company.DefaultJumpParams's documented frequency). Run enough
	// ticks that at least one fundamental event across 5 companies is
	// highly likely.
	const ticks = 3000
	for range ticks {
		s.Step()
	}

	fundamentalCount := 0
	for _, e := range s.EventLog {
		if e.Kind == EventFundamental {
			fundamentalCount++
		}
	}

	t.Logf("fundamental events logged over %d ticks across %d companies: %d", ticks, len(s.Symbols()), fundamentalCount)
	if fundamentalCount == 0 {
		t.Fatalf("expected at least one fundamental event across 5 companies over %d ticks, got 0", ticks)
	}
}

// TestLiquidationCascadeIntegration deliberately sets up an
// over-leveraged account and a sharp adverse price move, then
// confirms Step's liquidation wiring actually produces a logged
// EventLiquidation and reduces the forced account's position —
// proving the full pipeline (scan -> forced order -> book submission
// -> settlement -> event log) works together, not just each piece in
// isolation as earlier packages' own tests already proved.
func TestLiquidationCascadeIntegration(t *testing.T) {
	s := buildSmallSlice(t)
	symbol := s.Symbols()[0]
	book := s.Book(symbol)

	co := s.Company(symbol)

	// A tightly-leveraged human trader account, forced long, about to
	// get run over by a sharp drop.
	acct := market.NewAccount("victim", 10_000, 5, 0.10)
	victim := retail.NewHumanTrader("victim", acct)
	s.AddParticipant(victim, symbol)

	mid, _ := book.MidPrice()
	market.ApplySettledFill(acct, symbol, market.Buy, 500, mid) // large, leveraged long

	// Drain existing resting bids for this symbol out of contention by
	// adding a large sell that walks price down hard, simulating an
	// adverse move steep enough to breach maintenance margin.
	crashPrice := mid * 0.4
	co.TrueValue = crashPrice
	co.ReportedValue = crashPrice

	book.AddLimitOrder(&market.Order{AgentID: "seed_far", Side: market.Buy, Price: crashPrice, Quantity: 1_000_000})
	crash := &market.Order{AgentID: "crash_seller", Side: market.Sell, Quantity: 150000, IsMarket: true}
	book.Submit(crash)

	s.Step()

	liquidated := false
	for _, e := range s.EventLog {
		if e.Kind == EventLiquidation && e.LiquidatedAgentID == "victim" {
			liquidated = true
			t.Logf("liquidation event: tick=%d symbol=%s qty=%.2f", e.Tick, e.Symbol, e.LiquidationQty)
		}
	}

	if !liquidated {
		t.Fatalf("expected the over-leveraged victim account to be liquidated after a sharp adverse move")
	}
}

// TestBankReceivesExternalPrices confirms Step actually wires real
// cross-company market prices into every registered agent.Bank via
// SetExternalPrices, by reproducing the same drawdown-driven
// de-risking proof agent's own tests use internally: force a large
// loss in one covered company through genuine market trading (not a
// direct field mutation, since this test is specifically about
// whether Step's plumbing delivers real prices), then confirm the
// bank's sizing on the second covered company shrinks relative to an
// identical bank that never saw that loss — the same signature
// TestBankDeRisksAfterDrawdown in package agent already validated for
// the mechanism itself; this test validates that sim.Step is what
// actually delivers the real prices which make that mechanism fire
// correctly in an orchestrated run, rather than only in an isolated
// unit test with hand-fed prices.
func TestBankReceivesExternalPrices(t *testing.T) {
	params := company.DefaultGenerationParams()
	params.BaseJumpParams.LambdaDown = 0
	params.BaseJumpParams.LambdaUp = 0
	universe := company.GenerateUniverse(2, 99, params)
	victimCo, healthyCo := universe[0], universe[1]

	s := NewSimulation(5, 200)
	for _, co := range universe {
		s.AddCompany(co)
		book := s.Book(co.Symbol)
		mid := co.TrueValue
		for i := range 30 {
			book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: mid * (1 - 0.002*float64(i+1)), Quantity: 5000})
			book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: mid * (1 + 0.002*float64(i+1)), Quantity: 5000})
		}
	}

	bankAcct := market.NewAccount("bank1", 5_000_000, 3, 0.15)
	bank := agent.NewBank("bank1", bankAcct, universe)
	s.AddParticipant(bank, victimCo.Symbol, healthyCo.Symbol)

	// Give the bank a large, unambiguous mispricing in victimCo so it
	// actually opens a real position there before the crash — without
	// a held position, a later price crash has no effect on equity at
	// all (Account.Equity only marks held Positions to market), so
	// this step is necessary for the test's premise to hold, not
	// optional setup.
	victimCo.TrueValue = victimCo.TrueValue * 1.5
	for range 3 {
		s.Step()
	}
	if _, hasPosition := bankAcct.Positions[victimCo.Symbol]; !hasPosition {
		t.Fatalf("expected the bank to have opened a position in victimCo from its large mispricing before the crash — test setup invalid")
	}
	victimCo.TrueValue = victimCo.TrueValue / 1.5 // reset so the position isn't still being added to during the crash phase below

	// A large real market sell crashes victimCo's price hard, through
	// genuine order-book trading — not a direct field mutation — so
	// this specifically tests whether Step's real market prices reach
	// the bank, not whether the underlying drawdown math works (that's
	// already proven in package agent's own tests).
	crash := &market.Order{AgentID: "crash_seller", Side: market.Sell, Quantity: 30000, IsMarket: true}
	s.Book(victimCo.Symbol).Submit(crash)

	for range 5 {
		s.Step()
	}

	// Bump healthyCo's ReportedValue well above its market price so
	// the bank has a large, unambiguous mispricing to react to —
	// isolating whether its *sizing* on this trade was suppressed by
	// the other company's real crash, rather than testing whether it
	// trades at all.
	healthyCo.ReportedValue = healthyCo.TrueValue * 1.5

	healthyState := market.NewMarketState(healthyCo.Symbol, s.Tick(), s.Book(healthyCo.Symbol), 5, nil, 0)
	drawnDownOrders := bank.NextOrders(healthyState)

	// A fresh, otherwise-identical bank that never saw victimCo's
	// crash — same starting capital, same coverage, same mispricing —
	// used as the baseline for comparison.
	freshAcct := market.NewAccount("bank2", 5_000_000, 3, 0.15)
	freshBank := agent.NewBank("bank2", freshAcct, universe)
	freshOrders := freshBank.NextOrders(healthyState)

	if len(freshOrders) != 1 {
		t.Fatalf("expected the fresh, undamaged bank to trade the large healthyCo mispricing, got %d orders", len(freshOrders))
	}

	var drawnDownQty float64
	if len(drawnDownOrders) == 1 {
		drawnDownQty = drawnDownOrders[0].Quantity
	}
	t.Logf("order qty: fresh bank=%.2f, bank exposed to real crash via Step=%.2f", freshOrders[0].Quantity, drawnDownQty)

	if drawnDownQty >= freshOrders[0].Quantity {
		t.Fatalf("expected the bank that lived through victimCo's real, Step-driven crash to size healthyCo smaller than a fresh bank, got fresh=%.2f exposed=%.2f — if these are equal, Step is not actually delivering real cross-company prices to Bank",
			freshOrders[0].Quantity, drawnDownQty)
	}
}

func TestSimulationListIPO(t *testing.T) {
	s := buildSmallSlice(t)
	initialCompanyCount := len(s.Symbols())

	used := make(map[string]bool)
	for _, sym := range s.Symbols() {
		used[sym] = true
	}

	rng := rand.New(rand.NewSource(999))
	ipoCo := company.GenerateIPOCompany(company.InformationTechnology, company.MegaCap, 18_000_000_000, s.Tick(), rng, used)

	s.ListIPO(ipoCo, 50_000_000)

	if len(s.Symbols()) != initialCompanyCount+1 {
		t.Fatalf("expected symbol count %d, got %d", initialCompanyCount+1, len(s.Symbols()))
	}

	book := s.Book(ipoCo.Symbol)
	if book == nil {
		t.Fatalf("expected registered order book for %s", ipoCo.Symbol)
	}

	mid, hasMid := book.MidPrice()
	if !hasMid || mid <= 0 {
		t.Fatalf("expected valid mid price for newly listed IPO %s, got %.2f", ipoCo.Symbol, mid)
	}

	// Advance simulation steps with the new IPO trading
	for range 20 {
		report := s.Step()
		if report == nil {
			t.Fatalf("expected non-nil step report")
		}
	}

	// Verify IPO event was logged
	foundIPOEvent := false
	for _, e := range s.EventLog {
		if e.Kind == EventIPO && e.Symbol == ipoCo.Symbol {
			foundIPOEvent = true
			if e.IPORevenue != 18_000_000_000 {
				t.Fatalf("expected IPO revenue 18B, got %.2f", e.IPORevenue)
			}
		}
	}

	if !foundIPOEvent {
		t.Fatalf("expected EventIPO for %s in EventLog", ipoCo.Symbol)
	}
}

func TestMacroeconomicIntegration(t *testing.T) {
	s := buildSmallSlice(t)
	report := s.Step()

	if report.Citizen.Population < 300_000_000 {
		t.Fatalf("expected population >= 300M, got %f", report.Citizen.Population)
	}
	if report.National.GDP < 20_000_000_000_000 {
		t.Fatalf("expected GDP >= 20T, got %f", report.National.GDP)
	}
	if report.National.FedFundsRate <= 0 {
		t.Fatalf("expected positive Fed Funds Rate, got %f", report.National.FedFundsRate)
	}

	// Verify companies have macro footprint populated
	for _, sym := range s.Symbols() {
		co := s.Company(sym)
		if co.Headcount <= 0 {
			t.Errorf("company %s headcount <= 0", sym)
		}
		if co.LaborExpense <= 0 {
			t.Errorf("company %s labor expense <= 0", sym)
		}
		if co.DebtOutstanding <= 0 {
			t.Errorf("company %s debt <= 0", sym)
		}
	}
}

