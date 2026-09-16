package main

import (
	"testing"

	"economy/agent"
	"economy/company"
	"economy/market"
	"economy/retail"
	"economy/sim"
)

func TestPriceTrace(t *testing.T) {
	seed := int64(42)
	params := company.DefaultGenerationParams()
	universe := company.GenerateUniverse(10, seed, params)
	s := sim.NewSimulation(5, 200)

	for _, co := range universe {
		s.AddCompany(co)
	}

	for _, co := range universe {
		mm := agent.NewMarketMaker("mm_"+co.Symbol, 50_000_000, 10, 0.10)
		mm.MaxInventory = 500_000
		mm.QuoteSize = 1000
		s.AddParticipant(mm, co.Symbol)

		hfAcct := market.NewAccount("hf_"+co.Symbol, 2_000_000, 5, 0.10)
		hf := agent.NewHedgeFund("hf_"+co.Symbol, hfAcct, co)
		s.AddParticipant(hf, co.Symbol)

		bankAcct := market.NewAccount("bank_"+co.Symbol, 3_000_000, 3, 0.15)
		bank := agent.NewBank("bank_"+co.Symbol, bankAcct, []*company.Company{co})
		s.AddParticipant(bank, co.Symbol)
	}

	pool := retail.GenerateWatchlistPool(100, seed, universe, 1, 3, 1000, 20000)
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

	for _, symbol := range s.Symbols() {
		book := s.Book(symbol)
		mid := s.Company(symbol).TrueValue
		for i := range 100 {
			bid := mid * (1 - 0.002*float64(i+1))
			ask := mid * (1 + 0.002*float64(i+1))
			book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: bid, Quantity: 5000})
			book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: ask, Quantity: 5000})
		}
	}

	for tick := 1; tick <= 120; tick++ {
		report := s.Step()
		if tick >= 80 && tick <= 100 {
			rch := report.Prices["RCH"]
			t.Logf("Tick %3d: RCH Mid=%.2f Bid=%.2f Ask=%.2f Spread=%.2f", tick, rch.Mid, rch.Bid, rch.Ask, rch.Spread)
		}
	}
}
