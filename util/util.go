package util

import (
	"math"
)

// TruncateFloat drops any decimals beyond "precision"
func TruncateFloat(num float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	newNum := num * pow
	newNum = math.Floor(newNum)

	return newNum / pow
}
