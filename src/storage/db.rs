use std::collections::HashMap;
use std::path::Path;
use std::sync::Arc;
use parking_lot::Mutex;
use rusqlite::{params, Connection};

use crate::storage::models::*;

/// DB wraps a SQLite database with WAL configuration and analytical query helpers.
/// It uses a mutex-wrapped connection (or connections) for thread-safe access.
#[derive(Clone)]
pub struct DB {
    conn: Arc<Mutex<Connection>>,
}

impl DB {
    /// Open initializes a SQLite database connection with high-throughput WAL mode and ensures tables exist.
    pub fn open<P: AsRef<Path>>(db_path: P) -> rusqlite::Result<Self> {
        let path = db_path.as_ref();
        if let Some(parent) = path.parent() {
            if !parent.as_os_str().is_empty() {
                let _ = std::fs::create_dir_all(parent);
            }
        }

        let conn = Connection::open(path)?;

        // WAL Pragmas
        conn.pragma_update(None, "journal_mode", "WAL")?;
        conn.pragma_update(None, "synchronous", "NORMAL")?;
        conn.pragma_update(None, "busy_timeout", 5000)?;
        conn.pragma_update(None, "foreign_keys", "ON")?;

        let db = Self {
            conn: Arc::new(Mutex::new(conn)),
        };
        db.migrate()?;
        Ok(db)
    }

    /// Open in-memory database for testing.
    pub fn open_in_memory() -> rusqlite::Result<Self> {
        let conn = Connection::open_in_memory()?;
        conn.pragma_update(None, "journal_mode", "WAL")?;
        conn.pragma_update(None, "synchronous", "NORMAL")?;
        conn.pragma_update(None, "busy_timeout", 5000)?;
        conn.pragma_update(None, "foreign_keys", "ON")?;

        let db = Self {
            conn: Arc::new(Mutex::new(conn)),
        };
        db.migrate()?;
        Ok(db)
    }

    fn migrate(&self) -> rusqlite::Result<()> {
        let conn = self.conn.lock();
        let schema = r#"
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
        "#;
        conn.execute_batch(schema)?;
        Ok(())
    }

    /// Direct raw handle lock
    pub fn raw_conn(&self) -> Arc<Mutex<Connection>> {
        self.conn.clone()
    }

    /// Commit a Batch within an atomic transaction.
    pub fn commit_batch(&self, b: &Batch) -> rusqlite::Result<()> {
        if b.is_empty() {
            return Ok(());
        }

        let mut conn = self.conn.lock();
        let tx = conn.transaction()?;

        if !b.trades.is_empty() {
            let mut stmt = tx.prepare_cached(
                "INSERT INTO trades (tick, timestamp, symbol, taker_id, maker_id, side, price, quantity)
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)",
            )?;
            for t in &b.trades {
                stmt.execute(params![
                    t.tick,
                    t.timestamp,
                    t.symbol,
                    t.taker_id,
                    t.maker_id,
                    t.side,
                    t.price,
                    t.quantity
                ])?;
            }
        }

        if !b.prices.is_empty() {
            let mut stmt = tx.prepare_cached(
                "INSERT OR REPLACE INTO prices (tick, timestamp, symbol, mid, bid, ask, spread)
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7)",
            )?;
            for p in &b.prices {
                stmt.execute(params![
                    p.tick,
                    p.timestamp,
                    p.symbol,
                    p.mid,
                    p.bid,
                    p.ask,
                    p.spread
                ])?;
            }
        }

        if !b.candles.is_empty() {
            let mut stmt = tx.prepare_cached(
                "INSERT OR REPLACE INTO candles (symbol, timeframe, start_time, end_time, open, high, low, close, volume, tick_count)
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10)",
            )?;
            for c in &b.candles {
                stmt.execute(params![
                    c.symbol,
                    c.timeframe,
                    c.start_time,
                    c.end_time,
                    c.open,
                    c.high,
                    c.low,
                    c.close,
                    c.volume,
                    c.tick_count
                ])?;
            }
        }

        if !b.events.is_empty() {
            let mut stmt = tx.prepare_cached(
                "INSERT INTO events (tick, timestamp, kind, symbol, details)
                 VALUES (?1, ?2, ?3, ?4, ?5)",
            )?;
            for e in &b.events {
                stmt.execute(params![e.tick, e.timestamp, e.kind, e.symbol, e.details])?;
            }
        }

        if !b.accounts.is_empty() {
            let mut stmt = tx.prepare_cached(
                "INSERT INTO account_snapshots (tick, timestamp, agent_id, cash, equity, margin_ratio)
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
            )?;
            for a in &b.accounts {
                stmt.execute(params![
                    a.tick,
                    a.timestamp,
                    a.agent_id,
                    a.cash,
                    a.equity,
                    a.margin_ratio
                ])?;
            }
        }

        tx.commit()?;
        Ok(())
    }

    /// GetRecentTrades returns the most recent trades for a symbol, newest first.
    pub fn get_recent_trades(&self, symbol: &str, limit: usize) -> rusqlite::Result<Vec<TradeRecord>> {
        let limit = if limit == 0 { 50 } else { limit };
        let conn = self.conn.lock();
        let mut stmt = conn.prepare(
            "SELECT id, tick, timestamp, symbol, taker_id, maker_id, side, price, quantity
             FROM trades
             WHERE symbol = ?1
             ORDER BY id DESC
             LIMIT ?2",
        )?;
        let rows = stmt.query_map(params![symbol, limit as i64], |row| {
            Ok(TradeRecord {
                id: row.get(0)?,
                tick: row.get(1)?,
                timestamp: row.get(2)?,
                symbol: row.get(3)?,
                taker_id: row.get(4)?,
                maker_id: row.get(5)?,
                side: row.get(6)?,
                price: row.get(7)?,
                quantity: row.get(8)?,
            })
        })?;

        let mut out = Vec::new();
        for r in rows {
            out.push(r?);
        }
        Ok(out)
    }

    /// GetCandles returns historical OHLCV bars for a symbol and timeframe, newest first.
    pub fn get_candles(&self, symbol: &str, timeframe: &str, limit: usize) -> rusqlite::Result<Vec<CandleRecord>> {
        let limit = if limit == 0 { 100 } else { limit };
        let conn = self.conn.lock();
        let mut stmt = conn.prepare(
            "SELECT symbol, timeframe, start_time, end_time, open, high, low, close, volume, tick_count
             FROM candles
             WHERE symbol = ?1 AND timeframe = ?2
             ORDER BY start_time DESC
             LIMIT ?3",
        )?;
        let rows = stmt.query_map(params![symbol, timeframe, limit as i64], |row| {
            Ok(CandleRecord {
                symbol: row.get(0)?,
                timeframe: row.get(1)?,
                start_time: row.get(2)?,
                end_time: row.get(3)?,
                open: row.get(4)?,
                high: row.get(5)?,
                low: row.get(6)?,
                close: row.get(7)?,
                volume: row.get(8)?,
                tick_count: row.get(9)?,
            })
        })?;

        let mut out = Vec::new();
        for r in rows {
            out.push(r?);
        }
        Ok(out)
    }

    /// GetLatestPrices returns the latest price quote for each traded symbol.
    pub fn get_latest_prices(&self) -> rusqlite::Result<HashMap<String, PriceRecord>> {
        let conn = self.conn.lock();
        let query = "
            SELECT p.tick, p.timestamp, p.symbol, p.mid, p.bid, p.ask, p.spread
            FROM prices p
            INNER JOIN (
                SELECT symbol, MAX(tick) as max_tick
                FROM prices
                GROUP BY symbol
            ) latest ON p.symbol = latest.symbol AND p.tick = latest.max_tick
        ";
        let mut stmt = conn.prepare(query)?;
        let rows = stmt.query_map([], |row| {
            Ok(PriceRecord {
                tick: row.get(0)?,
                timestamp: row.get(1)?,
                symbol: row.get(2)?,
                mid: row.get(3)?,
                bid: row.get(4)?,
                ask: row.get(5)?,
                spread: row.get(6)?,
            })
        })?;

        let mut out = HashMap::new();
        for r in rows {
            let p = r?;
            out.insert(p.symbol.clone(), p);
        }
        Ok(out)
    }

    /// GetRecentEvents returns the most recent events, newest first.
    pub fn get_recent_events(&self, limit: usize) -> rusqlite::Result<Vec<EventRecord>> {
        let limit = if limit == 0 { 50 } else { limit };
        let conn = self.conn.lock();
        let mut stmt = conn.prepare(
            "SELECT id, tick, timestamp, kind, symbol, details
             FROM events
             ORDER BY id DESC
             LIMIT ?1",
        )?;
        let rows = stmt.query_map(params![limit as i64], |row| {
            Ok(EventRecord {
                id: row.get(0)?,
                tick: row.get(1)?,
                timestamp: row.get(2)?,
                kind: row.get(3)?,
                symbol: row.get(4)?,
                details: row.get(5)?,
            })
        })?;

        let mut out = Vec::new();
        for r in rows {
            out.push(r?);
        }
        Ok(out)
    }

    /// GetAccountSnapshots returns snapshots for a specific agent, newest first.
    pub fn get_account_snapshots(&self, agent_id: &str, limit: usize) -> rusqlite::Result<Vec<AccountRecord>> {
        let limit = if limit == 0 { 50 } else { limit };
        let conn = self.conn.lock();
        let mut stmt = conn.prepare(
            "SELECT tick, timestamp, agent_id, cash, equity, margin_ratio
             FROM account_snapshots
             WHERE agent_id = ?1
             ORDER BY id DESC
             LIMIT ?2",
        )?;
        let rows = stmt.query_map(params![agent_id, limit as i64], |row| {
            Ok(AccountRecord {
                tick: row.get(0)?,
                timestamp: row.get(1)?,
                agent_id: row.get(2)?,
                cash: row.get(3)?,
                equity: row.get(4)?,
                margin_ratio: row.get(5)?,
            })
        })?;

        let mut out = Vec::new();
        for r in rows {
            out.push(r?);
        }
        Ok(out)
    }
}
