package device

import "math"

// calcPercent calculates usage percentage rounded to 1 decimal.
func calcPercent(used, total int64) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(used)/float64(total)*1000) / 10
}
