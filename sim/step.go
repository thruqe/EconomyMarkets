package sim

import (
	"fmt"
	"time"

	"economy/company"
	"economy/market"
)

// maxLiquidationPasses bounds how many times the liquidation scan
// re-runs within a single Step call. A forced liquidation sell can
// itself move price enough to push another account under maintenance
// margin — this is the realistic cascade mechanic — so the scan loops
// until either no new liquidations trigger or this cap is hit,
// producing the dramatic, same-session cascade real forced-selling
// spirals exhibit, rather than one that only trickles out one
// liquidation per tick across many separate ticks. The cap exists so
// a pathological configuration (e.g. everyone simultaneously
// over-leveraged) can't loop indefinitely within one Step call; when
// hit, remaining liquidations simply continue naturally into
// subsequent ticks.
const maxLiquidationPasses = 5

// Step advances the simulation by exactly one tick, in order:
//
//  1. Advance every company's fundamentals (Company.Tick()), logging
//     any FundamentalEvent/RestatementEvent that fired.
//  2. Refresh every agent.Bank's cross-company price map
//     (SetExternalPrices) with current book mid-prices so Bank evaluates
//     multi-asset positions at real market prices during the trading round.
//  3. Build each company's MarketState from its current book and
//     rolling price history.
//  4. Poll every participant registered for each company, collecting
//     every order they return.
//  5. Submit all of this tick's orders to their respective books,
//     settling every resulting fill into both sides' accounts via
//     market.ApplySettledFill.
//  6. Refresh every agent.Bank's cross-company price map with this
//     tick's post-trade mid-prices and run the global liquidation scan
//     across every account, looping (up to maxLiquidationPasses) to let
//     cascades resolve within the tick — see maxLiquidationPasses' doc.
//  7. Record each company's new mid-price into rolling history for
//     next tick's MarketState.
//
// Returns a TickReport summarizing all trades, events, quotes, and account states from this tick.
func (s *Simulation) Step() *TickReport {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tick++
	eventsStart := len(s.EventLog)
	var tickTrades []Trade

	s.advanceFundamentals()
	s.refreshBankExternalPrices()
	s.runTradingRound(&tickTrades)
	s.refreshBankExternalPrices()
	s.runLiquidationCascade(&tickTrades)
	s.recordPrices()

	prices := make(map[string]PriceQuote, len(s.order))
	for _, symbol := range s.order {
		cm := s.markets[symbol]
		mid, hasMid := cm.book.MidPrice()
		bid, _ := cm.book.BestBid()
		ask, _ := cm.book.BestAsk()
		spread, _ := cm.book.Spread()
		prices[symbol] = PriceQuote{
			Symbol: symbol,
			Mid:    mid,
			Bid:    bid,
			Ask:    ask,
			Spread: spread,
			HasMid: hasMid,
		}
	}

	tickEvents := make([]Event, len(s.EventLog)-eventsStart)
	copy(tickEvents, s.EventLog[eventsStart:])

	currentPrices := s.currentMidPrices()
	var accts []AccountSnapshot
	for _, a := range s.allAccounts() {
		mr, _ := a.MarginRatio(currentPrices)
		accts = append(accts, AccountSnapshot{
			AgentID:     a.AgentID,
			Cash:        a.Cash,
			Equity:      a.Equity(currentPrices),
			MarginRatio: mr,
		})
	}

	return &TickReport{
		Tick:      s.tick,
		Timestamp: time.Now(),
		Prices:    prices,
		Trades:    tickTrades,
		Events:    tickEvents,
		Accounts:  accts,
		Citizen:   s.CitizenReport,
		National:  s.NationalReport,
	}
}

// ClearEventLog resets the in-memory EventLog slice to release memory in long-running simulations.
func (s *Simulation) ClearEventLog() {
	s.EventLog = s.EventLog[:0]
}

// advanceFundamentals ticks macroeconomic engines and every company's Company.Tick()
// while logging resulting events.
func (s *Simulation) advanceFundamentals() {
	// 1. Advance Macroeconomic Ecosystem (Citizen & Country / Federal Reserve)
	if s.Citizen != nil && s.National != nil {
		var totalHeadcount float64
		var totalCapEx float64
		var totalCorporateTaxes float64
		var sumReturns float64

		for _, symbol := range s.order {
			co := s.markets[symbol].co
			totalHeadcount += co.Headcount
			totalCapEx += co.CapEx
			totalCorporateTaxes += co.CorporateTaxPaid
			if co.SharesOutstanding > 0 {
				sumReturns += (co.ReportedValue - co.TrueValue) / co.TrueValue
			}
		}

		// Scale sample universe to national macro employment scale
		scaledHeadcount := totalHeadcount
		if totalHeadcount > 0 {
			scaledHeadcount = totalHeadcount * (161_700_000.0 / 25_000_000.0)
		}

		marketReturn := 0.0
		if len(s.order) > 0 {
			marketReturn = sumReturns / float64(len(s.order))
		}

		s.CitizenReport = s.Citizen.Tick(
			s.National.Labor.EmployedWorkers,
			s.National.Labor.AverageHourlyWage,
			s.National.CentralBank.CPIInflationRate,
			s.National.Labor.AnnualWageGrowth,
			marketReturn,
			1.0/252.0,
		)

		s.NationalReport = s.National.Tick(
			s.CitizenReport.ConsumerSpending,
			totalCapEx,
			scaledHeadcount,
			totalCorporateTaxes,
			s.CitizenReport.TaxesPaidThisTick,
			1.0/252.0,
		)

		// If FOMC rate announcement occurred, push event to EventLog
		if s.NationalReport.FOMCAnnouncement != "" {
			s.EventLog = append(s.EventLog, Event{
				Tick:          s.tick,
				Kind:          EventMacro,
				Symbol:        "FED",
				MacroHeadline: s.NationalReport.FOMCAnnouncement,
			})
		}
	}

	for _, symbol := range s.order {
		cm := s.markets[symbol]

		if s.National != nil {
			sectorStr := cm.co.Sector.String()
			compositeDemand := s.National.CompositeSectorDemand(
				sectorStr,
				s.CitizenReport.SectorDemand[sectorStr],
			)
			cm.co.UpdateMacro(
				s.National.Labor.AverageHourlyWage,
				s.National.Fiscal.CorporateTaxRate,
				s.National.CentralBank.TenYearYield,
				compositeDemand,
				1.0/252.0,
			)
		}

		fundamental, restatement := cm.co.Tick()

		if fundamental.Kind != company.NoJump {
			s.EventLog = append(s.EventLog, Event{
				Tick:            s.tick,
				Kind:            EventFundamental,
				Symbol:          symbol,
				FundamentalKind: fundamental.Kind,
				Multiplier:      fundamental.Multiplier,
			})
		}
		if restatement != nil {
			s.EventLog = append(s.EventLog, Event{
				Tick:               s.tick,
				Kind:               EventRestatement,
				Symbol:             symbol,
				RestatementProfile: restatement.Profile,
				PriorGapPercent:    restatement.PriorGapPercent,
				Severity:           restatement.Severity,
			})
		}

		// Corporate distress and workout mechanisms:
		// When a company faces severe distress, realistic financial interventions occur:
		if cm.co.ReportedValue < 1.00 {
			// 1. Reverse Stock Split: 1-for-10 ratio to prevent penny-stock delisting
			oldVal := cm.co.ReportedValue
			cm.co.TrueValue *= 10
			cm.co.ReportedValue *= 10
			cm.co.SharesOutstanding /= 10
			cm.co.Float /= 10
			s.EventLog = append(s.EventLog, Event{
				Tick:            s.tick,
				Kind:            EventReverseSplit,
				Symbol:          symbol,
				DistressDetails: fmt.Sprintf("1-for-10 reverse split executed to maintain listing standards ($%.2f -> $%.2f)", oldVal, cm.co.ReportedValue),
			})
		} else if cm.co.TrueValue < 5.00 && s.rng != nil {
			// Distress resolution chance per tick for distressed companies
			if s.rng.Float64() < 0.02 {
				workoutRoll := s.rng.Float64()
				switch {
				case workoutRoll < 0.30:
					// Emergency Bailout / Syndicate Credit Facility (+50% value recovery)
					mult := 1.40 + s.rng.Float64()*0.35
					cm.co.TrueValue *= mult
					cm.co.ReportedValue *= mult
					facility := 250_000_000 + s.rng.Float64()*500_000_000
					s.EventLog = append(s.EventLog, Event{
						Tick:            s.tick,
						Kind:            EventBailout,
						Symbol:          symbol,
						DistressDetails: fmt.Sprintf("Secured $%.0fM emergency credit line (+%.0f%% value recovery)", facility/1e6, (mult-1)*100),
					})

				case workoutRoll < 0.60:
					// Strategic Acquisition / Buyout Offer (+60% premium)
					mult := 1.50 + s.rng.Float64()*0.30
					cm.co.TrueValue *= mult
					cm.co.ReportedValue *= mult
					s.EventLog = append(s.EventLog, Event{
						Tick:            s.tick,
						Kind:            EventAcquisition,
						Symbol:          symbol,
						DistressDetails: fmt.Sprintf("Private equity syndicate submitted buyout offer @ $%.2f (+%.0f%% premium)", cm.co.ReportedValue, (mult-1)*100),
					})

				case workoutRoll < 0.85:
					// Operational Restructuring & Turnaround (+35% margin & valuation boost)
					mult := 1.30 + s.rng.Float64()*0.25
					cm.co.TrueValue *= mult
					cm.co.ReportedValue *= mult
					cm.co.NetMargin *= 1.25
					s.EventLog = append(s.EventLog, Event{
						Tick:            s.tick,
						Kind:            EventRestructuring,
						Symbol:          symbol,
						DistressDetails: fmt.Sprintf("Turnaround plan announced: closed non-core units, improved margins (+%.0f%% valuation)", (mult-1)*100),
					})

				default:
					// Chapter 11 Reorganization (+45% clean balance sheet rebound)
					mult := 1.35 + s.rng.Float64()*0.30
					cm.co.TrueValue *= mult
					cm.co.ReportedValue *= mult
					s.EventLog = append(s.EventLog, Event{
						Tick:            s.tick,
						Kind:            EventChapter11,
						Symbol:          symbol,
						DistressDetails: fmt.Sprintf("Court approved Chapter 11 plan: debt cleared, emerging with clean balance sheet"),
					})
				}
			}
		}
	}
}

// runTradingRound builds each company's MarketState, polls every
// registered participant for that company, and submits/settles every
// resulting order. Each company's full participant poll happens
// before any of that company's orders are submitted, so every
// participant for a given company sees the same MarketState this
// tick — but different companies are processed sequentially in a
// fixed order (s.order), meaning a participant covering multiple
// companies effectively sees them "settle" in that same fixed order
// within one tick. This is a real, worth-naming simplification: a
// true simultaneous-clearing market would resolve all companies'
// trading in parallel from one shared instant, whereas this
// orchestrator processes company by company. For a research tool this
// is a reasonable trade-off for a tractable, deterministic
// implementation; a fully simultaneous-clearing version is a possible
// future refinement, not something quietly claimed as already true
// here.
func (s *Simulation) runTradingRound(trades *[]Trade) {
	for _, symbol := range s.order {
		cm := s.markets[symbol]
		state := cm.state(s.tick, s.depthLevels)

		var allOrders []*market.Order
		for _, p := range cm.participants {
			orders := p.NextOrders(state)
			if len(orders) > 0 {
				// If participant submitted resting limit orders (e.g. MarketMaker refresh),
				// cancel their previous unexecuted orders so the book doesn't accumulate infinite stale liquidity.
				hasLimit := false
				for _, o := range orders {
					if !o.IsMarket {
						hasLimit = true
						break
					}
				}
				if hasLimit {
					cm.book.CancelAgentOrders(p.ID())
				}
				allOrders = append(allOrders, orders...)
			}
		}

		for _, o := range allOrders {
			fills := cm.book.Submit(o)
			s.settleFills(symbol, o, fills, trades)
		}
	}
}

// settleFills applies every fill from one submitted order to both the
// taker's and maker's accounts. The taker is whichever participant
// submitted o; the maker is whoever's resting order was matched
// against — settleFills looks up both by AgentID across every
// registered holder, since a fill's maker could be any participant
// registered for this symbol, not just the taker.
func (s *Simulation) settleFills(symbol string, o *market.Order, fills []market.Fill, trades *[]Trade) {
	sideStr := "BUY"
	if o.Side == market.Sell {
		sideStr = "SELL"
	}

	for _, f := range fills {
		if trades != nil {
			*trades = append(*trades, Trade{
				Tick:         s.tick,
				Symbol:       symbol,
				TakerAgentID: f.TakerAgentID,
				MakerAgentID: f.MakerAgentID,
				Side:         sideStr,
				Price:        f.Price,
				Quantity:     f.Quantity,
			})
		}

		if taker := s.findAccount(f.TakerAgentID); taker != nil {
			market.ApplySettledFill(taker, symbol, o.Side, f.Quantity, f.Price)
		}
		if maker := s.findAccount(f.MakerAgentID); maker != nil {
			makerSide := market.Sell
			if o.Side == market.Sell {
				makerSide = market.Buy // the maker's fill is always the opposite side of the taker's order
			}
			market.ApplySettledFill(maker, symbol, makerSide, f.Quantity, f.Price)
		}
	}
}

// findAccount looks up a registered participant's account by AgentID.
// Linear scan is acceptable at the participant counts this
// orchestrator is designed for (small slices first, per the project's
// own scaling plan); an id->holder map is a straightforward
// optimization if profiling ever shows this matters at full 500-
// company scale with large retail pools.
func (s *Simulation) findAccount(agentID string) *market.Account {
	for _, h := range s.holders {
		if h.ID() == agentID {
			return h.Account()
		}
	}
	return nil // e.g. a "seed" liquidity order with no real backing account, as used throughout this project's own tests
}

// refreshBankExternalPrices builds a global symbol->mid-price map from
// every company's current book state and hands it to every registered
// agent.Bank via SetExternalPrices, resolving the ReportedValue-based
// fallback approximation Bank previously had to use for any company
// outside the current tick's MarketState.
func (s *Simulation) refreshBankExternalPrices() {
	if len(s.banks) == 0 {
		return
	}
	prices := s.currentMidPrices()
	for _, b := range s.banks {
		b.SetExternalPrices(prices)
	}
}

// currentMidPrices builds a symbol->mid-price map covering every
// registered company, used both for refreshing Bank's external prices
// and for the global liquidation scan. Companies whose book currently
// has no valid mid-price (e.g. one side fully exhausted) are simply
// omitted — downstream consumers (Account.Equity, Account.MarginRatio)
// already handle a symbol being absent from a prices map by treating
// it as not contributing to equity/exposure for that symbol, matching
// how a real broker can't mark a position to a nonexistent price
// either.
func (s *Simulation) currentMidPrices() map[string]float64 {
	prices := make(map[string]float64, len(s.order))
	for _, symbol := range s.order {
		if mid, ok := s.markets[symbol].book.MidPrice(); ok {
			prices[symbol] = mid
		}
	}
	return prices
}

// runLiquidationCascade runs the global liquidation scan, submitting
// any forced orders it produces and settling their fills, looping up
// to maxLiquidationPasses times so a cascade can resolve within this
// tick (a forced sell can itself trigger the next liquidation) — see
// maxLiquidationPasses' doc for the reasoning and its cap.
func (s *Simulation) runLiquidationCascade(trades *[]Trade) {
	engine := market.LiquidationEngine{}

	for range maxLiquidationPasses {
		accounts := s.allAccounts()
		prices := s.currentMidPrices()

		forced := engine.ScanForLiquidations(accounts, prices)
		if len(forced) == 0 {
			return // cascade has resolved, nothing more to liquidate this tick
		}

		for _, fo := range forced {
			cm, ok := s.markets[fo.Symbol]
			if !ok {
				continue // defensive: shouldn't happen if every held symbol was registered via AddCompany
			}
			order := fo.Order
			fills := cm.book.Submit(&order)
			s.settleFills(fo.Symbol, &order, fills, trades)

			s.EventLog = append(s.EventLog, Event{
				Tick:              s.tick,
				Kind:              EventLiquidation,
				Symbol:            fo.Symbol,
				LiquidatedAgentID: fo.Account.AgentID,
				LiquidationQty:    order.Quantity,
			})
		}
	}
}

// allAccounts collects every registered participant's account,
// deduplicating by pointer identity — several participants sharing
// one market.Account (e.g. multiple HedgeFund instances against one
// desk's capital, as HedgeFund's own documentation notes is a
// supported setup) should only be liquidation-scanned once, not once
// per participant wrapping that account.
func (s *Simulation) allAccounts() []*market.Account {
	seen := make(map[*market.Account]bool)
	var out []*market.Account
	for _, h := range s.holders {
		acct := h.Account()
		if acct == nil || seen[acct] {
			continue
		}
		seen[acct] = true
		out = append(out, acct)
	}
	return out
}

// recordPrices appends this tick's settled mid-price into every
// company's rolling history, for next tick's MarketState.
func (s *Simulation) recordPrices() {
	for _, symbol := range s.order {
		s.markets[symbol].recordPrice()
	}
}
