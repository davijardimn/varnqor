package analytics

import (
	"math"
	"sort"
)

type BurnRate struct {
	Ready           bool    `json:"ready"`
	CurrentPerHour  int64   `json:"currentPerHour"`
	PreviousPerHour int64   `json:"previousPerHour"`
	BaselinePerHour int64   `json:"baselinePerHour"`
	Ratio           float64 `json:"ratio"`
	ChangePercent   float64 `json:"changePercent"`
	Trend           string  `json:"trend"`
	Level           string  `json:"level"`
	Confidence      string  `json:"confidence"`
	BaselineSamples int     `json:"baselineSamples"`
}

// ComputeBurnRate compares the latest rolling 60-minute window with a robust
// baseline built from completed active hours. Zero-usage hours are intentionally
// excluded from the baseline so idle time does not make normal work look like a
// spike. The median reduces the influence of a single unusually expensive hour.
func ComputeBurnRate(current, previous int64, baselineSamples []int64) BurnRate {
	result := BurnRate{
		CurrentPerHour:  max64(current, 0),
		PreviousPerHour: max64(previous, 0),
		Trend:           "idle",
		Level:           "learning",
		Confidence:      "insufficient",
	}

	samples := make([]int64, 0, len(baselineSamples))
	for _, value := range baselineSamples {
		if value > 0 {
			samples = append(samples, value)
		}
	}
	result.BaselineSamples = len(samples)
	if len(samples) > 0 {
		result.BaselinePerHour = median(samples)
	}

	result.Trend, result.ChangePercent = trend(result.CurrentPerHour, result.PreviousPerHour)
	result.Confidence = confidence(len(samples))
	result.Ready = len(samples) >= 2 && result.BaselinePerHour > 0
	if !result.Ready {
		return result
	}

	result.Ratio = float64(result.CurrentPerHour) / float64(result.BaselinePerHour)
	result.Ratio = math.Round(result.Ratio*100) / 100
	result.Level = level(result.Ratio)
	return result
}

func median(values []int64) int64 {
	ordered := append([]int64(nil), values...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	middle := len(ordered) / 2
	if len(ordered)%2 == 1 {
		return ordered[middle]
	}
	return (ordered[middle-1] + ordered[middle]) / 2
}

func trend(current, previous int64) (string, float64) {
	if previous <= 0 {
		if current > 0 {
			return "new-spike", 0
		}
		return "idle", 0
	}

	change := (float64(current-previous) / float64(previous)) * 100
	change = math.Round(change*10) / 10
	switch {
	case change > 10:
		return "accelerating", change
	case change < -10:
		return "cooling", change
	default:
		return "steady", change
	}
}

func confidence(samples int) string {
	switch {
	case samples >= 6:
		return "high"
	case samples >= 3:
		return "medium"
	case samples >= 2:
		return "low"
	default:
		return "insufficient"
	}
}

func level(ratio float64) string {
	switch {
	case ratio < 0.75:
		return "low"
	case ratio < 1.20:
		return "normal"
	case ratio < 1.50:
		return "elevated"
	case ratio < 2.00:
		return "high"
	default:
		return "critical"
	}
}

func max64(value, minimum int64) int64 {
	if value < minimum {
		return minimum
	}
	return value
}
