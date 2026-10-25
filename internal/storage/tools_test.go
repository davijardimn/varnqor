package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestToolMetadataRoundTrip(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "varnqor.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if rebuilt, err := store.EnsureToolIndex("claude"); err != nil {
		t.Fatal(err)
	} else if rebuilt {
		t.Fatal("fresh database should not require a tool-index backfill")
	}

	now := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	uses := []ToolUseRecord{{
		Provider: "claude", ToolUseID: "toolu_1", Timestamp: now, ToolName: "Read",
		Project: "-repo", Session: "s1", FilePath: "/repo/auth.go", ReadOffset: 1, ReadLimit: 200,
		InputFingerprint: "abc123", SourcePath: "/transcripts/s1.jsonl", SourceOffset: 100,
	}}
	results := []ToolResultRecord{{
		Provider: "claude", ToolUseID: "toolu_1", ResultBytes: 4096,
		SourcePath: "/transcripts/s1.jsonl", SourceOffset: 200,
	}}
	if err := store.ApplyToolEvents(uses, results); err != nil {
		t.Fatal(err)
	}

	calls, err := store.ToolCalls("claude", now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	call := calls[0]
	if call.Name != "Read" || call.FilePath != "/repo/auth.go" || call.ResultBytes != 4096 || call.InputFingerprint != "abc123" {
		t.Fatalf("unexpected tool metadata: %+v", call)
	}
	if call.SourcePath != "/transcripts/s1.jsonl" {
		t.Fatalf("expected internal source path to round-trip, got %q", call.SourcePath)
	}
}
