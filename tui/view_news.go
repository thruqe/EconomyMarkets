package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type NewsView struct {
	width  int
	height int
}

func NewNewsView() *NewsView {
	return &NewsView{width: 120, height: 35}
}

func (v *NewsView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *NewsView) Render(newsEvents []string, tapeTrades []string) string {
	halfWidth := (v.width - 6) / 2
	boxWidth := halfWidth
	visibleLines := max(v.height-8, 6)

	if v.width < 90 {
		boxWidth = v.width - 4
		visibleLines = max((v.height-10)/2, 4)
	}

	var nLines []string
	nLines = append(nLines, StyleHeader.Render(fmt.Sprintf(" %-8s %-40s", "TIME/TICK", "HEADLINE & CORPORATE DISCLOSURE")))
	if len(newsEvents) == 0 {
		nLines = append(nLines, lipgloss.NewStyle().Foreground(ColorMuted).Padding(1, 2).Render("Listening for corporate news and regulatory disclosures..."))
	} else {
		start := max(0, len(newsEvents)-visibleLines)
		for i := len(newsEvents) - 1; i >= start; i-- {
			nLines = append(nLines, "  • "+truncate(newsEvents[i], boxWidth-6))
		}
	}

	newsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSecondary).
		Width(boxWidth).
		Height(visibleLines + 4).
		Render(
			fmt.Sprintf("%s\n%s",
				StyleBold.Render("  CORPORATE NEWS WIRE & DISCLOSURES"),
				strings.Join(nLines, "\n"),
			),
		)

	var tLines []string
	tLines = append(tLines, StyleHeader.Render(FormatTableHeader(TapeTradeCols)))
	if len(tapeTrades) == 0 {
		tLines = append(tLines, lipgloss.NewStyle().Foreground(ColorMuted).Padding(1, 2).Render("Listening for cleared exchange executions..."))
	} else {
		start := max(0, len(tapeTrades)-visibleLines)
		for i := len(tapeTrades) - 1; i >= start; i-- {
			tLines = append(tLines, tapeTrades[i])
		}
	}

	tapeBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Width(boxWidth).
		Height(visibleLines + 4).
		Render(
			fmt.Sprintf("%s\n%s",
				StyleBold.Render("  GLOBAL TIME & SALES EXECUTION TAPE"),
				strings.Join(tLines, "\n"),
			),
		)

	if v.width < 90 {
		return lipgloss.JoinVertical(lipgloss.Left, newsBox, "", tapeBox)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, newsBox, " ", tapeBox)
}
