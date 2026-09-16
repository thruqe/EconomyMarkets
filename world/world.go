// Package world models the global multi-country economic simulation.
package world

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// SimDate represents a simulated calendar date (Y/M/D).
type SimDate struct {
	Year  int
	Month time.Month
	Day   int
}

// String formats the date as "Jan 02, 2026".
func (d SimDate) String() string {
	return fmt.Sprintf("%s %02d, %d", d.Month.String()[:3], d.Day, d.Year)
}

// SimEpoch is the calendar start of the simulation (Jan 1, 2020).
var SimEpoch = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

// DateFromMillis converts simulation-time milliseconds to a SimDate.
func DateFromMillis(ms int64) SimDate {
	t := SimEpoch.Add(time.Duration(ms) * time.Millisecond)
	return SimDate{Year: t.Year(), Month: t.Month(), Day: t.Day()}
}

// SpeedLabel returns a human-readable time-per-real-second label for a speed multiplier.
// At 1.0x speed, 4 ticks/s × 6 simulated hours/tick = 24 simulated hours/real-second (1 day/s).
func SpeedLabel(speed float64) string {
	if speed <= 0 {
		return "PAUSED"
	}
	daysPerSec := speed // 1.0x = 1 day/s
	switch {
	case daysPerSec >= 365:
		return fmt.Sprintf("~%.1f yr/s", daysPerSec/365.0)
	case daysPerSec >= 30:
		return fmt.Sprintf("~%.1f mo/s", daysPerSec/30.0)
	case daysPerSec >= 7:
		return fmt.Sprintf("~%.1f wk/s", daysPerSec/7.0)
	case daysPerSec >= 1:
		return fmt.Sprintf("~%.1f day/s", daysPerSec)
	default:
		hoursPerSec := daysPerSec * 24.0
		return fmt.Sprintf("~%.0f hr/s", hoursPerSec)
	}
}


// DiplomaticStance describes the political relationship between two countries.
type DiplomaticStance int

const (
	StanceNeutral DiplomaticStance = iota
	StanceFriendly
	StanceAllied
	StanceCold
	StanceHostile
)

func (s DiplomaticStance) String() string {
	switch s {
	case StanceAllied:
		return "Allied"
	case StanceFriendly:
		return "Friendly"
	case StanceCold:
		return "Cold"
	case StanceHostile:
		return "Hostile"
	default:
		return "Neutral"
	}
}

// Relation models bilateral foreign policy between home and a foreign country.
type Relation struct {
	TariffRate     float64          `json:"tariff_rate"`
	Stance         DiplomaticStance `json:"stance"`
	TradeVolume    float64          `json:"trade_volume"`
	AidFlow        float64          `json:"aid_flow"`
	SanctionsLevel int              `json:"sanctions_level"`
}

// Region represents a sub-national economic zone (state/province).
type Region struct {
	Name             string             `json:"name"`
	GDP              float64            `json:"gdp"`
	TaxRevenue       float64            `json:"tax_revenue"`
	UnemploymentRate float64            `json:"unemployment_rate"`
	Population       float64            `json:"population"`
	SectorStrengths  map[string]float64 `json:"sector_strengths"`
	Budget           float64            `json:"budget"`
	Debt             float64            `json:"debt"`
}

// ForeignCountry is a foreign nation with its own economic parameters.
type ForeignCountry struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	FlagCode         string           `json:"flag_code"`
	GDP              float64          `json:"gdp"`
	GDPGrowth        float64          `json:"gdp_growth"`
	Inflation        float64          `json:"inflation"`
	InterestRate     float64          `json:"interest_rate"`
	Unemployment     float64          `json:"unemployment"`
	Population       float64          `json:"population"`
	TradeBalance     float64          `json:"trade_balance"`
	CurrencyStrength float64          `json:"currency_strength"`
	NationalDebt     float64          `json:"national_debt"`
	DebtToGDP        float64          `json:"debt_to_gdp"`
	Relation         Relation         `json:"relation"`
	rng              *rand.Rand
}

// Tick advances the foreign country's economy by one sim tick.
func (fc *ForeignCountry) Tick(tickFracOfYear float64) {
	if fc.rng == nil {
		fc.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	growthShock := (fc.rng.Float64() - 0.495) * 0.006
	fc.GDPGrowth += growthShock
	if fc.GDPGrowth > 0.12 {
		fc.GDPGrowth = 0.12
	}
	if fc.GDPGrowth < -0.08 {
		fc.GDPGrowth = -0.08
	}
	fc.GDP *= (1.0 + fc.GDPGrowth*tickFracOfYear)

	inflShock := (fc.rng.Float64() - 0.5) * 0.001
	fc.Inflation += inflShock
	if fc.Inflation < -0.01 {
		fc.Inflation = -0.01
	}

	fc.NationalDebt *= (1.0 + 0.02*tickFracOfYear)
	if fc.GDP > 0 {
		fc.DebtToGDP = fc.NationalDebt / fc.GDP
	}

	// Hostile stance reduces bilateral trade
	if fc.Relation.Stance >= StanceCold {
		fc.Relation.TradeVolume *= 0.9999
	} else if fc.Relation.Stance == StanceAllied {
		fc.Relation.TradeVolume *= 1.00005
	}
}

// HomeCountry is the user's own country.
type HomeCountry struct {
	Name       string   `json:"name"`
	Currency   string   `json:"currency"`
	FlagCode   string   `json:"flag_code"`
	Founded    int      `json:"founded"`
	Regions    []Region `json:"regions"`
	Configured bool     `json:"configured"`
}

// IsConfigured returns true if the user has set up their home country.
func (h *HomeCountry) IsConfigured() bool {
	return h.Configured && h.Name != ""
}

// Configure sets up the home country with user-provided details.
func (h *HomeCountry) Configure(name, currency, flagCode string, founded int) {
	h.Name = strings.TrimSpace(name)
	h.Currency = strings.ToUpper(strings.TrimSpace(currency))
	h.FlagCode = strings.ToUpper(strings.TrimSpace(flagCode))
	h.Founded = founded
	h.Configured = true
}

// ForeignPolicyAction records a diplomatic action taken by the player.
type ForeignPolicyAction struct {
	Tick      int64
	CountryID string
	Action    string
	Detail    string
}

// World holds the entire multi-country simulation world.
type World struct {
	Home            HomeCountry             `json:"home"`
	Foreign         map[string]*ForeignCountry `json:"foreign"`
	PolicyLog       []ForeignPolicyAction   `json:"policy_log"`
	rng             *rand.Rand
}

// NewWorld creates a fresh world with 8 pre-seeded foreign countries.
func NewWorld(seed int64) *World {
	rng := rand.New(rand.NewSource(seed))
	w := &World{
		Foreign: make(map[string]*ForeignCountry),
		rng:     rng,
		Home: HomeCountry{
			Regions: defaultRegions(),
		},
	}

	foreignNations := []*ForeignCountry{
		{
			ID: "EU", Name: "European Union", FlagCode: "EU",
			GDP: 18_500_000_000_000, GDPGrowth: 0.012, Inflation: 0.028,
			InterestRate: 0.040, Unemployment: 0.062, Population: 450_000_000,
			TradeBalance: 300_000_000_000, CurrencyStrength: 1.08, NationalDebt: 14_000_000_000_000,
			Relation: Relation{TariffRate: 0.030, Stance: StanceFriendly, TradeVolume: 850_000_000_000},
		},
		{
			ID: "CHN", Name: "China", FlagCode: "CN",
			GDP: 17_700_000_000_000, GDPGrowth: 0.047, Inflation: 0.021,
			InterestRate: 0.031, Unemployment: 0.052, Population: 1_410_000_000,
			TradeBalance: 600_000_000_000, CurrencyStrength: 0.138, NationalDebt: 15_000_000_000_000,
			Relation: Relation{TariffRate: 0.085, Stance: StanceCold, TradeVolume: 600_000_000_000},
		},
		{
			ID: "GBR", Name: "United Kingdom", FlagCode: "GB",
			GDP: 3_100_000_000_000, GDPGrowth: 0.008, Inflation: 0.031,
			InterestRate: 0.052, Unemployment: 0.044, Population: 67_000_000,
			TradeBalance: -70_000_000_000, CurrencyStrength: 1.27, NationalDebt: 3_300_000_000_000,
			Relation: Relation{TariffRate: 0.020, Stance: StanceAllied, TradeVolume: 265_000_000_000},
		},
		{
			ID: "JPN", Name: "Japan", FlagCode: "JP",
			GDP: 4_200_000_000_000, GDPGrowth: 0.016, Inflation: 0.024,
			InterestRate: 0.0025, Unemployment: 0.026, Population: 126_000_000,
			TradeBalance: 50_000_000_000, CurrencyStrength: 0.0067, NationalDebt: 10_000_000_000_000,
			Relation: Relation{TariffRate: 0.025, Stance: StanceAllied, TradeVolume: 180_000_000_000},
		},
		{
			ID: "CAN", Name: "Canada", FlagCode: "CA",
			GDP: 2_100_000_000_000, GDPGrowth: 0.018, Inflation: 0.026,
			InterestRate: 0.045, Unemployment: 0.058, Population: 38_000_000,
			TradeBalance: -20_000_000_000, CurrencyStrength: 0.75, NationalDebt: 1_300_000_000_000,
			Relation: Relation{TariffRate: 0.000, Stance: StanceAllied, TradeVolume: 700_000_000_000},
		},
		{
			ID: "MEX", Name: "Mexico", FlagCode: "MX",
			GDP: 1_400_000_000_000, GDPGrowth: 0.022, Inflation: 0.047,
			InterestRate: 0.110, Unemployment: 0.028, Population: 130_000_000,
			TradeBalance: 30_000_000_000, CurrencyStrength: 0.058, NationalDebt: 900_000_000_000,
			Relation: Relation{TariffRate: 0.000, Stance: StanceFriendly, TradeVolume: 650_000_000_000},
		},
		{
			ID: "BRA", Name: "Brazil", FlagCode: "BR",
			GDP: 2_000_000_000_000, GDPGrowth: 0.028, Inflation: 0.052,
			InterestRate: 0.1075, Unemployment: 0.072, Population: 215_000_000,
			TradeBalance: 60_000_000_000, CurrencyStrength: 0.20, NationalDebt: 1_700_000_000_000,
			Relation: Relation{TariffRate: 0.040, Stance: StanceNeutral, TradeVolume: 85_000_000_000},
		},
		{
			ID: "IND", Name: "India", FlagCode: "IN",
			GDP: 3_700_000_000_000, GDPGrowth: 0.067, Inflation: 0.051,
			InterestRate: 0.065, Unemployment: 0.079, Population: 1_420_000_000,
			TradeBalance: -200_000_000_000, CurrencyStrength: 0.012, NationalDebt: 3_100_000_000_000,
			Relation: Relation{TariffRate: 0.055, Stance: StanceFriendly, TradeVolume: 130_000_000_000},
		},
	}
	for _, fc := range foreignNations {
		fc.rng = rand.New(rand.NewSource(rng.Int63()))
		if fc.GDP > 0 {
			fc.DebtToGDP = fc.NationalDebt / fc.GDP
		}
		w.Foreign[fc.ID] = fc
	}
	return w
}

// defaultRegions returns 5 home-country regions with baseline economic data.
func defaultRegions() []Region {
	return []Region{
		{
			Name: "Northeast", GDP: 4_800_000_000_000, TaxRevenue: 580_000_000_000,
			UnemploymentRate: 0.038, Population: 57_000_000, Budget: 420_000_000_000,
			SectorStrengths: map[string]float64{"Financials": 1.4, "Information Technology": 1.3, "Health Care": 1.2},
		},
		{
			Name: "South", GDP: 5_200_000_000_000, TaxRevenue: 490_000_000_000,
			UnemploymentRate: 0.042, Population: 84_000_000, Budget: 380_000_000_000,
			SectorStrengths: map[string]float64{"Energy": 1.5, "Industrials": 1.3, "Real Estate": 1.2},
		},
		{
			Name: "Midwest", GDP: 3_600_000_000_000, TaxRevenue: 360_000_000_000,
			UnemploymentRate: 0.036, Population: 67_000_000, Budget: 295_000_000_000,
			SectorStrengths: map[string]float64{"Industrials": 1.4, "Consumer Staples": 1.3, "Materials": 1.2},
		},
		{
			Name: "West", GDP: 6_100_000_000_000, TaxRevenue: 700_000_000_000,
			UnemploymentRate: 0.040, Population: 71_000_000, Budget: 510_000_000_000,
			SectorStrengths: map[string]float64{"Information Technology": 1.5, "Communication Services": 1.4, "Energy": 1.1},
		},
		{
			Name: "Pacific", GDP: 1_800_000_000_000, TaxRevenue: 190_000_000_000,
			UnemploymentRate: 0.035, Population: 16_000_000, Budget: 145_000_000_000,
			SectorStrengths: map[string]float64{"Real Estate": 1.3, "Consumer Discretionary": 1.2, "Utilities": 1.2},
		},
	}
}

// Tick advances the world economy by one simulation tick.
func (w *World) Tick(tickFracOfYear float64) {
	for _, fc := range w.Foreign {
		fc.Tick(tickFracOfYear)
	}
}

// SetRelation updates diplomatic stance and tariff with a foreign country.
func (w *World) SetRelation(countryID string, stance DiplomaticStance, tariffRate float64) bool {
	if fc, ok := w.Foreign[countryID]; ok {
		fc.Relation.Stance = stance
		fc.Relation.TariffRate = tariffRate
		return true
	}
	return false
}

// SendAid sends an aid package to a foreign country, improving diplomatic stance.
func (w *World) SendAid(countryID string, amount float64, tick int64) bool {
	fc, ok := w.Foreign[countryID]
	if !ok {
		return false
	}
	fc.Relation.AidFlow += amount
	if amount >= 1_000_000_000 && fc.Relation.Stance < StanceFriendly {
		fc.Relation.Stance++
	} else if amount >= 5_000_000_000 && fc.Relation.Stance < StanceAllied {
		fc.Relation.Stance++
	}
	w.PolicyLog = append(w.PolicyLog, ForeignPolicyAction{
		Tick: tick, CountryID: countryID, Action: "Aid",
		Detail: fmt.Sprintf("Sent %s in aid to %s", formatMoney(amount), fc.Name),
	})
	return true
}

// ImposeSanctions applies sanctions on a foreign country.
func (w *World) ImposeSanctions(countryID string, level int, tick int64) bool {
	fc, ok := w.Foreign[countryID]
	if !ok {
		return false
	}
	fc.Relation.SanctionsLevel = level
	if level >= 2 {
		fc.Relation.Stance = StanceHostile
	} else if level >= 1 && fc.Relation.Stance < StanceCold {
		fc.Relation.Stance = StanceCold
	}
	w.PolicyLog = append(w.PolicyLog, ForeignPolicyAction{
		Tick: tick, CountryID: countryID, Action: "Sanctions",
		Detail: fmt.Sprintf("Level %d sanctions imposed on %s", level, fc.Name),
	})
	return true
}

// NegotiateDeal improves tariff rate and stance with a country.
func (w *World) NegotiateDeal(countryID string, newTariff float64, tick int64) bool {
	fc, ok := w.Foreign[countryID]
	if !ok {
		return false
	}
	oldTariff := fc.Relation.TariffRate
	fc.Relation.TariffRate = newTariff
	if newTariff < oldTariff && fc.Relation.Stance < StanceAllied {
		fc.Relation.Stance++
	}
	w.PolicyLog = append(w.PolicyLog, ForeignPolicyAction{
		Tick: tick, CountryID: countryID, Action: "Trade Deal",
		Detail: fmt.Sprintf("Tariff with %s: %.1f%% → %.1f%%", fc.Name, oldTariff*100, newTariff*100),
	})
	return true
}

// ForeignCountriesSorted returns foreign countries sorted alphabetically by name.
func (w *World) ForeignCountriesSorted() []*ForeignCountry {
	result := make([]*ForeignCountry, 0, len(w.Foreign))
	for _, fc := range w.Foreign {
		result = append(result, fc)
	}
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Name > result[j].Name {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

func formatMoney(v float64) string {
	switch {
	case v >= 1e12:
		return fmt.Sprintf("$%.2fT", v/1e12)
	case v >= 1e9:
		return fmt.Sprintf("$%.2fB", v/1e9)
	case v >= 1e6:
		return fmt.Sprintf("$%.2fM", v/1e6)
	default:
		return fmt.Sprintf("$%.0f", v)
	}
}
