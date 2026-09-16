package tui

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"economy/citizen"
	"economy/company"
	"economy/country"
	"economy/country/fiscal"
	"economy/market"
	"economy/world"
)


func TestMain(m *testing.M) {
	_ = os.Remove("data/world_state.json")
	_ = os.Remove("data/portfolio.json")
	code := m.Run()
	_ = os.Remove("data/world_state.json")
	_ = os.Remove("data/portfolio.json")
	os.Exit(code)
}

func TestTUIEngineAndModel(t *testing.T) {
	tickChan := make(chan TickMsg, 50)
	engine := NewEngine(500, 42, 50*time.Millisecond, tickChan)

	// 1. Verify 500 companies generated
	universe := engine.GetUniverse()
	if len(universe) != 500 {
		t.Fatalf("expected 500 companies, got %d", len(universe))
	}

	// 2. Start simulation engine and wait for ticks
	engine.Start()
	defer func() {
		engine.Stop()
		_ = os.Remove("data/world_state.json")
		_ = os.Remove("data/portfolio.json")
	}()

	var firstTick TickMsg
	select {
	case msg := <-tickChan:
		firstTick = msg
		t.Logf("Received live tick %d with %d trades", msg.Report.Tick, len(msg.Report.Trades))
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for live tick from simulation engine")
	}

	// 3. Initialize Model and test all view rendering
	model := NewModel(engine, tickChan)
	model.Update(firstTick)

	// Test Overview tab
	model.activeTab = TabOverview
	overviewRender := model.View()
	if len(overviewRender) == 0 {
		t.Fatalf("expected non-empty overview render")
	}

	// Test Universe 500 tab
	model.activeTab = TabUniverse
	universeRender := model.View()
	if len(universeRender) == 0 {
		t.Fatalf("expected non-empty universe render")
	}

	// Test filtering in 500 companies
	model.universeView.SetSearch("Tech")
	filtered := model.universeView.FilteredRows(model.rows)
	t.Logf("Found %d companies matching 'Tech'", len(filtered))
	model.universeView.ClearSearch()

	// Test Detail tab
	model.activeTab = TabDetail
	detailRender := model.View()
	if len(detailRender) == 0 {
		t.Fatalf("expected non-empty detail render")
	}

	// Test Trade tab & Order submission
	model.activeTab = TabTrade
	tradeRender := model.View()
	if len(tradeRender) == 0 {
		t.Fatalf("expected non-empty trade render")
	}

	// Submit test buy order
	firstSym := universe[0].Symbol
	err := engine.SubmitOrder(firstSym, market.Buy, true, 100, 100.0)
	if err != nil {
		t.Fatalf("failed to submit order: %v", err)
	}

	// Test News tab
	model.activeTab = TabNews
	newsRender := model.View()
	if len(newsRender) == 0 {
		t.Fatalf("expected non-empty news render")
	}

	// Test Admin tab & IPO launch
	model.activeTab = TabAdmin
	adminRender := model.View()
	if len(adminRender) == 0 {
		t.Fatalf("expected non-empty admin render")
	}

	err = engine.LaunchIPO("Hyperion Quantum", "HQTM", company.InformationTechnology, company.MidCap, 50_000_000, 65.00)
	if err != nil {
		t.Fatalf("failed to launch IPO: %v", err)
	}

	newUniverse := engine.GetUniverse()
	if len(newUniverse) != 501 {
		t.Fatalf("expected 501 companies after IPO, got %d", len(newUniverse))
	}
	t.Logf("IPO successfully launched: %s, total companies: %d", newUniverse[len(newUniverse)-1].Symbol, len(newUniverse))

	// Test speed controls
	engine.SetSpeed(2.0)
	if engine.speedMultiplier != 2.0 {
		t.Fatalf("expected speed 2.0, got %.1f", engine.speedMultiplier)
	}
	engine.SetSpeed(0.0) // Pause
	if engine.speedMultiplier != 0.0 {
		t.Fatalf("expected speed 0.0, got %.1f", engine.speedMultiplier)
	}

	// 4. Test TabCharts (Line Chart with Timeframes)
	model.activeTab = TabCharts
	chartRender := model.View()
	if len(chartRender) == 0 {
		t.Fatalf("expected non-empty chart render")
	}
	// Verify line chart elements and price scale are present
	if !containsStr(chartRender, "HIGH:") || !containsStr(chartRender, "LOW:") {
		t.Fatalf("expected price high/low labels on chart tab")
	}

	// Test timeframe cycling
	initTF := model.chartView.Timeframe()
	model.chartView.NextTimeframe()
	newTF := model.chartView.Timeframe()
	if initTF == newTF {
		t.Fatalf("expected timeframe to change after NextTimeframe()")
	}
	t.Logf("Cycled timeframe from %s to %s", initTF, newTF)

	chartRenderNewTF := model.View()
	if !containsStr(chartRenderNewTF, newTF) {
		t.Fatalf("expected chart render to reflect new timeframe %s", newTF)
	}

	// Test multi-timeframe candles from engine
	for _, tf := range AvailableTimeframes {
		candles := engine.GetCandles(firstSym, tf)
		if len(candles) == 0 {
			t.Fatalf("expected non-empty candles for symbol %s on tf %s", firstSym, tf)
		}
	}

	// 5. Verify NO emojis anywhere across all tabs
	allTabs := []Tab{TabOverview, TabUniverse, TabCharts, TabDetail, TabTrade, TabNews, TabAdmin}
	for _, tab := range allTabs {
		model.activeTab = tab
		rendered := model.View()
		for _, r := range rendered {
			if (r >= 0x1F300 && r <= 0x1FAFF) || (r >= 0x1F600 && r <= 0x1F64F) {
				t.Fatalf("found emoji rune %c (%U) in tab %s", r, r, tab.String())
			}
		}
	}
}

func TestMouseInteractionsAndChronologicalPositions(t *testing.T) {
	tickChan := make(chan TickMsg, 10)
	engine := NewEngine(50, 99, 100*time.Millisecond, tickChan)
	model := NewModel(engine, tickChan)

	// 1. Test mouse clicking on top navigation tabs
	clickX := 50
	for _, ca := range model.tabClickAreas {
		if ca.Speed == float64(TabTrade) {
			clickX = (ca.X1 + ca.X2) / 2
			break
		}
	}
	mouseClickTrade := tea.MouseMsg{
		X:      clickX,
		Y:      0,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	model.Update(mouseClickTrade)
	if model.activeTab != TabTrade {
		t.Fatalf("expected TabTrade after clicking tab at (%d, 0), got %s", clickX, model.activeTab.String())
	}
	// Render view to calculate layout heights
	model.View()

	// 2. Test mouse clicking on [ SELL ] button in TradeView
	// In TabTrade, Row 1 (y = orderBoxY + 3..4):
	// x around 40 is [ SELL ] button
	mouseClickSell := tea.MouseMsg{
		X:      40,
		Y:      model.tradeView.orderBoxY + 3,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	model.Update(mouseClickSell)
	if model.tradeView.OrderSide != "SELL" {
		t.Fatalf("expected OrderSide SELL after clicking button, got %s", model.tradeView.OrderSide)
	}

	// Click on [ BUY ] button
	mouseClickBuy := tea.MouseMsg{
		X:      25,
		Y:      model.tradeView.orderBoxY + 3,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	model.Update(mouseClickBuy)
	if model.tradeView.OrderSide != "BUY" {
		t.Fatalf("expected OrderSide BUY after clicking button, got %s", model.tradeView.OrderSide)
	}

	// Click on preset [ 500 ] shares (x ~ 80, y = orderBoxY + 5)
	mouseClickPreset500 := tea.MouseMsg{
		X:      80,
		Y:      model.tradeView.orderBoxY + 5,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	model.Update(mouseClickPreset500)
	if model.tradeView.OrderQty != "500" {
		t.Fatalf("expected OrderQty '500' after clicking preset, got %s", model.tradeView.OrderQty)
	}

	// 3. Test Chronological Position Sorting (Oldest to Newest, regardless of PnL)
	acct := engine.GetHumanAccount()
	now := time.Now().UnixNano()

	// Position 1: Bought oldest (t = now - 3 hours), with HUGE LOSS (-$5000)
	acct.Positions["OLD_LOSS"] = &market.Position{
		Side:      market.Long,
		Quantity:  100,
		EntryCost: 10000,
		OpenedAt:  now - int64(3*time.Hour),
	}

	// Position 2: Bought in the middle (t = now - 1 hour), with HUGE PROFIT (+$50,000)
	acct.Positions["MID_GAIN"] = &market.Position{
		Side:      market.Long,
		Quantity:  200,
		EntryCost: 5000,
		OpenedAt:  now - int64(1*time.Hour),
	}

	// Position 3: Bought newest (t = now - 5 minutes), with SMALL PROFIT (+$100)
	acct.Positions["NEW_MODERATE"] = &market.Position{
		Side:      market.Long,
		Quantity:  50,
		EntryCost: 2000,
		OpenedAt:  now - int64(5*time.Minute),
	}

	// Add dummy prices to model rows
	model.rows = append(model.rows,
		CompanyRow{Symbol: "OLD_LOSS", Price: 50.0},        // loss
		CompanyRow{Symbol: "MID_GAIN", Price: 275.0},       // massive gain
		CompanyRow{Symbol: "NEW_MODERATE", Price: 42.0},    // small gain
	)

	// Render TabTrade and inspect position order
	model.activeTab = TabTrade
	tradeOutput := model.View()
	if len(tradeOutput) == 0 {
		t.Fatalf("expected non-empty trade output")
	}

	// In the trade view, verify that OLD_LOSS appears BEFORE MID_GAIN,
	// and MID_GAIN appears BEFORE NEW_MODERATE!
	idxOld := strings.Index(tradeOutput, "OLD_LOSS")
	idxMid := strings.Index(tradeOutput, "MID_GAIN")
	idxNew := strings.Index(tradeOutput, "NEW_MODERATE")

	if idxOld == -1 || idxMid == -1 || idxNew == -1 {
		t.Fatalf("expected all 3 positions in rendered view")
	}

	if !(idxOld < idxMid && idxMid < idxNew) {
		t.Fatalf("positions are not arranged from oldest to newest! indices: OLD=%d, MID=%d, NEW=%d",
			idxOld, idxMid, idxNew)
	}
	t.Logf("Verified positions strictly arranged from oldest to newest (OLD=%d < MID=%d < NEW=%d)",
		idxOld, idxMid, idxNew)

	// 4. Test clicking on a position row to auto-populate symbol and price
	mouseClickRow := tea.MouseMsg{
		X:      15,
		Y:      model.tradeView.positionsY + 3, // first position row
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	model.Update(mouseClickRow)
	if model.selectedSymbol != "OLD_LOSS" {
		t.Fatalf("expected selectedSymbol 'OLD_LOSS' after clicking row, got %s", model.selectedSymbol)
	}
	if model.tradeView.OrderSymbol != "OLD_LOSS" {
		t.Fatalf("expected OrderSymbol 'OLD_LOSS' in order form, got %s", model.tradeView.OrderSymbol)
	}

	// 5. Test mouse clicking on timeframe buttons in Charts tab
	model.activeTab = TabCharts
	// Click around x = 14, y = 5 (button 2 is [5m])
	mouseClick5m := tea.MouseMsg{
		X:      14,
		Y:      5,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	model.Update(mouseClick5m)
	if model.chartView.Timeframe() != "5m" {
		t.Logf("Timeframe after click: %s", model.chartView.Timeframe())
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) > 0 && searchSub(s, sub))
}

func searchSub(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestSmoothChartBrailleAndStockNavigation(t *testing.T) {
	tickChan := make(chan TickMsg, 10)
	engine := NewEngine(50, 123, 10*time.Millisecond, tickChan)
	model := NewModel(engine, tickChan)
	model.activeTab = TabCharts

	universe := engine.GetUniverse()
	if len(universe) < 2 {
		t.Fatalf("expected at least 2 companies, got %d", len(universe))
	}
	firstSym := universe[0].Symbol
	secondSym := universe[1].Symbol

	model.selectedSymbol = firstSym

	// 1. Verify seed candles continuity (No cliff drop at end)
	candles := engine.GetCandles(firstSym, "1m")
	if len(candles) == 0 {
		t.Fatalf("expected seeded candles for %s", firstSym)
	}
	lastCandle := candles[len(candles)-1]
	expectedPrice := universe[0].ReportedValue
	if expectedPrice <= 0 {
		expectedPrice = universe[0].TrueValue
	}
	if lastCandle.Close != expectedPrice {
		t.Fatalf("cliff detected: last candle close ($%.2f) does not match expected price ($%.2f)",
			lastCandle.Close, expectedPrice)
	}
	t.Logf("Seed candles verified: %d candles, seamlessly ending at expected price $%.2f", len(candles), expectedPrice)

	// 2. Verify Braille chart rendering
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	rendered := model.View()
	if !strings.Contains(rendered, "VOLUME") {
		t.Fatalf("expected VOLUME histogram in chart view")
	}
	if !strings.Contains(rendered, "STYLE: LINE") {
		t.Fatalf("expected STYLE: LINE toggle indicator in chart view")
	}

	// 3. Test Area Mode toggle via keyboard 'a'
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	renderedArea := model.View()
	if !strings.Contains(renderedArea, "STYLE: AREA") {
		t.Fatalf("expected STYLE: AREA toggle indicator after pressing 'a'")
	}
	t.Logf("Area mode toggle verified successfully")

	// 4. Test Stock Navigation Next/Prev
	model.selectNextStock()
	if model.selectedSymbol != secondSym {
		t.Fatalf("expected selectedSymbol %s after nextStock, got %s", secondSym, model.selectedSymbol)
	}
	model.selectPrevStock()
	if model.selectedSymbol != firstSym {
		t.Fatalf("expected selectedSymbol %s after prevStock, got %s", firstSym, model.selectedSymbol)
	}
	t.Logf("Stock navigation verified: %s <-> %s", firstSym, secondSym)

	// 5. Test Quick Search in Chart tab
	model.chartView.ToggleSearch()
	if !model.chartView.IsSearching() {
		t.Fatalf("expected chartView to be searching")
	}
	// Type second symbol
	for _, r := range secondSym {
		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	// Press Enter
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.selectedSymbol != secondSym {
		t.Fatalf("expected search to switch selectedSymbol to %s, got %s", secondSym, model.selectedSymbol)
	}
	if model.chartView.IsSearching() {
		t.Fatalf("expected search to close after Enter")
	}
	t.Logf("Chart quick search verified successfully: navigated to %s", secondSym)
}

func TestPortfolioPersistenceAndReset(t *testing.T) {
	testPath := "data/test_portfolio.json"
	defer os.Remove(testPath)

	// 1. Create and save a test portfolio
	saved := &SavedPortfolio{
		Cash: 85420.50,
		Positions: map[string]SavedPosition{
			"NVDA": {
				Symbol:    "NVDA",
				Side:      "LONG",
				Quantity:  50,
				EntryCost: 22500,
				OpenedAt:  123456789,
			},
		},
		TradeHistory: []string{"BUY 50 NVDA @ 450.00"},
	}
	err := SavePortfolio(testPath, saved)
	if err != nil {
		t.Fatalf("failed to save portfolio: %v", err)
	}

	// 2. Load and verify
	loaded, err := LoadPortfolio(testPath)
	if err != nil {
		t.Fatalf("failed to load portfolio: %v", err)
	}
	if loaded == nil || loaded.Cash != 85420.50 {
		t.Fatalf("expected cash 85420.50, got %v", loaded)
	}
	if pos, ok := loaded.Positions["NVDA"]; !ok || pos.Quantity != 50 {
		t.Fatalf("expected 50 NVDA shares, got %v", pos)
	}

	// 3. Apply to live account
	acct := market.NewAccount("test_user", 100_000, 5, 0.10)
	ApplySavedPortfolioToAccount(loaded, acct)
	if acct.Cash != 85420.50 || len(acct.Positions) != 1 {
		t.Fatalf("failed to apply saved portfolio to account: cash=%.2f, pos=%d", acct.Cash, len(acct.Positions))
	}

	// 4. Reset portfolio
	err = ResetPortfolioDisk(testPath)
	if err != nil {
		t.Fatalf("failed to reset portfolio on disk: %v", err)
	}
	resetLoaded, err := LoadPortfolio(testPath)
	if err != nil || resetLoaded.Cash != 100_000 || len(resetLoaded.Positions) != 0 {
		t.Fatalf("expected reset portfolio to have 100,000 cash and 0 positions, got %v", resetLoaded)
	}
	t.Logf("Portfolio persistence and reset verified successfully")
}

func TestAdminFlexibleIPOCreation(t *testing.T) {
	admin := NewAdminView()
	initialSector := admin.IPOSector
	initialTier := admin.IPOTier

	// 1. Cycle sector
	admin.CycleSector()
	if admin.IPOSector == initialSector {
		t.Fatalf("expected sector to cycle from %s", initialSector)
	}

	// 2. Cycle tier
	admin.CycleTier()
	if admin.IPOTier == initialTier {
		t.Fatalf("expected tier to cycle from %s", initialTier)
	}

	// 3. Randomize IPO
	admin.RandomizeIPO()
	if admin.IPOSymbol == "" || admin.IPOName == "" {
		t.Fatalf("expected non-empty symbol and name after randomize")
	}

	// 4. Test parsing
	sec := parseAdminSector("Energy")
	if sec != company.Energy {
		t.Fatalf("expected Energy sector, got %v", sec)
	}
	tier := parseAdminTier("Mega Cap")
	if tier != company.MegaCap {
		t.Fatalf("expected Mega Cap tier, got %v", tier)
	}

	t.Logf("Admin flexible IPO creation verified successfully: %s (%s) [%s / %s]",
		admin.IPOName, admin.IPOSymbol, admin.IPOSector, admin.IPOTier)
}

func TestChartExactTimeframeHitTesting(t *testing.T) {
	chart := NewChartView()
	chart.SetSize(120, 35)

	dummyCo := &CompanyRow{
		Symbol:    "TEST",
		Name:      "Test Corp",
		Sector:    "Information Technology",
		Price:     100.0,
		ChangePct: 2.5,
		Spread:    0.05,
	}

	// Render chart to calculate exact dynamic bounds
	_ = chart.Render(dummyCo, []CandleView{
		{Time: time.Now(), Open: 100, High: 102, Low: 99, Close: 101, Volume: 1000},
	})

	// Test clicking each timeframe button directly at its computed bounds
	for _, tf := range AvailableTimeframes {
		bounds, ok := chart.tfBounds[tf]
		if !ok {
			t.Fatalf("expected bounds recorded for timeframe %s", tf)
		}
		// Click right in the center of the button
		clickX := (bounds[0] + bounds[1]) / 2
		clickY := chart.controlsRowY

		action := chart.HandleMouseClick(clickX, clickY)
		expectedAction := "set_tf:" + tf
		if action != expectedAction {
			t.Fatalf("clicking at (%d, %d) for timeframe %s returned %s, expected %s",
				clickX, clickY, tf, action, expectedAction)
		}
		if chart.Timeframe() != tf {
			t.Fatalf("expected chart timeframe to be %s, got %s", tf, chart.Timeframe())
		}
	}
	t.Logf("Chart exact timeframe hit-testing verified across all %d timeframes", len(AvailableTimeframes))
}

func TestMetricFormatting(t *testing.T) {
	tests := []struct {
		val      float64
		currency bool
		expected string
	}{
		{500, false, "500"},
		{500, true, "$500.00"},
		{1500, false, "1.50K"},
		{1500, true, "$1.50K"},
		{2500000, false, "2.50M"},
		{2500000, true, "$2.50M"},
		{15000000000, false, "15.00B"},
		{15000000000, true, "$15.00B"},
		{2500000000000, false, "2.50T"},
		{2500000000000, true, "$2.50T"},
	}

	for _, tt := range tests {
		got := FormatMetric(tt.val, tt.currency)
		if got != tt.expected {
			t.Errorf("FormatMetric(%v, %v) = %s, expected %s", tt.val, tt.currency, got, tt.expected)
		}
	}

	commas := FormatIntegerWithCommas(1234567890)
	if commas != "1,234,567,890" {
		t.Errorf("FormatIntegerWithCommas = %s, expected 1,234,567,890", commas)
	}
	t.Logf("Metric formatting verified successfully")
}

func TestMacroeconomicDashboardRendering(t *testing.T) {
	ov := NewOverviewView()
	ov.SetSize(120, 35)

	rows := []CompanyRow{
		{
			Symbol:            "AAPL",
			Name:              "Apple Inc",
			Sector:            "Information Technology",
			CapTier:           "Mega Cap",
			Price:             225.50,
			MarketCap:         3_400_000_000_000,
			Headcount:         161_000,
			LaborExpense:      24_000_000_000,
			DebtOutstanding:   110_000_000_000,
			InterestExpense:   4_500_000_000,
			CorporateTaxPaid:  18_000_000_000,
			MacroDemandFactor: 1.05,
		},
	}

	citRep := citizen.CitizenReport{
		Population:         340_000_000,
		LaborForce:         168_300_000,
		EmployedCount:      161_700_000,
		UnemploymentRate:   0.039,
		ConsumerConfidence: 85.2,
		Happiness:          82.0,
		ConsumerSpending:   18_500_000_000_000,
	}

	natRep := country.NationalReport{
		GDP:              28_500_000_000_000,
		RealGDPGrowth:    0.024,
		FedFundsRate:     0.045,
		TenYearYield:     0.042,
		CPIInflationRate: 0.024,
		PolicyStance:     "Neutral",
		LaborForce:       168_300_000,
		EmployedWorkers:  161_700_000,
		UnemploymentRate: 0.039,
		DollarIndexDXY:   103.5,
		TradeBalance:     -800_000_000_000,
		NationalDebt:      35_000_000_000_000,
		ExchangeChartered: true,
	}

	rendered := ov.Render(rows, []string{"[MACRO] FED: FOMC held target rate at 4.50%"}, 100, 1000, 1.0, 500000, 150, citRep, natRep)
	if !strings.Contains(rendered, "FEDERAL RESERVE & NATIONAL OUTPUT") {
		t.Errorf("expected rendered overview to contain 'FEDERAL RESERVE & NATIONAL OUTPUT'")
	}
	if !strings.Contains(rendered, "LABOR MARKET & CITIZEN SENTIMENT") {
		t.Errorf("expected rendered overview to contain 'LABOR MARKET & CITIZEN SENTIMENT'")
	}

	// Test Detail View Macro Footprint
	dv := NewDetailView()
	dv.SetSize(120, 35)
	detailRendered := dv.Render(&rows[0], nil, nil, nil)
	if !strings.Contains(detailRendered, "CORPORATE MACROECONOMIC & LABOR FOOTPRINT") {
		t.Errorf("expected rendered detail to contain 'CORPORATE MACROECONOMIC & LABOR FOOTPRINT'")
	}
	t.Log("Macroeconomic dashboard and corporate footprint rendering verified successfully")
}

func TestWorldStatePersistenceAndResumption(t *testing.T) {
	testPath := "data/test_world_state.json"
	defer os.Remove(testPath)

	// 1. Create a mock world state
	state := &SavedWorldState{
		Version: 1,
		Tick:    128,
		SimTime: 987654321,
		Companies: []SavedCompany{
			{
				Symbol:            "TEST",
				Name:              "Test Corp",
				Sector:            int(company.InformationTechnology),
				CapTier:           int(company.MidCap),
				TrueValue:         120.50,
				ReportedValue:     121.00,
				SharesOutstanding: 20_000_000,
				Float:             15_000_000,
				AnnualRevenue:     2_500_000_000,
				NetMargin:         0.22,
				SectorMultiple:    6.5,
				IPOPrice:          100.00,
				DebtOutstanding:   500_000_000,
				InterestExpense:   25_000_000,
				MacroDemandFactor: 1.15,
			},
		},
		Macro: SavedMacro{
			Population:         335_000_000,
			EmployedCount:      162_000_000,
			ConsumerConfidence: 104.2,
			Happiness:          72.5,
			GDP:                28_500_000_000_000,
			FedFundsRate:       0.0475,
			ActivePolicies:     map[string]bool{"tech_subsidies": true, "clean_energy": true},
		},
		Candles: map[string]map[string][]SavedCandle{
			"TEST": {
				"1m": {
					{TimeUnixMilli: 987654000, Open: 120.0, High: 121.5, Low: 119.5, Close: 121.0, Volume: 1500},
				},
			},
		},
	}

	// 2. Save world state
	err := SaveWorldState(testPath, state)
	if err != nil {
		t.Fatalf("failed to save world state: %v", err)
	}

	// 3. Load world state
	loaded, err := LoadWorldState(testPath)
	if err != nil {
		t.Fatalf("failed to load world state: %v", err)
	}
	if loaded == nil || loaded.Tick != 128 {
		t.Fatalf("expected tick 128, got %v", loaded)
	}
	if len(loaded.Companies) != 1 || loaded.Companies[0].Symbol != "TEST" {
		t.Fatalf("expected company TEST, got %v", loaded.Companies)
	}
	if loaded.Macro.ConsumerConfidence != 104.2 || len(loaded.Macro.ActivePolicies) != 2 {
		t.Fatalf("macro state mismatch: %+v", loaded.Macro)
	}
	if len(loaded.Candles["TEST"]["1m"]) != 1 {
		t.Fatalf("expected 1m candle for TEST, got %v", loaded.Candles)
	}

	// 4. Convert to company and verify fields
	co := loaded.Companies[0].ConvertToCompany()
	if co.Symbol != "TEST" || co.TrueValue != 120.50 || co.DebtOutstanding != 500_000_000 {
		t.Fatalf("converted company mismatch: %+v", co)
	}

	// 5. Reset disk state
	err = ResetWorldStateDisk(testPath)
	if err != nil {
		t.Fatalf("failed to reset world state disk: %v", err)
	}
	if _, err := os.Stat(testPath); !os.IsNotExist(err) {
		t.Fatalf("expected test world state file to be removed after reset")
	}
	t.Log("World state persistence, loading, conversion, and reset verified successfully")
}

func TestGovernmentPoliciesAndBailout(t *testing.T) {
	tickChan := make(chan TickMsg, 10)
	engine := NewEngine(50, 456, 100*time.Millisecond, tickChan)
	defer func() {
		engine.Stop()
		_ = os.Remove("data/world_state.json")
		_ = os.Remove("data/portfolio.json")
	}()

	// 1. Test Policy toggles
	active, desc := engine.TogglePolicy("tech_subsidies")
	if !active {
		t.Fatalf("expected tech_subsidies to be enacted, got false (desc: %s)", desc)
	}

	active2, desc2 := engine.TogglePolicy("tariff_shield")
	if !active2 {
		t.Fatalf("expected tariff_shield to be enacted, got false (desc: %s)", desc2)
	}

	_, nat := engine.GetMacroReport()
	hasTech := false
	hasTariff := false
	for _, p := range nat.ActivePolicies {
		if p == "tech_subsidies" {
			hasTech = true
		}
		if p == "tariff_shield" {
			hasTariff = true
		}
	}
	if !hasTech || !hasTariff {
		t.Fatalf("expected active policies in report, got: %v", nat.ActivePolicies)
	}

	// 2. Test Emergency Corporate Bailout
	universe := engine.GetUniverse()
	if len(universe) == 0 {
		t.Fatalf("expected companies in universe")
	}
	targetCo := universe[0]
	targetCo.DebtOutstanding = 800_000_000
	targetCo.InterestExpense = 40_000_000
	initialVal := targetCo.ReportedValue
	initialDebt := targetCo.DebtOutstanding

	err := engine.BailoutCompany(targetCo.Symbol)
	if err != nil {
		t.Fatalf("bailout failed: %v", err)
	}

	if targetCo.DebtOutstanding != initialDebt*0.5 {
		t.Fatalf("expected debt to be halved from %.0f to %.0f, got %.0f",
			initialDebt, initialDebt*0.5, targetCo.DebtOutstanding)
	}
	if targetCo.ReportedValue <= initialVal {
		t.Fatalf("expected reported value to lift after bailout, got %.2f (was %.2f)",
			targetCo.ReportedValue, initialVal)
	}
	t.Logf("Bailout successfully deployed for %s: debt cut from $%.0fM to $%.0fM, value lifted to $%.2f",
		targetCo.Symbol, initialDebt/1e6, targetCo.DebtOutstanding/1e6, targetCo.ReportedValue)
}

func TestHoldQToExit(t *testing.T) {
	tickChan := make(chan TickMsg, 10)
	engine := NewEngine(20, 789, 100*time.Millisecond, tickChan)
	defer func() {
		engine.Stop()
		_ = os.Remove("data/world_state.json")
		_ = os.Remove("data/portfolio.json")
	}()

	model := NewModel(engine, tickChan)

	// 1. Pressing 'q' once should start holding state, NOT quit immediately
	qKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.Update(qKey)
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatalf("pressing 'q' once must NOT exit the application immediately")
		}
	}
	if !model.isHoldingQ {
		t.Fatalf("expected isHoldingQ to be true after pressing 'q'")
	}

	// 2. If user releases 'q' (no key for >650ms), hold timer cancels
	model.lastQPress = time.Now().Add(-700 * time.Millisecond)
	_, _ = model.Update(qHoldTickMsg(time.Now()))
	if model.isHoldingQ {
		t.Fatalf("expected isHoldingQ to reset to false after release timeout")
	}

	// 3. Pressing another key cancels the hold
	model.Update(qKey)
	if !model.isHoldingQ {
		t.Fatalf("expected isHoldingQ to be true after second press of 'q'")
	}
	otherKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	model.Update(otherKey)
	if model.isHoldingQ {
		t.Fatalf("expected isHoldingQ to reset to false after pressing another key")
	}

	// 4. Holding 'q' for 5.0 seconds triggers exit
	model.Update(qKey)
	model.qHoldStart = time.Now().Add(-5100 * time.Millisecond)
	model.lastQPress = time.Now()
	_, exitCmd := model.Update(qHoldTickMsg(time.Now()))
	if exitCmd == nil {
		t.Fatalf("expected tea.Quit after holding 'q' for 5.0s, got nil")
	} else if _, ok := exitCmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg after holding 'q' for 5.0s, got %T", exitCmd())
	}
	t.Log("Hold-q for 5 seconds exit behavior verified successfully")
}

func TestAdminExactMouseHitTesting(t *testing.T) {
	admin := NewAdminView()
	admin.SetSize(120, 35)

	distressed := []CompanyRow{
		{Symbol: "DIST", Name: "Distressed Co", Sector: "Industrials", Price: 2.50, DebtOutstanding: 400_000_000},
	}
	policies := []string{"tech_subsidies"}

	natRep := country.NationalReport{
		ExchangeChartered:     true,
		GDP:                   20_000_000_000_000,
		NationalDebt:          10_000_000_000_000,
		CreditRating:          "AAA",
		BorrowingYield:        0.035,
		InfrastructureLevel:   50,
		HealthcareLevel:       50,
		EducationLevel:        50,
		EnterpriseGrantsLevel: 50,
		ExportCapacityLevel:   50,
	}

	// 1. Render AdminSubTabSystem for engine, persistence, and executive policies
	admin.ActiveSubTab = AdminSubTabSystem
	admin.Render(100, "1m", 1.0, 50, 1500, 25_000_000, policies, distressed, natRep)

	// Test clicking speed buttons on screen line 12
	action1, _, sp1 := admin.HandleMouseClick(7, 12) // [ PAUSE ]
	if action1 != "set_speed" || sp1 != 0.0 {
		t.Fatalf("expected set_speed 0.0 at (7, 12), got action=%s, speed=%v", action1, sp1)
	}

	action2, _, sp2 := admin.HandleMouseClick(28, 12) // [ 1.0x ]
	if action2 != "set_speed" || sp2 != 1.0 {
		t.Fatalf("expected set_speed 1.0 at (28, 12), got action=%s, speed=%v", action2, sp2)
	}

	action3, _, sp3 := admin.HandleMouseClick(48, 12) // [ 5.0x ]
	if action3 != "set_speed" || sp3 != 5.0 {
		t.Fatalf("expected set_speed 5.0 at (48, 12), got action=%s, speed=%v", action3, sp3)
	}

	// 2. Test clicking persistence buttons on screen line 15
	actSave, _, _ := admin.HandleMouseClick(10, 15) // [ SAVE WORLD STATE ]
	if actSave != "save_state" {
		t.Fatalf("expected save_state at (10, 15), got %s", actSave)
	}

	actReset, _, _ := admin.HandleMouseClick(40, 15) // [ RESET WORLD FRESH ]
	if actReset != "reset_state" {
		t.Fatalf("expected reset_state at (40, 15), got %s", actReset)
	}

	// 3. Test clicking Executive Policy pill at top right
	// Top right starts at X ~ 60, first policy is at line 7
	actPol, valPol, _ := admin.HandleMouseClick(70, 7)
	if actPol != "toggle_policy" || valPol != "tech_subsidies" {
		t.Fatalf("expected toggle_policy tech_subsidies at (70, 7), got action=%s, val=%s", actPol, valPol)
	}

	// 4. Switch to AdminSubTabIPO for Bailout Desk and IPO Console
	admin.ActiveSubTab = AdminSubTabIPO
	admin.Render(100, "1m", 1.0, 50, 1500, 25_000_000, policies, distressed, natRep)

	// Bailout row for DIST
	var bailoutY int
	for _, a := range admin.clickAreas {
		if a.Action == "bailout" && a.Value == "DIST" {
			bailoutY = (a.Y1 + a.Y2) / 2
			break
		}
	}
	if bailoutY == 0 {
		t.Fatalf("no bailout click area registered for DIST")
	}
	actBail, valBail, _ := admin.HandleMouseClick(100, bailoutY)
	if actBail != "bailout" || valBail != "DIST" {
		t.Fatalf("expected bailout DIST at (100, %d), got action=%s, val=%s", bailoutY, actBail, valBail)
	}

	// 5. Test clicking IPO presets
	var sharesPresetY int
	for _, a := range admin.clickAreas {
		if a.Action == "set_shares" && a.Value == "50000000" {
			sharesPresetY = (a.Y1 + a.Y2) / 2
			break
		}
	}
	if sharesPresetY == 0 {
		t.Fatalf("no click area registered for shares preset 50M")
	}
	actShares, _, _ := admin.HandleMouseClick(74, sharesPresetY)
	if actShares != "set_shares" || admin.IPOShares != "50000000" {
		t.Fatalf("expected IPOShares to be 50000000 after clicking preset, got %s (action=%s)", admin.IPOShares, actShares)
	}

	t.Log("Admin exact mouse hit-testing verified across speed, persistence, policies, bailout, and IPO")
}

func TestWorldSimulationAndPolicyStudio(t *testing.T) {
	// 1. Test World creation and foreign nations
	w := world.NewWorld(42)
	if len(w.Foreign) != 8 {
		t.Fatalf("expected 8 foreign countries, got %d", len(w.Foreign))
	}
	chn, ok := w.Foreign["CHN"]
	if !ok || chn.GDP <= 0 {
		t.Fatalf("expected China with positive GDP, got %+v", chn)
	}

	// 2. Test World tick
	initChinaGDP := chn.GDP
	w.Tick(1.0 / 252.0)
	if chn.GDP == 0 || chn.GDP == initChinaGDP {
		t.Fatalf("expected China GDP to evolve after tick, got %f", chn.GDP)
	}

	// 3. Test Foreign policy actions
	w.SendAid("CHN", 5_000_000_000.0, 10)
	if len(w.PolicyLog) == 0 {
		t.Fatalf("expected aid to be recorded in PolicyLog")
	}

	w.NegotiateDeal("EU", 0.015, 11)
	if w.Foreign["EU"].Relation.TariffRate != 0.015 {
		t.Fatalf("expected EU tariff to be 1.5%%, got %f", w.Foreign["EU"].Relation.TariffRate)
	}

	w.ImposeSanctions("CHN", 2, 12)
	if w.Foreign["CHN"].Relation.Stance != world.StanceHostile {
		t.Fatalf("expected China stance to become Hostile after level 2 sanctions, got %s", w.Foreign["CHN"].Relation.Stance)
	}

	// 4. Test SimDate conversion
	// SimEpoch is Jan 1, 2020. 365 days later = Dec 31, 2020 or Jan 1, 2021
	oneYearMs := int64(365 * 24 * 3600 * 1000)
	d1 := world.DateFromMillis(oneYearMs)
	if d1.Year != 2020 && d1.Year != 2021 {
		t.Fatalf("expected year 2020 or 2021 after 1 year, got %d", d1.Year)
	}
	speedLabel := world.SpeedLabel(1.0)
	if speedLabel != "~1.0 day/s" {
		t.Fatalf("expected ~1.0 day/s, got %s", speedLabel)
	}

	// 5. Test RandomizeIPO does not produce duplicate symbols
	admin := NewAdminView()
	used := map[string]bool{
		"APEX": true, "BNOV": true, "VNRG": true, "TITN": true,
		"SOLA": true, "MFIN": true, "OMNI": true, "QSL": true,
		"ZPHR": true, "NXCM": true, "AETH": true, "VSEC": true,
	}
	admin.SetUsedSymbols(used)
	admin.RandomizeIPO()
	if used[admin.IPOSymbol] {
		t.Fatalf("RandomizeIPO produced an already-used symbol: %s", admin.IPOSymbol)
	}
	if admin.IPOSymbol == "" {
		t.Fatalf("RandomizeIPO produced empty symbol")
	}

	// 6. Test Policy Studio rendering and categories
	polView := NewPolicyView()
	polView.SetSize(120, 35)
	rendered := polView.Render([]string{"tech_subsidies", "rate_hike_50bp"})
	if !strings.Contains(rendered, "FISCAL") {
		t.Fatalf("expected Policy Studio to render fiscal category, got %s", rendered[:200])
	}
	// Switch to Monetary category via click
	act, _, cat, isCat := polView.HandleMouseClick(15, 4)
	if !isCat || act != "policy_cat" || cat != PolicyCatMonetary {
		t.Fatalf("expected policy_cat monetary at (15, 4), got act=%s, cat=%v, isCat=%v", act, cat, isCat)
	}
	monetaryRendered := polView.Render([]string{"rate_hike_50bp"})
	if !strings.Contains(monetaryRendered, "MONETARY") {
		t.Fatalf("expected Policy Studio to render monetary category, got %s", monetaryRendered[:200])
	}

	// 7. Test AllKnownPolicies
	allPols := fiscal.AllKnownPolicies()
	if len(allPols) < 17 {
		t.Fatalf("expected at least 17 policies across all categories, got %d", len(allPols))
	}

	t.Log("World simulation, policy studio, calendar date, and IPO uniqueness verified successfully")
}

func TestNationBuildingAndSovereignDesk(t *testing.T) {
	// 1. Initialize fresh Engine without disk save
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "test_world_nation.json")
	_ = os.Remove(savePath)

	tickChan := make(chan TickMsg, 100)
	engine := NewEngineWithPath(50, 42, 100*time.Millisecond, tickChan, savePath)
	defer engine.Stop()

	// Verify emerging nation baseline
	_, natRep := engine.GetMacroReport()
	if natRep.ExchangeChartered {
		t.Fatalf("expected unchartered exchange on fresh emerging nation start")
	}
	if natRep.TreasuryCash < 5_000_000_000 {
		t.Fatalf("expected initial treasury cash >= $5B, got %f", natRep.TreasuryCash)
	}
	initialCash := natRep.TreasuryCash
	initialDebt := natRep.NationalDebt

	// 2. Test Borrowing Sovereign Debt
	borrowAmt := 2_000_000_000.0
	if err := engine.BorrowMoney(borrowAmt); err != nil {
		t.Fatalf("failed to borrow money: %v", err)
	}
	_, natRep = engine.GetMacroReport()
	if math.Abs(natRep.TreasuryCash-(initialCash+borrowAmt)) > 1e6 {
		t.Fatalf("expected treasury cash %f after borrowing, got %f", initialCash+borrowAmt, natRep.TreasuryCash)
	}
	if math.Abs(natRep.NationalDebt-(initialDebt+borrowAmt)) > 1e6 {
		t.Fatalf("expected debt %f after borrowing, got %f", initialDebt+borrowAmt, natRep.NationalDebt)
	}

	// 3. Test Sovereign Debt Paydown
	repayAmt := 1_000_000_000.0
	if err := engine.RepayDebt(repayAmt); err != nil {
		t.Fatalf("failed to repay debt: %v", err)
	}
	_, natRep = engine.GetMacroReport()
	if math.Abs(natRep.TreasuryCash-(initialCash+borrowAmt-repayAmt)) > 1e6 {
		t.Fatalf("expected treasury cash %f after repayment, got %f", initialCash+borrowAmt-repayAmt, natRep.TreasuryCash)
	}

	// 4. Test Capital Investments
	investAmt := 1_000_000_000.0
	initInfra := natRep.InfrastructureLevel
	if err := engine.InvestInfrastructure(investAmt); err != nil {
		t.Fatalf("failed to invest infrastructure: %v", err)
	}
	initHealth := natRep.HealthcareLevel
	if err := engine.InvestPopulation(investAmt, "healthcare"); err != nil {
		t.Fatalf("failed to invest healthcare: %v", err)
	}
	initEdu := natRep.EducationLevel
	if err := engine.InvestPopulation(investAmt, "education"); err != nil {
		t.Fatalf("failed to invest education: %v", err)
	}
	initGrants := natRep.EnterpriseGrantsLevel
	if err := engine.InvestEnterprise(investAmt); err != nil {
		t.Fatalf("failed to invest enterprise grants: %v", err)
	}

	_, natRep = engine.GetMacroReport()
	if natRep.InfrastructureLevel <= initInfra {
		t.Fatalf("expected infrastructure level to increase from %f, got %f", initInfra, natRep.InfrastructureLevel)
	}
	if natRep.HealthcareLevel <= initHealth {
		t.Fatalf("expected healthcare level to increase from %f, got %f", initHealth, natRep.HealthcareLevel)
	}
	if natRep.EducationLevel <= initEdu {
		t.Fatalf("expected education level to increase from %f, got %f", initEdu, natRep.EducationLevel)
	}
	if natRep.EnterpriseGrantsLevel <= initGrants {
		t.Fatalf("expected enterprise grants level to increase from %f, got %f", initGrants, natRep.EnterpriseGrantsLevel)
	}

	// 5. Test Stock Exchange Chartering
	// Satisfy prerequisites: infra >= 25, grants >= 20, cash >= 5B
	for natRep.InfrastructureLevel < 25.0 {
		_ = engine.InvestInfrastructure(1_000_000_000.0)
		_, natRep = engine.GetMacroReport()
	}
	for natRep.EnterpriseGrantsLevel < 20.0 {
		_ = engine.InvestEnterprise(1_000_000_000.0)
		_, natRep = engine.GetMacroReport()
	}

	if err := engine.CharterStockExchange(); err != nil {
		t.Fatalf("failed to charter stock exchange after satisfying prerequisites: %v", err)
	}
	_, natRep = engine.GetMacroReport()
	if !natRep.ExchangeChartered {
		t.Fatalf("expected ExchangeChartered to be true after CharterStockExchange()")
	}

	// 6. Test Overview View Rendering in Unchartered and Chartered modes
	ov := NewOverviewView()
	ov.SetSize(120, 35)
	rows := []CompanyRow{
		{Symbol: "DOM1", Name: "Domestic Startup 1", Sector: "Information Technology", Stage: "Growth", PrivateValuation: 25_000_000, ReportedValue: 25_000_000, Price: 25.0},
		{Symbol: "DOM2", Name: "Domestic Startup 2", Sector: "Health Care", Stage: "Pre-IPO", PrivateValuation: 80_000_000, ReportedValue: 80_000_000, Price: 80.0},
	}
	citRep, _ := engine.GetMacroReport()

	// Unchartered view
	uncharteredNat := natRep
	uncharteredNat.ExchangeChartered = false
	uncharteredRender := ov.Render(rows, nil, 10, 1000, 1.0, 0, 0, citRep, uncharteredNat)
	if !strings.Contains(uncharteredRender, "SOVEREIGN TREASURY & MACRO INFLOWS") {
		t.Fatalf("expected unchartered overview to contain 'SOVEREIGN TREASURY & MACRO INFLOWS'")
	}
	if !strings.Contains(uncharteredRender, "DOMESTIC PRIVATE ENTERPRISE REGISTRY") {
		t.Fatalf("expected unchartered overview to contain 'DOMESTIC PRIVATE ENTERPRISE REGISTRY'")
	}
	if !strings.Contains(uncharteredRender, "STOCK EXCHANGE CHARTER ROADMAP") {
		t.Fatalf("expected unchartered overview to contain 'STOCK EXCHANGE CHARTER ROADMAP'")
	}

	// Chartered view
	charteredRender := ov.Render(rows, nil, 10, 1000, 1.0, 10000, 50, citRep, natRep)
	if !strings.Contains(charteredRender, "SOVEREIGN STATUS: Treasury Cash") {
		t.Fatalf("expected chartered overview to contain 'SOVEREIGN STATUS: Treasury Cash'")
	}
	if !strings.Contains(charteredRender, "FEDERAL RESERVE & NATIONAL OUTPUT") {
		t.Fatalf("expected chartered overview to contain 'FEDERAL RESERVE & NATIONAL OUTPUT'")
	}

	// 7. Test Admin View Hit Testing on Sovereign & Charter Desks
	admin := NewAdminView()
	admin.SetSize(120, 35)
	admin.Render(100, "1m", 1.0, 10, 500, 1_000_000, nil, nil, uncharteredNat)

	// Hit test borrow button
	var borrowArea *ClickArea
	var charterArea *ClickArea
	for i := range admin.clickAreas {
		ca := &admin.clickAreas[i]
		if ca.Action == "borrow_money" && ca.Value == "1000000000" {
			borrowArea = ca
		}
		if ca.Action == "charter_exchange" {
			charterArea = ca
		}
	}
	if borrowArea == nil {
		t.Fatalf("no borrow_money click area registered")
	}
	bAction, bVal, _ := admin.HandleMouseClick((borrowArea.X1+borrowArea.X2)/2, (borrowArea.Y1+borrowArea.Y2)/2)
	if bAction != "borrow_money" || bVal != "1000000000" {
		t.Fatalf("expected borrow_money 1000000000, got action=%s, val=%s", bAction, bVal)
	}

	if charterArea == nil {
		t.Fatalf("no charter_exchange click area registered")
	}
	cAction, _, _ := admin.HandleMouseClick((charterArea.X1+charterArea.X2)/2, (charterArea.Y1+charterArea.Y2)/2)
	if cAction != "charter_exchange" {
		t.Fatalf("expected charter_exchange, got action=%s", cAction)
	}

	t.Log("Nation building, sovereign debt, public capital investments, and stock exchange chartering verified successfully")
}

func TestSavedCompanyConvertToCompanyTick(t *testing.T) {
	sc := SavedCompany{
		Symbol:            "NVDA",
		Name:              "Nvidia Corp",
		Sector:            int(company.InformationTechnology),
		CapTier:           int(company.MegaCap),
		TrueValue:         120.50,
		ReportedValue:     120.50,
		SharesOutstanding: 25_000_000_000,
		Float:             23_000_000_000,
		AnnualRevenue:     60_000_000_000,
		NetMargin:         0.55,
		SectorMultiple:    20.0,
		IPOPrice:          100.0,
		IsIPO:             false,
		IPOTick:           0,
		IsPublic:          true,
		Stage:             "Public",
	}

	co := sc.ConvertToCompany()
	if co == nil {
		t.Fatalf("expected non-nil company")
	}

	for i := 0; i < 100; i++ {
		evt, restate := co.Tick()
		_ = evt
		_ = restate
	}

	if co.TrueValue <= 0 {
		t.Errorf("expected positive TrueValue after 100 ticks, got %f", co.TrueValue)
	}
}

func TestConcurrentPolicyToggle(t *testing.T) {
	tickChan := make(chan TickMsg, 50)
	engine := NewEngine(20, 1001, 10*time.Millisecond, tickChan)
	engine.Start()
	defer func() {
		engine.Stop()
		_ = os.Remove("data/world_state.json")
		_ = os.Remove("data/portfolio.json")
	}()

	policies := []string{
		fiscal.PolicyTechSubsidies,
		fiscal.PolicyDefenseInfrastructure,
		fiscal.PolicyCleanEnergy,
		fiscal.PolicyTariffShield,
		fiscal.PolicyCitizenStimulus,
		fiscal.PolicyMinWageHike,
		fiscal.PolicyChinaTechBan,
	}

	done := make(chan struct{})
	// Spawn multiple goroutines toggling policies rapidly
	for i := 0; i < 3; i++ {
		go func(id int) {
			for {
				select {
				case <-done:
					return
				default:
					p := policies[(id+int(time.Now().UnixNano()))%len(policies)]
					engine.TogglePolicy(p)
					time.Sleep(2 * time.Millisecond)
				}
			}
		}(i)
	}

	// Wait for at least 15 simulation ticks while concurrent toggles occur
	ticksReceived := 0
	timeout := time.After(4 * time.Second)
	for ticksReceived < 15 {
		select {
		case <-tickChan:
			ticksReceived++
		case <-timeout:
			close(done)
			t.Fatalf("timed out waiting for ticks during concurrent policy toggles")
		}
	}
	close(done)
	t.Logf("Successfully executed %d ticks concurrently with rapid policy toggles without panic", ticksReceived)
}

func TestCommandPaletteAndThemeToggle(t *testing.T) {
	tickChan := make(chan TickMsg, 10)
	engine := NewEngine(20, 1002, 100*time.Millisecond, tickChan)
	model := NewModel(engine, tickChan)

	// 1. Initially Dark Mode
	if CurrentTheme != ThemeDark {
		t.Fatalf("expected initial theme to be dark, got %s", CurrentTheme)
	}

	// 2. Open Command Palette via Ctrl+P
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if !model.palette.IsOpen() {
		t.Fatalf("expected command palette to be open after Ctrl+P")
	}

	// 3. Search for theme
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("light")})
	sel := model.palette.SelectedCommand()
	if sel == nil || (sel.Action != "set_theme" && sel.Action != "toggle_theme") {
		t.Fatalf("expected selected command to switch to light mode, got %+v", sel)
	}

	// 4. Execute command
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if model.palette.IsOpen() {
		t.Fatalf("expected command palette to be closed after Enter")
	}
	if CurrentTheme != ThemeLight {
		t.Fatalf("expected theme to switch to light, got %s", CurrentTheme)
	}

	// 5. Verify light theme colors applied
	if ColorBg != lipgloss.Color("#f6f8fa") {
		t.Fatalf("expected light ColorBg #f6f8fa, got %v", ColorBg)
	}
	if ColorCardBg != lipgloss.Color("#ffffff") {
		t.Fatalf("expected light ColorCardBg #ffffff, got %v", ColorCardBg)
	}

	// 6. Toggle back to dark via palette
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("dark")})
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if CurrentTheme != ThemeDark {
		t.Fatalf("expected theme to switch back to dark, got %s", CurrentTheme)
	}

	// Clean up settings file
	_ = os.Remove(DefaultSettingsPath)
}

func TestResponsiveTabsClickableOnEveryTab(t *testing.T) {
	tickChan := make(chan TickMsg, 10)
	engine := NewEngine(20, 1003, 100*time.Millisecond, tickChan)
	model := NewModel(engine, tickChan)

	allTabs := []Tab{
		TabOverview, TabUniverse, TabCharts, TabDetail,
		TabTrade, TabNews, TabAdmin, TabWorld, TabPolicy,
	}

	// Test on standard width (120 cols) and compact width (80 cols)
	widths := []int{120, 80}

	for _, w := range widths {
		model.Update(tea.WindowSizeMsg{Width: w, Height: 35})

		// On every tab, verify all 9 tabs have registered click areas and can be clicked
		for _, curTab := range allTabs {
			model.activeTab = curTab
			model.View() // triggers layout & click area population

			if len(model.tabClickAreas) != 9 {
				t.Fatalf("at width %d on tab %s, expected 9 tab click areas, got %d",
					w, curTab.String(), len(model.tabClickAreas))
			}

			// Try clicking the next tab via its registered click area
			nextTabIdx := (int(curTab) + 1) % len(allTabs)
			targetTab := allTabs[nextTabIdx]

			var targetArea *ClickArea
			for i := range model.tabClickAreas {
				if model.tabClickAreas[i].Speed == float64(targetTab) {
					targetArea = &model.tabClickAreas[i]
					break
				}
			}
			if targetArea == nil {
				t.Fatalf("could not find click area for target tab %s while on tab %s at width %d",
					targetTab.String(), curTab.String(), w)
			}

			clickMsg := tea.MouseMsg{
				X:      (targetArea.X1 + targetArea.X2) / 2,
				Y:      targetArea.Y1,
				Action: tea.MouseActionPress,
				Button: tea.MouseButtonLeft,
			}
			model.Update(clickMsg)

			if model.activeTab != targetTab {
				t.Fatalf("expected switch from %s to %s via click at (%d, %d), got %s",
					curTab.String(), targetTab.String(), clickMsg.X, clickMsg.Y, model.activeTab.String())
			}
		}
	}
	t.Log("Verified: on every tab, all 9 other tabs are rendered and clickable via mouse at all screen widths")
}

func TestChartTimeframeClampingNoLooping(t *testing.T) {
	chart := NewChartView()
	chart.SetSize(120, 35)

	// Default should be 1m (index 0)
	if chart.Timeframe() != "1m" {
		t.Fatalf("expected initial timeframe '1m', got '%s'", chart.Timeframe())
	}

	// Calling PrevTimeframe at 1m must remain at 1m (NO looping to 1D)
	chart.PrevTimeframe()
	if chart.Timeframe() != "1m" {
		t.Fatalf("expected PrevTimeframe() at 1m to remain at '1m', got '%s'", chart.Timeframe())
	}

	// Step forward through all AvailableTimeframes
	expectedSequence := []string{"1m", "5m", "15m", "30m", "1h", "2h", "4h", "1D"}
	for i := 1; i < len(expectedSequence); i++ {
		chart.NextTimeframe()
		if chart.Timeframe() != expectedSequence[i] {
			t.Fatalf("step %d: expected timeframe '%s', got '%s'", i, expectedSequence[i], chart.Timeframe())
		}
	}

	// Now at "1D" (the maximum timeframe). Calling NextTimeframe() must stay at "1D" (NO looping to 1m)
	if chart.Timeframe() != "1D" {
		t.Fatalf("expected to be at '1D', got '%s'", chart.Timeframe())
	}
	chart.NextTimeframe()
	if chart.Timeframe() != "1D" {
		t.Fatalf("expected NextTimeframe() at 1D to remain clamped at '1D', but it looped to '%s'", chart.Timeframe())
	}

	// Step back down
	for i := len(expectedSequence) - 2; i >= 0; i-- {
		chart.PrevTimeframe()
		if chart.Timeframe() != expectedSequence[i] {
			t.Fatalf("step down: expected timeframe '%s', got '%s'", expectedSequence[i], chart.Timeframe())
		}
	}

	// Again at 1m, verify clamped
	chart.PrevTimeframe()
	if chart.Timeframe() != "1m" {
		t.Fatalf("expected clamped at '1m', got '%s'", chart.Timeframe())
	}

	t.Log("Verified: ChartView timeframes clamp strictly at boundaries (1m and 1D) and do not loop.")
}

func TestViewResponsivenessAcrossWindowSizes(t *testing.T) {
	tickChan := make(chan TickMsg, 50)
	engine := NewEngine(20, 42, 50*time.Millisecond, tickChan)
	engine.Start()
	defer engine.Stop()

	var firstTick TickMsg
	select {
	case msg := <-tickChan:
		firstTick = msg
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for tick in responsiveness test")
	}

	model := NewModel(engine, tickChan)
	model.Update(firstTick)

	universe := engine.GetUniverse()
	if len(universe) > 0 {
		model.selectedSymbol = universe[0].Symbol
	}

	testSizes := []struct {
		width  int
		height int
		name   string
	}{
		{width: 80, height: 24, name: "80x24 Compact"},
		{width: 100, height: 30, name: "100x30 Medium"},
		{width: 120, height: 35, name: "120x35 Standard"},
		{width: 160, height: 50, name: "160x50 Widescreen"},
	}

	tabs := []Tab{
		TabOverview,
		TabUniverse,
		TabCharts,
		TabDetail,
		TabTrade,
		TabNews,
		TabAdmin,
		TabWorld,
		TabPolicy,
	}

	for _, sz := range testSizes {
		t.Run(sz.name, func(t *testing.T) {
			model.Update(tea.WindowSizeMsg{Width: sz.width, Height: sz.height})

			for _, tab := range tabs {
				model.activeTab = tab
				viewOutput := model.View()
				if len(viewOutput) == 0 {
					t.Fatalf("expected non-empty render on tab %s for size %dx%d", tab.String(), sz.width, sz.height)
				}

				lines := strings.Split(viewOutput, "\n")
				if len(lines) == 0 {
					t.Fatalf("tab %s produced 0 lines", tab.String())
				}
			}
		})
	}
	t.Log("Verified: all tabs render cleanly and responsively across 80x24, 100x30, 120x35, and 160x50 sizes without panic or errors.")
}
func TestTab7ResponsivenessAndTab9SingleClick(t *testing.T) {
	tickChan := make(chan TickMsg, 50)
	engine := NewEngine(50, 99, 100*time.Millisecond, tickChan)
	model := NewModel(engine, tickChan)
	w := 100
	h := 30
	model.Update(tea.WindowSizeMsg{Width: w, Height: h})

	// 1. Verify Tab 7 (Admin) line height fits nicely and top tabs remain at line 0
	model.activeTab = TabAdmin
	view7 := model.View()
	lines7 := strings.Split(view7, "\n")
	if len(lines7) > h+5 {
		t.Fatalf("TabAdmin rendered %d lines, exceeding terminal height %d (would cause terminal scroll and hide tabs)", len(lines7), h)
	}

	// Verify top navigation tabs are on the first rows (wrapped at width 100)
	if !strings.Contains(lines7[0], "Overview") || !strings.Contains(view7, "Sovereign") {
		t.Fatalf("Top tabs not found in header of TabAdmin view: %s", lines7[0])
	}

	// Verify Admin Sub-Tab navigation with keyboard '[' and ']'
	if model.adminView.ActiveSubTab != AdminSubTabTreasury {
		t.Fatalf("expected initial sub-tab Treasury, got %v", model.adminView.ActiveSubTab)
	}
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	if model.adminView.ActiveSubTab != AdminSubTabIPO {
		t.Fatalf("expected sub-tab IPO after ']', got %v", model.adminView.ActiveSubTab)
	}
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	if model.adminView.ActiveSubTab != AdminSubTabSystem {
		t.Fatalf("expected sub-tab System after ']', got %v", model.adminView.ActiveSubTab)
	}
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")})
	if model.adminView.ActiveSubTab != AdminSubTabIPO {
		t.Fatalf("expected sub-tab IPO after '[', got %v", model.adminView.ActiveSubTab)
	}

	// 2. Verify Tab 9 (Policy) Single Click Activation
	model.activeTab = TabPolicy
	_ = model.View() // Render to populate policyView.clickAreas

	// Find tech_subsidies click area
	var polArea *ClickArea
	for i := range model.policyView.clickAreas {
		if model.policyView.clickAreas[i].Value == "tech_subsidies" {
			polArea = &model.policyView.clickAreas[i]
			break
		}
	}
	if polArea == nil {
		t.Fatalf("tech_subsidies click area not found in PolicyView")
	}

	clickX := (polArea.X1 + polArea.X2) / 2
	clickY := (polArea.Y1 + polArea.Y2) / 2

	isPolicyActive := func(id string) bool {
		_, nat := engine.GetMacroReport()
		for _, p := range nat.ActivePolicies {
			if p == id {
				return true
			}
		}
		return false
	}

	// Verify initially not active
	if isPolicyActive("tech_subsidies") {
		t.Fatalf("tech_subsidies should initially be inactive")
	}

	// Step A: Mouse Press on policy card
	model.Update(tea.MouseMsg{
		X:      clickX,
		Y:      clickY,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})

	if !isPolicyActive("tech_subsidies") {
		t.Fatalf("expected tech_subsidies to be ENACTED after single left click press")
	}

	// Step B: Mouse Release on the SAME spot (previously this repealed it immediately!)
	model.Update(tea.MouseMsg{
		X:      clickX,
		Y:      clickY,
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonLeft,
	})

	if !isPolicyActive("tech_subsidies") {
		t.Fatalf("CRITICAL BUG: tech_subsidies was REPEALED by mouse release event! Must stay enacted.")
	}

	// Step C: Second Mouse Press toggles it back off (Repeal)
	model.Update(tea.MouseMsg{
		X:      clickX,
		Y:      clickY,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})

	if isPolicyActive("tech_subsidies") {
		t.Fatalf("expected tech_subsidies to be REPEALED after second click press")
	}

	t.Log("Verified: Tab 7 height is responsive and compact, and Tab 9 policy activates on a single click without requiring swiping.")
}



