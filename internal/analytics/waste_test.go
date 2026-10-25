package analytics

import (
	"testing"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

func TestAnalyzeRepeatedFileReadsFindsExactRangeRepeats(t *testing.T) {
	start := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	calls := []domain.ToolCallObservation{
		{Timestamp: start.Add(2 * time.Minute), Name: "Read", Project: "-repo", Session: "s1", FilePath: "/repo/auth.go", ReadOffset: 1, ReadLimit: 200, ResultBytes: 4000, SourcePath: "/transcript/s1.jsonl"},
		{Timestamp: start.Add(7 * time.Minute), Name: "Read", Project: "-repo", Session: "s1", FilePath: "/repo/auth.go", ReadOffset: 1, ReadLimit: 200, ResultBytes: 4000, SourcePath: "/transcript/s1.jsonl"},
		{Timestamp: start.Add(12 * time.Minute), Name: "Read", Project: "-repo", Session: "s1", FilePath: "/repo/auth.go", ReadOffset: 1, ReadLimit: 200, ResultBytes: 8000, SourcePath: "/transcript/s1.jsonl"},
	}

	report := AnalyzeRepeatedFileReads(calls, start, start.Add(time.Hour))
	if report.RepeatedReads != 2 {
		t.Fatalf("expected 2 repeated reads, got %d", report.RepeatedReads)
	}
	if report.EstimatedRepeatedTokens != 3000 {
		t.Fatalf("expected ~3000 repeated tokens, got %d", report.EstimatedRepeatedTokens)
	}
	if report.Confidence != "medium" {
		t.Fatalf("expected medium confidence, got %q", report.Confidence)
	}
	if len(report.Findings) != 1 || report.Findings[0].Reads != 3 || report.Findings[0].Repeats != 2 {
		t.Fatalf("unexpected findings: %+v", report.Findings)
	}
}

func TestAnalyzeRepeatedFileReadsMutationBreaksSequence(t *testing.T) {
	start := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	calls := []domain.ToolCallObservation{
		{Timestamp: start.Add(time.Minute), Name: "Read", FilePath: "/repo/auth.go", ResultBytes: 4000, SourcePath: "s1"},
		{Timestamp: start.Add(3 * time.Minute), Name: "Edit", FilePath: "/repo/auth.go", SourcePath: "s1"},
		{Timestamp: start.Add(5 * time.Minute), Name: "Read", FilePath: "/repo/auth.go", ResultBytes: 4000, SourcePath: "s1"},
	}

	report := AnalyzeRepeatedFileReads(calls, start, start.Add(time.Hour))
	if report.RepeatedReads != 0 || len(report.Findings) != 0 {
		t.Fatalf("expected mutation to reset repeated-read state: %+v", report)
	}
}

func TestAnalyzeRepeatedFileReadsDifferentRangesDoNotMatch(t *testing.T) {
	start := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	calls := []domain.ToolCallObservation{
		{Timestamp: start.Add(time.Minute), Name: "Read", FilePath: "/repo/auth.go", ReadOffset: 1, ReadLimit: 100, ResultBytes: 1000, SourcePath: "s1"},
		{Timestamp: start.Add(2 * time.Minute), Name: "Read", FilePath: "/repo/auth.go", ReadOffset: 101, ReadLimit: 100, ResultBytes: 1000, SourcePath: "s1"},
	}

	report := AnalyzeRepeatedFileReads(calls, start, start.Add(time.Hour))
	if report.RepeatedReads != 0 {
		t.Fatalf("expected different read ranges not to match: %+v", report)
	}
}

func TestAnalyzeWasteDetectsRepeatedSearch(t *testing.T) {
	start := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	calls := []domain.ToolCallObservation{
		{Timestamp: start.Add(time.Minute), Name: "Grep", Project: "-repo", Session: "s1", InputFingerprint: "same", ResultBytes: 8000, SourcePath: "s1"},
		{Timestamp: start.Add(5 * time.Minute), Name: "Grep", Project: "-repo", Session: "s1", InputFingerprint: "same", ResultBytes: 8000, SourcePath: "s1"},
		{Timestamp: start.Add(8 * time.Minute), Name: "Grep", Project: "-repo", Session: "s1", InputFingerprint: "same", ResultBytes: 8000, SourcePath: "s1"},
	}

	report := AnalyzeWaste(calls, start, start.Add(time.Hour))
	if report.RepeatedSearches != 2 {
		t.Fatalf("expected 2 repeated searches, got %d", report.RepeatedSearches)
	}
	if len(report.Findings) != 1 || report.Findings[0].RuleID != "repeated-search" {
		t.Fatalf("unexpected search findings: %+v", report.Findings)
	}
	if report.Findings[0].EstimatedTokens != 4000 {
		t.Fatalf("expected 4000 estimated tokens, got %d", report.Findings[0].EstimatedTokens)
	}
}

func TestAnalyzeWasteDetectsOversizedResponseUsingExcessOnly(t *testing.T) {
	start := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	calls := []domain.ToolCallObservation{{
		Timestamp: start.Add(time.Minute), Name: "Read", Project: "-repo", Session: "s1",
		FilePath: "/repo/large.go", ResultBytes: 40 * 1024, SourcePath: "s1",
	}}

	report := AnalyzeWaste(calls, start, start.Add(time.Hour))
	if report.OversizedResponses != 1 {
		t.Fatalf("expected one oversized response, got %d", report.OversizedResponses)
	}
	var found *domain.WasteFinding
	for i := range report.Findings {
		if report.Findings[i].RuleID == "oversized-tool-response" {
			found = &report.Findings[i]
		}
	}
	if found == nil || found.EstimatedTokens != 2048 {
		t.Fatalf("expected 2048 excess-context tokens, got %+v", found)
	}
}

func TestAnalyzeWasteDetectsRepeatedFailuresAndSuccessResets(t *testing.T) {
	start := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	calls := []domain.ToolCallObservation{
		{Timestamp: start.Add(time.Minute), Name: "Bash", InputFingerprint: "cmd-a", IsError: true, ResultBytes: 400, SourcePath: "s1"},
		{Timestamp: start.Add(4 * time.Minute), Name: "Bash", InputFingerprint: "cmd-a", IsError: true, ResultBytes: 400, SourcePath: "s1"},
		{Timestamp: start.Add(6 * time.Minute), Name: "Bash", InputFingerprint: "cmd-a", IsError: false, ResultBytes: 100, SourcePath: "s1"},
		{Timestamp: start.Add(8 * time.Minute), Name: "Bash", InputFingerprint: "cmd-a", IsError: true, ResultBytes: 400, SourcePath: "s1"},
	}

	report := AnalyzeWaste(calls, start, start.Add(time.Hour))
	if report.RepeatedFailures != 1 {
		t.Fatalf("expected one repeated failure before the success reset, got %d", report.RepeatedFailures)
	}
}
