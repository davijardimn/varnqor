package analytics

import (
	"math"
	"sort"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

// ComputeForecast uses a persistence forecast: the latest rolling 60-minute
// activity rate is projected forward unchanged. The envelope is not a
// statistical confidence interval; it is the observed range formed by the
// current pace, previous pace and active-hour baseline.
func ComputeForecast(burn BurnRate, horizon time.Duration) domain.ForecastReport {
	if horizon <= 0 {
		horizon = 2 * time.Hour
	}
	horizonMinutes := int(math.Round(horizon.Minutes()))
	if horizonMinutes < 1 {
		horizonMinutes = 1
	}

	report := domain.ForecastReport{
		Ready:           burn.CurrentPerHour > 0 || burn.BaselinePerHour > 0,
		Method:          "persistence",
		Pressure:        "learning",
		Confidence:      forecastConfidence(burn),
		CurrentPerHour:  max64(burn.CurrentPerHour, 0),
		PreviousPerHour: max64(burn.PreviousPerHour, 0),
		BaselinePerHour: max64(burn.BaselinePerHour, 0),
		Trend:           burn.Trend,
		HorizonMinutes:  horizonMinutes,
		RunwayAvailable: false,
		LimitSource:     "unavailable",
	}

	if burn.Ready && burn.BaselinePerHour > 0 {
		report.PressureRatio = math.Round((float64(report.CurrentPerHour)/float64(report.BaselinePerHour))*100) / 100
		report.Pressure = level(report.PressureRatio)
	}

	lowRate, highRate := observedRateEnvelope(report.CurrentPerHour, report.PreviousPerHour, report.BaselinePerHour)
	report.ProjectedTokens = projectRate(report.CurrentPerHour, horizonMinutes)
	report.BaselineProjection = projectRate(report.BaselinePerHour, horizonMinutes)
	report.EnvelopeLow = projectRate(lowRate, horizonMinutes)
	report.EnvelopeHigh = projectRate(highRate, horizonMinutes)
	report.Points = forecastPoints(report, lowRate, highRate)
	return report
}

func forecastPoints(report domain.ForecastReport, lowRate, highRate int64) []domain.ForecastPoint {
	minutes := []int{30, 60, report.HorizonMinutes}
	seen := make(map[int]bool)
	unique := make([]int, 0, len(minutes))
	for _, value := range minutes {
		if value <= 0 || value > report.HorizonMinutes || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	if !seen[report.HorizonMinutes] {
		unique = append(unique, report.HorizonMinutes)
	}
	sort.Ints(unique)

	points := make([]domain.ForecastPoint, 0, len(unique))
	for _, minute := range unique {
		points = append(points, domain.ForecastPoint{
			Minutes:         minute,
			ProjectedTokens: projectRate(report.CurrentPerHour, minute),
			BaselineTokens:  projectRate(report.BaselinePerHour, minute),
			EnvelopeLow:     projectRate(lowRate, minute),
			EnvelopeHigh:    projectRate(highRate, minute),
		})
	}
	return points
}

func observedRateEnvelope(current, previous, baseline int64) (int64, int64) {
	rates := []int64{max64(current, 0)}
	if previous > 0 {
		rates = append(rates, previous)
	}
	if baseline > 0 {
		rates = append(rates, baseline)
	}
	low, high := rates[0], rates[0]
	for _, value := range rates[1:] {
		if value < low {
			low = value
		}
		if value > high {
			high = value
		}
	}
	return low, high
}

func projectRate(rate int64, minutes int) int64 {
	if rate <= 0 || minutes <= 0 {
		return 0
	}
	return int64(math.Round(float64(rate) * float64(minutes) / 60.0))
}

func forecastConfidence(burn BurnRate) string {
	if burn.Confidence == "high" || burn.Confidence == "medium" || burn.Confidence == "low" {
		return burn.Confidence
	}
	if burn.CurrentPerHour > 0 {
		return "low"
	}
	return "insufficient"
}
