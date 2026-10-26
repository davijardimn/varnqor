package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

func TestExploreFiltersAndComparesWindows(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "varnqor.db"))
	if err != nil { t.Fatal(err) }
	defer store.Close()
	now := time.Date(2026,10,26,12,0,0,0,time.UTC)
	makeRecord := func(provider,key,project,session,model string, when time.Time, total int64) UsageRecord {
		return UsageRecord{Key:key,Provider:provider,SourcePath:provider+".jsonl",SourceOffset:int64(len(key)),Call:domain.UsageCall{MessageID:key,RequestID:key,Timestamp:when,Model:model,Project:project,Session:session,StopReason:"end",Tokens:domain.TokenTotals{Input:total-10,Output:10,Total:total}}}
	}
	claude := []UsageRecord{
		makeRecord("claude","old","alpha","s0","claude-sonnet",now.Add(-90*time.Minute),100),
		makeRecord("claude","c1","alpha","s1","claude-sonnet",now.Add(-30*time.Minute),100),
		makeRecord("claude","c2","alpha","s1","claude-sonnet",now.Add(-15*time.Minute),100),
	}
	if err:=store.ApplyFile(SourceState{Provider:"claude",Path:"claude.jsonl",Size:10,ModTimeNS:1,Offset:10,Checkpoint:"x"},claude);err!=nil{t.Fatal(err)}
	codex := []UsageRecord{makeRecord("codex","x1","beta","x1","gpt-codex",now.Add(-10*time.Minute),50)}
	if err:=store.ApplyFile(SourceState{Provider:"codex",Path:"codex.jsonl",Size:10,ModTimeNS:1,Offset:10,Checkpoint:"x"},codex);err!=nil{t.Fatal(err)}

	report,err:=store.Explore(domain.ExploreFilter{RangeMinutes:60,Limit:20},now);if err!=nil{t.Fatal(err)}
	if report.Current.Tokens.Total!=250||report.Previous.Tokens.Total!=100{t.Fatalf("unexpected windows: current=%d previous=%d",report.Current.Tokens.Total,report.Previous.Tokens.Total)}
	if report.Comparison.TokensChangePercent!=150{t.Fatalf("delta=%v",report.Comparison.TokensChangePercent)}
	if len(report.Timeline)!=12{t.Fatalf("timeline buckets=%d",len(report.Timeline))}
	if len(report.Projects)!=2||report.Projects[0].Name!="alpha"{t.Fatalf("projects=%+v",report.Projects)}

	filtered,err:=store.Explore(domain.ExploreFilter{Provider:"codex",RangeMinutes:60},now);if err!=nil{t.Fatal(err)}
	if filtered.Current.Tokens.Total!=50||filtered.Current.Calls!=1{t.Fatalf("filtered=%+v",filtered.Current)}
	if len(filtered.Providers)!=1||filtered.Providers[0].Provider!="codex"{t.Fatalf("providers=%+v",filtered.Providers)}
}
