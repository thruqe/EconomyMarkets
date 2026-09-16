package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"economy/market"
)

// SavedPosition represents a serialized held position.
type SavedPosition struct {
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"` // "LONG" or "SHORT"
	Quantity  float64 `json:"quantity"`
	EntryCost float64 `json:"entry_cost"`
	OpenedAt  int64   `json:"opened_at"`
}

// SavedPortfolio holds persisted user portfolio state.
type SavedPortfolio struct {
	Cash         float64                  `json:"cash"`
	Positions    map[string]SavedPosition `json:"positions"`
	TradeHistory []string                 `json:"trade_history"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

var portfolioMutex sync.Mutex

const DefaultPortfolioPath = "data/portfolio.json"

// LoadPortfolio loads saved portfolio from disk if it exists.
// Returns nil if file doesn't exist.
func LoadPortfolio(filePath string) (*SavedPortfolio, error) {
	portfolioMutex.Lock()
	defer portfolioMutex.Unlock()

	if filePath == "" {
		filePath = DefaultPortfolioPath
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var p SavedPortfolio
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// SavePortfolio saves the portfolio state to disk.
func SavePortfolio(filePath string, p *SavedPortfolio) error {
	portfolioMutex.Lock()
	defer portfolioMutex.Unlock()

	if filePath == "" {
		filePath = DefaultPortfolioPath
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	p.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// ResetPortfolioDisk resets the saved portfolio to starting baseline.
func ResetPortfolioDisk(filePath string) error {
	fresh := &SavedPortfolio{
		Cash:         100_000.0,
		Positions:    make(map[string]SavedPosition),
		TradeHistory: []string{},
		UpdatedAt:    time.Now(),
	}
	return SavePortfolio(filePath, fresh)
}

// ApplySavedPortfolioToAccount restores positions and cash into a live Account.
func ApplySavedPortfolioToAccount(saved *SavedPortfolio, acct *market.Account) {
	if saved == nil || acct == nil {
		return
	}
	acct.Cash = saved.Cash
	if acct.Positions == nil {
		acct.Positions = make(map[string]*market.Position)
	} else {
		clear(acct.Positions)
	}

	for sym, sp := range saved.Positions {
		side := market.Long
		if sp.Side == "SHORT" {
			side = market.Short
		}
		opened := sp.OpenedAt
		if opened == 0 {
			opened = time.Now().UnixNano()
		}
		acct.Positions[sym] = &market.Position{
			Side:      side,
			Quantity:  sp.Quantity,
			EntryCost: sp.EntryCost,
			OpenedAt:  opened,
		}
	}
}

// ExtractSavedPortfolioFromAccount dumps a live Account to SavedPortfolio.
func ExtractSavedPortfolioFromAccount(acct *market.Account, tradeHistory []string) *SavedPortfolio {
	if acct == nil {
		return &SavedPortfolio{
			Cash:         100_000.0,
			Positions:    make(map[string]SavedPosition),
			TradeHistory: tradeHistory,
			UpdatedAt:    time.Now(),
		}
	}

	positions := make(map[string]SavedPosition)
	for sym, pos := range acct.Positions {
		sideStr := "LONG"
		if pos.Side == market.Short {
			sideStr = "SHORT"
		}
		positions[sym] = SavedPosition{
			Symbol:    sym,
			Side:      sideStr,
			Quantity:  pos.Quantity,
			EntryCost: pos.EntryCost,
			OpenedAt:  pos.OpenedAt,
		}
	}

	return &SavedPortfolio{
		Cash:         acct.Cash,
		Positions:    positions,
		TradeHistory: tradeHistory,
		UpdatedAt:    time.Now(),
	}
}
