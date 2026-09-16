package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TableColumn specifies the title, display width, and alignment of a table column.
type TableColumn struct {
	Title      string
	Width      int
	AlignRight bool
}

// FormatCell formats a single cell string to exact visual width width.
// It accounts for existing ANSI styling and wide runes, truncates cleanly if exceeding width,
// pads with spaces (left for right-aligned, right for left-aligned),
// and applies the given lipgloss style to the padded block.
func FormatCell(text string, width int, alignRight bool, style lipgloss.Style) string {
	visW := lipgloss.Width(text)
	if visW > width {
		text = truncate(text, width)
		visW = lipgloss.Width(text)
	}
	pad := width - visW
	if pad < 0 {
		pad = 0
	}

	var padded string
	if alignRight {
		padded = strings.Repeat(" ", pad) + text
	} else {
		padded = text + strings.Repeat(" ", pad)
	}

	// Only invoke style.Render if style has attributes to avoid extra empty ANSI escapes
	if style.GetBold() || style.GetForeground() != nil || style.GetBackground() != nil || style.GetUnderline() {
		return style.Render(padded)
	}
	return padded
}

// FormatTableRow formats an entire row from a list of TableColumn specs, string values, and cell styles.
// The resulting row has guaranteed identical visual column offsets to FormatTableHeader.
func FormatTableRow(cols []TableColumn, values []string, styles []lipgloss.Style) string {
	cells := make([]string, len(cols))
	for i, col := range cols {
		val := ""
		if i < len(values) {
			val = values[i]
		}
		var st lipgloss.Style
		if i < len(styles) {
			st = styles[i]
		}
		cells[i] = FormatCell(val, col.Width, col.AlignRight, st)
	}
	return " " + strings.Join(cells, " ")
}

// FormatTableHeader formats the header string for a list of TableColumn specs.
func FormatTableHeader(cols []TableColumn) string {
	cells := make([]string, len(cols))
	for i, col := range cols {
		cells[i] = FormatCell(col.Title, col.Width, col.AlignRight, lipgloss.NewStyle())
	}
	return " " + strings.Join(cells, " ")
}

// TapeTradeCols defines the standard columns for executed trades on the tape.
var TapeTradeCols = []TableColumn{
	{Title: "SYM", Width: 6, AlignRight: false},
	{Title: "SIDE", Width: 5, AlignRight: false},
	{Title: "QTY", Width: 7, AlignRight: true},
	{Title: "PRICE", Width: 9, AlignRight: true},
	{Title: "TAKER", Width: 12, AlignRight: false},
}

// FormatTapeTrade formats an executed trade row consistently with TapeTradeCols.
func FormatTapeTrade(sym, side string, qty, price float64, taker string) string {
	sideStyle := StyleGreen
	if strings.EqualFold(side, "SELL") || strings.EqualFold(side, "SHORT") {
		sideStyle = StyleRed
	}
	values := []string{
		sym,
		side,
		fmt.Sprintf("%.0f", qty),
		fmt.Sprintf("%.2f", price),
		taker,
	}
	styles := []lipgloss.Style{
		StyleBold,
		sideStyle,
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		StyleMuted,
	}
	return FormatTableRow(TapeTradeCols, values, styles)
}

