package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStorageLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_market.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Verify WAL mode
	var journalMode string
	if err := db.db.QueryRow("PRAGMA journal_mode;").Scan(&journalMode); err != nil {
		t.Fatalf("failed to query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("expected journal_mode=wal, got %s", journalMode)
	}

	writer := NewAsyncWriter(db, 100, 50*time.Millisecond)

	now := time.Now().UnixMilli()

	batch := &Batch{
		Trades: []TradeRecord{
			{Tick: 1, Timestamp: now, Symbol: "TEST", TakerID: "buyer1", MakerID: "mm1", Side: "BUY", Price: 100.50, Quantity: 200},
			{Tick: 1, Timestamp: now, Symbol: "TEST", TakerID: "seller1", MakerID: "mm1", Side: "SELL", Price: 100.40, Quantity: 150},
		},
		Prices: []PriceRecord{
			{Tick: 1, Timestamp: now, Symbol: "TEST", Mid: 100.45, Bid: 100.40, Ask: 100.50, Spread: 0.10},
		},
		Candles: []CandleRecord{
			{Symbol: "TEST", Timeframe: "1s", StartTime: now - 1000, EndTime: now, Open: 100.0, High: 101.0, Low: 99.5, Close: 100.45, Volume: 350, TickCount: 10},
		},
		Events: []EventRecord{
			{Tick: 1, Timestamp: now, Kind: "Fundamental", Symbol: "TEST", Details: `{"multiplier":1.05}`},
		},
		Accounts: []AccountRecord{
			{Tick: 1, Timestamp: now, AgentID: "buyer1", Cash: 50000.0, Equity: 70000.0, MarginRatio: 0.25},
		},
	}

	if !writer.Enqueue(batch) {
		t.Fatalf("failed to enqueue batch")
	}

	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Verify queries
	trades, err := db.GetRecentTrades("TEST", 10)
	if err != nil {
		t.Fatalf("GetRecentTrades failed: %v", err)
	}
	if len(trades) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(trades))
	}
	if trades[0].Quantity != 150 { // newest first
		t.Errorf("expected newest trade quantity 150, got %.2f", trades[0].Quantity)
	}

	prices, err := db.GetLatestPrices()
	if err != nil {
		t.Fatalf("GetLatestPrices failed: %v", err)
	}
	p, ok := prices["TEST"]
	if !ok {
		t.Fatalf("expected price record for TEST")
	}
	if p.Mid != 100.45 {
		t.Errorf("expected mid 100.45, got %.2f", p.Mid)
	}

	candles, err := db.GetCandles("TEST", "1s", 10)
	if err != nil {
		t.Fatalf("GetCandles failed: %v", err)
	}
	if len(candles) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(candles))
	}

	events, err := db.GetRecentEvents(10)
	if err != nil {
		t.Fatalf("GetRecentEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	accounts, err := db.GetAccountSnapshots("buyer1", 10)
	if err != nil {
		t.Fatalf("GetAccountSnapshots failed: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account snapshot, got %d", len(accounts))
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close failed: %v", err)
	}
}
