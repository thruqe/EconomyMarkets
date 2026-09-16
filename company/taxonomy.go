package company

// Sector follows the standard GICS (Global Industry Classification
// Standard) sector taxonomy used by real exchanges and index
// providers, rather than an invented category scheme. Each sector
// carries baseline drift/volatility/jump characteristics loosely
// grounded in how that sector actually tends to behave in real
// markets (e.g. Utilities: low drift, low vol; Energy: high vol from
// commodity exposure; Information Technology: higher drift and vol).
type Sector int

const (
	Energy Sector = iota
	Materials
	Industrials
	ConsumerDiscretionary
	ConsumerStaples
	HealthCare
	Financials
	InformationTechnology
	CommunicationServices
	Utilities
	RealEstate
)

func (s Sector) String() string {
	switch s {
	case Energy:
		return "Energy"
	case Materials:
		return "Materials"
	case Industrials:
		return "Industrials"
	case ConsumerDiscretionary:
		return "Consumer Discretionary"
	case ConsumerStaples:
		return "Consumer Staples"
	case HealthCare:
		return "Health Care"
	case Financials:
		return "Financials"
	case InformationTechnology:
		return "Information Technology"
	case CommunicationServices:
		return "Communication Services"
	case Utilities:
		return "Utilities"
	case RealEstate:
		return "Real Estate"
	default:
		return "Unknown Sector"
	}
}

// AllSectors lists every sector, in a stable order, for generators
// that need to iterate or pick uniformly among them.
var AllSectors = []Sector{
	Energy, Materials, Industrials, ConsumerDiscretionary, ConsumerStaples,
	HealthCare, Financials, InformationTechnology, CommunicationServices,
	Utilities, RealEstate,
}

// CapTier is a market-capitalization tier. Independent of sector, cap
// tier scales volatility on top of the sector baseline — the "size
// effect", one of the more robust empirical patterns in equity
// markets: smaller companies are reliably more volatile than larger
// ones in the same sector, not just noisier versions of the same risk.
type CapTier int

const (
	MegaCap CapTier = iota
	LargeCap
	MidCap
	SmallCap
)

func (c CapTier) String() string {
	switch c {
	case MegaCap:
		return "Mega Cap"
	case LargeCap:
		return "Large Cap"
	case MidCap:
		return "Mid Cap"
	case SmallCap:
		return "Small Cap"
	default:
		return "Unknown Cap Tier"
	}
}

// SectorProfile holds the baseline GBM parameters and jump-risk
// character for a sector, before any cap-tier scaling is applied.
type SectorProfile struct {
	// DriftMin/DriftMax bound the annualized-style drift (μ) range this
	// sector's companies are generated from, per tick in this sim's
	// units. Expressed as a range rather than a fixed number so
	// individual companies within a sector still vary.
	DriftMin, DriftMax float64

	// VolMin/VolMax bound the baseline volatility (σ) range, before
	// CapTier scaling is applied.
	VolMin, VolMax float64

	// JumpRiskMultiplier scales this sector's baseline jump
	// probabilities (see JumpParams) up or down — e.g. Information
	// Technology and Health Care (biotech-style binary outcomes) skew
	// higher; Utilities and Consumer Staples skew lower.
	JumpRiskMultiplier float64

	// HighUncertaintyWeight scales how likely this sector is to
	// generate companies with a HighUncertainty reporting profile,
	// reflecting real-world variation in how knowable a sector's
	// fundamentals genuinely are — early-stage biotech and speculative
	// technology names are honestly harder to value even for
	// insiders, versus a utility with predictable, well-understood
	// cash flows. This is a multiplier on the baseline reporting-
	// profile weight (see ReportingProfileWeights), not an absolute
	// probability.
	HighUncertaintyWeight float64

	// PSMultipleMin/Max bound typical Price-to-Sales valuation multiples
	// for this sector in equilibrium.
	PSMultipleMin, PSMultipleMax float64

	// NetMarginMin/Max bound typical net profit margins for this sector.
	NetMarginMin, NetMarginMax float64
}

// sectorProfiles gives each sector a defensible baseline, directionally
// consistent with how these sectors actually tend to behave (Utilities
// low drift/vol, Energy high vol from commodity exposure, Information
// Technology higher drift/vol, etc.). These are reasonable defaults
// for simulation purposes, not a claim of precise empirical
// calibration — tune them once you're observing simulated behavior
// and forming opinions about feel.
var sectorProfiles = map[Sector]SectorProfile{
	Energy:                {DriftMin: 0.00001, DriftMax: 0.00005, VolMin: 0.0008, VolMax: 0.0018, JumpRiskMultiplier: 1.1, HighUncertaintyWeight: 0.9, PSMultipleMin: 1.2, PSMultipleMax: 2.5, NetMarginMin: 0.08, NetMarginMax: 0.18},
	Materials:             {DriftMin: 0.00001, DriftMax: 0.00005, VolMin: 0.0007, VolMax: 0.0016, JumpRiskMultiplier: 1.0, HighUncertaintyWeight: 0.8, PSMultipleMin: 1.2, PSMultipleMax: 2.5, NetMarginMin: 0.07, NetMarginMax: 0.16},
	Industrials:           {DriftMin: 0.00002, DriftMax: 0.00006, VolMin: 0.0007, VolMax: 0.0015, JumpRiskMultiplier: 0.9, HighUncertaintyWeight: 0.6, PSMultipleMin: 1.5, PSMultipleMax: 3.0, NetMarginMin: 0.08, NetMarginMax: 0.15},
	ConsumerDiscretionary: {DriftMin: 0.00002, DriftMax: 0.00006, VolMin: 0.0008, VolMax: 0.0017, JumpRiskMultiplier: 1.1, HighUncertaintyWeight: 0.9, PSMultipleMin: 1.5, PSMultipleMax: 3.5, NetMarginMin: 0.06, NetMarginMax: 0.15},
	ConsumerStaples:       {DriftMin: 0.00002, DriftMax: 0.00005, VolMin: 0.0004, VolMax: 0.0009, JumpRiskMultiplier: 0.6, HighUncertaintyWeight: 0.3, PSMultipleMin: 1.5, PSMultipleMax: 2.8, NetMarginMin: 0.07, NetMarginMax: 0.13},
	HealthCare:            {DriftMin: 0.00002, DriftMax: 0.00007, VolMin: 0.0008, VolMax: 0.0018, JumpRiskMultiplier: 1.2, HighUncertaintyWeight: 1.8, PSMultipleMin: 3.5, PSMultipleMax: 7.5, NetMarginMin: 0.12, NetMarginMax: 0.25},
	Financials:            {DriftMin: 0.00002, DriftMax: 0.00005, VolMin: 0.0006, VolMax: 0.0014, JumpRiskMultiplier: 1.0, HighUncertaintyWeight: 0.6, PSMultipleMin: 2.0, PSMultipleMax: 4.5, NetMarginMin: 0.15, NetMarginMax: 0.28},
	InformationTechnology: {DriftMin: 0.00003, DriftMax: 0.00008, VolMin: 0.0009, VolMax: 0.0020, JumpRiskMultiplier: 1.3, HighUncertaintyWeight: 1.5, PSMultipleMin: 4.5, PSMultipleMax: 9.5, NetMarginMin: 0.18, NetMarginMax: 0.32},
	CommunicationServices: {DriftMin: 0.00002, DriftMax: 0.00006, VolMin: 0.0007, VolMax: 0.0016, JumpRiskMultiplier: 1.1, HighUncertaintyWeight: 0.9, PSMultipleMin: 2.5, PSMultipleMax: 5.0, NetMarginMin: 0.10, NetMarginMax: 0.22},
	Utilities:             {DriftMin: 0.00002, DriftMax: 0.00006, VolMin: 0.0005, VolMax: 0.0010, JumpRiskMultiplier: 0.4, HighUncertaintyWeight: 0.2, PSMultipleMin: 1.5, PSMultipleMax: 2.8, NetMarginMin: 0.08, NetMarginMax: 0.14},
	RealEstate:            {DriftMin: 0.00002, DriftMax: 0.00005, VolMin: 0.0005, VolMax: 0.0012, JumpRiskMultiplier: 0.7, HighUncertaintyWeight: 0.4, PSMultipleMin: 3.0, PSMultipleMax: 6.0, NetMarginMin: 0.15, NetMarginMax: 0.30},
}

// CapTierProfile scales sector-baseline volatility and jump risk by
// company size. Smaller companies are more volatile and more
// jump-prone (less diversified, thinner analyst coverage, more binary
// outcomes) regardless of sector.
type CapTierProfile struct {
	VolMultiplier  float64
	JumpMultiplier float64

	// SharesOutstandingMin/Max bound the share-count range typical of
	// this tier, used by the generator when constructing a company.
	SharesOutstandingMin, SharesOutstandingMax float64

	// FloatFractionMin/Max bound what fraction of shares outstanding
	// are actually tradeable float, as opposed to closely held. Larger
	// companies tend to have a higher tradeable fraction; smaller/
	// newer companies often have more concentrated closely-held stakes.
	FloatFractionMin, FloatFractionMax float64

	// RevenueMin/Max bound the annual revenue in USD for companies in this tier.
	RevenueMin, RevenueMax float64
}

var capTierProfiles = map[CapTier]CapTierProfile{
	MegaCap:  {VolMultiplier: 0.7, JumpMultiplier: 0.5, SharesOutstandingMin: 2_000_000_000, SharesOutstandingMax: 8_000_000_000, FloatFractionMin: 0.80, FloatFractionMax: 0.95, RevenueMin: 25_000_000_000, RevenueMax: 150_000_000_000},
	LargeCap: {VolMultiplier: 0.9, JumpMultiplier: 0.8, SharesOutstandingMin: 300_000_000, SharesOutstandingMax: 2_000_000_000, FloatFractionMin: 0.70, FloatFractionMax: 0.90, RevenueMin: 5_000_000_000, RevenueMax: 25_000_000_000},
	MidCap:   {VolMultiplier: 1.15, JumpMultiplier: 1.2, SharesOutstandingMin: 50_000_000, SharesOutstandingMax: 300_000_000, FloatFractionMin: 0.55, FloatFractionMax: 0.85, RevenueMin: 1_000_000_000, RevenueMax: 5_000_000_000},
	SmallCap: {VolMultiplier: 1.5, JumpMultiplier: 1.8, SharesOutstandingMin: 5_000_000, SharesOutstandingMax: 50_000_000, FloatFractionMin: 0.35, FloatFractionMax: 0.75, RevenueMin: 100_000_000, RevenueMax: 1_000_000_000},
}
