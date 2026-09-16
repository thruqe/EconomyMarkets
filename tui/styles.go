package tui

import (
	"github.com/charmbracelet/lipgloss"
)

type ThemeMode string

const (
	ThemeDark  ThemeMode = "dark"
	ThemeLight ThemeMode = "light"
)

var CurrentTheme = ThemeDark

var (
	// Brand Colors
	ColorPrimary   lipgloss.TerminalColor
	ColorSecondary lipgloss.TerminalColor
	ColorSuccess   lipgloss.TerminalColor
	ColorDanger    lipgloss.TerminalColor
	ColorWarning   lipgloss.TerminalColor
	ColorMuted     lipgloss.TerminalColor
	ColorBg        lipgloss.TerminalColor
	ColorCardBg    lipgloss.TerminalColor
	ColorBorder    lipgloss.TerminalColor
	ColorHighlight lipgloss.TerminalColor
	ColorText      lipgloss.TerminalColor
	ColorTextBold  lipgloss.TerminalColor
	ColorBtnBg     lipgloss.TerminalColor

	// Styles
	StyleTitle            lipgloss.Style
	StyleHeader           lipgloss.Style
	StyleTabActive        lipgloss.Style
	StyleTabInactive      lipgloss.Style
	StyleCard             lipgloss.Style
	StyleGreen            lipgloss.Style
	StyleRed              lipgloss.Style
	StyleWarning          lipgloss.Style
	StyleCyan             lipgloss.Style
	StyleMuted            lipgloss.Style
	StyleBold             lipgloss.Style
	StyleTableRowSelected lipgloss.Style
	StyleTableRow         lipgloss.Style
	StyleStatusBar        lipgloss.Style
	StyleBadge            lipgloss.Style
)

func init() {
	ApplyTheme(ThemeDark)
}

func ApplyTheme(mode ThemeMode) {
	CurrentTheme = mode
	if mode == ThemeLight {
		ColorPrimary = lipgloss.Color("#0969da")   // Classic Blue
		ColorSecondary = lipgloss.Color("#8250df") // Purple
		ColorSuccess = lipgloss.Color("#1a7f37")   // Forest Green
		ColorDanger = lipgloss.Color("#cf222e")    // Crimson Red
		ColorWarning = lipgloss.Color("#9a6700")   // Amber/Brown
		ColorMuted = lipgloss.Color("#656d76")     // Medium Gray
		ColorBg = lipgloss.Color("#f6f8fa")        // Clean light background
		ColorCardBg = lipgloss.Color("#ffffff")    // White card background
		ColorBorder = lipgloss.Color("#d0d7de")    // Light gray border
		ColorHighlight = lipgloss.Color("#0969da") // Blue selection
		ColorText = lipgloss.Color("#24292f")      // Crisp dark text
		ColorTextBold = lipgloss.Color("#1f2328")  // High-contrast heading text
		ColorBtnBg = lipgloss.Color("#eaeef2")     // Light button background

		StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

		StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorTextBold).
			Background(ColorBtnBg).
			Padding(0, 1)

		StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(ColorHighlight).
			Padding(0, 1).
			MarginRight(1)

		StyleTabInactive = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(ColorBtnBg).
			Padding(0, 1).
			MarginRight(1)

		StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorCardBg).
			Padding(0, 1)

		StyleGreen = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
		StyleRed = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)
		StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
		StyleCyan = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
		StyleMuted = lipgloss.NewStyle().Foreground(ColorMuted)
		StyleBold = lipgloss.NewStyle().Bold(true).Foreground(ColorTextBold)

		StyleTableRowSelected = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(ColorHighlight)

		StyleTableRow = lipgloss.NewStyle().
			Foreground(ColorText)

		StyleStatusBar = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(ColorBtnBg).
			Padding(0, 1)

		StyleBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(ColorSecondary).
			Padding(0, 1)
	} else {
		// Dark theme
		ColorPrimary = lipgloss.Color("#58a6ff")
		ColorSecondary = lipgloss.Color("#bc8cff")
		ColorSuccess = lipgloss.Color("#3fb950")
		ColorDanger = lipgloss.Color("#f85149")
		ColorWarning = lipgloss.Color("#d29922")
		ColorMuted = lipgloss.Color("#8b949e")
		ColorBg = lipgloss.Color("#0d1117")
		ColorCardBg = lipgloss.Color("#161b22")
		ColorBorder = lipgloss.Color("#30363d")
		ColorHighlight = lipgloss.Color("#1f6feb")
		ColorText = lipgloss.Color("#c9d1d9")
		ColorTextBold = lipgloss.Color("#f0f6fc")
		ColorBtnBg = lipgloss.Color("#21262d")

		StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

		StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorTextBold).
			Background(ColorBtnBg).
			Padding(0, 1)

		StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(ColorHighlight).
			Padding(0, 1).
			MarginRight(1)

		StyleTabInactive = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(ColorCardBg).
			Padding(0, 1).
			MarginRight(1)

		StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorCardBg).
			Padding(0, 1)

		StyleGreen = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
		StyleRed = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)
		StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
		StyleCyan = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
		StyleMuted = lipgloss.NewStyle().Foreground(ColorMuted)
		StyleBold = lipgloss.NewStyle().Bold(true).Foreground(ColorTextBold)

		StyleTableRowSelected = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(ColorHighlight)

		StyleTableRow = lipgloss.NewStyle().
			Foreground(ColorText)

		StyleStatusBar = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(lipgloss.Color("#090d13")).
			Padding(0, 1)

		StyleBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(ColorSecondary).
			Padding(0, 1)
	}
}

