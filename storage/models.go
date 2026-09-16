package storage

// TradeRecord represents an individual matched fill between taker and maker orders.
type TradeRecord struct {
	ID        int64   `json:"id"`
	Tick      int     `json:"tick"`
	Timestamp int64   `json:"timestamp"` // Unix timestamp in milliseconds
	Symbol    string  `json:"symbol"`
	TakerID   string  `json:"taker_id"`
	MakerID   string  `json:"maker_id"`
	Side      string  `json:"side"` // "BUY" or "SELL"
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
}

// PriceRecord represents top-of-book market data for a symbol at a specific tick.
type PriceRecord struct {
	Tick      int     `json:"tick"`
	Timestamp int64   `json:"timestamp"`
	Symbol    string  `json:"symbol"`
	Mid       float64 `json:"mid"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Spread    float64 `json:"spread"`
}

// CandleRecord represents an aggregated OHLCV bar across a timeframe window.
type CandleRecord struct {
	Symbol    string  `json:"symbol"`
	Timeframe string  `json:"timeframe"` // e.g. "1s", "5s", "1m", "5m"
	StartTime int64   `json:"start_time"`
	EndTime   int64   `json:"end_time"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
	TickCount int     `json:"tick_count"`
}

// EventRecord captures fundamental shocks, restatements, or liquidation events.
type EventRecord struct {
	ID        int64  `json:"id"`
	Tick      int    `json:"tick"`
	Timestamp int64  `json:"timestamp"`
	Kind      string `json:"kind"` // "Fundamental", "Restatement", "Liquidation"
	Symbol    string `json:"symbol"`
	Details   string `json:"details"`
}

// AccountRecord records periodic financial telemetry of a participant account.
type AccountRecord struct {
	Tick        int     `json:"tick"`
	Timestamp   int64   `json:"timestamp"`
	AgentID     string  `json:"agent_id"`
	Cash        float64 `json:"cash"`
	Equity      float64 `json:"equity"`
	MarginRatio float64 `json:"margin_ratio"`
}

// Batch bundles market records to be committed together in an atomic SQLite transaction.
type Batch struct {
	Trades   []TradeRecord
	Prices   []PriceRecord
	Candles  []CandleRecord
	Events   []EventRecord
	Accounts []AccountRecord
}

// Empty returns true if the batch contains no records.
func (b *Batch) Empty() bool {
	return len(b.Trades) == 0 &&
		len(b.Prices) == 0 &&
		len(b.Candles) == 0 &&
		len(b.Events) == 0 &&
		len(b.Accounts) == 0
}
