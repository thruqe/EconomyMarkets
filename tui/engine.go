package tui

import (
	"fmt"
	"maps"
	"math"
	"math/rand"
	"sync"
	"time"

	"economy/agent"
	"economy/citizen"
	"economy/company"
	"economy/country"
	"economy/market"
	"economy/retail"
	"economy/sim"
	"economy/world"
)

var tfDurations = map[string]int64{
	"1m":  60 * 1000,
	"5m":  5 * 60 * 1000,
	"15m": 15 * 60 * 1000,
	"30m": 30 * 60 * 1000,
	"1h":  60 * 60 * 1000,
	"2h":  2 * 60 * 60 * 1000,
	"4h":  4 * 60 * 60 * 1000,
	"1D":  24 * 60 * 60 * 1000,
}

type Engine struct {
	mu              sync.RWMutex
	simulation      *sim.Simulation
	universe        []*company.Company
	humanTrader     *retail.HumanTrader
	humanAcct       *market.Account
	latestPrices    map[string]sim.PriceQuote
	candles         map[string]map[string][]CandleView // [symbol][timeframe][]CandleView
	usedSymbols     map[string]bool
	rng             *rand.Rand
	speedMultiplier float64
	tickInterval    time.Duration
	simTime         int64
	running         bool
	stopChan        chan struct{}
	tickChan        chan TickMsg
	worldSim        *world.World // multi-country world simulation
	statePath       string
}

func NewEngine(universeSize int, seed int64, tickInterval time.Duration, tickChan chan TickMsg) *Engine {
	return NewEngineWithPath(universeSize, seed, tickInterval, tickChan, DefaultWorldStatePath)
}

func NewEngineWithPath(universeSize int, seed int64, tickInterval time.Duration, tickChan chan TickMsg, statePath string) *Engine {
	if statePath == "" {
		statePath = DefaultWorldStatePath
	}
	rng := rand.New(rand.NewSource(seed))

	var universe []*company.Company
	usedSymbols := make(map[string]bool)
	var savedWorld *SavedWorldState

	sw, err := LoadWorldState(statePath)
	if err == nil && sw != nil && len(sw.Companies) > 0 {
		savedWorld = sw
		universe = make([]*company.Company, len(sw.Companies))
		for i, sc := range sw.Companies {
			co := sc.ConvertToCompany()
			universe[i] = co
			usedSymbols[co.Symbol] = true
		}
	} else {
		params := company.DefaultGenerationParams()
		universe = company.GenerateUniverse(universeSize, seed, params)
		for _, co := range universe {
			usedSymbols[co.Symbol] = true
		}
	}

	s := sim.NewSimulation(5, 200)
	if savedWorld != nil {
		s.SetTick(savedWorld.Tick)
		if s.Citizen != nil {
			s.Citizen.Demographics.TotalPopulation = savedWorld.Macro.Population
			s.Citizen.Demographics.EmployedCount = savedWorld.Macro.EmployedCount
			s.Citizen.Demographics.AverageHourlyEarnings = savedWorld.Macro.AverageHourlyEarnings
			s.Citizen.Demographics.AggregateDisposableIncome = savedWorld.Macro.AggregateDisposable
			s.Citizen.Sentiment.Index = savedWorld.Macro.ConsumerConfidence
			s.Citizen.Sentiment.Happiness = savedWorld.Macro.Happiness
			s.Citizen.Consumer.TotalConsumerSpending = savedWorld.Macro.ConsumerSpending
		}
		if s.National != nil {
			s.National.NominalGDP = savedWorld.Macro.GDP
			s.National.RealGDPGrowth = savedWorld.Macro.RealGDPGrowth
			s.National.CentralBank.FedFundsRate = savedWorld.Macro.FedFundsRate
			s.National.CentralBank.TenYearYield = savedWorld.Macro.TenYearYield
			s.National.CentralBank.CPIInflationRate = savedWorld.Macro.CPIInflationRate
			s.National.CentralBank.PolicyStance = savedWorld.Macro.PolicyStance
			s.National.Trade.DollarIndexDXY = savedWorld.Macro.DollarIndexDXY
			s.National.Trade.TradeBalance = savedWorld.Macro.TradeBalance
			s.National.Fiscal.NationalDebt = savedWorld.Macro.NationalDebt
			if savedWorld.Macro.ActivePolicies != nil {
				s.National.Fiscal.ActivePolicies = savedWorld.Macro.ActivePolicies
			}
			s.National.Fiscal.TreasuryCash = savedWorld.Macro.TreasuryCash
			s.National.Fiscal.CreditRating = savedWorld.Macro.CreditRating
			s.National.Fiscal.BorrowingYield = savedWorld.Macro.BorrowingYield
			s.National.Fiscal.InfrastructureLevel = savedWorld.Macro.InfrastructureLevel
			s.National.Fiscal.HealthcareLevel = savedWorld.Macro.HealthcareLevel
			s.National.Fiscal.EducationLevel = savedWorld.Macro.EducationLevel
			s.National.Fiscal.EnterpriseGrantsLevel = savedWorld.Macro.EnterpriseGrantsLevel
			s.National.Fiscal.ExportCapacityLevel = savedWorld.Macro.ExportCapacityLevel
			s.National.Fiscal.FDIInflow = savedWorld.Macro.FDIInflow
			s.National.Fiscal.ExportRevenue = savedWorld.Macro.ExportRevenue
			s.National.Fiscal.ExchangeChartered = savedWorld.Macro.ExchangeChartered
			// If existing save has many companies, keep exchange chartered
			if len(savedWorld.Companies) > 20 {
				s.National.Fiscal.ExchangeChartered = true
			}
		}
	} else {
		// Fresh start in Nation Building mode
		if s.Citizen != nil {
			s.Citizen.Demographics.TotalPopulation = 4_500_000.0
			s.Citizen.Demographics.WorkingAgePopulation = 2_800_000.0
			s.Citizen.Demographics.EmployedCount = 2_500_000.0
			s.Citizen.Demographics.AverageHourlyEarnings = 18.50
			s.Citizen.Demographics.AggregateDisposableIncome = 2_500_000.0 * 18.50 * 2000.0 * 0.85
			s.Citizen.Sentiment.Index = 68.0
			s.Citizen.Sentiment.Happiness = 72.0
		}
		if s.National != nil {
			s.National.NominalGDP = 14_000_000_000.0 // $14B initial GDP
			s.National.RealGDPGrowth = 0.045
			s.National.Fiscal.MandatorySpending = 1_200_000_000.0   // $1.2B mandatory spending
			s.National.Fiscal.DiscretionarySpending = 800_000_000.0 // $800M discretionary spending
			s.National.Fiscal.TreasuryCash = 10_000_000_000.0       // $10B starting sovereign treasury
			s.National.Fiscal.NationalDebt = 5_000_000_000.0        // $5B starting sovereign debt
			s.National.Fiscal.InfrastructureLevel = 15.0
			s.National.Fiscal.HealthcareLevel = 20.0
			s.National.Fiscal.EducationLevel = 20.0
			s.National.Fiscal.EnterpriseGrantsLevel = 12.0
			s.National.Fiscal.ExportCapacityLevel = 15.0
			s.National.Fiscal.FDIInflow = 1_200_000_000.0
			s.National.Fiscal.ExportRevenue = 2_500_000_000.0
			s.National.Fiscal.ExchangeChartered = false

			// Mark early cohort as emerging private domestic ventures until chartered
			for i, co := range universe {
				if i < 20 {
					co.IsPublic = false
					co.Stage = "Seed"
					if i%2 == 0 {
						co.Stage = "Growth"
					}
					co.PrivateValuation = co.AnnualRevenue * co.SectorMultiple
				}
			}
		}
	}

	if s.Citizen != nil && s.National != nil {
		citRep := s.Citizen.Tick(s.National.Labor.EmployedWorkers, s.National.Labor.AverageHourlyWage, s.National.CentralBank.CPIInflationRate, s.National.Labor.AnnualWageGrowth, 0.0, 1.0/252.0)
		natRep := s.National.Tick(citRep.ConsumerSpending, 500_000_000.0, s.National.Labor.EmployedWorkers, 100_000_000.0, citRep.TaxesPaidThisTick, 1.0/252.0)
		s.CitizenReport = citRep
		s.NationalReport = natRep
	}

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
		hf.MinTradeThreshold = 0.003
		hf.FullConvictionThreshold = 0.025
		hf.RebalanceThreshold = 0.003
		hf.ExecutionRate = 0.06
		hf.MaxOrderShares = 200
		s.AddParticipant(hf, co.Symbol)

		bankAcct := market.NewAccount("bank_"+co.Symbol, 3_000_000, 3, 0.15)
		bank := agent.NewBank("bank_"+co.Symbol, bankAcct, []*company.Company{co})
		bank.MinTradeThreshold = 0.006
		bank.FullConvictionThreshold = 0.040
		bank.RebalanceThreshold = 0.006
		bank.ExecutionRate = 0.05
		bank.MaxOrderShares = 250
		s.AddParticipant(bank, co.Symbol)
	}

	// Retail bots
	pool := retail.GenerateWatchlistPool(100, seed, universe, 1, 3, 1000, 20000)
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

	// Human Trader
	humanAcct := market.NewAccount("human_trader", 100_000, 5, 0.10)
	if saved, err := LoadPortfolio(DefaultPortfolioPath); err == nil && saved != nil {
		ApplySavedPortfolioToAccount(saved, humanAcct)
	}
	humanTrader := retail.NewHumanTrader("human_trader", humanAcct)
	s.AddParticipant(humanTrader, s.Symbols()...)

	nowMillis := time.Now().UnixMilli()
	if savedWorld != nil && savedWorld.SimTime > 0 {
		nowMillis = savedWorld.SimTime
	}

	// Initialize multi-country world simulation
	worldSim := world.NewWorld(seed)
	if savedWorld != nil {
		// Restore home country configuration
		if savedWorld.HomeCountry.Configured && savedWorld.HomeCountry.Name != "" {
			worldSim.Home.Configure(
				savedWorld.HomeCountry.Name,
				savedWorld.HomeCountry.Currency,
				savedWorld.HomeCountry.FlagCode,
				savedWorld.HomeCountry.Founded,
			)
		}
		// Restore foreign country states
		for _, sfc := range savedWorld.Foreign {
			if fc, ok := worldSim.Foreign[sfc.ID]; ok {
				fc.GDP = sfc.GDP
				fc.GDPGrowth = sfc.GDPGrowth
				fc.Inflation = sfc.Inflation
				fc.NationalDebt = sfc.NationalDebt
				fc.DebtToGDP = sfc.DebtToGDP
				fc.TradeBalance = sfc.TradeBalance
				fc.Relation.TariffRate = sfc.TariffRate
				fc.Relation.Stance = world.DiplomaticStance(sfc.Stance)
				fc.Relation.TradeVolume = sfc.TradeVol
				fc.Relation.AidFlow = sfc.AidFlow
				fc.Relation.SanctionsLevel = sfc.SanctionsLevel
			}
		}
		// Restore policy log
		if len(savedWorld.PolicyLog) > 0 {
			worldSim.PolicyLog = append(worldSim.PolicyLog, savedWorld.PolicyLog...)
		}
	}

	e := &Engine{
		simulation:      s,
		universe:        universe,
		humanTrader:     humanTrader,
		humanAcct:       humanAcct,
		latestPrices:    make(map[string]sim.PriceQuote),
		candles:         make(map[string]map[string][]CandleView),
		usedSymbols:     usedSymbols,
		rng:             rng,
		speedMultiplier: 1.0,
		tickInterval:    tickInterval,
		simTime:         nowMillis,
		stopChan:        make(chan struct{}),
		tickChan:        tickChan,
		worldSim:        worldSim,
		statePath:       statePath,
	}

	// Seed or restore candles
	if savedWorld != nil && len(savedWorld.Candles) > 0 {
		for sym, tfMap := range savedWorld.Candles {
			e.candles[sym] = make(map[string][]CandleView)
			for tf, scList := range tfMap {
				var cvList []CandleView
				for _, sc := range scList {
					cvList = append(cvList, CandleView{
						Time:   time.UnixMilli(sc.TimeUnixMilli),
						Open:   sc.Open,
						High:   sc.High,
						Low:    sc.Low,
						Close:  sc.Close,
						Volume: sc.Volume,
					})
				}
				e.candles[sym][tf] = cvList
			}
		}
	} else {
		// Initialize baseline candle points across timeframes for every company
		for _, co := range universe {
			e.candles[co.Symbol] = make(map[string][]CandleView)
			p := co.ReportedValue
			if p <= 0 {
				p = co.TrueValue
			}
			for tf, dur := range tfDurations {
				numPoints := 40
				tfCandles := make([]CandleView, numPoints)
				curClose := p
				for i := numPoints - 1; i >= 0; i-- {
					t := time.UnixMilli(nowMillis - int64(numPoints-1-i)*dur)
					noise := (rng.Float64() - 0.485) * (curClose * 0.003)
					prevClose := curClose - noise
					if prevClose < 1.0 {
						prevClose = 1.0
					}
					openP := prevClose
					closeP := curClose
					highP := math.Max(openP, closeP) + rng.Float64()*(curClose*0.001)
					lowP := math.Min(openP, closeP) - rng.Float64()*(curClose*0.001)
					tfCandles[i] = CandleView{
						Time:   t,
						Open:   openP,
						High:   highP,
						Low:    lowP,
						Close:  closeP,
						Volume: 500 + rng.Float64()*2500,
					}
					curClose = prevClose
				}
				e.candles[co.Symbol][tf] = tfCandles
			}
		}
	}

	return e
}

func (e *Engine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.mu.Unlock()

	go e.runLoop()
}

func (e *Engine) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}
	e.running = false
	close(e.stopChan)
	e.mu.Unlock()
	_ = e.SaveWorldState()
	_ = e.SavePortfolio()
}

func (e *Engine) runLoop() {
	ticker := time.NewTicker(e.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			return
		case <-ticker.C:
			e.mu.RLock()
			speed := e.speedMultiplier
			e.mu.RUnlock()

			if speed <= 0 {
				continue // Paused
			}

			// Execute simulation step
			report := e.simulation.Step()

			e.mu.Lock()
			e.latestPrices = report.Prices
			// 1 tick = 6 hours of simulated time at 1.0x (4 ticks/sec = 24 simulated hours/real-sec)
			simMillisPerTick := int64(6 * 3600 * 1000)
			e.simTime += int64(float64(simMillisPerTick) * speed)
			curSimTime := e.simTime

			// Update in-memory multi-timeframe candles for each symbol
			for sym, p := range report.Prices {
				if p.HasMid && p.Mid > 0 {
					symCandles, exists := e.candles[sym]
					if !exists {
						symCandles = make(map[string][]CandleView)
						e.candles[sym] = symCandles
					}
					for tf, dur := range tfDurations {
						history := symCandles[tf]
						bucketStart := (curSimTime / dur) * dur
						bucketTime := time.UnixMilli(bucketStart)

						if len(history) == 0 {
							symCandles[tf] = []CandleView{
								{Time: bucketTime, Open: p.Mid, High: p.Mid, Low: p.Mid, Close: p.Mid, Volume: 0},
							}
						} else {
							lastIdx := len(history) - 1
							last := &history[lastIdx]
							if bucketStart > last.Time.UnixMilli() {
								// Rollover to new candle for this timeframe
								symCandles[tf] = append(symCandles[tf], CandleView{
									Time:   bucketTime,
									Open:   p.Mid,
									High:   p.Mid,
									Low:    p.Mid,
									Close:  p.Mid,
									Volume: 0,
								})
								if len(symCandles[tf]) > 100 {
									symCandles[tf] = symCandles[tf][len(symCandles[tf])-100:]
								}
							} else {
								// Update current candle
								if p.Mid > last.High {
									last.High = p.Mid
								}
								if p.Mid < last.Low {
									last.Low = p.Mid
								}
								last.Close = p.Mid
							}
						}
					}
				}
			}

			// Add trade volume to current candle across timeframes
			for _, tr := range report.Trades {
				if symCandles, ok := e.candles[tr.Symbol]; ok {
					for _, tfList := range symCandles {
						if len(tfList) > 0 {
							tfList[len(tfList)-1].Volume += tr.Quantity
						}
					}
				}
			}

			activeUniverseCount := len(e.universe)

			// Route healthcare and education levels to citizen demographics
			if e.simulation.Citizen != nil && e.simulation.National != nil {
				e.simulation.Citizen.SetPublicFactors(
					e.simulation.National.Fiscal.HealthcareLevel,
					e.simulation.National.Fiscal.EducationLevel,
				)
			}

			// Advance world simulation (foreign countries, diplomatic relations)
			tickFracOfYear := float64(e.tickInterval.Milliseconds()) * speed / (365.25 * 24 * 3600 * 1000)
			if e.worldSim != nil {
				e.worldSim.Tick(tickFracOfYear)
			}

			// Probabilistic emerging company spawning
			// Base: ~1 new company per 500 ticks at 1x speed, scaled by GDP growth & enterprise grants
			gdpGrowth := e.simulation.NationalReport.RealGDPGrowth
			baseSpawnProb := 1.0 / 450.0
			if e.simulation.National != nil {
				baseSpawnProb *= (1.0 + e.simulation.National.Fiscal.EnterpriseGrantsLevel/50.0)
			}
			if gdpGrowth > 0.03 {
				baseSpawnProb *= 1.5 // Faster growth → more company formation
			} else if gdpGrowth < 0.0 {
				baseSpawnProb *= 0.3 // Recession → fewer startups
			}
			baseSpawnProb *= speed // Scale with simulation speed
			if e.rng.Float64() < baseSpawnProb && len(e.universe) < 1000 {
				sectors := []company.Sector{
					company.InformationTechnology, company.HealthCare, company.Financials,
					company.ConsumerDiscretionary, company.Industrials, company.Energy,
					company.Materials, company.CommunicationServices, company.Utilities,
					company.RealEstate, company.ConsumerStaples,
				}
				sector := sectors[e.rng.Intn(len(sectors))]
				tiers := []company.CapTier{company.SmallCap, company.MidCap, company.SmallCap}
				tier := tiers[e.rng.Intn(len(tiers))]

				isChartered := false
				if e.simulation.National != nil {
					isChartered = e.simulation.National.Fiscal.ExchangeChartered
				}

				if !isChartered {
					// Emerging private domestic enterprise
					seedRev := 4_000_000 + e.rng.Float64()*36_000_000
					newCo := company.GeneratePrivateEnterprise(sector, seedRev, int(report.Tick), e.rng, e.usedSymbols)
					e.usedSymbols[newCo.Symbol] = true
					e.universe = append(e.universe, newCo)
				} else {
					// Public IPO listing on chartered exchange
					revenue := 500_000_000 + e.rng.Float64()*4_500_000_000
					newCo := company.GenerateIPOCompany(sector, tier, revenue, int(report.Tick), e.rng, e.usedSymbols)
					newCo.IsIPO = true
					newCo.IPOTick = int(report.Tick)
					newCo.IsPublic = true
					newCo.Stage = "Public"
					e.usedSymbols[newCo.Symbol] = true
					e.universe = append(e.universe, newCo)
					e.simulation.ListIPO(newCo, 10_000_000)
					e.simulation.AddParticipant(e.humanTrader, newCo.Symbol)
					// Initialize candles for the new company
					e.candles[newCo.Symbol] = make(map[string][]CandleView)
					initP := newCo.ReportedValue
					if initP <= 0 {
						initP = 15.0
					}
					for tf, dur := range tfDurations {
						var tfCandles []CandleView
						for i := 15; i >= 0; i-- {
							t := time.UnixMilli(curSimTime - int64(i)*dur)
							tfCandles = append(tfCandles, CandleView{
								Time: t, Open: initP, High: initP, Low: initP, Close: initP, Volume: 0,
							})
						}
						e.candles[newCo.Symbol][tf] = tfCandles
					}
				}
			}


			e.mu.Unlock()

			// Check if human trader executed any trades this tick
			for _, tr := range report.Trades {
				if tr.TakerAgentID == "human_trader" || tr.MakerAgentID == "human_trader" {
					_ = e.SavePortfolio()
					break
				}
			}

			// Periodic auto-save of world state and portfolio every 50 ticks
			if report.Tick%50 == 0 {
				_ = e.SaveWorldState()
				_ = e.SavePortfolio()
			}

			// Broadcast tick to TUI
			select {
			case e.tickChan <- TickMsg{
				Report:      report,
				SimTime:     curSimTime,
				Speed:       speed,
				ActiveCount: activeUniverseCount,
			}:
			default:
				// Skip if channel full
			}
		}
	}
}

// SimBridge implementation

func (e *Engine) GetUniverse() []*company.Company {
	e.mu.RLock()
	defer e.mu.RUnlock()
	res := make([]*company.Company, len(e.universe))
	copy(res, e.universe)
	return res
}

func (e *Engine) GetPrices() map[string]sim.PriceQuote {
	e.mu.RLock()
	defer e.mu.RUnlock()
	res := make(map[string]sim.PriceQuote, len(e.latestPrices))
	maps.Copy(res, e.latestPrices)
	return res
}

func (e *Engine) GetBook(symbol string) *market.OrderBook {
	return e.simulation.Book(symbol)
}

func (e *Engine) GetHumanAccount() *market.Account {
	return e.humanAcct
}

func (e *Engine) GetCandles(symbol string, tf string) []CandleView {
	e.mu.RLock()
	defer e.mu.RUnlock()
	symCandles, ok := e.candles[symbol]
	if !ok {
		return nil
	}
	source, ok := symCandles[tf]
	if !ok {
		source = symCandles["1m"]
	}
	res := make([]CandleView, len(source))
	copy(res, source)
	return res
}

func (e *Engine) SubmitOrder(symbol string, side market.Side, isMarket bool, qty float64, price float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ord := &market.Order{
		ID:       uint64(time.Now().UnixNano()),
		AgentID:  e.humanTrader.Account().AgentID,
		Side:     side,
		IsMarket: isMarket,
		Quantity: qty,
		Price:    price,
	}

	e.humanTrader.SubmitOrderForSymbol(symbol, ord)
	return nil
}

func (e *Engine) ClosePosition(symbol string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	pos, ok := e.humanAcct.Positions[symbol]
	if !ok || pos.Quantity == 0 {
		return fmt.Errorf("no open position in %s", symbol)
	}

	side := market.Sell
	if pos.Quantity < 0 {
		side = market.Buy
	}
	qty := mathAbs(pos.Quantity)

	ord := &market.Order{
		ID:       uint64(time.Now().UnixNano()),
		AgentID:  e.humanTrader.Account().AgentID,
		Side:     side,
		IsMarket: true,
		Quantity: qty,
	}

	e.humanTrader.SubmitOrderForSymbol(symbol, ord)
	return nil
}

func mathAbs(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}

func (e *Engine) SetSpeed(speed float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.speedMultiplier = speed
}

func (e *Engine) LaunchIPO(name, symbol string, sector company.Sector, tier company.CapTier, shares, price float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.usedSymbols[symbol] {
		candidate := symbol
		suffix := 2
		for e.usedSymbols[candidate] {
			candidate = fmt.Sprintf("%s%d", symbol, suffix)
			suffix++
		}
		symbol = candidate
	}

	revenue := 8_000_000_000 + e.rng.Float64()*37_000_000_000
	ipoCo := company.GenerateIPOCompany(sector, tier, revenue, 1, e.rng, e.usedSymbols)
	if symbol != "" {
		ipoCo.Symbol = symbol
	}
	if name != "" {
		ipoCo.Name = name
	}
	if price > 0 {
		ipoCo.IPOPrice = price
		ipoCo.ReportedValue = price
		ipoCo.TrueValue = price
	}
	if shares > 0 {
		ipoCo.SharesOutstanding = shares
		ipoCo.Float = shares * 0.70
	}

	e.usedSymbols[ipoCo.Symbol] = true
	e.universe = append(e.universe, ipoCo)

	// List in simulation
	e.simulation.ListIPO(ipoCo, 50_000_000)
	e.simulation.AddParticipant(e.humanTrader, ipoCo.Symbol)

	// Retail bots
	retailCohort := retail.GenerateWatchlistPool(30, e.rng.Int63(), []*company.Company{ipoCo}, 1, 1, 2500, 50000)
	for _, bot := range retailCohort {
		e.simulation.AddParticipant(bot, ipoCo.Symbol)
	}

	// Initialize candles across timeframes for the newly listed IPO
	e.candles[ipoCo.Symbol] = make(map[string][]CandleView)
	for tf, dur := range tfDurations {
		var tfCandles []CandleView
		curP := price
		for i := 15; i >= 0; i-- {
			t := time.UnixMilli(e.simTime - int64(i)*dur)
			tfCandles = append(tfCandles, CandleView{
				Time:   t,
				Open:   curP,
				High:   curP,
				Low:    curP,
				Close:  curP,
				Volume: shares * 0.001,
			})
		}
		e.candles[ipoCo.Symbol][tf] = tfCandles
	}

	return nil
}

func (e *Engine) SavePortfolio() error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	saved := ExtractSavedPortfolioFromAccount(e.humanAcct, nil)
	return SavePortfolio(DefaultPortfolioPath, saved)
}

func (e *Engine) ResetPortfolio() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.humanAcct.Cash = 100_000.0
	clear(e.humanAcct.Positions)
	return ResetPortfolioDisk(DefaultPortfolioPath)
}

func (e *Engine) GetMacroReport() (citizen.CitizenReport, country.NationalReport) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.simulation != nil {
		citRep := e.simulation.CitizenReport
		natRep := e.simulation.NationalReport
		if e.simulation.National != nil {
			fisc := e.simulation.National.Fiscal
			natRep.TreasuryCash = fisc.TreasuryCash
			natRep.NationalDebt = fisc.NationalDebt
			natRep.CreditRating = fisc.CreditRating
			natRep.BorrowingYield = fisc.BorrowingYield
			natRep.InfrastructureLevel = fisc.InfrastructureLevel
			natRep.HealthcareLevel = fisc.HealthcareLevel
			natRep.EducationLevel = fisc.EducationLevel
			natRep.EnterpriseGrantsLevel = fisc.EnterpriseGrantsLevel
			natRep.ExportCapacityLevel = fisc.ExportCapacityLevel
			natRep.FDIInflow = fisc.FDIInflow
			natRep.ExportRevenue = fisc.ExportRevenue
			natRep.ExchangeChartered = fisc.ExchangeChartered
		}
		return citRep, natRep
	}
	return citizen.CitizenReport{}, country.NationalReport{}
}

func (e *Engine) SaveWorldState() error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.simulation == nil {
		return nil
	}

	savedCos := make([]SavedCompany, len(e.universe))
	for i, co := range e.universe {
		mid := co.ReportedValue
		bid := mid * 0.999
		ask := mid * 1.001
		if bk := e.simulation.Book(co.Symbol); bk != nil {
			if m, ok := bk.MidPrice(); ok && m > 0 {
				mid = m
			}
			if b, ok := bk.BestBid(); ok && b > 0 {
				bid = b
			}
			if a, ok := bk.BestAsk(); ok && a > 0 {
				ask = a
			}
		}

		savedCos[i] = SavedCompany{
			Symbol:            co.Symbol,
			Name:              co.Name,
			Sector:            int(co.Sector),
			CapTier:           int(co.CapTier),
			TrueValue:         co.TrueValue,
			ReportedValue:     co.ReportedValue,
			SharesOutstanding: co.SharesOutstanding,
			Float:             co.Float,
			AnnualRevenue:     co.AnnualRevenue,
			NetMargin:         co.NetMargin,
			SectorMultiple:    co.SectorMultiple,
			IPOPrice:          co.IPOPrice,
			IsIPO:             co.IsIPO,
			IPOTick:           co.IPOTick,
			Headcount:         co.Headcount,
			AverageWage:       co.AverageWage,
			LaborExpense:      co.LaborExpense,
			DebtOutstanding:   co.DebtOutstanding,
			InterestExpense:   co.InterestExpense,
			CorporateTaxPaid:  co.CorporateTaxPaid,
			CapEx:             co.CapEx,
			MacroDemandFactor: co.MacroDemandFactor,
			MidPrice:          mid,
			Bid:               bid,
			Ask:               ask,
			IsPublic:          co.IsPublic,
			Stage:             co.Stage,
			PrivateValuation:  co.PrivateValuation,
		}
	}

	savedCandles := make(map[string]map[string][]SavedCandle)
	for sym, tfMap := range e.candles {
		savedCandles[sym] = make(map[string][]SavedCandle)
		for tf, cvList := range tfMap {
			var scList []SavedCandle
			for _, cv := range cvList {
				scList = append(scList, SavedCandle{
					TimeUnixMilli: cv.Time.UnixMilli(),
					Open:          cv.Open,
					High:          cv.High,
					Low:           cv.Low,
					Close:         cv.Close,
					Volume:        cv.Volume,
				})
			}
			savedCandles[sym][tf] = scList
		}
	}

	citRep, natRep := e.simulation.CitizenReport, e.simulation.NationalReport
	fisc := e.simulation.National.Fiscal
	savedMacro := SavedMacro{
		Population:            citRep.Population,
		EmployedCount:         citRep.EmployedCount,
		AverageHourlyEarnings: citRep.AverageHourlyEarnings,
		AggregateDisposable:   citRep.DisposableIncome,
		ConsumerConfidence:    citRep.ConsumerConfidence,
		Happiness:             citRep.Happiness,
		ConsumerSpending:      citRep.ConsumerSpending,
		GDP:                   natRep.GDP,
		RealGDPGrowth:         natRep.RealGDPGrowth,
		FedFundsRate:          natRep.FedFundsRate,
		TenYearYield:          natRep.TenYearYield,
		CPIInflationRate:      natRep.CPIInflationRate,
		PolicyStance:          natRep.PolicyStance,
		DollarIndexDXY:        natRep.DollarIndexDXY,
		TradeBalance:          natRep.TradeBalance,
		NationalDebt:          natRep.NationalDebt,
		ActivePolicies:        fisc.ActivePolicies,

		TreasuryCash:          fisc.TreasuryCash,
		CreditRating:          fisc.CreditRating,
		BorrowingYield:        fisc.BorrowingYield,
		InfrastructureLevel:   fisc.InfrastructureLevel,
		HealthcareLevel:       fisc.HealthcareLevel,
		EducationLevel:        fisc.EducationLevel,
		EnterpriseGrantsLevel: fisc.EnterpriseGrantsLevel,
		ExportCapacityLevel:   fisc.ExportCapacityLevel,
		FDIInflow:             natRep.FDIInflow,
		ExportRevenue:         natRep.ExportRevenue,
		ExchangeChartered:     fisc.ExchangeChartered,
	}

	// Persist world state
	var savedHome SavedHomeCountry
	var savedForeign []SavedForeignCountry
	var policyLog []world.ForeignPolicyAction
	if e.worldSim != nil {
		savedHome = SavedHomeCountry{
			Name:       e.worldSim.Home.Name,
			Currency:   e.worldSim.Home.Currency,
			FlagCode:   e.worldSim.Home.FlagCode,
			Founded:    e.worldSim.Home.Founded,
			Configured: e.worldSim.Home.Configured,
		}
		for _, fc := range e.worldSim.Foreign {
			savedForeign = append(savedForeign, SavedForeignCountry{
				ID: fc.ID, Name: fc.Name, FlagCode: fc.FlagCode,
				GDP: fc.GDP, GDPGrowth: fc.GDPGrowth, Inflation: fc.Inflation,
				InterestRate: fc.InterestRate, Unemployment: fc.Unemployment, Population: fc.Population,
				TradeBalance: fc.TradeBalance, CurrencyStrength: fc.CurrencyStrength,
				NationalDebt: fc.NationalDebt, DebtToGDP: fc.DebtToGDP,
				TariffRate: fc.Relation.TariffRate, Stance: int(fc.Relation.Stance),
				TradeVol: fc.Relation.TradeVolume, AidFlow: fc.Relation.AidFlow,
				SanctionsLevel: fc.Relation.SanctionsLevel,
			})
		}
		policyLog = e.worldSim.PolicyLog
	}

	state := &SavedWorldState{
		Version:     1,
		Tick:        e.simulation.Tick(),
		SimTime:     e.simTime,
		Macro:       savedMacro,
		Companies:   savedCos,
		Candles:     savedCandles,
		HomeCountry: savedHome,
		Foreign:     savedForeign,
		PolicyLog:   policyLog,
	}

	path := e.statePath
	if path == "" {
		path = DefaultWorldStatePath
	}
	return SaveWorldState(path, state)
}

func (e *Engine) ResetWorldState() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	path := e.statePath
	if path == "" {
		path = DefaultWorldStatePath
	}
	_ = ResetWorldStateDisk(path)
	return e.ResetPortfolio()
}

func (e *Engine) BailoutCompany(symbol string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.simulation == nil {
		return fmt.Errorf("simulation not running")
	}
	return e.simulation.BailoutCompany(symbol, 500_000_000.0)
}

func (e *Engine) TogglePolicy(policyID string) (bool, string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.simulation == nil {
		return false, "simulation not running"
	}
	return e.simulation.TogglePolicy(policyID)
}

// GetWorld returns the multi-country world simulation state (read-only snapshot).
func (e *Engine) GetWorld() *world.World {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.worldSim
}

// SetForeignRelation updates the diplomatic stance and tariff rate with a foreign country.
func (e *Engine) SetForeignRelation(countryID string, stance world.DiplomaticStance, tariffRate float64) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.worldSim == nil {
		return false
	}
	return e.worldSim.SetRelation(countryID, stance, tariffRate)
}

// NegotiateTradeDeal lowers the tariff rate and improves stance with a country.
func (e *Engine) NegotiateTradeDeal(countryID string, newTariff float64) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.worldSim == nil {
		return false
	}
	return e.worldSim.NegotiateDeal(countryID, newTariff, int64(e.simulation.Tick()))
}

// SendForeignAid sends an aid package improving diplomatic stance.
func (e *Engine) SendForeignAid(countryID string, amount float64) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.worldSim == nil {
		return false
	}
	return e.worldSim.SendAid(countryID, amount, int64(e.simulation.Tick()))
}

// ImposeSanctions applies sanctions on a foreign country.
func (e *Engine) ImposeSanctions(countryID string, level int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.worldSim == nil {
		return false
	}
	return e.worldSim.ImposeSanctions(countryID, level, int64(e.simulation.Tick()))
}

// GetSimDate returns the current simulated calendar date.
func (e *Engine) GetSimDate() world.SimDate {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return world.DateFromMillis(e.simTime)
}

// ConfigureHomeCountry sets up the user's home country.
func (e *Engine) ConfigureHomeCountry(name, currency, flagCode string, founded int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.worldSim != nil {
		e.worldSim.Home.Configure(name, currency, flagCode, founded)
	}
}

// BorrowMoney issues new sovereign bonds to increase Treasury Cash.
func (e *Engine) BorrowMoney(amount float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.simulation == nil || e.simulation.National == nil {
		return fmt.Errorf("simulation not running")
	}
	return e.simulation.National.Fiscal.BorrowDebt(amount)
}

// RepayDebt uses Treasury Cash to pay down National Debt.
func (e *Engine) RepayDebt(amount float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.simulation == nil || e.simulation.National == nil {
		return fmt.Errorf("simulation not running")
	}
	return e.simulation.National.Fiscal.RepayDebt(amount)
}

// InvestInfrastructure invests Treasury Cash into national infrastructure.
func (e *Engine) InvestInfrastructure(amount float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.simulation == nil || e.simulation.National == nil {
		return fmt.Errorf("simulation not running")
	}
	return e.simulation.National.Fiscal.InvestInfrastructure(amount)
}

// InvestPopulation invests Treasury Cash into healthcare or education.
func (e *Engine) InvestPopulation(amount float64, pillar string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.simulation == nil || e.simulation.National == nil {
		return fmt.Errorf("simulation not running")
	}
	return e.simulation.National.Fiscal.InvestPopulation(amount, pillar)
}

// InvestEnterprise grants seed capital to domestic startups.
func (e *Engine) InvestEnterprise(amount float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.simulation == nil || e.simulation.National == nil {
		return fmt.Errorf("simulation not running")
	}
	return e.simulation.National.Fiscal.InvestEnterpriseGrants(amount)
}

// CharterStockExchange inaugurates the official National Stock Exchange once prerequisites are met.
func (e *Engine) CharterStockExchange() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.simulation == nil || e.simulation.National == nil {
		return fmt.Errorf("simulation not running")
	}
	if err := e.simulation.National.Fiscal.CharterExchange(); err != nil {
		return err
	}
	// When chartered, convert private enterprises that have reached Pre-IPO or Growth into public listings
	for _, co := range e.universe {
		if !co.IsPublic && (co.Stage == "Pre-IPO" || co.Stage == "Growth") {
			co.IsPublic = true
			co.Stage = "Public"
			co.IsIPO = true
			if co.SharesOutstanding > 0 {
				co.IPOPrice = co.ReportedValue / co.SharesOutstanding
			}
			co.IPOTick = e.simulation.Tick()
		}
	}
	return nil
}
