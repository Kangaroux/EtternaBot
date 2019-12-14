package util

import (
	"math"
)

// GetEqualitySign compares the two values and returns the corresponding equality sign
func GetEqualitySign(a, b float64) rune {
	if a > b {
		return '>'
	} else if a < b {
		return '<'
	} else {
		return '='
	}
}

// RoundToPrecision works like math.Round except that it can round to a specific
// decimal point. e.g. RoundToPrecision(0.125, 2) == 0.13
func RoundToPrecision(num float64, precision int) float64 {
	if precision < 0 {
		panic("precision must be greater than zero")
	}

	scalar := math.Pow10(precision)
	f := num * scalar

	// This fixes a floating point issue where a leading 5 turns into 4999... which causes
	// the number to be incorrectly rounded down. 1.255 * 100 => 125.4999...
	// This behavior only appears to happen when the value exists in a variable. If you test
	// it in the Go playground with constants it works as expected
	f = math.Nextafter(f, f+1)

	return math.Round(f) / scalar
}
