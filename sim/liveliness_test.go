package sim

import (
	"math/rand"
	"testing"

	"economy/agent"
	"economy/company"
	"economy/market"
	"economy/retail"
)

func TestSimulationLivelinessOverLongRun(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	universe := company.GenerateUniverse(10, 42, company.DefaultGenerationParams())

	s := NewSimulation(10, 100)
	for _, co := range universe {
		s.AddCompany(co)
	}

	for _, co := range universe {
		mm := agent.NewMarketMaker("mm_"+co.Symbol, 50_000_000, 10, 0.05)
		mm.MaxInventory = 250_000
		mm.QuoteSize = 500
		s.AddParticipant(mm, co.Symbol)

		hfAcct := market.NewAccount("hf_"+co.Symbol, 20_000_000, 5, 0.10)
		hf := agent.NewHedgeFund("hf_"+co.Symbol, hfAcct, co)
		s.AddParticipant(hf, co.Symbol)

		bankAcct := market.NewAccount("bank_"+co.Symbol, 100_000_000, 3, 0.15)
		bank := agent.NewBank("bank_"+co.Symbol, bankAcct, []*company.Company{co})
		s.AddParticipant(bank, co.Symbol)

		// Seed initial book depth
		bk := s.Book(co.Symbol)
		for i := range 30 {
			bid := co.TrueValue * (1 - 0.002*float64(i+1))
			ask := co.TrueValue * (1 + 0.002*float64(i+1))
			bk.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: bid, Quantity: 500})
			bk.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: ask, Quantity: 500})
		}
	}

	retailBots := retail.GenerateWatchlistPool(100, rng.Int63(), universe, 1, 3, 1000, 20000)
	for _, bot := range retailBots {
		for _, co := range universe {
			if bot.Watches(co.Symbol) {
				s.AddParticipant(bot, co.Symbol)
			}
		}
	}

	// Warm up 80 ticks
	for i := range 80 {
		report := s.Step()
		if i == 0 || i == 79 {
			t.Logf("Warmup tick %d: trades=%d", i, len(report.Trades))
			for sym, p := range report.Prices {
				t.Logf("  sym=%s mid=%.2f bid=%.2f ask=%.2f hasMid=%v", sym, p.Mid, p.Bid, p.Ask, p.HasMid)
			}
		}
	}

	initialPrices := make(map[string]float64)
	for _, co := range universe {
		if bk := s.Book(co.Symbol); bk != nil {
			if mid, ok := bk.MidPrice(); ok {
				initialPrices[co.Symbol] = mid
			}
		}
	}

	// Run for 1,000 ticks and record trade counts in slices
	firstHalfTrades := 0
	secondHalfTrades := 0

	for i := range 1000 {
		report := s.Step()
		if i < 500 {
			firstHalfTrades += len(report.Trades)
		} else {
			secondHalfTrades += len(report.Trades)
		}
	}

	t.Logf("First 500 ticks trades: %d, Second 500 ticks trades: %d", firstHalfTrades, secondHalfTrades)

	if firstHalfTrades == 0 || secondHalfTrades == 0 {
		t.Fatalf("Market froze! Trades in first half: %d, second half: %d", firstHalfTrades, secondHalfTrades)
	}

	// Ensure second half trading volume is healthy (at least 30% of first half, not frozen to 0)
	if float64(secondHalfTrades) < float64(firstHalfTrades)*0.3 {
		t.Fatalf("Trading volume collapsed! First half: %d, second half: %d", firstHalfTrades, secondHalfTrades)
	}

	// Ensure prices actually moved for all companies
	movedCount := 0
	for _, co := range universe {
		if bk := s.Book(co.Symbol); bk != nil {
			if mid, ok := bk.MidPrice(); ok {
				initMid := initialPrices[co.Symbol]
				diff := mid - initMid
				t.Logf("Company %s: initial mid=%.2f, final mid=%.2f, delta=%.2f", co.Symbol, initMid, mid, diff)
				if diff != 0 {
					movedCount++
				}
			}
		}
	}

	if movedCount < len(universe) {
		t.Fatalf("Only %d/%d companies moved price!", movedCount, len(universe))
	}
}
