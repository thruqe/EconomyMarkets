package main

import (
	"math"
	"strings"
	"sync"

	"economy/storage"
)

// Standard trader timeframes:
// 1m, 5m, 15, 30, 45, 1h, 2h, 4h, 1D, 1W, 1M, 1Y
const (
	TF1m  = "1m"
	TF5m  = "5m"
	TF15m = "15m"
	TF30m = "30m"
	TF45m = "45m"
	TF1h  = "1h"
	TF2h  = "2h"
	TF4h  = "4h"
	TF1D  = "1D"
	TF1W  = "1W"
	TF1M  = "1M"
	TF1Y  = "1Y"
)

var AllTimeframes = []string{
	TF1m, TF5m, TF15m, TF30m, TF45m, TF1h, TF2h, TF4h, TF1D, TF1W, TF1M, TF1Y,
}

// TimeframeDurationMillis returns the exact duration of a candle in milliseconds
var TimeframeDurationMillis = map[string]int64{
	TF1m:  60 * 1000,
	TF5m:  5 * 60 * 1000,
	TF15m: 15 * 60 * 1000,
	TF30m: 30 * 60 * 1000,
	TF45m: 45 * 60 * 1000,
	TF1h:  60 * 60 * 1000,
	TF2h:  2 * 60 * 60 * 1000,
	TF4h:  4 * 60 * 60 * 1000,
	TF1D:  24 * 60 * 60 * 1000,
	TF1W:  7 * 24 * 60 * 60 * 1000,
	TF1M:  30 * 24 * 60 * 60 * 1000,
	TF1Y:  365 * 24 * 60 * 60 * 1000,
}

// NormalizeTimeframe maps user aliases ("15", "w", "m", "y", etc.) to canonical IDs
func NormalizeTimeframe(tf string) string {
	raw := strings.TrimSpace(tf)
	// Check Minute: "1m", "m", "1min", "minute"
	if raw == "1m" || raw == "m" || strings.EqualFold(raw, "1min") || strings.EqualFold(raw, "min") || strings.EqualFold(raw, "minute") {
		return TF1m
	}
	// Check Monthly: "1M", "M", "1mo", "mo", "month", "monthly"
	if raw == "1M" || raw == "M" || strings.EqualFold(raw, "1mo") || strings.EqualFold(raw, "mo") || strings.EqualFold(raw, "month") || strings.EqualFold(raw, "monthly") {
		return TF1M
	}

	lower := strings.ToLower(raw)
	switch lower {
	case "5m", "5min":
		return TF5m
	case "15", "15m", "15min":
		return TF15m
	case "30", "30m", "30min":
		return TF30m
	case "45", "45m", "45min":
		return TF45m
	case "1h", "60m":
		return TF1h
	case "2h", "120m":
		return TF2h
	case "4h", "240m":
		return TF4h
	case "d", "1d", "day", "daily":
		return TF1D
	case "w", "1w", "week", "weekly":
		return TF1W
	case "y", "1y", "year", "yearly":
		return TF1Y
	}

	if raw == "D" {
		return TF1D
	}
	if raw == "W" {
		return TF1W
	}
	if raw == "Y" {
		return TF1Y
	}

	return TF1m
}

// SymbolTimeframeBars stores the historical completed bars and live in-progress bar
type SymbolTimeframeBars struct {
	mu            sync.RWMutex
	symbol        string
	completedBars map[string][]storage.CandleRecord
	currentBar    map[string]*storage.CandleRecord
	currentBarEnd map[string]int64
}

func NewSymbolTimeframeBars(symbol string) *SymbolTimeframeBars {
	stb := &SymbolTimeframeBars{
		symbol:        symbol,
		completedBars: make(map[string][]storage.CandleRecord),
		currentBar:    make(map[string]*storage.CandleRecord),
		currentBarEnd: make(map[string]int64),
	}

	for _, tf := range AllTimeframes {
		stb.completedBars[tf] = make([]storage.CandleRecord, 0, 3000)
	}

	return stb
}

// MultiTimeframeManager manages all symbols across all standard timeframes
type MultiTimeframeManager struct {
	mu      sync.RWMutex
	symbols map[string]*SymbolTimeframeBars
	db      *storage.DB
}

func NewMultiTimeframeManager(db *storage.DB) *MultiTimeframeManager {
	return &MultiTimeframeManager{
		symbols: make(map[string]*SymbolTimeframeBars),
		db:      db,
	}
}

func (m *MultiTimeframeManager) RegisterSymbol(symbol string) *SymbolTimeframeBars {
	m.mu.Lock()
	defer m.mu.Unlock()

	if stb, exists := m.symbols[symbol]; exists {
		return stb
	}

	stb := NewSymbolTimeframeBars(symbol)
	m.symbols[symbol] = stb
	return stb
}

func (m *MultiTimeframeManager) Get(symbol string) *SymbolTimeframeBars {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.symbols[symbol]
}

// InitLiveCandles initializes live in-progress candles for all timeframes at the starting price.
func (stb *SymbolTimeframeBars) InitLiveCandles(symbol string, startPrice float64, simStartTime int64) {
	stb.mu.Lock()
	defer stb.mu.Unlock()

	if startPrice <= 0 {
		startPrice = 100.0
	}
	for _, tf := range AllTimeframes {
		durationMs := TimeframeDurationMillis[tf]
		stb.completedBars[tf] = nil
		stb.currentBar[tf] = &storage.CandleRecord{
			Symbol:    symbol,
			Timeframe: tf,
			StartTime: simStartTime,
			EndTime:   simStartTime + durationMs,
			Open:      startPrice,
			High:      startPrice,
			Low:       startPrice,
			Close:     startPrice,
			Volume:    0,
			TickCount: 0,
		}
		stb.currentBarEnd[tf] = simStartTime + durationMs
	}
}

// AddTick advances the multi-timeframe candles based on simulated market time.
// If a candle's timeframe duration has elapsed, it is closed and added to history.
// Otherwise, the current in-progress candle updates in real-time.
func (stb *SymbolTimeframeBars) AddTick(simTimeMillis int64, price float64, tradeVolume float64, retailVolume float64) []storage.CandleRecord {
	if price <= 0 {
		return nil
	}

	stb.mu.Lock()
	defer stb.mu.Unlock()

	var closedBars []storage.CandleRecord

	for _, tf := range AllTimeframes {
		durationMs := TimeframeDurationMillis[tf]
		cur := stb.currentBar[tf]
		if cur == nil {
			cur = &storage.CandleRecord{
				Symbol:    stb.symbol,
				Timeframe: tf,
				StartTime: simTimeMillis,
				EndTime:   simTimeMillis + durationMs,
				Open:      price,
				High:      price,
				Low:       price,
				Close:     price,
				Volume:    tradeVolume,
				TickCount: 1,
			}
			stb.currentBar[tf] = cur
			stb.currentBarEnd[tf] = simTimeMillis + durationMs
			continue
		}

		if simTimeMillis >= stb.currentBarEnd[tf] {
			// Current candle period has finished! Complete and close it.
			prevClose := cur.Close
			cur.EndTime = stb.currentBarEnd[tf]
			stb.completedBars[tf] = append(stb.completedBars[tf], *cur)
			if len(stb.completedBars[tf]) > 3000 {
				stb.completedBars[tf] = stb.completedBars[tf][len(stb.completedBars[tf])-3000:]
			}
			closedBars = append(closedBars, *cur)

			// Start next candle seamlessly from prevClose
			nextStart := stb.currentBarEnd[tf]
			nextEnd := nextStart + durationMs
			for simTimeMillis >= nextEnd {
				nextStart = nextEnd
				nextEnd = nextStart + durationMs
			}

			stb.currentBar[tf] = &storage.CandleRecord{
				Symbol:    stb.symbol,
				Timeframe: tf,
				StartTime: nextStart,
				EndTime:   nextEnd,
				Open:      prevClose,
				High:      math.Max(prevClose, price),
				Low:       math.Min(prevClose, price),
				Close:     price,
				Volume:    tradeVolume,
				TickCount: 1,
			}
			stb.currentBarEnd[tf] = nextEnd
		} else {
			// Update the forming in-progress candle
			if cur.TickCount == 0 {
				cur.Open = price
				cur.High = price
				cur.Low = price
				cur.Close = price
				cur.Volume = tradeVolume
				cur.TickCount = 1
			} else {
				if price > cur.High {
					cur.High = price
				}
				if price < cur.Low {
					cur.Low = price
				}
				cur.Close = price
				cur.Volume += tradeVolume
				cur.TickCount++
			}
		}
	}

	return closedBars
}

// GetCurrentCandle returns a copy of the live in-progress candle for the given timeframe
func (stb *SymbolTimeframeBars) GetCurrentCandle(tf string) (storage.CandleRecord, bool) {
	stb.mu.RLock()
	defer stb.mu.RUnlock()

	normTF := NormalizeTimeframe(tf)
	cur := stb.currentBar[normTF]
	if cur == nil {
		return storage.CandleRecord{}, false
	}
	return *cur, true
}

// GetBars returns the historical completed bars plus the live in-progress bar
func (stb *SymbolTimeframeBars) GetBars(tf string, limit int) []storage.CandleRecord {
	stb.mu.RLock()
	defer stb.mu.RUnlock()

	normTF := NormalizeTimeframe(tf)
	list := stb.completedBars[normTF]
	if limit <= 0 || limit > len(list) {
		limit = len(list)
	}

	result := make([]storage.CandleRecord, 0, limit+1)
	if len(list) > 0 {
		startIdx := max(len(list)-limit, 0)
		result = append(result, list[startIdx:]...)
	}

	// Append the currently active forming bar
	if cur := stb.currentBar[normTF]; cur != nil {
		result = append(result, *cur)
	}

	return result
}
