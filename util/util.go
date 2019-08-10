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

// TruncateFloat drops any decimals beyond "precision"
func TruncateFloat(num float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	newNum := num * pow
	newNum = math.Floor(newNum)

	return newNum / pow
}
