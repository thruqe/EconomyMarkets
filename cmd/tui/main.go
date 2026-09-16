package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"economy/tui"
)

func main() {
	numCompanies := flag.Int("companies", 500, "Number of companies in the simulation universe")
	seed := flag.Int64("seed", 42, "Random seed for reproducible universe generation")
	tickInterval := flag.Duration("tick-interval", 250*time.Millisecond, "Wall-clock duration per simulation tick")
	flag.Parse()

	// Channel for streaming simulation tick messages to the TUI program
	tickChan := make(chan tui.TickMsg, 100)

	// Initialize the simulation engine with 500 companies
	engine := tui.NewEngine(*numCompanies, *seed, *tickInterval, tickChan)
	engine.Start()
	defer engine.Stop()

	// Initialize the interactive Bubble Tea model
	model := tui.NewModel(engine, tickChan)

	// Run full-screen interactive TUI with alternate screen buffer
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running EconomyMarkets TUI: %v\n", err)
		os.Exit(1)
	}
}
