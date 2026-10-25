package analytics

import (
	"math"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

func ConfiguredBudgetWindow(now time.Time, resetHour int) (time.Time, time.Time) {
	local := now.In(now.Location())
	if resetHour < 0 {
		resetHour = 0
	}
	if resetHour > 23 {
		resetHour = 23
	}
	start := time.Date(local.Year(), local.Month(), local.Day(), resetHour, 0, 0, 0, local.Location())
	if local.Before(start) {
		start = start.AddDate(0, 0, -1)
	}
	return start, start.AddDate(0, 0, 1)
}

func ComputeConfiguredQuota(now, start, resetAt time.Time, label string, limit, used, currentPerHour int64) domain.QuotaStatus {
	result := domain.QuotaStatus{
		Source:         "configured",
		Metric:         "tokens",
		Label:          label,
		Limit:          max64(limit, 0),
		Used:           max64(used, 0),
		WindowStart:    start,
		ResetAt:        resetAt,
		CurrentPerHour: max64(currentPerHour, 0),
		Confidence:     "configured",
	}
	if result.Label == "" {
		result.Label = "Configured daily budget"
	}
	if result.Limit <= 0 {
		return result
	}
	result.Available = true
	result.Remaining = result.Limit - result.Used
	if result.Remaining < 0 {
		result.Remaining = 0
	}
	result.UtilizationPercent = math.Round((float64(result.Used)/float64(result.Limit))*1000) / 10

	hoursToReset := resetAt.Sub(now).Hours()
	if hoursToReset < 0 {
		hoursToReset = 0
	}
	result.ProjectedAtReset = result.Used + int64(math.Round(float64(result.CurrentPerHour)*hoursToReset))
	result.WillExceedBeforeReset = result.ProjectedAtReset > result.Limit

	if result.Used >= result.Limit {
		result.RunwayAvailable = true
		result.RunwayHours = 0
		return result
	}
	if result.CurrentPerHour > 0 {
		result.RunwayAvailable = true
		result.RunwayHours = math.Round((float64(result.Remaining)/float64(result.CurrentPerHour))*100) / 100
	}
	return result
}
