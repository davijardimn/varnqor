package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTranscriptChunkExtractsUsageAndToolMetadata(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "session.jsonl")
	content := "" +
		`{"type":"assistant","timestamp":"2026-10-25T18:00:00Z","requestId":"req-1","message":{"id":"msg-1","model":"claude-sonnet-4-6","stop_reason":"tool_use","usage":{"input_tokens":10,"output_tokens":20,"cache_creation_input_tokens":30,"cache_read_input_tokens":40},"content":[{"type":"tool_use","id":"toolu_1","name":"Read","input":{"file_path":"/repo/auth.go","offset":10,"limit":20}}]}}` + "\n" +
		`{"type":"user","timestamp":"2026-10-25T18:00:01Z","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"hello","is_error":false}]}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	chunk, err := ReadTranscriptChunk(TranscriptFile{Path: path, Project: "-repo", Session: "session-1", Size: int64(len(content))}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if chunk.Warnings != 0 || len(chunk.Usage) != 1 || len(chunk.ToolUses) != 1 {
		t.Fatalf("unexpected transcript chunk: %+v", chunk)
	}
	use := chunk.ToolUses[0]
	if use.ID != "toolu_1" || use.Name != "Read" || use.FilePath != "/repo/auth.go" || use.ReadOffset != 10 || use.ReadLimit != 20 {
		t.Fatalf("unexpected tool use: %+v", use)
	}
	if len(use.InputFingerprint) != 64 {
		t.Fatalf("expected SHA-256 input fingerprint, got %q", use.InputFingerprint)
	}
	if len(chunk.ToolResults) != 1 || chunk.ToolResults[0].ResultBytes != 5 {
		t.Fatalf("unexpected tool result metadata: %+v", chunk.ToolResults)
	}
	if chunk.NextOffset != int64(len(content)) {
		t.Fatalf("expected next offset %d, got %d", len(content), chunk.NextOffset)
	}
}

func TestReadTranscriptChunkCapturesSearchScopeWithoutQueryText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	content := `{"type":"assistant","timestamp":"2026-10-25T18:00:00Z","message":{"content":[{"type":"tool_use","id":"toolu_grep","name":"Grep","input":{"pattern":"VERY_PRIVATE_PATTERN","path":"/repo/src","glob":"*.go"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	chunk, err := ReadTranscriptChunk(TranscriptFile{Path: path, Project: "-repo", Session: "s1"}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunk.ToolUses) != 1 {
		t.Fatalf("expected one Grep use, got %d", len(chunk.ToolUses))
	}
	use := chunk.ToolUses[0]
	if use.FilePath != "/repo/src" || len(use.InputFingerprint) != 64 {
		t.Fatalf("unexpected Grep metadata: %+v", use)
	}
	if strings.Contains(use.InputFingerprint, "VERY_PRIVATE_PATTERN") {
		t.Fatal("fingerprint must not contain raw query text")
	}
}

func TestReadTranscriptChunkNeverStoresWriteContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	content := `{"type":"assistant","timestamp":"2026-10-25T18:00:00Z","message":{"content":[{"type":"tool_use","id":"toolu_write","name":"Write","input":{"file_path":"/repo/secret.txt","content":"DO_NOT_STORE_ME"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	chunk, err := ReadTranscriptChunk(TranscriptFile{Path: path, Project: "-repo", Session: "session-1"}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunk.ToolUses) != 1 {
		t.Fatalf("expected one tool use, got %d", len(chunk.ToolUses))
	}
	use := chunk.ToolUses[0]
	if use.Name != "Write" || use.FilePath != "/repo/secret.txt" || len(use.InputFingerprint) != 64 {
		t.Fatalf("unexpected write metadata: %+v", use)
	}
}
