package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// AsyncWriter provides decoupled, asynchronous batch writing to SQLite.
type AsyncWriter struct {
	db            *DB
	queue         chan *Batch
	flushChan     chan chan error
	done          chan struct{}
	wg            sync.WaitGroup
	flushInterval time.Duration
	maxBatchSize  int
	closeOnce     sync.Once
}

// NewAsyncWriter constructs and starts an AsyncWriter.
func NewAsyncWriter(db *DB, bufferCap int, flushInterval time.Duration) *AsyncWriter {
	if bufferCap <= 0 {
		bufferCap = 1000
	}
	if flushInterval <= 0 {
		flushInterval = 250 * time.Millisecond
	}

	w := &AsyncWriter{
		db:            db,
		queue:         make(chan *Batch, bufferCap),
		flushChan:     make(chan chan error),
		done:          make(chan struct{}),
		flushInterval: flushInterval,
		maxBatchSize:  500,
	}

	w.wg.Add(1)
	go w.worker()

	return w
}

// Enqueue sends a batch to the asynchronous ingestion channel.
// Returns false if the queue is full or closed.
func (w *AsyncWriter) Enqueue(b *Batch) bool {
	if b == nil || b.Empty() {
		return true
	}
	select {
	case w.queue <- b:
		return true
	default:
		// Queue full - non-blocking drop or caller can handle
		return false
	}
}

// Flush triggers an immediate transaction commit of all currently queued batches and waits for completion.
func (w *AsyncWriter) Flush() error {
	resp := make(chan error, 1)
	select {
	case w.flushChan <- resp:
		return <-resp
	case <-w.done:
		return nil
	}
}

// Close stops ingestion, flushes all remaining items in the queue to disk, and closes the worker.
func (w *AsyncWriter) Close() error {
	var err error
	w.closeOnce.Do(func() {
		close(w.done)
		close(w.queue)
		w.wg.Wait()
	})
	return err
}

func (w *AsyncWriter) worker() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	var pending []*Batch

	flushPending := func() error {
		if len(pending) == 0 {
			return nil
		}
		merged := mergeBatches(pending)
		pending = pending[:0]
		return w.commitBatch(merged)
	}

	for {
		select {
		case b, ok := <-w.queue:
			if !ok {
				// Channel closed: flush remaining batches and exit
				_ = flushPending()
				return
			}
			if b != nil && !b.Empty() {
				pending = append(pending, b)
				if len(pending) >= w.maxBatchSize {
					_ = flushPending()
				}
			}

		case <-ticker.C:
			_ = flushPending()

		case resp := <-w.flushChan:
			// Drain all currently available items in the queue before flushing
			for {
				select {
				case b, ok := <-w.queue:
					if ok && b != nil && !b.Empty() {
						pending = append(pending, b)
						continue
					}
				default:
				}
				break
			}
			err := flushPending()
			resp <- err
		}
	}
}

func mergeBatches(batches []*Batch) *Batch {
	out := &Batch{}
	for _, b := range batches {
		out.Trades = append(out.Trades, b.Trades...)
		out.Prices = append(out.Prices, b.Prices...)
		out.Candles = append(out.Candles, b.Candles...)
		out.Events = append(out.Events, b.Events...)
		out.Accounts = append(out.Accounts, b.Accounts...)
	}
	return out
}

func (w *AsyncWriter) commitBatch(b *Batch) error {
	if b == nil || b.Empty() {
		return nil
	}

	tx, err := w.db.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // no-op if committed
	}()

	if len(b.Trades) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO trades (tick, timestamp, symbol, taker_id, maker_id, side, price, quantity)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("prepare trades: %w", err)
		}
		defer stmt.Close()

		for _, t := range b.Trades {
			if _, err := stmt.Exec(t.Tick, t.Timestamp, t.Symbol, t.TakerID, t.MakerID, t.Side, t.Price, t.Quantity); err != nil {
				return fmt.Errorf("exec trade: %w", err)
			}
		}
	}

	if len(b.Prices) > 0 {
		stmt, err := tx.Prepare(`
			INSERT OR REPLACE INTO prices (tick, timestamp, symbol, mid, bid, ask, spread)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("prepare prices: %w", err)
		}
		defer stmt.Close()

		for _, p := range b.Prices {
			if _, err := stmt.Exec(p.Tick, p.Timestamp, p.Symbol, p.Mid, p.Bid, p.Ask, p.Spread); err != nil {
				return fmt.Errorf("exec price: %w", err)
			}
		}
	}

	if len(b.Candles) > 0 {
		stmt, err := tx.Prepare(`
			INSERT OR REPLACE INTO candles (symbol, timeframe, start_time, end_time, open, high, low, close, volume, tick_count)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("prepare candles: %w", err)
		}
		defer stmt.Close()

		for _, c := range b.Candles {
			if _, err := stmt.Exec(c.Symbol, c.Timeframe, c.StartTime, c.EndTime, c.Open, c.High, c.Low, c.Close, c.Volume, c.TickCount); err != nil {
				return fmt.Errorf("exec candle: %w", err)
			}
		}
	}

	if len(b.Events) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO events (tick, timestamp, kind, symbol, details)
			VALUES (?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("prepare events: %w", err)
		}
		defer stmt.Close()

		for _, e := range b.Events {
			if _, err := stmt.Exec(e.Tick, e.Timestamp, e.Kind, e.Symbol, e.Details); err != nil {
				return fmt.Errorf("exec event: %w", err)
			}
		}
	}

	if len(b.Accounts) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO account_snapshots (tick, timestamp, agent_id, cash, equity, margin_ratio)
			VALUES (?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("prepare account_snapshots: %w", err)
		}
		defer stmt.Close()

		for _, a := range b.Accounts {
			if _, err := stmt.Exec(a.Tick, a.Timestamp, a.AgentID, a.Cash, a.Equity, a.MarginRatio); err != nil {
				return fmt.Errorf("exec account snapshot: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit batch: %w", err)
	}

	return nil
}
