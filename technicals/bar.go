// Package technicals computes standard technical-analysis indicators
// (SMA, EMA, RSI, MACD, Bollinger Bands, ATR, swing highs/lows) from
// OHLC price bars, mirroring how real trading platforms and technical
// literature define them — not simplified approximations. It has no
// knowledge of orders, accounts, or agents; it operates purely on
// price history, which is why it sits alongside package market rather
// than inside it.
//
// This package is also what gives multi-timeframe behavior real
// meaning in this simulation: an Aggregator with a small window
// produces bars a "scalper"-style trader would react to, while a
// larger window produces bars a "swing"-style trader would react to,
// from the same underlying tick stream — matching how real markets
// let a 1-minute chart and a daily chart tell very different stories
// about the same instrument.
package technicals

// Bar is one OHLC (open/high/low/close) price bar aggregated from a
// fixed number of underlying ticks, mirroring a real candlestick.
type Bar struct {
	Open  float64
	High  float64
	Low   float64
	Close float64

	// TickCount is how many raw ticks were aggregated into this bar —
	// mainly useful for diagnostics/tests; a fully-formed bar should
	// always have TickCount equal to the Aggregator's configured
	// window size.
	TickCount int
}

// Aggregator rolls a stream of raw tick-level prices into fixed-size
// OHLC bars, maintaining a rolling history of completed bars up to
// maxHistory. This is the multi-timeframe mechanism: constructing two
// Aggregators with different windowSize values over the same
// underlying tick stream produces genuinely different bar series
// (e.g. one bar every 5 ticks vs. one every 50), exactly as a 5-minute
// chart and a 1-hour chart are built from the same underlying trades
// in a real market.
type Aggregator struct {
	windowSize int
	maxHistory int

	current        Bar
	ticksInCurrent int
	hasCurrent     bool

	completed []Bar // oldest first
}

// NewAggregator constructs an Aggregator that closes a bar every
// windowSize ticks and retains up to maxHistory completed bars
// (oldest dropped once exceeded).
func NewAggregator(windowSize, maxHistory int) *Aggregator {
	if windowSize < 1 {
		windowSize = 1
	}
	if maxHistory < 1 {
		maxHistory = 1
	}
	return &Aggregator{windowSize: windowSize, maxHistory: maxHistory}
}

// AddTick feeds one raw tick price into the aggregator. Once
// windowSize ticks have been accumulated, the in-progress bar closes
// and is appended to the completed history.
func (a *Aggregator) AddTick(price float64) {
	if !a.hasCurrent {
		a.current = Bar{Open: price, High: price, Low: price, Close: price, TickCount: 0}
		a.hasCurrent = true
	}

	if price > a.current.High {
		a.current.High = price
	}
	if price < a.current.Low {
		a.current.Low = price
	}
	a.current.Close = price
	a.current.TickCount++
	a.ticksInCurrent++

	if a.ticksInCurrent >= a.windowSize {
		a.completed = append(a.completed, a.current)
		if len(a.completed) > a.maxHistory {
			a.completed = a.completed[len(a.completed)-a.maxHistory:]
		}
		a.hasCurrent = false
		a.ticksInCurrent = 0
	}
}

// Bars returns the completed bar history, oldest first. The
// in-progress (not yet closed) bar is intentionally excluded —
// indicators should be computed from settled bars, matching how a
// real chart's current, still-forming candle is typically treated
// separately from closed history.
func (a *Aggregator) Bars() []Bar {
	return a.completed
}

// WindowSize reports how many raw ticks this aggregator rolls into
// each bar.
func (a *Aggregator) WindowSize() int {
	return a.windowSize
}
