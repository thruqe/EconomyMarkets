use std::collections::HashMap;
use economy::market::margin::{Account, LiquidationEngine, Position, PositionSide};
use economy::market::orderbook::{Order, OrderBook, Side};

#[test]
fn test_slippage_from_depth() {
    let mut book = OrderBook::new();

    // Market maker seeds resting liquidity on the ask side.
    book.submit(Order {
        id: 0,
        agent_id: "mm".into(),
        side: Side::Sell,
        price: 100.10,
        quantity: 500.0,
        is_market: false,
    });
    book.submit(Order {
        id: 0,
        agent_id: "mm".into(),
        side: Side::Sell,
        price: 100.15,
        quantity: 500.0,
        is_market: false,
    });
    book.submit(Order {
        id: 0,
        agent_id: "mm".into(),
        side: Side::Sell,
        price: 100.20,
        quantity: 500.0,
        is_market: false,
    });

    // Hedge fund submits one large market buy of 1000 shares.
    let fills = book.submit(Order {
        id: 0,
        agent_id: "hedge_fund_1".into(),
        side: Side::Buy,
        price: 0.0,
        quantity: 1000.0,
        is_market: true,
    });

    assert_eq!(fills.len(), 2);
    let mut total_cost = 0.0;
    let mut total_qty = 0.0;
    for f in &fills {
        total_cost += f.price * f.quantity;
        total_qty += f.quantity;
    }
    let avg_price = total_cost / total_qty;
    assert!(
        avg_price > 100.10,
        "expected slippage above best ask 100.10, got avg {}",
        avg_price
    );
    assert_eq!(avg_price, 100.125);

    let new_ask = book.best_ask().unwrap();
    assert!(new_ask > 100.10);
    assert_eq!(new_ask, 100.20);
}

#[test]
fn test_margin_cascade() {
    let mut book = OrderBook::new();

    // Resting bids for the forced sell to hit.
    book.submit(Order {
        id: 0,
        agent_id: "mm".into(),
        side: Side::Buy,
        price: 99.50,
        quantity: 200.0,
        is_market: false,
    });
    book.submit(Order {
        id: 0,
        agent_id: "mm".into(),
        side: Side::Buy,
        price: 99.00,
        quantity: 300.0,
        is_market: false,
    });
    book.submit(Order {
        id: 0,
        agent_id: "mm".into(),
        side: Side::Buy,
        price: 98.50,
        quantity: 500.0,
        is_market: false,
    });

    // Leveraged trader: 5x max leverage, called at 10% maintenance margin.
    let mut acct = Account::new("leveraged_trader", 1000.0, 5.0, 0.10);
    acct.positions.insert(
        "SYN".into(),
        Position {
            side: PositionSide::Long,
            quantity: 100.0,
            entry_cost: 10000.0,
            opened_at: 0,
        },
    );

    let mut prices = HashMap::new();
    prices.insert("SYN".into(), 100.0);
    let ratio = acct.margin_ratio(&prices).unwrap();
    assert!((ratio - 0.10).abs() < 1e-4);

    // Price drops sharply to 91.00
    prices.insert("SYN".into(), 91.0);
    let ratio_drop = acct.margin_ratio(&prices).unwrap();
    assert!(ratio_drop < 0.10);

    let forced = LiquidationEngine::scan_for_liquidations(&[&acct], &prices);
    assert_eq!(forced.len(), 1);

    let fo = &forced[0];
    assert_eq!(fo.order.side, Side::Sell);
    assert_eq!(fo.order.quantity, 100.0);

    let fills = book.submit(fo.order.clone());
    assert!(!fills.is_empty());
    assert_eq!(fills[0].price, 99.50);
}
