package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"economy/world"
)

type WorldView struct {
	width          int
	height         int
	clickAreas     []ClickArea
	statusMsg      string
	isStatusErr    bool
	scrollOffset   int
	// Setup fields for blank slate country configuration
	SetupName      string
	SetupCurrency  string
	SetupFlag      string
	SetupFounded   string
	ActiveField    int
}

func NewWorldView() *WorldView {
	return &WorldView{
		width:         120,
		height:        35,
		SetupName:     "Republic of Aurora",
		SetupCurrency: "AUR",
		SetupFlag:     "AU",
		SetupFounded:  "2026",
		ActiveField:   0,
	}
}

func (v *WorldView) SetSize(w, h int) {
	v.width = w
	v.height = h
}

func (v *WorldView) SetStatus(msg string, isErr bool) {
	v.statusMsg = msg
	v.isStatusErr = isErr
}

// Render draws the World & Nations tab.
func (v *WorldView) Render(w *world.World, simDate world.SimDate) string {
	v.clickAreas = v.clickAreas[:0]

	if w == nil {
		return StyleMuted.Render("  World simulation not initialized.")
	}

	// If home country is not configured, render the Country Setup Wizard
	if !w.Home.IsConfigured() {
		return v.renderCountrySetup(w)
	}

	// ----- Home Country Section -----
	homeWidth := v.width/2 - 4
	regionWidth := v.width/2 - 2
	if v.width < 100 {
		homeWidth = v.width - 4
		regionWidth = v.width - 4
	}

	homeLines := buildHomeCountryLines(w, simDate)
	homeBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Width(homeWidth).
		Padding(0, 1).
		Render(fmt.Sprintf("%s\n\n%s",
			StyleTitle.Render("HOME COUNTRY: "+w.Home.Name),
			strings.Join(homeLines, "\n"),
		))

	// ----- Regions Section -----
	regionLines := buildRegionLines(w)
	regionBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSecondary).
		Width(regionWidth).
		Padding(0, 1).
		Render(fmt.Sprintf("%s\n\n%s",
			StyleTitle.Render("DOMESTIC REGIONS (5 ZONES)"),
			strings.Join(regionLines, "\n"),
		))

	var topRow string
	if v.width < 100 {
		topRow = lipgloss.JoinVertical(lipgloss.Left, homeBox, "", regionBox)
	} else {
		topRow = lipgloss.JoinHorizontal(lipgloss.Top, homeBox, " ", regionBox)
	}

	// ----- Foreign Relations Section -----
	foreignLines := buildForeignLines(w, v)
	foreignBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Width(v.width - 4).
		Padding(0, 1).
		Render(fmt.Sprintf("%s\n\n%s",
			StyleTitle.Render("FOREIGN RELATIONS & DIPLOMATIC STANDING"),
			strings.Join(foreignLines, "\n"),
		))

	// ----- Policy Log Section -----
	logLines := buildPolicyLogLines(w)
	logBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Width(v.width - 4).
		Padding(0, 1).
		Render(fmt.Sprintf("%s\n\n%s",
			StyleBold.Render("FOREIGN POLICY ACTION LOG"),
			strings.Join(logLines, "\n"),
		))

	// Status bar
	var statusLine string
	if v.statusMsg != "" {
		if v.isStatusErr {
			statusLine = "\n" + StyleRed.Render("  [ERROR] "+v.statusMsg)
		} else {
			statusLine = "\n" + StyleGreen.Render("  [OK] "+v.statusMsg)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		topRow,
		"",
		foreignBox,
		"",
		logBox,
		statusLine,
	)
}

func buildHomeCountryLines(w *world.World, simDate world.SimDate) []string {
	h := &w.Home
	if !h.IsConfigured() {
		return []string{
			StyleMuted.Render("  Home country not configured yet."),
			"",
			StyleMuted.Render("  Press [C] to set up your country name, currency, and flag."),
		}
	}
	flag := h.FlagCode
	if flag == "" {
		flag = "--"
	}
	return []string{
		fmt.Sprintf("  Name      : %s", StyleBold.Render(h.Name)),
		fmt.Sprintf("  Currency  : %s", StyleBold.Render(h.Currency)),
		fmt.Sprintf("  Flag Code : %s", StyleCyan.Render(flag)),
		fmt.Sprintf("  Founded   : %s", StyleBold.Render(fmt.Sprintf("%d", h.Founded))),
		fmt.Sprintf("  Sim Date  : %s", StyleGreen.Render(simDate.String())),
		"",
		fmt.Sprintf("  Regions   : %d domestic economic zones", len(h.Regions)),
	}
}

func buildRegionLines(w *world.World) []string {
	if len(w.Home.Regions) == 0 {
		return []string{StyleMuted.Render("  No regions defined.")}
	}
	regionCols := []TableColumn{
		{Title: "REGION", Width: 12, AlignRight: false},
		{Title: "GDP", Width: 10, AlignRight: true},
		{Title: "TAX REV", Width: 10, AlignRight: true},
		{Title: "UNEMP", Width: 8, AlignRight: true},
		{Title: "POPULATION", Width: 10, AlignRight: true},
	}
	lines := []string{StyleHeader.Render(FormatTableHeader(regionCols))}
	for _, r := range w.Home.Regions {
		values := []string{
			r.Name,
			FormatCurrencyShort(r.GDP),
			FormatCurrencyShort(r.TaxRevenue),
			fmt.Sprintf("%.1f%%", r.UnemploymentRate*100),
			FormatLargeNumber(r.Population),
		}
		lines = append(lines, FormatTableRow(regionCols, values, nil))
	}
	return lines
}

func buildForeignLines(w *world.World, v *WorldView) []string {
	countries := w.ForeignCountriesSorted()
	var foreignCols []TableColumn
	if v.width < 95 {
		foreignCols = []TableColumn{
			{Title: "COUNTRY", Width: 14, AlignRight: false},
			{Title: "GDP", Width: 8, AlignRight: true},
			{Title: "GROWTH", Width: 8, AlignRight: true},
			{Title: "TARIFF", Width: 8, AlignRight: true},
			{Title: "STANCE", Width: 10, AlignRight: false},
			{Title: "SANCTIONS", Width: 11, AlignRight: false},
		}
	} else {
		foreignCols = []TableColumn{
			{Title: "COUNTRY", Width: 18, AlignRight: false},
			{Title: "GDP", Width: 9, AlignRight: true},
			{Title: "GROWTH", Width: 9, AlignRight: true},
			{Title: "INFLATION", Width: 9, AlignRight: true},
			{Title: "TARIFF", Width: 8, AlignRight: true},
			{Title: "TRADE VOL", Width: 11, AlignRight: true},
			{Title: "STANCE", Width: 10, AlignRight: false},
			{Title: "SANCTIONS", Width: 11, AlignRight: false},
		}
	}
	lines := []string{StyleHeader.Render(FormatTableHeader(foreignCols))}

	// Foreign relations start after: header=2 rows, top section, empty, foreignBox border + title + blank + header = 4 rows
	topRowHeight := 0
	if len(w.Home.Regions) > 0 {
		topRowHeight = 10 + len(w.Home.Regions)
	} else {
		topRowHeight = 12
	}
	foreignStartY := 2 + topRowHeight + 5

	for i, fc := range countries {
		stanceStyle := StyleMuted
		switch fc.Relation.Stance {
		case world.StanceAllied:
			stanceStyle = StyleGreen
		case world.StanceFriendly:
			stanceStyle = StyleCyan
		case world.StanceCold:
			stanceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f0883e"))
		case world.StanceHostile:
			stanceStyle = StyleRed
		}

		sanctionLabel := "-"
		sanctionStyle := StyleMuted
		if fc.Relation.SanctionsLevel >= 2 {
			sanctionLabel = "SANCTIONED"
			sanctionStyle = StyleRed
		} else if fc.Relation.SanctionsLevel == 1 {
			sanctionLabel = "PARTIAL"
			sanctionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f0883e"))
		}

		growthStyle := StyleGreen
		if fc.GDPGrowth < 0 {
			growthStyle = StyleRed
		}

		dealBtn := lipgloss.NewStyle().Background(lipgloss.Color("#1f6feb")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render("[Deal]")
		aidBtn := lipgloss.NewStyle().Background(lipgloss.Color("#238636")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render("[Aid]")
		sanctionBtn := lipgloss.NewStyle().Background(lipgloss.Color("#da3633")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render("[Sanction]")

		var values []string
		var styles []lipgloss.Style
		if v.width < 95 {
			values = []string{
				truncate(fc.Name, 14),
				FormatCurrencyShort(fc.GDP),
				fmt.Sprintf("%+.1f%%", fc.GDPGrowth*100),
				fmt.Sprintf("%.1f%%", fc.Relation.TariffRate*100),
				fc.Relation.Stance.String(),
				sanctionLabel,
			}
			styles = []lipgloss.Style{
				lipgloss.NewStyle(),
				lipgloss.NewStyle(),
				growthStyle,
				lipgloss.NewStyle(),
				stanceStyle,
				sanctionStyle,
			}
		} else {
			values = []string{
				fc.Name,
				FormatCurrencyShort(fc.GDP),
				fmt.Sprintf("%+.1f%%", fc.GDPGrowth*100),
				fmt.Sprintf("%.1f%%", fc.Inflation*100),
				fmt.Sprintf("%.1f%%", fc.Relation.TariffRate*100),
				FormatCurrencyShort(fc.Relation.TradeVolume),
				fc.Relation.Stance.String(),
				sanctionLabel,
			}
			styles = []lipgloss.Style{
				lipgloss.NewStyle(),
				lipgloss.NewStyle(),
				growthStyle,
				lipgloss.NewStyle(),
				lipgloss.NewStyle(),
				lipgloss.NewStyle(),
				stanceStyle,
				sanctionStyle,
			}
		}
		dataLine := FormatTableRow(foreignCols, values, styles)
		actionLine := fmt.Sprintf("   %s  %s  %s", dealBtn, aidBtn, sanctionBtn)
		lines = append(lines, dataLine, actionLine)

		// Register click areas for action buttons
		btnY := foreignStartY + 2 + i*2 + 1
		dealX := 4
		aidX := dealX + 8
		sanctionX := aidX + 7
		v.clickAreas = append(v.clickAreas,
			ClickArea{X1: dealX, Y1: btnY - 1, X2: dealX + 7, Y2: btnY + 1, Action: "negotiate_deal", Value: fc.ID},
			ClickArea{X1: aidX, Y1: btnY - 1, X2: aidX + 6, Y2: btnY + 1, Action: "send_aid", Value: fc.ID},
			ClickArea{X1: sanctionX, Y1: btnY - 1, X2: sanctionX + 11, Y2: btnY + 1, Action: "impose_sanction", Value: fc.ID},
		)
	}
	return lines
}

func buildPolicyLogLines(w *world.World) []string {
	if len(w.PolicyLog) == 0 {
		return []string{StyleMuted.Render("  No foreign policy actions taken yet. Use the action buttons above.")}
	}
	var lines []string
	start := len(w.PolicyLog) - 8
	if start < 0 {
		start = 0
	}
	for _, entry := range w.PolicyLog[start:] {
		line := fmt.Sprintf("  [Tick %d] %-12s %s", entry.Tick, entry.Action, entry.Detail)
		lines = append(lines, line)
	}
	return lines
}

var sampleCountries = []struct{ name, curr, flag string }{
	{"Republic of Aurora", "AUR", "AU"},
	{"Commonwealth of Solaria", "SOL", "SL"},
	{"United Federation of Zephyr", "ZPH", "ZF"},
	{"Free State of Pacifica", "PAC", "PC"},
	{"Nordic Sovereignty of Borealis", "BOR", "BL"},
	{"Meridian Republic", "MER", "MR"},
}

func (v *WorldView) randomizeCountry() {
	c := sampleCountries[time.Now().UnixNano()%int64(len(sampleCountries))]
	v.SetupName = c.name
	v.SetupCurrency = c.curr
	v.SetupFlag = c.flag
	v.SetupFounded = "2026"
}

func (v *WorldView) renderCountrySetup(w *world.World) string {
	fieldStyle := func(idx int) lipgloss.Style {
		if v.ActiveField == idx {
			return lipgloss.NewStyle().Background(ColorHighlight).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1)
		}
		return lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(lipgloss.Color("#c9d1d9")).Padding(0, 1)
	}

	nameField := fmt.Sprintf("  Country Name : %s", fieldStyle(0).Render(" "+v.SetupName+" "))
	currField := fmt.Sprintf("  Currency     : %s", fieldStyle(1).Render(" "+v.SetupCurrency+" "))
	flagField := fmt.Sprintf("  Flag Code    : %s", fieldStyle(2).Render(" "+v.SetupFlag+" "))
	fndField  := fmt.Sprintf("  Founded Year : %s", fieldStyle(3).Render(" "+v.SetupFounded+" "))

	initBtn := lipgloss.NewStyle().Background(ColorSuccess).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 3).Render("[ FOUND SOVEREIGN COUNTRY & ENTER WORLD (Enter) ]")
	presetBtn := lipgloss.NewStyle().Background(lipgloss.Color("#1f6feb")).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 2).Render("[ RANDOMIZE NATION IDEA ]")

	v.clickAreas = append(v.clickAreas,
		ClickArea{X1: 4, Y1: 6, X2: 45, Y2: 8, Action: "focus_setup_field", Speed: 0},
		ClickArea{X1: 4, Y1: 8, X2: 30, Y2: 10, Action: "focus_setup_field", Speed: 1},
		ClickArea{X1: 4, Y1: 10, X2: 30, Y2: 12, Action: "focus_setup_field", Speed: 2},
		ClickArea{X1: 4, Y1: 12, X2: 30, Y2: 14, Action: "focus_setup_field", Speed: 3},
		ClickArea{X1: 4, Y1: 15, X2: 35, Y2: 17, Action: "randomize_country"},
		ClickArea{X1: 36, Y1: 15, X2: 85, Y2: 17, Action: "confirm_country"},
	)

	statusLine := ""
	if v.statusMsg != "" {
		if v.isStatusErr {
			statusLine = "\n  " + StyleRed.Render("[ERROR] "+v.statusMsg)
		} else {
			statusLine = "\n  " + StyleGreen.Render("[OK] "+v.statusMsg)
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Width(v.width - 4).
		Padding(1, 2).
		Render(
			fmt.Sprintf("%s\n\n%s\n\n%s\n%s\n%s\n%s\n\n%s  %s%s\n\n%s",
				StyleTitle.Render("SOVEREIGN STATE CREATION WIZARD (CUSTOM ECONOMY)"),
				StyleMuted.Render("Design and establish your own nation from a blank slate. You govern policies, regional investment, and foreign relations."),
				nameField,
				currField,
				flagField,
				fndField,
				presetBtn,
				initBtn,
				statusLine,
				StyleMuted.Render("Controls: Click any field or button  •  [Tab] Next field  •  [Enter] Confirm & Found Country"),
			),
		)
}

// HandleMouseClick processes clicks in the world view.
func (v *WorldView) HandleMouseClick(x, y int) (action, value string) {
	for _, a := range v.clickAreas {
		if x >= a.X1 && x <= a.X2 && y >= a.Y1 && y <= a.Y2 {
			switch a.Action {
			case "focus_setup_field":
				v.ActiveField = int(a.Speed)
				return a.Action, a.Value
			case "randomize_country":
				v.randomizeCountry()
				return a.Action, a.Value
			default:
				return a.Action, a.Value
			}
		}
	}
	return "", ""
}


// FormatCurrencyShort formats large currency values with T/B/M suffixes.
func FormatCurrencyShort(v float64) string {
	switch {
	case v >= 1e12:
		return fmt.Sprintf("$%.2fT", v/1e12)
	case v >= 1e9:
		return fmt.Sprintf("$%.1fB", v/1e9)
	case v >= 1e6:
		return fmt.Sprintf("$%.1fM", v/1e6)
	default:
		return fmt.Sprintf("$%.0f", v)
	}
}

// FormatLargeNumber formats a float as a compact number (M/B).
func FormatLargeNumber(v float64) string {
	switch {
	case v >= 1e9:
		return fmt.Sprintf("%.2fB", v/1e9)
	case v >= 1e6:
		return fmt.Sprintf("%.1fM", v/1e6)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}
