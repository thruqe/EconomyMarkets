package retail

import (
	"math/rand"

	"economy/company"
	"economy/market"
	"economy/technicals"
)

// stopLoss represents one resting protective order a bot has placed
// against its own position. Checked unconditionally every tick before
// any discretionary logic runs — this mechanical, non-negotiable
// check is what makes stop-losses the mechanism that produces genuine
// liquidity sweeps: enough bots independently placing stops at
// similar, predictable levels (see StopPlacementStyle) creates a
// cluster of resting forced-exit triggers that, once price reaches
// them, fire together and accelerate the move exactly as a real
// sweep does.
type stopLoss struct {
	active    bool
	triggerAt float64 // price level, above (for a short) or below (for a long) entry
	isLong    bool
}

// SimulatedRetailTrader is a behavioral, non-conviction-driven market
// participant: it does not compute a fair-value estimate the way
// HedgeFund/Bank do (even its FundamentalOnly variant reads
// company.ReportedValue, never TrueValue — see package company's
// documentation on why). Its size, information source, reaction
// shape, and stop discipline are all governed by its SkillTier,
// InformationStyle, and Archetype, generated at construction rather
// than hand-tuned per instance.
type SimulatedRetailTrader struct {
	id      string
	account *market.Account
	target  *company.Company // the single company this bot trades

	tier      SkillTier
	style     InformationStyle
	archetype Archetype

	agg *technicals.Aggregator

	cooldownRemaining int
	stop              stopLoss

	rng *rand.Rand
}

// NewSimulatedRetailTrader constructs one bot. Exported mainly so
// GenerateRetailCrowd (see crowd.go) has something to call
// repeatedly; direct construction is also fine for tests wanting
// precise control over a single bot's tier/style/archetype.
func NewSimulatedRetailTrader(
	id string,
	acct *market.Account,
	target *company.Company,
	tier SkillTier,
	style InformationStyle,
	archetype Archetype,
	rng *rand.Rand,
) *SimulatedRetailTrader {
	tp := tierProfiles[tier]
	return &SimulatedRetailTrader{
		id:        id,
		account:   acct,
		target:    target,
		tier:      tier,
		style:     style,
		archetype: archetype,
		agg:       technicals.NewAggregator(tp.AggregatorWindow, 200),
		rng:       rng,
	}
}

func (b *SimulatedRetailTrader) ID() string               { return b.id }
func (b *SimulatedRetailTrader) Account() *market.Account { return b.account }

// netInventory reads this bot's own signed position size for its
// target company — a small private equivalent of agent.NetInventory,
// deliberately not shared across the package boundary (see package
// doc: retail imports nothing from agent, and vice versa).
func (b *SimulatedRetailTrader) netInventory() float64 {
	pos, ok := b.account.Positions[b.target.Symbol]
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
// position and price has crossed its resting stop, unconditionally —
// no archetype/style discretion applies here, matching how a real
// stop-loss is a mechanical order, not a judgment call made fresh
// each time it's tested.
func (b *SimulatedRetailTrader) checkStopLoss(currentPrice float64) *market.Order {
	if !b.stop.active {
		return nil
	}

	triggered := false
	if b.stop.isLong && currentPrice <= b.stop.triggerAt {
		triggered = true
	}
	if !b.stop.isLong && currentPrice >= b.stop.triggerAt {
		triggered = true
	}
	if !triggered {
		return nil
	}

	qty := absF(b.netInventory())
	if qty <= 0 {
		b.stop.active = false
		return nil
	}

	side := market.Sell
	if !b.stop.isLong {
		side = market.Buy // covering a short
	}

	b.stop.active = false
	tp := tierProfiles[b.tier]
	b.cooldownRemaining = tp.CooldownTicksMin + b.rng.Intn(tp.CooldownTicksMax-tp.CooldownTicksMin+1)

	return &market.Order{AgentID: b.id, Side: side, Quantity: qty, IsMarket: true}
}

// placeStop sets a fresh stop-loss after a new position is opened, at
// a level determined by this bot's tier-driven StopPlacementStyle.
func (b *SimulatedRetailTrader) placeStop(entryPrice float64, isLong bool, bars []technicals.Bar) {
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
		b.stop = stopLoss{active: true, triggerAt: entryPrice - distance, isLong: true}
	} else {
		b.stop = stopLoss{active: true, triggerAt: entryPrice + distance, isLong: false}
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

// computeSignal builds this tick's Signal from the bot's own
// technicals window and (if its style/tier allow) fundamentals — see
// InformationStyle doc for what each style actually reads.
func (b *SimulatedRetailTrader) computeSignal(state market.MarketState) (Signal, bool) {
	bars := b.agg.Bars()
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
		gap := (b.target.ReportedValue - state.Mid) / state.Mid
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

// NextOrders implements market.OrderSource.
func (b *SimulatedRetailTrader) NextOrders(state market.MarketState) []*market.Order {
	if !state.HasMid || state.Mid <= 0 {
		return nil
	}
	b.agg.AddTick(state.Mid)

	// Stop-loss check is unconditional and comes before cooldown or
	// any discretionary logic — a real stop doesn't wait for the
	// trader to feel like checking it.
	if order := b.checkStopLoss(state.Mid); order != nil {
		return []*market.Order{order}
	}

	if b.cooldownRemaining > 0 {
		b.cooldownRemaining--
		return nil
	}

	signal, ok := b.computeSignal(state)
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

	b.placeStop(state.Mid, side == market.Buy, b.agg.Bars())
	b.cooldownRemaining = tp.CooldownTicksMin + b.rng.Intn(tp.CooldownTicksMax-tp.CooldownTicksMin+1)

	return []*market.Order{{AgentID: b.id, Side: side, Quantity: quantity, IsMarket: true}}
}
