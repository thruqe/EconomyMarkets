use serde::{Deserialize, Serialize};

/// Bar is one OHLCV price bar aggregated from underlying ticks and trades.
#[derive(Debug, Clone, Copy, PartialEq, Serialize, Deserialize)]
pub struct Bar {
    pub open: f64,
    pub high: f64,
    pub low: f64,
    pub close: f64,
    pub volume: f64,
    pub retail_volume: f64,
    pub tick_count: usize,
}

impl Bar {
    pub fn new(price: f64, volume: f64, retail_volume: f64) -> Self {
        Self {
            open: price,
            high: price,
            low: price,
            close: price,
            volume,
            retail_volume,
            tick_count: 1,
        }
    }
}

/// Aggregator rolls a stream of raw tick-level prices and trade volumes into fixed-size
/// OHLCV bars, maintaining a rolling history of completed bars up to max_history.
#[derive(Debug, Clone)]
pub struct Aggregator {
    window_size: usize,
    max_history: usize,
    current: Option<Bar>,
    ticks_in_current: usize,
    completed: Vec<Bar>, // oldest first
}

impl Aggregator {
    pub fn new(window_size: usize, max_history: usize) -> Self {
        Self {
            window_size: window_size.max(1),
            max_history: max_history.max(1),
            current: None,
            ticks_in_current: 0,
            completed: Vec::new(),
        }
    }

    pub fn add_tick(&mut self, price: f64) {
        self.add_tick_with_volume(price, 0.0, 0.0);
    }

    pub fn add_tick_with_volume(&mut self, price: f64, volume: f64, retail_volume: f64) {
        match &mut self.current {
            None => {
                self.current = Some(Bar::new(price, volume, retail_volume));
                self.ticks_in_current = 1;
            }
            Some(curr) => {
                if price > curr.high {
                    curr.high = price;
                }
                if price < curr.low {
                    curr.low = price;
                }
                curr.close = price;
                curr.volume += volume;
                curr.retail_volume += retail_volume;
                curr.tick_count += 1;
                self.ticks_in_current += 1;
            }
        }

        if self.ticks_in_current >= self.window_size {
            if let Some(bar) = self.current.take() {
                self.completed.push(bar);
                if self.completed.len() > self.max_history {
                    let excess = self.completed.len() - self.max_history;
                    self.completed.drain(0..excess);
                }
            }
            self.ticks_in_current = 0;
        }
    }

    pub fn current(&self) -> Option<&Bar> {
        self.current.as_ref()
    }

    pub fn completed(&self) -> &[Bar] {
        &self.completed
    }

    pub fn all_bars(&self) -> Vec<Bar> {
        let mut bars = self.completed.clone();
        if let Some(ref curr) = self.current {
            bars.push(*curr);
        }
        bars
    }

    pub fn last_close(&self) -> Option<f64> {
        if let Some(ref curr) = self.current {
            Some(curr.close)
        } else {
            self.completed.last().map(|b| b.close)
        }
    }
}
