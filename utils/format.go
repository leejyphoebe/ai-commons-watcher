package utils

import (
	"fmt"
	"math"
)

func FormatToK(value float64, decimalPlace float64, kFormat, belowKFormat string) string {
	if value < 1000 {
		return fmt.Sprintf(belowKFormat, value)
	}
	k := value / 1000
	dec := math.Pow(10, decimalPlace)
	k = math.Floor(k*dec) / dec // Round to 1 decimal place
	return fmt.Sprintf(kFormat, k)
}
