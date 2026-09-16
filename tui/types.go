package tui

import (
	"time"

	"economy/citizen"
	"economy/company"
	"economy/country"
	"economy/market"
	"economy/sim"
	"economy/world"
)

type Tab int

const (
	TabOverview Tab = iota
	TabUniverse
	TabCharts
	TabDetail
	TabTrade
	TabNews
	TabAdmin
	TabWorld
	TabPolicy
)

func (t Tab) String() string {
	switch t {
	case TabOverview:
		return "1. Overview"
	case TabUniverse:
		return "2. 500 Stocks"
	case TabCharts:
		return "3. Charts"
	case TabDetail:
		return "4. Deep Dive & DOM"
	case TabTrade:
		return "5. Portfolio & Trade"
	case TabNews:
		return "6. News & Tape"
	case TabAdmin:
		return "7. Government & Admin"
	case TabWorld:
		return "8. World & Nations"
	case TabPolicy:
		return "9. Policy Studio"
	default:
		return "Unknown"
	}
}

type CompanyRow struct {
	Symbol            string
	Name              string
	Sector            string
	CapTier           string
	Price             float64
	Bid               float64
	Ask               float64
	Spread            float64
	PrevPrice         float64
	Change            float64
	ChangePct         float64
	Volume            float64
	ReportedValue     float64
	TrueValue         float64
	PE                float64
	PS                float64
	MarketCap         float64
	IsIPO             bool
	Headcount         float64
	LaborExpense      float64
	DebtOutstanding   float64
	InterestExpense   float64
	CorporateTaxPaid  float64
	MacroDemandFactor float64
	IsPublic          bool
	Stage             string
	PrivateValuation  float64
}

type OrderBookLevel struct {
	Price float64
	Size  float64
}

type OrderBookView struct {
	Symbol string
	Bids   []OrderBookLevel
	Asks   []OrderBookLevel
	Mid    float64
	Spread float64
}

type PositionView struct {
	Symbol        string
	Side          string
	Quantity      float64
	EntryPrice    float64
	MarkPrice     float64
	TotalValue    float64
	Unrealized    float64
	UnrealizedPct float64
	OpenedAt      int64
	OpenedTime    time.Time
}

type CandleView struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// TickMsg is delivered to the Bubble Tea update loop on every simulation step
type TickMsg struct {
	Report      *sim.TickReport
	SimTime     int64
	Speed       float64
	ActiveCount int
}

// SimBridge provides thread-safe access to simulation data and actions
type SimBridge interface {
	GetUniverse() []*company.Company
	GetPrices() map[string]sim.PriceQuote
	GetBook(symbol string) *market.OrderBook
	GetHumanAccount() *market.Account
	GetCandles(symbol string, tf string) []CandleView
	SubmitOrder(symbol string, side market.Side, isMarket bool, qty float64, price float64) error
	ClosePosition(symbol string) error
	SetSpeed(speed float64)
	LaunchIPO(name, symbol string, sector company.Sector, tier company.CapTier, shares, price float64) error
	ResetPortfolio() error
	SavePortfolio() error
	GetMacroReport() (citizen.CitizenReport, country.NationalReport)
	BailoutCompany(symbol string) error
	TogglePolicy(policyID string) (bool, string)
	SaveWorldState() error
	ResetWorldState() error
	// World & foreign policy
	GetWorld() *world.World
	SetForeignRelation(countryID string, stance world.DiplomaticStance, tariffRate float64) bool
	NegotiateTradeDeal(countryID string, newTariff float64) bool
	SendForeignAid(countryID string, amount float64) bool
	ImposeSanctions(countryID string, level int) bool
	GetSimDate() world.SimDate
	// Sovereign Treasury & Nation Building
	BorrowMoney(amount float64) error
	RepayDebt(amount float64) error
	InvestInfrastructure(amount float64) error
	InvestPopulation(amount float64, pillar string) error
	InvestEnterprise(amount float64) error
	CharterStockExchange() error
}

