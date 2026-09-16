package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type TradeView struct {
	width          int
	height         int
	OrderSymbol    string
	OrderSide      string // "BUY" or "SELL"
	OrderType      string // "MARKET" or "LIMIT"
	OrderQty       string // e.g. "100"
	OrderPrice     string // e.g. "105.50"
	ActiveField    int    // 0: Symbol, 1: Side, 2: Type, 3: Qty, 4: Price, 5: Submit
	StatusMsg      string
	IsStatusErr    bool
	SelectedPosIdx int

	// Cached box heights for mouse hit testing
	summaryHeight int
	positionsY    int
	positionsH    int
	orderBoxY     int
	orderBoxH     int
}

func NewTradeView() *TradeView {
	return &TradeView{
		width:       120,
		height:      35,
		OrderSide:   "BUY",
		OrderType:   "MARKET",
		OrderQty:    "100",
		OrderPrice:  "100.00",
		ActiveField: 0,
	}
}

func (v *TradeView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *TradeView) SetStatus(msg string, isErr bool) {
	v.StatusMsg = msg
	v.IsStatusErr = isErr
}

func (v *TradeView) SetSelectedSymbol(sym string, price float64) {
	v.OrderSymbol = sym
	if price > 0 {
		v.OrderPrice = fmt.Sprintf("%.2f", price)
	}
}

func (v *TradeView) AdjustShares(delta float64) {
	cur, _ := strconv.ParseFloat(strings.TrimSpace(v.OrderQty), 64)
	newQty := cur + delta
	if newQty < 1 {
		newQty = 1
	}
	v.OrderQty = fmt.Sprintf("%.0f", newQty)
}

func (v *TradeView) AdjustPrice(delta float64) {
	cur, _ := strconv.ParseFloat(strings.TrimSpace(v.OrderPrice), 64)
	newPrice := cur + delta
	if newPrice < 0.01 {
		newPrice = 0.01
	}
	v.OrderPrice = fmt.Sprintf("%.2f", newPrice)
}

func (v *TradeView) SetMaxShares(cash float64) {
	p, _ := strconv.ParseFloat(strings.TrimSpace(v.OrderPrice), 64)
	if p <= 0 {
		p = 1.0
	}
	maxShares := (cash * 0.98) / p
	if maxShares < 1 {
		maxShares = 1
	}
	v.OrderQty = fmt.Sprintf("%.0f", maxShares)
}

func (v *TradeView) Render(cash, equity, marginRatio float64, positions []PositionView) string {
	var totalUnrealized float64
	for _, p := range positions {
		totalUnrealized += p.Unrealized
	}

	pnlStyle := StyleGreen
	if totalUnrealized < 0 {
		pnlStyle = StyleRed
	}

	marginStatus := "Healthy"
	marginStyle := StyleGreen
	if marginRatio < 1.5 && marginRatio > 0 {
		marginStatus = "Warning"
		marginStyle = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
	} else if marginRatio < 1.1 && marginRatio > 0 {
		marginStatus = "MARGIN CALL"
		marginStyle = StyleRed
	}

	var summaryRow string
	if v.width < 100 {
		cardW := (v.width - 8) / 2
		otherCardW := v.width - 8 - cardW
		c1 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(cardW).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Cash Balance"), StyleBold.Render(formatCurrency(cash))),
		)
		c2 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(otherCardW).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Total Equity"), StyleBold.Render(formatCurrency(equity))),
		)
		c3 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(cardW).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Unrealized PnL"), pnlStyle.Render(formatCurrency(totalUnrealized))),
		)
		c4 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(otherCardW).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s (%s)", StyleMuted.Render("Margin Ratio"), marginStyle.Render(fmt.Sprintf("%.2fx", marginRatio)), marginStyle.Render(marginStatus)),
		)
		summaryRow = lipgloss.JoinVertical(lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, c1, " ", c2), "", lipgloss.JoinHorizontal(lipgloss.Top, c3, " ", c4))
	} else {
		colWidth := max((v.width-12)/4, 20)
		c1 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(colWidth).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Cash Balance"), StyleBold.Render(formatCurrency(cash))),
		)
		c2 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(colWidth).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Total Equity"), StyleBold.Render(formatCurrency(equity))),
		)
		c3 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(colWidth).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Total Unrealized PnL"), pnlStyle.Render(formatCurrency(totalUnrealized))),
		)
		c4 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(colWidth).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s (%s)", StyleMuted.Render("Margin Ratio"), marginStyle.Render(fmt.Sprintf("%.2fx", marginRatio)), marginStyle.Render(marginStatus)),
		)
		summaryRow = lipgloss.JoinHorizontal(lipgloss.Top, c1, " ", c2, " ", c3, " ", c4)
	}
	v.summaryHeight = lipgloss.Height(summaryRow)

	var posCols []TableColumn
	if v.width < 95 {
		posCols = []TableColumn{
			{Title: "SYM", Width: 12, AlignRight: false},
			{Title: "SIDE", Width: 6, AlignRight: false},
			{Title: "SHARES", Width: 8, AlignRight: true},
			{Title: "ENTRY", Width: 9, AlignRight: true},
			{Title: "MARK", Width: 9, AlignRight: true},
			{Title: "UNREALIZED", Width: 16, AlignRight: true},
			{Title: "ACTION", Width: 8, AlignRight: false},
		}
	} else {
		posCols = []TableColumn{
			{Title: "SYM", Width: 12, AlignRight: false},
			{Title: "SIDE", Width: 6, AlignRight: false},
			{Title: "SHARES", Width: 10, AlignRight: true},
			{Title: "ENTRY", Width: 10, AlignRight: true},
			{Title: "MARK", Width: 10, AlignRight: true},
			{Title: "VALUE", Width: 12, AlignRight: true},
			{Title: "UNREALIZED", Width: 18, AlignRight: true},
			{Title: "OPENED", Width: 10, AlignRight: false},
			{Title: "ACTION", Width: 8, AlignRight: false},
		}
	}

	var posLines []string
	posLines = append(posLines, StyleHeader.Render(FormatTableHeader(posCols)))

	if len(positions) == 0 {
		posLines = append(posLines, lipgloss.NewStyle().Foreground(ColorMuted).Padding(1, 2).Render("No open positions. Use the order form to enter trades."))
	} else {
		for i, p := range positions {
			sideStyle := StyleGreen
			if p.Side == "SHORT" {
				sideStyle = StyleRed
			}
			pnlStr := fmt.Sprintf("%+9.2f (%+.1f%%)", p.Unrealized, p.UnrealizedPct)
			posPnlStyle := StyleGreen
			if p.Unrealized < 0 {
				posPnlStyle = StyleRed
			}
			itemStyle := StyleTableRow
			if i == v.SelectedPosIdx {
				itemStyle = StyleTableRowSelected
			}

			openedStr := "-"
			if p.OpenedAt > 0 {
				openedStr = p.OpenedTime.Format("15:04:05")
			}

			closeBtnStyle := lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)

			var values []string
			var styles []lipgloss.Style
			if v.width < 95 {
				values = []string{
					p.Symbol,
					p.Side,
					fmt.Sprintf("%.0f", p.Quantity),
					fmt.Sprintf("%.2f", p.EntryPrice),
					fmt.Sprintf("%.2f", p.MarkPrice),
					pnlStr,
					"[Close]",
				}
				styles = []lipgloss.Style{
					lipgloss.NewStyle(),
					sideStyle,
					lipgloss.NewStyle(),
					lipgloss.NewStyle(),
					lipgloss.NewStyle(),
					posPnlStyle,
					closeBtnStyle,
				}
			} else {
				values = []string{
					p.Symbol,
					p.Side,
					fmt.Sprintf("%.0f", p.Quantity),
					fmt.Sprintf("%.2f", p.EntryPrice),
					fmt.Sprintf("%.2f", p.MarkPrice),
					fmt.Sprintf("%.2f", p.TotalValue),
					pnlStr,
					openedStr,
					"[Close]",
				}
				styles = []lipgloss.Style{
					lipgloss.NewStyle(),
					sideStyle,
					lipgloss.NewStyle(),
					lipgloss.NewStyle(),
					lipgloss.NewStyle(),
					lipgloss.NewStyle(),
					posPnlStyle,
					lipgloss.NewStyle(),
					closeBtnStyle,
				}
			}
			rowContent := FormatTableRow(posCols, values, styles)
			posLines = append(posLines, itemStyle.Width(v.width-6).Render(rowContent))
		}
	}

	positionsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(v.width - 4).
		Render(
			fmt.Sprintf("%s\n%s",
				StyleBold.Render("  OPEN PORTFOLIO POSITIONS (Sorted by order time: Oldest to Newest)"),
				strings.Join(posLines, "\n"),
			),
		)
	v.positionsY = 2 + v.summaryHeight + 1
	v.positionsH = lipgloss.Height(positionsBox)

	// Interactive Button Helpers
	btnStyle := func(active bool, activeBg lipgloss.TerminalColor) lipgloss.Style {
		if active {
			return lipgloss.NewStyle().Background(activeBg).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1)
		}
		return lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorMuted).Padding(0, 1)
	}

	inputFieldStyle := func(fieldIdx int) lipgloss.Style {
		if v.ActiveField == fieldIdx {
			return lipgloss.NewStyle().Background(ColorHighlight).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1)
		}
		return lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(lipgloss.Color("#c9d1d9")).Padding(0, 1)
	}

	// Line 1: Symbol, Side, Type
	symLabel := StyleBold.Render("Symbol:")
	symInput := inputFieldStyle(0).Render(" " + v.OrderSymbol + " ")

	sideLabel := StyleBold.Render("Action:")
	buyBtn := btnStyle(v.OrderSide == "BUY", ColorSuccess).Render("[ BUY ]")
	sellBtn := btnStyle(v.OrderSide == "SELL", ColorDanger).Render("[ SELL ]")

	typeLabel := StyleBold.Render("Type:")
	mktBtn := btnStyle(v.OrderType == "MARKET", ColorHighlight).Render("[ MARKET ]")
	lmtBtn := btnStyle(v.OrderType == "LIMIT", ColorHighlight).Render("[ LIMIT ]")

	row1 := fmt.Sprintf("  %s %s     %s %s %s     %s %s %s",
		symLabel, symInput, sideLabel, buyBtn, sellBtn, typeLabel, mktBtn, lmtBtn)

	// Line 2: Shares and Quick Presets
	qtyLabel := StyleBold.Render("Shares:")
	minusSharesBtn := btnStyle(false, ColorCardBg).Render("[-]")
	qtyInput := inputFieldStyle(3).Render(" " + v.OrderQty + " ")
	plusSharesBtn := btnStyle(false, ColorCardBg).Render("[+]")

	presetsLabel := StyleMuted.Render("Presets:")
	p10 := btnStyle(v.OrderQty == "10", ColorCardBg).Render("[ 10 ]")
	p50 := btnStyle(v.OrderQty == "50", ColorCardBg).Render("[ 50 ]")
	p100 := btnStyle(v.OrderQty == "100", ColorCardBg).Render("[ 100 ]")
	p500 := btnStyle(v.OrderQty == "500", ColorCardBg).Render("[ 500 ]")
	p1k := btnStyle(v.OrderQty == "1000", ColorCardBg).Render("[ 1K ]")
	pMax := btnStyle(false, ColorCardBg).Render("[ MAX ]")

	row2 := fmt.Sprintf("  %s %s %s %s     %s %s %s %s %s %s %s",
		qtyLabel, minusSharesBtn, qtyInput, plusSharesBtn, presetsLabel, p10, p50, p100, p500, p1k, pMax)

	// Line 3: Price
	priceLabel := StyleBold.Render("Price: ")
	minusPriceBtn := btnStyle(false, ColorCardBg).Render("[-0.10]")
	priceInput := inputFieldStyle(4).Render(" " + v.OrderPrice + " ")
	plusPriceBtn := btnStyle(false, ColorCardBg).Render("[+0.10]")
	markBtn := btnStyle(false, ColorCardBg).Render("[ Use Mark Price ]")

	row3 := fmt.Sprintf("  %s %s %s %s     %s",
		priceLabel, minusPriceBtn, priceInput, plusPriceBtn, markBtn)

	// Line 4: Action Buttons
	submitColor := ColorSuccess
	if v.OrderSide == "SELL" {
		submitColor = ColorDanger
	}
	submitBtn := lipgloss.NewStyle().
		Background(submitColor).
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true).
		Padding(0, 3).
		Render(fmt.Sprintf("[ SUBMIT %s ORDER ]", v.OrderSide))

	closePosBtn := lipgloss.NewStyle().
		Background(lipgloss.Color("#30363d")).
		Foreground(ColorDanger).
		Bold(true).
		Padding(0, 2).
		Render("[ CLOSE SELECTED POSITION ]")

	resetBtn := lipgloss.NewStyle().
		Background(lipgloss.Color("#490206")).
		Foreground(lipgloss.Color("#ff7b72")).
		Bold(true).
		Padding(0, 2).
		Render("[ RESET PORTFOLIO ]")

	row4 := fmt.Sprintf("  %s      %s      %s", submitBtn, closePosBtn, resetBtn)

	statusText := ""
	if v.StatusMsg != "" {
		if v.IsStatusErr {
			statusText = "  " + StyleRed.Render("[ERROR] "+v.StatusMsg) + "\n\n"
		} else {
			statusText = "  " + StyleGreen.Render("[OK] "+v.StatusMsg) + "\n\n"
		}
	}

	footerText := StyleMuted.Render("  Mouse Controls: Click any button, preset, or position row  •  Keyboard: [Tab] Field  •  [Enter] Submit  •  [c] Close")

	orderBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Width(v.width - 4).
		Padding(1, 1).
		Render(
			fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s%s",
				StyleTitle.Render("RAPID ORDER ENTRY (Clickable with Mouse)"),
				row1,
				row2,
				row3,
				row4,
				statusText,
				footerText,
			),
		)

	v.orderBoxY = v.positionsY + v.positionsH + 1
	v.orderBoxH = lipgloss.Height(orderBox)

	return lipgloss.JoinVertical(lipgloss.Left,
		summaryRow,
		"",
		positionsBox,
		"",
		orderBox,
	)
}

// HandleMouseClick handles a mouse click inside the Trade tab.
// Returns an action string if a button/row was tapped.
func (v *TradeView) HandleMouseClick(x, y int, positions []PositionView, markPrices map[string]float64, cash float64) string {
	// 1. Check if click is inside Open Positions table
	if y >= v.positionsY && y < v.positionsY+v.positionsH {
		// Inside positions box
		// Top border is line 0, title is line 1, table header is line 2
		rowIdx := y - (v.positionsY + 3)
		if rowIdx >= 0 && rowIdx < len(positions) {
			p := positions[rowIdx]
			// Check if clicked the [Close] button on the right edge
			if x >= v.width-22 {
				return "close_pos:" + p.Symbol
			}
			// Clicked the row to select position
			v.SelectedPosIdx = rowIdx
			v.SetSelectedSymbol(p.Symbol, p.MarkPrice)
			return "select_pos:" + p.Symbol
		}
		return ""
	}

	// 2. Check if click is inside Order Entry Box
	if y >= v.orderBoxY && y < v.orderBoxY+v.orderBoxH {
		relY := y - v.orderBoxY
		switch {
		case relY >= 3 && relY <= 4:
			// Row 1: Symbol, Side, Type
			if x >= 2 && x <= 20 {
				v.ActiveField = 0 // Focus Symbol
				return "focus_symbol"
			}
			if x >= 22 && x <= 36 {
				v.OrderSide = "BUY"
				return "set_side:BUY"
			}
			if x >= 37 && x <= 50 {
				v.OrderSide = "SELL"
				return "set_side:SELL"
			}
			if x >= 52 && x <= 72 {
				v.OrderType = "MARKET"
				return "set_type:MARKET"
			}
			if x >= 73 && x <= 90 {
				v.OrderType = "LIMIT"
				return "set_type:LIMIT"
			}

		case relY >= 5 && relY <= 6:
			// Row 2: Shares, Presets
			if x >= 10 && x <= 16 {
				v.AdjustShares(-50)
				return "adj_shares:-50"
			}
			if x >= 17 && x <= 28 {
				v.ActiveField = 3 // Focus Shares
				return "focus_shares"
			}
			if x >= 29 && x <= 35 {
				v.AdjustShares(+50)
				return "adj_shares:+50"
			}
			// Presets
			if x >= 46 && x <= 54 {
				v.OrderQty = "10"
				return "set_shares:10"
			}
			if x >= 55 && x <= 64 {
				v.OrderQty = "50"
				return "set_shares:50"
			}
			if x >= 65 && x <= 75 {
				v.OrderQty = "100"
				return "set_shares:100"
			}
			if x >= 76 && x <= 86 {
				v.OrderQty = "500"
				return "set_shares:500"
			}
			if x >= 87 && x <= 96 {
				v.OrderQty = "1000"
				return "set_shares:1000"
			}
			if x >= 97 && x <= 110 {
				v.SetMaxShares(cash)
				return "set_shares:max"
			}

		case relY >= 7 && relY <= 8:
			// Row 3: Price
			if x >= 10 && x <= 20 {
				v.AdjustPrice(-0.10)
				return "adj_price:-0.10"
			}
			if x >= 21 && x <= 34 {
				v.ActiveField = 4 // Focus Price
				return "focus_price"
			}
			if x >= 35 && x <= 46 {
				v.AdjustPrice(+0.10)
				return "adj_price:+0.10"
			}
			if x >= 48 && x <= 75 {
				if mk, ok := markPrices[v.OrderSymbol]; ok && mk > 0 {
					v.OrderPrice = fmt.Sprintf("%.2f", mk)
				}
				return "set_price:mark"
			}

		case relY >= 9 && relY <= 11:
			// Row 4: Action Buttons
			if x >= 2 && x <= 35 {
				return "submit_order"
			}
			if x >= 36 && x <= 75 {
				return "close_selected"
			}
			if x >= 76 && x <= 110 {
				return "reset_portfolio"
			}
		}
	}

	return ""
}
