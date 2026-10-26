package analytics

import (
	"fmt"
	"math"

	"github.com/davijardimn/varnqor/internal/domain"
)

type PulsePolicy struct {
	Enabled                     bool
	SpikePercent                float64
	ProjectConcentrationPercent float64
	SessionConcentrationPercent float64
	CacheReadMinimumPercent     float64
	MinimumTokens               int64
}

func BuildPulse(report domain.ExploreReport, policy PulsePolicy) []domain.PulseSignal {
	if !policy.Enabled || report.Current.Tokens.Total < policy.MinimumTokens {
		return nil
	}
	items := make([]domain.PulseSignal, 0, 5)
	if report.Previous.Tokens.Total > 0 && report.Comparison.TokensChangePercent >= policy.SpikePercent {
		severity := "medium"
		if report.Comparison.TokensChangePercent >= policy.SpikePercent*2 { severity = "high" }
		items = append(items, domain.PulseSignal{ID:"token-spike",Severity:severity,Kind:"pace",Title:"Usage spike",Detail:fmt.Sprintf("Token activity is %.1f%% above the previous comparable window.",report.Comparison.TokensChangePercent),Value:report.Comparison.TokensChangePercent,Threshold:policy.SpikePercent})
	}
	if len(report.Projects)>0 && report.Projects[0].SharePercent >= policy.ProjectConcentrationPercent {
		top:=report.Projects[0]
		severity:="medium";if top.SharePercent>=90{severity="high"}
		items=append(items,domain.PulseSignal{ID:"project-concentration",Severity:severity,Kind:"concentration",Title:"Project concentration",Detail:fmt.Sprintf("%s accounts for %.1f%% of tokens in this window.",top.Name,top.SharePercent),Value:top.SharePercent,Threshold:policy.ProjectConcentrationPercent,Provider:top.Provider,Project:top.Name})
	}
	if len(report.Sessions)>0 && report.Sessions[0].SharePercent >= policy.SessionConcentrationPercent {
		top:=report.Sessions[0]
		severity:="medium";if top.SharePercent>=85{severity="high"}
		items=append(items,domain.PulseSignal{ID:"session-concentration",Severity:severity,Kind:"concentration",Title:"Hot session",Detail:fmt.Sprintf("One session accounts for %.1f%% of token activity in this window.",top.SharePercent),Value:top.SharePercent,Threshold:policy.SessionConcentrationPercent,Provider:top.Provider,Project:top.Project,Session:top.Session})
	}
	prompt:=report.Current.Tokens.Input+report.Current.Tokens.CacheCreation+report.Current.Tokens.CacheRead
	if report.Current.Tokens.CacheRead>0 && prompt>=policy.MinimumTokens && report.Current.CacheReadShare<policy.CacheReadMinimumPercent {
		items=append(items,domain.PulseSignal{ID:"cache-read-low",Severity:"low",Kind:"cache",Title:"Low cache reuse",Detail:fmt.Sprintf("Cache reads represent %.1f%% of prompt-side tokens in this window.",report.Current.CacheReadShare),Value:report.Current.CacheReadShare,Threshold:policy.CacheReadMinimumPercent})
	}
	if report.Current.Sessions >= 8 && report.Current.Calls > 0 {
		callsPerSession:=float64(report.Current.Calls)/float64(report.Current.Sessions)
		if callsPerSession<2 {
			items=append(items,domain.PulseSignal{ID:"session-fragmentation",Severity:"low",Kind:"fragmentation",Title:"Fragmented activity",Detail:fmt.Sprintf("%d sessions produced only %.1f token-bearing calls per session on average.",report.Current.Sessions,callsPerSession),Value:math.Round(callsPerSession*10)/10,Threshold:2})
		}
	}
	return items
}
