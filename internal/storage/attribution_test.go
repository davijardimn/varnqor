package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

func TestAttributionRanksWindowDrivers(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "attribution.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	now := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	records := []UsageRecord{
		fakeAttributionRecord("a", now.Add(-50*time.Minute), "project-a", "session-a", "opus", false, 600),
		fakeAttributionRecord("b", now.Add(-40*time.Minute), "project-a", "session-a", "opus", true, 200),
		fakeAttributionRecord("c", now.Add(-30*time.Minute), "project-b", "session-b", "sonnet", false, 200),
		fakeAttributionRecord("old", now.Add(-2*time.Hour), "project-old", "session-old", "sonnet", false, 900),
	}
	if err := store.ApplyFile(SourceState{Provider: "claude", Path: "fixture.jsonl", Size: 1, Offset: 1}, records); err != nil {
		t.Fatal(err)
	}

	report, err := store.Attribution("claude", now.Add(-time.Hour), now, 5)
	if err != nil {
		t.Fatal(err)
	}
	if report.Tokens.Total != 1000 || report.Calls != 3 {
		t.Fatalf("unexpected window total: calls=%d tokens=%d", report.Calls, report.Tokens.Total)
	}
	if len(report.Projects) != 2 || report.Projects[0].Key != "project-a" || report.Projects[0].Tokens.Total != 800 {
		t.Fatalf("unexpected project attribution: %+v", report.Projects)
	}
	if report.Projects[0].SharePercent != 80 {
		t.Fatalf("expected 80%% project share, got %.1f", report.Projects[0].SharePercent)
	}
	if len(report.Sessions) != 2 || report.Sessions[0].Session != "session-a" {
		t.Fatalf("unexpected session attribution: %+v", report.Sessions)
	}
	if len(report.Models) != 2 || report.Models[0].Key != "opus" || report.Models[0].SharePercent != 80 {
		t.Fatalf("unexpected model attribution: %+v", report.Models)
	}
	if len(report.Agents) != 2 || report.Agents[0].Key != "main" || report.Agents[0].Tokens.Total != 800 {
		t.Fatalf("unexpected agent attribution: %+v", report.Agents)
	}
}

func TestAttributionRespectsLimit(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "limit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	now := time.Now()
	records := []UsageRecord{
		fakeAttributionRecord("a", now.Add(-10*time.Minute), "a", "a", "m1", false, 300),
		fakeAttributionRecord("b", now.Add(-9*time.Minute), "b", "b", "m2", false, 200),
		fakeAttributionRecord("c", now.Add(-8*time.Minute), "c", "c", "m3", false, 100),
	}
	if err := store.ApplyFile(SourceState{Provider: "claude", Path: "limit.jsonl", Size: 1, Offset: 1}, records); err != nil {
		t.Fatal(err)
	}

	report, err := store.Attribution("claude", now.Add(-time.Hour), now, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Projects) != 2 || len(report.Sessions) != 2 || len(report.Models) != 2 {
		t.Fatalf("expected top 2 dimensions, got projects=%d sessions=%d models=%d", len(report.Projects), len(report.Sessions), len(report.Models))
	}
}

func fakeAttributionRecord(key string, at time.Time, project, session, model string, subagent bool, total int64) UsageRecord {
	return UsageRecord{
		Key: key,
		Provider: "claude",
		SourcePath: "fixture.jsonl",
		Call: domain.UsageCall{
			MessageID: key,
			Timestamp: at,
			Project: project,
			Session: session,
			Model: model,
			Subagent: subagent,
			StopReason: "end_turn",
			Tokens: domain.TokenTotals{Input: total, Total: total},
		},
	}
}
