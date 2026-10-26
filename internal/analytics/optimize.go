package analytics

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

// BuildOptimization turns observed resource signals into bounded, explainable
// playbooks. Savings ranges are deliberately non-additive because multiple
// actions can affect the same future tokens. Varnqor 3 never executes these
// actions against an external agent automatically.
func BuildOptimization(now time.Time, report domain.ExploreReport, burn BurnRate, forecast domain.ForecastReport, waste domain.WasteReport, policy domain.OptimizationPolicy) domain.OptimizationReport {
	if policy.Goal == "" { policy.Goal = "balanced" }
	if policy.HorizonMinutes <= 0 { policy.HorizonMinutes = 120 }
	if policy.TargetReductionPercent <= 0 { policy.TargetReductionPercent = 20 }
	if policy.ExecutionMode == "" { policy.ExecutionMode = "preview" }

	currentPerHour := burn.CurrentPerHour
	if currentPerHour <= 0 && report.Filter.RangeMinutes > 0 && report.Current.Tokens.Total > 0 {
		currentPerHour = int64(math.Round(float64(report.Current.Tokens.Total) * 60 / float64(report.Filter.RangeMinutes)))
	}
	projected := currentPerHour * int64(policy.HorizonMinutes) / 60
	if forecast.CurrentPerHour > 0 && policy.HorizonMinutes == forecast.HorizonMinutes {
		projected = forecast.ProjectedTokens
	}
	targetSavings := int64(math.Round(float64(projected) * policy.TargetReductionPercent / 100))
	if targetSavings > projected { targetSavings = projected }
	targetTokens := projected - targetSavings
	if targetTokens < 0 { targetTokens = 0 }

	result := domain.OptimizationReport{
		GeneratedAt: now,
		State: "stable",
		Goal: policy.Goal,
		Mode: policy.ExecutionMode,
		ExecutionAvailable: false,
		Scenario: domain.OptimizationScenario{
			CurrentPerHour: currentPerHour,
			HorizonMinutes: policy.HorizonMinutes,
			ProjectedTokens: projected,
			TargetReductionPercent: policy.TargetReductionPercent,
			TargetTokens: targetTokens,
			TargetSavings: targetSavings,
			Feasibility: optimizationFeasibility(policy.TargetReductionPercent),
		},
		Actions: []domain.OptimizationAction{},
		ProvidersObserved: len(report.Providers),
		ProjectsObserved: len(report.Projects),
		SessionsObserved: len(report.Sessions),
		Note: "Savings ranges are heuristic, overlap is possible, and actions are never executed automatically.",
	}
	if !policy.Enabled {
		result.State = "disabled"
		return result
	}
	if report.Current.Tokens.Total <= 0 || currentPerHour <= 0 {
		result.State = "idle"
		return result
	}

	actions := make([]domain.OptimizationAction, 0, 6)
	add := func(action domain.OptimizationAction) {
		action.NonAdditive = true
		if action.EstimatedSavingsLow < 0 { action.EstimatedSavingsLow = 0 }
		if action.EstimatedSavingsHigh < action.EstimatedSavingsLow { action.EstimatedSavingsHigh = action.EstimatedSavingsLow }
		if action.EstimatedSavingsHigh > projected { action.EstimatedSavingsHigh = projected }
		if action.EstimatedSavingsLow > action.EstimatedSavingsHigh { action.EstimatedSavingsLow = action.EstimatedSavingsHigh }
		actions = append(actions, action)
	}

	if waste.EstimatedSignalTokens > 0 {
		scale := math.Max(1, float64(policy.HorizonMinutes)/60)
		signal := int64(math.Round(float64(waste.EstimatedSignalTokens) * scale))
		low, high := boundedImpact(signal, 0.30, 0.75, projected)
		priority := "medium"
		if waste.EstimatedSignalTokens >= max64(2500, report.Current.Tokens.Total/10) { priority = "high" }
		add(domain.OptimizationAction{
			ID: "trim-tool-context", Kind: "tool-context", Priority: priority,
			Title: "Trim repeated tool context",
			Summary: "Remove repeated reads, searches, failed retries, and oversized tool-result context before it compounds.",
			Evidence: fmt.Sprintf("Claude tool heuristics found ~%d signal-context tokens in the latest analysis window.", waste.EstimatedSignalTokens),
			Confidence: string(waste.Confidence), EstimatedSavingsLow: low, EstimatedSavingsHigh: high,
			Steps: []string{"Inspect the highest-confidence Waste findings.", "Narrow repeated reads/searches to the smallest useful scope.", "Checkpoint the result before repeating the same tool path."},
		})
	}

	if len(report.Sessions) > 0 {
		top := report.Sessions[0]
		if top.SharePercent >= 45 {
			base := int64(math.Round(float64(projected) * top.SharePercent / 100))
			low, high := boundedImpact(base, 0.05, 0.15, projected)
			priority := "medium"
			if top.SharePercent >= 70 { priority = "high" }
			add(domain.OptimizationAction{
				ID: "checkpoint-dominant-session", Kind: "session-checkpoint", Priority: priority,
				Title: "Checkpoint the dominant session",
				Summary: "A single session owns a large part of the selected window. Preserve the useful state, then continue from a tighter context boundary.",
				Evidence: fmt.Sprintf("Session %s in %s accounts for %.1f%% of selected-window tokens.", shortID(top.Session), top.Project, top.SharePercent),
				Provider: top.Provider, Project: top.Project, Session: top.Session, Confidence: "medium", EstimatedSavingsLow: low, EstimatedSavingsHigh: high,
				Steps: []string{"Capture the current objective, decisions, and unresolved work in a compact checkpoint.", "Start the next phase from that checkpoint instead of the full historical context.", "Keep the original session available for audit rather than deleting it."},
			})
		}
	}

	if burn.Ready && burn.Ratio >= 1.35 {
		low := int64(math.Round(float64(targetSavings) * 0.35))
		high := targetSavings
		priority := "medium"
		if burn.Ratio >= 1.75 || burn.Level == "critical" { priority = "high" }
		add(domain.OptimizationAction{
			ID: "pace-envelope", Kind: "pace-control", Priority: priority,
			Title: "Bring the next work block back inside the learned pace envelope",
			Summary: "Current token pace is materially above the active baseline. Use a bounded work block before expanding scope again.",
			Evidence: fmt.Sprintf("Current pace is %.2fx the learned active baseline (%d vs %d tokens/h).", burn.Ratio, burn.CurrentPerHour, burn.BaselinePerHour),
			Confidence: burn.Confidence, EstimatedSavingsLow: low, EstimatedSavingsHigh: high,
			Steps: []string{"Finish the current atomic task before opening a new workstream.", "Prefer targeted retrieval over broad repository reads during the next block.", "Re-check Burn and Pulse after one refresh interval."},
		})
	}

	promptTokens := report.Current.Tokens.Input + report.Current.Tokens.CacheCreation + report.Current.Tokens.CacheRead
	cacheTelemetry := report.Current.Tokens.CacheCreation > 0 || report.Current.Tokens.CacheRead > 0
	if cacheTelemetry && promptTokens > 0 && report.Current.CacheReadShare < 20 {
		low, high := boundedImpact(projected, 0.02, 0.08, projected)
		add(domain.OptimizationAction{
			ID: "cache-reuse", Kind: "cache-reuse", Priority: "low",
			Title: "Improve reusable context boundaries",
			Summary: "Cache-read share is low in a window where cache telemetry is available. Stable context boundaries may reduce repeated prompt transfer.",
			Evidence: fmt.Sprintf("Cache reads account for %.1f%% of prompt-side tokens in the selected window.", report.Current.CacheReadShare),
			Confidence: "low", EstimatedSavingsLow: low, EstimatedSavingsHigh: high,
			Steps: []string{"Keep stable instructions and reference context together where the provider can reuse them.", "Avoid needlessly changing large invariant context between adjacent calls.", "Treat this as provider-dependent; verify the cache-read share improves before relying on it."},
		})
	}

	if report.Current.Sessions >= 6 && report.Current.Calls >= report.Current.Sessions && float64(report.Current.Calls)/float64(report.Current.Sessions) < 3.5 {
		low, high := boundedImpact(projected, 0.02, 0.07, projected)
		add(domain.OptimizationAction{
			ID: "session-fragmentation", Kind: "session-fragmentation", Priority: "low",
			Title: "Reduce session fragmentation",
			Summary: "The selected window contains many short sessions. Consolidating closely related work can reduce repeated setup and rediscovery.",
			Evidence: fmt.Sprintf("%d calls are spread across %d token-bearing sessions (%.1f calls/session).", report.Current.Calls, report.Current.Sessions, float64(report.Current.Calls)/float64(report.Current.Sessions)),
			Confidence: "medium", EstimatedSavingsLow: low, EstimatedSavingsHigh: high,
			Steps: []string{"Group closely related edits into one deliberate session.", "Carry forward a compact checkpoint when switching agents.", "Do not merge unrelated tasks merely to improve this metric."},
		})
	}

	if policy.Goal == "economy" && report.Current.Tokens.Total > 0 {
		outputShare := float64(report.Current.Tokens.Output) / float64(report.Current.Tokens.Total) * 100
		if outputShare >= 30 {
			low, high := boundedImpact(projected, 0.03, 0.10, projected)
			add(domain.OptimizationAction{
				ID: "output-budget", Kind: "output-budget", Priority: "low",
				Title: "Use an explicit response budget for routine agent turns",
				Summary: "Economy mode can trim verbose intermediate responses while leaving complex reasoning untouched when it is needed.",
				Evidence: fmt.Sprintf("Output tokens are %.1f%% of selected-window activity.", outputShare),
				Confidence: "low", EstimatedSavingsLow: low, EstimatedSavingsHigh: high,
				Steps: []string{"Ask for concise progress updates during routine tool loops.", "Keep full explanations for decisions, failures, and review points.", "Measure whether output share falls without reducing task quality."},
			})
		}
	}

	sort.SliceStable(actions, func(i, j int) bool {
		pi, pj := optimizationPriorityRank(actions[i].Priority), optimizationPriorityRank(actions[j].Priority)
		if pi != pj { return pi > pj }
		return actions[i].EstimatedSavingsHigh > actions[j].EstimatedSavingsHigh
	})
	if len(actions) > 6 { actions = actions[:6] }
	result.Actions = actions
	if len(actions) > 0 { result.State = "opportunity" }
	return result
}

func boundedImpact(base int64, lowRate, highRate float64, capValue int64) (int64, int64) {
	if base <= 0 || capValue <= 0 { return 0, 0 }
	low := int64(math.Round(float64(base) * lowRate))
	high := int64(math.Round(float64(base) * highRate))
	if high > capValue { high = capValue }
	if low > high { low = high }
	return low, high
}

func optimizationFeasibility(target float64) string {
	switch {
	case target <= 15:
		return "conservative"
	case target <= 30:
		return "balanced"
	default:
		return "ambitious"
	}
}

func optimizationPriorityRank(value string) int {
	switch value {
	case "high": return 3
	case "medium": return 2
	case "low": return 1
	default: return 0
	}
}

func shortID(value string) string {
	if len(value) <= 10 { return value }
	return value[:8] + "…"
}
