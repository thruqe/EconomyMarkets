package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type UniverseView struct {
	width          int
	height         int
	cursor         int
	scrollOffset   int
	searchQuery    string
	searching      bool
	selectedSector string
	selectedTier   string
}

func NewUniverseView() *UniverseView {
	return &UniverseView{
		width:          120,
		height:         35,
		cursor:         0,
		scrollOffset:   0,
		selectedSector: "All",
		selectedTier:   "All",
	}
}

func (v *UniverseView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *UniverseView) Cursor() int {
	return v.cursor
}

func (v *UniverseView) MoveUp() {
	if v.cursor > 0 {
		v.cursor--
		if v.cursor < v.scrollOffset {
			v.scrollOffset = v.cursor
		}
	}
}

func (v *UniverseView) MoveDown(totalItems int) {
	if v.cursor < totalItems-1 {
		v.cursor++
		visibleRows := max(v.height-12, 5)
		if v.cursor >= v.scrollOffset+visibleRows {
			v.scrollOffset = v.cursor - visibleRows + 1
		}
	}
}

func (v *UniverseView) PageUp() {
	visibleRows := max(v.height-12, 5)
	v.cursor -= visibleRows
	if v.cursor < 0 {
		v.cursor = 0
	}
	v.scrollOffset = v.cursor
}

func (v *UniverseView) PageDown(totalItems int) {
	visibleRows := max(v.height-12, 5)
	v.cursor += visibleRows
	if v.cursor >= totalItems {
		v.cursor = max(0, totalItems-1)
	}
	v.scrollOffset = max(0, v.cursor-visibleRows+1)
}

func (v *UniverseView) SetSearch(query string) {
	v.searchQuery = query
	v.cursor = 0
	v.scrollOffset = 0
}

func (v *UniverseView) ClearSearch() {
	v.searchQuery = ""
	v.cursor = 0
	v.scrollOffset = 0
}

func (v *UniverseView) SetSector(sec string) {
	v.selectedSector = sec
	v.cursor = 0
	v.scrollOffset = 0
}

func (v *UniverseView) SetTier(tier string) {
	v.selectedTier = tier
	v.cursor = 0
	v.scrollOffset = 0
}

func (v *UniverseView) FilteredRows(rows []CompanyRow) []CompanyRow {
	var filtered []CompanyRow
	q := strings.ToLower(v.searchQuery)

	for _, r := range rows {
		if v.selectedSector != "All" && !strings.EqualFold(r.Sector, v.selectedSector) {
			continue
		}
		if v.selectedTier != "All" && !strings.EqualFold(r.CapTier, v.selectedTier) {
			continue
		}
		if q != "" {
			if !strings.Contains(strings.ToLower(r.Symbol), q) &&
				!strings.Contains(strings.ToLower(r.Name), q) &&
				!strings.Contains(strings.ToLower(r.Sector), q) {
				continue
			}
		}
		filtered = append(filtered, r)
	}
	return filtered
}

func (v *UniverseView) SelectedCompany(rows []CompanyRow) *CompanyRow {
	filtered := v.FilteredRows(rows)
	if len(filtered) == 0 || v.cursor < 0 || v.cursor >= len(filtered) {
		return nil
	}
	return &filtered[v.cursor]
}

func (v *UniverseView) Render(rows []CompanyRow, isSearching bool, isChartered bool) string {
	filtered := v.FilteredRows(rows)
	total := len(filtered)

	var statusHeader string
	if !isChartered {
		statusHeader = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarning).
			Width(v.width - 4).
			Padding(0, 1).
			Render(fmt.Sprintf(" %s  │  %s\n %s",
				lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render("NATIONAL PRIVATE ENTERPRISE REGISTRY (PRE-MARKET)"),
				StyleMuted.Render("Stock Exchange not yet chartered"),
				StyleCyan.Render("Domestic businesses operate privately. Use Administration [Tab 7] to charter the Stock Exchange once ready."),
			))
	}

	// Search & Filter header
	searchPrompt := "/"
	if isSearching {
		searchPrompt = "SEARCH (Type query, press Enter): " + v.searchQuery + "_"
	} else if v.searchQuery != "" {
		searchPrompt = fmt.Sprintf("Filter: %s [Esc to clear, / to edit]", v.searchQuery)
	} else {
		searchPrompt = "Press '/' to search symbols or companies | 's' sector filter | 't' cap tier filter"
	}

	searchBar := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Width(v.width-4).
		Padding(0, 1).
		Render(searchPrompt)

	// Filter tabs bar
	filterBar := lipgloss.NewStyle().Foreground(ColorMuted).Render(
		fmt.Sprintf(" Sector: %s  |  Cap Tier: %s  |  Showing %d of %d enterprises",
			StyleBold.Render(v.selectedSector),
			StyleBold.Render(v.selectedTier),
			total,
			len(rows),
		),
	)

	// Table Dimensions
	visibleRows := max(v.height-12, 4)
	if !isChartered {
		visibleRows = max(v.height-15, 3)
	}

	if v.cursor >= total {
		v.cursor = max(0, total-1)
	}
	if v.scrollOffset > v.cursor {
		v.scrollOffset = v.cursor
	}

	var cols []TableColumn
	if !isChartered {
		if v.width < 85 {
			cols = []TableColumn{
				{Title: "SYM", Width: 6, AlignRight: false},
				{Title: "ENTERPRISE", Width: 16, AlignRight: false},
				{Title: "STAGE", Width: 10, AlignRight: false},
				{Title: "VALUATION", Width: 12, AlignRight: true},
				{Title: "STATUS", Width: 10, AlignRight: false},
			}
		} else if v.width < 115 {
			cols = []TableColumn{
				{Title: "SYM", Width: 6, AlignRight: false},
				{Title: "ENTERPRISE", Width: 18, AlignRight: false},
				{Title: "SECTOR", Width: 14, AlignRight: false},
				{Title: "STAGE", Width: 10, AlignRight: false},
				{Title: "VALUATION", Width: 12, AlignRight: true},
				{Title: "WORKFORCE", Width: 10, AlignRight: true},
				{Title: "STATUS", Width: 10, AlignRight: false},
			}
		} else {
			cols = []TableColumn{
				{Title: "SYM", Width: 6, AlignRight: false},
				{Title: "ENTERPRISE NAME", Width: 22, AlignRight: false},
				{Title: "SECTOR", Width: 18, AlignRight: false},
				{Title: "STAGE", Width: 10, AlignRight: false},
				{Title: "VALUATION", Width: 14, AlignRight: true},
				{Title: "ANNUAL REV", Width: 14, AlignRight: true},
				{Title: "WORKFORCE", Width: 10, AlignRight: true},
				{Title: "STATUS", Width: 12, AlignRight: false},
			}
		}
	} else {
		if v.width < 85 {
			cols = []TableColumn{
				{Title: "SYM", Width: 6, AlignRight: false},
				{Title: "NAME", Width: 16, AlignRight: false},
				{Title: "PRICE", Width: 10, AlignRight: true},
				{Title: "CHG %", Width: 10, AlignRight: true},
				{Title: "VOLUME", Width: 10, AlignRight: true},
			}
		} else if v.width < 115 {
			cols = []TableColumn{
				{Title: "SYM", Width: 6, AlignRight: false},
				{Title: "NAME", Width: 18, AlignRight: false},
				{Title: "SECTOR", Width: 14, AlignRight: false},
				{Title: "PRICE", Width: 10, AlignRight: true},
				{Title: "SPREAD", Width: 8, AlignRight: true},
				{Title: "CHG %", Width: 10, AlignRight: true},
				{Title: "VOLUME", Width: 10, AlignRight: true},
			}
		} else {
			cols = []TableColumn{
				{Title: "SYM", Width: 6, AlignRight: false},
				{Title: "NAME", Width: 20, AlignRight: false},
				{Title: "SECTOR", Width: 16, AlignRight: false},
				{Title: "CAP TIER", Width: 10, AlignRight: false},
				{Title: "PRICE", Width: 10, AlignRight: true},
				{Title: "BID/ASK", Width: 11, AlignRight: true},
				{Title: "SPREAD", Width: 8, AlignRight: true},
				{Title: "CHG %", Width: 10, AlignRight: true},
				{Title: "VOLUME", Width: 12, AlignRight: true},
				{Title: "P/E", Width: 8, AlignRight: true},
			}
		}
	}

	header := StyleHeader.Width(v.width - 6).Render(FormatTableHeader(cols))

	var tableLines []string
	tableLines = append(tableLines, header)

	end := min(v.scrollOffset+visibleRows, total)
	for i := v.scrollOffset; i < end; i++ {
		r := filtered[i]

		var formattedRow string
		if !isChartered {
			stage := r.Stage
			if stage == "" {
				stage = "Seed"
			}
			val := r.PrivateValuation
			if val <= 0 {
				val = r.ReportedValue
			}
			statusLabel := "Private"
			statusStyle := StyleMuted
			if stage == "Pre-IPO" {
				statusLabel = "IPO Ready"
				statusStyle = StyleGreen.Bold(true)
			} else if stage == "Growth" {
				statusLabel = "Expanding"
				statusStyle = StyleCyan
			}

			rowStyle := StyleTableRow
			if i == v.cursor {
				rowStyle = StyleTableRowSelected
			}

			symStyle := lipgloss.NewStyle()
			if i != v.cursor {
				symStyle = StyleBold
			}

			var values []string
			var styles []lipgloss.Style
			if v.width < 85 {
				values = []string{r.Symbol, r.Name, stage, FormatCurrency(val), statusLabel}
				styles = []lipgloss.Style{symStyle, lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), statusStyle}
			} else if v.width < 115 {
				values = []string{r.Symbol, r.Name, r.Sector, stage, FormatCurrency(val), fmt.Sprintf("%.0f", r.Headcount), statusLabel}
				styles = []lipgloss.Style{symStyle, lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), statusStyle}
			} else {
				values = []string{r.Symbol, r.Name, r.Sector, stage, FormatCurrency(val), FormatCurrency(val * 0.25), fmt.Sprintf("%.0f", r.Headcount), statusLabel}
				styles = []lipgloss.Style{symStyle, lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), statusStyle}
			}

			rowContent := FormatTableRow(cols, values, styles)
			formattedRow = rowStyle.Width(v.width - 6).Render(rowContent)
		} else {
			bidAsk := fmt.Sprintf("%.2f/%.2f", r.Bid, r.Ask)
			chgStr := fmt.Sprintf("%+6.2f%%", r.ChangePct)
			chgStyled := StyleGreen
			if r.ChangePct < 0 {
				chgStyled = StyleRed
			}

			rowStyle := StyleTableRow
			if i == v.cursor {
				rowStyle = StyleTableRowSelected
			}

			symStyle := lipgloss.NewStyle()
			if i != v.cursor {
				symStyle = StyleBold
			}

			var values []string
			var styles []lipgloss.Style
			if v.width < 85 {
				values = []string{r.Symbol, r.Name, fmt.Sprintf("%.2f", r.Price), chgStr, formatVolume(r.Volume)}
				styles = []lipgloss.Style{symStyle, lipgloss.NewStyle(), lipgloss.NewStyle(), chgStyled, lipgloss.NewStyle()}
			} else if v.width < 115 {
				values = []string{r.Symbol, r.Name, r.Sector, fmt.Sprintf("%.2f", r.Price), fmt.Sprintf("%.2f", r.Spread), chgStr, formatVolume(r.Volume)}
				styles = []lipgloss.Style{symStyle, lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), chgStyled, lipgloss.NewStyle()}
			} else {
				values = []string{r.Symbol, r.Name, r.Sector, r.CapTier, fmt.Sprintf("%.2f", r.Price), bidAsk, fmt.Sprintf("%.2f", r.Spread), chgStr, formatVolume(r.Volume), fmt.Sprintf("%.1f", r.PE)}
				styles = []lipgloss.Style{symStyle, lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle(), chgStyled, lipgloss.NewStyle(), lipgloss.NewStyle()}
			}

			rowContent := FormatTableRow(cols, values, styles)
			formattedRow = rowStyle.Width(v.width - 6).Render(rowContent)
		}
		tableLines = append(tableLines, formattedRow)
	}

	if total == 0 {
		tableLines = append(tableLines, lipgloss.NewStyle().Foreground(ColorMuted).Padding(1, 2).Render("No matching enterprises found."))
	}

	tableBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(v.width - 4).
		Render(strings.Join(tableLines, "\n"))

	var footer string
	if v.width < 90 {
		footer = lipgloss.NewStyle().Foreground(ColorMuted).Render(
			" [j/k] Nav  •  [Enter] Chart  •  [/] Search  •  [s] Sector  •  [t] Tier",
		)
	} else {
		footer = lipgloss.NewStyle().Foreground(ColorMuted).Render(
			" [↑/↓ or j/k] Navigate  •  [Enter] View Chart  •  [PgUp/PgDn] Page  •  [/] Search  •  [s] Sector  •  [t] Tier",
		)
	}

	var elements []string
	if statusHeader != "" {
		elements = append(elements, statusHeader)
	}
	elements = append(elements, searchBar, filterBar, tableBox, footer)
	return lipgloss.JoinVertical(lipgloss.Left, elements...)
}

// HandleMouseClick handles clicking on rows or filter buttons in Universe tab
func (v *UniverseView) HandleMouseClick(x, y int, rows []CompanyRow) (selected *CompanyRow, action string) {
	// y=2 is searchBar, y=3 is filterBar
	if y == 3 {
		if x >= 10 && x <= 45 {
			return nil, "cycle_sector"
		}
		if x >= 46 && x <= 80 {
			return nil, "cycle_tier"
		}
	}

	// Table rows start at y=6 (after top header 0-1, search 2, filter 3, table border 4, column headers 5)
	pageSize := v.height - 10
	if pageSize < 5 {
		pageSize = 5
	}
	rowOnScreen := y - 6
	filtered := v.FilteredRows(rows)
	if rowOnScreen >= 0 && rowOnScreen < pageSize {
		targetIdx := v.scrollOffset + rowOnScreen
		if targetIdx >= 0 && targetIdx < len(filtered) {
			v.cursor = targetIdx
			return &filtered[targetIdx], "select_stock"
		}
	}
	return nil, ""
}
