package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a SQLite database with WAL configuration and analytical query helpers.
type DB struct {
	db *sql.DB
}

// Open initializes a SQLite database connection with high-throughput WAL mode and ensures tables exist.
func Open(dbPath string) (*DB, error) {
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database at %s: %w", dbPath, err)
	}

	// SQLite connection pool: allow concurrent connections for WAL readers,
	// while the writer executes inside serialized transactions.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA foreign_keys = ON;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to apply pragma %q: %w", p, err)
		}
	}

	s := &DB{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run database migration: %w", err)
	}

	return s, nil
}

// SQLDB returns the underlying *sql.DB.
func (d *DB) SQLDB() *sql.DB {
	return d.db
}

// Close closes the underlying database.
func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS trades (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tick INTEGER NOT NULL,
		timestamp INTEGER NOT NULL,
		symbol TEXT NOT NULL,
		taker_id TEXT NOT NULL,
		maker_id TEXT NOT NULL,
		side TEXT NOT NULL,
		price REAL NOT NULL,
		quantity REAL NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_trades_sym_tick ON trades(symbol, tick);

	CREATE TABLE IF NOT EXISTS prices (
		tick INTEGER NOT NULL,
		timestamp INTEGER NOT NULL,
		symbol TEXT NOT NULL,
		mid REAL NOT NULL,
		bid REAL NOT NULL,
		ask REAL NOT NULL,
		spread REAL NOT NULL,
		PRIMARY KEY (tick, symbol)
	);
	CREATE INDEX IF NOT EXISTS idx_prices_sym_tick ON prices(symbol, tick);

	CREATE TABLE IF NOT EXISTS candles (
		symbol TEXT NOT NULL,
		timeframe TEXT NOT NULL,
		start_time INTEGER NOT NULL,
		end_time INTEGER NOT NULL,
		open REAL NOT NULL,
		high REAL NOT NULL,
		low REAL NOT NULL,
		close REAL NOT NULL,
		volume REAL NOT NULL,
		tick_count INTEGER NOT NULL,
		PRIMARY KEY (symbol, timeframe, start_time)
	);
	CREATE INDEX IF NOT EXISTS idx_candles_lookup ON candles(symbol, timeframe, start_time DESC);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tick INTEGER NOT NULL,
		timestamp INTEGER NOT NULL,
		kind TEXT NOT NULL,
		symbol TEXT NOT NULL,
		details TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_events_tick ON events(tick);

	CREATE TABLE IF NOT EXISTS account_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tick INTEGER NOT NULL,
		timestamp INTEGER NOT NULL,
		agent_id TEXT NOT NULL,
		cash REAL NOT NULL,
		equity REAL NOT NULL,
		margin_ratio REAL NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_accts_agent_tick ON account_snapshots(agent_id, tick);
	`
	_, err := d.db.Exec(schema)
	return err
}

// GetRecentTrades returns the most recent trades for a symbol, newest first.
func (d *DB) GetRecentTrades(symbol string, limit int) ([]TradeRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, tick, timestamp, symbol, taker_id, maker_id, side, price, quantity
		FROM trades
		WHERE symbol = ?
		ORDER BY id DESC
		LIMIT ?
	`
	rows, err := d.db.Query(query, symbol, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []TradeRecord
	for rows.Next() {
		var t TradeRecord
		if err := rows.Scan(&t.ID, &t.Tick, &t.Timestamp, &t.Symbol, &t.TakerID, &t.MakerID, &t.Side, &t.Price, &t.Quantity); err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}
	return trades, rows.Err()
}

// GetCandles returns historical OHLCV bars for a symbol and timeframe, newest first.
func (d *DB) GetCandles(symbol, timeframe string, limit int) ([]CandleRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT symbol, timeframe, start_time, end_time, open, high, low, close, volume, tick_count
		FROM candles
		WHERE symbol = ? AND timeframe = ?
		ORDER BY start_time DESC
		LIMIT ?
	`
	rows, err := d.db.Query(query, symbol, timeframe, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candles []CandleRecord
	for rows.Next() {
		var c CandleRecord
		if err := rows.Scan(&c.Symbol, &c.Timeframe, &c.StartTime, &c.EndTime, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume, &c.TickCount); err != nil {
			return nil, err
		}
		candles = append(candles, c)
	}
	return candles, rows.Err()
}

// GetLatestPrices returns the latest price quote for each traded symbol.
func (d *DB) GetLatestPrices() (map[string]PriceRecord, error) {
	query := `
		SELECT p.tick, p.timestamp, p.symbol, p.mid, p.bid, p.ask, p.spread
		FROM prices p
		INNER JOIN (
			SELECT symbol, MAX(tick) as max_tick
			FROM prices
			GROUP BY symbol
		) latest ON p.symbol = latest.symbol AND p.tick = latest.max_tick
	`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prices := make(map[string]PriceRecord)
	for rows.Next() {
		var p PriceRecord
		if err := rows.Scan(&p.Tick, &p.Timestamp, &p.Symbol, &p.Mid, &p.Bid, &p.Ask, &p.Spread); err != nil {
			return nil, err
		}
		prices[p.Symbol] = p
	}
	return prices, rows.Err()
}

// GetRecentEvents returns the most recent events, newest first.
func (d *DB) GetRecentEvents(limit int) ([]EventRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, tick, timestamp, kind, symbol, details
		FROM events
		ORDER BY id DESC
		LIMIT ?
	`
	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventRecord
	for rows.Next() {
		var e EventRecord
		if err := rows.Scan(&e.ID, &e.Tick, &e.Timestamp, &e.Kind, &e.Symbol, &e.Details); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// GetAccountSnapshots returns snapshots for a specific agent, newest first.
func (d *DB) GetAccountSnapshots(agentID string, limit int) ([]AccountRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT tick, timestamp, agent_id, cash, equity, margin_ratio
		FROM account_snapshots
		WHERE agent_id = ?
		ORDER BY id DESC
		LIMIT ?
	`
	rows, err := d.db.Query(query, agentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []AccountRecord
	for rows.Next() {
		var a AccountRecord
		if err := rows.Scan(&a.Tick, &a.Timestamp, &a.AgentID, &a.Cash, &a.Equity, &a.MarginRatio); err != nil {
			return nil, err
		}
		records = append(records, a)
	}
	return records, rows.Err()
}
