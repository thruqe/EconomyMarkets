package retail

import (
	"math/rand"

	"economy/company"
	"economy/market"
	"economy/technicals"
)

// stopLoss represents one resting protective order a bot has placed
// against its own position in one specific company. Checked
// unconditionally every tick before any discretionary logic runs —
// this mechanical, non-negotiable check is what makes stop-losses the
// mechanism that produces genuine liquidity sweeps: enough bots
// independently placing stops at similar, predictable levels (see
// StopPlacementStyle) creates a cluster of resting forced-exit
// triggers that, once price reaches them, fire together and
// accelerate the move exactly as a real sweep does.
type stopLoss struct {
	active    bool
	triggerAt float64 // price level, above (for a short) or below (for a long) entry
	isLong    bool
}

// watchState holds everything specific to one company on a bot's
// watchlist: its own technicals aggregator (momentum on company A
// tells you nothing about company B, so each watched company needs
// independent bar history), its own cooldown counter, and its own
// stop-loss. A bot with several companies on its watchlist is not
// "one bot split several ways" — it genuinely tracks each company
// independently, exactly as a real retail trader watching a handful
// of names does.
type watchState struct {
	agg               *technicals.Aggregator
	cooldownRemaining int
	stop              stopLoss
}

// SimulatedRetailTrader is a behavioral, non-conviction-driven market
// participant: it does not compute a fair-value estimate the way
// HedgeFund/Bank do (even its FundamentalOnly variant reads
// company.ReportedValue, never TrueValue — see package company's
// documentation on why). Its size, information source, reaction
// shape, and stop discipline are all governed by its SkillTier,
// InformationStyle, and Archetype, generated at construction rather
// than hand-tuned per instance.
//
// A single bot trades a small watchlist of companies (1-5, typically),
// not the full universe — matching how real retail traders follow a
// personal handful of names rather than the whole market. See
// crowd.go's GenerateWatchlistPool for how watchlists are assigned
// with realistic, attention-skewed concentration.
type SimulatedRetailTrader struct {
	id      string
	account *market.Account

	watchlist map[string]*company.Company // symbol -> company, this bot's covered names
	state     map[string]*watchState      // symbol -> this bot's per-company tracking state

	tier      SkillTier
	style     InformationStyle
	archetype Archetype

	rng *rand.Rand
}

// NewSimulatedRetailTrader constructs one bot watching the given
// companies. Exported mainly so GenerateWatchlistPool (see crowd.go)
// has something to call repeatedly; direct construction is also fine
// for tests wanting precise control over a single bot's watchlist,
// tier, style, and archetype.
func NewSimulatedRetailTrader(
	id string,
	acct *market.Account,
	watchlist []*company.Company,
	tier SkillTier,
	style InformationStyle,
	archetype Archetype,
	rng *rand.Rand,
) *SimulatedRetailTrader {
	tp := tierProfiles[tier]

	companies := make(map[string]*company.Company, len(watchlist))
	states := make(map[string]*watchState, len(watchlist))
	for _, c := range watchlist {
		companies[c.Symbol] = c
		states[c.Symbol] = &watchState{agg: technicals.NewAggregator(tp.AggregatorWindow, 200)}
	}

	return &SimulatedRetailTrader{
		id:        id,
		account:   acct,
		watchlist: companies,
		state:     states,
		tier:      tier,
		style:     style,
		archetype: archetype,
		rng:       rng,
	}
}

func (b *SimulatedRetailTrader) ID() string               { return b.id }
func (b *SimulatedRetailTrader) Account() *market.Account { return b.account }

// Watches reports whether this bot's watchlist includes the given
// symbol — exposed mainly for tests and orchestrator bookkeeping
// (e.g. deciding which bots to poll for which company without calling
// NextOrders speculatively on every bot for every symbol).
func (b *SimulatedRetailTrader) Watches(symbol string) bool {
	_, ok := b.watchlist[symbol]
	return ok
}

// netInventory reads this bot's own signed position size for the
// given symbol — a small private equivalent of agent.NetInventory,
// deliberately not shared across the package boundary (see package
// doc: retail imports nothing from agent, and vice versa).
func (b *SimulatedRetailTrader) netInventory(symbol string) float64 {
	pos, ok := b.account.Positions[symbol]
	if !ok {
		return 0
	}
	if pos.Side == market.Short {
		return -pos.Quantity
	}
	return pos.Quantity
}

func absF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// checkStopLoss returns a forced market order if this bot holds a
// position in symbol and price has crossed that company's resting
// stop, unconditionally — no archetype/style discretion applies here,
// matching how a real stop-loss is a mechanical order, not a judgment
// call made fresh each time it's tested.
func (b *SimulatedRetailTrader) checkStopLoss(symbol string, currentPrice float64) *market.Order {
	ws := b.state[symbol]
	if ws == nil || !ws.stop.active {
		return nil
	}

	triggered := false
	if ws.stop.isLong && currentPrice <= ws.stop.triggerAt {
		triggered = true
	}
	if !ws.stop.isLong && currentPrice >= ws.stop.triggerAt {
		triggered = true
	}
	if !triggered {
		return nil
	}

	qty := absF(b.netInventory(symbol))
	if qty <= 0 {
		ws.stop.active = false
		return nil
	}

	side := market.Sell
	if !ws.stop.isLong {
		side = market.Buy // covering a short
	}

	ws.stop.active = false
	tp := tierProfiles[b.tier]
	ws.cooldownRemaining = tp.CooldownTicksMin + b.rng.Intn(tp.CooldownTicksMax-tp.CooldownTicksMin+1)

	return &market.Order{AgentID: b.id, Side: side, Quantity: qty, IsMarket: true}
}

// placeStop sets a fresh stop-loss for symbol after a new position is
// opened, at a level determined by this bot's tier-driven
// StopPlacementStyle.
func (b *SimulatedRetailTrader) placeStop(ws *watchState, entryPrice float64, isLong bool, bars []technicals.Bar) {
	tp := tierProfiles[b.tier]

	var distance float64
	switch tp.StopPlacement {
	case RoundNumberStop:
		// Nearest round number below (long) or above (short) entry —
		// the most predictable, most sweepable placement, matching
		// how real inexperienced retail commonly places stops.
		roundTo := roundNumberIncrement(entryPrice)
		if isLong {
			distance = entryPrice - (roundDown(entryPrice, roundTo))
		} else {
			distance = roundUp(entryPrice, roundTo) - entryPrice
		}
		if distance <= 0 {
			distance = entryPrice * 0.02 // fallback if already sitting on a round number
		}
	case SwingLevelStop:
		high, low, ok := technicals.SwingHighLow(bars, min(len(bars), 10))
		if ok && isLong && low < entryPrice {
			distance = entryPrice - low
		} else if ok && !isLong && high > entryPrice {
			distance = high - entryPrice
		} else {
			distance = entryPrice * 0.03
		}
	case BufferedStop:
		atr, ok := technicals.ATR(bars, min(len(bars)-1, 14))
		if ok && atr > 0 {
			distance = atr * 2.5 // wider, ATR-scaled buffer — less predictable, less sweepable
		} else {
			distance = entryPrice * 0.05
		}
	}

	if isLong {
		ws.stop = stopLoss{active: true, triggerAt: entryPrice - distance, isLong: true}
	} else {
		ws.stop = stopLoss{active: true, triggerAt: entryPrice + distance, isLong: false}
	}
}

func roundNumberIncrement(price float64) float64 {
	switch {
	case price < 10:
		return 0.5
	case price < 100:
		return 5
	case price < 1000:
		return 50
	default:
		return 500
	}
}

func roundDown(price, increment float64) float64 {
	return float64(int(price/increment)) * increment
}

func roundUp(price, increment float64) float64 {
	down := roundDown(price, increment)
	if down == price {
		return price
	}
	return down + increment
}

// computeSignal builds this tick's Signal for one watched company
// from that company's own technicals window and (if this bot's
// style/tier allow) its fundamentals — see InformationStyle doc for
// what each style actually reads.
func (b *SimulatedRetailTrader) computeSignal(symbol string, target *company.Company, ws *watchState, state market.MarketState) (Signal, bool) {
	bars := ws.agg.Bars()
	tp := tierProfiles[b.tier]

	var s Signal
	haveTechnical := false

	if momentum, ok := technicals.Momentum(bars, tp.MomentumWindow); ok {
		s.Momentum = momentum
		haveTechnical = true
	}

	if bb, ok := technicals.BollingerBands(bars, min(len(bars), 20), 2); ok && bb.Middle > 0 {
		s.Volatility = (bb.Upper - bb.Lower) / bb.Middle
	}

	haveFundamental := false
	if tp.ReactsToFundamentalEvents && state.HasMid && state.Mid > 0 {
		gap := (target.ReportedValue - state.Mid) / state.Mid
		s.FundamentalGap = gap
		haveFundamental = true
	}

	switch b.style {
	case TechnicalOnly:
		return s, haveTechnical
	case FundamentalOnly:
		if !haveFundamental {
			return s, false
		}
		// Repurpose Momentum as the drive signal for archetype
		// reaction functions, which are written in terms of Momentum/
		// Volatility — a fundamental gap plays the same directional
		// role technical momentum would for a TechnicalOnly bot.
		s.Momentum = s.FundamentalGap
		return s, true
	case Blended:
		if !haveTechnical && !haveFundamental {
			return s, false
		}
		if haveTechnical && haveFundamental {
			s.Momentum = (s.Momentum + s.FundamentalGap) / 2
		} else if haveFundamental {
			s.Momentum = s.FundamentalGap
		}
		return s, true
	case Confused:
		// Inconsistent by design: each tick, randomly pick which
		// signal (if any) actually drives this bot's reaction,
		// modeling a genuinely unstable process rather than a
		// diluted blend of both.
		roll := b.rng.Float64()
		switch {
		case roll < 0.4 && haveTechnical:
			return s, true
		case roll < 0.7 && haveFundamental:
			s.Momentum = s.FundamentalGap
			return s, true
		default:
			return s, false // this tick, nothing coherent drives it
		}
	}
	return s, false
}

// NextOrders implements market.OrderSource. Called once per tick per
// watched company by the orchestrator (mirroring how agent.Bank is
// driven — see Bank.NextOrders for the identical pattern). If
// state.Symbol isn't on this bot's watchlist, it has nothing to do
// this call, matching how a real retail trader simply doesn't react
// to a company they don't follow.
func (b *SimulatedRetailTrader) NextOrders(state market.MarketState) []*market.Order {
	target, watched := b.watchlist[state.Symbol]
	if !watched {
		return nil
	}
	ws := b.state[state.Symbol]

	if !state.HasMid || state.Mid <= 0 {
		return nil
	}
	ws.agg.AddTick(state.Mid)

	// Stop-loss check is unconditional and comes before cooldown or
	// any discretionary logic — a real stop doesn't wait for the
	// trader to feel like checking it.
	if order := b.checkStopLoss(state.Symbol, state.Mid); order != nil {
		return []*market.Order{order}
	}

	if ws.cooldownRemaining > 0 {
		ws.cooldownRemaining--
		return nil
	}

	signal, ok := b.computeSignal(state.Symbol, target, ws, state)
	if !ok {
		return nil // not enough bar history yet, or (Confused) nothing driving this tick
	}

	tp := tierProfiles[b.tier]
	baseActivity := 0.15 // overall per-tick willingness to act at all, before archetype shaping
	probs := ReactionFor(b.archetype)(signal, baseActivity)

	roll := b.rng.Float64()
	var side market.Side
	switch {
	case roll < probs.Buy:
		side = market.Buy
	case roll < probs.Buy+probs.Sell:
		side = market.Sell
	default:
		return nil // hold
	}

	sizeFraction := tp.PositionSizeFractionMin + b.rng.Float64()*(tp.PositionSizeFractionMax-tp.PositionSizeFractionMin)
	buyingPower := b.account.AvailableBuyingPower(map[string]float64{state.Symbol: state.Mid})
	notional := buyingPower * sizeFraction
	if notional <= 0 {
		return nil
	}
	quantity := notional / state.Mid

	b.placeStop(ws, state.Mid, side == market.Buy, ws.agg.Bars())
	ws.cooldownRemaining = tp.CooldownTicksMin + b.rng.Intn(tp.CooldownTicksMax-tp.CooldownTicksMin+1)

	return []*market.Order{{AgentID: b.id, Side: side, Quantity: quantity, IsMarket: true}}
}
