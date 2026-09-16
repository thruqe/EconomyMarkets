package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var AvailableTimeframes = []string{"1m", "5m", "15m", "30m", "1h", "2h", "4h", "1D"}

// Standard Unicode Braille 2x4 dot bitmasks (U+2800)
var brailleDots = [2][4]rune{
	{0x01, 0x02, 0x04, 0x40}, // Column 0: dots 1, 2, 3, 7
	{0x08, 0x10, 0x20, 0x80}, // Column 1: dots 4, 5, 6, 8
}

type ChartView struct {
	width          int
	height         int
	selectedTF     string
	selectedTFIdx  int
	isAreaMode     bool
	inspectCol     int // -1 when not inspecting
	isSearching    bool
	searchQuery    string
	tfBounds       map[string][2]int
	prevBtnBound   [2]int
	nextBtnBound   [2]int
	searchBtnBound [2]int
	styleBtnBound  [2]int
	deepDiveBound  [2]int
	headerRowY     int
	controlsRowY   int
}

func NewChartView() *ChartView {
	return &ChartView{
		width:         120,
		height:        35,
		selectedTF:    "1m",
		selectedTFIdx: 0,
		inspectCol:    -1,
		tfBounds:      make(map[string][2]int),
		headerRowY:    3,
		controlsRowY:  5,
	}
}

func (v *ChartView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *ChartView) SetTimeframe(tf string) {
	for i, avail := range AvailableTimeframes {
		if avail == tf {
			v.SetTimeframeIndex(i)
			return
		}
	}
}

func (v *ChartView) Timeframe() string {
	if v.selectedTF == "" {
		return "1m"
	}
	return v.selectedTF
}

func (v *ChartView) NextTimeframe() {
	if v.selectedTFIdx < len(AvailableTimeframes)-1 {
		v.selectedTFIdx++
		v.selectedTF = AvailableTimeframes[v.selectedTFIdx]
		v.inspectCol = -1
	}
}

func (v *ChartView) PrevTimeframe() {
	if v.selectedTFIdx > 0 {
		v.selectedTFIdx--
		v.selectedTF = AvailableTimeframes[v.selectedTFIdx]
		v.inspectCol = -1
	}
}

func (v *ChartView) SetTimeframeIndex(idx int) {
	if idx >= 0 && idx < len(AvailableTimeframes) {
		v.selectedTFIdx = idx
		v.selectedTF = AvailableTimeframes[idx]
		v.inspectCol = -1
	}
}

func (v *ChartView) ToggleAreaMode() {
	v.isAreaMode = !v.isAreaMode
}

func (v *ChartView) ToggleSearch() {
	v.isSearching = !v.isSearching
	if !v.isSearching {
		v.searchQuery = ""
	}
}

func (v *ChartView) IsSearching() bool {
	return v.isSearching
}

func (v *ChartView) SearchQuery() string {
	return v.searchQuery
}

func (v *ChartView) SetSearch(q string) {
	v.searchQuery = q
}

func (v *ChartView) AppendSearchRune(r rune) {
	v.searchQuery += string(r)
}

func (v *ChartView) PopSearchRune() {
	if len(v.searchQuery) > 0 {
		v.searchQuery = v.searchQuery[:len(v.searchQuery)-1]
	}
}

func (v *ChartView) ClearSearch() {
	v.isSearching = false
	v.searchQuery = ""
}

func (v *ChartView) MoveInspectLeft() {
	if v.inspectCol <= 0 {
		v.inspectCol = 0
	} else {
		v.inspectCol--
	}
}

func (v *ChartView) MoveInspectRight(maxCol int) {
	if v.inspectCol < 0 {
		v.inspectCol = maxCol / 2
	} else if v.inspectCol < maxCol-1 {
		v.inspectCol++
	}
}

func (v *ChartView) ClearInspect() {
	v.inspectCol = -1
}

func (v *ChartView) Render(company *CompanyRow, candles []CandleView) string {
	if company == nil {
		return lipgloss.NewStyle().Foreground(ColorMuted).Padding(2, 4).Render(
			"No company selected. Please select a company from [2. 500 Stocks].",
		)
	}

	// 1. Header Row 1: Company details + Quick Stock Switcher
	chgStr := fmt.Sprintf("%+6.2f%%", company.ChangePct)
	chgStyled := StyleGreen.Render(chgStr)
	if company.ChangePct < 0 {
		chgStyled = StyleRed.Render(chgStr)
	}

	var companyTitle string
	if v.width < 95 {
		companyTitle = fmt.Sprintf(" %s $%.2f %s Spr: $%.2f",
			StyleTitle.Render(company.Symbol),
			company.Price,
			chgStyled,
			company.Spread,
		)
	} else {
		companyTitle = fmt.Sprintf(" %s - %s   |   Sector: %s   |   Price: $%.2f %s   |   Spread: $%.2f",
			StyleTitle.Render(company.Symbol),
			truncate(company.Name, 20),
			company.Sector,
			company.Price,
			chgStyled,
			company.Spread,
		)
	}

	var prevBtn, nextBtn, searchBtn string
	if v.width < 85 {
		prevBtn = lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorPrimary).Padding(0, 1).Render("[<]")
		nextBtn = lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorPrimary).Padding(0, 1).Render("[>]")
		searchBtn = lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(lipgloss.Color("#f0f6fc")).Padding(0, 1).Render("[/]")
	} else {
		prevBtn = lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorPrimary).Padding(0, 1).Render("[ < PREV ]")
		nextBtn = lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorPrimary).Padding(0, 1).Render("[ NEXT > ]")
		searchBtn = lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(lipgloss.Color("#f0f6fc")).Padding(0, 1).Render("[ / SEARCH ]")
	}

	stockNav := lipgloss.JoinHorizontal(lipgloss.Center, prevBtn, " ", nextBtn, " ", searchBtn)
	headerRow1 := lipgloss.JoinHorizontal(lipgloss.Center, companyTitle, "  ", stockNav)

	// Dynamically calculate stock navigation button bounds
	v.headerRowY = 3
	navStartX := 2 + lipgloss.Width(companyTitle) + 2
	v.prevBtnBound = [2]int{navStartX, navStartX + lipgloss.Width(prevBtn)}
	navStartX += lipgloss.Width(prevBtn) + 1
	v.nextBtnBound = [2]int{navStartX, navStartX + lipgloss.Width(nextBtn)}
	navStartX += lipgloss.Width(nextBtn) + 1
	v.searchBtnBound = [2]int{navStartX, navStartX + lipgloss.Width(searchBtn)}

	// Search bar if search mode active
	searchRow := ""
	if v.isSearching {
		v.controlsRowY = 6
		searchPrompt := lipgloss.NewStyle().Background(ColorHighlight).Foreground(lipgloss.Color("#ffffff")).Bold(true).Padding(0, 1).Render(
			" QUICK SEARCH TICKER: " + strings.ToUpper(v.searchQuery) + "_ (Type symbol, press Enter) ",
		)
		searchRow = "\n" + searchPrompt
	} else {
		v.controlsRowY = 5
	}

	// 2. Header Row 2: Timeframe selector buttons + Style Toggle + Deep Dive
	if v.tfBounds == nil {
		v.tfBounds = make(map[string][2]int)
	}
	curTfX := 2
	var tfButtons []string
	for _, tf := range AvailableTimeframes {
		var btnRendered string
		if tf == v.selectedTF {
			btnRendered = StyleTabActive.Render(fmt.Sprintf("[%s]", tf))
		} else {
			btnRendered = StyleTabInactive.Render(fmt.Sprintf("[%s]", tf))
		}
		btnW := lipgloss.Width(btnRendered)
		v.tfBounds[tf] = [2]int{curTfX, curTfX + btnW}
		curTfX += btnW
		tfButtons = append(tfButtons, btnRendered)
	}
	tfBar := lipgloss.JoinHorizontal(lipgloss.Center, tfButtons...)

	styleLabel := "[ STYLE: LINE ]"
	if v.isAreaMode {
		styleLabel = "[ STYLE: AREA ]"
	}
	styleBtn := lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorSecondary).Bold(true).Padding(0, 1).Render(styleLabel)
	deepDiveBtn := lipgloss.NewStyle().Background(lipgloss.Color("#21262d")).Foreground(ColorPrimary).Bold(true).Padding(0, 1).Render("[ 4. DEEP DIVE & DOM ]")

	curTfX += 4 // "    "
	v.styleBtnBound = [2]int{curTfX, curTfX + lipgloss.Width(styleBtn)}
	curTfX += lipgloss.Width(styleBtn) + 1
	v.deepDiveBound = [2]int{curTfX, curTfX + lipgloss.Width(deepDiveBtn)}

	controlsBar := lipgloss.JoinHorizontal(lipgloss.Center, tfBar, "    ", styleBtn, " ", deepDiveBtn)

	banner := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Width(v.width - 4).
		Padding(0, 1).
		Render(fmt.Sprintf("%s%s\n\n%s", headerRow1, searchRow, controlsBar))

	// 3. Extract OHLCV and compute statistics
	var prices []float64
	var volumes []float64
	var times []time.Time

	if len(candles) == 0 {
		prices = []float64{company.Price}
		volumes = []float64{0}
		times = []time.Time{time.Now()}
	} else {
		for _, c := range candles {
			prices = append(prices, c.Close)
			volumes = append(volumes, c.Volume)
			times = append(times, c.Time)
		}
	}

	openPrice := prices[0]
	currentClose := prices[len(prices)-1]
	highPrice := currentClose
	lowPrice := currentClose
	var totalVol float64
	for i, p := range prices {
		if p > highPrice {
			highPrice = p
		}
		if p < lowPrice && p > 0 {
			lowPrice = p
		}
		totalVol += volumes[i]
	}

	periodChg := currentClose - openPrice
	periodPct := 0.0
	if openPrice > 0 {
		periodPct = (periodChg / openPrice) * 100.0
	}
	periodStyle := StyleGreen
	if periodChg < 0 {
		periodStyle = StyleRed
	}

	// 4. Metrics & Inspector Bar
	var metricsText string
	if v.width < 90 {
		metricsText = fmt.Sprintf("  O: $%.2f  H: $%.2f  L: $%.2f  C: $%.2f (%s)  V: %s",
			openPrice, highPrice, lowPrice, currentClose, periodStyle.Render(fmt.Sprintf("%+.2f%%", periodPct)), formatVolume(totalVol))
	} else {
		metricsText = fmt.Sprintf("  OPEN: $%.2f   HIGH: $%.2f   LOW: $%.2f   CLOSE: $%.2f (%s)   RANGE: $%.2f   VOL: %s",
			openPrice, highPrice, lowPrice, currentClose, periodStyle.Render(fmt.Sprintf("%+.2f%%", periodPct)), highPrice-lowPrice, formatVolume(totalVol))
	}

	chartHeight := v.height - 13
	if chartHeight < 8 {
		chartHeight = 8
	}
	chartWidth := v.width - 6

	chartBoxContent := renderSmoothTerminalChart(
		prices,
		volumes,
		times,
		chartWidth,
		chartHeight,
		company.Price,
		v.isAreaMode,
		v.inspectCol,
	)

	chartBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(v.width - 4).
		Render(fmt.Sprintf("%s\n\n%s", StyleBold.Render(metricsText), chartBoxContent))

	var footer string
	if v.width < 90 {
		footer = lipgloss.NewStyle().Foreground(ColorMuted).Render(
			" Controls: [t] Timeframe  •  [a] Area/Line  •  [j/k] Change Stock  •  [/] Search  •  [d] Detail",
		)
	} else {
		footer = lipgloss.NewStyle().Foreground(ColorMuted).Render(
			" Controls: Click any button/timeframe with mouse  •  [t or [/]] Timeframe  •  [a] Area/Line  •  [↑/↓ or j/k] Change Stock  •  [/] Search  •  [d] Deep Dive",
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		banner,
		chartBox,
		footer,
	)
}

// renderSmoothTerminalChart generates a high-resolution sub-pixel Braille curve with volume sub-panel and time axis
func renderSmoothTerminalChart(
	prices []float64,
	volumes []float64,
	times []time.Time,
	width, height int,
	currentPrice float64,
	isAreaMode bool,
	inspectCol int,
) string {
	n := len(prices)
	if n == 0 {
		prices = []float64{currentPrice}
		volumes = []float64{0}
		times = []time.Time{time.Now()}
		n = 1
	}

	// Determine min and max prices
	minP := math.MaxFloat64
	maxP := -math.MaxFloat64
	for _, p := range prices {
		if p < minP && p > 0 {
			minP = p
		}
		if p > maxP {
			maxP = p
		}
	}
	if maxP <= minP {
		maxP = minP + 1.0
	}
	padding := (maxP - minP) * 0.06
	if padding < 0.05 {
		padding = 0.05
	}
	maxP += padding
	minP -= padding
	priceRange := maxP - minP

	gridWidth := width - 14
	if gridWidth < 30 {
		gridWidth = 30
	}
	plotHeight := height - 6
	if plotHeight < 8 {
		plotHeight = 8
	}

	// Sub-pixel dimensions (Braille: 2 horizontal dots x 4 vertical dots per character cell)
	subWidth := gridWidth * 2
	subHeight := plotHeight * 4

	// 1. Resample prices across all subWidth dot columns with linear interpolation
	subPrices := make([]float64, subWidth)
	for x := 0; x < subWidth; x++ {
		if n == 1 {
			subPrices[x] = prices[0]
			continue
		}
		frac := float64(x) / float64(subWidth-1) * float64(n-1)
		i0 := int(frac)
		if i0 >= n-1 {
			i0 = n - 2
		}
		t := frac - float64(i0)
		subPrices[x] = prices[i0]*(1.0-t) + prices[i0+1]*t
	}

	// Determine overall chart color
	isOverallGreen := subPrices[subWidth-1] >= subPrices[0]
	lineColor := StyleGreen
	if !isOverallGreen {
		lineColor = StyleRed
	}

	// 2. Map prices to Y sub-pixel coordinates (0 = top = maxP, subHeight-1 = bottom = minP)
	subYCoords := make([]int, subWidth)
	for x, p := range subPrices {
		y := int(math.Round((maxP - p) / priceRange * float64(subHeight-1)))
		if y < 0 {
			y = 0
		}
		if y >= subHeight {
			y = subHeight - 1
		}
		subYCoords[x] = y
	}

	// 3. Braille grid (character cells)
	cellGrid := make([][]rune, plotHeight)
	for r := 0; r < plotHeight; r++ {
		cellGrid[r] = make([]rune, gridWidth)
	}

	// Helper to illuminate a sub-pixel dot
	setDot := func(px, py int) {
		if px >= 0 && px < subWidth && py >= 0 && py < subHeight {
			cx := px / 2
			cy := py / 4
			dotCol := px % 2
			dotRow := py % 4
			cellGrid[cy][cx] |= brailleDots[dotCol][dotRow]
		}
	}

	// 4. Draw smooth continuous line using DDA / Bresenham between sub-pixels
	for x := 0; x < subWidth; x++ {
		if x == 0 {
			setDot(0, subYCoords[0])
		} else {
			x0, y0 := x-1, subYCoords[x-1]
			x1, y1 := x, subYCoords[x]
			dx := float64(x1 - x0)
			dy := float64(y1 - y0)
			steps := int(math.Max(math.Abs(dx), math.Abs(dy)))
			if steps < 1 {
				steps = 1
			}
			for s := 0; s <= steps; s++ {
				t := float64(s) / float64(steps)
				px := int(math.Round(float64(x0) + dx*t))
				py := int(math.Round(float64(y0) + dy*t))
				setDot(px, py)
			}
		}

		// Area mode: fill all dots below the line to the bottom
		if isAreaMode {
			topY := subYCoords[x]
			for fillY := topY + 1; fillY < subHeight; fillY++ {
				setDot(x, fillY)
			}
		}
	}

	// Calculate row index for current price guideline
	currentPriceRow := int(math.Round((maxP - currentPrice) / priceRange * float64(plotHeight-1)))
	if currentPriceRow < 0 {
		currentPriceRow = 0
	}
	if currentPriceRow >= plotHeight {
		currentPriceRow = plotHeight - 1
	}

	// 5. Render character lines with Y-axis price labels
	var lines []string

	for r := 0; r < plotHeight; r++ {
		rowPrice := maxP - (float64(r)/float64(plotHeight-1))*priceRange
		axis := fmt.Sprintf("%9.2f ┤ ", rowPrice)

		var rowChars []string
		for c := 0; c < gridWidth; c++ {
			b := cellGrid[r][c]
			if b > 0 {
				rowChars = append(rowChars, lineColor.Render(string(rune(0x2800+b))))
			} else if inspectCol >= 0 && c == inspectCol {
				// Vertical crosshair guide
				rowChars = append(rowChars, StyleMuted.Render("┆"))
			} else if r == currentPriceRow {
				// Subtle horizontal guideline at current price
				rowChars = append(rowChars, lipgloss.NewStyle().Foreground(lipgloss.Color("#2d333b")).Render("┄"))
			} else {
				rowChars = append(rowChars, " ")
			}
		}

		// Show current price pointer on the current price row
		rowSuffix := ""
		if r == currentPriceRow {
			rowSuffix = lineColor.Render(fmt.Sprintf(" ◄ $%.2f", currentPrice))
		}

		lines = append(lines, StyleMuted.Render(axis)+strings.Join(rowChars, "")+rowSuffix)
	}

	// 6. Bottom axis line
	bottomAxis := fmt.Sprintf("%10s┴%s", "", strings.Repeat("─", gridWidth))
	lines = append(lines, StyleMuted.Render(bottomAxis))

	// 7. Time axis labels
	startTimeStr := times[0].Format("15:04")
	endTimeStr := times[len(times)-1].Format("15:04")
	midTimeStr := times[len(times)/2].Format("15:04")

	spacing := (gridWidth - len(startTimeStr) - len(endTimeStr) - len(midTimeStr)) / 2
	if spacing < 2 {
		spacing = 2
	}
	timeRow := fmt.Sprintf("%11s%s%s%s%s%s",
		"",
		startTimeStr,
		strings.Repeat(" ", spacing),
		midTimeStr,
		strings.Repeat(" ", spacing),
		endTimeStr,
	)
	lines = append(lines, StyleMuted.Render(timeRow))

	// 8. Volume sub-panel (Histogram at the bottom)
	maxVol := 1.0
	for _, v := range volumes {
		if v > maxVol {
			maxVol = v
		}
	}

	volBlocks := []rune{' ', ' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var volChars []string
	for c := 0; c < gridWidth; c++ {
		origIdx := int(float64(c) / float64(gridWidth-1) * float64(n-1))
		if origIdx >= n {
			origIdx = n - 1
		}
		volVal := volumes[origIdx]
		volLvl := int((volVal / maxVol) * 7)
		if volLvl < 1 && volVal > 0 {
			volLvl = 1
		}
		if volLvl > 8 {
			volLvl = 8
		}
		volChar := string(volBlocks[volLvl])
		volChars = append(volChars, lipgloss.NewStyle().Foreground(lipgloss.Color("#58a6ff")).Render(volChar))
	}
	lines = append(lines, StyleMuted.Render("   VOLUME ┤ ")+strings.Join(volChars, ""))

	return strings.Join(lines, "\n")
}

// HandleMouseClick checks if the user clicked on timeframe buttons, stock navigation, style toggle, or chart area
func (v *ChartView) HandleMouseClick(x, y int) string {
	// 1. Stock navigation header row
	if y >= v.headerRowY-1 && y <= v.headerRowY+1 {
		if x >= v.prevBtnBound[0] && x < v.prevBtnBound[1] {
			return "prev_stock"
		}
		if x >= v.nextBtnBound[0] && x < v.nextBtnBound[1] {
			return "next_stock"
		}
		if x >= v.searchBtnBound[0] && x < v.searchBtnBound[1] {
			v.ToggleSearch()
			return "toggle_search"
		}
	}

	// 2. Controls row (Timeframes & Style toggle & Deep Dive)
	if y >= v.controlsRowY-1 && y <= v.controlsRowY+1 {
		for _, tf := range AvailableTimeframes {
			if b, ok := v.tfBounds[tf]; ok {
				if x >= b[0] && x < b[1] {
					v.SetTimeframe(tf)
					return "set_tf:" + tf
				}
			}
		}
		// Style toggle (Line / Area)
		if x >= v.styleBtnBound[0] && x < v.styleBtnBound[1] {
			v.ToggleAreaMode()
			return "toggle_area"
		}
		// Deep Dive button
		if x >= v.deepDiveBound[0] && x < v.deepDiveBound[1] {
			return "open_detail"
		}
	}

	// 3. Chart area canvas for inspection
	if y >= v.controlsRowY+2 && y <= v.height-4 {
		plotCol := x - 13
		if plotCol >= 0 && plotCol < v.width-18 {
			v.inspectCol = plotCol
			return "inspect"
		}
	}

	return ""
}

