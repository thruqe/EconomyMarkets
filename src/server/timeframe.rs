use std::collections::HashMap;
use parking_lot::RwLock;
use crate::storage::models::CandleRecord;

pub const TF_1M: &str = "1m";
pub const TF_5M: &str = "5m";
pub const TF_15M: &str = "15m";
pub const TF_30M: &str = "30m";
pub const TF_45M: &str = "45m";
pub const TF_1H: &str = "1h";
pub const TF_2H: &str = "2h";
pub const TF_4H: &str = "4h";
pub const TF_1D: &str = "1D";
pub const TF_1W: &str = "1W";
pub const TF_1MO: &str = "1M";
pub const TF_1Y: &str = "1Y";

pub const ALL_TIMEFRAMES: &[&str] = &[
    TF_1M, TF_5M, TF_15M, TF_30M, TF_45M, TF_1H, TF_2H, TF_4H, TF_1D, TF_1W, TF_1MO, TF_1Y,
];

pub fn timeframe_duration_millis(tf: &str) -> i64 {
    match tf {
        TF_1M => 60 * 1000,
        TF_5M => 5 * 60 * 1000,
        TF_15M => 15 * 60 * 1000,
        TF_30M => 30 * 60 * 1000,
        TF_45M => 45 * 60 * 1000,
        TF_1H => 60 * 60 * 1000,
        TF_2H => 2 * 60 * 60 * 1000,
        TF_4H => 4 * 60 * 60 * 1000,
        TF_1D => 24 * 60 * 60 * 1000,
        TF_1W => 7 * 24 * 60 * 60 * 1000,
        TF_1MO => 30 * 24 * 60 * 60 * 1000,
        TF_1Y => 365 * 24 * 60 * 60 * 1000,
        _ => 60 * 1000,
    }
}

pub fn normalize_timeframe(tf: &str) -> String {
    let raw = tf.trim();
    if raw == "1m" || raw == "m" || raw.eq_ignore_ascii_case("1min") || raw.eq_ignore_ascii_case("min") || raw.eq_ignore_ascii_case("minute") {
        return TF_1M.to_string();
    }
    if raw == "1M" || raw == "M" || raw.eq_ignore_ascii_case("1mo") || raw.eq_ignore_ascii_case("mo") || raw.eq_ignore_ascii_case("month") || raw.eq_ignore_ascii_case("monthly") {
        return TF_1MO.to_string();
    }

    let lower = raw.to_lowercase();
    match lower.as_str() {
        "5m" | "5min" => TF_5M.to_string(),
        "15" | "15m" | "15min" => TF_15M.to_string(),
        "30" | "30m" | "30min" => TF_30M.to_string(),
        "45" | "45m" | "45min" => TF_45M.to_string(),
        "1h" | "60m" => TF_1H.to_string(),
        "2h" | "120m" => TF_2H.to_string(),
        "4h" | "240m" => TF_4H.to_string(),
        "d" | "1d" | "day" | "daily" => TF_1D.to_string(),
        "w" | "1w" | "week" | "weekly" => TF_1W.to_string(),
        "y" | "1y" | "year" | "yearly" => TF_1Y.to_string(),
        _ => {
            if raw == "D" {
                TF_1D.to_string()
            } else if raw == "W" {
                TF_1W.to_string()
            } else if raw == "Y" {
                TF_1Y.to_string()
            } else {
                TF_1M.to_string()
            }
        }
    }
}

pub struct SymbolTimeframeBars {
    pub symbol: String,
    completed_bars: RwLock<HashMap<String, Vec<CandleRecord>>>,
    current_bar: RwLock<HashMap<String, CandleRecord>>,
    current_bar_end: RwLock<HashMap<String, i64>>,
}

impl SymbolTimeframeBars {
    pub fn new(symbol: &str) -> Self {
        let mut completed = HashMap::new();
        for &tf in ALL_TIMEFRAMES {
            completed.insert(tf.to_string(), Vec::with_capacity(3000));
        }
        Self {
            symbol: symbol.to_string(),
            completed_bars: RwLock::new(completed),
            current_bar: RwLock::new(HashMap::new()),
            current_bar_end: RwLock::new(HashMap::new()),
        }
    }

    pub fn init_live_candles(&self, symbol: &str, start_price: f64, sim_start_time: i64) {
        let start_price = if start_price <= 0.0 { 100.0 } else { start_price };
        let mut completed = self.completed_bars.write();
        let mut current = self.current_bar.write();
        let mut ends = self.current_bar_end.write();

        let sym_seed = symbol.bytes().fold(14695981039346656037u64, |acc, b| {
            acc.wrapping_mul(1099511628211) ^ (b as u64)
        });

        for &tf in ALL_TIMEFRAMES {
            let dur = timeframe_duration_millis(tf);
            let num_historical = 60;
            let mut hist_bars = Vec::with_capacity(num_historical + 5);

            let mut xorshift = sym_seed.wrapping_add(tf.len() as u64 * 31);
            let mut next_f64 = || -> f64 {
                xorshift ^= xorshift << 13;
                xorshift ^= xorshift >> 7;
                xorshift ^= xorshift << 17;
                ((xorshift & 0xFFFFFFFF) as f64 + 1.0) / 4294967296.0
            };
            let mut next_normal = || -> f64 {
                let u1 = next_f64();
                let u2 = next_f64();
                (-2.0 * u1.ln()).sqrt() * (2.0 * std::f64::consts::PI * u2).cos()
            };

            // Determine market archetype for this symbol & timeframe
            let archetype = ((sym_seed ^ (tf.len() as u64 * 7919)) % 4) as usize;
            // Archetypes:
            // 0 = Strong Bullish Markup (steady impulse with shallow flags)
            // 1 = Cyclical Swing Channel (tests resistance, pulls back to support, rebounds)
            // 2 = Turnaround / Breakout Rally (consolidation then strong breakout wave)
            // 3 = Orderly Mean-Reversion / Consolidation

            let mut raw_opens = Vec::with_capacity(num_historical);
            let mut raw_closes = Vec::with_capacity(num_historical);
            let mut cur_p = 100.0;
            let mut prev_ret = 0.002;

            for i in 0..num_historical {
                let bar_open = cur_p;
                raw_opens.push(bar_open);

                // Define 5-phase market wave structure across the 60 bars
                let (phase_drift, phase_vol, persistence) = if i < 12 {
                    // Phase 1: Base / Accumulation
                    (0.0003, 0.0025, 0.65)
                } else if i < 28 {
                    // Phase 2: Primary Trend Impulse Wave
                    match archetype {
                        0 => (0.0048, 0.0035, 0.76), // Strong Bull Wave
                        1 => (0.0038, 0.0032, 0.72), // Rally to Range High
                        2 => (0.0010, 0.0028, 0.68), // Pre-breakout base
                        _ => (0.0028, 0.0030, 0.70), // Moderate trend
                    }
                } else if i < 38 {
                    // Phase 3: Corrective Retracement / Counter-Trend Flag
                    match archetype {
                        0 => (-0.0028, 0.0025, 0.74), // Orderly 38% pullback
                        1 => (-0.0042, 0.0030, 0.75), // Pullback to channel support
                        2 => (0.0055, 0.0040, 0.80),  // Sharp breakout expansion
                        _ => (-0.0022, 0.0024, 0.70), // Mild consolidation
                    }
                } else if i < 50 {
                    // Phase 4: Secondary Impulse / Breakout Wave
                    match archetype {
                        0 => (0.0042, 0.0036, 0.75), // Higher High breakout
                        1 => (0.0035, 0.0030, 0.72), // Rebound off support
                        2 => (0.0038, 0.0034, 0.75), // Breakout continuation
                        _ => (0.0020, 0.0028, 0.68), // Equilibrium
                    }
                } else {
                    // Phase 5: Convergence Wave into live start price
                    (0.0005, 0.0022, 0.65)
                };

                // Autoregressive AR(1) momentum ensures multi-bar directional persistence (no zigzag)
                let z = next_normal();
                let ret = (persistence * prev_ret) + ((1.0 - persistence) * phase_drift) + (phase_vol * z);
                let ret = ret.clamp(-0.022, 0.025);
                prev_ret = ret;

                cur_p = (cur_p * (1.0 + ret)).max(20.0);
                raw_closes.push(cur_p);
            }

            // Scale all bars so that the final historical bar's close EXACTLY equals start_price
            let last_close = *raw_closes.last().unwrap_or(&100.0);
            let scale_factor = if last_close > 0.0 { start_price / last_close } else { 1.0 };

            let base_vol = (start_price * 2500.0).clamp(20_000.0, 350_000.0);

            for i in 0..num_historical {
                let bar_start = sim_start_time - ((num_historical - i) as i64 * dur);
                let bar_end = bar_start + dur;

                let open = raw_opens[i] * scale_factor;
                let close = raw_closes[i] * scale_factor;
                let body = (close - open).abs();
                let is_up = close >= open;

                let (wick_high, wick_low, vol_mult) = if i >= 35 && i <= 37 && archetype <= 1 {
                    // Classic rejection hammer candle at the bottom of the pullback wave
                    let lower = body * (1.8 + next_f64() * 1.4) + open * 0.0008;
                    let upper = body * (0.08 + next_f64() * 0.18) + open * 0.0004;
                    (upper, lower, 1.75)
                } else if i >= 38 && i <= 40 {
                    // Breakout expansion candle with volume surge
                    let upper = body * (0.06 + next_f64() * 0.14) + open * 0.0004;
                    let lower = body * (0.06 + next_f64() * 0.14) + open * 0.0004;
                    (upper, lower, 2.4 + next_f64() * 0.8)
                } else if i >= 28 && i < 38 {
                    // Corrective pullback: lower volume dry-up
                    let upper = body * (0.12 + next_f64() * 0.30) + open * 0.0006;
                    let lower = body * (0.12 + next_f64() * 0.30) + open * 0.0006;
                    (upper, lower, 0.55 + next_f64() * 0.25)
                } else if is_up {
                    // Clean bullish trend candle: small wicks, solid body
                    let upper = body * (0.08 + next_f64() * 0.22) + open * 0.0005;
                    let lower = body * (0.04 + next_f64() * 0.14) + open * 0.0003;
                    (upper, lower, 1.2 + next_f64() * 0.6)
                } else {
                    // Clean bearish trend candle: small wicks, solid body
                    let upper = body * (0.04 + next_f64() * 0.14) + open * 0.0003;
                    let lower = body * (0.08 + next_f64() * 0.22) + open * 0.0005;
                    (upper, lower, 1.0 + next_f64() * 0.5)
                };

                let high = open.max(close) + wick_high;
                let low = (open.min(close) - wick_low).max(0.10);
                let vol = base_vol * vol_mult;

                hist_bars.push(CandleRecord {
                    symbol: symbol.to_string(),
                    timeframe: tf.to_string(),
                    start_time: bar_start,
                    end_time: bar_end,
                    open,
                    high,
                    low,
                    close,
                    volume: vol,
                    tick_count: 24,
                });
            }

            completed.insert(tf.to_string(), hist_bars);

            // Current live forming bar at sim_start_time
            current.insert(
                tf.to_string(),
                CandleRecord {
                    symbol: symbol.to_string(),
                    timeframe: tf.to_string(),
                    start_time: sim_start_time,
                    end_time: sim_start_time + dur,
                    open: start_price,
                    high: start_price,
                    low: start_price,
                    close: start_price,
                    volume: 0.0,
                    tick_count: 0,
                },
            );
            ends.insert(tf.to_string(), sim_start_time + dur);
        }
    }

    pub fn add_tick(&self, sim_time_millis: i64, price: f64, trade_volume: f64, _retail_volume: f64) -> Vec<CandleRecord> {
        if price <= 0.0 {
            return Vec::new();
        }

        let mut completed = self.completed_bars.write();
        let mut current = self.current_bar.write();
        let mut ends = self.current_bar_end.write();
        let mut closed_bars = Vec::new();

        for &tf in ALL_TIMEFRAMES {
            let dur = timeframe_duration_millis(tf);
            let bar_end = *ends.get(tf).unwrap_or(&(sim_time_millis + dur));

            if let Some(cur) = current.get_mut(tf) {
                if sim_time_millis >= bar_end {
                    let prev_close = cur.close;
                    cur.end_time = bar_end;
                    let closed = cur.clone();

                    let list = completed.entry(tf.to_string()).or_default();
                    list.push(closed.clone());
                    if list.len() > 3000 {
                        list.remove(0);
                    }
                    closed_bars.push(closed);

                    let mut next_start = bar_end;
                    let mut next_end = next_start + dur;
                    while sim_time_millis >= next_end {
                        next_start = next_end;
                        next_end = next_start + dur;
                    }

                    *cur = CandleRecord {
                        symbol: self.symbol.clone(),
                        timeframe: tf.to_string(),
                        start_time: next_start,
                        end_time: next_end,
                        open: prev_close,
                        high: prev_close.max(price),
                        low: prev_close.min(price),
                        close: price,
                        volume: trade_volume,
                        tick_count: 1,
                    };
                    ends.insert(tf.to_string(), next_end);
                } else {
                    if cur.tick_count == 0 {
                        cur.open = price;
                        cur.high = price;
                        cur.low = price;
                        cur.close = price;
                        cur.volume = trade_volume;
                        cur.tick_count = 1;
                    } else {
                        if price > cur.high {
                            cur.high = price;
                        }
                        if price < cur.low {
                            cur.low = price;
                        }
                        cur.close = price;
                        cur.volume += trade_volume;
                        cur.tick_count += 1;
                    }
                }
            } else {
                current.insert(
                    tf.to_string(),
                    CandleRecord {
                        symbol: self.symbol.clone(),
                        timeframe: tf.to_string(),
                        start_time: sim_time_millis,
                        end_time: sim_time_millis + dur,
                        open: price,
                        high: price,
                        low: price,
                        close: price,
                        volume: trade_volume,
                        tick_count: 1,
                    },
                );
                ends.insert(tf.to_string(), sim_time_millis + dur);
            }
        }

        closed_bars
    }

    pub fn get_bars(&self, tf: &str, limit: usize) -> Vec<CandleRecord> {
        let norm_tf = normalize_timeframe(tf);
        let completed = self.completed_bars.read();
        let current = self.current_bar.read();

        let list = completed.get(&norm_tf);
        let count = list.map(|l| l.len()).unwrap_or(0);
        let lim = if limit == 0 || limit > count { count } else { limit };

        let mut res = Vec::with_capacity(lim + 1);
        if let Some(l) = list {
            if !l.is_empty() {
                let start = l.len().saturating_sub(lim);
                res.extend_from_slice(&l[start..]);
            }
        }
        if let Some(cur) = current.get(&norm_tf) {
            res.push(cur.clone());
        }
        res
    }
}

pub struct MultiTimeframeManager {
    symbols: RwLock<HashMap<String, SymbolTimeframeBars>>,
}

impl MultiTimeframeManager {
    pub fn new() -> Self {
        Self {
            symbols: RwLock::new(HashMap::new()),
        }
    }

    pub fn register_symbol(&self, symbol: &str) {
        let mut symbols = self.symbols.write();
        if !symbols.contains_key(symbol) {
            symbols.insert(symbol.to_string(), SymbolTimeframeBars::new(symbol));
        }
    }

    pub fn with_symbol<F, R>(&self, symbol: &str, f: F) -> Option<R>
    where
        F: FnOnce(&SymbolTimeframeBars) -> R,
    {
        let symbols = self.symbols.read();
        symbols.get(symbol).map(f)
    }

    pub fn get_bars(&self, symbol: &str, tf: &str, limit: usize) -> Option<Vec<CandleRecord>> {
        let symbols = self.symbols.read();
        symbols.get(symbol).map(|stb| stb.get_bars(tf, limit))
    }
}
