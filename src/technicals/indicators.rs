use super::bar::Bar;
use serde::{Deserialize, Serialize};

/// Extracts closing prices from bars.
pub fn closes(bars: &[Bar]) -> Vec<f64> {
    bars.iter().map(|b| b.close).collect()
}

/// SMA computes the simple moving average of closing price over the most recent window bars.
pub fn sma(bars: &[Bar], window: usize) -> Option<f64> {
    if window < 1 || bars.len() < window {
        return None;
    }
    let recent = &bars[bars.len() - window..];
    let sum: f64 = recent.iter().map(|b| b.close).sum();
    Some(sum / window as f64)
}

/// EMA computes the exponential moving average of closing price over window bars.
pub fn ema(bars: &[Bar], window: usize) -> Option<f64> {
    if window < 1 || bars.len() < window {
        return None;
    }
    let alpha = 2.0 / (window as f64 + 1.0);
    let seed = sma(&bars[..window], window)?;

    let mut current_ema = seed;
    for b in &bars[window..] {
        current_ema = alpha * b.close + (1.0 - alpha) * current_ema;
    }
    Some(current_ema)
}

/// RSI computes the Relative Strength Index over window bars using Wilder's original method.
pub fn rsi(bars: &[Bar], window: usize) -> Option<f64> {
    if window < 1 || bars.len() < window + 1 {
        return None;
    }

    let c = closes(bars);

    let mut sum_gain = 0.0;
    let mut sum_loss = 0.0;
    for i in 1..=window {
        let change = c[i] - c[i - 1];
        if change > 0.0 {
            sum_gain += change;
        } else {
            sum_loss += -change;
        }
    }
    let mut avg_gain = sum_gain / window as f64;
    let mut avg_loss = sum_loss / window as f64;

    for i in (window + 1)..c.len() {
        let change = c[i] - c[i - 1];
        let (gain, loss) = if change > 0.0 {
            (change, 0.0)
        } else {
            (0.0, -change)
        };
        avg_gain = (avg_gain * (window as f64 - 1.0) + gain) / window as f64;
        avg_loss = (avg_loss * (window as f64 - 1.0) + loss) / window as f64;
    }

    if avg_loss == 0.0 {
        if avg_gain == 0.0 {
            return Some(50.0); // Neutral
        }
        return Some(100.0); // Maximally overbought
    }

    let rs = avg_gain / avg_loss;
    Some(100.0 - (100.0 / (1.0 + rs)))
}

/// MACDResult holds the three standard MACD output values.
#[derive(Debug, Clone, Copy, PartialEq, Serialize, Deserialize)]
pub struct MacdResult {
    pub macd: f64,
    pub signal: f64,
    pub histogram: f64,
}

/// MACD computes Moving Average Convergence/Divergence.
pub fn macd(bars: &[Bar], fast_window: usize, slow_window: usize, signal_window: usize) -> Option<MacdResult> {
    if bars.len() < slow_window + signal_window {
        return None;
    }

    let mut macd_series = Vec::with_capacity(bars.len() - slow_window + 1);
    for i in slow_window..=bars.len() {
        let window = &bars[..i];
        let fast = ema(window, fast_window)?;
        let slow = ema(window, slow_window)?;
        macd_series.push(Bar {
            open: fast - slow,
            high: fast - slow,
            low: fast - slow,
            close: fast - slow,
            volume: 0.0,
            retail_volume: 0.0,
            tick_count: 1,
        });
    }

    if macd_series.len() < signal_window {
        return None;
    }

    let signal = ema(&macd_series, signal_window)?;
    let macd_now = macd_series.last()?.close;

    Some(MacdResult {
        macd: macd_now,
        signal,
        histogram: macd_now - signal,
    })
}

/// BollingerResult holds the three bands.
#[derive(Debug, Clone, Copy, PartialEq, Serialize, Deserialize)]
pub struct BollingerResult {
    pub middle: f64,
    pub upper: f64,
    pub lower: f64,
}

/// BollingerBands computes Bollinger Bands over window with k standard deviations.
pub fn bollinger_bands(bars: &[Bar], window: usize, k: f64) -> Option<BollingerResult> {
    let middle = sma(bars, window)?;
    let recent = &bars[bars.len() - window..];
    let sum_sq: f64 = recent.iter().map(|b| {
        let diff = b.close - middle;
        diff * diff
    }).sum();
    let stddev = (sum_sq / window as f64).sqrt();

    Some(BollingerResult {
        middle,
        upper: middle + k * stddev,
        lower: middle - k * stddev,
    })
}

fn true_range(current: Bar, previous: Bar) -> f64 {
    let high_low = current.high - current.low;
    let high_prev_close = (current.high - previous.close).abs();
    let low_prev_close = (current.low - previous.close).abs();
    high_low.max(high_prev_close.max(low_prev_close))
}

/// ATR computes the Average True Range over window bars.
pub fn atr(bars: &[Bar], window: usize) -> Option<f64> {
    if window < 1 || bars.len() < window + 1 {
        return None;
    }

    let recent = &bars[bars.len() - window - 1..];
    let mut sum = 0.0;
    for i in 1..recent.len() {
        sum += true_range(recent[i], recent[i - 1]);
    }
    Some(sum / window as f64)
}

/// SwingHighLow finds the local extremum swing high and swing low within recent window bars.
pub fn swing_high_low(bars: &[Bar], window: usize) -> Option<(f64, f64)> {
    if window < 1 || bars.len() < window {
        return None;
    }
    let recent = &bars[bars.len() - window..];
    let mut swing_high = recent[0].high;
    let mut swing_low = recent[0].low;
    for b in &recent[1..] {
        if b.high > swing_high {
            swing_high = b.high;
        }
        if b.low < swing_low {
            swing_low = b.low;
        }
    }
    Some((swing_high, swing_low))
}

/// Momentum computes simple percentage change in closing price: (latest - past) / past.
pub fn momentum(bars: &[Bar], window: usize) -> Option<f64> {
    if window < 1 || bars.len() <= window {
        return None;
    }
    let past = bars[bars.len() - window - 1].close;
    let latest = bars[bars.len() - 1].close;
    if past == 0.0 {
        return None;
    }
    Some((latest - past) / past)
}
