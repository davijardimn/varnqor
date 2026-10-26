package analytics

import (
	"testing"

	"github.com/davijardimn/varnqor/internal/domain"
)

func TestBuildPulseSignals(t *testing.T){
	report:=domain.ExploreReport{
		Current:domain.WindowSummary{Calls:20,Sessions:4,Tokens:domain.TokenTotals{Input:7000,CacheRead:1000,Output:2000,Total:10000},CacheReadShare:12.5},
		Previous:domain.WindowSummary{Tokens:domain.TokenTotals{Total:4000}},
		Comparison:domain.CompareReport{TokensChangePercent:150},
		Projects:[]domain.DimensionActivity{{Provider:"claude",Name:"alpha",SharePercent:80}},
		Sessions:[]domain.SessionActivity{{Provider:"claude",Project:"alpha",Session:"s1",SharePercent:70}},
	}
	policy:=PulsePolicy{Enabled:true,SpikePercent:75,ProjectConcentrationPercent:70,SessionConcentrationPercent:60,CacheReadMinimumPercent:15,MinimumTokens:1000}
	items:=BuildPulse(report,policy)
	if len(items)<4{t.Fatalf("expected at least four signals, got %+v",items)}
	seen:=map[string]bool{};for _,item:=range items{seen[item.ID]=true}
	for _,id:=range []string{"token-spike","project-concentration","session-concentration","cache-read-low"}{if !seen[id]{t.Fatalf("missing signal %s: %+v",id,items)}}
}

func TestBuildPulseHonorsMinimumVolume(t *testing.T){
	report:=domain.ExploreReport{Current:domain.WindowSummary{Tokens:domain.TokenTotals{Total:100}},Comparison:domain.CompareReport{TokensChangePercent:900}}
	items:=BuildPulse(report,PulsePolicy{Enabled:true,SpikePercent:75,MinimumTokens:1000})
	if len(items)!=0{t.Fatalf("expected no signals below minimum volume: %+v",items)}
}
