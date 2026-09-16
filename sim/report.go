package sim

import (
	"time"

	"economy/citizen"
	"economy/country"
)

// PriceQuote represents top-of-book market quote for one symbol at one tick.
type PriceQuote struct {
	Symbol string  `json:"symbol"`
	Mid    float64 `json:"mid"`
	Bid    float64 `json:"bid"`
	Ask    float64 `json:"ask"`
	Spread float64 `json:"spread"`
	HasMid bool    `json:"has_mid"`
}

// Trade represents an executed match between taker and maker on a specific symbol.
type Trade struct {
	Tick         int     `json:"tick"`
	Symbol       string  `json:"symbol"`
	TakerAgentID string  `json:"taker_id"`
	MakerAgentID string  `json:"maker_id"`
	Side         string  `json:"side"` // "BUY" or "SELL"
	Price        float64 `json:"price"`
	Quantity     float64 `json:"quantity"`
}

// AccountSnapshot captures an account's financial telemetry at a tick.
type AccountSnapshot struct {
	AgentID     string  `json:"agent_id"`
	Cash        float64 `json:"cash"`
	Equity      float64 `json:"equity"`
	MarginRatio float64 `json:"margin_ratio"`
}

// TickReport is returned by Simulation.Step(), summarizing all activity during that tick.
type TickReport struct {
	Tick      int                    `json:"tick"`
	Timestamp time.Time              `json:"timestamp"`
	Prices    map[string]PriceQuote  `json:"prices"`
	Trades    []Trade                `json:"trades"`
	Events    []Event                `json:"events"`
	Accounts  []AccountSnapshot      `json:"accounts"`
	Citizen   citizen.CitizenReport  `json:"citizen"`
	National  country.NationalReport `json:"national"`
}
