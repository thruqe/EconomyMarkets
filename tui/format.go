package tui

import (
	"fmt"
	"math"
	"strings"
)

// FormatMetric formats any numerical value with human-readable SI metric suffixes (K, M, B, T).
// If currency is true, prepends a '$' sign.
func FormatMetric(val float64, currency bool) string {
	sign := ""
	if val < 0 {
		sign = "-"
		val = math.Abs(val)
	}

	prefix := ""
	if currency {
		prefix = "$"
	}

	switch {
	case val >= 1e12:
		return fmt.Sprintf("%s%s%.2fT", sign, prefix, val/1e12)
	case val >= 1e9:
		return fmt.Sprintf("%s%s%.2fB", sign, prefix, val/1e9)
	case val >= 1e6:
		return fmt.Sprintf("%s%s%.2fM", sign, prefix, val/1e6)
	case val >= 1e3:
		return fmt.Sprintf("%s%s%.2fK", sign, prefix, val/1e3)
	default:
		if currency {
			return fmt.Sprintf("%s$%.2f", sign, val)
		}
		if val == math.Floor(val) {
			return fmt.Sprintf("%s%.0f", sign, val)
		}
		return fmt.Sprintf("%s%.2f", sign, val)
	}
}

// FormatCurrency returns formatted currency with metric suffix when large (e.g. $1.25M)
func FormatCurrency(val float64) string {
	return FormatMetric(val, true)
}

// FormatShares returns formatted share volume or count with metric suffix (e.g. 50.00M, 2.50B)
func FormatShares(val float64) string {
	return FormatMetric(val, false)
}

// FormatIntegerWithCommas returns comma-delimited integers (e.g. 1,250,000)
func FormatIntegerWithCommas(n int64) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return sign + s
	}

	var parts []string
	remainder := len(s) % 3
	if remainder > 0 {
		parts = append(parts, s[:remainder])
	}
	for i := remainder; i < len(s); i += 3 {
		parts = append(parts, s[i:i+3])
	}
	return sign + strings.Join(parts, ",")
}
