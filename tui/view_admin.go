package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"economy/country"
	"economy/country/fiscal"
	"economy/world"
)


var AdminSectors = []string{
	"Information Technology",
	"Health Care",
	"Financials",
	"Consumer Discretionary",
	"Consumer Staples",
	"Industrials",
	"Energy",
	"Materials",
	"Communication Services",
	"Utilities",
	"Real Estate",
}

var AdminTiers = []string{
	"Mega Cap",
	"Large Cap",
	"Mid Cap",
	"Small Cap",
}

type IPOTemplate struct {
	Name   string
	Symbol string
	Sector string
	Tier   string
	Shares string
	Price  string
}

var sampleIPOs = []IPOTemplate{
	{"Apex Technologies", "APEX", "Information Technology", "Mid Cap", "25000000", "45.00"},
	{"BioNova Therapeutics", "BNOV", "Health Care", "Small Cap", "12000000", "28.50"},
	{"Vanguard Energy Core", "VNRG", "Energy", "Large Cap", "120000000", "78.00"},
	{"Titan Heavy Industries", "TITN", "Industrials", "Large Cap", "90000000", "64.00"},
	{"Solaria CleanTech", "SOLA", "Utilities", "Mid Cap", "40000000", "36.00"},
	{"Meridian Global Bank", "MFIN", "Financials", "Large Cap", "150000000", "92.00"},
	{"OmniConsumer Brands", "OMNI", "Consumer Staples", "Mega Cap", "400000000", "110.00"},
	{"Quantum Systems Lab", "QSL", "Information Technology", "Small Cap", "15000000", "19.50"},
	{"Zephyr Aerospace", "ZPHR", "Industrials", "Mid Cap", "50000000", "85.00"},
	{"Nexus Communications", "NXCM", "Communication Services", "Mid Cap", "60000000", "42.00"},
	{"Aether Robotics", "AETH", "Information Technology", "Mid Cap", "35000000", "52.00"},
	{"Veritas CyberSec", "VSEC", "Information Technology", "Small Cap", "18000000", "31.00"},
}

type ClickArea struct {
	X1, Y1, X2, Y2 int
	Action         string
	Value          string
	Speed          float64
}

type AdminSubTab int

const (
	AdminSubTabTreasury AdminSubTab = iota
	AdminSubTabIPO
	AdminSubTabSystem
)

func (t AdminSubTab) String() string {
	switch t {
	case AdminSubTabIPO:
		return "Stock Exchange & IPO"
	case AdminSubTabSystem:
		return "Simulation & Policies"
	default:
		return "Sovereign Treasury & Capital"
	}
}

type AdminView struct {
	width        int
	height       int
	ActiveSubTab AdminSubTab
	IPOSymbol    string
	IPOName      string
	IPOSector    string
	IPOTier      string
	IPOShares    string
	IPOPrice     string
	ActiveField  int
	StatusMsg    string
	IsStatusErr  bool
	rng          *rand.Rand
	clickAreas   []ClickArea
	usedSymbols  map[string]bool
}

func NewAdminView() *AdminView {
	return &AdminView{
		width:        120,
		height:       35,
		ActiveSubTab: AdminSubTabTreasury,
		IPOSymbol:    "APEX",
		IPOName:      "Apex Technologies",
		IPOSector:    "Information Technology",
		IPOTier:      "Mid Cap",
		IPOShares:    "25000000",
		IPOPrice:     "45.00",
		ActiveField:  -1,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		clickAreas:   make([]ClickArea, 0),
		usedSymbols:  make(map[string]bool),
	}
}

func (v *AdminView) NextSubTab() {
	v.ActiveSubTab = (v.ActiveSubTab + 1) % 3
	v.ActiveField = -1
}

func (v *AdminView) PrevSubTab() {
	v.ActiveSubTab = (v.ActiveSubTab + 2) % 3
	v.ActiveField = -1
}

func (v *AdminView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *AdminView) SetStatus(msg string, isErr bool) {
	v.StatusMsg = msg
	v.IsStatusErr = isErr
}

func (v *AdminView) SetUsedSymbols(used map[string]bool) {
	v.usedSymbols = used
}

func (v *AdminView) CycleSector() {
	for i, s := range AdminSectors {
		if strings.EqualFold(s, v.IPOSector) {
			v.IPOSector = AdminSectors[(i+1)%len(AdminSectors)]
			return
		}
	}
	v.IPOSector = AdminSectors[0]
}

func (v *AdminView) CycleTier() {
	for i, t := range AdminTiers {
		if strings.EqualFold(t, v.IPOTier) {
			v.IPOTier = AdminTiers[(i+1)%len(AdminTiers)]
			return
		}
	}
	v.IPOTier = AdminTiers[0]
}

func (v *AdminView) RandomizeIPO() {
	// Filter available samples by those whose symbol isn't already used
	var available []IPOTemplate
	for _, tmpl := range sampleIPOs {
		if v.usedSymbols == nil || !v.usedSymbols[tmpl.Symbol] {
			available = append(available, tmpl)
		}
	}

	if len(available) > 0 {
		idx := v.rng.Intn(len(available))
		tmpl := available[idx]
		v.IPOName = tmpl.Name
		v.IPOSymbol = tmpl.Symbol
		v.IPOSector = tmpl.Sector
		v.IPOTier = tmpl.Tier
		v.IPOShares = tmpl.Shares
		v.IPOPrice = tmpl.Price
		v.StatusMsg = "Randomized idea: " + tmpl.Name + " (" + tmpl.Symbol + ")"
		v.IsStatusErr = false
		return
	}

	// All sample symbols taken — generate a completely random unique ticker
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for attempt := 0; attempt < 100; attempt++ {
		length := 3 + v.rng.Intn(2) // 3 or 4 letters
		sym := make([]byte, length)
		for i := range sym {
			sym[i] = letters[v.rng.Intn(len(letters))]
		}
		symStr := string(sym)
		if v.usedSymbols == nil || !v.usedSymbols[symStr] {
			v.IPOSymbol = symStr
			v.IPOName = "New " + symStr + " Corp"
			v.StatusMsg = "Randomized idea: " + v.IPOName + " (" + symStr + ")"
			v.IsStatusErr = false
			return
		}
	}
	v.StatusMsg = "Could not find a free symbol; try entering one manually"
	v.IsStatusErr = true
}

func (v *AdminView) Render(
	tick int64,
	uptime string,
	speed float64,
	totalCompanies int,
	totalTrades int64,
	totalVol float64,
	activePolicies []string,
	distressedCompanies []CompanyRow,
	natRep country.NationalReport,
	headerRows ...int,
) string {
	v.clickAreas = v.clickAreas[:0]

	hHeight := 2
	if len(headerRows) > 0 && headerRows[0] > 0 {
		hHeight = headerRows[0]
	}

	// 1. Sub-Tab Navigation Bar
	subTabs := []struct {
		tab   AdminSubTab
		label string
	}{
		{AdminSubTabTreasury, "[1] Sovereign Treasury & Capital"},
		{AdminSubTabIPO, "[2] Stock Exchange & IPO"},
		{AdminSubTabSystem, "[3] Simulation & Policies"},
	}

	var subTabButtons []string
	curX := 2
	subTabY := hHeight
	for _, st := range subTabs {
		var btn string
		if v.ActiveSubTab == st.tab {
			btn = StyleTabActive.Render(" " + st.label + " ")
		} else {
			btn = StyleTabInactive.Render(" " + st.label + " ")
		}
		subTabButtons = append(subTabButtons, btn)
		btnW := lipgloss.Width(btn)
		v.clickAreas = append(v.clickAreas, ClickArea{
			X1:     curX,
			Y1:     subTabY,
			X2:     curX + btnW,
			Y2:     subTabY + 1,
			Action: "admin_subtab",
			Speed:  float64(st.tab),
		})
		curX += btnW + 1
	}
	subTabBar := " " + strings.Join(subTabButtons, " ")

	var statusLine string
	if v.StatusMsg != "" {
		if v.IsStatusErr {
			statusLine = "\n  " + StyleRed.Render("[ERROR] "+v.StatusMsg)
		} else {
			statusLine = "\n  " + StyleGreen.Render("[OK] "+v.StatusMsg)
		}
	}

	contentStartY := hHeight + 2

	switch v.ActiveSubTab {
	case AdminSubTabTreasury:
		// Sovereign Treasury & Capital Desk
		debtToGDP := 0.0
		if natRep.GDP > 0 {
			debtToGDP = (natRep.NationalDebt / natRep.GDP) * 100.0
		}

		borrow1Label := "[ Borrow $1B ]"
		borrow5Label := "[ Borrow $5B ]"
		repay1Label := "[ Repay $1B ]"
		repay5Label := "[ Repay $5B ]"
		if v.width < 110 {
			borrow1Label = "[ +$1B ]"
			borrow5Label = "[ +$5B ]"
			repay1Label = "[ -$1B ]"
			repay5Label = "[ -$5B ]"
		}
		borrow1Btn := lipgloss.NewStyle().Background(lipgloss.Color("#1f6feb")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render(borrow1Label)
		borrow5Btn := lipgloss.NewStyle().Background(lipgloss.Color("#1f6feb")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render(borrow5Label)
		repay1Btn := lipgloss.NewStyle().Background(lipgloss.Color("#238636")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render(repay1Label)
		repay5Btn := lipgloss.NewStyle().Background(lipgloss.Color("#238636")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render(repay5Label)

		debtDeskLines := []string{
			fmt.Sprintf(" %-22s : %s", "Treasury Liquid Cash", StyleGreen.Bold(true).Render(FormatCurrency(natRep.TreasuryCash))),
			fmt.Sprintf(" %-22s : %s (%s)", "National Debt", StyleBold.Render(FormatCurrency(natRep.NationalDebt)), StyleMuted.Render(fmt.Sprintf("%.1f%% GDP", debtToGDP))),
			fmt.Sprintf(" %-22s : %s (Yield: %.2f%%)", "Credit Rating", StyleCyan.Bold(true).Render(natRep.CreditRating), natRep.BorrowingYield*100),
			"",
			fmt.Sprintf("  Issue Debt:   %s   %s", borrow1Btn, borrow5Btn),
			fmt.Sprintf("  Paydown Debt: %s    %s", repay1Btn, repay5Btn),
		}

		debtBoxWidth := (v.width - 6) / 2
		capBoxWidth := v.width - 6 - debtBoxWidth
		if v.width < 100 {
			debtBoxWidth = v.width - 4
			capBoxWidth = v.width - 4
		}

		debtDeskBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Width(debtBoxWidth).
			Padding(0, 1).
			Render(
				fmt.Sprintf("%s\n\n%s",
					StyleTitle.Render("SOVEREIGN TREASURY & DEBT ISSUANCE DESK"),
					strings.Join(debtDeskLines, "\n"),
				),
			)

		investBtn := func(label string) string {
			return lipgloss.NewStyle().Background(lipgloss.Color("#238636")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render(label)
		}

		capitalLines := []string{
			fmt.Sprintf(" %-20s : %4.1f/100  %s", "Infrastructure", natRep.InfrastructureLevel, investBtn("[ +$1B Infra ]")),
			fmt.Sprintf(" %-20s : %4.1f/100  %s", "Public Healthcare", natRep.HealthcareLevel, investBtn("[ +$1B Health ]")),
			fmt.Sprintf(" %-20s : %4.1f/100  %s", "Workforce Education", natRep.EducationLevel, investBtn("[ +$1B Edu ]")),
			fmt.Sprintf(" %-20s : %4.1f/100  %s", "Enterprise Grants", natRep.EnterpriseGrantsLevel, investBtn("[ +$1B Grants ]")),
			fmt.Sprintf(" %-20s : %4.1f/100  %s", "Export Logistics", natRep.ExportCapacityLevel, investBtn("[ +$1B Ports ]")),
		}

		capitalDeskBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Width(capBoxWidth).
			Padding(0, 1).
			Render(
				fmt.Sprintf("%s\n\n%s",
					StyleTitle.Render("NATIONAL CAPITAL INVESTMENTS (from Treasury)"),
					strings.Join(capitalLines, "\n"),
				),
			)

		var treasuryRow string
		if v.width < 100 {
			treasuryRow = lipgloss.JoinVertical(lipgloss.Left, debtDeskBox, "", capitalDeskBox)
		} else {
			treasuryRow = lipgloss.JoinHorizontal(lipgloss.Top, debtDeskBox, " ", capitalDeskBox)
		}

		debtBorrowY := contentStartY + 6
		debtRepayY := contentStartY + 7
		v.clickAreas = append(v.clickAreas,
			ClickArea{X1: 15, Y1: debtBorrowY - 1, X2: 32, Y2: debtBorrowY + 1, Action: "borrow_money", Value: "1000000000"},
			ClickArea{X1: 34, Y1: debtBorrowY - 1, X2: 51, Y2: debtBorrowY + 1, Action: "borrow_money", Value: "5000000000"},
			ClickArea{X1: 15, Y1: debtRepayY - 1, X2: 30, Y2: debtRepayY + 1, Action: "repay_debt", Value: "1000000000"},
			ClickArea{X1: 32, Y1: debtRepayY - 1, X2: 47, Y2: debtRepayY + 1, Action: "repay_debt", Value: "5000000000"},
		)

		investActions := []string{"invest_infrastructure", "invest_healthcare", "invest_education", "invest_enterprise", "invest_exports"}
		topRightOffsetX := debtBoxWidth + 1
		for i, act := range investActions {
			capY := contentStartY + 3 + i
			investX1 := topRightOffsetX + 30
			investX2 := topRightOffsetX + capBoxWidth - 2
			if v.width < 100 {
				capY = contentStartY + lipgloss.Height(debtDeskBox) + 1 + 3 + i
				investX1 = 30
				investX2 = v.width - 6
			}
			v.clickAreas = append(v.clickAreas, ClickArea{
				X1:     investX1,
				Y1:     capY - 1,
				X2:     investX2,
				Y2:     capY + 1,
				Action: act,
				Value:  "1000000000",
			})
		}

		treasuryHeight := lipgloss.Height(treasuryRow)
		charterBoxStartY := contentStartY + treasuryHeight + 1

		passTag := func(ok bool) string {
			if ok {
				return lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render("PASS")
			}
			return lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("REQ'D")
		}

		infraPass := natRep.InfrastructureLevel >= 25.0
		grantsPass := natRep.EnterpriseGrantsLevel >= 20.0
		cashPass := natRep.TreasuryCash >= 5_000_000_000.0
		cosPass := totalCompanies >= 3
		canCharter := infraPass && grantsPass && cashPass && cosPass

		var charterBox string
		if !natRep.ExchangeChartered {
			charterBtnStyle := lipgloss.NewStyle().Background(lipgloss.Color("#238636")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 2)
			charterBtnLabel := "[ INAUGURATE & CHARTER NATIONAL STOCK EXCHANGE ]"
			if !canCharter {
				charterBtnStyle = lipgloss.NewStyle().Background(lipgloss.Color("#484f58")).Foreground(lipgloss.Color("#8b949e")).Padding(0, 2)
				charterBtnLabel = "[ CHARTER NATIONAL STOCK EXCHANGE (Requirements Pending) ]"
			}

			var prereqLine string
			if v.width < 105 {
				prereqLine = fmt.Sprintf("  Prerequisites:\n    Infra >= 25.0 [%s]  •  Grants >= 20.0 [%s]\n    Cash >= $5B   [%s]  •  Enterprises >= 3 [%s]",
					passTag(infraPass), passTag(grantsPass), passTag(cashPass), passTag(cosPass),
				)
			} else {
				prereqLine = fmt.Sprintf("  Prerequisites: Infra >= 25.0 [%s]  •  Grants >= 20.0 [%s]  •  Cash >= $5B [%s]  •  Enterprises >= 3 [%s]",
					passTag(infraPass), passTag(grantsPass), passTag(cashPass), passTag(cosPass),
				)
			}

			charterLines := []string{
				prereqLine,
				"",
				"  " + charterBtnStyle.Render(charterBtnLabel),
			}

			charterBox = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorHighlight).
				Width(v.width - 4).
				Padding(0, 1).
				Render(
					fmt.Sprintf("%s\n\n%s",
						StyleTitle.Render("NATIONAL STOCK EXCHANGE CHARTERING CONSOLE"),
						strings.Join(charterLines, "\n"),
					),
				)

			btnY := charterBoxStartY + 4
			if v.width < 105 {
				btnY = charterBoxStartY + 5
			}
			v.clickAreas = append(v.clickAreas, ClickArea{
				X1:     3,
				Y1:     btnY - 1,
				X2:     min(v.width-4, 65),
				Y2:     btnY + 2,
				Action: "charter_exchange",
			})
		} else {
			charterBox = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess).
				Width(v.width - 4).
				Padding(0, 1).
				Render(
					fmt.Sprintf("%s\n\n  %s  %s",
						StyleTitle.Render("NATIONAL STOCK EXCHANGE CHARTER STATUS"),
						lipgloss.NewStyle().Background(ColorSuccess).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 2).Render("[ ACTIVE & CHARTERED ]"),
						StyleMuted.Render("The domestic capital markets and public stock trading are fully operational."),
					),
				)
		}

		return lipgloss.JoinVertical(lipgloss.Left,
			subTabBar,
			"",
			treasuryRow,
			"",
			charterBox,
			statusLine,
		)

	case AdminSubTabIPO:
		// Emergency Corporate Bailout Facility + IPO Issuer
		var bailoutLines []string
		if len(distressedCompanies) > 0 {
			bailoutCols := []TableColumn{
				{Title: "SYM", Width: 6, AlignRight: false},
				{Title: "NAME", Width: 22, AlignRight: false},
				{Title: "SECTOR", Width: 18, AlignRight: false},
				{Title: "PRICE", Width: 10, AlignRight: true},
				{Title: "DEBT", Width: 14, AlignRight: true},
				{Title: "EXECUTIVE ACTION", Width: 26, AlignRight: false},
			}
			bailoutLines = append(bailoutLines, StyleHeader.Render(FormatTableHeader(bailoutCols)))
			displayCount := min(4, len(distressedCompanies))
			for i := 0; i < displayCount; i++ {
				c := distressedCompanies[i]
				bailoutBtn := lipgloss.NewStyle().
					Background(ColorWarning).
					Foreground(lipgloss.Color("#000000")).
					Bold(true).
					Padding(0, 2).
					Render(fmt.Sprintf("[ BAIL OUT %s ($500M) ]", c.Symbol))

				values := []string{
					c.Symbol,
					c.Name,
					c.Sector,
					fmt.Sprintf("%.2f", c.Price),
					FormatCurrency(c.DebtOutstanding),
					bailoutBtn,
				}
				styles := []lipgloss.Style{
					StyleBold,
					lipgloss.NewStyle(),
					lipgloss.NewStyle(),
					lipgloss.NewStyle(),
					lipgloss.NewStyle(),
					lipgloss.NewStyle(),
				}
				bailoutLines = append(bailoutLines, FormatTableRow(bailoutCols, values, styles))

				btnY := contentStartY + 4 + i
				v.clickAreas = append(v.clickAreas, ClickArea{
					X1:     65,
					Y1:     btnY - 1,
					X2:     v.width - 2,
					Y2:     btnY + 1,
					Action: "bailout",
					Value:  c.Symbol,
				})
			}
		} else {
			bailoutLines = append(bailoutLines,
				"  All enterprise universe corporations maintain liquid balance sheets.",
				"  No companies currently require emergency federal Chapter 11/Bailout intervention ($ < $5.00).",
				"  "+StyleMuted.Render("Note: As business cycles and sessions evolve, distressed companies will appear here with one-click federal bailout options."),
			)
		}

		bailoutBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarning).
			Width(v.width - 4).
			Padding(0, 1).
			Render(
				fmt.Sprintf("%s\n\n%s",
					StyleBold.Render("EMERGENCY CORPORATE BAILOUT FACILITY (TREASURY RESCUE DESK)"),
					strings.Join(bailoutLines, "\n"),
				),
			)

		bailoutHeight := lipgloss.Height(bailoutBox)
		ipoBoxStartY := contentStartY + bailoutHeight + 1

		var ipoBox string
		if !natRep.ExchangeChartered {
			ipoBox = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder).
				Width(v.width - 4).
				Padding(1, 2).
				Render(
					fmt.Sprintf("%s\n\n  %s\n  %s",
						lipgloss.NewStyle().Foreground(ColorMuted).Bold(true).Render("INITIAL PUBLIC OFFERING (IPO) ISSUER [LOCKED]"),
						StyleWarning.Render("Stock Exchange is currently unchartered."),
						StyleMuted.Render("Charter the National Stock Exchange in the [1] Treasury tab once prerequisites are met to unlock public IPO offerings."),
					),
				)
		} else {
			fieldStyle := func(idx int) lipgloss.Style {
				if v.ActiveField == idx {
					return lipgloss.NewStyle().Background(ColorHighlight).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1)
				}
				return lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(lipgloss.Color("#c9d1d9")).Padding(0, 1)
			}
			pillBtn := func(label string) string {
				return lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorSecondary).Bold(true).Padding(0, 1).Render(label)
			}
			presetBtn := func(label string) string {
				return lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorMuted).Padding(0, 1).Render(label)
			}

			symField := fmt.Sprintf("Symbol: %s", fieldStyle(0).Render(" "+v.IPOSymbol+" "))
			nameField := fmt.Sprintf("Company Name: %s", fieldStyle(1).Render(" "+v.IPOName+" "))
			randomBtn := lipgloss.NewStyle().Background(lipgloss.Color("#1f6feb")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 2).Render("[ RANDOMIZE IDEA ]")
			formRow1 := lipgloss.JoinHorizontal(lipgloss.Center, symField, "    ", nameField, "    ", randomBtn)

			secPill := pillBtn(fmt.Sprintf("[ < %s > ]", v.IPOSector))
			tierPill := pillBtn(fmt.Sprintf("[ < %s > ]", v.IPOTier))
			secField := fmt.Sprintf("Sector: %s", secPill)
			tierField := fmt.Sprintf("Cap Tier: %s", tierPill)
			formRow2 := lipgloss.JoinHorizontal(lipgloss.Center, secField, "    ", tierField)

			sharesField := fmt.Sprintf("Shares: %s", fieldStyle(4).Render(" "+v.IPOShares+" "))
			sharesPresets := lipgloss.JoinHorizontal(lipgloss.Center,
				presetBtn("[ 5M ]"), " ",
				presetBtn("[ 10M ]"), " ",
				presetBtn("[ 25M ]"), " ",
				presetBtn("[ 50M ]"), " ",
				presetBtn("[ 100M ]"),
			)
			formRow3 := lipgloss.JoinHorizontal(lipgloss.Center, sharesField, "  Presets: ", sharesPresets)

			priceField := fmt.Sprintf("Offering Price: %s", fieldStyle(5).Render(" $"+v.IPOPrice+" "))
			pricePresets := lipgloss.JoinHorizontal(lipgloss.Center,
				presetBtn("[ $10 ]"), " ",
				presetBtn("[ $25 ]"), " ",
				presetBtn("[ $50 ]"), " ",
				presetBtn("[ $100 ]"), " ",
				presetBtn("[ $250 ]"),
			)
			formRow4 := lipgloss.JoinHorizontal(lipgloss.Center, priceField, "  Presets: ", pricePresets)

			launchBtn := lipgloss.NewStyle().Background(ColorSuccess).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 3).Render("[ LAUNCH IPO ONTO EXCHANGE (Enter) ]")

			statusText := ""
			if v.StatusMsg != "" {
				if v.IsStatusErr {
					statusText = "\n  " + StyleRed.Render("[ERROR] "+v.StatusMsg)
				} else {
					statusText = "\n  " + StyleGreen.Render("[OK] "+v.StatusMsg)
				}
			}

			ipoBox = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess).
				Width(v.width - 4).
				Padding(1, 2).
				Render(
					fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s%s\n\n%s",
						StyleGreen.Render("INITIAL PUBLIC OFFERING (IPO) ISSUER (Click any field, pill, or preset with mouse)"),
						formRow1,
						formRow2,
						formRow3,
						formRow4,
						launchBtn,
						statusText,
						StyleMuted.Render("Controls: Mouse click to edit or pick  •  [Tab] Next field  •  [Esc] Unfocus  •  [Enter] Launch IPO"),
					),
				)

			v.registerIPOClickAreas(ipoBoxStartY)
		}

		return lipgloss.JoinVertical(lipgloss.Left,
			subTabBar,
			"",
			bailoutBox,
			"",
			ipoBox,
			statusLine,
		)

	default: // AdminSubTabSystem
		activePolicyMap := make(map[string]bool, len(activePolicies))
		for _, p := range activePolicies {
			activePolicyMap[p] = true
		}

		statusSpeed := fmt.Sprintf("%.1fx (%s)", speed, world.SpeedLabel(speed))
		if speed <= 0 {
			statusSpeed = "PAUSED"
		}

		engineLines := []string{
			fmt.Sprintf(" %-24s : %s", "Current Simulation Tick", StyleBold.Render(fmt.Sprintf("%d", tick))),
			fmt.Sprintf(" %-24s : %s", "Active Speed Multiplier", StyleBold.Render(statusSpeed)),
			fmt.Sprintf(" %-24s : %s", "Active Companies Universe", StyleBold.Render(fmt.Sprintf("%d stocks", totalCompanies))),
			fmt.Sprintf(" %-24s : %s", "Total Executed Trades", StyleBold.Render(FormatIntegerWithCommas(totalTrades))),
			fmt.Sprintf(" %-24s : %s", "Total Cumulative Volume", StyleBold.Render(FormatCurrency(totalVol))),
			"",
			StyleBold.Render("  SPEED CONTROLS (Click to set):"),
			"  [ PAUSE ]  [ 0.5x ]  [ 1.0x ]  [ 2.0x ]  [ 5.0x ]  [ 10.0x ]",
			"",
			StyleBold.Render("  STATE PERSISTENCE (Click to execute):"),
			fmt.Sprintf("  %s    %s",
				lipgloss.NewStyle().Background(lipgloss.Color("#1f6feb")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render("[ SAVE WORLD STATE ]"),
				lipgloss.NewStyle().Background(lipgloss.Color("#da3633")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render("[ RESET WORLD FRESH ]"),
			),
		}

		halfWidth := (v.width - 6) / 2
		otherHalfWidth := v.width - 6 - halfWidth
		engineBoxWidth := halfWidth
		policyBoxWidth := otherHalfWidth
		if v.width < 100 {
			engineBoxWidth = v.width - 4
			policyBoxWidth = v.width - 4
		}

		engineBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Width(engineBoxWidth).
			Padding(0, 1).
			Render(
				fmt.Sprintf("%s\n\n%s",
					StyleTitle.Render("SIMULATION ENGINE & PERSISTENCE"),
					strings.Join(engineLines, "\n"),
				),
			)

		// Register click areas for speed buttons and persistence buttons (matching lines 11-16)
		v.clickAreas = append(v.clickAreas,
			ClickArea{X1: 3, Y1: 11, X2: 13, Y2: 13, Action: "set_speed", Speed: 0.0},
			ClickArea{X1: 14, Y1: 11, X2: 23, Y2: 13, Action: "set_speed", Speed: 0.5},
			ClickArea{X1: 24, Y1: 11, X2: 33, Y2: 13, Action: "set_speed", Speed: 1.0},
			ClickArea{X1: 34, Y1: 11, X2: 43, Y2: 13, Action: "set_speed", Speed: 2.0},
			ClickArea{X1: 44, Y1: 11, X2: 53, Y2: 13, Action: "set_speed", Speed: 5.0},
			ClickArea{X1: 54, Y1: 11, X2: 65, Y2: 13, Action: "set_speed", Speed: 10.0},
			ClickArea{X1: 3, Y1: 14, X2: 27, Y2: 16, Action: "save_state"},
			ClickArea{X1: 28, Y1: 14, X2: 55, Y2: 16, Action: "reset_state"},
		)

		policyLines := []string{
			StyleMuted.Render("Click policy pill to Enact or Repeal executive federal programs:"),
			"",
		}

		topRightOffsetX := lipgloss.Width(engineBox) + 1
		const tabHeaderRows = 2
		const policyBoxPaddingRows = 4
		curPolicyY := tabHeaderRows + policyBoxPaddingRows + 1
		if v.width < 100 {
			topRightOffsetX = 0
			curPolicyY = tabHeaderRows + lipgloss.Height(engineBox) + 1 + policyBoxPaddingRows
		}

		for i, p := range fiscal.KnownPolicies {
			isActive := activePolicyMap[p.ID]
			statusPill := lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorMuted).Padding(0, 1).Render("[ INACTIVE ]")
			if isActive {
				statusPill = lipgloss.NewStyle().Background(ColorSuccess).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render("[ ENACTED  ]")
			}

			costStr := fmt.Sprintf("-$%.0fB/yr", p.AnnualCost/1e9)
			if p.AnnualCost < 0 {
				costStr = fmt.Sprintf("+$%.0fB/yr rev", -p.AnnualCost/1e9)
			}

			nameCell := FormatCell(p.Name, 26, false, StyleBold)
			costCell := FormatCell(costStr, 14, true, StyleMuted)
			sectorCell := FormatCell(p.FavoredSector, 16, false, StyleCyan)
			line := fmt.Sprintf("  %s %s %s  %s", statusPill, nameCell, costCell, sectorCell)
			policyLines = append(policyLines, line)

			v.clickAreas = append(v.clickAreas, ClickArea{
				X1:     topRightOffsetX + 1,
				Y1:     curPolicyY - 2,
				X2:     topRightOffsetX + policyBoxWidth,
				Y2:     curPolicyY + 2,
				Action: "toggle_policy",
				Value:  p.ID,
			})
			curPolicyY += 2
			if i < len(fiscal.KnownPolicies)-1 {
				policyLines = append(policyLines, "  "+StyleMuted.Render(p.Description))
			}
		}

		policyBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Width(policyBoxWidth).
			Padding(0, 1).
			Render(
				fmt.Sprintf("%s\n\n%s",
					StyleTitle.Render("EXECUTIVE GOVERNMENT ECONOMIC POLICIES"),
					strings.Join(policyLines, "\n"),
				),
			)

		var topRow string
		if v.width < 100 {
			topRow = lipgloss.JoinVertical(lipgloss.Left, engineBox, "", policyBox)
		} else {
			topRow = lipgloss.JoinHorizontal(lipgloss.Top, engineBox, " ", policyBox)
		}

		return lipgloss.JoinVertical(lipgloss.Left,
			subTabBar,
			"",
			topRow,
			statusLine,
		)
	}
}

func (v *AdminView) registerIPOClickAreas(baseY int) {
	// Row 1: Symbol, Name, Randomize (centered at baseY + 4)
	v.clickAreas = append(v.clickAreas,
		ClickArea{X1: 3, Y1: baseY + 3, X2: 24, Y2: baseY + 5, Action: "focus_field", Speed: 0},
		ClickArea{X1: 25, Y1: baseY + 3, X2: 70, Y2: baseY + 5, Action: "focus_field", Speed: 1},
		ClickArea{X1: 71, Y1: baseY + 3, X2: 100, Y2: baseY + 5, Action: "randomize_ipo"},
	)

	// Row 2: Sector pill, Tier pill (centered at baseY + 6)
	v.clickAreas = append(v.clickAreas,
		ClickArea{X1: 3, Y1: baseY + 5, X2: 44, Y2: baseY + 7, Action: "cycle_sector"},
		ClickArea{X1: 45, Y1: baseY + 5, X2: 85, Y2: baseY + 7, Action: "cycle_tier"},
	)

	// Row 3: Shares field & presets (centered at baseY + 8)
	v.clickAreas = append(v.clickAreas,
		ClickArea{X1: 3, Y1: baseY + 7, X2: 28, Y2: baseY + 9, Action: "focus_field", Speed: 4},
		ClickArea{X1: 38, Y1: baseY + 7, X2: 49, Y2: baseY + 9, Action: "set_shares", Value: "5000000"},
		ClickArea{X1: 50, Y1: baseY + 7, X2: 59, Y2: baseY + 9, Action: "set_shares", Value: "10000000"},
		ClickArea{X1: 60, Y1: baseY + 7, X2: 69, Y2: baseY + 9, Action: "set_shares", Value: "25000000"},
		ClickArea{X1: 70, Y1: baseY + 7, X2: 79, Y2: baseY + 9, Action: "set_shares", Value: "50000000"},
		ClickArea{X1: 80, Y1: baseY + 7, X2: 95, Y2: baseY + 9, Action: "set_shares", Value: "100000000"},
	)

	// Row 4: Price field & presets (centered at baseY + 10)
	v.clickAreas = append(v.clickAreas,
		ClickArea{X1: 3, Y1: baseY + 9, X2: 32, Y2: baseY + 11, Action: "focus_field", Speed: 5},
		ClickArea{X1: 38, Y1: baseY + 9, X2: 52, Y2: baseY + 11, Action: "set_price", Value: "10.00"},
		ClickArea{X1: 53, Y1: baseY + 9, X2: 61, Y2: baseY + 11, Action: "set_price", Value: "25.00"},
		ClickArea{X1: 62, Y1: baseY + 9, X2: 70, Y2: baseY + 11, Action: "set_price", Value: "50.00"},
		ClickArea{X1: 71, Y1: baseY + 9, X2: 80, Y2: baseY + 11, Action: "set_price", Value: "100.00"},
		ClickArea{X1: 81, Y1: baseY + 9, X2: 95, Y2: baseY + 11, Action: "set_price", Value: "250.00"},
	)

	// Row 5: Launch button (centered at baseY + 12)
	v.clickAreas = append(v.clickAreas,
		ClickArea{X1: 3, Y1: baseY + 11, X2: 50, Y2: baseY + 14, Action: "launch_ipo"},
	)
}

// HandleMouseClick evaluates a mouse click against recorded click areas
func (v *AdminView) HandleMouseClick(x, y int) (action string, value string, speed float64) {
	for _, a := range v.clickAreas {
		if x >= a.X1 && x <= a.X2 && y >= a.Y1 && y <= a.Y2 {
			switch a.Action {
			case "admin_subtab":
				v.ActiveSubTab = AdminSubTab(int(a.Speed))
				v.ActiveField = -1
				return a.Action, a.Value, a.Speed
			case "focus_field":
				v.ActiveField = int(a.Speed)
				return a.Action, a.Value, a.Speed
			case "cycle_sector":
				v.CycleSector()
				return a.Action, a.Value, a.Speed
			case "cycle_tier":
				v.CycleTier()
				return a.Action, a.Value, a.Speed
			case "randomize_ipo":
				v.RandomizeIPO()
				return a.Action, a.Value, a.Speed
			case "set_shares":
				v.IPOShares = a.Value
				return a.Action, a.Value, a.Speed
			case "set_price":
				v.IPOPrice = a.Value
				return a.Action, a.Value, a.Speed
			default:
				return a.Action, a.Value, a.Speed
			}
		}
	}
	return "", "", 0
}
