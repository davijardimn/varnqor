package gemini

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONLUsage(t *testing.T){dir:=t.TempDir();path:=filepath.Join(dir,"session-a.jsonl");line:=`{"id":"m1","timestamp":"2026-10-25T20:00:00Z","type":"gemini","model":"gemini-3.0-pro","tokens":{"input":100,"output":20,"cached":40,"total":120}}`+"\n";if err:=os.WriteFile(path,[]byte(line),0o600);err!=nil{t.Fatal(err)};chunk,err:=ReadChunk(TranscriptFile{Path:path,Project:"demo",Session:"s1",Size:int64(len(line))},0);if err!=nil{t.Fatal(err)};if len(chunk.Usage)!=1{t.Fatalf("usage=%d",len(chunk.Usage))};call:=chunk.Usage[0].Call;if call.Tokens.Input!=60||call.Tokens.CacheRead!=40||call.Tokens.Output!=20||call.Tokens.Total!=120{t.Fatalf("tokens=%+v",call.Tokens)}}

func TestReadLegacyUsage(t *testing.T){dir:=t.TempDir();path:=filepath.Join(dir,"session-old.json");data:=`{"sessionId":"s2","directories":["C:/work/demo"],"kind":"subagent","messages":[{"id":"m2","timestamp":"2026-10-25T20:00:00Z","type":"gemini","tokens":{"input":10,"output":5,"cached":0,"total":15}}]}`;if err:=os.WriteFile(path,[]byte(data),0o600);err!=nil{t.Fatal(err)};chunk,err:=ReadLegacy(TranscriptFile{Path:path,Project:"hash",Session:"old",LegacyJSON:true,Size:int64(len(data))});if err!=nil{t.Fatal(err)};if len(chunk.Usage)!=1||chunk.Usage[0].Call.Project!="demo"||!chunk.Usage[0].Call.Subagent{t.Fatalf("chunk=%+v",chunk)}}
