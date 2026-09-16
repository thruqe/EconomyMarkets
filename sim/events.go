package sim

import "economy/company"

// EventKind categorizes an Event for filtering/display — e.g. a
// research notebook wanting only liquidations, or only fraud
// restatements, across a run.
type EventKind int

const (
	EventFundamental EventKind = iota // a company.FundamentalEvent with Kind != NoJump
	EventRestatement                  // a company.RestatementEvent
	EventLiquidation                  // a forced liquidation order was generated and submitted
	EventIPO                          // an enterprise public listing event
	EventBailout                      // emergency liquidity bailout / syndicate credit line
	EventAcquisition                  // strategic buyout offer at premium
	EventRestructuring                // turnaround cost cutting & operational reorganization
	EventReverseSplit                 // 1-for-10 reverse stock split to maintain listing standards
	EventChapter11                    // court-supervised debt restructuring & clean balance sheet
	EventMacro                        // macroeconomic event (FOMC policy, CPI, labor)
)

func (k EventKind) String() string {
	switch k {
	case EventFundamental:
		return "Fundamental"
	case EventRestatement:
		return "Restatement"
	case EventLiquidation:
		return "Liquidation"
	case EventIPO:
		return "IPO"
	case EventBailout:
		return "Bailout"
	case EventAcquisition:
		return "Acquisition"
	case EventRestructuring:
		return "Restructuring"
	case EventReverseSplit:
		return "ReverseSplit"
	case EventChapter11:
		return "Chapter11"
	case EventMacro:
		return "Macro"
	default:
		return "Unknown"
	}
}

// Event is one notable occurrence during a simulation run, collected
// into Simulation.EventLog. Fields not relevant to a given Kind are
// left at their zero value rather than the log needing a different
// struct per kind — simpler for a first pass at observability, at the
// cost of some unused fields per entry; worth revisiting if the log
// needs to get more structured later.
type Event struct {
	Tick   int
	Kind   EventKind
	Symbol string

	// FundamentalKind/Multiplier populated when Kind == EventFundamental.
	FundamentalKind company.JumpKind
	Multiplier      float64

	// RestatementProfile/PriorGapPercent/Severity populated when
	// Kind == EventRestatement.
	RestatementProfile company.ReportingProfile
	PriorGapPercent    float64
	Severity           float64

	// LiquidatedAgentID/LiquidationQty populated when
	// Kind == EventLiquidation.
	LiquidatedAgentID string
	LiquidationQty    float64

	// IPO details populated when Kind == EventIPO
	IPORevenue float64
	IPOPrice   float64
	IPOShares  float64

	// DistressDetails populated when Kind is a distress event
	DistressDetails string

	// MacroHeadline populated when Kind == EventMacro
	MacroHeadline string
}
