package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestTableAlignmentNoAnsiDrift(t *testing.T) {
	cols := []TableColumn{
		{Title: "SYM", Width: 6, AlignRight: false},
		{Title: "NAME", Width: 20, AlignRight: false},
		{Title: "PRICE", Width: 10, AlignRight: true},
		{Title: "CHG %", Width: 10, AlignRight: true},
		{Title: "STATUS", Width: 12, AlignRight: false},
	}

	header := FormatTableHeader(cols)
	headerWidth := lipgloss.Width(header)

	testCases := []struct {
		name   string
		values []string
		styles []lipgloss.Style
	}{
		{
			name:   "plain values without styling",
			values: []string{"AAPL", "Apple Inc.", "182.50", "+1.25%", "Active"},
			styles: nil,
		},
		{
			name:   "styled values with bold and colors",
			values: []string{"AAPL", "Apple Inc.", "182.50", "+1.25%", "Active"},
			styles: []lipgloss.Style{
				StyleBold,
				lipgloss.NewStyle(),
				lipgloss.NewStyle(),
				StyleGreen,
				StyleCyan,
			},
		},
		{
			name:   "styled negative values with red",
			values: []string{"TSLA", "Tesla Inc.", "195.10", "-4.80%", "Halted"},
			styles: []lipgloss.Style{
				StyleBold,
				lipgloss.NewStyle(),
				lipgloss.NewStyle(),
				StyleRed,
				StyleRed,
			},
		},
		{
			name:   "short values",
			values: []string{"F", "Ford", "12.00", "0.00%", "OK"},
			styles: []lipgloss.Style{
				StyleBold,
				lipgloss.NewStyle(),
				lipgloss.NewStyle(),
				lipgloss.NewStyle(),
				lipgloss.NewStyle(),
			},
		},
		{
			name:   "values exceeding column widths (truncated)",
			values: []string{"VERYLONGSYMBOL", "Super Long Company Name That Exceeds Width", "123456789.99", "+1234.56%", "SuperLongStatus"},
			styles: []lipgloss.Style{
				StyleBold,
				lipgloss.NewStyle(),
				lipgloss.NewStyle(),
				StyleGreen,
				StyleRed,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			row := FormatTableRow(cols, tc.values, tc.styles)
			rowWidth := lipgloss.Width(row)

			if rowWidth != headerWidth {
				t.Errorf("Visual width mismatch: header=%d, row=%d\nHeader: %q\nRow:    %q",
					headerWidth, rowWidth, header, row)
			}

			// Verify selected row with background styling preserves width
			selectedRow := StyleTableRowSelected.Render(row)
			selectedWidth := lipgloss.Width(selectedRow)
			if selectedWidth != headerWidth {
				t.Errorf("Selected row visual width mismatch: header=%d, selectedRow=%d",
					headerWidth, selectedWidth)
			}
		})
	}
}

func TestTapeTradeAlignment(t *testing.T) {
	header := FormatTableHeader(TapeTradeCols)
	headerWidth := lipgloss.Width(header)

	trades := []struct {
		sym   string
		side  string
		qty   float64
		price float64
		taker string
	}{
		{"AAPL", "BUY", 100, 182.50, "Retail#1"},
		{"MSFT", "SELL", 5000, 420.15, "Inst#99"},
		{"NVDA", "SHORT", 25, 890.00, "MM#4"},
		{"AMZN", "BUY", 1, 175.20, "Algo#12"},
	}

	for _, tr := range trades {
		row := FormatTapeTrade(tr.sym, tr.side, tr.qty, tr.price, tr.taker)
		rowWidth := lipgloss.Width(row)
		if rowWidth != headerWidth {
			t.Errorf("Tape trade visual width mismatch: header=%d, trade=%d\nHeader: %q\nTrade:  %q",
				headerWidth, rowWidth, header, row)
		}
	}
}

func TestAllTableColumnSpecs(t *testing.T) {
	specs := []struct {
		name string
		cols []TableColumn
	}{
		{
			name: "DOM columns",
			cols: []TableColumn{
				{Title: "SIDE", Width: 6, AlignRight: false},
				{Title: "PRICE", Width: 10, AlignRight: true},
				{Title: "SIZE", Width: 10, AlignRight: true},
				{Title: "DEPTH", Width: 16, AlignRight: false},
			},
		},
		{
			name: "Domestic Regions",
			cols: []TableColumn{
				{Title: "REGION", Width: 12, AlignRight: false},
				{Title: "GDP", Width: 10, AlignRight: true},
				{Title: "TAX REV", Width: 10, AlignRight: true},
				{Title: "UNEMP", Width: 8, AlignRight: true},
				{Title: "POPULATION", Width: 10, AlignRight: true},
			},
		},
		{
			name: "Foreign Relations (Compact)",
			cols: []TableColumn{
				{Title: "COUNTRY", Width: 14, AlignRight: false},
				{Title: "GDP", Width: 8, AlignRight: true},
				{Title: "GROWTH", Width: 8, AlignRight: true},
				{Title: "TARIFF", Width: 8, AlignRight: true},
				{Title: "STANCE", Width: 10, AlignRight: false},
				{Title: "SANCTIONS", Width: 11, AlignRight: false},
			},
		},
		{
			name: "Foreign Relations (Wide)",
			cols: []TableColumn{
				{Title: "COUNTRY", Width: 18, AlignRight: false},
				{Title: "GDP", Width: 9, AlignRight: true},
				{Title: "GROWTH", Width: 9, AlignRight: true},
				{Title: "INFLATION", Width: 9, AlignRight: true},
				{Title: "TARIFF", Width: 8, AlignRight: true},
				{Title: "TRADE VOL", Width: 11, AlignRight: true},
				{Title: "STANCE", Width: 10, AlignRight: false},
				{Title: "SANCTIONS", Width: 11, AlignRight: false},
			},
		},
		{
			name: "Admin Bailout",
			cols: []TableColumn{
				{Title: "SYM", Width: 6, AlignRight: false},
				{Title: "NAME", Width: 22, AlignRight: false},
				{Title: "SECTOR", Width: 18, AlignRight: false},
				{Title: "PRICE", Width: 10, AlignRight: true},
				{Title: "DEBT", Width: 14, AlignRight: true},
				{Title: "EXECUTIVE ACTION", Width: 26, AlignRight: false},
			},
		},
	}

	for _, s := range specs {
		t.Run(s.name, func(t *testing.T) {
			hdr := FormatTableHeader(s.cols)
			hdrWidth := lipgloss.Width(hdr)

			vals := make([]string, len(s.cols))
			styles := make([]lipgloss.Style, len(s.cols))
			for i, col := range s.cols {
				if col.AlignRight {
					vals[i] = fmt.Sprintf("%d.00", (i+1)*10)
				} else {
					vals[i] = fmt.Sprintf("Val%d", i+1)
				}
				if i%2 == 0 {
					styles[i] = StyleBold
				} else {
					styles[i] = StyleGreen
				}
			}

			row := FormatTableRow(s.cols, vals, styles)
			rowWidth := lipgloss.Width(row)

			if hdrWidth != rowWidth {
				t.Fatalf("[%s] width mismatch: header=%d, row=%d", s.name, hdrWidth, rowWidth)
			}
		})
	}
}

func TestFormatCellRuneSafety(t *testing.T) {
	tests := []struct {
		input      string
		width      int
		alignRight bool
		expectedW  int
	}{
		{"Hello World", 5, false, 5},
		{"ASCII", 10, false, 10},
		{"123.45", 10, true, 10},
		{"€100", 8, true, 8},
		{"", 5, false, 5},
	}

	for _, tt := range tests {
		res := FormatCell(tt.input, tt.width, tt.alignRight, lipgloss.NewStyle())
		visW := lipgloss.Width(res)
		if visW != tt.expectedW {
			t.Errorf("FormatCell(%q, %d, %v) visual width = %d, expected %d",
				tt.input, tt.width, tt.alignRight, visW, tt.expectedW)
		}
		if tt.alignRight && !strings.HasSuffix(res, strings.TrimSpace(tt.input)) && len(strings.TrimSpace(tt.input)) <= tt.width {
			t.Errorf("Right alignment failed for %q in %q", tt.input, res)
		}
	}
}
