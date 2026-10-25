package analytics

import (
	"testing"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

func TestBuildAdvisorPrioritizesPressureAndWasteEvidence(t *testing.T) {
	forecast := domain.ForecastReport{Ready: true, Pressure: "high", PressureRatio: 1.7, ProjectedTokens: 340000, HorizonMinutes: 120}
	attribution := domain.AttributionReport{
		Sessions: []domain.AttributionItem{{Session: "session-123456789", SharePercent: 72}},
		Agents: []domain.AttributionItem{{Key: "subagents", Label: "Subagents", SharePercent: 52}},
	}
	waste := domain.WasteReport{Rules: []domain.WasteRuleSummary{
		{RuleID: "repeated-search", Findings: 2, Signals: 3, EstimatedTokens: 8000},
		{RuleID: "oversized-tool-response", Findings: 1, Signals: 1, EstimatedTokens: 30000},
	}}

	report := BuildAdvisor(time.Now(), BurnRate{}, forecast, attribution, waste)
	if report.Mode != "advisory" || len(report.Recommendations) < 4 {
		t.Fatalf("unexpected advisor report: %+v", report)
	}
	ids := make(map[string]bool)
	for _, item := range report.Recommendations { ids[item.ID] = true }
	for _, id := range []string{"resource-pressure", "repeated-search", "oversized-tool-response", "subagent-pressure"} {
		if !ids[id] { t.Fatalf("expected recommendation %q in %+v", id, report.Recommendations) }
	}
	if report.High == 0 {
		t.Fatalf("expected at least one high-priority recommendation: %+v", report)
	}
}

func TestBuildAdvisorStaysQuietForNormalCleanWindow(t *testing.T) {
	forecast := domain.ForecastReport{Ready: true, Pressure: "normal", PressureRatio: 1.0, ProjectedTokens: 1000, HorizonMinutes: 120}
	report := BuildAdvisor(time.Now(), BurnRate{}, forecast, domain.AttributionReport{}, domain.WasteReport{})
	if len(report.Recommendations) != 0 || report.High != 0 || report.Medium != 0 || report.Low != 0 {
		t.Fatalf("expected no recommendations: %+v", report)
	}
}
