package company

import (
	"math/rand"
	"strings"
)

// nameParts holds sector-flavored word banks used to construct
// plausible company names, so generated names read as at least
// loosely appropriate to their sector (Energy names skew toward
// "Petro"/"Basin"/"Field"-style words, Information Technology skews
// toward "Nex"/"Cyber"/"Logic"-style words, etc.) rather than being
// generic across all 500 companies.
type nameParts struct {
	prefixes []string
	roots    []string
	suffixes []string
}

var sectorNameParts = map[Sector]nameParts{
	Energy: {
		prefixes: []string{"Petro", "Basin", "Ridge", "Delta", "Summit", "Anchor"},
		roots:    []string{"Field", "Well", "Reserve", "Pipeline", "Crest"},
		suffixes: []string{"Energy", "Resources", "Petroleum", "Holdings"},
	},
	Materials: {
		prefixes: []string{"Iron", "Granite", "Copper", "Continental", "Bedrock"},
		roots:    []string{"Mining", "Alloy", "Mineral", "Quarry"},
		suffixes: []string{"Materials", "Industries", "Corp", "Group"},
	},
	Industrials: {
		prefixes: []string{"Union", "Atlas", "Precision", "Ironclad", "Vanguard"},
		roots:    []string{"Machine", "Works", "Fabrication", "Systems"},
		suffixes: []string{"Industries", "Manufacturing", "Corp", "Holdings"},
	},
	ConsumerDiscretionary: {
		prefixes: []string{"Urban", "Bright", "Luxe", "Metro", "Pacific"},
		roots:    []string{"Retail", "Goods", "Style", "Living"},
		suffixes: []string{"Brands", "Group", "Co.", "Holdings"},
	},
	ConsumerStaples: {
		prefixes: []string{"Harvest", "Golden", "Homestead", "Heritage", "Clearwater"},
		roots:    []string{"Foods", "Provisions", "Pantry", "Mills"},
		suffixes: []string{"Foods", "Brands", "Co.", "Group"},
	},
	HealthCare: {
		prefixes: []string{"Vita", "Gene", "Nova", "Helix", "Cortex"},
		roots:    []string{"Health", "Bio", "Therapeutics", "Med"},
		suffixes: []string{"Health", "Sciences", "Therapeutics", "Labs"},
	},
	Financials: {
		prefixes: []string{"Sterling", "Meridian", "Trust", "Capital", "Ledger"},
		roots:    []string{"Financial", "Trust", "Capital", "Credit"},
		suffixes: []string{"Financial", "Group", "Holdings", "Partners"},
	},
	InformationTechnology: {
		prefixes: []string{"Nex", "Cyber", "Byte", "Logic", "Quantum"},
		roots:    []string{"Soft", "Data", "Systems", "Cloud"},
		suffixes: []string{"Technologies", "Systems", "Software", "Inc."},
	},
	CommunicationServices: {
		prefixes: []string{"Signal", "Relay", "Beacon", "Pulse", "Echo"},
		roots:    []string{"Media", "Network", "Stream", "Connect"},
		suffixes: []string{"Media", "Communications", "Networks", "Group"},
	},
	Utilities: {
		prefixes: []string{"Northern", "Coastal", "Public", "Regional", "Continental"},
		roots:    []string{"Power", "Grid", "Water", "Light"},
		suffixes: []string{"Utilities", "Power", "Energy", "Co."},
	},
	RealEstate: {
		prefixes: []string{"Cornerstone", "Skyline", "Landmark", "Harborview", "Prairie"},
		roots:    []string{"Realty", "Properties", "Estates", "Trust"},
		suffixes: []string{"Realty", "Properties", "REIT", "Trust"},
	},
}

// GenerateName builds a plausible, sector-flavored company name from
// word banks rather than a generic placeholder, using rng so name
// generation is reproducible under a seeded generator.
func GenerateName(sector Sector, rng *rand.Rand) string {
	parts := sectorNameParts[sector]
	prefix := parts.prefixes[rng.Intn(len(parts.prefixes))]
	suffix := parts.suffixes[rng.Intn(len(parts.suffixes))]

	// Occasionally include a root word between prefix and suffix for
	// variety (e.g. "Nex Data Systems" vs just "Nex Technologies").
	if rng.Float64() < 0.4 {
		root := parts.roots[rng.Intn(len(parts.roots))]
		return prefix + " " + root + " " + suffix
	}
	return prefix + " " + suffix
}

// GenerateSymbol derives a ticker-style symbol from a generated name,
// mirroring how real tickers are usually shortened forms of the
// company name rather than arbitrary letters. Falls back to padding
// with consonants from the name if the natural initials run short,
// and de-duplicates against usedSymbols by appending a numeral suffix
// on collision.
func GenerateSymbol(name string, usedSymbols map[string]bool) string {
	words := strings.Fields(name)
	var initials []byte
	for _, w := range words {
		if len(w) > 0 {
			initials = append(initials, w[0])
		}
	}

	symbol := strings.ToUpper(string(initials))
	if len(symbol) > 4 {
		symbol = symbol[:4]
	}
	if len(symbol) < 2 && len(words) > 0 {
		// Very short name: pad using leading letters of the first word.
		first := strings.ToUpper(words[0])
		for i := 0; i < len(first) && len(symbol) < 3; i++ {
			symbol += string(first[i])
		}
	}

	base := symbol
	candidate := base
	suffixNum := 1
	for usedSymbols[candidate] {
		suffixNum++
		candidate = base + itoa(suffixNum)
	}
	usedSymbols[candidate] = true
	return candidate
}

// itoa is a tiny local integer-to-string helper to avoid pulling in
// strconv for a single-digit-almost-always case; falls back correctly
// for multi-digit collision counts too.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
