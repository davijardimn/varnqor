package claude

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanUsageTodayDeduplicatesStreamingSnapshots(t *testing.T) {
	root := t.TempDir()
	projects := filepath.Join(root, ".claude", "projects")
	project := filepath.Join(projects, "-work-api")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}

	transcript := filepath.Join(project, "session-1.jsonl")
	content := fmt.Sprintf("%s\n%s\n%s\n%s\n",
		usageLine("msg-1", "req-1", "2026-10-25T12:00:00Z", "claude-sonnet-4-6", "", 10, 9, 20, 30, 12, 8),
		usageLine("msg-1", "req-1", "2026-10-25T12:00:01Z", "claude-sonnet-4-6", "tool_use", 10, 159, 20, 30, 12, 8),
		usageLine("msg-2", "", "2026-10-25T12:05:00Z", "claude-sonnet-4-6", "", 5, 1, 0, 95, 0, 0),
		usageLine("msg-2", "", "2026-10-25T12:05:01Z", "claude-sonnet-4-6", "end_turn", 5, 20, 0, 95, 0, 0),
	)
	if err := os.WriteFile(transcript, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	overview, err := ScanUsageToday(projects, now)
	if err != nil {
		t.Fatal(err)
	}

	if overview.Calls != 2 {
		t.Fatalf("expected 2 unique calls, got %d", overview.Calls)
	}
	if overview.Projects != 1 || overview.Sessions != 1 {
		t.Fatalf("unexpected scope counts: projects=%d sessions=%d", overview.Projects, overview.Sessions)
	}
	if overview.Tokens.Input != 15 || overview.Tokens.Output != 179 {
		t.Fatalf("unexpected input/output totals: %+v", overview.Tokens)
	}
	if overview.Tokens.CacheCreation != 20 || overview.Tokens.CacheRead != 125 {
		t.Fatalf("unexpected cache totals: %+v", overview.Tokens)
	}
	if overview.Tokens.CacheCreation5m != 12 || overview.Tokens.CacheCreation1h != 8 {
		t.Fatalf("unexpected cache TTL split: %+v", overview.Tokens)
	}
	if overview.Tokens.Total != 339 {
		t.Fatalf("expected 339 total activity tokens, got %d", overview.Tokens.Total)
	}
	if math.Abs(overview.CacheReadShare-78.125) > 0.001 {
		t.Fatalf("unexpected cache read share: %f", overview.CacheReadShare)
	}
	if len(overview.Models) != 1 || overview.Models[0].Model != "claude-sonnet-4-6" {
		t.Fatalf("unexpected model breakdown: %+v", overview.Models)
	}
}

func TestScanUsageTodayDeduplicatesAcrossFiles(t *testing.T) {
	projects := filepath.Join(t.TempDir(), "projects")
	project := filepath.Join(projects, "-repo")
	subagents := filepath.Join(project, "session-2", "subagents")
	if err := os.MkdirAll(subagents, 0o755); err != nil {
		t.Fatal(err)
	}

	line := usageLine("msg-shared", "req-shared", "2026-10-25T10:00:00Z", "claude-opus-4-6", "end_turn", 7, 11, 13, 17, 13, 0)
	if err := os.WriteFile(filepath.Join(project, "session-2.jsonl"), []byte(line+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subagents, "agent-a.jsonl"), []byte(line+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	overview, err := ScanUsageToday(projects, time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if overview.Calls != 1 {
		t.Fatalf("expected duplicate call across files to count once, got %d", overview.Calls)
	}
	if overview.Tokens.Total != 48 {
		t.Fatalf("expected 48 total activity tokens, got %d", overview.Tokens.Total)
	}
	if overview.FilesScanned != 2 {
		t.Fatalf("expected 2 files scanned, got %d", overview.FilesScanned)
	}
}

func TestScanUsageTodayCountsMalformedUsageAsWarning(t *testing.T) {
	projects := filepath.Join(t.TempDir(), "projects")
	project := filepath.Join(projects, "-repo")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "session.jsonl"), []byte("{not-json \"usage\": true}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	overview, err := ScanUsageToday(projects, time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if overview.ParseWarnings != 1 {
		t.Fatalf("expected 1 parse warning, got %d", overview.ParseWarnings)
	}
}

func usageLine(messageID, requestID, timestamp, model, stopReason string, input, output, cacheCreation, cacheRead, cache5m, cache1h int64) string {
	stop := "null"
	if stopReason != "" {
		stop = fmt.Sprintf("%q", stopReason)
	}
	return fmt.Sprintf(`{"type":"assistant","timestamp":%q,"requestId":%q,"message":{"id":%q,"model":%q,"stop_reason":%s,"usage":{"input_tokens":%d,"output_tokens":%d,"cache_creation_input_tokens":%d,"cache_read_input_tokens":%d,"cache_creation":{"ephemeral_5m_input_tokens":%d,"ephemeral_1h_input_tokens":%d}}}}`, timestamp, requestID, messageID, model, stop, input, output, cacheCreation, cacheRead, cache5m, cache1h)
}
