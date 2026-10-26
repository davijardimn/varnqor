package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davijardimn/varnqor/internal/analytics"
	"github.com/davijardimn/varnqor/internal/domain"
	"github.com/davijardimn/varnqor/internal/settings"
)

type ControlPlaneResponse struct {
	Status       SystemStatus              `json:"status"`
	Explore      domain.ExploreReport      `json:"explore"`
	Optimization domain.OptimizationReport `json:"optimization"`
	Control      domain.ControlPlaneReport `json:"control"`
}

func (a *App) GetControlPlane(filter domain.ExploreFilter) (ControlPlaneResponse, error) {
	if a.storeErr != nil { return ControlPlaneResponse{}, a.storeErr }
	if a.store == nil { return ControlPlaneResponse{}, errors.New("local usage store is not initialized") }
	now:=time.Now()
	cfg,cfgErr:=settings.LoadDefault()
	if cfgErr!=nil { cfg=settings.Default() }
	status:=collectSystemStatus(a.store,a.storeErr,cfg,cfgErr,now)
	explore,err:=a.GetExplore(filter)
	if err!=nil{return ControlPlaneResponse{},err}
	optimization,err:=a.GetOptimization(filter)
	if err!=nil{return ControlPlaneResponse{},err}

	burn:=status.Burn
	// A scoped mission/project/model cannot inherit the global learned baseline.
	if explore.Filter.Provider!=""||explore.Filter.Project!=""||explore.Filter.Session!=""||explore.Filter.Model!="" {
		minutes:=explore.Filter.RangeMinutes;if minutes<=0{minutes=60}
		burn=analytics.ComputeBurnRate(normalizeWindowRate(explore.Current.Tokens.Total,minutes),normalizeWindowRate(explore.Previous.Tokens.Total,minutes),nil)
	}
	c:=cfg.Control
	policy:=domain.ControlPolicySnapshot{
		Enabled:c.Enabled,Profile:c.Profile,BurnCeilingRatio:c.BurnCeilingRatio,
		SessionDominancePercent:c.SessionDominancePercent,ProjectDominancePercent:c.ProjectDominancePercent,
		CacheFloorPercent:c.CacheFloorPercent,BudgetWarningPercent:c.BudgetWarningPercent,
		MissionActiveMinutes:c.MissionActiveMinutes,MissionCoolingMinutes:c.MissionCoolingMinutes,
		DecisionHorizonMinutes:c.DecisionHorizonMinutes,MinimumEvaluationTokens:c.MinimumEvaluationTokens,
	}
	control:=analytics.BuildControlPlane(now,status.Providers,burn,status.Quota,explore,optimization,policy)
	return ControlPlaneResponse{Status:status,Explore:explore,Optimization:optimization,Control:control},nil
}

// ExportControlBrief creates a privacy-safe handoff that can be pasted into an
// agent or issue. It contains only resource metadata and Varnqor decisions; no
// prompt text, source code or raw tool-result bodies are reconstructed.
func (a *App) ExportControlBrief(filter domain.ExploreFilter) (string,error) {
	report,err:=a.GetControlPlane(filter);if err!=nil{return "",err}
	var b strings.Builder
	fmt.Fprintln(&b,"# Varnqor Control Brief")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b,"- State: **%s**\n",strings.ToUpper(report.Control.State))
	fmt.Fprintf(&b,"- Window: **%d min**\n",report.Control.RangeMinutes)
	fmt.Fprintf(&b,"- Current pace: **%d tokens/h**\n",report.Control.CurrentPerHour)
	fmt.Fprintf(&b,"- Decision horizon: **%d min**\n",report.Control.Policy.DecisionHorizonMinutes)
	fmt.Fprintf(&b,"- Projected activity: **%d tokens**\n",report.Control.ProjectedTokens)
	fmt.Fprintf(&b,"- Missions: **%d active / %d cooling**\n",report.Control.ActiveMissions,report.Control.CoolingMissions)
	fmt.Fprintf(&b,"- Policy: **%s**\n",strings.ToUpper(report.Control.Policy.Profile))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b,"## Incidents")
	if len(report.Control.Incidents)==0 { fmt.Fprintln(&b,"No control-policy incident is active.") }
	for _,item:=range report.Control.Incidents {
		fmt.Fprintf(&b,"- **%s · %s** — %s\n",strings.ToUpper(item.Severity),item.Title,item.Detail)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b,"## Decisions")
	if len(report.Control.Decisions)==0 { fmt.Fprintln(&b,"No intervention is currently strong enough to recommend.") }
	for i,item:=range report.Control.Decisions {
		fmt.Fprintf(&b,"### %d. %s [%s]\n",i+1,item.Title,strings.ToUpper(item.Priority))
		fmt.Fprintf(&b,"%s\n",item.Reason)
		if item.EstimatedSavingsHigh>0 { fmt.Fprintf(&b,"\nEstimated bounded impact: ~%d–%d tokens (non-additive).\n",item.EstimatedSavingsLow,item.EstimatedSavingsHigh) }
		if item.Project!=""||item.Session!=""||item.Provider!="" { fmt.Fprintf(&b,"\nScope: provider=%s project=%s session=%s\n",blankDash(item.Provider),blankDash(item.Project),blankDash(shortBriefID(item.Session))) }
		if len(item.Steps)>0 { fmt.Fprintln(&b,"\nSteps:");for _,step:=range item.Steps{fmt.Fprintf(&b,"- %s\n",step)} }
		fmt.Fprintln(&b)
	}
	fmt.Fprintln(&b,"---")
	fmt.Fprintln(&b,"This brief contains local resource metadata only. Savings ranges may overlap. Varnqor does not execute external agent actions automatically.")
	return b.String(),nil
}

func blankDash(value string)string{if strings.TrimSpace(value)==""{return "-"};return value}
func shortBriefID(value string)string{if len(value)<=16{return value};return value[:14]+"…"}
