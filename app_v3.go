package main

import (
	"errors"
	"time"

	"github.com/davijardimn/varnqor/internal/analytics"
	"github.com/davijardimn/varnqor/internal/domain"
	"github.com/davijardimn/varnqor/internal/settings"
)

func (a *App) GetOptimization(filter domain.ExploreFilter) (domain.OptimizationReport, error) {
	if a.storeErr != nil { return domain.OptimizationReport{}, a.storeErr }
	if a.store == nil { return domain.OptimizationReport{}, errors.New("local usage store is not initialized") }
	now := time.Now()
	report, err := a.GetExplore(filter)
	if err != nil { return domain.OptimizationReport{}, err }
	cfg, cfgErr := settings.LoadDefault()
	if cfgErr != nil { cfg = settings.Default() }

	minutes := report.Filter.RangeMinutes
	if minutes <= 0 { minutes = 60 }
	currentPerHour := normalizeWindowRate(report.Current.Tokens.Total, minutes)
	previousPerHour := normalizeWindowRate(report.Previous.Tokens.Total, minutes)
	burn := analytics.ComputeBurnRate(currentPerHour, previousPerHour, nil)

	// A global baseline is valid only for an unscoped aggregate report. Using it
	// for one provider/project/session/model would imply a precision we do not have.
	if report.Filter.Provider == "" && report.Filter.Project == "" && report.Filter.Session == "" && report.Filter.Model == "" {
		_, hourly, usageErr := a.store.DailyUsageAll(now)
		if usageErr == nil {
			localNow := now.In(now.Location())
			currentHour := time.Date(localNow.Year(),localNow.Month(),localNow.Day(),localNow.Hour(),0,0,0,localNow.Location())
			baseline := make([]int64,0,len(hourly))
			for _, bucket := range hourly {
				if bucket.Start.Before(currentHour) && bucket.Tokens.Total > 0 { baseline = append(baseline,bucket.Tokens.Total) }
			}
			burn = analytics.ComputeBurnRate(currentPerHour,previousPerHour,baseline)
		}
	}

	horizon := cfg.Optimization.HorizonMinutes
	if horizon <= 0 { horizon = 120 }
	forecast := analytics.ComputeForecast(burn,time.Duration(horizon)*time.Minute)
	waste := domain.WasteReport{}
	if (report.Filter.Provider == "" || report.Filter.Provider == "claude") && report.Filter.Project == "" && report.Filter.Session == "" && report.Filter.Model == "" {
		if toolCalls, toolErr := a.store.ToolCalls("claude",now.Add(-time.Hour),now); toolErr == nil {
			waste = analytics.AnalyzeWaste(toolCalls,now.Add(-time.Hour),now)
		}
	}
	policy := domain.OptimizationPolicy{
		Enabled: cfg.Optimization.Enabled,
		Goal: cfg.Optimization.Goal,
		TargetReductionPercent: cfg.Optimization.TargetReductionPercent,
		HorizonMinutes: cfg.Optimization.HorizonMinutes,
		ExecutionMode: cfg.Optimization.ExecutionMode,
	}
	return analytics.BuildOptimization(now,report,burn,forecast,waste,policy),nil
}

func normalizeWindowRate(tokens int64, minutes int) int64 {
	if tokens <= 0 || minutes <= 0 { return 0 }
	return tokens * 60 / int64(minutes)
}
