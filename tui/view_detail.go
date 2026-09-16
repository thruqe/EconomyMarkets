package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type DetailView struct {
	width  int
	height int
}

func NewDetailView() *DetailView {
	return &DetailView{width: 120, height: 35}
}

func (v *DetailView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *DetailView) Render(company *CompanyRow, book *OrderBookView, candles []CandleView, trades []string) string {
	if company == nil {
		return lipgloss.NewStyle().Foreground(ColorMuted).Padding(2, 4).Render(
			"No company selected. Please return to [2. 500 Stocks] and press Enter on a company.",
		)
	}

	chgStr := fmt.Sprintf("%+6.2f%%", company.ChangePct)
	chgStyled := StyleGreen.Render(chgStr)
	if company.ChangePct < 0 {
		chgStyled = StyleRed.Render(chgStr)
	}

	headerText := fmt.Sprintf(" %s - %s  │  %s  │  %s  │  Price: $%.2f %s  │  Spread: $%.2f",
		StyleTitle.Render(company.Symbol),
		company.Name,
		company.Sector,
		company.CapTier,
		company.Price,
		chgStyled,
		company.Spread,
	)
	banner := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Width(v.width - 4).
		Render(headerText)

	valDiff := company.ReportedValue - company.TrueValue
	valDiffPct := 0.0
	if company.TrueValue > 0 {
		valDiffPct = (valDiff / company.TrueValue) * 100.0
	}
	biasStr := fmt.Sprintf("%+.2f%% gap", valDiffPct)
	if math.Abs(valDiffPct) < 0.5 {
		biasStr = "Fairly Stated"
	} else if valDiffPct > 0 {
		biasStr = fmt.Sprintf("Optimistic (%+.1f%%)", valDiffPct)
	} else {
		biasStr = fmt.Sprintf("Conservative (%+.1f%%)", valDiffPct)
	}

	leftWidth := (v.width - 8) / 2
	rightWidth := v.width - 8 - leftWidth
	if v.width < 95 {
		leftWidth = v.width - 4
		rightWidth = v.width - 4
	}

	fundamentalsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(leftWidth).
		Padding(0, 1).
		Render(
			fmt.Sprintf("%s\n\n"+
				" %-22s : %s\n"+
				" %-22s : %s\n"+
				" %-22s : $%.2f\n"+
				" %-22s : $%.2f (%s)\n"+
				" %-22s : %.2fx\n"+
				" %-22s : %.2fx\n"+
				" %-22s : $%.2f / $%.2f",
				StyleBold.Render("FUNDAMENTAL VALUATION METRICS"),
				"Market Capitalization", formatCurrency(company.MarketCap),
				"Reporting Profile", biasStr,
				"Enterprise True Value", company.TrueValue,
				"Reported Book Value", company.ReportedValue, biasStr,
				"Price to Earnings (P/E)", company.PE,
				"Price to Sales (P/S)", company.PS,
				"Current Bid / Ask", company.Bid, company.Ask,
			),
		)

	var domLines []string
	domCols := []TableColumn{
		{Title: "SIDE", Width: 6, AlignRight: false},
		{Title: "PRICE", Width: 10, AlignRight: true},
		{Title: "SIZE", Width: 10, AlignRight: true},
		{Title: "DEPTH", Width: 16, AlignRight: false},
	}
	domLines = append(domLines, StyleHeader.Render(FormatTableHeader(domCols)))

	if book != nil && (len(book.Asks) > 0 || len(book.Bids) > 0) {
		for i := len(book.Asks) - 1; i >= 0; i-- {
			ask := book.Asks[i]
			barLen := int(math.Min(16, ask.Size/50))
			bar := strings.Repeat("█", barLen)
			values := []string{"ASK", fmt.Sprintf("%.2f", ask.Price), fmt.Sprintf("%.0f", ask.Size), bar}
			styles := []lipgloss.Style{StyleRed, lipgloss.NewStyle(), lipgloss.NewStyle(), StyleRed}
			domLines = append(domLines, FormatTableRow(domCols, values, styles))
		}

		domLines = append(domLines, StyleMuted.Render(fmt.Sprintf(" ────── SPREAD: $%.2f ──────", book.Spread)))

		for _, bid := range book.Bids {
			barLen := int(math.Min(16, bid.Size/50))
			bar := strings.Repeat("█", barLen)
			values := []string{"BID", fmt.Sprintf("%.2f", bid.Price), fmt.Sprintf("%.0f", bid.Size), bar}
			styles := []lipgloss.Style{StyleGreen, lipgloss.NewStyle(), lipgloss.NewStyle(), StyleGreen}
			domLines = append(domLines, FormatTableRow(domCols, values, styles))
		}
	} else {
		domLines = append(domLines, "  Waiting for order book depth...")
	}

	domBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(rightWidth).
		Render(
			fmt.Sprintf("%s\n%s",
				StyleBold.Render("  ORDER BOOK DEPTH OF MARKET (DOM)"),
				strings.Join(domLines, "\n"),
			),
		)

	var topRow string
	if v.width < 95 {
		topRow = lipgloss.JoinVertical(lipgloss.Left, fundamentalsBox, "", domBox)
	} else {
		topRow = lipgloss.JoinHorizontal(lipgloss.Top, fundamentalsBox, " ", domBox)
	}

	var macroText string
	if v.width < 95 {
		macroText = fmt.Sprintf(
			" %-22s : %s\n"+
				" %-22s : %s\n"+
				" %-22s : %s\n"+
				" %-22s : %s\n"+
				" %-22s : %s\n"+
				" %-22s : %.2fx",
			"Headcount", fmt.Sprintf("%s employees", FormatMetric(company.Headcount, false)),
			"Debt Outstanding", FormatCurrency(company.DebtOutstanding),
			"Annual Payroll", FormatCurrency(company.LaborExpense),
			"Interest Burden", FormatCurrency(company.InterestExpense),
			"Federal Taxes (IRS)", FormatCurrency(company.CorporateTaxPaid),
			"Demand Multiplier", company.MacroDemandFactor,
		)
	} else {
		macroText = fmt.Sprintf(
			" %-24s : %-18s   %-24s : %s\n"+
				" %-24s : %-18s   %-24s : %s\n"+
				" %-24s : %-18s   %-24s : %.2fx (Consumer & Trade)",
			"Total Headcount", fmt.Sprintf("%s employees", FormatMetric(company.Headcount, false)),
			"Total Debt Outstanding", FormatCurrency(company.DebtOutstanding),
			"Annual Payroll Expense", FormatCurrency(company.LaborExpense),
			"Annual Interest Burden", FormatCurrency(company.InterestExpense),
			"Federal Taxes (IRS)", FormatCurrency(company.CorporateTaxPaid),
			"Macro Demand Multiplier", company.MacroDemandFactor,
		)
	}

	macroBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(v.width - 4).
		Padding(0, 1).
		Render(
			fmt.Sprintf("%s\n\n%s",
				StyleBold.Render("CORPORATE MACROECONOMIC & LABOR FOOTPRINT"),
				macroText,
			),
		)

	chartHint := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(v.width - 4).
		Padding(1, 2).
		Render(
			fmt.Sprintf("%s\n%s",
				StyleBold.Render("INTERACTIVE TIME & PRICE CHARTS"),
				StyleMuted.Render("Press '3' to view the dedicated high-resolution multi-timeframe line chart for "+company.Symbol),
			),
		)

	tradesText := "No trades recorded for this symbol yet."
	if len(trades) > 0 {
		var trLines []string
		start := max(0, len(trades)-5)
		for i := len(trades) - 1; i >= start; i-- {
			trLines = append(trLines, "  • "+trades[i])
		}
		tradesText = strings.Join(trLines, "\n")
	}
	tradesBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(v.width - 4).
		Render(
			fmt.Sprintf("%s\n%s",
				StyleBold.Render("  RECENT EXECUTED TRADES"),
				tradesText,
			),
		)

	return lipgloss.JoinVertical(lipgloss.Left,
		banner,
		topRow,
		macroBox,
		chartHint,
		tradesBox,
	)
}
