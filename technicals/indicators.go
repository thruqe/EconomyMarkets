package technicals

import "math"

// closes extracts closing prices from bars, the input most indicators
// operate on.
func closes(bars []Bar) []float64 {
	out := make([]float64, len(bars))
	for i, b := range bars {
		out[i] = b.Close
	}
	return out
}

// SMA computes the simple moving average of closing price over the
// most recent window bars. Returns (0, false) if fewer than window
// bars are available.
func SMA(bars []Bar, window int) (float64, bool) {
	if window < 1 || len(bars) < window {
		return 0, false
	}
	recent := bars[len(bars)-window:]
	var sum float64
	for _, b := range recent {
		sum += b.Close
	}
	return sum / float64(window), true
}

// EMA computes the exponential moving average of closing price over
// window bars, using the standard smoothing factor α = 2/(window+1)
// and seeding the initial value with the SMA of the first window
// bars — the conventional way real platforms initialize an EMA series
// rather than seeding from a single point.
func EMA(bars []Bar, window int) (float64, bool) {
	if window < 1 || len(bars) < window {
		return 0, false
	}
	alpha := 2.0 / (float64(window) + 1.0)

	seed, ok := SMA(bars[:window], window)
	if !ok {
		return 0, false
	}

	ema := seed
	for _, b := range bars[window:] {
		ema = alpha*b.Close + (1-alpha)*ema
	}
	return ema, true
}

// RSI computes the Relative Strength Index over window bars using
// Wilder's original method: average gains and average losses over the
// window (via Wilder smoothing, not a simple average, matching the
// standard definition), giving RSI = 100 - 100/(1+RS) where
// RS = avgGain/avgLoss. Returns a value in [0, 100]; conventionally,
// values above 70 are read as overbought and below 30 as oversold.
// Returns (0, false) if fewer than window+1 bars are available (RSI
// needs window price *changes*, which requires window+1 prices).
func RSI(bars []Bar, window int) (float64, bool) {
	if window < 1 || len(bars) < window+1 {
		return 0, false
	}

	c := closes(bars)

	// Wilder smoothing: seed with a simple average of the first
	// window changes, then smooth subsequent changes in with weight
	// 1/window — this is Wilder's original method, distinct from
	// (and more standard than) a plain rolling average of gains/losses.
	var sumGain, sumLoss float64
	for i := 1; i <= window; i++ {
		change := c[i] - c[i-1]
		if change > 0 {
			sumGain += change
		} else {
			sumLoss += -change
		}
	}
	avgGain := sumGain / float64(window)
	avgLoss := sumLoss / float64(window)

	for i := window + 1; i < len(c); i++ {
		change := c[i] - c[i-1]
		gain, loss := 0.0, 0.0
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}
		avgGain = (avgGain*float64(window-1) + gain) / float64(window)
		avgLoss = (avgLoss*float64(window-1) + loss) / float64(window)
	}

	if avgLoss == 0 {
		if avgGain == 0 {
			return 50, true // no movement at all: neutral RSI by convention
		}
		return 100, true // no losses at all over the window: maximally overbought
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	return rsi, true
}

// MACDResult holds the three standard MACD output series values at
// the current point: the MACD line itself, its signal line, and their
// difference (the "histogram").
type MACDResult struct {
	MACD      float64
	Signal    float64
	Histogram float64
}

// MACD computes the standard Moving Average Convergence/Divergence:
// MACD line = EMA(fastWindow) - EMA(slowWindow), Signal = EMA of the
// MACD line over signalWindow bars, Histogram = MACD - Signal.
// Conventional defaults in real platforms are fast=12, slow=26,
// signal=9 bars, but this implementation takes them as parameters
// rather than hardcoding them, since this simulation's bar windows
// (see Aggregator) don't necessarily correspond to the daily bars
// those defaults were designed around.
func MACD(bars []Bar, fastWindow, slowWindow, signalWindow int) (MACDResult, bool) {
	if len(bars) < slowWindow+signalWindow {
		return MACDResult{}, false
	}

	// Build the MACD line series (one value per bar position once
	// enough history exists) so the signal line can be an EMA of that
	// series rather than of raw price.
	macdSeries := make([]Bar, 0, len(bars)-slowWindow+1)
	for i := slowWindow; i <= len(bars); i++ {
		window := bars[:i]
		fast, ok1 := EMA(window, fastWindow)
		slow, ok2 := EMA(window, slowWindow)
		if !ok1 || !ok2 {
			continue
		}
		macdSeries = append(macdSeries, Bar{Close: fast - slow})
	}

	if len(macdSeries) < signalWindow {
		return MACDResult{}, false
	}

	signal, ok := EMA(macdSeries, signalWindow)
	if !ok {
		return MACDResult{}, false
	}

	macdNow := macdSeries[len(macdSeries)-1].Close
	return MACDResult{
		MACD:      macdNow,
		Signal:    signal,
		Histogram: macdNow - signal,
	}, true
}

// BollingerBands computes the standard Bollinger Bands: a middle band
// (SMA over window), and upper/lower bands offset by k standard
// deviations of closing price over that same window. k=2 is the
// conventional default in real usage, but is a parameter here rather
// than hardcoded.
type BollingerResult struct {
	Middle float64
	Upper  float64
	Lower  float64
}

func BollingerBands(bars []Bar, window int, k float64) (BollingerResult, bool) {
	middle, ok := SMA(bars, window)
	if !ok {
		return BollingerResult{}, false
	}

	recent := bars[len(bars)-window:]
	var sumSquaredDiff float64
	for _, b := range recent {
		diff := b.Close - middle
		sumSquaredDiff += diff * diff
	}
	stddev := math.Sqrt(sumSquaredDiff / float64(window))

	return BollingerResult{
		Middle: middle,
		Upper:  middle + k*stddev,
		Lower:  middle - k*stddev,
	}, true
}

// trueRange computes a single bar's true range against the previous
// bar's close: max(high-low, |high-prevClose|, |low-prevClose|). This
// is the standard definition (Wilder) that correctly accounts for
// gaps between bars, which a plain high-low range would miss.
func trueRange(current, previous Bar) float64 {
	highLow := current.High - current.Low
	highPrevClose := math.Abs(current.High - previous.Close)
	lowPrevClose := math.Abs(current.Low - previous.Close)
	return math.Max(highLow, math.Max(highPrevClose, lowPrevClose))
}

// ATR computes the Average True Range over window bars: the average
// of each bar's true range (see trueRange) across the window. Used in
// real markets as the standard input to volatility-aware stop-loss
// placement, since it accounts for gap risk that a simple high-low
// range or close-to-close volatility measure would miss.
func ATR(bars []Bar, window int) (float64, bool) {
	if window < 1 || len(bars) < window+1 {
		return 0, false
	}

	recent := bars[len(bars)-window-1:] // need one extra bar for the first true range's "previous close"
	var sum float64
	for i := 1; i < len(recent); i++ {
		sum += trueRange(recent[i], recent[i-1])
	}
	return sum / float64(window), true
}

// SwingHighLow finds the local extremum swing high and swing low
// within the most recent window bars: a swing high is a bar whose
// High is greater than or equal to every other bar's High in the
// window (and symmetrically for swing low), matching the standard
// technical-analysis definition of a swing point as a local extreme
// within a lookback window, rather than just "the single highest
// close" or an arbitrary boundary value.
func SwingHighLow(bars []Bar, window int) (swingHigh, swingLow float64, ok bool) {
	if window < 1 || len(bars) < window {
		return 0, 0, false
	}
	recent := bars[len(bars)-window:]

	swingHigh = recent[0].High
	swingLow = recent[0].Low
	for _, b := range recent[1:] {
		if b.High > swingHigh {
			swingHigh = b.High
		}
		if b.Low < swingLow {
			swingLow = b.Low
		}
	}
	return swingHigh, swingLow, true
}

// Momentum computes the simple percentage change in closing price
// from window bars ago to the most recent bar:
// (latest - past) / past. This is the plain rate-of-change measure
// many retail-style "is it trending" heuristics use directly, distinct
// from RSI's more involved gain/loss-averaged oscillator.
func Momentum(bars []Bar, window int) (float64, bool) {
	if window < 1 || len(bars) <= window {
		return 0, false
	}
	past := bars[len(bars)-window-1].Close
	latest := bars[len(bars)-1].Close
	if past == 0 {
		return 0, false
	}
	return (latest - past) / past, true
}
