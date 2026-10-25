package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davijardimn/varnqor/internal/storage"
)

func TestSyncClaudeIndexesIncrementallyAndBuildsHourlyHistory(t *testing.T) {
	root := t.TempDir()
	projects := filepath.Join(root, "projects")
	project := filepath.Join(projects, "-repo")
	if err := os.MkdirAll(project, 0o755); err != nil { t.Fatal(err) }

	transcript := filepath.Join(project, "session-1.jsonl")
	first := testUsageLine("msg-1", "req-1", "2026-10-25T12:05:00Z", 10, 20, 30, 40)
	if err := os.WriteFile(transcript, []byte(first+"\n"), 0o600); err != nil { t.Fatal(err) }

	store, err := storage.Open(filepath.Join(root, "varnqor.db"))
	if err != nil { t.Fatal(err) }
	defer store.Close()

	stats, err := SyncClaude(store, projects)
	if err != nil { t.Fatal(err) }
	if stats.FilesChanged != 1 || stats.RecordsObserved != 1 || stats.BytesIndexed == 0 {
		t.Fatalf("unexpected first sync stats: %+v", stats)
	}

	now := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	overview, hourly, err := store.DailyUsage("claude", now)
	if err != nil { t.Fatal(err) }
	if overview.Calls != 1 || overview.Tokens.Total != 100 {
		t.Fatalf("unexpected first overview: %+v", overview)
	}
	if len(hourly) != 24 || hourly[12].Calls != 1 || hourly[12].Tokens.Total != 100 {
		t.Fatalf("unexpected hourly series: %+v", hourly)
	}

	stats, err = SyncClaude(store, projects)
	if err != nil { t.Fatal(err) }
	if stats.FilesChanged != 0 || stats.BytesIndexed != 0 {
		t.Fatalf("unchanged sync should be idle: %+v", stats)
	}

	second := testUsageLine("msg-2", "req-2", "2026-10-25T13:10:00Z", 5, 6, 7, 8)
	file, err := os.OpenFile(transcript, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil { t.Fatal(err) }
	if _, err := file.WriteString(second + "\n"); err != nil { t.Fatal(err) }
	_ = file.Close()

	stats, err = SyncClaude(store, projects)
	if err != nil { t.Fatal(err) }
	if stats.FilesChanged != 1 || stats.RecordsObserved != 1 {
		t.Fatalf("append sync should only observe new record: %+v", stats)
	}
	overview, hourly, err = store.DailyUsage("claude", now)
	if err != nil { t.Fatal(err) }
	if overview.Calls != 2 || overview.Tokens.Total != 126 {
		t.Fatalf("unexpected appended overview: %+v", overview)
	}
	if hourly[13].Calls != 1 || hourly[13].Tokens.Total != 26 {
		t.Fatalf("unexpected second hour: %+v", hourly[13])
	}
}

func TestSyncClaudeRebuildsAfterTranscriptRewrite(t *testing.T) {
	root := t.TempDir()
	projects := filepath.Join(root, "projects")
	project := filepath.Join(projects, "-repo")
	if err := os.MkdirAll(project, 0o755); err != nil { t.Fatal(err) }
	transcript := filepath.Join(project, "session.jsonl")
	if err := os.WriteFile(transcript, []byte(testUsageLine("old", "req-old", "2026-10-25T10:00:00Z", 10, 10, 10, 10)+"\n"), 0o600); err != nil { t.Fatal(err) }

	store, err := storage.Open(filepath.Join(root, "varnqor.db"))
	if err != nil { t.Fatal(err) }
	defer store.Close()
	if _, err := SyncClaude(store, projects); err != nil { t.Fatal(err) }

	if err := os.WriteFile(transcript, []byte(testUsageLine("new", "req-new", "2026-10-25T11:00:00Z", 1, 2, 3, 4)+"\n"), 0o600); err != nil { t.Fatal(err) }
	stats, err := SyncClaude(store, projects)
	if err != nil { t.Fatal(err) }
	if !stats.Rebuilt {
		t.Fatalf("expected safe rebuild after rewrite: %+v", stats)
	}

	overview, _, err := store.DailyUsage("claude", time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC))
	if err != nil { t.Fatal(err) }
	if overview.Calls != 1 || overview.Tokens.Total != 10 {
		t.Fatalf("stale records survived rebuild: %+v", overview)
	}
}

func testUsageLine(messageID, requestID, timestamp string, input, output, cacheCreation, cacheRead int64) string {
	return fmt.Sprintf(`{"type":"assistant","timestamp":%q,"requestId":%q,"message":{"id":%q,"model":"claude-sonnet-4-6","stop_reason":"end_turn","usage":{"input_tokens":%d,"output_tokens":%d,"cache_creation_input_tokens":%d,"cache_read_input_tokens":%d}}}`, timestamp, requestID, messageID, input, output, cacheCreation, cacheRead)
}
