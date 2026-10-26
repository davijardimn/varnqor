package analytics

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

// BuildControlPlane turns local telemetry, learned burn, configured runway and
// optimization playbooks into one operational decision surface. It deliberately
// has no opaque score: every state transition is backed by an incident or a
// visible policy threshold. External execution remains unavailable.
func BuildControlPlane(
	now time.Time,
	providers []domain.ProviderStatus,
	burn BurnRate,
	quota domain.QuotaStatus,
	explore domain.ExploreReport,
	optimization domain.OptimizationReport,
	policy domain.ControlPolicySnapshot,
) domain.ControlPlaneReport {
	minutes := explore.Filter.RangeMinutes
	if minutes <= 0 { minutes = 60 }
	currentPerHour := normalizeControlRate(explore.Current.Tokens.Total, minutes)
	projected := currentPerHour * int64(policy.DecisionHorizonMinutes) / 60
	if optimization.Scenario.ProjectedTokens > 0 { projected = optimization.Scenario.ProjectedTokens }

	result := domain.ControlPlaneReport{
		GeneratedAt: now,
		State: "nominal",
		RangeMinutes: minutes,
		CurrentPerHour: currentPerHour,
		ProjectedTokens: projected,
		Policy: policy,
		Missions: []domain.ControlMission{},
		Incidents: []domain.ControlIncident{},
		Decisions: []domain.ControlDecision{},
		ExternalExecution: false,
		Note: "Control decisions are local, explainable and advisory. Varnqor does not silently execute actions against external agents.",
	}
	if !policy.Enabled {
		result.State = "disabled"
		return result
	}

	for _, provider := range providers {
		if provider.State == "connected" || provider.State == "partial" { result.ActiveProviders++ }
		switch provider.State {
		case "error":
			result.Incidents = append(result.Incidents, domain.ControlIncident{
				ID: "provider-error-"+provider.ID, Severity: "high", Kind: "provider-health",
				Title: provider.Name+" ingestion needs attention",
				Detail: firstNonEmpty(provider.Error, "Local provider data could not be indexed cleanly."), Source: "provider",
				Provider: provider.ID,
			})
			result.Decisions = append(result.Decisions, domain.ControlDecision{
				ID: "repair-provider-"+provider.ID, Priority: "high", Kind: "provider-repair",
				Title: "Repair "+provider.Name+" telemetry before trusting fleet totals",
				Reason: "Provider ingestion is in an error state, so aggregate resource decisions can be incomplete.",
				Provider: provider.ID, Confidence: "high",
				Steps: []string{"Open Agents and inspect the provider error.", "Run Doctor to verify paths and local storage.", "Refresh the Control Room after the provider indexes successfully."},
			})
		case "partial":
			result.Incidents = append(result.Incidents, domain.ControlIncident{
				ID: "provider-partial-"+provider.ID, Severity: "medium", Kind: "provider-health",
				Title: provider.Name+" is only partially observable",
				Detail: firstNonEmpty(provider.Note, "Some local provider data is unsupported or unavailable."), Source: "provider",
				Provider: provider.ID,
			})
		}
	}

	for _, session := range explore.Sessions {
		age := ageMinutes(now, session.LastActivity)
		state := missionState(age, policy.MissionActiveMinutes, policy.MissionCoolingMinutes)
		pressure := missionPressure(session.SharePercent, state, policy.SessionDominancePercent)
		mission := domain.ControlMission{
			ID: session.Provider+"|"+session.Session,
			Provider: session.Provider, Project: session.Project, Session: session.Session,
			State: state, Pressure: pressure, Calls: session.Calls, Tokens: session.Tokens,
			SharePercent: session.SharePercent, VelocityHour: normalizeControlRate(session.Tokens.Total, minutes),
			FirstActivity: session.FirstActivity, LastActivity: session.LastActivity, AgeMinutes: age, Subagent: session.Subagent,
		}
		result.Missions = append(result.Missions, mission)
		if state == "active" { result.ActiveMissions++ }
		if state == "cooling" { result.CoolingMissions++ }
	}
	sort.SliceStable(result.Missions, func(i,j int) bool {
		si,sj:=missionStateRank(result.Missions[i].State),missionStateRank(result.Missions[j].State)
		if si!=sj{return si>sj}
		if result.Missions[i].SharePercent!=result.Missions[j].SharePercent{return result.Missions[i].SharePercent>result.Missions[j].SharePercent}
		return result.Missions[i].LastActivity.After(result.Missions[j].LastActivity)
	})
	if len(result.Missions)>16 { result.Missions=result.Missions[:16] }

	if explore.Current.Tokens.Total >= policy.MinimumEvaluationTokens {
		if burn.Ready && burn.Ratio >= policy.BurnCeilingRatio {
			severity := "medium"
			if burn.Ratio >= policy.BurnCeilingRatio*1.25 || burn.Level == "critical" { severity = "high" }
			result.Incidents = append(result.Incidents, domain.ControlIncident{
				ID:"burn-ceiling",Severity:severity,Kind:"burn-ceiling",Title:"Burn pace crossed the control ceiling",
				Detail:fmt.Sprintf("Current pace is %.2fx the learned active baseline; the control ceiling is %.2fx.",burn.Ratio,policy.BurnCeilingRatio),
				Source:"burn",Value:burn.Ratio,Threshold:policy.BurnCeilingRatio,
			})
		}
		if len(explore.Projects)>0 && explore.Projects[0].SharePercent >= policy.ProjectDominancePercent {
			top:=explore.Projects[0]; severity:="medium"; if top.SharePercent>=90{severity="high"}
			result.Incidents=append(result.Incidents,domain.ControlIncident{
				ID:"project-dominance",Severity:severity,Kind:"project-dominance",Title:"One project dominates the control window",
				Detail:fmt.Sprintf("%s owns %.1f%% of selected-window tokens.",top.Name,top.SharePercent),Source:"attribution",Value:top.SharePercent,Threshold:policy.ProjectDominancePercent,Project:top.Name,Provider:top.Provider,
			})
		}
		if len(explore.Sessions)>0 && explore.Sessions[0].SharePercent >= policy.SessionDominancePercent {
			top:=explore.Sessions[0]; severity:="medium"; if top.SharePercent>=85{severity="high"}
			result.Incidents=append(result.Incidents,domain.ControlIncident{
				ID:"session-dominance",Severity:severity,Kind:"session-dominance",Title:"A single mission owns most of the burn",
				Detail:fmt.Sprintf("Session %s in %s owns %.1f%% of selected-window tokens.",shortControlID(top.Session),top.Project,top.SharePercent),Source:"attribution",Value:top.SharePercent,Threshold:policy.SessionDominancePercent,Provider:top.Provider,Project:top.Project,Session:top.Session,
			})
		}
		promptSide:=explore.Current.Tokens.Input+explore.Current.Tokens.CacheCreation+explore.Current.Tokens.CacheRead
		cacheTelemetry:=explore.Current.Tokens.CacheCreation>0||explore.Current.Tokens.CacheRead>0
		if cacheTelemetry && promptSide>0 && explore.Current.CacheReadShare < policy.CacheFloorPercent {
			severity:="low"; if explore.Current.CacheReadShare < policy.CacheFloorPercent/2 { severity="medium" }
			result.Incidents=append(result.Incidents,domain.ControlIncident{
				ID:"cache-floor",Severity:severity,Kind:"cache-floor",Title:"Reusable context is below policy floor",
				Detail:fmt.Sprintf("Cache-read share is %.1f%%; policy floor is %.1f%%.",explore.Current.CacheReadShare,policy.CacheFloorPercent),Source:"cache",Value:explore.Current.CacheReadShare,Threshold:policy.CacheFloorPercent,
			})
		}
	}

	if quota.Available && quota.UtilizationPercent >= policy.BudgetWarningPercent {
		severity:="medium"; if quota.WillExceedBeforeReset || quota.UtilizationPercent>=100 { severity="high" }
		result.Incidents=append(result.Incidents,domain.ControlIncident{
			ID:"budget-risk",Severity:severity,Kind:"budget-risk",Title:"Configured budget is entering the control zone",
			Detail:fmt.Sprintf("%.1f%% of the configured local budget is used; warning threshold is %.1f%%.",quota.UtilizationPercent,policy.BudgetWarningPercent),Source:"budget",Value:quota.UtilizationPercent,Threshold:policy.BudgetWarningPercent,
		})
		if severity=="high" {
			result.Decisions=append(result.Decisions,domain.ControlDecision{
				ID:"contain-budget-window",Priority:"high",Kind:"budget-containment",Title:"Contain the next work block before reset",
				Reason:"Projected activity can exceed the configured local budget before reset.",Confidence:quota.Confidence,
				Steps:[]string{"Finish the current atomic task before opening another workstream.","Use Optimize to select the highest-impact bounded intervention.","Re-check runway after the next refresh; the configured budget is not a provider quota."},
			})
		}
	}

	for _, action := range optimization.Actions {
		result.Decisions=append(result.Decisions,domain.ControlDecision{
			ID:"opt-"+action.ID,Priority:action.Priority,Kind:action.Kind,Title:action.Title,Reason:action.Evidence,
			Provider:action.Provider,Project:action.Project,Session:action.Session,Confidence:action.Confidence,
			EstimatedSavingsLow:action.EstimatedSavingsLow,EstimatedSavingsHigh:action.EstimatedSavingsHigh,NonAdditive:action.NonAdditive,Steps:action.Steps,
		})
	}

	sort.SliceStable(result.Incidents,func(i,j int)bool{
		pi,pj:=controlPriorityRank(result.Incidents[i].Severity),controlPriorityRank(result.Incidents[j].Severity)
		if pi!=pj{return pi>pj}; return result.Incidents[i].Title<result.Incidents[j].Title
	})
	sort.SliceStable(result.Decisions,func(i,j int)bool{
		pi,pj:=controlPriorityRank(result.Decisions[i].Priority),controlPriorityRank(result.Decisions[j].Priority)
		if pi!=pj{return pi>pj}; return result.Decisions[i].EstimatedSavingsHigh>result.Decisions[j].EstimatedSavingsHigh
	})
	if len(result.Incidents)>10 { result.Incidents=result.Incidents[:10] }
	if len(result.Decisions)>8 { result.Decisions=result.Decisions[:8] }
	for _,incident:=range result.Incidents{switch incident.Severity{case"high":result.HighIncidents++;case"medium":result.MediumIncidents++;default:result.LowIncidents++}}

	// Health and trusted-budget failures always outrank a quiet token window.
	// Otherwise a broken provider could incorrectly make the Control Room look idle.
	if hasCriticalControlIncident(result.Incidents,quota) {
		result.State="critical"
	} else if result.HighIncidents>0 || hasDecisionPriority(result.Decisions,"high") {
		result.State="intervene"
	} else if result.MediumIncidents>0 || hasDecisionPriority(result.Decisions,"medium") {
		result.State="watch"
	} else if explore.Current.Tokens.Total==0 && result.ActiveProviders==0 {
		result.State="idle"
	}
	return result
}

func normalizeControlRate(tokens int64, minutes int) int64 { if tokens<=0||minutes<=0{return 0}; return int64(math.Round(float64(tokens)*60/float64(minutes))) }
func ageMinutes(now,last time.Time) int { if last.IsZero(){return 1<<30}; d:=now.Sub(last); if d<0{return 0}; return int(math.Round(d.Minutes())) }
func missionState(age,active,cooling int)string{if age<=active{return"active"};if age<=cooling{return"cooling"};return"idle"}
func missionStateRank(value string)int{switch value{case"active":return 3;case"cooling":return 2;default:return 1}}
func missionPressure(share float64,state string,threshold float64)string{if state=="idle"{return"low"};if share>=threshold{return"high"};if share>=threshold*.72{return"elevated"};return"normal"}
func controlPriorityRank(value string)int{switch strings.ToLower(value){case"high","critical":return 3;case"medium","elevated","watch":return 2;case"low","normal":return 1;default:return 0}}
func shortControlID(value string)string{if len(value)<=10{return value};return value[:8]+"…"}
func firstNonEmpty(values ...string)string{for _,value:=range values{if strings.TrimSpace(value)!=""{return value}};return""}
func hasDecisionPriority(items []domain.ControlDecision,priority string)bool{for _,item:=range items{if item.Priority==priority{return true}};return false}
func hasCriticalControlIncident(items []domain.ControlIncident,quota domain.QuotaStatus)bool{for _,item:=range items{if item.Severity!="high"{continue};if item.Kind=="provider-health"{return true};if item.Kind=="budget-risk"&&(quota.WillExceedBeforeReset||quota.UtilizationPercent>=100){return true}};return false}
