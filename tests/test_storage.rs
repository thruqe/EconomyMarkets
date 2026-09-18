use std::time::Duration;
use economy::storage::db::DB;
use economy::storage::models::*;
use economy::storage::writer::AsyncWriter;

#[tokio::test]
async fn test_storage_lifecycle() {
    let db = DB::open_in_memory().expect("failed to open memory db");
    let (writer, handle) = AsyncWriter::new(db.clone(), 100, Duration::from_millis(50));

    let now = chrono::Utc::now().timestamp_millis();
    let batch = Batch {
        trades: vec![
            TradeRecord {
                id: 0,
                tick: 1,
                timestamp: now,
                symbol: "TEST".into(),
                taker_id: "buyer1".into(),
                maker_id: "mm1".into(),
                side: "BUY".into(),
                price: 100.50,
                quantity: 200.0,
            },
            TradeRecord {
                id: 0,
                tick: 1,
                timestamp: now,
                symbol: "TEST".into(),
                taker_id: "seller1".into(),
                maker_id: "mm1".into(),
                side: "SELL".into(),
                price: 100.40,
                quantity: 150.0,
            },
        ],
        prices: vec![PriceRecord {
            tick: 1,
            timestamp: now,
            symbol: "TEST".into(),
            mid: 100.45,
            bid: 100.40,
            ask: 100.50,
            spread: 0.10,
        }],
        candles: vec![CandleRecord {
            symbol: "TEST".into(),
            timeframe: "1s".into(),
            start_time: now - 1000,
            end_time: now,
            open: 100.0,
            high: 101.0,
            low: 99.5,
            close: 100.45,
            volume: 350.0,
            tick_count: 10,
        }],
        events: vec![EventRecord {
            id: 0,
            tick: 1,
            timestamp: now,
            kind: "Fundamental".into(),
            symbol: "TEST".into(),
            details: r#"{"multiplier":1.05}"#.into(),
        }],
        accounts: vec![AccountRecord {
            tick: 1,
            timestamp: now,
            agent_id: "buyer1".into(),
            cash: 50000.0,
            equity: 70000.0,
            margin_ratio: 0.25,
        }],
    };

    assert!(writer.enqueue(batch));
    writer.flush().await.expect("flush failed");

    let trades = db.get_recent_trades("TEST", 10).unwrap();
    assert_eq!(trades.len(), 2);
    assert_eq!(trades[0].quantity, 150.0); // newest first (descending id)

    let candles = db.get_candles("TEST", "1s", 10).unwrap();
    assert_eq!(candles.len(), 1);
    assert_eq!(candles[0].open, 100.0);

    let prices = db.get_latest_prices().unwrap();
    assert_eq!(prices.get("TEST").unwrap().mid, 100.45);

    let events = db.get_recent_events(10).unwrap();
    assert_eq!(events.len(), 1);
    assert_eq!(events[0].kind, "Fundamental");

    let accounts = db.get_account_snapshots("buyer1", 10).unwrap();
    assert_eq!(accounts.len(), 1);
    assert_eq!(accounts[0].cash, 50000.0);

    drop(writer);
    let _ = handle.await;
}
