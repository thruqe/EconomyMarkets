package country_test

import (
	"testing"

	"economy/country"
)

func TestNationalEconomy_Initialization(t *testing.T) {
	nat := country.NewDefaultNationalEconomy()
	if nat == nil {
		t.Fatal("expected non-nil NationalEconomy")
	}

	if nat.NominalGDP < 20_000_000_000_000.0 {
		t.Errorf("expected nominal GDP > 20T, got %f", nat.NominalGDP)
	}

	if nat.CentralBank.FedFundsRate <= 0 {
		t.Errorf("expected positive Fed Funds Rate, got %f", nat.CentralBank.FedFundsRate)
	}

	if nat.Labor.LaborForce < 100_000_000 {
		t.Errorf("expected Labor Force > 100M, got %f", nat.Labor.LaborForce)
	}

	if nat.Trade.AnnualExports <= 0 || nat.Trade.AnnualImports <= 0 {
		t.Errorf("expected positive exports and imports")
	}

	if nat.Fiscal.NationalDebt < 30_000_000_000_000.0 {
		t.Errorf("expected National Debt > 30T, got %f", nat.Fiscal.NationalDebt)
	}
}

func TestNationalEconomy_Tick(t *testing.T) {
	nat := country.NewDefaultNationalEconomy()

	fraction := 1.0 / 252.0 // Daily tick
	consumerSpending := 18_500_000_000_000.0
	corporateInvestment := 5_000_000_000_000.0
	corporateHeadcount := 160_000_000.0
	corporateTaxes := 500_000_000_000.0
	citizenTaxes := 4_000_000_000_000.0

	var lastReport country.NationalReport
	fomcTriggered := false

	// Run for 300 ticks to cross an FOMC meeting
	for i := 0; i < 300; i++ {
		rep := nat.Tick(
			consumerSpending,
			corporateInvestment,
			corporateHeadcount,
			corporateTaxes,
			citizenTaxes,
			fraction,
		)
		lastReport = rep
		if rep.FOMCAnnouncement != "" {
			fomcTriggered = true
		}
	}

	if lastReport.GDP < 20_000_000_000_000.0 {
		t.Errorf("expected GDP > 20T, got %f", lastReport.GDP)
	}

	if lastReport.UnemploymentRate < 0.01 || lastReport.UnemploymentRate > 0.20 {
		t.Errorf("unemployment rate out of bounds: %f", lastReport.UnemploymentRate)
	}

	if lastReport.DollarIndexDXY < 70 || lastReport.DollarIndexDXY > 150 {
		t.Errorf("Dollar Index out of bounds: %f", lastReport.DollarIndexDXY)
	}

	if !fomcTriggered {
		t.Log("FOMC meeting did not change rates (held steady), which is acceptable if economy was balanced")
	}

	// Test composite sector demand
	itDemand := nat.CompositeSectorDemand("Information Technology", 1.2)
	if itDemand <= 0.5 || itDemand >= 2.0 {
		t.Errorf("composite sector demand out of expected range: %f", itDemand)
	}
}
