package sim

import (
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
//  2. Build each company's MarketState from its current book and
//     rolling price history.
//  3. Poll every participant registered for each company, collecting
//     every order they return.
//  4. Submit all of this tick's orders to their respective books,
//     settling every resulting fill into both sides' accounts via
//     market.ApplySettledFill.
//  5. Refresh every agent.Bank's cross-company price map
//     (SetExternalPrices) with this tick's real mid-prices, resolving
//     the approximation Bank previously had to fall back to.
//  6. Run the global liquidation scan across every account, looping
//     (up to maxLiquidationPasses) to let cascades resolve within the
//     tick — see maxLiquidationPasses' doc.
//  7. Record each company's new mid-price into rolling history for
//     next tick's MarketState.
func (s *Simulation) Step() {
	s.tick++

	s.advanceFundamentals()
	s.runTradingRound()
	s.refreshBankExternalPrices()
	s.runLiquidationCascade()
	s.recordPrices()
}

// advanceFundamentals ticks every company's Company.Tick() and logs
// any resulting fundamental/restatement events.
func (s *Simulation) advanceFundamentals() {
	for _, symbol := range s.order {
		cm := s.markets[symbol]
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
func (s *Simulation) runTradingRound() {
	for _, symbol := range s.order {
		cm := s.markets[symbol]
		state := cm.state(s.tick, s.depthLevels)

		var allOrders []*market.Order
		for _, p := range cm.participants {
			orders := p.NextOrders(state)
			allOrders = append(allOrders, orders...)
		}

		for _, o := range allOrders {
			fills := cm.book.Submit(o)
			s.settleFills(symbol, o, fills)
		}
	}
}

// settleFills applies every fill from one submitted order to both the
// taker's and maker's accounts. The taker is whichever participant
// submitted o; the maker is whoever's resting order was matched
// against — settleFills looks up both by AgentID across every
// registered holder, since a fill's maker could be any participant
// registered for this symbol, not just the taker.
func (s *Simulation) settleFills(symbol string, o *market.Order, fills []market.Fill) {
	for _, f := range fills {
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
func (s *Simulation) runLiquidationCascade() {
	engine := market.LiquidationEngine{}

	for pass := 0; pass < maxLiquidationPasses; pass++ {
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
			s.settleFills(fo.Symbol, &order, fills)

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
