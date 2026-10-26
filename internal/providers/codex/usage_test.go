package codex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadChunkTokenCount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rollout.jsonl")
	data := `{"timestamp":"2026-10-25T20:00:00Z","type":"session_meta","payload":{"id":"sess-1","cwd":"C:/work/varnqor","source":"cli"}}` + "\n" +
		`{"timestamp":"2026-10-25T20:00:01Z","type":"turn_context","payload":{"model":"gpt-5.6-codex"}}` + "\n" +
		`{"timestamp":"2026-10-25T20:00:02Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":120,"cached_input_tokens":20,"output_tokens":30,"total_tokens":150},"last_token_usage":{"input_tokens":120,"cached_input_tokens":20,"output_tokens":30,"total_tokens":150}}}}` + "\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil { t.Fatal(err) }
	chunk, err := ReadChunk(TranscriptFile{Path:path,Project:"varnqor",Session:"sess-1",Size:int64(len(data))}, 0)
	if err != nil { t.Fatal(err) }
	if len(chunk.Usage) != 1 { t.Fatalf("usage=%d", len(chunk.Usage)) }
	call := chunk.Usage[0].Call
	if call.Model != "gpt-5.6-codex" { t.Fatalf("model=%q", call.Model) }
	if call.Tokens.Input != 100 || call.Tokens.CacheRead != 20 || call.Tokens.Output != 30 || call.Tokens.Total != 150 { t.Fatalf("tokens=%+v", call.Tokens) }
}
