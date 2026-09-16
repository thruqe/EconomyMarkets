package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"economy/country/fiscal"
)

// PolicyCategory navigation tabs for the Policy Studio
type PolicyCategory int

const (
	PolicyCatFiscal PolicyCategory = iota
	PolicyCatMonetary
	PolicyCatTrade
	PolicyCatRegulatory
	PolicyCatLabor
)

func (c PolicyCategory) String() string {
	switch c {
	case PolicyCatMonetary:
		return "Monetary"
	case PolicyCatTrade:
		return "Trade"
	case PolicyCatRegulatory:
		return "Regulatory"
	case PolicyCatLabor:
		return "Labor"
	default:
		return "Fiscal"
	}
}

// PolicyView renders the Policy Studio tab (Tab 9).
type PolicyView struct {
	width       int
	height      int
	activeCategory PolicyCategory
	clickAreas  []ClickArea
	statusMsg   string
	isStatusErr bool
}

func NewPolicyView() *PolicyView {
	return &PolicyView{width: 120, height: 35, activeCategory: PolicyCatFiscal}
}

func (v *PolicyView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

func (v *PolicyView) SetStatus(msg string, isErr bool) {
	v.statusMsg = msg
	v.isStatusErr = isErr
}

func (v *PolicyView) Render(activePolicies []string, headerRows ...int) string {
	v.clickAreas = v.clickAreas[:0]

	hHeight := 2
	if len(headerRows) > 0 && headerRows[0] > 0 {
		hHeight = headerRows[0]
	}

	activePolicyMap := make(map[string]bool, len(activePolicies))
	for _, p := range activePolicies {
		activePolicyMap[p] = true
	}

	// Category navigation bar
	categories := []PolicyCategory{
		PolicyCatFiscal, PolicyCatMonetary, PolicyCatTrade, PolicyCatRegulatory, PolicyCatLabor,
	}
	catColors := map[PolicyCategory]string{
		PolicyCatFiscal:     "#f0883e",
		PolicyCatMonetary:   "#58a6ff",
		PolicyCatTrade:      "#3fb950",
		PolicyCatRegulatory: "#bc8cff",
		PolicyCatLabor:      "#e3b341",
	}

	var catButtons []string
	catStartX := 3
	catBarY := hHeight + 2
	for ci, cat := range categories {
		label := " " + cat.String() + " "
		style := lipgloss.NewStyle().Padding(0, 1)
		if cat == v.activeCategory {
			style = style.Background(lipgloss.Color(catColors[cat])).Foreground(lipgloss.Color("#0d1117")).Bold(true)
		} else {
			style = style.Foreground(lipgloss.Color(catColors[cat])).Background(lipgloss.Color("#21262d"))
		}
		btn := style.Render(label)
		catButtons = append(catButtons, btn)
		btnWidth := lipgloss.Width(btn)
		v.clickAreas = append(v.clickAreas, ClickArea{
			X1: catStartX, Y1: catBarY - 1, X2: catStartX + btnWidth, Y2: catBarY + 1,
			Action: "policy_cat", Speed: float64(ci),
		})
		catStartX += btnWidth + 1
	}
	catBar := "  " + strings.Join(catButtons, " ")

	// Get the right policy list for current category
	var policies []fiscal.PolicyInfo
	var catTitle string
	switch v.activeCategory {
	case PolicyCatMonetary:
		policies = fiscal.KnownMonetaryPolicies
		catTitle = "MONETARY POLICY — Central Bank & Money Supply Controls"
	case PolicyCatTrade:
		policies = fiscal.KnownTradePolicies
		catTitle = "TRADE POLICY — Import/Export & Bilateral Trade Agreements"
	case PolicyCatRegulatory:
		policies = fiscal.KnownRegulatoryPolicies
		catTitle = "REGULATORY POLICY — Antitrust, Deregulation & Sector Rules"
	case PolicyCatLabor:
		policies = fiscal.KnownLaborPolicies
		catTitle = "LABOR POLICY — Wages, Immigration & Workers' Rights"
	default:
		policies = fiscal.KnownPolicies
		catTitle = "FISCAL POLICY — Government Spending & Subsidies"
	}

	// Render policy cards
	var policyCards []string
	policyStartY := hHeight + 6 // headerHeight + border(1) + pad(1) + catBar(1) + blank(1) + title(1) + blank(1)

	for i, p := range policies {
		isActive := activePolicyMap[p.ID]

		var pillStyle lipgloss.Style
		var pillLabel string
		if isActive {
			pillStyle = lipgloss.NewStyle().Background(ColorSuccess).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 2)
			pillLabel = "[ ENACTED  ]"
		} else {
			pillStyle = lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorMuted).Padding(0, 2)
			pillLabel = "[ INACTIVE ]"
		}
		pill := pillStyle.Render(pillLabel)

		costStr := ""
		if p.AnnualCost > 0 {
			costStr = fmt.Sprintf("Cost: $%.0fB/yr", p.AnnualCost/1e9)
		} else if p.AnnualCost < 0 {
			costStr = fmt.Sprintf("Revenue: +$%.0fB/yr", -p.AnnualCost/1e9)
		} else {
			costStr = "No direct cost"
		}

		card := fmt.Sprintf("  %s  %s  %s\n     %s\n     Favors: %s",
			pill,
			StyleBold.Render(p.Name),
			StyleMuted.Render(costStr),
			StyleMuted.Render(p.Description),
			StyleCyan.Render(p.FavoredSector),
		)
		policyCards = append(policyCards, card)

		// Register click area for toggle: spans whole card width and height
		cardY := policyStartY + i*4
		v.clickAreas = append(v.clickAreas, ClickArea{
			X1:     2,
			Y1:     cardY - 1,
			X2:     v.width - 4,
			Y2:     cardY + 3,
			Action: "toggle_policy",
			Value:  p.ID,
		})
	}

	// Status bar
	var statusLine string
	if v.statusMsg != "" {
		if v.isStatusErr {
			statusLine = "\n  " + StyleRed.Render("[ERROR] "+v.statusMsg)
		} else {
			statusLine = "\n  " + StyleGreen.Render("[OK] "+v.statusMsg)
		}
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(catColors[v.activeCategory])).
		Width(v.width - 4).
		Padding(1, 2).
		Render(
			fmt.Sprintf("%s\n\n%s\n\n%s%s\n\n%s",
				catBar,
				StyleTitle.Render(catTitle),
				strings.Join(policyCards, "\n\n"),
				statusLine,
				StyleMuted.Render("Click any policy card to ENACT or REPEAL it immediately  •  [1-5] Category keys"),
			),
		)

	return contentBox
}

// HandleMouseClick processes clicks in the policy view.
func (v *PolicyView) HandleMouseClick(x, y int) (action, value string, cat PolicyCategory, catSet bool) {
	for _, a := range v.clickAreas {
		if x >= a.X1 && x <= a.X2 && y >= a.Y1 && y <= a.Y2 {
			if a.Action == "policy_cat" {
				v.activeCategory = PolicyCategory(int(a.Speed))
				return a.Action, a.Value, v.activeCategory, true
			}
			return a.Action, a.Value, v.activeCategory, false
		}
	}
	return "", "", v.activeCategory, false
}
