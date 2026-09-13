package market

import (
	"fmt"
	"testing"
)

// TestSlippageFromDepth demonstrates that a large market order walks
// through multiple price levels and gets a worse average fill price
// than the best ask — with no separate "impact formula", just the
// mechanical consequence of consuming resting liquidity.
func TestSlippageFromDepth(t *testing.T) {
	book := NewOrderBook()

	// Market maker seeds resting liquidity on the ask side.
	book.AddLimitOrder(&Order{AgentID: "mm", Side: Sell, Price: 100.10, Quantity: 500})
	book.AddLimitOrder(&Order{AgentID: "mm", Side: Sell, Price: 100.15, Quantity: 500})
	book.AddLimitOrder(&Order{AgentID: "mm", Side: Sell, Price: 100.20, Quantity: 500})

	// A hedge fund submits one large market buy.
	fills := book.Submit(&Order{AgentID: "hedge_fund_1", Side: Buy, Quantity: 1000, IsMarket: true})

	var totalCost, totalQty float64
	for _, f := range fills {
		fmt.Printf("fill: %.0f @ %.2f\n", f.Quantity, f.Price)
		totalCost += f.Price * f.Quantity
		totalQty += f.Quantity
	}
	avgPrice := totalCost / totalQty
	fmt.Printf("average fill price: %.4f\n", avgPrice)

	if avgPrice <= 100.10 {
		t.Fatalf("expected slippage above best ask 100.10, got avg %.4f", avgPrice)
	}

	newAsk, _ := book.BestAsk()
	fmt.Printf("new best ask after trade: %.2f\n", newAsk)
	if newAsk <= 100.10 {
		t.Fatalf("expected best ask to have moved up after depth was consumed")
	}
}

// TestMarginCascade demonstrates a forced liquidation: an over-leveraged
// long position gets marked down as price falls, breaches maintenance
// margin, and the LiquidationEngine produces a forced sell order. That
// sell then gets submitted to the same book, which can push price down
// further and could (in a fuller sim with more accounts) trigger the
// next account's liquidation in turn.
func TestMarginCascade(t *testing.T) {
	book := NewOrderBook()

	// Resting bids for the forced sell to hit.
	book.AddLimitOrder(&Order{AgentID: "mm", Side: Buy, Price: 99.50, Quantity: 200})
	book.AddLimitOrder(&Order{AgentID: "mm", Side: Buy, Price: 99.00, Quantity: 300})
	book.AddLimitOrder(&Order{AgentID: "mm", Side: Buy, Price: 98.50, Quantity: 500})

	// A leveraged trader: 5x max leverage, called at 10% maintenance margin.
	acct := NewAccount("leveraged_trader", 1000, 5.0, 0.10)

	// They open a long of 100 shares at 100.00 (position value 10,000 —
	// that's 10x their cash, beyond their stated max, but for this demo
	// we're just placing the position directly to test the margin check
	// itself rather than the entry-sizing logic).
	acct.Positions["SYN"] = &Position{Side: Long, Quantity: 100, EntryCost: 10000}

	prices := map[string]float64{"SYN": 100.00}
	ratio, _ := acct.MarginRatio(prices)
	fmt.Printf("margin ratio at entry price: %.3f\n", ratio)

	// Price drops sharply.
	prices["SYN"] = 91.00
	ratio, _ = acct.MarginRatio(prices)
	fmt.Printf("margin ratio after drop to 91.00: %.3f\n", ratio)

	engine := LiquidationEngine{}
	forced := engine.ScanForLiquidations([]*Account{acct}, prices)

	if len(forced) == 0 {
		t.Fatalf("expected a forced liquidation order once under maintenance margin")
	}

	fo := forced[0]
	fmt.Printf("forced order: side=%v qty=%.0f (account %s)\n", fo.Order.Side, fo.Order.Quantity, fo.Account.AgentID)

	// Submit the forced sell — this is the step that would push price
	// down further and could trip the next leveraged account in a
	// fuller simulation.
	fills := book.Submit(&fo.Order)
	for _, f := range fills {
		fmt.Printf("liquidation fill: %.0f @ %.2f\n", f.Quantity, f.Price)
	}

	newBid, _ := book.BestBid()
	fmt.Printf("best bid after forced liquidation: %.2f\n", newBid)
}
