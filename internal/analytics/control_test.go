package analytics

import (
	"testing"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

func controlPolicy() domain.ControlPolicySnapshot {
	return domain.ControlPolicySnapshot{
		Enabled:true, Profile:"balanced", BurnCeilingRatio:1.7,
		SessionDominancePercent:60, ProjectDominancePercent:70, CacheFloorPercent:15,
		BudgetWarningPercent:80, MissionActiveMinutes:12, MissionCoolingMinutes:45,
		DecisionHorizonMinutes:120, MinimumEvaluationTokens:1000,
	}
}

func TestBuildControlPlaneCreatesMissionAndDominanceIncident(t *testing.T) {
	now:=time.Date(2026,10,26,12,0,0,0,time.UTC)
	explore:=domain.ExploreReport{
		Filter:domain.ExploreFilter{RangeMinutes:60},
		Current:domain.WindowSummary{Tokens:domain.TokenTotals{Total:10_000},Calls:8,Sessions:2,Projects:1},
		Projects:[]domain.DimensionActivity{{Provider:"claude",Name:"varnqor",SharePercent:92,Tokens:domain.TokenTotals{Total:9_200}}},
		Sessions:[]domain.SessionActivity{{Provider:"claude",Project:"varnqor",Session:"session-123456789",Calls:7,Tokens:domain.TokenTotals{Total:8_000},SharePercent:80,FirstActivity:now.Add(-40*time.Minute),LastActivity:now.Add(-3*time.Minute)}},
	}
	opt:=domain.OptimizationReport{Scenario:domain.OptimizationScenario{ProjectedTokens:20_000},Actions:[]domain.OptimizationAction{{ID:"checkpoint-dominant-session",Kind:"session-checkpoint",Priority:"high",Title:"Checkpoint the dominant session",Evidence:"80% share",Confidence:"medium",EstimatedSavingsLow:500,EstimatedSavingsHigh:1500,NonAdditive:true,Steps:[]string{"checkpoint"}}}}
	burn:=ComputeBurnRate(10_000,4_000,[]int64{5_000,5_500,5_200})
	report:=BuildControlPlane(now,[]domain.ProviderStatus{{ID:"claude",Name:"Claude Code",State:"connected"}},burn,domain.QuotaStatus{},explore,opt,controlPolicy())
	if report.ActiveMissions!=1 || len(report.Missions)!=1 || report.Missions[0].State!="active" { t.Fatalf("unexpected missions: %+v",report.Missions) }
	if report.Missions[0].Pressure!="high" { t.Fatalf("expected high mission pressure, got %s",report.Missions[0].Pressure) }
	if report.HighIncidents==0 { t.Fatalf("expected high incident: %+v",report.Incidents) }
	if len(report.Decisions)==0 || report.Decisions[0].Priority!="high" { t.Fatalf("expected high decision: %+v",report.Decisions) }
	if report.State!="intervene" { t.Fatalf("expected intervene state, got %s",report.State) }
}

func TestBuildControlPlaneProviderErrorIsCritical(t *testing.T) {
	now:=time.Now()
	report:=BuildControlPlane(now,[]domain.ProviderStatus{{ID:"codex",Name:"Codex",State:"error",Error:"bad session index"}},BurnRate{},domain.QuotaStatus{},domain.ExploreReport{Filter:domain.ExploreFilter{RangeMinutes:60}},domain.OptimizationReport{},controlPolicy())
	if report.State!="critical" { t.Fatalf("expected critical, got %s",report.State) }
	if len(report.Incidents)!=1 || report.Incidents[0].Kind!="provider-health" { t.Fatalf("unexpected incidents: %+v",report.Incidents) }
	if len(report.Decisions)!=1 || report.Decisions[0].Kind!="provider-repair" { t.Fatalf("unexpected decisions: %+v",report.Decisions) }
}

func TestBuildControlPlaneBudgetRiskCreatesContainmentDecision(t *testing.T) {
	now:=time.Now()
	quota:=domain.QuotaStatus{Available:true,UtilizationPercent:91,WillExceedBeforeReset:true,Confidence:"medium"}
	explore:=domain.ExploreReport{Filter:domain.ExploreFilter{RangeMinutes:60},Current:domain.WindowSummary{Tokens:domain.TokenTotals{Total:2000}}}
	report:=BuildControlPlane(now,nil,BurnRate{},quota,explore,domain.OptimizationReport{},controlPolicy())
	if report.State!="critical" { t.Fatalf("expected critical budget state, got %s",report.State) }
	found:=false
	for _,decision:=range report.Decisions{if decision.Kind=="budget-containment"{found=true}}
	if !found { t.Fatalf("expected budget containment decision: %+v",report.Decisions) }
}
