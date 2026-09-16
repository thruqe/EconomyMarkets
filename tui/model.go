package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"economy/citizen"
	"economy/company"
	"economy/country"
	"economy/market"
	"economy/sim"
	"economy/world"
)

type Model struct {
	activeTab     Tab
	bridge        SimBridge
	tickChan      chan TickMsg
	width         int
	height        int
	startTime     time.Time

	// View Models
	overviewView *OverviewView
	universeView *UniverseView
	chartView    *ChartView
	detailView   *DetailView
	tradeView    *TradeView
	newsView     *NewsView
	adminView    *AdminView
	worldView    *WorldView
	policyView   *PolicyView

	// Cached Simulation State
	rows            []CompanyRow
	selectedSymbol  string
	newsEvents      []string
	tapeTrades      []string
	symbolTrades    map[string][]string
	currentTick     int64
	simTime         int64
	speedMultiplier float64
	totalTrades     int64
	totalVolume     float64

	// Macroeconomic Ecosystem State
	citizenReport  citizen.CitizenReport
	nationalReport country.NationalReport

	// Input modes
	isSearching bool

	// Hold 'q' for 5s to exit
	isHoldingQ bool
	qHoldStart time.Time
	lastQPress time.Time

	// Country setup modal state
	isConfiguringCountry bool
	countrySetupField    int
	countrySetupName     string
	countrySetupCurrency string
	countrySetupFlag     string
	countrySetupFounded  string

	// Responsive Tab Bar
	tabClickAreas []ClickArea
	headerHeight  int

	// Command Palette & Settings
	palette *CommandPalette
}

type qHoldTickMsg time.Time

func tickQHold() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return qHoldTickMsg(t)
	})
}

func NewModel(bridge SimBridge, tickChan chan TickMsg) *Model {
	settings := LoadSettings("")
	ApplyTheme(settings.Theme)

	m := &Model{
		activeTab:       TabOverview,
		bridge:          bridge,
		tickChan:        tickChan,
		width:           120,
		height:          35,
		startTime:       time.Now(),
		overviewView:    NewOverviewView(),
		universeView:    NewUniverseView(),
		chartView:       NewChartView(),
		detailView:      NewDetailView(),
		tradeView:       NewTradeView(),
		newsView:        NewNewsView(),
		adminView:       NewAdminView(),
		worldView:       NewWorldView(),
		policyView:      NewPolicyView(),
		palette:         NewCommandPalette(),
		symbolTrades:    make(map[string][]string),
		speedMultiplier: 1.0,
	}

	if bridge != nil {
		cit, nat := bridge.GetMacroReport()
		m.citizenReport = cit
		m.nationalReport = nat
	}

	m.refreshUniverse()
	if len(m.rows) > 0 {
		m.selectedSymbol = m.rows[0].Symbol
		m.tradeView.SetSelectedSymbol(m.rows[0].Symbol, m.rows[0].Price)
	}

	m.renderTabBar()
	return m
}

func (m *Model) Init() tea.Cmd {
	return waitForTick(m.tickChan)
}

func waitForTick(tickChan chan TickMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-tickChan
		if !ok {
			return nil
		}
		return msg
	}
}

func (m *Model) refreshUniverse() {
	if m.bridge == nil {
		return
	}
	companies := m.bridge.GetUniverse()
	prices := m.bridge.GetPrices()

	rows := make([]CompanyRow, len(companies))
	for i, co := range companies {
		p, hasP := prices[co.Symbol]
		mid := co.ReportedValue
		if mid <= 0 {
			mid = co.TrueValue
		}
		bid := mid - 0.02
		ask := mid + 0.02
		spread := 0.04
		if hasP && p.HasMid && p.Mid > 0 {
			mid = p.Mid
			bid = p.Bid
			ask = p.Ask
			spread = p.Spread
		}

		basePrice := co.IPOPrice
		if basePrice <= 0 {
			basePrice = co.TrueValue
		}
		chg := mid - basePrice
		chgPct := 0.0
		if basePrice > 0 {
			chgPct = (chg / basePrice) * 100.0
		}

		rows[i] = CompanyRow{
			Symbol:        co.Symbol,
			Name:          co.Name,
			Sector:        co.Sector.String(),
			CapTier:       co.CapTier.String(),
			Price:         mid,
			Bid:           bid,
			Ask:           ask,
			Spread:        spread,
			PrevPrice:     basePrice,
			Change:        chg,
			ChangePct:     chgPct,
			Volume:            co.SharesOutstanding * 0.02,
			ReportedValue:     co.ReportedValue,
			TrueValue:         co.TrueValue,
			PE:                mid / mathMax(0.01, (co.AnnualRevenue*co.NetMargin)/co.SharesOutstanding),
			PS:                mid / mathMax(0.01, co.AnnualRevenue/co.SharesOutstanding),
			MarketCap:         mid * co.SharesOutstanding,
			IsIPO:             co.IsIPO,
			Headcount:         co.Headcount,
			LaborExpense:      co.LaborExpense,
			DebtOutstanding:   co.DebtOutstanding,
			InterestExpense:   co.InterestExpense,
			CorporateTaxPaid:  co.CorporateTaxPaid,
			MacroDemandFactor: co.MacroDemandFactor,
			IsPublic:          co.IsPublic,
			Stage:             co.Stage,
			PrivateValuation:  co.PrivateValuation,
		}
	}
	m.rows = rows
	used := make(map[string]bool, len(rows))
	for _, r := range rows {
		used[r.Symbol] = true
	}
	m.adminView.SetUsedSymbols(used)
}

func mathMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.renderTabBar()
		availHeight := m.height - m.headerHeight - 2
		if availHeight < 10 {
			availHeight = 10
		}
		m.overviewView.SetSize(msg.Width, availHeight)
		m.universeView.SetSize(msg.Width, availHeight)
		m.chartView.SetSize(msg.Width, availHeight)
		m.detailView.SetSize(msg.Width, availHeight)
		m.tradeView.SetSize(msg.Width, availHeight)
		m.newsView.SetSize(msg.Width, availHeight)
		m.adminView.SetSize(msg.Width, availHeight)
		m.worldView.SetSize(msg.Width, availHeight)
		m.policyView.SetSize(msg.Width, availHeight)
		return m, nil

	case qHoldTickMsg:
		if !m.isHoldingQ {
			return m, nil
		}
		if time.Since(m.lastQPress) > 650*time.Millisecond {
			m.isHoldingQ = false
			return m, nil
		}
		if time.Since(m.qHoldStart) >= 5*time.Second {
			if m.bridge != nil {
				_ = m.bridge.SaveWorldState()
			}
			return m, tea.Quit
		}
		return m, tickQHold()

	case TickMsg:
		m.currentTick = int64(msg.Report.Tick)
		m.simTime = msg.SimTime
		m.speedMultiplier = msg.Speed
		m.totalTrades += int64(len(msg.Report.Trades))
		m.citizenReport = msg.Report.Citizen
		m.nationalReport = msg.Report.National

		// Process trades
		for _, tr := range msg.Report.Trades {
			m.totalVolume += tr.Quantity * tr.Price
			tradeStr := FormatTapeTrade(tr.Symbol, tr.Side, tr.Quantity, tr.Price, tr.TakerAgentID)
			m.tapeTrades = append(m.tapeTrades, tradeStr)
			if len(m.tapeTrades) > 200 {
				m.tapeTrades = m.tapeTrades[len(m.tapeTrades)-200:]
			}

			// Symbol-specific trade history
			m.symbolTrades[tr.Symbol] = append(m.symbolTrades[tr.Symbol], tradeStr)
			if len(m.symbolTrades[tr.Symbol]) > 50 {
				m.symbolTrades[tr.Symbol] = m.symbolTrades[tr.Symbol][len(m.symbolTrades[tr.Symbol])-50:]
			}
		}

		// Process corporate events
		for _, ev := range msg.Report.Events {
			evStr := formatEvent(ev)
			m.newsEvents = append(m.newsEvents, evStr)
			if len(m.newsEvents) > 200 {
				m.newsEvents = m.newsEvents[len(m.newsEvents)-200:]
			}
		}

		// Refresh universe prices
		m.refreshUniverse()

		return m, waitForTick(m.tickChan)

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Check if command palette is open
			if m.palette != nil && m.palette.IsOpen() {
				cmd := m.palette.HandleMouseClick(msg.X, msg.Y)
				if cmd != nil {
					return m.executePaletteCommand(cmd)
				}
				m.palette.Close()
				return m, nil
			}

			// 1. Top navigation tabs click (checks dynamic bounding boxes for all tabs)
			for _, ca := range m.tabClickAreas {
				if msg.X >= ca.X1 && msg.X <= ca.X2 && msg.Y >= ca.Y1 && msg.Y <= ca.Y2 {
					m.activeTab = Tab(ca.Speed)
					return m, nil
				}
			}
		}

		// 2. Mouse Wheel
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			switch m.activeTab {
			case TabUniverse:
				m.universeView.MoveUp()
				return m, nil
			case TabCharts:
				m.chartView.PrevTimeframe()
				return m, nil
			}
		case tea.MouseButtonWheelDown:
			switch m.activeTab {
			case TabUniverse:
				m.universeView.MoveDown(len(m.universeView.FilteredRows(m.rows)))
				return m, nil
			case TabCharts:
				m.chartView.NextTimeframe()
				return m, nil
			}
		}

		// 3. Tab-specific mouse interactions
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch m.activeTab {
			case TabUniverse:
				selected, action := m.universeView.HandleMouseClick(msg.X, msg.Y, m.rows)
				if action == "cycle_sector" {
					sectors := []string{"All", "Information Technology", "Health Care", "Financials", "Consumer Discretionary", "Industrials", "Energy", "Materials"}
					for i, s := range sectors {
						if strings.EqualFold(s, m.universeView.SelectedCompany(m.rows).Sector) {
							next := sectors[(i+1)%len(sectors)]
							m.universeView.SetSector(next)
							break
						}
					}
					return m, nil
				} else if action == "cycle_tier" {
					tiers := []string{"All", "Mega Cap", "Large Cap", "Mid Cap", "Small Cap"}
					m.universeView.SetTier(tiers[1])
					return m, nil
				} else if selected != nil {
					m.selectedSymbol = selected.Symbol
					m.tradeView.SetSelectedSymbol(selected.Symbol, selected.Price)
					m.activeTab = TabCharts
					return m, nil
				}

			case TabCharts:
				action := m.chartView.HandleMouseClick(msg.X, msg.Y)
				switch action {
				case "open_detail":
					m.activeTab = TabDetail
					return m, nil
				case "prev_stock":
					m.selectPrevStock()
					return m, nil
				case "next_stock":
					m.selectNextStock()
					return m, nil
				}
				return m, nil

			case TabTrade:
				prices := make(map[string]float64)
				for _, r := range m.rows {
					prices[r.Symbol] = r.Price
				}
				var cash float64
				var posViews []PositionView
				if m.bridge != nil {
					acct := m.bridge.GetHumanAccount()
					if acct != nil {
						cash = acct.Cash
						for sym, pos := range acct.Positions {
							p := prices[sym]
							posViews = append(posViews, PositionView{
								Symbol:     sym,
								Side:       "LONG",
								Quantity:   pos.Quantity,
								EntryPrice: pos.EntryCost / mathMax(1, pos.Quantity),
								MarkPrice:  p,
								OpenedAt:   pos.OpenedAt,
								OpenedTime: time.Unix(0, pos.OpenedAt),
							})
						}
						// Chronological: Oldest to Newest
						sort.Slice(posViews, func(i, j int) bool {
							if posViews[i].OpenedAt != posViews[j].OpenedAt {
								return posViews[i].OpenedAt < posViews[j].OpenedAt
							}
							return posViews[i].Symbol < posViews[j].Symbol
						})
					}
				}

				action := m.tradeView.HandleMouseClick(msg.X, msg.Y, posViews, prices, cash)
				switch {
				case strings.HasPrefix(action, "select_pos:"):
					sym := strings.TrimPrefix(action, "select_pos:")
					m.selectedSymbol = sym
					return m, nil
				case strings.HasPrefix(action, "close_pos:"):
					sym := strings.TrimPrefix(action, "close_pos:")
					err := m.bridge.ClosePosition(sym)
					if err != nil {
						m.tradeView.SetStatus("Close error: "+err.Error(), true)
					} else {
						m.tradeView.SetStatus("Position closed successfully for "+sym, false)
					}
					return m, nil
				case action == "submit_order":
					return m.submitTradeOrder()
				case action == "close_selected":
					if m.selectedSymbol != "" {
						err := m.bridge.ClosePosition(m.selectedSymbol)
						if err != nil {
							m.tradeView.SetStatus("Close error: "+err.Error(), true)
						} else {
							m.tradeView.SetStatus("Position closed successfully for "+m.selectedSymbol, false)
						}
					}
					return m, nil
				case action == "reset_portfolio":
					err := m.bridge.ResetPortfolio()
					if err != nil {
						m.tradeView.SetStatus("Reset error: "+err.Error(), true)
					} else {
						m.tradeView.SetStatus("Portfolio reset: Cash restored to $100,000, all positions cleared", false)
					}
					return m, nil
				}
				return m, nil

			case TabAdmin:
				action, val, speed := m.adminView.HandleMouseClick(msg.X, msg.Y)
				switch action {
				case "set_speed":
					m.bridge.SetSpeed(speed)
					m.speedMultiplier = speed
					return m, nil
				case "save_state":
					err := m.bridge.SaveWorldState()
					if err != nil {
						m.adminView.SetStatus("Save error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus("World state saved successfully to data/world_state.json", false)
					}
					return m, nil
				case "reset_state":
					err := m.bridge.ResetWorldState()
					if err != nil {
						m.adminView.SetStatus("Reset error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus("World reset fresh! New companies generated and books cleared.", false)
						m.refreshUniverse()
					}
					return m, nil
				case "toggle_policy":
					enacted, desc := m.bridge.TogglePolicy(val)
					statusVerb := "Repealed"
					if enacted {
						statusVerb = "Enacted"
					}
					m.adminView.SetStatus(fmt.Sprintf("%s policy: %s", statusVerb, desc), false)
					return m, nil
				case "bailout":
					err := m.bridge.BailoutCompany(val)
					if err != nil {
						m.adminView.SetStatus("Bailout error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus(fmt.Sprintf("Federal bailout injected into %s ($500M capital, debt halved)", val), false)
						m.refreshUniverse()
					}
					return m, nil
				case "borrow_money":
					amt, _ := strconv.ParseFloat(val, 64)
					if err := m.bridge.BorrowMoney(amt); err != nil {
						m.adminView.SetStatus("Borrow error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus(fmt.Sprintf("Successfully borrowed %s. Treasury cash increased.", FormatCurrency(amt)), false)
					}
					return m, nil
				case "repay_debt":
					amt, _ := strconv.ParseFloat(val, 64)
					if err := m.bridge.RepayDebt(amt); err != nil {
						m.adminView.SetStatus("Repay error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus(fmt.Sprintf("Successfully repaid %s of sovereign debt.", FormatCurrency(amt)), false)
					}
					return m, nil
				case "invest_infrastructure":
					amt, _ := strconv.ParseFloat(val, 64)
					if err := m.bridge.InvestInfrastructure(amt); err != nil {
						m.adminView.SetStatus("Investment error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus(fmt.Sprintf("Invested %s in Infrastructure (Power, Transport, Ports).", FormatCurrency(amt)), false)
					}
					return m, nil
				case "invest_healthcare":
					amt, _ := strconv.ParseFloat(val, 64)
					if err := m.bridge.InvestPopulation(amt, "healthcare"); err != nil {
						m.adminView.SetStatus("Investment error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus(fmt.Sprintf("Invested %s in Public Healthcare (Boosts population growth).", FormatCurrency(amt)), false)
					}
					return m, nil
				case "invest_education":
					amt, _ := strconv.ParseFloat(val, 64)
					if err := m.bridge.InvestPopulation(amt, "education"); err != nil {
						m.adminView.SetStatus("Investment error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus(fmt.Sprintf("Invested %s in Workforce Education (Boosts productivity & wages).", FormatCurrency(amt)), false)
					}
					return m, nil
				case "invest_enterprise":
					amt, _ := strconv.ParseFloat(val, 64)
					if err := m.bridge.InvestEnterprise(amt); err != nil {
						m.adminView.SetStatus("Investment error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus(fmt.Sprintf("Injected %s in Enterprise Seed Capital (Accelerates startup formation).", FormatCurrency(amt)), false)
					}
					return m, nil
				case "invest_exports":
					amt, _ := strconv.ParseFloat(val, 64)
					if err := m.bridge.InvestInfrastructure(amt); err != nil {
						m.adminView.SetStatus("Investment error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus(fmt.Sprintf("Invested %s in Export Logistics & Shipping Ports.", FormatCurrency(amt)), false)
					}
					return m, nil
				case "charter_exchange":
					if err := m.bridge.CharterStockExchange(); err != nil {
						m.adminView.SetStatus("Charter error: "+err.Error(), true)
					} else {
						m.adminView.SetStatus("NATIONAL STOCK EXCHANGE INAUGURATED & CHARTERED! Public trading and IPO markets are now open.", false)
						m.refreshUniverse()
					}
					return m, nil
				case "launch_ipo":
					return m.launchIPOFromView()
				}
				return m, nil

			case TabWorld:
				action, val := m.worldView.HandleMouseClick(msg.X, msg.Y)
				switch action {
				case "confirm_country":
					name := strings.TrimSpace(m.worldView.SetupName)
					curr := strings.TrimSpace(m.worldView.SetupCurrency)
					flag := strings.TrimSpace(m.worldView.SetupFlag)
					fnd, _ := strconv.Atoi(strings.TrimSpace(m.worldView.SetupFounded))
					if fnd == 0 {
						fnd = 2026
					}
					if name == "" {
						m.worldView.SetStatus("Please provide a country name", true)
						return m, nil
					}
					if m.bridge != nil && m.bridge.GetWorld() != nil {
						m.bridge.GetWorld().Home.Configure(name, curr, flag, fnd)
						_ = m.bridge.SaveWorldState()
					}
					m.worldView.SetStatus(fmt.Sprintf("Sovereign nation established: %s (%s)", name, curr), false)
					return m, nil
				case "negotiate_deal":
					ok := m.bridge.NegotiateTradeDeal(val, 0.015)
					if ok {
						m.worldView.SetStatus(fmt.Sprintf("Bilateral trade pact ratified with %s (Tariff reduced to 1.5%%)", val), false)
					} else {
						m.worldView.SetStatus("Failed to negotiate trade agreement", true)
					}
					return m, nil
				case "send_aid":
					ok := m.bridge.SendForeignAid(val, 5_000_000_000.0)
					if ok {
						m.worldView.SetStatus(fmt.Sprintf("$5B foreign aid package dispatched to %s (Diplomatic standing improved)", val), false)
					} else {
						m.worldView.SetStatus("Aid dispatch failed", true)
					}
					return m, nil
				case "impose_sanction":
					ok := m.bridge.ImposeSanctions(val, 2)
					if ok {
						m.worldView.SetStatus(fmt.Sprintf("Comprehensive economic sanctions levied against %s", val), false)
					} else {
						m.worldView.SetStatus("Sanctions order failed", true)
					}
					return m, nil
				}
				return m, nil

			case TabPolicy:
				action, val, _, isCat := m.policyView.HandleMouseClick(msg.X, msg.Y)
				if isCat {
					return m, nil
				}
				if action == "toggle_policy" {
					enacted, desc := m.bridge.TogglePolicy(val)
					verb := "Repealed"
					if enacted {
						verb = "Enacted"
					}
					m.policyView.SetStatus(fmt.Sprintf("%s policy: %s", verb, desc), false)
					return m, nil
				}
				return m, nil
			}
		}
		return m, nil

	case tea.KeyMsg:
		// Command Palette Toggle (Ctrl+P)
		if msg.Type == tea.KeyCtrlP || msg.String() == "ctrl+p" || msg.String() == "ctrl+P" {
			m.palette.Toggle()
			return m, nil
		}

		// When Command Palette is open, capture all input
		if m.palette != nil && m.palette.IsOpen() {
			switch msg.Type {
			case tea.KeyEsc:
				m.palette.Close()
				return m, nil
			case tea.KeyEnter:
				cmd := m.palette.SelectedCommand()
				return m.executePaletteCommand(cmd)
			case tea.KeyUp:
				m.palette.MoveUp()
				return m, nil
			case tea.KeyDown:
				m.palette.MoveDown()
				return m, nil
			case tea.KeyBackspace:
				m.palette.BackspaceQuery()
				return m, nil
			default:
				if len(msg.Runes) > 0 {
					for _, r := range msg.Runes {
						m.palette.AppendQuery(r)
					}
					return m, nil
				}
			}
			return m, nil
		}

		// Search input mode in Universe tab
		if m.isSearching && m.activeTab == TabUniverse {
			switch msg.Type {
			case tea.KeyEnter:
				m.isSearching = false
				return m, nil
			case tea.KeyEsc:
				m.isSearching = false
				m.universeView.ClearSearch()
				return m, nil
			case tea.KeyBackspace:
				cur := m.universeView.FilteredRows(m.rows)
				_ = cur
				// Pop character
				return m, nil
			default:
				if len(msg.Runes) > 0 {
					m.universeView.SetSearch(string(msg.Runes))
				}
				return m, nil
			}
		}

		// Search input mode in Charts tab
		if m.activeTab == TabCharts && m.chartView.IsSearching() {
			switch msg.Type {
			case tea.KeyEnter:
				q := strings.ToUpper(strings.TrimSpace(m.chartView.SearchQuery()))
				if q != "" {
					for _, r := range m.rows {
						if strings.HasPrefix(strings.ToUpper(r.Symbol), q) || strings.Contains(strings.ToUpper(r.Name), q) {
							m.selectedSymbol = r.Symbol
							m.tradeView.SetSelectedSymbol(r.Symbol, r.Price)
							break
						}
					}
				}
				m.chartView.ClearSearch()
				return m, nil
			case tea.KeyEsc:
				m.chartView.ClearSearch()
				return m, nil
			case tea.KeyBackspace:
				m.chartView.PopSearchRune()
				return m, nil
			default:
				if len(msg.Runes) > 0 {
					m.chartView.AppendSearchRune(msg.Runes[0])
				}
				return m, nil
			}
		}

		if msg.String() == "ctrl+c" {
			if m.bridge != nil {
				_ = m.bridge.SaveWorldState()
			}
			return m, tea.Quit
		}

		isEditingText := false
		if m.activeTab == TabTrade && (m.tradeView.ActiveField == 0 || m.tradeView.ActiveField == 3 || m.tradeView.ActiveField == 4) {
			isEditingText = true
		}
		if m.activeTab == TabAdmin && (m.adminView.ActiveField == 0 || m.adminView.ActiveField == 1 || m.adminView.ActiveField == 4 || m.adminView.ActiveField == 5) {
			isEditingText = true
		}
		if m.activeTab == TabWorld && m.bridge != nil && m.bridge.GetWorld() != nil && !m.bridge.GetWorld().Home.IsConfigured() {
			isEditingText = true
		}

		if !isEditingText {
			if msg.String() == "q" {
				if !m.isHoldingQ {
					m.isHoldingQ = true
					m.qHoldStart = time.Now()
					m.lastQPress = time.Now()
					return m, tickQHold()
				}
				m.lastQPress = time.Now()
				if time.Since(m.qHoldStart) >= 5*time.Second {
					if m.bridge != nil {
						_ = m.bridge.SaveWorldState()
					}
					return m, tea.Quit
				}
				return m, nil
			} else {
				m.isHoldingQ = false
			}
		}

		// Global Hotkeys
		switch msg.String() {
		case "1":
			m.activeTab = TabOverview
			return m, nil
		case "2":
			m.activeTab = TabUniverse
			return m, nil
		case "3":
			m.activeTab = TabCharts
			return m, nil
		case "4":
			m.activeTab = TabDetail
			return m, nil
		case "5":
			m.activeTab = TabTrade
			return m, nil
		case "6":
			m.activeTab = TabNews
			return m, nil
		case "7":
			m.activeTab = TabAdmin
			return m, nil
		case "8":
			m.activeTab = TabWorld
			return m, nil
		case "9":
			m.activeTab = TabPolicy
			return m, nil
		case " ":
			// Toggle pause/resume
			if m.speedMultiplier > 0 {
				m.bridge.SetSpeed(0)
				m.speedMultiplier = 0
			} else {
				m.bridge.SetSpeed(1.0)
				m.speedMultiplier = 1.0
			}
			return m, nil
		case "+", "=":
			newSpeed := m.speedMultiplier * 2.0
			if newSpeed > 10.0 {
				newSpeed = 10.0
			}
			if newSpeed < 0.5 {
				newSpeed = 0.5
			}
			m.bridge.SetSpeed(newSpeed)
			m.speedMultiplier = newSpeed
			return m, nil
		case "-", "_":
			newSpeed := m.speedMultiplier / 2.0
			if newSpeed < 0.5 {
				newSpeed = 0.5
			}
			m.bridge.SetSpeed(newSpeed)
			m.speedMultiplier = newSpeed
			return m, nil
		}

		// Tab-specific key handlers
		switch m.activeTab {
		case TabUniverse:
			switch msg.String() {
			case "/":
				m.isSearching = true
				return m, nil
			case "esc":
				m.universeView.ClearSearch()
				return m, nil
			case "up", "k":
				m.universeView.MoveUp()
				return m, nil
			case "down", "j":
				m.universeView.MoveDown(len(m.universeView.FilteredRows(m.rows)))
				return m, nil
			case "pgup":
				m.universeView.PageUp()
				return m, nil
			case "pgdown":
				m.universeView.PageDown(len(m.universeView.FilteredRows(m.rows)))
				return m, nil
			case "enter":
				selected := m.universeView.SelectedCompany(m.rows)
				if selected != nil {
					m.selectedSymbol = selected.Symbol
					m.tradeView.SetSelectedSymbol(selected.Symbol, selected.Price)
					m.activeTab = TabCharts
				}
				return m, nil
			case "s":
				// Cycle sector filter
				sectors := []string{"All", "Information Technology", "Health Care", "Financials", "Consumer Discretionary", "Industrials", "Energy", "Materials"}
				for i, s := range sectors {
					if strings.EqualFold(s, m.universeView.SelectedCompany(m.rows).Sector) {
						next := sectors[(i+1)%len(sectors)]
						m.universeView.SetSector(next)
						break
					}
				}
				return m, nil
			case "t":
				// Cycle cap tier filter
				tiers := []string{"All", "Mega Cap", "Large Cap", "Mid Cap", "Small Cap"}
				next := tiers[1]
				m.universeView.SetTier(next)
				return m, nil
			}

		case TabCharts:
			switch msg.String() {
			case "/":
				m.chartView.ToggleSearch()
				return m, nil
			case "a":
				m.chartView.ToggleAreaMode()
				return m, nil
			case "t", "]":
				m.chartView.NextTimeframe()
				return m, nil
			case "[":
				m.chartView.PrevTimeframe()
				return m, nil
			case "left", "h":
				m.chartView.MoveInspectLeft()
				return m, nil
			case "right", "l":
				m.chartView.MoveInspectRight(m.width - 18)
				return m, nil
			case "esc":
				m.chartView.ClearInspect()
				return m, nil
			case "up", "k":
				m.selectPrevStock()
				return m, nil
			case "down", "j":
				m.selectNextStock()
				return m, nil
			case "d", "enter":
				m.activeTab = TabDetail
				return m, nil
			}

		case TabDetail:
			switch msg.String() {
			case "c", "3":
				m.activeTab = TabCharts
				return m, nil
			case "t", "5":
				m.activeTab = TabTrade
				return m, nil
			}

		case TabTrade:
			// If user is editing a text field (0: Symbol, 3: Shares, 4: Price)
			switch m.tradeView.ActiveField {
			case 0:
				switch msg.Type {
				case tea.KeyBackspace:
					if len(m.tradeView.OrderSymbol) > 0 {
						m.tradeView.OrderSymbol = m.tradeView.OrderSymbol[:len(m.tradeView.OrderSymbol)-1]
					}
					return m, nil
				case tea.KeyEsc:
					m.tradeView.OrderSymbol = ""
					return m, nil
				case tea.KeyEnter:
					return m.submitTradeOrder()
				case tea.KeyTab:
					m.tradeView.ActiveField = 1
					return m, nil
				default:
					if len(msg.Runes) > 0 {
						r := msg.Runes[0]
						if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
							m.tradeView.OrderSymbol += strings.ToUpper(string(r))
							return m, nil
						}
					}
				}
			case 3:
				switch msg.Type {
				case tea.KeyBackspace:
					if len(m.tradeView.OrderQty) > 0 {
						m.tradeView.OrderQty = m.tradeView.OrderQty[:len(m.tradeView.OrderQty)-1]
					}
					return m, nil
				case tea.KeyEsc:
					m.tradeView.OrderQty = "100"
					return m, nil
				case tea.KeyEnter:
					return m.submitTradeOrder()
				case tea.KeyTab:
					m.tradeView.ActiveField = 4
					return m, nil
				default:
					if len(msg.Runes) > 0 {
						r := msg.Runes[0]
						if r >= '0' && r <= '9' {
							m.tradeView.OrderQty += string(r)
							return m, nil
						}
					}
				}
			case 4:
				switch msg.Type {
				case tea.KeyBackspace:
					if len(m.tradeView.OrderPrice) > 0 {
						m.tradeView.OrderPrice = m.tradeView.OrderPrice[:len(m.tradeView.OrderPrice)-1]
					}
					return m, nil
				case tea.KeyEsc:
					m.tradeView.OrderPrice = "100.00"
					return m, nil
				case tea.KeyEnter:
					return m.submitTradeOrder()
				case tea.KeyTab:
					m.tradeView.ActiveField = 5
					return m, nil
				default:
					if len(msg.Runes) > 0 {
						r := msg.Runes[0]
						if (r >= '0' && r <= '9') || r == '.' {
							m.tradeView.OrderPrice += string(r)
							return m, nil
						}
					}
				}
			}

			switch msg.String() {
			case "tab":
				m.tradeView.ActiveField = (m.tradeView.ActiveField + 1) % 6
				return m, nil
			case "b":
				m.tradeView.OrderSide = "BUY"
				return m, nil
			case "s":
				m.tradeView.OrderSide = "SELL"
				return m, nil
			case "m":
				m.tradeView.OrderType = "MARKET"
				return m, nil
			case "l":
				m.tradeView.OrderType = "LIMIT"
				return m, nil
			case "c":
				// Close selected position
				if m.selectedSymbol != "" {
					err := m.bridge.ClosePosition(m.selectedSymbol)
					if err != nil {
						m.tradeView.SetStatus("Close error: "+err.Error(), true)
					} else {
						m.tradeView.SetStatus("Position closed successfully for "+m.selectedSymbol, false)
					}
				}
				return m, nil
			case "enter":
				return m.submitTradeOrder()
			}

		case TabAdmin:
			if msg.Type == tea.KeyEsc {
				m.adminView.ActiveField = -1
				return m, nil
			}

			if m.adminView.ActiveField == -1 {
				switch msg.String() {
				case "[", "h", "left":
					m.adminView.PrevSubTab()
					return m, nil
				case "]", "l", "right":
					m.adminView.NextSubTab()
					return m, nil
				case "tab":
					if m.adminView.ActiveSubTab == AdminSubTabIPO {
						m.adminView.ActiveField = 0
					}
					return m, nil
				}
			}

			switch m.adminView.ActiveField {
			case 0: // Symbol
				switch msg.Type {
				case tea.KeyBackspace:
					if len(m.adminView.IPOSymbol) > 0 {
						m.adminView.IPOSymbol = m.adminView.IPOSymbol[:len(m.adminView.IPOSymbol)-1]
					}
					return m, nil
				case tea.KeyEsc:
					m.adminView.ActiveField = -1
					return m, nil
				case tea.KeyTab:
					m.adminView.ActiveField = 1
					return m, nil
				case tea.KeyEnter:
					return m.launchIPOFromView()
				default:
					if len(msg.Runes) > 0 {
						r := msg.Runes[0]
						if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
							m.adminView.IPOSymbol += strings.ToUpper(string(r))
							return m, nil
						}
					}
				}
			case 1: // Company Name
				switch msg.Type {
				case tea.KeyBackspace:
					if len(m.adminView.IPOName) > 0 {
						m.adminView.IPOName = m.adminView.IPOName[:len(m.adminView.IPOName)-1]
					}
					return m, nil
				case tea.KeyEsc:
					m.adminView.ActiveField = -1
					return m, nil
				case tea.KeyTab:
					m.adminView.ActiveField = 4
					return m, nil
				case tea.KeyEnter:
					return m.launchIPOFromView()
				default:
					if len(msg.Runes) > 0 {
						m.adminView.IPOName += string(msg.Runes)
						return m, nil
					}
				}
			case 4: // Shares
				switch msg.Type {
				case tea.KeyBackspace:
					if len(m.adminView.IPOShares) > 0 {
						m.adminView.IPOShares = m.adminView.IPOShares[:len(m.adminView.IPOShares)-1]
					}
					return m, nil
				case tea.KeyEsc:
					m.adminView.ActiveField = -1
					return m, nil
				case tea.KeyTab:
					m.adminView.ActiveField = 5
					return m, nil
				case tea.KeyEnter:
					return m.launchIPOFromView()
				default:
					if len(msg.Runes) > 0 {
						r := msg.Runes[0]
						if r >= '0' && r <= '9' {
							m.adminView.IPOShares += string(r)
							return m, nil
						}
					}
				}
			case 5: // Price
				switch msg.Type {
				case tea.KeyBackspace:
					if len(m.adminView.IPOPrice) > 0 {
						m.adminView.IPOPrice = m.adminView.IPOPrice[:len(m.adminView.IPOPrice)-1]
					}
					return m, nil
				case tea.KeyEsc:
					m.adminView.ActiveField = -1
					return m, nil
				case tea.KeyTab:
					m.adminView.ActiveField = 0
					return m, nil
				case tea.KeyEnter:
					return m.launchIPOFromView()
				default:
					if len(msg.Runes) > 0 {
						r := msg.Runes[0]
						if (r >= '0' && r <= '9') || r == '.' {
							m.adminView.IPOPrice += string(r)
							return m, nil
						}
					}
				}
			}

			switch msg.String() {
			case "tab":
				m.adminView.ActiveField = (m.adminView.ActiveField + 1) % 6
				return m, nil
			case "r":
				m.adminView.RandomizeIPO()
				return m, nil
			case "s":
				m.adminView.CycleSector()
				return m, nil
			case "t":
				m.adminView.CycleTier()
				return m, nil
			case "enter":
				return m.launchIPOFromView()
			}

		case TabWorld:
			if m.bridge != nil && m.bridge.GetWorld() != nil && !m.bridge.GetWorld().Home.IsConfigured() {
				switch m.worldView.ActiveField {
				case 0: // Name
					switch msg.Type {
					case tea.KeyBackspace:
						if len(m.worldView.SetupName) > 0 {
							m.worldView.SetupName = m.worldView.SetupName[:len(m.worldView.SetupName)-1]
						}
						return m, nil
					case tea.KeyEnter:
						m.worldView.ActiveField = 1
						return m, nil
					default:
						if len(msg.Runes) > 0 {
							m.worldView.SetupName += string(msg.Runes)
							return m, nil
						}
					}
				case 1: // Currency
					switch msg.Type {
					case tea.KeyBackspace:
						if len(m.worldView.SetupCurrency) > 0 {
							m.worldView.SetupCurrency = m.worldView.SetupCurrency[:len(m.worldView.SetupCurrency)-1]
						}
						return m, nil
					case tea.KeyEnter:
						m.worldView.ActiveField = 2
						return m, nil
					default:
						if len(msg.Runes) > 0 {
							m.worldView.SetupCurrency += strings.ToUpper(string(msg.Runes))
							return m, nil
						}
					}
				case 2: // Flag
					switch msg.Type {
					case tea.KeyBackspace:
						if len(m.worldView.SetupFlag) > 0 {
							m.worldView.SetupFlag = m.worldView.SetupFlag[:len(m.worldView.SetupFlag)-1]
						}
						return m, nil
					case tea.KeyEnter:
						m.worldView.ActiveField = 3
						return m, nil
					default:
						if len(msg.Runes) > 0 && len(m.worldView.SetupFlag) < 3 {
							m.worldView.SetupFlag += strings.ToUpper(string(msg.Runes))
							return m, nil
						}
					}
				case 3: // Founded
					switch msg.Type {
					case tea.KeyBackspace:
						if len(m.worldView.SetupFounded) > 0 {
							m.worldView.SetupFounded = m.worldView.SetupFounded[:len(m.worldView.SetupFounded)-1]
						}
						return m, nil
					case tea.KeyEnter:
						name := strings.TrimSpace(m.worldView.SetupName)
						curr := strings.TrimSpace(m.worldView.SetupCurrency)
						flag := strings.TrimSpace(m.worldView.SetupFlag)
						fnd, _ := strconv.Atoi(strings.TrimSpace(m.worldView.SetupFounded))
						if fnd == 0 {
							fnd = 2026
						}
						if name == "" {
							name = "Republic of Aurora"
						}
						m.bridge.GetWorld().Home.Configure(name, curr, flag, fnd)
						_ = m.bridge.SaveWorldState()
						m.worldView.SetStatus("Sovereign nation established: "+name, false)
						return m, nil
					default:
						if len(msg.Runes) > 0 && msg.Runes[0] >= '0' && msg.Runes[0] <= '9' {
							m.worldView.SetupFounded += string(msg.Runes)
							return m, nil
						}
					}
				}
				switch msg.String() {
				case "tab":
					m.worldView.ActiveField = (m.worldView.ActiveField + 1) % 4
					return m, nil
				case "r":
					m.worldView.randomizeCountry()
					return m, nil
				case "enter":
					name := strings.TrimSpace(m.worldView.SetupName)
					curr := strings.TrimSpace(m.worldView.SetupCurrency)
					flag := strings.TrimSpace(m.worldView.SetupFlag)
					fnd, _ := strconv.Atoi(strings.TrimSpace(m.worldView.SetupFounded))
					if fnd == 0 {
						fnd = 2026
					}
					if name == "" {
						name = "Republic of Aurora"
					}
					m.bridge.GetWorld().Home.Configure(name, curr, flag, fnd)
					_ = m.bridge.SaveWorldState()
					m.worldView.SetStatus("Sovereign nation established: "+name, false)
					return m, nil
				}
			}

		case TabPolicy:
			switch msg.String() {
			case "tab", "right", "l":
				m.policyView.activeCategory = (m.policyView.activeCategory + 1) % 5
				return m, nil
			case "left", "h":
				m.policyView.activeCategory = (m.policyView.activeCategory + 4) % 5
				return m, nil
			}
		}
	}

	return m, nil
}

func (m *Model) View() string {
	if m.palette != nil && m.palette.IsOpen() {
		modal := m.palette.Render(m.width, m.height)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
	}

	header, headerHeight := m.renderTabBar()
	m.headerHeight = headerHeight

	// 2. Active Tab Content
	var content string
	switch m.activeTab {
	case TabOverview:
		content = m.overviewView.Render(m.rows, m.newsEvents, m.currentTick, m.simTime, m.speedMultiplier, m.totalVolume, m.totalTrades, m.citizenReport, m.nationalReport)

	case TabUniverse:
		content = m.universeView.Render(m.rows, m.isSearching, m.nationalReport.ExchangeChartered)

	case TabCharts:
		var selCo *CompanyRow
		for i := range m.rows {
			if m.rows[i].Symbol == m.selectedSymbol {
				selCo = &m.rows[i]
				break
			}
		}
		var candleViews []CandleView
		if m.bridge != nil && m.selectedSymbol != "" {
			candleViews = m.bridge.GetCandles(m.selectedSymbol, m.chartView.Timeframe())
		}
		content = m.chartView.Render(selCo, candleViews)

	case TabDetail:
		var selCo *CompanyRow
		for i := range m.rows {
			if m.rows[i].Symbol == m.selectedSymbol {
				selCo = &m.rows[i]
				break
			}
		}
		var bookView *OrderBookView
		var candleViews []CandleView
		if m.bridge != nil && m.selectedSymbol != "" {
			if bk := m.bridge.GetBook(m.selectedSymbol); bk != nil {
				mid, _ := bk.MidPrice()
				sp, _ := bk.Spread()
				bids, asks := bk.TopLevels(5)
				bookView = &OrderBookView{
					Symbol: m.selectedSymbol,
					Mid:    mid,
					Spread: sp,
				}
				for _, b := range bids {
					bookView.Bids = append(bookView.Bids, OrderBookLevel{Price: b.Price, Size: b.Quantity})
				}
				for _, a := range asks {
					bookView.Asks = append(bookView.Asks, OrderBookLevel{Price: a.Price, Size: a.Quantity})
				}
			}
			candleViews = m.bridge.GetCandles(m.selectedSymbol, "1m")
		}
		content = m.detailView.Render(selCo, bookView, candleViews, m.symbolTrades[m.selectedSymbol])

	case TabTrade:
		var cash, equity, marginRatio float64
		var posViews []PositionView
		if m.bridge != nil {
			acct := m.bridge.GetHumanAccount()
			if acct != nil {
				cash = acct.Cash
				prices := make(map[string]float64)
				for _, r := range m.rows {
					prices[r.Symbol] = r.Price
				}
				equity = acct.Equity(prices)
				marginRatio, _ = acct.MarginRatio(prices)

				for sym, pos := range acct.Positions {
					curPrice := prices[sym]
					val := pos.PositionValue(curPrice)
					unrealized := pos.UnrealizedPnL(curPrice)
					avgEntry := 0.0
					if pos.Quantity > 0 {
						avgEntry = pos.EntryCost / pos.Quantity
					}
					unrealizedPct := 0.0
					if pos.EntryCost > 0 {
						unrealizedPct = (unrealized / pos.EntryCost) * 100.0
					}
					sideStr := "LONG"
					if pos.Side == market.Short {
						sideStr = "SHORT"
					}
					posViews = append(posViews, PositionView{
						Symbol:        sym,
						Side:          sideStr,
						Quantity:      pos.Quantity,
						EntryPrice:    avgEntry,
						MarkPrice:     curPrice,
						TotalValue:    val,
						Unrealized:    unrealized,
						UnrealizedPct: unrealizedPct,
						OpenedAt:      pos.OpenedAt,
						OpenedTime:    time.Unix(0, pos.OpenedAt),
					})
				}

				// Sort positions strictly by time bought or sold: OLDEST to NEWEST!
				sort.Slice(posViews, func(i, j int) bool {
					if posViews[i].OpenedAt != posViews[j].OpenedAt {
						return posViews[i].OpenedAt < posViews[j].OpenedAt
					}
					return posViews[i].Symbol < posViews[j].Symbol
				})
			}
		}
		content = m.tradeView.Render(cash, equity, marginRatio, posViews)

	case TabNews:
		content = m.newsView.Render(m.newsEvents, m.tapeTrades)

	case TabAdmin:
		uptime := time.Since(m.startTime).Round(time.Second).String()
		var distressed []CompanyRow
		for _, r := range m.rows {
			if r.Price < 5.00 || r.ReportedValue < 5.00 {
				distressed = append(distressed, r)
			}
		}
		content = m.adminView.Render(
			m.currentTick,
			uptime,
			m.speedMultiplier,
			len(m.rows),
			m.totalTrades,
			m.totalVolume,
			m.nationalReport.ActivePolicies,
			distressed,
			m.nationalReport,
			headerHeight,
		)

	case TabWorld:
		if m.bridge != nil {
			content = m.worldView.Render(m.bridge.GetWorld(), m.bridge.GetSimDate())
		} else {
			content = StyleMuted.Render("  World simulation offline.")
		}

	case TabPolicy:
		content = m.policyView.Render(m.nationalReport.ActivePolicies, headerHeight)
	}

	// 3. Global Status Bar
	var statusBar string
	if m.isHoldingQ {
		elapsed := time.Since(m.qHoldStart).Seconds()
		if elapsed > 5.0 {
			elapsed = 5.0
		}
		pct := elapsed / 5.0
		barWidth := 20
		filled := int(pct * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		if filled < 0 {
			filled = 0
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		holdStatus := fmt.Sprintf("  HOLD 'q' FOR 5.0s TO EXIT: [%s]  %.1fs / 5.0s  (Release 'q' to cancel exit)  ", bar, elapsed)
		statusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#da3633")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).
			Width(m.width).
			Render(holdStatus)
	} else {
		simDateStr := ""
		if m.bridge != nil {
			simDateStr = fmt.Sprintf("Date: %s  │  ", m.bridge.GetSimDate().String())
		}
		speedStr := world.SpeedLabel(m.speedMultiplier)
		var status string
		if m.width < 90 {
			status = fmt.Sprintf(" %sTick: %d │ %s │ Cos: %d │ [Ctrl+P] Menu │ Hold [q] Exit",
				simDateStr, m.currentTick, speedStr, len(m.rows))
		} else if m.width < 135 {
			status = fmt.Sprintf(" EconomyMarkets │ %sTick: %d │ %s │ Cos: %d │ Trd: %d │ [Ctrl+P] Settings │ Hold [q] Exit",
				simDateStr, m.currentTick, speedStr, len(m.rows), m.totalTrades)
		} else {
			status = fmt.Sprintf(" EconomyMarkets TUI  │  %sTick: %d  │  Speed: %s  │  Companies: %d  │  Trades: %d  │  [Ctrl+P] Settings & Theme  │  [1-9] Tabs  │  [Space] Pause  │  Hold [q] 5s to Exit",
				simDateStr, m.currentTick, speedStr, len(m.rows), m.totalTrades)
		}
		if lipgloss.Width(status) > m.width {
			status = truncate(status, m.width)
		}
		statusBar = StyleStatusBar.Width(m.width).Render(status)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		"",
		statusBar,
	)
}

func (m *Model) executePaletteCommand(cmd *PaletteCommand) (tea.Model, tea.Cmd) {
	if cmd == nil {
		return m, nil
	}
	m.palette.Close()

	switch cmd.Action {
	case "toggle_theme":
		if CurrentTheme == ThemeLight {
			ApplyTheme(ThemeDark)
		} else {
			ApplyTheme(ThemeLight)
		}
		_ = SaveSettings("", Settings{Theme: CurrentTheme})
		return m, nil

	case "set_theme":
		if cmd.Value == "light" {
			ApplyTheme(ThemeLight)
		} else {
			ApplyTheme(ThemeDark)
		}
		_ = SaveSettings("", Settings{Theme: CurrentTheme})
		return m, nil

	case "go_tab":
		idx, _ := strconv.Atoi(cmd.Value)
		if idx >= 0 && idx <= int(TabPolicy) {
			m.activeTab = Tab(idx)
		}
		return m, nil

	case "set_speed":
		sp, _ := strconv.ParseFloat(cmd.Value, 64)
		if m.bridge != nil {
			m.bridge.SetSpeed(sp)
			m.speedMultiplier = sp
		}
		return m, nil

	case "save_state":
		if m.bridge != nil {
			_ = m.bridge.SaveWorldState()
		}
		return m, nil

	case "save_portfolio":
		if m.bridge != nil {
			_ = m.bridge.SavePortfolio()
		}
		return m, nil

	case "reset_portfolio":
		if m.bridge != nil {
			_ = m.bridge.ResetPortfolio()
		}
		return m, nil

	case "reset_world":
		if m.bridge != nil {
			_ = m.bridge.ResetWorldState()
			m.refreshUniverse()
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) renderTabBar() (string, int) {
	m.tabClickAreas = m.tabClickAreas[:0]

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

	tabPillLabels := map[Tab]string{
		TabOverview: "[1] Overview",
		TabUniverse: "[2] Stocks",
		TabCharts:   "[3] Charts",
		TabDetail:   "[4] Depth",
		TabTrade:    "[5] Trade",
		TabNews:     "[6] News",
		TabAdmin:    "[7] Sovereign",
		TabWorld:    "[8] World",
		TabPolicy:   "[9] Policy",
	}

	if m.width >= 165 {
		tabPillLabels[TabUniverse] = "[2] 500 Stocks"
		tabPillLabels[TabDetail] = "[4] Deep Dive & DOM"
		tabPillLabels[TabTrade] = "[5] Portfolio & Trade"
		tabPillLabels[TabNews] = "[6] News & Tape"
		tabPillLabels[TabAdmin] = "[7] Sovereign & Admin"
		tabPillLabels[TabWorld] = "[8] World & Nations"
		tabPillLabels[TabPolicy] = "[9] Policy Studio"
	}

	type renderedPill struct {
		tab   Tab
		text  string
		width int
	}

	var pills []renderedPill
	for _, t := range tabs {
		label := " " + tabPillLabels[t] + " "
		var pillText string
		if t == m.activeTab {
			pillText = StyleTabActive.Render(label)
		} else {
			pillText = StyleTabInactive.Render(label)
		}
		pills = append(pills, renderedPill{
			tab:   t,
			text:  pillText,
			width: lipgloss.Width(pillText),
		})
	}

	// Layout pills into rows that fit within m.width
	var rows [][]renderedPill
	var currentRow []renderedPill
	currentWidth := 0
	availWidth := m.width - 2
	if availWidth < 40 {
		availWidth = 40
	}

	for _, p := range pills {
		if len(currentRow) > 0 && currentWidth+p.width+1 > availWidth {
			rows = append(rows, currentRow)
			currentRow = nil
			currentWidth = 0
		}
		currentRow = append(currentRow, p)
		currentWidth += p.width + 1
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	var rowStrings []string
	curY := 0
	for _, row := range rows {
		var rowParts []string
		curX := 1
		for _, p := range row {
			rowParts = append(rowParts, p.text)
			m.tabClickAreas = append(m.tabClickAreas, ClickArea{
				X1:     curX,
				Y1:     curY,
				X2:     curX + p.width,
				Y2:     curY,
				Action: "switch_tab",
				Speed:  float64(p.tab),
			})
			curX += p.width + 1
		}
		rowStrings = append(rowStrings, " "+strings.Join(rowParts, " "))
		curY++
	}

	navBar := strings.Join(rowStrings, "\n")
	header := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ColorBorder).
		Width(m.width).
		Render(navBar)

	headerHeight := curY + 1
	return header, headerHeight
}

func formatEvent(ev sim.Event) string {
	switch ev.Kind {
	case sim.EventFundamental:
		return fmt.Sprintf("[%d] %s: %s shock (x%.2f)", ev.Tick, ev.Symbol, ev.FundamentalKind.String(), ev.Multiplier)
	case sim.EventRestatement:
		return fmt.Sprintf("[%d] %s: Restatement (%s, gap: %+.1f%%, severity: %.0f%%)", ev.Tick, ev.Symbol, ev.RestatementProfile.String(), ev.PriorGapPercent*100, ev.Severity*100)
	case sim.EventLiquidation:
		return fmt.Sprintf("[%d] %s: Margin liquidation of %s (%.0f shares)", ev.Tick, ev.Symbol, ev.LiquidatedAgentID, ev.LiquidationQty)
	case sim.EventIPO:
		return fmt.Sprintf("[%d] %s: Listed via IPO @ $%.2f (%.0f shares)", ev.Tick, ev.Symbol, ev.IPOPrice, ev.IPOShares)
	case sim.EventBailout:
		return fmt.Sprintf("[%d] %s: Bailout - %s", ev.Tick, ev.Symbol, ev.DistressDetails)
	case sim.EventAcquisition:
		return fmt.Sprintf("[%d] %s: Acquisition - %s", ev.Tick, ev.Symbol, ev.DistressDetails)
	case sim.EventRestructuring:
		return fmt.Sprintf("[%d] %s: Restructuring - %s", ev.Tick, ev.Symbol, ev.DistressDetails)
	case sim.EventReverseSplit:
		return fmt.Sprintf("[%d] %s: Reverse Split - %s", ev.Tick, ev.Symbol, ev.DistressDetails)
	case sim.EventChapter11:
		return fmt.Sprintf("[%d] %s: Chapter 11 - %s", ev.Tick, ev.Symbol, ev.DistressDetails)
	case sim.EventMacro:
		return fmt.Sprintf("[%d] [MACRO] %s: %s", ev.Tick, ev.Symbol, ev.MacroHeadline)
	default:
		return fmt.Sprintf("[%d] %s event", ev.Tick, ev.Symbol)
	}
}


func (m *Model) submitTradeOrder() (tea.Model, tea.Cmd) {
	sym := strings.ToUpper(strings.TrimSpace(m.tradeView.OrderSymbol))
	qty, err := strconv.ParseFloat(strings.TrimSpace(m.tradeView.OrderQty), 64)
	if err != nil || qty <= 0 {
		m.tradeView.SetStatus("Invalid shares quantity", true)
		return m, nil
	}
	price, err := strconv.ParseFloat(strings.TrimSpace(m.tradeView.OrderPrice), 64)
	if err != nil || price <= 0 {
		m.tradeView.SetStatus("Invalid order price", true)
		return m, nil
	}
	side := market.Buy
	if m.tradeView.OrderSide == "SELL" {
		side = market.Sell
	}
	isMarket := (m.tradeView.OrderType == "MARKET")

	err = m.bridge.SubmitOrder(sym, side, isMarket, qty, price)
	if err != nil {
		m.tradeView.SetStatus("Order failed: "+err.Error(), true)
	} else {
		m.tradeView.SetStatus(fmt.Sprintf("Executed: %s %.0f %s @ $%.2f", m.tradeView.OrderSide, qty, sym, price), false)
	}
	return m, nil
}

func parseAdminSector(s string) company.Sector {
	switch strings.TrimSpace(s) {
	case "Energy":
		return company.Energy
	case "Materials":
		return company.Materials
	case "Industrials":
		return company.Industrials
	case "Consumer Discretionary":
		return company.ConsumerDiscretionary
	case "Consumer Staples":
		return company.ConsumerStaples
	case "Health Care":
		return company.HealthCare
	case "Financials":
		return company.Financials
	case "Information Technology":
		return company.InformationTechnology
	case "Communication Services":
		return company.CommunicationServices
	case "Utilities":
		return company.Utilities
	case "Real Estate":
		return company.RealEstate
	default:
		return company.InformationTechnology
	}
}

func parseAdminTier(t string) company.CapTier {
	switch strings.TrimSpace(t) {
	case "Mega Cap":
		return company.MegaCap
	case "Large Cap":
		return company.LargeCap
	case "Mid Cap":
		return company.MidCap
	case "Small Cap":
		return company.SmallCap
	default:
		return company.MidCap
	}
}

func (m *Model) launchIPOFromView() (tea.Model, tea.Cmd) {
	sym := strings.ToUpper(strings.TrimSpace(m.adminView.IPOSymbol))
	name := strings.TrimSpace(m.adminView.IPOName)
	shares, err1 := strconv.ParseFloat(strings.TrimSpace(m.adminView.IPOShares), 64)
	price, err2 := strconv.ParseFloat(strings.TrimSpace(m.adminView.IPOPrice), 64)
	if sym == "" || name == "" || err1 != nil || err2 != nil || shares <= 0 || price <= 0 {
		m.adminView.SetStatus("Invalid IPO parameters", true)
		return m, nil
	}
	sec := parseAdminSector(m.adminView.IPOSector)
	tier := parseAdminTier(m.adminView.IPOTier)
	err := m.bridge.LaunchIPO(name, sym, sec, tier, shares, price)
	if err != nil {
		m.adminView.SetStatus("IPO Failed: "+err.Error(), true)
	} else {
		m.adminView.SetStatus(fmt.Sprintf("IPO Launched: %s (%s) [%s] @ $%.2f", name, sym, m.adminView.IPOSector, price), false)
		m.refreshUniverse()
	}
	return m, nil
}

func (m *Model) selectPrevStock() {
	if len(m.rows) == 0 {
		return
	}
	for i, r := range m.rows {
		if r.Symbol == m.selectedSymbol {
			prevIdx := (i - 1 + len(m.rows)) % len(m.rows)
			m.selectedSymbol = m.rows[prevIdx].Symbol
			m.tradeView.SetSelectedSymbol(m.rows[prevIdx].Symbol, m.rows[prevIdx].Price)
			return
		}
	}
	m.selectedSymbol = m.rows[0].Symbol
	m.tradeView.SetSelectedSymbol(m.rows[0].Symbol, m.rows[0].Price)
}

func (m *Model) selectNextStock() {
	if len(m.rows) == 0 {
		return
	}
	for i, r := range m.rows {
		if r.Symbol == m.selectedSymbol {
			nextIdx := (i + 1) % len(m.rows)
			m.selectedSymbol = m.rows[nextIdx].Symbol
			m.tradeView.SetSelectedSymbol(m.rows[nextIdx].Symbol, m.rows[nextIdx].Price)
			return
		}
	}
	m.selectedSymbol = m.rows[0].Symbol
	m.tradeView.SetSelectedSymbol(m.rows[0].Symbol, m.rows[0].Price)
}

