package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"economy/company"
	"economy/world"
)

const DefaultWorldStatePath = "data/world_state.json"

type SavedCompany struct {
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	Sector            int     `json:"sector"`
	CapTier           int     `json:"cap_tier"`
	TrueValue         float64 `json:"true_value"`
	ReportedValue     float64 `json:"reported_value"`
	SharesOutstanding float64 `json:"shares_outstanding"`
	Float             float64 `json:"float"`
	AnnualRevenue     float64 `json:"annual_revenue"`
	NetMargin         float64 `json:"net_margin"`
	SectorMultiple    float64 `json:"sector_multiple"`
	IPOPrice          float64 `json:"ipo_price"`
	IsIPO             bool    `json:"is_ipo"`
	IPOTick           int     `json:"ipo_tick"`
	Headcount         float64 `json:"headcount"`
	AverageWage       float64 `json:"average_wage"`
	LaborExpense      float64 `json:"labor_expense"`
	DebtOutstanding   float64 `json:"debt_outstanding"`
	InterestExpense   float64 `json:"interest_expense"`
	CorporateTaxPaid  float64 `json:"corporate_tax_paid"`
	CapEx             float64 `json:"capex"`
	MacroDemandFactor float64 `json:"macro_demand_factor"`
	MidPrice          float64 `json:"mid_price"`
	Bid               float64 `json:"bid"`
	Ask               float64 `json:"ask"`
	IsPublic          bool    `json:"is_public"`
	Stage             string  `json:"stage"`
	PrivateValuation  float64 `json:"private_valuation"`
}

type SavedMacro struct {
	// Citizen
	Population            float64 `json:"population"`
	EmployedCount         float64 `json:"employed_count"`
	AverageHourlyEarnings float64 `json:"avg_hourly_earnings"`
	AggregateDisposable   float64 `json:"aggregate_disposable"`
	ConsumerConfidence    float64 `json:"consumer_confidence"`
	Happiness             float64 `json:"happiness"`
	ConsumerSpending      float64 `json:"consumer_spending"`

	// National
	GDP              float64         `json:"gdp"`
	RealGDPGrowth    float64         `json:"real_gdp_growth"`
	FedFundsRate     float64         `json:"fed_funds_rate"`
	TenYearYield     float64         `json:"ten_year_yield"`
	CPIInflationRate float64         `json:"cpi_inflation_rate"`
	PolicyStance     string          `json:"policy_stance"`
	DollarIndexDXY   float64         `json:"dollar_index_dxy"`
	TradeBalance     float64         `json:"trade_balance"`
	NationalDebt     float64         `json:"national_debt"`
	ActivePolicies   map[string]bool `json:"active_policies"`

	// Sovereign Treasury & Nation Building
	TreasuryCash          float64 `json:"treasury_cash"`
	CreditRating          string  `json:"credit_rating"`
	BorrowingYield        float64 `json:"borrowing_yield"`
	InfrastructureLevel   float64 `json:"infrastructure_level"`
	HealthcareLevel       float64 `json:"healthcare_level"`
	EducationLevel        float64 `json:"education_level"`
	EnterpriseGrantsLevel float64 `json:"enterprise_grants_level"`
	ExportCapacityLevel   float64 `json:"export_capacity_level"`
	FDIInflow             float64 `json:"fdi_inflow"`
	ExportRevenue         float64 `json:"export_revenue"`
	ExchangeChartered     bool    `json:"exchange_chartered"`
}

type SavedCandle struct {
	TimeUnixMilli int64   `json:"time"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        float64 `json:"volume"`
}

// SavedForeignCountry persists a foreign country's economic state.
type SavedForeignCountry struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	FlagCode         string             `json:"flag_code"`
	GDP              float64            `json:"gdp"`
	GDPGrowth        float64            `json:"gdp_growth"`
	Inflation        float64            `json:"inflation"`
	InterestRate     float64            `json:"interest_rate"`
	Unemployment     float64            `json:"unemployment"`
	Population       float64            `json:"population"`
	TradeBalance     float64            `json:"trade_balance"`
	CurrencyStrength float64            `json:"currency_strength"`
	NationalDebt     float64            `json:"national_debt"`
	DebtToGDP        float64            `json:"debt_to_gdp"`
	TariffRate       float64            `json:"tariff_rate"`
	Stance           int                `json:"stance"`
	TradeVol         float64            `json:"trade_vol"`
	AidFlow          float64            `json:"aid_flow"`
	SanctionsLevel   int                `json:"sanctions_level"`
}

// SavedHomeCountry persists the user's home country configuration.
type SavedHomeCountry struct {
	Name       string `json:"name"`
	Currency   string `json:"currency"`
	FlagCode   string `json:"flag_code"`
	Founded    int    `json:"founded"`
	Configured bool   `json:"configured"`
}

type SavedWorldState struct {
	Version     int                                            `json:"version"`
	Tick        int                                            `json:"tick"`
	SimTime     int64                                          `json:"sim_time"`
	TotalTrades int64                                          `json:"total_trades"`
	TotalVolume float64                                        `json:"total_volume"`
	Macro       SavedMacro                                     `json:"macro"`
	Companies   []SavedCompany                                 `json:"companies"`
	Candles     map[string]map[string][]SavedCandle            `json:"candles"`
	HomeCountry SavedHomeCountry                               `json:"home_country"`
	Foreign     []SavedForeignCountry                          `json:"foreign_countries"`
	PolicyLog   []world.ForeignPolicyAction                    `json:"policy_log"`
	SavedAt     time.Time                                      `json:"saved_at"`
}

// LoadWorldState reads the serialized simulation world state from disk.
func LoadWorldState(path string) (*SavedWorldState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state SavedWorldState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// SaveWorldState atomic writes the simulation world state to disk.
func SaveWorldState(path string, state *SavedWorldState) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	state.SavedAt = time.Now()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, path)
}

// ResetWorldStateDisk deletes the saved world state file.
func ResetWorldStateDisk(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ConvertToCompany instantiates a live *company.Company from a SavedCompany record.
func (sc *SavedCompany) ConvertToCompany() *company.Company {
	co := &company.Company{
		Symbol:            sc.Symbol,
		Name:              sc.Name,
		Sector:            company.Sector(sc.Sector),
		CapTier:           company.CapTier(sc.CapTier),
		TrueValue:         sc.TrueValue,
		ReportedValue:     sc.ReportedValue,
		SharesOutstanding: sc.SharesOutstanding,
		Float:             sc.Float,
		AnnualRevenue:     sc.AnnualRevenue,
		NetMargin:         sc.NetMargin,
		SectorMultiple:    sc.SectorMultiple,
		IPOPrice:          sc.IPOPrice,
		IsIPO:             sc.IsIPO,
		IPOTick:           sc.IPOTick,
		Headcount:         sc.Headcount,
		AverageWage:       sc.AverageWage,
		LaborExpense:      sc.LaborExpense,
		DebtOutstanding:   sc.DebtOutstanding,
		InterestExpense:   sc.InterestExpense,
		CorporateTaxPaid:  sc.CorporateTaxPaid,
		CapEx:             sc.CapEx,
		MacroDemandFactor: sc.MacroDemandFactor,
		IsPublic:          sc.IsPublic,
		Stage:             sc.Stage,
		PrivateValuation:  sc.PrivateValuation,
	}
	co.InitRuntime(nil)
	return co
}
