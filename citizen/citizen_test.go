package citizen

import (
	"testing"
)

func TestCitizenEconomyLifecycle(t *testing.T) {
	ce := NewDefaultCitizenEconomy()

	// 1. Verify baseline setup
	if ce.Demographics.TotalPopulation < 300_000_000 {
		t.Fatalf("expected population >= 300M, got %f", ce.Demographics.TotalPopulation)
	}
	if ce.Sentiment.Index != 85.0 {
		t.Fatalf("expected baseline sentiment 85.0, got %f", ce.Sentiment.Index)
	}

	// 2. Tick simulation
	// 1 tick = 1/250th of a business year (~1 trading day)
	tickFraction := 1.0 / 250.0
	report := ce.Tick(
		161_500_000, // 161.5M employed (~4.0% unemployment)
		35.00,       // $35/hr
		0.024,       // 2.4% inflation
		0.040,       // 4.0% wage growth
		0.05,        // +5% stock market gain
		tickFraction,
	)

	// 3. Verify report contents
	if report.Population <= 0 {
		t.Fatalf("invalid population in report")
	}
	if report.UnemploymentRate <= 0 || report.UnemploymentRate > 0.15 {
		t.Fatalf("unemployment rate out of realistic bounds: %f", report.UnemploymentRate)
	}
	if report.ConsumerConfidence < 20 || report.ConsumerConfidence > 120 {
		t.Fatalf("consumer confidence out of bounds: %f", report.ConsumerConfidence)
	}
	if report.Happiness < 20 || report.Happiness > 100 {
		t.Fatalf("happiness out of bounds: %f", report.Happiness)
	}
	if report.TaxesPaidThisTick <= 0 {
		t.Fatalf("expected positive taxes paid this tick, got %f", report.TaxesPaidThisTick)
	}

	// Check elastic vs inelastic sector demand
	discretionaryDemand := report.SectorDemand["Consumer Discretionary"]
	staplesDemand := report.SectorDemand["Consumer Staples"]
	if discretionaryDemand <= 0 || staplesDemand <= 0 {
		t.Fatalf("expected positive sector demand factors")
	}

	t.Logf("Citizen economy tick verified: Pop=%.1fM, Unemp=%.2f%%, Conf=%.1f, Happiness=%.1f, Taxes=$%.2fM",
		report.Population/1e6, report.UnemploymentRate*100, report.ConsumerConfidence, report.Happiness, report.TaxesPaidThisTick/1e6)
}

func TestCitizenEconomy_SetPublicFactors(t *testing.T) {
	ce := NewDefaultCitizenEconomy()
	if ce == nil || ce.Demographics == nil {
		t.Fatal("expected non-nil CitizenEconomy and Demographics")
	}

	initGrowth := ce.Demographics.PopulationGrowthRate
	ce.SetPublicFactors(80.0, 90.0)

	if ce.Demographics.PopulationGrowthRate <= initGrowth {
		t.Fatalf("expected population growth rate to increase from %f, got %f", initGrowth, ce.Demographics.PopulationGrowthRate)
	}

	expectedMinWage := 15.0 + (90.0/100.0)*35.0
	if ce.Demographics.AverageHourlyEarnings < expectedMinWage {
		t.Fatalf("expected average wage >= %f, got %f", expectedMinWage, ce.Demographics.AverageHourlyEarnings)
	}
}
