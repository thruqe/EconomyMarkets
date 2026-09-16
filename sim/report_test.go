package sim

import (
	"testing"

	"economy/agent"
	"economy/company"
	"economy/market"
)

func TestStepReturnsPopulatedTickReport(t *testing.T) {
	params := company.DefaultGenerationParams()
	universe := company.GenerateUniverse(2, 42, params)
	co1, co2 := universe[0], universe[1]

	s := NewSimulation(5, 100)
	s.AddCompany(co1)
	s.AddCompany(co2)

	// Add market maker
	mm := agent.NewMarketMaker("mm1", 1_000_000, 5, 0.10)
	s.AddParticipant(mm, co1.Symbol, co2.Symbol)

	// Seed book
	for _, sym := range []string{co1.Symbol, co2.Symbol} {
		book := s.Book(sym)
		book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 100.0, Quantity: 1000})
		book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 102.0, Quantity: 1000})
	}

	report := s.Step()
	if report == nil {
		t.Fatalf("expected non-nil TickReport from Step()")
	}
	if report.Tick != 1 {
		t.Errorf("expected Tick=1, got %d", report.Tick)
	}
	if len(report.Prices) != 2 {
		t.Errorf("expected 2 price quotes, got %d", len(report.Prices))
	}
	p1, ok := report.Prices[co1.Symbol]
	if !ok || !p1.HasMid {
		t.Errorf("expected valid mid quote for %s", co1.Symbol)
	}
	if len(report.Accounts) != 1 {
		t.Errorf("expected 1 account snapshot, got %d", len(report.Accounts))
	}
	if report.Accounts[0].AgentID != "mm1" {
		t.Errorf("expected account mm1, got %s", report.Accounts[0].AgentID)
	}
}
