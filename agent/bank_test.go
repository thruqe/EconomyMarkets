package agent

import (
	"testing"

	"economy/company"
	"economy/market"
)

// TestBankIgnoresCompaniesOutsideCoverage confirms a Bank produces no
// orders when handed a MarketState for a symbol it doesn't cover.
func TestBankIgnoresCompaniesOutsideCoverage(t *testing.T) {
	covered := makeTestCompany("COVERED", 150.0)
	acct := market.NewAccount("bank1", 10_000_000, 3.0, 0.15)
	bank := NewBank("bank1", acct, []*company.Company{covered})

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 99.90, Quantity: 1000})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 100.10, Quantity: 1000})

	state := market.NewMarketState("OTHER", 1, book, 5, nil, 0)

	orders := bank.NextOrders(state)
	if len(orders) != 0 {
		t.Fatalf("expected no orders for an uncovered symbol, got %d", len(orders))
	}
}

// TestBankMoreConservativeThanHedgeFund confirms Bank's higher
// MinTradeThreshold means it stays out of a mispricing an ordinary
// HedgeFund would already act on.
func TestBankMoreConservativeThanHedgeFund(t *testing.T) {
	// A company reported meaningfully above market price: after each
	// agent's own SkepticismDiscount haircut, the perceived mispricing
	// lands around 6.5% for HedgeFund (35% discount — clears its 3%
	// threshold and rebalance floor) but only around 5.0% for Bank
	// (50% discount — stays under its 6% threshold). This is what the
	// discount mechanism is meant to produce: the same report read
	// two different ways by two different levels of skepticism.
	comp := makeTestCompany("SYN", 110.0)

	bankAcct := market.NewAccount("bank1", 10_000_000, 3.0, 0.15)
	bank := NewBank("bank1", bankAcct, []*company.Company{comp})

	fundAcct := market.NewAccount("fund1", 10_000_000, 5.0, 0.10)
	fund := NewHedgeFund("fund1", fundAcct, comp)

	book := market.NewOrderBook()
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: 99.90, Quantity: 100000})
	book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: 100.10, Quantity: 100000})

	state := market.NewMarketState("SYN", 1, book, 5, nil, 0)

	bankOrders := bank.NextOrders(state)
	fundOrders := fund.NextOrders(state)

	if len(bankOrders) != 0 {
		t.Fatalf("expected Bank to stay out of a ~4%% mispricing (below its 6%% threshold), got %d orders", len(bankOrders))
	}
	if len(fundOrders) != 1 {
		t.Fatalf("expected HedgeFund to act on the same ~4%% mispricing (above its 3%% threshold), got %d orders", len(fundOrders))
	}
}

// TestBankDeRisksAfterDrawdown is the important one: it proves that
// once a Bank's account has suffered a large enough drawdown from its
// equity high-water mark, it scales down (or entirely suppresses) new
// conviction-sized exposure to a *different* company with a perfectly
// good, large mispricing — i.e. the de-risking is genuinely account-
// wide and thesis-independent, not something that only responds to
// losses in the same position being sized.
func TestBankDeRisksAfterDrawdown(t *testing.T) {
	compA := makeTestCompany("AAA", 100.0) // will be the source of the loss
	compB := makeTestCompany("BBB", 150.0) // separate company, strong standalone thesis

	acct := market.NewAccount("bank1", 1_000_000, 3.0, 0.15)
	bank := NewBank("bank1", acct, []*company.Company{compA, compB})

	// Establish the high-water mark first: give the account an initial
	// mark at full health with no positions.
	bank.riskScale(bank.account.Equity(map[string]float64{"AAA": 100.0, "BBB": 150.0}))

	// Simulate a large loss: open a big long in AAA, then mark AAA's
	// reported value down hard. Note this test must move
	// compA.ReportedValue itself (not just a local price map) because
	// Bank.NextOrders's internal equity recomputation
	// (currentPricesAcrossCoverage) marks every non-current-tick
	// symbol using the company's own ReportedValue — a known
	// approximation documented on that method (SetExternalPrices lets
	// an orchestrator override this with real market prices, unset
	// here on purpose to exercise the fallback path directly), since a
	// single tick's MarketState only carries a real market price for
	// one symbol. Moving ReportedValue directly is what actually
	// exercises that path honestly, rather than working around the
	// approximation.
	market.ApplySettledFill(acct, "AAA", market.Buy, 5000, 100.0) // 500,000 notional long
	compA.ReportedValue = 60.0                                    // -40%, will be picked up internally too

	// Mark AAA down 40% - equity now reflects a large unrealized loss.
	lossPrices := map[string]float64{"AAA": 60.0, "BBB": 150.0}
	equityAfterLoss := acct.Equity(lossPrices)
	t.Logf("equity after loss: %.2f (high-water mark: %.2f)", equityAfterLoss, bank.equityHighWaterMark)

	scale := bank.riskScale(equityAfterLoss)
	t.Logf("risk scale after drawdown: %.4f", scale)

	if scale >= 1.0 {
		t.Fatalf("expected risk scale to have shrunk below 1.0 after a large drawdown, got %.4f", scale)
	}

	// Now check that this suppressed scale actually reduces (or fully
	// suppresses) the order Bank would otherwise place on BBB's large,
	// perfectly good mispricing.
	book := market.NewOrderBook()
	for i := range 40 {
		bidPrice := 99.90 - float64(i)*0.10
		askPrice := 100.10 + float64(i)*0.10
		book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Buy, Price: bidPrice, Quantity: 20000})
		book.AddLimitOrder(&market.Order{AgentID: "seed", Side: market.Sell, Price: askPrice, Quantity: 20000})
	}
	stateB := market.NewMarketState("BBB", 2, book, 5, nil, 0)

	ordersAfterDrawdown := bank.NextOrders(stateB)

	// Compare against what an otherwise-identical Bank (no drawdown)
	// would have ordered on the same BBB mispricing, to confirm the
	// drawdown genuinely suppressed sizing rather than this being a
	// coincidental zero.
	freshAcct := market.NewAccount("bank2", 1_000_000, 3.0, 0.15)
	freshBank := NewBank("bank2", freshAcct, []*company.Company{compB})
	freshBank.riskScale(freshAcct.Equity(map[string]float64{"BBB": 150.0})) // establish HWM at full health
	ordersFresh := freshBank.NextOrders(stateB)

	if len(ordersFresh) != 1 {
		t.Fatalf("expected the undrawn-down bank to place an order on BBB's large mispricing, got %d", len(ordersFresh))
	}

	var drawdownQty float64
	if len(ordersAfterDrawdown) == 1 {
		drawdownQty = ordersAfterDrawdown[0].Quantity
	}
	t.Logf("order qty: fresh bank=%.2f, drawn-down bank=%.2f", ordersFresh[0].Quantity, drawdownQty)

	if drawdownQty >= ordersFresh[0].Quantity {
		t.Fatalf("expected drawn-down bank to size BBB's position smaller (or zero) than a fresh bank, got fresh=%.2f drawdown=%.2f",
			ordersFresh[0].Quantity, drawdownQty)
	}
}

// TestBankFullyStopsNewRiskBeyondCutoff confirms that beyond
// DrawdownFullCutoff, risk scale reaches exactly zero.
func TestBankFullyStopsNewRiskBeyondCutoff(t *testing.T) {
	acct := market.NewAccount("bank1", 1_000_000, 3.0, 0.15)
	bank := NewBank("bank1", acct, nil)

	bank.riskScale(1_000_000) // establish high-water mark

	scale := bank.riskScale(650_000) // 35% drawdown, beyond the 30% cutoff
	t.Logf("risk scale at 35%% drawdown (cutoff is 30%%): %.4f", scale)

	if scale != 0 {
		t.Fatalf("expected risk scale to be exactly 0 beyond DrawdownFullCutoff, got %.4f", scale)
	}
}
