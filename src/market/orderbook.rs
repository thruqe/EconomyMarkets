use serde::{Deserialize, Serialize};

/// Side represents the direction of an order.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Side {
    Buy,
    Sell,
}

impl std::fmt::Display for Side {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Side::Buy => write!(f, "BUY"),
            Side::Sell => write!(f, "SELL"),
        }
    }
}

/// Order is a resting or incoming order in the book.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: u64,
    pub agent_id: String,
    pub side: Side,
    pub price: f64, // ignored for market orders
    pub quantity: f64,
    pub is_market: bool,
}

/// Fill represents one match between an incoming order and a resting order.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Fill {
    pub taker_agent_id: String,
    pub maker_agent_id: String,
    pub price: f64,
    pub quantity: f64,
}

/// PriceLevel holds all resting quantity at a single price, FIFO by arrival.
#[derive(Debug, Clone)]
pub struct PriceLevel {
    pub price: f64,
    pub orders: Vec<Order>, // queue, front = oldest = filled first
}

/// OrderBook is a lightweight limit order book: discrete price levels,
/// each holding a FIFO queue of resting orders.
#[derive(Debug, Clone, Default)]
pub struct OrderBook {
    pub bids: Vec<PriceLevel>, // sorted descending by price (best bid first)
    pub asks: Vec<PriceLevel>, // sorted ascending by price (best ask first)
    next_order_id: u64,
}

/// BookLevel summarizes price, cumulative quantity, and order count at a single price tier.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BookLevel {
    pub price: f64,
    pub quantity: f64,
    pub orders_count: usize,
}

impl OrderBook {
    pub fn new() -> Self {
        Self::default()
    }

    /// Best bid returns the highest resting buy price, or None if none.
    pub fn best_bid(&self) -> Option<f64> {
        self.bids.first().map(|lvl| lvl.price)
    }

    /// Best ask returns the lowest resting sell price, or None if none.
    pub fn best_ask(&self) -> Option<f64> {
        self.asks.first().map(|lvl| lvl.price)
    }

    /// Spread returns ask - bid, or None if either side is empty.
    pub fn spread(&self) -> Option<f64> {
        match (self.best_bid(), self.best_ask()) {
            (Some(bid), Some(ask)) => Some(ask - bid),
            _ => None,
        }
    }

    /// MidPrice returns (bid + ask) / 2, or if only one side has liquidity, returns that side's best price.
    pub fn mid_price(&self) -> Option<f64> {
        match (self.best_bid(), self.best_ask()) {
            (Some(bid), Some(ask)) => Some((bid + ask) / 2.0),
            (Some(bid), None) => Some(bid),
            (None, Some(ask)) => Some(ask),
            (None, None) => None,
        }
    }

    /// AddLimitOrder places a resting limit order on the book without matching.
    pub fn add_limit_order(&mut self, mut o: Order) -> u64 {
        self.next_order_id += 1;
        o.id = self.next_order_id;
        let side = o.side;
        if side == Side::Buy {
            insert_level(&mut self.bids, o, true);
        } else {
            insert_level(&mut self.asks, o, false);
        }
        self.next_order_id
    }

    /// CancelAgentOrders removes all unexecuted resting orders belonging to agent_id from both sides.
    pub fn cancel_agent_orders(&mut self, agent_id: &str) {
        self.bids = filter_levels(std::mem::take(&mut self.bids), agent_id);
        self.asks = filter_levels(std::mem::take(&mut self.asks), agent_id);
    }

    /// Submit processes an incoming order (market or aggressive limit),
    /// walking the opposite side of the book and returning every fill generated.
    pub fn submit(&mut self, mut o: Order) -> Vec<Fill> {
        let mut fills = Vec::new();
        let mut remaining = o.quantity;

        let is_buy = o.side == Side::Buy;
        let opposite = if is_buy {
            &mut self.asks
        } else {
            &mut self.bids
        };

        while remaining > 0.0 && !opposite.is_empty() {
            let lvl_price = opposite[0].price;

            // Price limit check for non-market orders
            if !o.is_market {
                if is_buy && lvl_price > o.price {
                    break;
                }
                if !is_buy && lvl_price < o.price {
                    break;
                }
            }

            let lvl = &mut opposite[0];
            while !lvl.orders.is_empty() && remaining > 0.0 {
                let resting = &mut lvl.orders[0];
                let qty = remaining.min(resting.quantity);

                fills.push(Fill {
                    taker_agent_id: o.agent_id.clone(),
                    maker_agent_id: resting.agent_id.clone(),
                    price: lvl_price,
                    quantity: qty,
                });

                resting.quantity -= qty;
                remaining -= qty;

                if resting.quantity <= 1e-9 {
                    lvl.orders.remove(0);
                }
            }

            if opposite[0].orders.is_empty() {
                opposite.remove(0);
            }
        }

        // Resting remainder for limit orders only
        if remaining > 1e-9 && !o.is_market {
            o.quantity = remaining;
            self.add_limit_order(o);
        }

        fills
    }

    /// DepthAtLevels returns cumulative quantity available within the given
    /// number of price levels on each side.
    pub fn depth_at_levels(&self, n: usize) -> (f64, f64) {
        let bid_qty: f64 = self
            .bids
            .iter()
            .take(n)
            .flat_map(|lvl| &lvl.orders)
            .map(|o| o.quantity)
            .sum();

        let ask_qty: f64 = self
            .asks
            .iter()
            .take(n)
            .flat_map(|lvl| &lvl.orders)
            .map(|o| o.quantity)
            .sum();

        (bid_qty, ask_qty)
    }

    /// TopLevels returns the top n resting price levels for bids (highest first) and asks (lowest first).
    pub fn top_levels(&self, n: usize) -> (Vec<BookLevel>, Vec<BookLevel>) {
        let n = if n == 0 { 15 } else { n };

        let bids = self
            .bids
            .iter()
            .take(n)
            .map(|lvl| BookLevel {
                price: lvl.price,
                quantity: lvl.orders.iter().map(|o| o.quantity).sum(),
                orders_count: lvl.orders.len(),
            })
            .collect();

        let asks = self
            .asks
            .iter()
            .take(n)
            .map(|lvl| BookLevel {
                price: lvl.price,
                quantity: lvl.orders.iter().map(|o| o.quantity).sum(),
                orders_count: lvl.orders.len(),
            })
            .collect();

        (bids, asks)
    }
}

fn filter_levels(levels: Vec<PriceLevel>, agent_id: &str) -> Vec<PriceLevel> {
    let mut remaining = Vec::with_capacity(levels.len());
    for mut lvl in levels {
        lvl.orders.retain(|o| o.agent_id != agent_id);
        if !lvl.orders.is_empty() {
            remaining.push(lvl);
        }
    }
    remaining
}

fn insert_level(levels: &mut Vec<PriceLevel>, o: Order, descending: bool) {
    for lvl in levels.iter_mut() {
        if (lvl.price - o.price).abs() < 1e-9 {
            lvl.orders.push(o);
            return;
        }
    }

    levels.push(PriceLevel {
        price: o.price,
        orders: vec![o],
    });

    if descending {
        levels.sort_by(|a, b| b.price.partial_cmp(&a.price).unwrap_or(std::cmp::Ordering::Equal));
    } else {
        levels.sort_by(|a, b| a.price.partial_cmp(&b.price).unwrap_or(std::cmp::Ordering::Equal));
    }
}
