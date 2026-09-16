package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type PaletteCommand struct {
	ID          string
	Title       string
	Category    string
	Description string
	Action      string
	Value       string
}

type CommandPalette struct {
	isOpen      bool
	query       string
	cursorIndex int
	width       int
	height      int
	commands    []PaletteCommand
	filtered    []PaletteCommand
	clickAreas  []ClickArea
}

func NewCommandPalette() *CommandPalette {
	cp := &CommandPalette{
		width: 70,
	}
	cp.initCommands()
	cp.filterCommands()
	return cp
}

func (p *CommandPalette) initCommands() {
	themeTitle := "Toggle Theme (Switch to Light Mode)"
	if CurrentTheme == ThemeLight {
		themeTitle = "Toggle Theme (Switch to Dark Mode)"
	}

	p.commands = []PaletteCommand{
		{ID: "toggle_theme", Title: themeTitle, Category: "Settings", Description: "Toggle between Dark and Light color themes", Action: "toggle_theme"},
		{ID: "theme_light", Title: "Theme: Light Mode", Category: "Settings", Description: "Switch to clean, high-contrast light theme", Action: "set_theme", Value: "light"},
		{ID: "theme_dark", Title: "Theme: Dark Mode", Category: "Settings", Description: "Switch to dark slate terminal theme", Action: "set_theme", Value: "dark"},

		// Navigation
		{ID: "nav_1", Title: "Go to: [1] Overview", Category: "Navigation", Description: "Market summaries, sovereign status & macro indicators", Action: "go_tab", Value: "0"},
		{ID: "nav_2", Title: "Go to: [2] 500 Stocks Universe", Category: "Navigation", Description: "Equity taxonomy, quotes and corporate search", Action: "go_tab", Value: "1"},
		{ID: "nav_3", Title: "Go to: [3] Candlestick Charts", Category: "Navigation", Description: "Multi-timeframe OHLC technical candlestick charts", Action: "go_tab", Value: "2"},
		{ID: "nav_4", Title: "Go to: [4] Deep Dive & DOM", Category: "Navigation", Description: "Full Level 2 orderbook depth and fundamental balance sheet", Action: "go_tab", Value: "3"},
		{ID: "nav_5", Title: "Go to: [5] Portfolio & Trade", Category: "Navigation", Description: "Order execution, open positions, margin & PnL", Action: "go_tab", Value: "4"},
		{ID: "nav_6", Title: "Go to: [6] News & Wire Tape", Category: "Navigation", Description: "Realtime news stream, corporate filings and tape", Action: "go_tab", Value: "5"},
		{ID: "nav_7", Title: "Go to: [7] Sovereign & Admin Desk", Category: "Navigation", Description: "Treasury debt, capital investments, bailouts & IPO listing", Action: "go_tab", Value: "6"},
		{ID: "nav_8", Title: "Go to: [8] World & Nations", Category: "Navigation", Description: "Global map, foreign trade treaties, aid & sanctions", Action: "go_tab", Value: "7"},
		{ID: "nav_9", Title: "Go to: [9] Policy Studio", Category: "Navigation", Description: "Executive fiscal, monetary, trade, regulatory & labor policies", Action: "go_tab", Value: "8"},

		// Simulation Speeds
		{ID: "speed_0", Title: "Simulation: Pause (0x)", Category: "Simulation", Description: "Pause simulated market clock", Action: "set_speed", Value: "0"},
		{ID: "speed_1", Title: "Simulation: 1x Speed", Category: "Simulation", Description: "Standard pace (4 ticks = 1 day)", Action: "set_speed", Value: "1"},
		{ID: "speed_5", Title: "Simulation: 5x Fast", Category: "Simulation", Description: "Accelerated market clock", Action: "set_speed", Value: "5"},
		{ID: "speed_12", Title: "Simulation: 12x Day/sec", Category: "Simulation", Description: "1 day per second high speed", Action: "set_speed", Value: "12"},
		{ID: "speed_60", Title: "Simulation: 60x Ultra Speed", Category: "Simulation", Description: "5 days per second macroeconomic accelerator", Action: "set_speed", Value: "60"},

		// Persistence / Maintenance
		{ID: "save_state", Title: "Save World State", Category: "Data", Description: "Snapshot all macro, country, and stock states to disk", Action: "save_state"},
		{ID: "save_portfolio", Title: "Save Portfolio", Category: "Data", Description: "Save human trading balance and positions to disk", Action: "save_portfolio"},
		{ID: "reset_portfolio", Title: "Reset Portfolio", Category: "Data", Description: "Clear holdings and reset cash to $100,000", Action: "reset_portfolio"},
		{ID: "reset_world", Title: "Reset World State", Category: "Data", Description: "Wipe world save and generate fresh nation from tick 0", Action: "reset_world"},
	}
}

func (p *CommandPalette) Toggle() {
	p.isOpen = !p.isOpen
	if p.isOpen {
		p.query = ""
		p.cursorIndex = 0
		p.initCommands()
		p.filterCommands()
	}
}

func (p *CommandPalette) IsOpen() bool {
	return p.isOpen
}

func (p *CommandPalette) Close() {
	p.isOpen = false
}

func (p *CommandPalette) MoveUp() {
	if p.cursorIndex > 0 {
		p.cursorIndex--
	}
}

func (p *CommandPalette) MoveDown() {
	if p.cursorIndex < len(p.filtered)-1 {
		p.cursorIndex++
	}
}

func (p *CommandPalette) SetQuery(q string) {
	p.query = q
	p.cursorIndex = 0
	p.filterCommands()
}

func (p *CommandPalette) AppendQuery(r rune) {
	p.query += string(r)
	p.cursorIndex = 0
	p.filterCommands()
}

func (p *CommandPalette) BackspaceQuery() {
	if len(p.query) > 0 {
		p.query = p.query[:len(p.query)-1]
		p.cursorIndex = 0
		p.filterCommands()
	}
}

func (p *CommandPalette) SelectedCommand() *PaletteCommand {
	if len(p.filtered) == 0 || p.cursorIndex < 0 || p.cursorIndex >= len(p.filtered) {
		return nil
	}
	return &p.filtered[p.cursorIndex]
}

func (p *CommandPalette) filterCommands() {
	q := strings.ToLower(strings.TrimSpace(p.query))
	if q == "" {
		p.filtered = make([]PaletteCommand, len(p.commands))
		copy(p.filtered, p.commands)
		return
	}

	p.filtered = nil
	for _, cmd := range p.commands {
		text := strings.ToLower(cmd.Title + " " + cmd.Category + " " + cmd.Description)
		if strings.Contains(text, q) {
			p.filtered = append(p.filtered, cmd)
		}
	}
}

func (p *CommandPalette) HandleMouseClick(x, y int) *PaletteCommand {
	for _, ca := range p.clickAreas {
		if x >= ca.X1 && x <= ca.X2 && y >= ca.Y1 && y <= ca.Y2 {
			idx := int(ca.Speed)
			if idx >= 0 && idx < len(p.filtered) {
				return &p.filtered[idx]
			}
		}
	}
	return nil
}

func (p *CommandPalette) Render(screenWidth, screenHeight int) string {
	p.clickAreas = p.clickAreas[:0]

	modalWidth := 74
	if modalWidth > screenWidth-4 {
		modalWidth = screenWidth - 4
	}
	if modalWidth < 40 {
		modalWidth = 40
	}

	titleBar := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render("  COMMAND PALETTE  [Ctrl+P / Esc to Close]")

	searchPrompt := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight).
		Render(" > ")

	queryDisp := p.query
	if queryDisp == "" {
		queryDisp = lipgloss.NewStyle().Foreground(ColorMuted).Render("Type to search commands or settings (e.g. 'theme', 'stocks', 'reset')...")
	} else {
		queryDisp = lipgloss.NewStyle().Bold(true).Foreground(ColorTextBold).Render(queryDisp)
	}

	inputLine := searchPrompt + queryDisp
	divider := lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", modalWidth-4))

	maxVisible := 10
	startIdx := 0
	if p.cursorIndex >= maxVisible {
		startIdx = p.cursorIndex - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(p.filtered) {
		endIdx = len(p.filtered)
	}

	var itemLines []string
	startX := (screenWidth - modalWidth) / 2
	startY := (screenHeight - 16) / 2
	if startY < 2 {
		startY = 2
	}

	if len(p.filtered) == 0 {
		itemLines = append(itemLines, lipgloss.NewStyle().Foreground(ColorMuted).Padding(1, 2).Render("No matching commands found."))
	} else {
		for i := startIdx; i < endIdx; i++ {
			cmd := p.filtered[i]
			prefix := "   "
			var rowStyle lipgloss.Style
			if i == p.cursorIndex {
				prefix = " ▶ "
				rowStyle = lipgloss.NewStyle().
					Background(ColorHighlight).
					Foreground(lipgloss.Color("#ffffff")).
					Bold(true).
					Width(modalWidth - 6)
			} else {
				rowStyle = lipgloss.NewStyle().
					Foreground(ColorText).
					Width(modalWidth - 6)
			}

			catBadge := lipgloss.NewStyle().
				Foreground(ColorMuted).
				Render(fmt.Sprintf("[%s]", cmd.Category))

			titlePart := fmt.Sprintf("%-38s", cmd.Title)
			if len(titlePart) > 38 {
				titlePart = titlePart[:35] + "..."
			}

			lineContent := fmt.Sprintf("%s%-38s  %s", prefix, titlePart, catBadge)
			itemLines = append(itemLines, rowStyle.Render(lineContent))

			rowY := startY + 3 + (i - startIdx)
			p.clickAreas = append(p.clickAreas, ClickArea{
				X1:     startX + 2,
				Y1:     rowY,
				X2:     startX + modalWidth - 2,
				Y2:     rowY,
				Action: "palette_item",
				Speed:  float64(i),
			})
		}
	}

	footer := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Padding(0, 1).
		Render("  [↑/↓] Navigate   [Enter/Click] Execute   [Esc] Close")

	bodyParts := []string{
		titleBar,
		inputLine,
		divider,
	}
	bodyParts = append(bodyParts, strings.Join(itemLines, "\n"))
	bodyParts = append(bodyParts, divider, footer)

	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorHighlight).
		Background(ColorCardBg).
		Padding(1, 1).
		Width(modalWidth).
		Render(strings.Join(bodyParts, "\n"))

	return modalBox
}
