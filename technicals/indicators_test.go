package technicals

import (
	"math"
	"testing"
)

func closeBars(prices []float64) []Bar {
	bars := make([]Bar, len(prices))
	for i, p := range prices {
		bars[i] = Bar{Open: p, High: p, Low: p, Close: p, TickCount: 1}
	}
	return bars
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

// TestAggregatorProducesCorrectOHLC confirms bars capture true
// open/high/low/close over their window, not just averages.
func TestAggregatorProducesCorrectOHLC(t *testing.T) {
	agg := NewAggregator(5, 10)
	ticks := []float64{100, 105, 98, 102, 101} // open=100, high=105, low=98, close=101

	for _, tick := range ticks {
		agg.AddTick(tick)
	}

	bars := agg.Bars()
	if len(bars) != 1 {
		t.Fatalf("expected exactly 1 completed bar after 5 ticks with window 5, got %d", len(bars))
	}

	b := bars[0]
	if b.Open != 100 || b.High != 105 || b.Low != 98 || b.Close != 101 {
		t.Fatalf("bar OHLC incorrect: got O=%.2f H=%.2f L=%.2f C=%.2f, want O=100 H=105 L=98 C=101",
			b.Open, b.High, b.Low, b.Close)
	}
}

// TestAggregatorRespectsMaxHistory confirms old bars get dropped once
// history exceeds the configured cap.
func TestAggregatorRespectsMaxHistory(t *testing.T) {
	agg := NewAggregator(2, 3) // bar every 2 ticks, keep only 3 bars

	for i := range 20 { // 10 bars worth of ticks
		agg.AddTick(float64(i))
	}

	bars := agg.Bars()
	if len(bars) != 3 {
		t.Fatalf("expected history capped at 3 bars, got %d", len(bars))
	}
}

// TestSMAKnownValue checks SMA against a hand-computed reference.
func TestSMAKnownValue(t *testing.T) {
	bars := closeBars([]float64{10, 20, 30, 40, 50})
	sma, ok := SMA(bars, 5)
	if !ok {
		t.Fatalf("expected SMA to compute with exactly enough bars")
	}
	want := 30.0 // (10+20+30+40+50)/5
	if !almostEqual(sma, want, 1e-9) {
		t.Fatalf("SMA(5) = %.4f, want %.4f", sma, want)
	}
}

// TestSMAInsufficientBars confirms SMA correctly refuses to compute
// with fewer bars than the requested window rather than silently
// averaging over what's available.
func TestSMAInsufficientBars(t *testing.T) {
	bars := closeBars([]float64{10, 20})
	_, ok := SMA(bars, 5)
	if ok {
		t.Fatalf("expected SMA to report insufficient data for window 5 with only 2 bars")
	}
}

// TestRSIAllGainsIsMaximallyOverbought confirms RSI correctly reaches
// 100 when there are no losses at all in the window — a basic sanity
// check on the Wilder formula's boundary behavior.
func TestRSIAllGainsIsMaximallyOverbought(t *testing.T) {
	bars := closeBars([]float64{10, 11, 12, 13, 14, 15, 16}) // strictly increasing
	rsi, ok := RSI(bars, 5)
	if !ok {
		t.Fatalf("expected RSI to compute")
	}
	if rsi != 100 {
		t.Fatalf("expected RSI=100 for a strictly increasing series with no losses, got %.4f", rsi)
	}
}

// TestRSIAllLossesIsMaximallyOversold is the symmetric case.
func TestRSIAllLossesIsMaximallyOversold(t *testing.T) {
	bars := closeBars([]float64{16, 15, 14, 13, 12, 11, 10}) // strictly decreasing
	rsi, ok := RSI(bars, 5)
	if !ok {
		t.Fatalf("expected RSI to compute")
	}
	if rsi != 0 {
		t.Fatalf("expected RSI=0 for a strictly decreasing series with no gains, got %.4f", rsi)
	}
}

// TestRSIFlatIsNeutral confirms a perfectly flat price series produces
// the conventional neutral RSI of 50.
func TestRSIFlatIsNeutral(t *testing.T) {
	bars := closeBars([]float64{100, 100, 100, 100, 100, 100})
	rsi, ok := RSI(bars, 5)
	if !ok {
		t.Fatalf("expected RSI to compute")
	}
	if rsi != 50 {
		t.Fatalf("expected RSI=50 for a flat series, got %.4f", rsi)
	}
}

// TestEMAReactsFasterThanSMA confirms the defining property of an EMA
// versus an SMA: after a sudden price jump, EMA should have moved
// further toward the new price than SMA over the same window, since
// EMA weights recent observations more heavily.
func TestEMAReactsFasterThanSMA(t *testing.T) {
	// Flat at 100 for a while, then a jump to 200 for the last few bars.
	prices := []float64{100, 100, 100, 100, 100, 100, 100, 100, 200, 200, 200}
	bars := closeBars(prices)

	sma, ok1 := SMA(bars, 10)
	ema, ok2 := EMA(bars, 10)
	if !ok1 || !ok2 {
		t.Fatalf("expected both SMA and EMA to compute")
	}

	t.Logf("SMA(10)=%.4f EMA(10)=%.4f after a jump from 100 to 200", sma, ema)

	if ema <= sma {
		t.Fatalf("expected EMA to have moved further toward the recent jump than SMA, got SMA=%.4f EMA=%.4f", sma, ema)
	}
}

// TestBollingerBandsWidenWithVolatility confirms bands are wider for a
// more volatile series than a calm one, holding window and k constant.
func TestBollingerBandsWidenWithVolatility(t *testing.T) {
	calm := closeBars([]float64{100, 100.1, 99.9, 100.05, 99.95, 100.02, 99.98, 100.01, 99.99, 100})
	volatile := closeBars([]float64{100, 105, 95, 108, 92, 106, 94, 107, 93, 100})

	calmBands, ok1 := BollingerBands(calm, 10, 2)
	volatileBands, ok2 := BollingerBands(volatile, 10, 2)
	if !ok1 || !ok2 {
		t.Fatalf("expected both to compute")
	}

	calmWidth := calmBands.Upper - calmBands.Lower
	volatileWidth := volatileBands.Upper - volatileBands.Lower

	t.Logf("calm band width=%.4f, volatile band width=%.4f", calmWidth, volatileWidth)

	if volatileWidth <= calmWidth {
		t.Fatalf("expected volatile series to produce wider bands, got calm=%.4f volatile=%.4f", calmWidth, volatileWidth)
	}
}

// TestATRAccountsForGaps confirms ATR is sensitive to gaps between
// bars (a jump between one bar's close and the next bar's high/low),
// not just each bar's own high-low range — the specific property that
// distinguishes true range from a naive range calculation.
func TestATRAccountsForGaps(t *testing.T) {
	// Each individual bar has a small high-low range, but there's a
	// large gap between bar 1's close (100) and bar 2's open/high/low
	// (120) — a naive high-low-only range would miss this entirely.
	bars := []Bar{
		{Open: 99, High: 101, Low: 99, Close: 100},
		{Open: 120, High: 121, Low: 119, Close: 120},
		{Open: 120, High: 122, Low: 119, Close: 121},
	}

	atr, ok := ATR(bars, 2)
	if !ok {
		t.Fatalf("expected ATR to compute")
	}

	t.Logf("ATR with a gap between bars: %.4f", atr)

	// If ATR only considered each bar's own high-low range (roughly
	// 2-3 points each), it would be small; true range correctly
	// captures the ~20-point gap, so ATR should be large.
	if atr < 10 {
		t.Fatalf("expected ATR to reflect the large gap between bars via true range, got %.4f (too small — likely not accounting for gaps)", atr)
	}
}

// TestSwingHighLowFindsLocalExtremes confirms swing high/low reflect
// genuine local extremes within the window, not just first/last bar.
func TestSwingHighLowFindsLocalExtremes(t *testing.T) {
	bars := []Bar{
		{High: 100, Low: 95},
		{High: 110, Low: 90}, // the extremes are in the middle of the window
		{High: 105, Low: 98},
	}

	high, low, ok := SwingHighLow(bars, 3)
	if !ok {
		t.Fatalf("expected SwingHighLow to compute")
	}
	if high != 110 {
		t.Fatalf("expected swing high 110, got %.4f", high)
	}
	if low != 90 {
		t.Fatalf("expected swing low 90, got %.4f", low)
	}
}

// TestMomentumSignMatchesDirection confirms Momentum is positive for
// a rising series and negative for a falling one, with correct
// magnitude.
func TestMomentumSignMatchesDirection(t *testing.T) {
	rising := closeBars([]float64{100, 105, 110})
	falling := closeBars([]float64{100, 95, 90})

	upMom, ok1 := Momentum(rising, 2)
	downMom, ok2 := Momentum(falling, 2)
	if !ok1 || !ok2 {
		t.Fatalf("expected both to compute")
	}

	t.Logf("rising momentum=%.4f, falling momentum=%.4f", upMom, downMom)

	if upMom <= 0 {
		t.Fatalf("expected positive momentum for a rising series, got %.4f", upMom)
	}
	if downMom >= 0 {
		t.Fatalf("expected negative momentum for a falling series, got %.4f", downMom)
	}

	wantUp := (110.0 - 100.0) / 100.0
	if !almostEqual(upMom, wantUp, 1e-9) {
		t.Fatalf("rising momentum = %.4f, want %.4f", upMom, wantUp)
	}
}

// TestMACDSignalLineLagsMACDLine is a basic sanity check that MACD
// computes without error and produces a non-trivial histogram once a
// clear trend is present, rather than asserting an exact reference
// value (MACD's dependence on EMA seeding makes hand-computed exact
// values impractical to verify by hand; this checks its qualitative
// behavior instead).
func TestMACDProducesNonTrivialSignalOnTrend(t *testing.T) {
	prices := make([]float64, 60)
	for i := range prices {
		prices[i] = 100 + float64(i)*0.5 // steady uptrend
	}
	bars := closeBars(prices)

	result, ok := MACD(bars, 12, 26, 9)
	if !ok {
		t.Fatalf("expected MACD to compute with 60 bars (needs slow+signal=35 minimum)")
	}

	t.Logf("MACD=%.4f Signal=%.4f Histogram=%.4f", result.MACD, result.Signal, result.Histogram)

	// In a steady uptrend, the fast EMA should sit above the slow EMA,
	// so the MACD line itself should be positive.
	if result.MACD <= 0 {
		t.Fatalf("expected a positive MACD line during a steady uptrend, got %.4f", result.MACD)
	}
}
