package analytics

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

// BuildAdvisor converts existing Varnqor signals into prioritized, read-only
// recommendations. It does not mutate agent settings, prompts, files or model
// selection. Recommendations remain suggestions the user can inspect.
func BuildAdvisor(now time.Time, burn BurnRate, forecast domain.ForecastReport, attribution domain.AttributionReport, waste domain.WasteReport) domain.AdvisorReport {
	report := domain.AdvisorReport{GeneratedAt: now, Mode: "advisory", Recommendations: []domain.AdvisorRecommendation{}}

	if pressureRank(forecast.Pressure) >= pressureRank("elevated") {
		priority := "medium"
		if forecast.Pressure == "high" || forecast.Pressure == "critical" {
			priority = "high"
		}
		report.Recommendations = append(report.Recommendations, domain.AdvisorRecommendation{
			ID: "resource-pressure", Priority: priority, Category: "pace", Title: "Reduce near-term resource pressure",
			Evidence: fmt.Sprintf("Current pace is %.2fx the active baseline; persistence projection is %d tokens over %d minutes.", forecast.PressureRatio, forecast.ProjectedTokens, forecast.HorizonMinutes),
			Recommendation: "Consider narrowing the next agent step or deferring nonessential parallel exploration until the burn rate cools.",
		})
	}

	for _, summary := range waste.Rules {
		if summary.Signals <= 0 && summary.Findings <= 0 {
			continue
		}
		recommendation, ok := wasteRecommendation(summary)
		if ok {
			report.Recommendations = append(report.Recommendations, recommendation)
		}
	}

	if pressureRank(forecast.Pressure) >= pressureRank("elevated") {
		for _, agent := range attribution.Agents {
			if (agent.Key == "subagents" || strings.EqualFold(agent.Label, "Subagents")) && agent.SharePercent >= 35 {
				priority := "low"
				if agent.SharePercent >= 50 {
					priority = "medium"
				}
				report.Recommendations = append(report.Recommendations, domain.AdvisorRecommendation{
					ID: "subagent-pressure", Priority: priority, Category: "agents", Title: "Review subagent fan-out",
					Evidence: fmt.Sprintf("Subagents account for %.1f%% of activity in the current burn window.", agent.SharePercent),
					Recommendation: "Check whether every concurrent subagent is still needed and avoid duplicate exploration across agents.",
				})
				break
			}
		}

		if len(attribution.Sessions) > 0 && attribution.Sessions[0].SharePercent >= 60 {
			top := attribution.Sessions[0]
			report.Recommendations = append(report.Recommendations, domain.AdvisorRecommendation{
				ID: "session-concentration", Priority: "low", Category: "attribution", Title: "Inspect the dominant session",
				Evidence: fmt.Sprintf("Session %s accounts for %.1f%% of activity in the current burn window.", shortAdvisorSession(top.Session), top.SharePercent),
				Recommendation: "Inspect that session before broadening the workload; concentration is not waste, but it is the best place to explain the current burn.",
			})
		}
	}

	sort.SliceStable(report.Recommendations, func(i, j int) bool {
		left, right := report.Recommendations[i], report.Recommendations[j]
		if priorityRank(left.Priority) != priorityRank(right.Priority) {
			return priorityRank(left.Priority) > priorityRank(right.Priority)
		}
		if left.EstimatedTokens != right.EstimatedTokens {
			return left.EstimatedTokens > right.EstimatedTokens
		}
		return left.ID < right.ID
	})
	if len(report.Recommendations) > 5 {
		report.Recommendations = report.Recommendations[:5]
	}
	for _, recommendation := range report.Recommendations {
		switch recommendation.Priority {
		case "high":
			report.High++
		case "medium":
			report.Medium++
		default:
			report.Low++
		}
	}
	return report
}

func wasteRecommendation(summary domain.WasteRuleSummary) (domain.AdvisorRecommendation, bool) {
	priority := signalPriority(summary.EstimatedTokens, summary.Signals)
	base := domain.AdvisorRecommendation{
		Priority: priority, Category: "waste", EstimatedTokens: summary.EstimatedTokens, RuleID: summary.RuleID,
		Evidence: fmt.Sprintf("%d signals across %d findings; ~%d heuristic context tokens.", summary.Signals, summary.Findings, summary.EstimatedTokens),
	}
	switch summary.RuleID {
	case "repeated-file-read":
		base.ID, base.Title = "repeated-file-read", "Reuse recent file context when still valid"
		base.Recommendation = "Before another identical Read, prefer the already-returned range unless the file changed or context was compacted."
	case "repeated-search":
		base.ID, base.Title = "repeated-search", "Reuse or narrow repeated searches"
		base.Recommendation = "Reuse recent Grep/Glob results when still valid, or narrow the search scope before repeating the same query."
	case "oversized-tool-response":
		base.ID, base.Title = "oversized-tool-response", "Narrow large tool responses"
		base.Recommendation = "Use smaller Read ranges or tighter search scopes when the full result is not necessary for the next step."
	case "repeated-failed-operation":
		base.ID, base.Title = "repeated-failed-operation", "Change approach after repeated failures"
		base.Recommendation = "After the same operation fails repeatedly, inspect the failure cause or change the input instead of retrying unchanged."
	default:
		return domain.AdvisorRecommendation{}, false
	}
	return base, true
}

func signalPriority(tokens int64, signals int) string {
	switch {
	case tokens >= 25_000 || signals >= 5:
		return "high"
	case tokens >= 5_000 || signals >= 2:
		return "medium"
	default:
		return "low"
	}
}

func priorityRank(value string) int {
	switch value {
	case "high": return 3
	case "medium": return 2
	default: return 1
	}
}

func pressureRank(value string) int {
	switch value {
	case "critical": return 5
	case "high": return 4
	case "elevated": return 3
	case "normal": return 2
	case "low": return 1
	default: return 0
	}
}

func shortAdvisorSession(value string) string {
	if len(value) <= 10 {
		if value == "" { return "unknown" }
		return value
	}
	return value[:8] + "…"
}
