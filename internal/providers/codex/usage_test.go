package codex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadChunkLegacyTokenCount(t *testing.T) {
	dir := t.TempDir(); path := filepath.Join(dir, "rollout.jsonl")
	data := `{"timestamp":"2026-10-25T20:00:00Z","type":"session_meta","payload":{"id":"sess-1","cwd":"C:/work/varnqor","source":"cli"}}` + "\n" +
		`{"timestamp":"2026-10-25T20:00:01Z","type":"turn_context","payload":{"model":"gpt-5.6-codex"}}` + "\n" +
		`{"timestamp":"2026-10-25T20:00:02Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":120,"cached_input_tokens":20,"output_tokens":30,"total_tokens":150},"last_token_usage":{"input_tokens":120,"cached_input_tokens":20,"output_tokens":30,"total_tokens":150}}}}` + "\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil { t.Fatal(err) }
	chunk, err := ReadChunk(TranscriptFile{Path:path,Project:"varnqor",Session:"sess-1",Size:int64(len(data))}, 0); if err != nil { t.Fatal(err) }
	if len(chunk.Usage) != 1 { t.Fatalf("usage=%d", len(chunk.Usage)) }
	call := chunk.Usage[0].Call
	if call.Model != "gpt-5.6-codex" { t.Fatalf("model=%q", call.Model) }
	if call.Tokens.Input != 100 || call.Tokens.CacheRead != 20 || call.Tokens.Output != 30 || call.Tokens.Total != 150 { t.Fatalf("tokens=%+v", call.Tokens) }
}

func TestReadChunkPrefersPerResponseUsageAndAvoidsDoubleCount(t *testing.T) {
	dir := t.TempDir(); path := filepath.Join(dir, "rollout-modern.jsonl")
	data := `{"timestamp":"2026-10-25T20:00:00Z","type":"session_meta","payload":{"id":"sess-modern","cwd":"C:/work/varnqor","source":"cli"}}` + "\n" +
		`{"timestamp":"2026-10-25T20:00:01Z","type":"turn_context","payload":{"model":"gpt-6-astra"}}` + "\n" +
		`{"timestamp":"2026-10-25T20:00:02Z","type":"token_usage_record","payload":{"response_id":"resp-1","usage":{"input_tokens":100,"cached_input_tokens":80,"cache_write_input_tokens":5,"output_tokens":20,"reasoning_output_tokens":3,"total_tokens":120},"turn_token_usage":{"input_tokens":100,"output_tokens":20,"total_tokens":120},"thread_token_usage":{"input_tokens":100,"output_tokens":20,"total_tokens":120}}}` + "\n" +
		`{"timestamp":"2026-10-25T20:00:03Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":100,"cached_input_tokens":80,"output_tokens":20,"total_tokens":120},"last_token_usage":{"input_tokens":100,"cached_input_tokens":80,"output_tokens":20,"total_tokens":120}}}}` + "\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil { t.Fatal(err) }
	chunk, err := ReadChunk(TranscriptFile{Path:path,Project:"varnqor",Session:"sess-modern",Size:int64(len(data))}, 0); if err != nil { t.Fatal(err) }
	if len(chunk.Usage) != 1 { t.Fatalf("modern rollout double-counted: usage=%d", len(chunk.Usage)) }
	call := chunk.Usage[0].Call
	if chunk.Usage[0].Key != "resp-1" || call.RequestID != "resp-1" { t.Fatalf("response identity=%q/%q", chunk.Usage[0].Key, call.RequestID) }
	if call.Model != "gpt-6-astra" { t.Fatalf("model=%q", call.Model) }
	if call.Tokens.Input != 20 || call.Tokens.CacheRead != 80 || call.Tokens.CacheCreation != 5 || call.Tokens.Output != 20 || call.Tokens.Total != 120 { t.Fatalf("tokens=%+v", call.Tokens) }
}

func TestReadChunkRestoresModelAcrossIncrementalOffset(t *testing.T) {
	dir := t.TempDir(); path := filepath.Join(dir, "rollout-offset.jsonl")
	prefix := `{"timestamp":"2026-10-25T20:00:00Z","type":"session_meta","payload":{"id":"sess-offset","cwd":"C:/work/varnqor","source":"cli"}}` + "\n" +
		`{"timestamp":"2026-10-25T20:00:01Z","type":"turn_context","payload":{"model":"gpt-6-sol"}}` + "\n"
	suffix := `{"timestamp":"2026-10-25T20:00:02Z","type":"token_usage_record","payload":{"response_id":"resp-offset","usage":{"input_tokens":20,"cached_input_tokens":0,"output_tokens":5,"total_tokens":25}}}` + "\n"
	data := prefix + suffix
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil { t.Fatal(err) }
	chunk, err := ReadChunk(TranscriptFile{Path:path,Project:"varnqor",Session:"sess-offset",Size:int64(len(data))}, int64(len(prefix))); if err != nil { t.Fatal(err) }
	if len(chunk.Usage) != 1 || chunk.Usage[0].Call.Model != "gpt-6-sol" { t.Fatalf("model context was not restored: %+v", chunk.Usage) }
}
