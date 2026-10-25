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
