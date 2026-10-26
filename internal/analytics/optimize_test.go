package analytics

import (
	"testing"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

func TestBuildOptimizationCreatesBoundedPlaybooks(t *testing.T) {
	now := time.Date(2026,10,26,12,0,0,0,time.UTC)
	report := domain.ExploreReport{
		Filter: domain.ExploreFilter{RangeMinutes:60},
		Current: domain.WindowSummary{Calls:18,Projects:2,Sessions:3,CacheReadShare:10,Tokens:domain.TokenTotals{Input:6000,Output:2000,CacheCreation:1000,CacheRead:1000,Total:10000}},
		Providers: []domain.DimensionActivity{{Provider:"claude",Name:"Claude Code"}},
		Projects: []domain.DimensionActivity{{Provider:"claude",Name:"alpha",SharePercent:80}},
		Sessions: []domain.SessionActivity{{Provider:"claude",Project:"alpha",Session:"session-123456789",Calls:12,SharePercent:72,Tokens:domain.TokenTotals{Total:7200}}},
	}
	burn := BurnRate{Ready:true,CurrentPerHour:10000,BaselinePerHour:5000,Ratio:2,Level:"critical",Confidence:"medium"}
	waste := domain.WasteReport{EstimatedSignalTokens:2500,Confidence:"medium"}
	policy := domain.OptimizationPolicy{Enabled:true,Goal:"balanced",TargetReductionPercent:20,HorizonMinutes:120,ExecutionMode:"preview"}
	got := BuildOptimization(now,report,burn,domain.ForecastReport{},waste,policy)
	if got.State != "opportunity" { t.Fatalf("state=%s",got.State) }
	if got.ExecutionAvailable { t.Fatal("v3 optimizer must not claim external execution") }
	if got.Scenario.ProjectedTokens != 20000 || got.Scenario.TargetSavings != 4000 || got.Scenario.TargetTokens != 16000 { t.Fatalf("scenario=%+v",got.Scenario) }
	if len(got.Actions) < 3 { t.Fatalf("expected multiple actions: %+v",got.Actions) }
	for _, action := range got.Actions {
		if !action.NonAdditive { t.Fatalf("action must be non-additive: %+v",action) }
		if action.EstimatedSavingsHigh > got.Scenario.ProjectedTokens { t.Fatalf("impact exceeds projection: %+v",action) }
	}
}

func TestBuildOptimizationIdleAndDisabledStates(t *testing.T) {
	now:=time.Now()
	idle:=BuildOptimization(now,domain.ExploreReport{Filter:domain.ExploreFilter{RangeMinutes:60}},BurnRate{},domain.ForecastReport{},domain.WasteReport{},domain.OptimizationPolicy{Enabled:true,Goal:"balanced",TargetReductionPercent:20,HorizonMinutes:120,ExecutionMode:"preview"})
	if idle.State!="idle"||len(idle.Actions)!=0{t.Fatalf("idle=%+v",idle)}
	disabled:=BuildOptimization(now,domain.ExploreReport{},BurnRate{},domain.ForecastReport{},domain.WasteReport{},domain.OptimizationPolicy{Enabled:false,Goal:"balanced",TargetReductionPercent:20,HorizonMinutes:120,ExecutionMode:"preview"})
	if disabled.State!="disabled"{t.Fatalf("disabled=%+v",disabled)}
}

func TestBuildOptimizationEconomyCanSurfaceOutputBudget(t *testing.T) {
	report:=domain.ExploreReport{Filter:domain.ExploreFilter{RangeMinutes:60},Current:domain.WindowSummary{Calls:4,Projects:1,Sessions:1,Tokens:domain.TokenTotals{Input:3000,Output:3000,Total:6000}}}
	got:=BuildOptimization(time.Now(),report,BurnRate{CurrentPerHour:6000},domain.ForecastReport{},domain.WasteReport{},domain.OptimizationPolicy{Enabled:true,Goal:"economy",TargetReductionPercent:20,HorizonMinutes:60,ExecutionMode:"preview"})
	found:=false
	for _,action:=range got.Actions{if action.Kind=="output-budget"{found=true}}
	if !found{t.Fatalf("expected economy output action: %+v",got.Actions)}
}
