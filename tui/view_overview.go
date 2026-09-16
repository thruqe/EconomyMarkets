package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"economy/citizen"
	"economy/country"
)

type OverviewView struct {
	width  int
	height int
}

func NewOverviewView() *OverviewView {
	return &OverviewView{width: 120, height: 35}
}

func (v *OverviewView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func renderProgressBar(val float64, maxVal float64, width int) string {
	pct := math.Max(0, math.Min(1.0, val/maxVal))
	filled := int(math.Round(pct * float64(width)))
	empty := width - filled
	if empty < 0 {
		empty = 0
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func (v *OverviewView) Render(
	rows []CompanyRow,
	events []string,
	tick int64,
	simTime int64,
	speed float64,
	totalVol float64,
	totalTrades int64,
	citRep citizen.CitizenReport,
	natRep country.NationalReport,
) string {
	if len(rows) == 0 {
		return lipgloss.NewStyle().Foreground(ColorMuted).Render("No market data available yet...")
	}

	halfWidth := (v.width - 6) / 2
	otherHalfWidth := v.width - 6 - halfWidth

	// =========================================================================
	// MODE A: UNCHARTERED NATION BUILDING (Pre-Market Emerging Economy)
	// =========================================================================
	if !natRep.ExchangeChartered {
		debtToGDP := 0.0
		if natRep.GDP > 0 {
			debtToGDP = (natRep.NationalDebt / natRep.GDP) * 100.0
		}

		var sovereignCards string
		if v.width < 100 {
			cardW := (v.width - 8) / 2
			otherCardW := v.width - 8 - cardW
			c1 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorSuccess).Width(cardW).Padding(0, 1).Render(
				fmt.Sprintf("%s\n%s", StyleMuted.Render("Treasury Cash"), StyleGreen.Bold(true).Render(formatCurrency(natRep.TreasuryCash))),
			)
			c2 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorWarning).Width(otherCardW).Padding(0, 1).Render(
				fmt.Sprintf("%s\n%s (%s)", StyleMuted.Render("National Debt"), StyleBold.Render(formatCurrency(natRep.NationalDebt)), StyleMuted.Render(fmt.Sprintf("%.1f%% GDP", debtToGDP))),
			)
			c3 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorPrimary).Width(cardW).Padding(0, 1).Render(
				fmt.Sprintf("%s\n%s (Yield: %.2f%%)", StyleMuted.Render("Rating & Yield"), StyleCyan.Bold(true).Render(natRep.CreditRating), natRep.BorrowingYield*100),
			)
			c4 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorPrimary).Width(otherCardW).Padding(0, 1).Render(
				fmt.Sprintf("%s\n%s", StyleMuted.Render("Exchange Status"), lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render("[ UNCHARTERED ]")),
			)
			sovereignCards = lipgloss.JoinVertical(lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, c1, " ", c2), "", lipgloss.JoinHorizontal(lipgloss.Top, c3, " ", c4))
		} else {
			colWidth := max((v.width-12)/4, 20)
			c1 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorSuccess).Width(colWidth).Padding(0, 1).Render(
				fmt.Sprintf("%s\n%s", StyleMuted.Render("Sovereign Treasury Cash"), StyleGreen.Bold(true).Render(formatCurrency(natRep.TreasuryCash))),
			)
			c2 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorWarning).Width(colWidth).Padding(0, 1).Render(
				fmt.Sprintf("%s\n%s (%s)", StyleMuted.Render("Sovereign National Debt"), StyleBold.Render(formatCurrency(natRep.NationalDebt)), StyleMuted.Render(fmt.Sprintf("%.1f%% GDP", debtToGDP))),
			)
			c3 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorPrimary).Width(colWidth).Padding(0, 1).Render(
				fmt.Sprintf("%s\n%s (Yield: %.2f%%)", StyleMuted.Render("Credit Rating & Yield"), StyleCyan.Bold(true).Render(natRep.CreditRating), natRep.BorrowingYield*100),
			)
			c4 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorPrimary).Width(colWidth).Padding(0, 1).Render(
				fmt.Sprintf("%s\n%s", StyleMuted.Render("Stock Exchange Status"), lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render("[ UNCHARTERED ]")),
			)
			sovereignCards = lipgloss.JoinHorizontal(lipgloss.Top, c1, " ", c2, " ", c3, " ", c4)
		}

		leftTreasuryLines := fmt.Sprintf(
			" %-24s : %s (%s)\n"+
				" %-24s : %s\n"+
				" %-24s : %s (%.1f%% of GDP)\n"+
				" %-24s : %s / yr\n"+
				" %-24s : %s / yr\n"+
				" %-24s : %s / yr\n"+
				" %-24s : %.2f%% [%s Rating]",
			"Gross Domestic Product", FormatCurrency(natRep.GDP), fmt.Sprintf("%+.1f%% real", natRep.RealGDPGrowth*100),
			"Liquid Sovereign Treasury", FormatCurrency(natRep.TreasuryCash),
			"National Sovereign Debt", FormatCurrency(natRep.NationalDebt), debtToGDP,
			"Federal Tax Revenues", FormatCurrency(natRep.FederalRevenue),
			"Sovereign Export Revenue", FormatCurrency(natRep.ExportRevenue),
			"Foreign Direct Investment", FormatCurrency(natRep.FDIInflow),
			"Bond Borrowing Yield", natRep.BorrowingYield*100, natRep.CreditRating,
		)

		rightCapitalLines := fmt.Sprintf(
			" %-24s : %s  │  Labor Force: %s\n"+
				" %-24s : %.2f%%  │  Avg Wage: $%.2f/hr\n"+
				" %-24s : %s %4.1f/100\n"+
				" %-24s : %s %4.1f/100\n"+
				" %-24s : %s %4.1f/100\n"+
				" %-24s : %s %4.1f/100\n"+
				" %-24s : %s %4.1f/100",
			"Citizen Population", FormatMetric(citRep.Population, false), FormatMetric(natRep.LaborForce, false),
			"Civilian Unemployment", natRep.UnemploymentRate*100, natRep.AverageHourlyWage,
			"Infrastructure (Power/Logistics)", renderProgressBar(natRep.InfrastructureLevel, 100, 10), natRep.InfrastructureLevel,
			"Healthcare & Hospitals", renderProgressBar(natRep.HealthcareLevel, 100, 10), natRep.HealthcareLevel,
			"Workforce Education", renderProgressBar(natRep.EducationLevel, 100, 10), natRep.EducationLevel,
			"Enterprise Seed Grants", renderProgressBar(natRep.EnterpriseGrantsLevel, 100, 10), natRep.EnterpriseGrantsLevel,
			"Export Ports & Terminals", renderProgressBar(natRep.ExportCapacityLevel, 100, 10), natRep.ExportCapacityLevel,
		)

		var middleRow string
		if v.width < 95 {
			boxLeft := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Width(v.width - 4).
				Padding(0, 1).
				Render(fmt.Sprintf("%s\n%s", StyleBold.Render("SOVEREIGN TREASURY & MACRO INFLOWS"), leftTreasuryLines))

			boxRight := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess).
				Width(v.width - 4).
				Padding(0, 1).
				Render(fmt.Sprintf("%s\n%s", StyleBold.Render("NATIONAL INFRASTRUCTURE & HUMAN CAPITAL"), rightCapitalLines))
			middleRow = lipgloss.JoinVertical(lipgloss.Left, boxLeft, "", boxRight)
		} else {
			boxLeft := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Width(halfWidth).
				Padding(0, 1).
				Render(fmt.Sprintf("%s\n%s", StyleBold.Render("SOVEREIGN TREASURY & MACRO INFLOWS"), leftTreasuryLines))

			boxRight := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess).
				Width(otherHalfWidth).
				Padding(0, 1).
				Render(fmt.Sprintf("%s\n%s", StyleBold.Render("NATIONAL INFRASTRUCTURE & HUMAN CAPITAL"), rightCapitalLines))
			middleRow = lipgloss.JoinHorizontal(lipgloss.Top, boxLeft, " ", boxRight)
		}

		privateCount := min(6, len(rows))
		if v.height < 32 {
			privateCount = min(3, len(rows))
		}
		var privateCols []TableColumn
		if v.width < 85 {
			privateCols = []TableColumn{
				{Title: "SYM", Width: 6, AlignRight: false},
				{Title: "NAME", Width: 16, AlignRight: false},
				{Title: "STAGE", Width: 10, AlignRight: false},
				{Title: "VALUATION", Width: 12, AlignRight: true},
			}
		} else {
			privateCols = []TableColumn{
				{Title: "SYM", Width: 6, AlignRight: false},
				{Title: "NAME", Width: 18, AlignRight: false},
				{Title: "SECTOR", Width: 16, AlignRight: false},
				{Title: "STAGE", Width: 10, AlignRight: false},
				{Title: "VALUATION", Width: 12, AlignRight: true},
				{Title: "REVENUE", Width: 12, AlignRight: true},
			}
		}

		privateLines := []string{
			StyleHeader.Render(FormatTableHeader(privateCols)),
		}
		for i := 0; i < privateCount; i++ {
			r := rows[i]
			stage := r.Stage
			if stage == "" {
				stage = "Seed"
			}
			val := r.PrivateValuation
			if val <= 0 {
				val = r.ReportedValue
			}
			stageStyle := StyleMuted
			if stage == "Pre-IPO" {
				stageStyle = StyleGreen.Bold(true)
			} else if stage == "Growth" {
				stageStyle = StyleCyan
			}

			var values []string
			var styles []lipgloss.Style
			if v.width < 85 {
				values = []string{r.Symbol, r.Name, stage, FormatCurrency(val)}
				styles = []lipgloss.Style{StyleBold, lipgloss.NewStyle(), stageStyle, lipgloss.NewStyle()}
			} else {
				values = []string{r.Symbol, r.Name, r.Sector, stage, FormatCurrency(val), FormatCurrency(val * 0.25)}
				styles = []lipgloss.Style{StyleBold, lipgloss.NewStyle(), lipgloss.NewStyle(), stageStyle, lipgloss.NewStyle(), lipgloss.NewStyle()}
			}
			privateLines = append(privateLines, FormatTableRow(privateCols, values, styles))
		}

		infraPass := natRep.InfrastructureLevel >= 25.0
		grantsPass := natRep.EnterpriseGrantsLevel >= 20.0
		cashPass := natRep.TreasuryCash >= 5_000_000_000.0
		cosPass := len(rows) >= 3

		passTag := func(ok bool) string {
			if ok {
				return lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render("[ PASS ]")
			}
			return lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("[ REQ'D ]")
		}

		roadmapLines := []string{
			"  To establish public trading, the sovereign government must charter the exchange:",
			"",
			fmt.Sprintf("  %s Infrastructure Index >= 25.0  (Current: %.1f)", passTag(infraPass), natRep.InfrastructureLevel),
			fmt.Sprintf("  %s Enterprise Grants Level >= 20.0 (Current: %.1f)", passTag(grantsPass), natRep.EnterpriseGrantsLevel),
			fmt.Sprintf("  %s Liquid Treasury Cash >= $5.00B (Current: %s)", passTag(cashPass), FormatCurrency(natRep.TreasuryCash)),
			fmt.Sprintf("  %s Domestic Enterprises >= 3    (Current: %d)", passTag(cosPass), len(rows)),
			"",
			StyleCyan.Render("  -> Action: Go to Administration [Tab 7] to borrow sovereign capital,"),
			StyleCyan.Render("     invest in Infrastructure & Grants, and Inaugurate the Exchange!"),
		}

		var bottomRow string
		if v.width < 95 {
			privateBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Width(v.width - 4).
				Render(
					fmt.Sprintf("%s\n%s",
						StyleBold.Render("  DOMESTIC PRIVATE ENTERPRISE REGISTRY"),
						strings.Join(privateLines, "\n"),
					),
				)
			roadmapBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSecondary).
				Width(v.width - 4).
				Render(
					fmt.Sprintf("%s\n%s",
						StyleTitle.Render("  STOCK EXCHANGE CHARTER ROADMAP"),
						strings.Join(roadmapLines, "\n"),
					),
				)
			bottomRow = lipgloss.JoinVertical(lipgloss.Left, privateBox, "", roadmapBox)
		} else {
			privateBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Width(halfWidth).
				Render(
					fmt.Sprintf("%s\n%s",
						StyleBold.Render("  DOMESTIC PRIVATE ENTERPRISE REGISTRY"),
						strings.Join(privateLines, "\n"),
					),
				)
			roadmapBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSecondary).
				Width(otherHalfWidth).
				Render(
					fmt.Sprintf("%s\n%s",
						StyleTitle.Render("  STOCK EXCHANGE CHARTER ROADMAP"),
						strings.Join(roadmapLines, "\n"),
					),
				)
			bottomRow = lipgloss.JoinHorizontal(lipgloss.Top, privateBox, " ", roadmapBox)
		}

		recentEvents := "No major national events yet."
		if len(events) > 0 {
			var evLines []string
			start := max(0, len(events)-4)
			for i := len(events) - 1; i >= start; i-- {
				evLines = append(evLines, "  • "+events[i])
			}
			recentEvents = strings.Join(evLines, "\n")
		}
		eventsBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Width(v.width - 4).
			Render(
				fmt.Sprintf("%s\n%s",
					StyleBold.Render("  NATIONAL DISPATCHES & SOVEREIGN TRANSMISSIONS"),
					recentEvents,
				),
			)

		outElements := []string{
			sovereignCards,
			"",
			middleRow,
			"",
			bottomRow,
		}
		if v.height >= 28 {
			outElements = append(outElements, "", eventsBox)
		}

		return lipgloss.JoinVertical(lipgloss.Left, outElements...)
	}

	// =========================================================================
	// MODE B: CHARTERED NATIONAL STOCK EXCHANGE (Active Public Market)
	// =========================================================================
	var totalMarketCap float64
	var advancers, decliners, unchanged int
	var sumPE float64
	var peCount int

	sortedByGain := make([]CompanyRow, len(rows))
	copy(sortedByGain, rows)
	sort.Slice(sortedByGain, func(i, j int) bool {
		return sortedByGain[i].ChangePct > sortedByGain[j].ChangePct
	})

	sortedByVol := make([]CompanyRow, len(rows))
	copy(sortedByVol, rows)
	sort.Slice(sortedByVol, func(i, j int) bool {
		return sortedByVol[i].Volume > sortedByVol[j].Volume
	})

	for _, r := range rows {
		totalMarketCap += r.MarketCap
		if r.ChangePct > 0.05 {
			advancers++
		} else if r.ChangePct < -0.05 {
			decliners++
		} else {
			unchanged++
		}
		if r.PE > 0 && r.PE < 200 {
			sumPE += r.PE
			peCount++
		}
	}
	avgPE := 0.0
	if peCount > 0 {
		avgPE = sumPE / float64(peCount)
	}

	debtToGDP := 0.0
	if natRep.GDP > 0 {
		debtToGDP = (natRep.NationalDebt / natRep.GDP) * 100.0
	}
	bannerText := fmt.Sprintf(" SOVEREIGN STATUS: Treasury Cash: %s  │  National Debt: %s (%.1f%% GDP)  │  Rating: %s (%.2f%%)  │  Exchange: %s",
		StyleGreen.Bold(true).Render(FormatCurrency(natRep.TreasuryCash)),
		FormatCurrency(natRep.NationalDebt),
		debtToGDP,
		StyleCyan.Bold(true).Render(natRep.CreditRating),
		natRep.BorrowingYield*100,
		StyleGreen.Bold(true).Render("[ CHARTERED & OPEN ]"),
	)
	sovereignBanner := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Width(v.width - 4).
		Padding(0, 1).
		Render(bannerText)

	var macroRow string
	if v.width < 100 {
		cardW := (v.width - 8) / 2
		otherCardW := v.width - 8 - cardW
		c1 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(cardW).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Total Market Cap"), StyleBold.Render(formatCurrency(totalMarketCap))),
		)
		c2 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(otherCardW).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s / %s", StyleMuted.Render("Advancers / Decliners"), StyleGreen.Render(fmt.Sprintf("+ %d", advancers)), StyleRed.Render(fmt.Sprintf("- %d", decliners))),
		)
		c3 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(cardW).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Avg Universe P/E"), StyleBold.Render(fmt.Sprintf("%.1fx", avgPE))),
		)
		c4 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(otherCardW).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Executed Trades"), StyleBold.Render(fmt.Sprintf("%d", totalTrades))),
		)
		macroRow = lipgloss.JoinVertical(lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, c1, " ", c2), "", lipgloss.JoinHorizontal(lipgloss.Top, c3, " ", c4))
	} else {
		colWidth := max((v.width-12)/4, 20)
		c1 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(colWidth).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Total Market Cap"), StyleBold.Render(formatCurrency(totalMarketCap))),
		)
		c2 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(colWidth).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s / %s", StyleMuted.Render("Advancers / Decliners"), StyleGreen.Render(fmt.Sprintf("+ %d", advancers)), StyleRed.Render(fmt.Sprintf("- %d", decliners))),
		)
		c3 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(colWidth).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Average Universe P/E"), StyleBold.Render(fmt.Sprintf("%.1fx", avgPE))),
		)
		c4 := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Width(colWidth).Padding(0, 1).Render(
			fmt.Sprintf("%s\n%s", StyleMuted.Render("Total Executed Trades"), StyleBold.Render(fmt.Sprintf("%d", totalTrades))),
		)
		macroRow = lipgloss.JoinHorizontal(lipgloss.Top, c1, " ", c2, " ", c3, " ", c4)
	}

	fedStanceColor := StyleCyan
	switch natRep.PolicyStance {
	case "Hawkish":
		fedStanceColor = StyleRed
	case "Dovish":
		fedStanceColor = StyleGreen
	}

	leftMacroLines := fmt.Sprintf(
		" %-22s : %s (%s)\n"+
			" %-22s : %.2f%% [%s stance]\n"+
			" %-22s : %.2f%%\n"+
			" %-22s : %.2f%% (Fed Target: 2.00%%)\n"+
			" %-22s : %.1f  │  Trade Balance: %s",
		"Gross Domestic Product", FormatCurrency(natRep.GDP), fmt.Sprintf("%+.1f%% real", natRep.RealGDPGrowth*100),
		"Federal Funds Rate", natRep.FedFundsRate*100, fedStanceColor.Render(natRep.PolicyStance),
		"10-Yr Benchmark Yield", natRep.TenYearYield*100,
		"CPI Inflation Rate", natRep.CPIInflationRate*100,
		"U.S. Dollar Index (DXY)", natRep.DollarIndexDXY, FormatCurrency(natRep.TradeBalance),
	)

	rightMacroLines := fmt.Sprintf(
		" %-22s : %s  │  Labor Force: %s\n"+
			" %-22s : %.2f%% (Employed: %s)\n"+
			" %-22s : $%.2f/hr (Growth: %+.1f%%)\n"+
			" %-22s : %.1f/100  │  Happiness: %.1f/100\n"+
			" %-22s : %s  │  PCE Spending: %s",
		"Citizen Population", FormatMetric(citRep.Population, false), FormatMetric(natRep.LaborForce, false),
		"Civilian Unemployment", natRep.UnemploymentRate*100, FormatMetric(natRep.EmployedWorkers, false),
		"Avg Hourly Earnings", natRep.AverageHourlyWage, natRep.AnnualWageGrowth*100,
		"Consumer Confidence", citRep.ConsumerConfidence, citRep.Happiness,
		"Federal National Debt", FormatCurrency(natRep.NationalDebt), FormatCurrency(citRep.ConsumerSpending),
	)

	var macroBoxRow string
	if v.width < 95 {
		boxLeft := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Width(v.width - 4).
			Padding(0, 1).
			Render(fmt.Sprintf("%s\n%s", StyleBold.Render("FEDERAL RESERVE & NATIONAL OUTPUT"), leftMacroLines))

		boxRight := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Width(v.width - 4).
			Padding(0, 1).
			Render(fmt.Sprintf("%s\n%s", StyleBold.Render("LABOR MARKET & CITIZEN SENTIMENT"), rightMacroLines))

		macroBoxRow = lipgloss.JoinVertical(lipgloss.Left, boxLeft, "", boxRight)
	} else {
		boxLeft := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Width(halfWidth).
			Padding(0, 1).
			Render(fmt.Sprintf("%s\n%s", StyleBold.Render("FEDERAL RESERVE & NATIONAL OUTPUT"), leftMacroLines))

		boxRight := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Width(otherHalfWidth).
			Padding(0, 1).
			Render(fmt.Sprintf("%s\n%s", StyleBold.Render("LABOR MARKET & CITIZEN SENTIMENT"), rightMacroLines))

		macroBoxRow = lipgloss.JoinHorizontal(lipgloss.Top, boxLeft, " ", boxRight)
	}

	gainersCount := min(6, len(sortedByGain))
	if v.height < 32 {
		gainersCount = min(3, len(sortedByGain))
	}
	gainerCols := []TableColumn{
		{Title: "SYM", Width: 6, AlignRight: false},
		{Title: "NAME", Width: 18, AlignRight: false},
		{Title: "PRICE", Width: 10, AlignRight: true},
		{Title: "GAIN", Width: 10, AlignRight: true},
	}
	topGainersLines := []string{
		StyleHeader.Render(FormatTableHeader(gainerCols)),
	}
	for i := range gainersCount {
		r := sortedByGain[i]
		topGainersLines = append(topGainersLines, FormatTableRow(
			gainerCols,
			[]string{r.Symbol, r.Name, fmt.Sprintf("%.2f", r.Price), fmt.Sprintf("+%.2f%%", r.ChangePct)},
			[]lipgloss.Style{StyleBold, lipgloss.NewStyle(), lipgloss.NewStyle(), StyleGreen},
		))
	}

	loserCols := []TableColumn{
		{Title: "SYM", Width: 6, AlignRight: false},
		{Title: "NAME", Width: 18, AlignRight: false},
		{Title: "PRICE", Width: 10, AlignRight: true},
		{Title: "DROP", Width: 10, AlignRight: true},
	}
	losersLines := []string{
		StyleHeader.Render(FormatTableHeader(loserCols)),
	}
	for i := len(sortedByGain) - 1; i >= max(0, len(sortedByGain)-gainersCount); i-- {
		r := sortedByGain[i]
		losersLines = append(losersLines, FormatTableRow(
			loserCols,
			[]string{r.Symbol, r.Name, fmt.Sprintf("%.2f", r.Price), fmt.Sprintf("%.2f%%", r.ChangePct)},
			[]lipgloss.Style{StyleBold, lipgloss.NewStyle(), lipgloss.NewStyle(), StyleRed},
		))
	}

	var tablesRow string
	if v.width < 95 {
		gainersBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Width(v.width - 4).
			Render(
				fmt.Sprintf("%s\n%s",
					StyleGreen.Render("  [+] TOP GAINERS"),
					strings.Join(topGainersLines, "\n"),
				),
			)
		losersBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDanger).
			Width(v.width - 4).
			Render(
				fmt.Sprintf("%s\n%s",
					StyleRed.Render("  [-] TOP DECLINERS"),
					strings.Join(losersLines, "\n"),
				),
			)
		tablesRow = lipgloss.JoinVertical(lipgloss.Left, gainersBox, "", losersBox)
	} else {
		gainersBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Width(halfWidth).
			Render(
				fmt.Sprintf("%s\n%s",
					StyleGreen.Render("  [+] TOP GAINERS"),
					strings.Join(topGainersLines, "\n"),
				),
			)
		losersBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDanger).
			Width(otherHalfWidth).
			Render(
				fmt.Sprintf("%s\n%s",
					StyleRed.Render("  [-] TOP DECLINERS"),
					strings.Join(losersLines, "\n"),
				),
			)
		tablesRow = lipgloss.JoinHorizontal(lipgloss.Top, gainersBox, " ", losersBox)
	}

	activeCount := min(5, len(sortedByVol))
	if v.height < 32 {
		activeCount = min(3, len(sortedByVol))
	}
	var activeCols []TableColumn
	if v.width < 85 {
		activeCols = []TableColumn{
			{Title: "SYM", Width: 6, AlignRight: false},
			{Title: "NAME", Width: 16, AlignRight: false},
			{Title: "PRICE", Width: 10, AlignRight: true},
			{Title: "VOLUME", Width: 12, AlignRight: true},
		}
	} else {
		activeCols = []TableColumn{
			{Title: "SYM", Width: 6, AlignRight: false},
			{Title: "NAME", Width: 20, AlignRight: false},
			{Title: "SECTOR", Width: 16, AlignRight: false},
			{Title: "PRICE", Width: 10, AlignRight: true},
			{Title: "VOLUME", Width: 14, AlignRight: true},
			{Title: "SPREAD", Width: 8, AlignRight: true},
		}
	}
	activeLines := []string{StyleHeader.Render(FormatTableHeader(activeCols))}
	for i := range activeCount {
		r := sortedByVol[i]
		var values []string
		var styles []lipgloss.Style
		if v.width < 85 {
			values = []string{r.Symbol, r.Name, fmt.Sprintf("%.2f", r.Price), formatVolume(r.Volume)}
			styles = []lipgloss.Style{StyleBold, lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle()}
		} else {
			values = []string{r.Symbol, r.Name, r.Sector, fmt.Sprintf("%.2f", r.Price), formatVolume(r.Volume), fmt.Sprintf("%.2f", r.Spread)}
			styles = []lipgloss.Style{StyleBold, lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle()}
		}
		activeLines = append(activeLines, FormatTableRow(activeCols, values, styles))
	}
	activeBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Width(v.width - 4).
		Render(
			fmt.Sprintf("%s\n%s",
				StyleTitle.Render("  MOST ACTIVE BY VOLUME"),
				strings.Join(activeLines, "\n"),
			),
		)

	recentEvents := "No major market events yet."
	if len(events) > 0 {
		var evLines []string
		start := max(0, len(events)-4)
		for i := len(events) - 1; i >= start; i-- {
			evLines = append(evLines, "  • "+events[i])
		}
		recentEvents = strings.Join(evLines, "\n")
	}
	eventsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(v.width - 4).
		Render(
			fmt.Sprintf("%s\n%s",
				StyleBold.Render("  REAL-TIME CORPORATE WIRE & MACRO TRANSMISSIONS"),
				recentEvents,
			),
		)

	outElements := []string{
		sovereignBanner,
		"",
		macroRow,
		"",
		macroBoxRow,
		"",
		tablesRow,
	}
	if v.height >= 26 {
		outElements = append(outElements, "", activeBox)
	}
	if v.height >= 32 {
		outElements = append(outElements, "", eventsBox)
	}

	return lipgloss.JoinVertical(lipgloss.Left, outElements...)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return string(runes[:n])
	}
	return string(runes[:n-1]) + "…"
}

func formatCurrency(val float64) string {
	return FormatCurrency(val)
}

func formatVolume(val float64) string {
	return FormatMetric(val, false)
}
