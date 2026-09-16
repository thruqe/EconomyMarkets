package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"economy/agent"
	"economy/company"
	"economy/market"
	"economy/retail"
	"economy/sim"
	"economy/storage"
	"economy/technicals"
)

type serverState struct {
	mu              sync.RWMutex
	startTime       time.Time
	currentTick     int
	totalTrades     int64
	totalVolume     float64
	latestPrices    map[string]sim.PriceQuote
	companies       []*company.Company
	sim             *sim.Simulation
	humanTrader     *retail.HumanTrader
	humanAcct       *market.Account
	aggregators     map[string]*technicals.Aggregator
	lastBarCount    map[string]int
	candlesWindow   int
	usedSymbols     map[string]bool
	rng             *rand.Rand
	tfManager       *MultiTimeframeManager
	speedMultiplier float64
	tickInterval    time.Duration
	simTime         int64
	lastTradePrice  map[string]float64
}

func main() {
	dbPath := flag.String("db", "data/market.db", "Path to SQLite database file")
	numCompanies := flag.Int("companies", 500, "Number of companies in the simulation universe")
	tickInterval := flag.Duration("tick-interval", 250*time.Millisecond, "Wall-clock duration per simulation tick")
	httpPort := flag.Int("port", 8080, "HTTP server port for dashboard and REST API")
	seed := flag.Int64("seed", 42, "Random seed for reproducible universe generation")
	candlesWindow := flag.Int("candle-window", 10, "Number of ticks per OHLC candlestick bar")
	flag.Parse()

	log.Printf("Starting Economy Markets Simulation Daemon...")
	log.Printf("Configuration: companies=%d, tickInterval=%v, candleWindow=%d, port=%d, db=%s",
		*numCompanies, *tickInterval, *candlesWindow, *httpPort, *dbPath)

	// 1. Initialize SQLite storage layer
	db, err := storage.Open(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	writer := storage.NewAsyncWriter(db, 2000, 500*time.Millisecond)
	defer writer.Close()

	// 2. Generate universe and configure simulation
	params := company.DefaultGenerationParams()
	params.BaseJumpParams.LambdaDown = 1.0 / 20000 // Rare corporate shock (~once every 80 wall-clock minutes per company)
	params.BaseJumpParams.LambdaUp = 1.0 / 35000
	params.BaseJumpParams.DownJumpMin = 0.03
	params.BaseJumpParams.DownJumpMax = 0.08
	params.BaseJumpParams.UpJumpMin = 0.03
	params.BaseJumpParams.UpJumpMax = 0.08
	universe := company.GenerateUniverse(*numCompanies, *seed, params)

	s := sim.NewSimulation(5, 200)
	for _, co := range universe {
		s.AddCompany(co)
	}

	for _, co := range universe {
		mm := agent.NewMarketMaker("mm_"+co.Symbol, 50_000_000, 10, 0.10)
		mm.MaxInventory = 500_000
		mm.QuoteSize = 250
		mm.LadderLevels = 5
		mm.SetCompany(co)
		s.AddParticipant(mm, co.Symbol)

		hfAcct := market.NewAccount("hf_"+co.Symbol, 2_000_000, 5, 0.10)
		hf := agent.NewHedgeFund("hf_"+co.Symbol, hfAcct, co)
		hf.MinTradeThreshold = 0.003 // 30 bps: active value discovery as fundamentals drift
		hf.FullConvictionThreshold = 0.025
		hf.RebalanceThreshold = 0.003
		hf.ExecutionRate = 0.06
		hf.MaxOrderShares = 200
		s.AddParticipant(hf, co.Symbol)

		bankAcct := market.NewAccount("bank_"+co.Symbol, 3_000_000, 3, 0.15)
		bank := agent.NewBank("bank_"+co.Symbol, bankAcct, []*company.Company{co})
		bank.MinTradeThreshold = 0.006 // 60 bps
		bank.FullConvictionThreshold = 0.040
		bank.RebalanceThreshold = 0.006
		bank.ExecutionRate = 0.05
		bank.MaxOrderShares = 250
		s.AddParticipant(bank, co.Symbol)
	}

	pool := retail.GenerateWatchlistPool(100, *seed, universe, 1, 3, 1000, 20000)
	for _, bot := range pool {
		var watched []string
		for _, symbol := range s.Symbols() {
			if bot.Watches(symbol) {
				watched = append(watched, symbol)
			}
		}
		if len(watched) > 0 {
			s.AddParticipant(bot, watched...)
		}
	}

	// 3. User session / HumanTrader participating in the live market
	humanAcct := market.NewAccount("human_trader", 100_000, 5, 0.10)
	humanTrader := retail.NewHumanTrader("human_trader", humanAcct)
	s.AddParticipant(humanTrader, s.Symbols()...)

	// 4. Technical aggregators for candlesticks
	aggregators := make(map[string]*technicals.Aggregator, len(universe))
	lastBarCount := make(map[string]int, len(universe))
	for _, co := range universe {
		aggregators[co.Symbol] = technicals.NewAggregator(*candlesWindow, 500)
	}

	usedSymbols := make(map[string]bool)
	for _, co := range universe {
		usedSymbols[co.Symbol] = true
	}
	serverRng := rand.New(rand.NewSource(*seed + 777))

	// 4. Multi-Timeframe Manager with 10-Year Historical Backfill
	// Align simStartTime to the minute boundary so candle intervals start and end cleanly on clock minutes
	simStartTime := (time.Now().UnixMilli() / 60000) * 60000
	tfManager := NewMultiTimeframeManager(db)
	log.Printf("Initializing live multi-timeframe candles for %d companies...", len(universe))
	for _, co := range universe {
		stb := tfManager.RegisterSymbol(co.Symbol)
		settledMid := co.ReportedValue
		if settledMid <= 0 {
			settledMid = co.TrueValue
		}
		if bk := s.Book(co.Symbol); bk != nil {
			if m, ok := bk.MidPrice(); ok && m > 0 {
				settledMid = m
			} else if bestBid, ok := bk.BestBid(); ok && bestBid > 0 {
				settledMid = bestBid
			} else if bestAsk, ok := bk.BestAsk(); ok && bestAsk > 0 {
				settledMid = bestAsk
			}
		}
		stb.InitLiveCandles(co.Symbol, settledMid, simStartTime)
	}
	log.Printf("✓ Live multi-timeframe candles initialized for %d companies.", len(universe))

	state := &serverState{
		startTime:       time.Now(),
		simTime:         simStartTime,
		latestPrices:    make(map[string]sim.PriceQuote),
		companies:       universe,
		sim:             s,
		humanTrader:     humanTrader,
		humanAcct:       humanAcct,
		aggregators:     aggregators,
		lastBarCount:    lastBarCount,
		candlesWindow:   *candlesWindow,
		usedSymbols:     usedSymbols,
		rng:             serverRng,
		tfManager:       tfManager,
		speedMultiplier: 1.0,
		tickInterval:    *tickInterval,
		lastTradePrice:  make(map[string]float64),
	}
	for _, co := range universe {
		state.lastTradePrice[co.Symbol] = co.TrueValue
	}

	// 5. Setup HTTP Server with REST API
	mux := http.NewServeMux()
	setupRoutes(mux, db, state)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *httpPort),
		Handler: enableCORS(mux),
	}

	go func() {
		log.Printf("HTTP REST API listening on http://localhost:%d", *httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// 5. Signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("Simulation running continuously. Press Ctrl+C to exit gracefully.")

	ticker := time.NewTicker(*tickInterval)
	defer ticker.Stop()

	// 6. Main continuous simulation loop
	for {
		select {
		case <-ctx.Done():
			log.Println("\nShutdown signal received. Stopping simulation loop...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_ = httpServer.Shutdown(shutdownCtx)

			log.Println("Flushing pending records to SQLite database...")
			if err := writer.Flush(); err != nil {
				log.Printf("Error during final flush: %v", err)
			}
			log.Println("Database flushed. Exiting cleanly.")
			return

		case now := <-ticker.C:
			report := s.Step()

			// Update in-memory telemetry
			state.mu.Lock()
			state.currentTick = report.Tick
			state.latestPrices = report.Prices
			state.totalTrades += int64(len(report.Trades))
			for _, tr := range report.Trades {
				state.totalVolume += tr.Quantity * tr.Price
			}
			state.mu.Unlock()

			// Prepare storage batch
			nowMillis := now.UnixMilli()
			batch := &storage.Batch{}

			// 1. Group trade volumes per symbol, specifically identifying retail volume
			symbolTradeVol := make(map[string]float64)
			symbolRetailVol := make(map[string]float64)
			symbolTrades := make(map[string][]storage.TradeRecord)

			for _, tr := range report.Trades {
				tradeRec := storage.TradeRecord{
					Tick:      tr.Tick,
					Timestamp: nowMillis,
					Symbol:    tr.Symbol,
					TakerID:   tr.TakerAgentID,
					MakerID:   tr.MakerAgentID,
					Side:      tr.Side,
					Price:     tr.Price,
					Quantity:  tr.Quantity,
				}
				batch.Trades = append(batch.Trades, tradeRec)
				symbolTrades[tr.Symbol] = append(symbolTrades[tr.Symbol], tradeRec)

				symbolTradeVol[tr.Symbol] += tr.Quantity
				isRetail := strings.HasPrefix(tr.TakerAgentID, "retail_") ||
					strings.HasPrefix(tr.MakerAgentID, "retail_") ||
					tr.TakerAgentID == "human_trader" ||
					tr.MakerAgentID == "human_trader"
				if isRetail {
					symbolRetailVol[tr.Symbol] += tr.Quantity
				}
			}

			// Advance simulated market time realistically based on tick interval and speed multiplier
			// At 1x speed with 250ms tickInterval, 1 tick = 250ms of sim time.
			// Exactly 4 ticks = 1 second, 240 ticks = 60 seconds (1 full 1m candle).
			state.mu.Lock()
			simTimeStep := int64(float64(state.tickInterval.Milliseconds()) * state.speedMultiplier)
			if simTimeStep <= 0 {
				simTimeStep = state.tickInterval.Milliseconds()
			}
			state.simTime += simTimeStep
			currentSimTime := state.simTime
			state.mu.Unlock()

			// 2. Feed prices and genuine trade volume into MultiTimeframeManager
			for sym, p := range report.Prices {
				if p.HasMid && p.Mid > 0 {
					batch.Prices = append(batch.Prices, storage.PriceRecord{
						Tick:      report.Tick,
						Timestamp: currentSimTime,
						Symbol:    sym,
						Mid:       p.Mid,
						Bid:       p.Bid,
						Ask:       p.Ask,
						Spread:    p.Spread,
					})

					if stb := state.tfManager.Get(sym); stb != nil {
						tradeVol := symbolTradeVol[sym]
						retailVol := symbolRetailVol[sym]

						// Authentic candlestick pricing: use continuous cleared mid-market price
						// to represent genuine price discovery, eliminating artificial bid-ask bounce
						// wicks across all timeframes.
						feedPrice := p.Mid
						if feedPrice <= 0 {
							feedPrice = state.lastTradePrice[sym]
						}
						if feedPrice > 0 {
							state.lastTradePrice[sym] = feedPrice
						}

						closedBars := stb.AddTick(currentSimTime, feedPrice, tradeVol, retailVol)
						batch.Candles = append(batch.Candles, closedBars...)
					}
				}
			}

			for _, ev := range report.Events {
				detailsBytes, _ := json.Marshal(map[string]any{
					"fundamental_kind": ev.FundamentalKind.String(),
					"multiplier":       ev.Multiplier,
					"liquidated_agent": ev.LiquidatedAgentID,
					"liquidation_qty":  ev.LiquidationQty,
					"ipo_revenue":      ev.IPORevenue,
					"ipo_price":        ev.IPOPrice,
					"ipo_shares":       ev.IPOShares,
				})
				batch.Events = append(batch.Events, storage.EventRecord{
					Tick:      ev.Tick,
					Timestamp: nowMillis,
					Kind:      ev.Kind.String(),
					Symbol:    ev.Symbol,
					Details:   string(detailsBytes),
				})
			}

			// Sample account snapshots every 20 ticks to keep snapshot size bounded
			if report.Tick%20 == 0 {
				for _, acct := range report.Accounts {
					batch.Accounts = append(batch.Accounts, storage.AccountRecord{
						Tick:        report.Tick,
						Timestamp:   nowMillis,
						AgentID:     acct.AgentID,
						Cash:        acct.Cash,
						Equity:      acct.Equity,
						MarginRatio: acct.MarginRatio,
					})
				}
			}

			writer.Enqueue(batch)

			// Clean event log periodically to avoid memory growth in forever runs
			if report.Tick%1000 == 0 {
				s.ClearEventLog()
				log.Printf("Simulation Tick %d: totalTrades=%d, totalVolume=$%.2f",
					report.Tick, state.totalTrades, state.totalVolume)
			}
		}
	}
}

type PriceItem struct {
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	Sector            string  `json:"sector"`
	CapTier           string  `json:"cap_tier"`
	TrueValue         float64 `json:"true_value"`
	ReportedValue     float64 `json:"reported_value"`
	Mid               float64 `json:"mid"`
	Bid               float64 `json:"bid"`
	Ask               float64 `json:"ask"`
	Spread            float64 `json:"spread"`
	AnnualRevenue     float64 `json:"annual_revenue"`
	NetMargin         float64 `json:"net_margin"`
	SectorMultiple    float64 `json:"sector_multiple"`
	MarketCap         float64 `json:"market_cap"`
	ReportedMarketCap float64 `json:"reported_market_cap"`
	SharesOutstanding float64 `json:"shares_outstanding"`
	FloatShares       float64 `json:"float_shares"`
	PERatio           float64 `json:"pe_ratio"`
	PSRatio           float64 `json:"ps_ratio"`
	IsIPO             bool    `json:"is_ipo"`
	IPOTick           int     `json:"ipo_tick"`
	IPOPrice          float64 `json:"ipo_price"`
}

type PositionDTO struct {
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Quantity      float64 `json:"quantity"`
	EntryPrice    float64 `json:"entry_price"`
	CurrentPrice  float64 `json:"current_price"`
	Value         float64 `json:"value"`
	UnrealizedPnL float64 `json:"unrealized_pnl"`
	PnLPercent    float64 `json:"pnl_percent"`
}

func (state *serverState) buildPriceItems() []PriceItem {
	items := make([]PriceItem, 0, len(state.companies))
	for _, co := range state.companies {
		q := state.latestPrices[co.Symbol]
		marketPrice := q.Mid
		if marketPrice <= 0 {
			marketPrice = co.TrueValue
		}
		items = append(items, PriceItem{
			Symbol:            co.Symbol,
			Name:              co.Name,
			Sector:            co.Sector.String(),
			CapTier:           co.CapTier.String(),
			TrueValue:         co.TrueValue,
			ReportedValue:     co.ReportedValue,
			Mid:               q.Mid,
			Bid:               q.Bid,
			Ask:               q.Ask,
			Spread:            q.Spread,
			AnnualRevenue:     co.AnnualRevenue,
			NetMargin:         co.NetMargin,
			SectorMultiple:    co.SectorMultiple,
			MarketCap:         co.MarketCap(),
			ReportedMarketCap: co.ReportedMarketCap(),
			SharesOutstanding: co.SharesOutstanding,
			FloatShares:       co.Float,
			PERatio:           co.PriceToEarnings(marketPrice),
			PSRatio:           co.PriceToSales(marketPrice),
			IsIPO:             co.IsIPO,
			IPOTick:           co.IPOTick,
			IPOPrice:          co.IPOPrice,
		})
	}
	return items
}

func (state *serverState) buildPortfolioDTO() map[string]any {
	currentPrices := make(map[string]float64)
	for sym, q := range state.latestPrices {
		if q.HasMid {
			currentPrices[sym] = q.Mid
		}
	}
	for sym := range state.humanAcct.Positions {
		if _, ok := currentPrices[sym]; !ok {
			if bk := state.sim.Book(sym); bk != nil {
				if mid, hasMid := bk.MidPrice(); hasMid {
					currentPrices[sym] = mid
				}
			}
		}
	}

	equity := state.humanAcct.Equity(currentPrices)
	buyingPower := state.humanAcct.AvailableBuyingPower(currentPrices)
	marginRatio, _ := state.humanAcct.MarginRatio(currentPrices)

	positions := make([]PositionDTO, 0)
	for sym, pos := range state.humanAcct.Positions {
		curPrice := currentPrices[sym]
		entryPrice := 0.0
		if pos.Quantity > 0 {
			entryPrice = pos.EntryCost / pos.Quantity
		}
		if curPrice == 0 {
			curPrice = entryPrice
		}
		sideStr := "LONG"
		var pnl float64
		if pos.Side == market.Short {
			sideStr = "SHORT"
			pnl = pos.Quantity * (entryPrice - curPrice)
		} else {
			pnl = pos.Quantity * (curPrice - entryPrice)
		}
		pnlPct := 0.0
		if pos.EntryCost > 0 {
			pnlPct = (pnl / pos.EntryCost) * 100.0
		}
		positions = append(positions, PositionDTO{
			Symbol:        sym,
			Side:          sideStr,
			Quantity:      pos.Quantity,
			EntryPrice:    entryPrice,
			CurrentPrice:  curPrice,
			Value:         pos.Quantity * curPrice,
			UnrealizedPnL: pnl,
			PnLPercent:    pnlPct,
		})
	}

	return map[string]any{
		"cash":                     state.humanAcct.Cash,
		"equity":                   equity,
		"buying_power":             buyingPower,
		"margin_ratio":             marginRatio,
		"maintenance_margin_ratio": state.humanAcct.MaintenanceMarginRatio,
		"max_leverage":             state.humanAcct.MaxLeverage,
		"positions":                positions,
	}
}

func setupRoutes(mux *http.ServeMux, db *storage.DB, state *serverState) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		state.mu.RLock()
		defer state.mu.RUnlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "healthy",
			"tick":   state.currentTick,
			"uptime": time.Since(state.startTime).String(),
		})
	})

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		state.mu.RLock()
		defer state.mu.RUnlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tick":         state.currentTick,
			"uptime":       time.Since(state.startTime).Round(time.Second).String(),
			"companies":    len(state.companies),
			"total_trades": state.totalTrades,
			"total_volume": state.totalVolume,
		})
	})

	mux.HandleFunc("/api/prices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		state.mu.RLock()
		defer state.mu.RUnlock()
		items := state.buildPriceItems()
		_ = json.NewEncoder(w).Encode(items)
	})

	mux.HandleFunc("/api/trades", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		sym := r.URL.Query().Get("symbol")
		if sym == "" {
			http.Error(w, `{"error":"symbol parameter required"}`, http.StatusBadRequest)
			return
		}
		trades, err := db.GetRecentTrades(sym, 50)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		if trades == nil {
			trades = []storage.TradeRecord{}
		}
		_ = json.NewEncoder(w).Encode(trades)
	})

	mux.HandleFunc("/api/candles", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		sym := r.URL.Query().Get("symbol")
		tf := r.URL.Query().Get("tf")
		if tf == "" {
			tf = r.URL.Query().Get("timeframe")
		}
		if sym == "" {
			http.Error(w, `{"error":"symbol parameter required"}`, http.StatusBadRequest)
			return
		}
		if tf == "" {
			tf = TF1m
		}
		normTF := NormalizeTimeframe(tf)

		limit := 3000
		if lStr := r.URL.Query().Get("limit"); lStr != "" {
			if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		var candles []storage.CandleRecord
		if stb := state.tfManager.Get(sym); stb != nil {
			candles = stb.GetBars(normTF, limit)
		} else {
			var err error
			candles, err = db.GetCandles(sym, normTF, limit)
			if err != nil || candles == nil {
				candles = []storage.CandleRecord{}
			}
		}
		_ = json.NewEncoder(w).Encode(candles)
	})

	mux.HandleFunc("/api/sim/speed", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			Multiplier float64 `json:"multiplier"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		state.mu.Lock()
		if req.Multiplier > 0 {
			state.speedMultiplier = req.Multiplier
		}
		mult := state.speedMultiplier
		state.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"speed_multiplier": mult})
	})

	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		events, err := db.GetRecentEvents(50)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		if events == nil {
			events = []storage.EventRecord{}
		}
		_ = json.NewEncoder(w).Encode(events)
	})

	mux.HandleFunc("/api/sim/ipo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Sector  string  `json:"sector"`
			CapTier string  `json:"cap_tier"`
			Revenue float64 `json:"revenue"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		state.mu.Lock()
		defer state.mu.Unlock()

		sector := company.InformationTechnology
		if req.Sector != "" {
			for _, s := range company.AllSectors {
				if strings.EqualFold(s.String(), req.Sector) {
					sector = s
					break
				}
			}
		} else {
			sector = company.AllSectors[state.rng.Intn(len(company.AllSectors))]
		}

		tier := company.MegaCap
		if req.CapTier != "" {
			switch strings.ToLower(strings.ReplaceAll(req.CapTier, " ", "")) {
			case "smallcap":
				tier = company.SmallCap
			case "midcap":
				tier = company.MidCap
			case "largecap":
				tier = company.LargeCap
			default:
				tier = company.MegaCap
			}
		}

		revenue := req.Revenue
		if revenue <= 0 {
			// Multi-billion dollar enterprise revenue between $8B and $45B
			revenue = 8_000_000_000 + state.rng.Float64()*37_000_000_000
		}

		ipoCo := company.GenerateIPOCompany(sector, tier, revenue, state.currentTick, state.rng, state.usedSymbols)
		state.usedSymbols[ipoCo.Symbol] = true

		state.sim.ListIPO(ipoCo, 50_000_000)
		state.sim.AddParticipant(state.humanTrader, ipoCo.Symbol)
		state.companies = append(state.companies, ipoCo)

		// Register in MultiTimeframeManager and seed historical data
		if state.tfManager != nil {
			stb := state.tfManager.RegisterSymbol(ipoCo.Symbol)
			stb.InitLiveCandles(ipoCo.Symbol, ipoCo.IPOPrice, state.simTime)
		}

		// Provision retail crowd with high attention score for the newly listed IPO
		retailCohort := retail.GenerateWatchlistPool(60, state.rng.Int63(), []*company.Company{ipoCo}, 1, 1, 2500, 50000)
		for _, bot := range retailCohort {
			state.sim.AddParticipant(bot, ipoCo.Symbol)
		}

		if state.aggregators != nil {
			state.aggregators[ipoCo.Symbol] = technicals.NewAggregator(state.candlesWindow, 500)
			state.lastBarCount[ipoCo.Symbol] = 0
		}

		log.Printf("🚀 NEW ENTERPRISE IPO LISTED: %s (%s) - Sector: %s, Tier: %s, Revenue: $%.2fB, Offering Price: $%.2f",
			ipoCo.Name, ipoCo.Symbol, ipoCo.Sector.String(), ipoCo.CapTier.String(), ipoCo.AnnualRevenue/1e9, ipoCo.IPOPrice)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":             "listed",
			"symbol":             ipoCo.Symbol,
			"name":               ipoCo.Name,
			"sector":             ipoCo.Sector.String(),
			"cap_tier":           ipoCo.CapTier.String(),
			"annual_revenue":     ipoCo.AnnualRevenue,
			"net_margin":         ipoCo.NetMargin,
			"sector_multiple":    ipoCo.SectorMultiple,
			"market_cap":         ipoCo.MarketCap(),
			"shares_outstanding": ipoCo.SharesOutstanding,
			"float_shares":       ipoCo.Float,
			"ipo_price":          ipoCo.IPOPrice,
			"ipo_tick":           state.currentTick,
		})
	})

	mux.HandleFunc("/api/book", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		sym := r.URL.Query().Get("symbol")
		if sym == "" {
			http.Error(w, `{"error":"symbol parameter required"}`, http.StatusBadRequest)
			return
		}
		state.mu.RLock()
		defer state.mu.RUnlock()
		book := state.sim.Book(sym)
		if book == nil {
			http.Error(w, `{"error":"symbol not found"}`, http.StatusNotFound)
			return
		}
		bids, asks := book.TopLevels(15)
		mid, hasMid := book.MidPrice()
		spread, _ := book.Spread()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"symbol":  sym,
			"has_mid": hasMid,
			"mid":     mid,
			"spread":  spread,
			"bids":    bids,
			"asks":    asks,
		})
	})

	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Symbol   string  `json:"symbol"`
			Side     string  `json:"side"` // "BUY" or "SELL"
			Type     string  `json:"type"` // "MARKET" or "LIMIT"
			Price    float64 `json:"price"`
			Quantity float64 `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		if req.Symbol == "" || req.Quantity <= 0 {
			http.Error(w, `{"error":"symbol and positive quantity required"}`, http.StatusBadRequest)
			return
		}
		side := market.Buy
		if req.Side == "SELL" {
			side = market.Sell
		}
		isMarket := req.Type == "MARKET" || req.Price <= 0

		state.mu.Lock()
		state.humanTrader.SubmitOrderForSymbol(req.Symbol, &market.Order{
			AgentID:  "human_trader",
			Side:     side,
			Price:    req.Price,
			Quantity: req.Quantity,
			IsMarket: isMarket,
		})
		state.mu.Unlock()

		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   "queued",
			"symbol":   req.Symbol,
			"side":     req.Side,
			"quantity": req.Quantity,
			"price":    req.Price,
			"type":     req.Type,
		})
	})

	mux.HandleFunc("/api/portfolio", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		state.mu.RLock()
		defer state.mu.RUnlock()
		dto := state.buildPortfolioDTO()
		_ = json.NewEncoder(w).Encode(dto)
	})

	mux.HandleFunc("/api/portfolio/close", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"POST required"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Symbol string `json:"symbol"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		state.mu.Lock()
		defer state.mu.Unlock()
		pos, ok := state.humanAcct.Positions[req.Symbol]
		if !ok || pos.Quantity <= 0 {
			http.Error(w, `{"error":"no open position for symbol"}`, http.StatusBadRequest)
			return
		}
		closeSide := market.Sell
		if pos.Side == market.Short {
			closeSide = market.Buy
		}
		state.humanTrader.SubmitOrderForSymbol(req.Symbol, &market.Order{
			AgentID:  "human_trader",
			Side:     closeSide,
			Quantity: pos.Quantity,
			IsMarket: true,
		})
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   "close_submitted",
			"symbol":   req.Symbol,
			"quantity": pos.Quantity,
		})
	})

	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		name := "Retail Trader"
		if req.Email != "" {
			name = req.Email
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id":          "human_trader",
			"name":             name,
			"account_id":       "human_trader",
			"starting_capital": 100000,
			"max_leverage":     5,
		})
	})

	mux.HandleFunc("/api/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		name := "Retail Trader"
		if req.Name != "" {
			name = req.Name
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id":          "human_trader",
			"name":             name,
			"account_id":       "human_trader",
			"starting_capital": 100000,
			"max_leverage":     5,
		})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service": "EconomyMarkets Simulation API",
			"status":  "online",
			"tui":     "Launch interactive terminal client via ./bin/economy or go run ./cmd/tui",
		})
	})
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
