package company

import "math/rand"

// ReportingProfile categorizes how a company's publicly reported
// value relates to its true fundamental value (TrueValue). This is a
// genuinely distinct axis from Sector/CapTier/jump risk: it answers
// "how much can this company's own disclosures be trusted", not "how
// volatile or risky is this business." A company can be a stable,
// low-volatility Utility and still have a misleading reporting
// profile, or a high-growth, high-jump-risk tech company that
// reports with scrupulous honesty — the two axes are independent.
//
// This exists because, in reality, nobody — not even sophisticated
// retail, and arguably not even institutional analysts — observes
// TrueValue directly. Everyone observes ReportedValue and has to
// judge how much to trust it. A fundamentals-driven trader who reads
// ReportedValue on a company with hidden fraud will look, for a long
// stretch, exactly like a trader who is right — until a restatement
// event violently reveals otherwise. This is a real, well-documented
// market phenomenon (Enron, Wirecard, Luckin Coffee), not a
// simplification.
type ReportingProfile int

const (
	// Honest: ReportedValue tracks TrueValue closely, with only small
	// natural noise (estimation error, timing lag) — the majority
	// case.
	Honest ReportingProfile = iota

	// OptimisticBias: a small, persistent positive bias — routine
	// corporate optimism or aggressive-but-legal accounting choices,
	// not fraud. Very common in reality and not scandalous on its own.
	OptimisticBias

	// ConservativeBias: the mirror case — a small, persistent
	// negative bias, e.g. deliberately sandbagged guidance so the
	// company can reliably "beat" expectations later. Also a real,
	// common, non-scandalous corporate tactic.
	ConservativeBias

	// HighUncertainty: honest reporting, but with wide variance
	// rather than a directional bias — some companies' true value is
	// genuinely hard to know even for insiders (early-stage biotech,
	// speculative technology), independent of any intent to mislead.
	HighUncertainty

	// Fraudulent: a large, hidden, and growing gap between
	// ReportedValue and TrueValue, rare by design, which eventually
	// resolves in a violent restatement event that corrects
	// ReportedValue back toward TrueValue all at once — the
	// Enron/Wirecard case.
	Fraudulent
)

func (r ReportingProfile) String() string {
	switch r {
	case Honest:
		return "Honest"
	case OptimisticBias:
		return "Optimistic Bias"
	case ConservativeBias:
		return "Conservative Bias"
	case HighUncertainty:
		return "High Uncertainty"
	case Fraudulent:
		return "Fraudulent"
	default:
		return "Unknown Reporting Profile"
	}
}

// ReportingParams configures the noise, bias, and correction dynamics
// for one company's reporting profile. Values here are generated
// per-company (see reportingParamsFor), not shared globally, so two
// companies with the same profile still vary in exact magnitude.
type ReportingParams struct {
	Profile ReportingProfile

	// NoiseStdDev is the standard deviation (as a fraction of
	// TrueValue) of ordinary per-tick reporting noise applied on top
	// of any persistent bias — present for every profile, since even
	// honest reporting isn't perfectly precise.
	NoiseStdDev float64

	// PersistentBias is a per-tick fractional drift applied to the
	// gap between ReportedValue and TrueValue: positive for
	// OptimisticBias, negative for ConservativeBias, zero for Honest
	// and HighUncertainty (whose divergence is noise-driven, not
	// directional), and used as the *hidden accumulation rate* for
	// Fraudulent before its eventual correction.
	PersistentBias float64

	// RestatementProbability is the per-tick probability of a
	// correction event firing. Honest/HighUncertainty get a small
	// probability of gentle, non-violent corrections (case: "we now
	// have better information"); Fraudulent gets a much smaller
	// probability of a single large, violent correction; Optimistic/
	// ConservativeBias correct only slowly through ordinary mean
	// reversion of PersistentBias rather than discrete events (kept
	// at 0 here).
	RestatementProbability float64

	// RestatementSeverity bounds how much of the accumulated
	// ReportedValue/TrueValue gap gets closed when a restatement
	// fires, expressed as a fraction of the gap (1.0 = fully closes
	// the gap immediately; Fraudulent uses a high severity for a
	// violent correction, gentle corrections use a partial value).
	RestatementSeverityMin, RestatementSeverityMax float64
}

// reportingProfileWeights gives the baseline (pre-sector-scaling)
// likelihood of each profile when generating a company. Fraudulent is
// deliberately rare — calibrated similarly to ManiaJump's rarity, so
// across a 500-company universe only a small handful of companies
// carry hidden fraud risk, matching how large-scale accounting fraud
// is rare but not vanishingly so in reality.
var reportingProfileWeights = map[ReportingProfile]float64{
	Honest:           0.55,
	OptimisticBias:   0.20,
	ConservativeBias: 0.15,
	HighUncertainty:  0.08, // baseline; scaled further by sector via HighUncertaintyWeight
	Fraudulent:       0.02,
}

// chooseReportingProfile draws a ReportingProfile for a company,
// scaling HighUncertainty's weight by the sector's
// HighUncertaintyWeight so sectors with genuinely harder-to-know
// fundamentals (biotech, speculative tech) produce that profile more
// often, matching real-world variation in how knowable different
// businesses' fundamentals are.
func chooseReportingProfile(sector Sector, rng *rand.Rand) ReportingProfile {
	sp := sectorProfiles[sector]

	weights := map[ReportingProfile]float64{
		Honest:           reportingProfileWeights[Honest],
		OptimisticBias:   reportingProfileWeights[OptimisticBias],
		ConservativeBias: reportingProfileWeights[ConservativeBias],
		HighUncertainty:  reportingProfileWeights[HighUncertainty] * sp.HighUncertaintyWeight,
		Fraudulent:       reportingProfileWeights[Fraudulent],
	}

	var total float64
	for _, w := range weights {
		total += w
	}

	r := rng.Float64() * total
	var cumulative float64
	// Iterate in a fixed order so profile selection is reproducible
	// under a seeded rng (map iteration order is not guaranteed in Go).
	order := []ReportingProfile{Honest, OptimisticBias, ConservativeBias, HighUncertainty, Fraudulent}
	for _, p := range order {
		cumulative += weights[p]
		if r <= cumulative {
			return p
		}
	}
	return Honest
}

// reportingParamsFor builds the concrete, per-company ReportingParams
// for a chosen profile, drawing specific magnitudes from reasonable
// ranges so companies sharing a profile still vary individually.
func reportingParamsFor(profile ReportingProfile, rng *rand.Rand) ReportingParams {
	switch profile {
	case Honest:
		return ReportingParams{
			Profile:                Honest,
			NoiseStdDev:            0.01 + rng.Float64()*0.02, // 1-3%
			PersistentBias:         0,
			RestatementProbability: 1.0 / 4000, // occasional gentle "better information" correction
			RestatementSeverityMin: 0.2,
			RestatementSeverityMax: 0.5,
		}
	case OptimisticBias:
		return ReportingParams{
			Profile:                OptimisticBias,
			NoiseStdDev:            0.015 + rng.Float64()*0.02,
			PersistentBias:         0.00006 + rng.Float64()*0.00014, // compounds to roughly 3-10% over ~500 ticks
			RestatementProbability: 1.0 / 4000,
			RestatementSeverityMin: 0.2,
			RestatementSeverityMax: 0.5,
		}
	case ConservativeBias:
		return ReportingParams{
			Profile:                ConservativeBias,
			NoiseStdDev:            0.015 + rng.Float64()*0.02,
			PersistentBias:         -(0.00006 + rng.Float64()*0.00014), // mirror of OptimisticBias
			RestatementProbability: 1.0 / 4000,
			RestatementSeverityMin: 0.2,
			RestatementSeverityMax: 0.5,
		}
	case HighUncertainty:
		return ReportingParams{
			Profile:                HighUncertainty,
			NoiseStdDev:            0.05 + rng.Float64()*0.08, // 5-13%: wide, honest uncertainty, no directional bias
			PersistentBias:         0,
			RestatementProbability: 1.0 / 2500, // more frequent gentle corrections — genuinely evolving information
			RestatementSeverityMin: 0.3,
			RestatementSeverityMax: 0.7,
		}
	case Fraudulent:
		return ReportingParams{
			Profile:                Fraudulent,
			NoiseStdDev:            0.01 + rng.Float64()*0.02,       // reports look just as clean as an honest company
			PersistentBias:         0.00015 + rng.Float64()*0.00025, // compounds to roughly 15-40% over ~1000 ticks before discovery — meaningfully larger than OptimisticBias, but not absurd
			RestatementProbability: 1.0 / 6000,                      // rare — mirrors ManiaJump-scale rarity
			RestatementSeverityMin: 0.85,
			RestatementSeverityMax: 1.0, // violent, near-total correction when discovered
		}
	default:
		return ReportingParams{Profile: Honest, NoiseStdDev: 0.02}
	}
}
