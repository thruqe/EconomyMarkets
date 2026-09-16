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

// Bar is one OHLCV price bar aggregated from underlying ticks and trades.
type Bar struct {
	Open         float64
	High         float64
	Low          float64
	Close        float64
	Volume       float64
	RetailVolume float64

	// TickCount is how many raw ticks were aggregated into this bar.
	TickCount int
}

// Aggregator rolls a stream of raw tick-level prices and trade volumes into fixed-size
// OHLCV bars, maintaining a rolling history of completed bars up to maxHistory.
type Aggregator struct {
	windowSize int
	maxHistory int

	current        Bar
	ticksInCurrent int
	hasCurrent     bool

	completed []Bar // oldest first
}

// NewAggregator constructs an Aggregator that closes a bar every
// windowSize ticks and retains up to maxHistory completed bars.
func NewAggregator(windowSize, maxHistory int) *Aggregator {
	if windowSize < 1 {
		windowSize = 1
	}
	if maxHistory < 1 {
		maxHistory = 1
	}
	return &Aggregator{windowSize: windowSize, maxHistory: maxHistory}
}

// AddTick feeds one raw tick price into the aggregator with zero volume.
func (a *Aggregator) AddTick(price float64) {
	a.AddTickWithVolume(price, 0, 0)
}

// AddTickWithVolume feeds one raw tick price and executed volume into the aggregator.
func (a *Aggregator) AddTickWithVolume(price float64, volume float64, retailVolume float64) {
	if !a.hasCurrent {
		a.current = Bar{
			Open:         price,
			High:         price,
			Low:          price,
			Close:        price,
			Volume:       volume,
			RetailVolume: retailVolume,
			TickCount:    1,
		}
		a.hasCurrent = true
		a.ticksInCurrent = 1
	} else {
		if price > a.current.High {
			a.current.High = price
		}
		if price < a.current.Low {
			a.current.Low = price
		}
		a.current.Close = price
		a.current.Volume += volume
		a.current.RetailVolume += retailVolume
		a.current.TickCount++
		a.ticksInCurrent++
	}

	if a.ticksInCurrent >= a.windowSize {
		a.completed = append(a.completed, a.current)
		if len(a.completed) > a.maxHistory {
			a.completed = a.completed[len(a.completed)-a.maxHistory:]
		}
		a.hasCurrent = false
		a.ticksInCurrent = 0
	}
}

// Current returns the in-progress forming bar, if any.
func (a *Aggregator) Current() (Bar, bool) {
	if !a.hasCurrent {
		return Bar{}, false
	}
	return a.current, true
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
