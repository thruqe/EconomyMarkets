use economy::technicals::{
    atr, bollinger_bands, ema, macd, momentum, rsi, sma, Aggregator, Bar,
};

fn make_bars(closes: &[f64]) -> Vec<Bar> {
    closes
        .iter()
        .map(|&p| Bar {
            open: p,
            high: p + 1.0,
            low: p - 1.0,
            close: p,
            volume: 100.0,
            retail_volume: 50.0,
            tick_count: 1,
        })
        .collect()
}

#[test]
fn test_aggregator_and_indicators() {
    let mut agg = Aggregator::new(5, 10);
    let ticks = [100.0, 105.0, 98.0, 102.0, 101.0];
    for t in ticks {
        agg.add_tick(t);
    }
    let bars = agg.all_bars();
    assert_eq!(bars.len(), 1);
    assert_eq!(bars[0].open, 100.0);
    assert_eq!(bars[0].high, 105.0);
    assert_eq!(bars[0].low, 98.0);
    assert_eq!(bars[0].close, 101.0);

    // Indicators
    let bars_series = make_bars(&[10.0, 20.0, 30.0, 40.0, 50.0]);
    let sma_val = sma(&bars_series, 5).unwrap();
    assert_eq!(sma_val, 30.0);

    let ema_val = ema(&bars_series, 5).unwrap();
    assert!(ema_val > 0.0);

    let mom = momentum(&bars_series, 4).unwrap();
    assert_eq!(mom, 4.0);

    let bb = bollinger_bands(&bars_series, 5, 2.0).unwrap();
    assert_eq!(bb.middle, 30.0);
    assert!(bb.upper > bb.middle);
    assert!(bb.lower < bb.middle);

    let atr_val = atr(&bars_series, 4).unwrap();
    assert!(atr_val > 0.0);

    let long_bars = make_bars(&[
        44.0, 44.5, 45.0, 45.5, 46.0, 46.5, 47.0, 47.5, 48.0, 48.5, 49.0, 49.5, 50.0, 50.5, 51.0, 51.5,
    ]);
    let rsi_val = rsi(&long_bars, 14).unwrap();
    assert!(rsi_val > 70.0, "Uptrending series should produce high RSI");

    let macd_bars: Vec<Bar> = (0..40).map(|i| Bar {
        open: i as f64,
        high: i as f64 + 1.0,
        low: i as f64 - 1.0,
        close: i as f64,
        volume: 100.0,
        retail_volume: 50.0,
        tick_count: 1,
    }).collect();
    let macd_res = macd(&macd_bars, 12, 26, 9).unwrap();
    assert!(macd_res.macd > 0.0);
}
