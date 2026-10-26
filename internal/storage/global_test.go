package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

func TestPrepareV1CleansZeroTokenRowsAndAggregatesProviders(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "varnqor.db"))
	if err != nil { t.Fatal(err) }
	defer store.Close()

	now := time.Date(2026, 10, 25, 20, 0, 0, 0, time.Local)
	rows := []struct{ provider, key, model, project, session string; total int64 }{
		{"claude", "c-real", "claude-sonnet", "alpha", "s1", 120},
		{"codex", "x-real", "gpt-codex", "alpha", "s2", 80},
		{"gemini", "g-real", "gemini-pro", "beta", "s3", 50},
		{"claude", "legacy-zero", "<synthetic>", "ghost", "ghost", 0},
	}
	for i, row := range rows {
		call := domain.UsageCall{Timestamp: now.Add(time.Duration(i) * time.Minute), Model: row.model, Project: row.project, Session: row.session, Tokens: domain.TokenTotals{Input: row.total, Total: row.total}}
		if err := store.ApplyFile(SourceState{Provider: row.provider, Path: row.provider + ".jsonl", Size: 10, ModTimeNS: 1, Offset: 10}, []UsageRecord{{Key: row.key, Provider: row.provider, SourcePath: row.provider + ".jsonl", Call: call}}); err != nil { t.Fatal(err) }
	}
	if err := store.PrepareV1(); err != nil { t.Fatal(err) }
	usage, _, err := store.DailyUsageAll(now)
	if err != nil { t.Fatal(err) }
	if usage.Calls != 3 { t.Fatalf("calls=%d, want 3", usage.Calls) }
	if usage.Tokens.Total != 250 { t.Fatalf("tokens=%d, want 250", usage.Tokens.Total) }
	if usage.Sessions != 3 { t.Fatalf("sessions=%d, want 3", usage.Sessions) }
	if usage.Projects != 3 { t.Fatalf("provider-aware projects=%d, want 3", usage.Projects) }
	for _, model := range usage.Models { if model.Model == "<synthetic>" { t.Fatal("synthetic model leaked into v1 usage") } }
	if err := store.PrepareV1(); err != nil { t.Fatalf("PrepareV1 must be idempotent: %v", err) }
}

func TestAttributionAllKeepsProviderIdentity(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "varnqor.db")); if err != nil { t.Fatal(err) }; defer store.Close()
	now := time.Now()
	for _, provider := range []string{"claude", "codex"} {
		call := domain.UsageCall{Timestamp: now.Add(-time.Minute), Model: "model", Project: "same-project", Session: "same-session", Tokens: domain.TokenTotals{Input: 100, Total: 100}}
		if err := store.ApplyFile(SourceState{Provider:provider,Path:provider+".jsonl",Size:1,ModTimeNS:1,Offset:1}, []UsageRecord{{Key:provider+"-1",Provider:provider,SourcePath:provider+".jsonl",Call:call}}); err != nil { t.Fatal(err) }
	}
	report, err := store.AttributionAll(now.Add(-time.Hour), now, 10); if err != nil { t.Fatal(err) }
	if len(report.Projects) != 2 { t.Fatalf("projects=%d, want 2 provider-separated rows", len(report.Projects)) }
	if report.Projects[0].Provider == report.Projects[1].Provider { t.Fatalf("provider identity was lost: %+v", report.Projects) }
}
