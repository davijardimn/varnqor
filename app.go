package main

import (
	"context"
	"fmt"
	"time"

	"github.com/davijardimn/varnqor/internal/analytics"
	"github.com/davijardimn/varnqor/internal/doctor"
	"github.com/davijardimn/varnqor/internal/domain"
	"github.com/davijardimn/varnqor/internal/indexer"
	"github.com/davijardimn/varnqor/internal/providers/claude"
	"github.com/davijardimn/varnqor/internal/settings"
	"github.com/davijardimn/varnqor/internal/storage"
)

type App struct {
	ctx      context.Context
	store    *storage.Store
	storeErr error
}

type SystemStatus struct {
	Version       string                   `json:"version"`
	GeneratedAt   time.Time                `json:"generatedAt"`
	Claude        claude.Detection         `json:"claude"`
	Usage         domain.UsageOverview     `json:"usage"`
	Hourly        []domain.HourUsage       `json:"hourly"`
	Burn          analytics.BurnRate       `json:"burn"`
	Forecast      domain.ForecastReport    `json:"forecast"`
	Quota         domain.QuotaStatus       `json:"quota"`
	Attribution   domain.AttributionReport `json:"attribution"`
	Waste         domain.WasteReport       `json:"waste"`
	Advisor       domain.AdvisorReport     `json:"advisor"`
	Sync          domain.SyncStats         `json:"sync"`
	StorageReady  bool                     `json:"storageReady"`
	UsageError    string                   `json:"usageError,omitempty"`
	SettingsError string                   `json:"settingsError,omitempty"`
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.store, a.storeErr = storage.OpenDefault()
}

func (a *App) shutdown(context.Context) {
	if a.store != nil {
		_ = a.store.Close()
	}
}

func (a *App) GetSettings() (settings.Settings, error) {
	return settings.LoadDefault()
}

func (a *App) SaveSettings(value settings.Settings) (settings.Settings, error) {
	return settings.SaveDefault(value)
}

func (a *App) RunDoctor() domain.DoctorReport {
	return doctor.Run(a.store, a.storeErr, Version)
}

func (a *App) GetSystemStatus() SystemStatus {
	cfg, cfgErr := settings.LoadDefault()
	return collectSystemStatus(a.store, a.storeErr, cfg, cfgErr, time.Now())
}

func collectSystemStatus(store *storage.Store, storeErr error, cfg settings.Settings, cfgErr error, now time.Time) SystemStatus {
	detection := claude.Detect()
	status := SystemStatus{
		Version:      Version,
		GeneratedAt:  now,
		Claude:       detection,
		StorageReady: store != nil && storeErr == nil,
	}
	if cfgErr != nil {
		status.SettingsError = cfgErr.Error()
		cfg = settings.Default()
	}
	if storeErr != nil {
		status.UsageError = fmt.Sprintf("open local usage store: %v", storeErr)
		return status
	}
	if store == nil {
		status.UsageError = "local usage store is not initialized"
		return status
	}
	if !detection.DataDetected {
		applyConfiguredQuota(&status, store, cfg, now)
		return status
	}

	syncStats, err := indexer.SyncClaude(store, detection.ProjectsPath)
	if err != nil {
		status.UsageError = err.Error()
		return status
	}
	status.Sync = syncStats

	usage, hourly, err := store.DailyUsage("claude", now)
	if err != nil {
		status.UsageError = fmt.Sprintf("query local usage: %v", err)
		return status
	}
	files, warnings, err := store.SourceSummary("claude")
	if err != nil {
		status.UsageError = fmt.Sprintf("query source summary: %v", err)
		return status
	}
	usage.FilesScanned, usage.ParseWarnings = files, warnings
	status.Usage, status.Hourly = usage, hourly

	currentStart := now.Add(-time.Hour)
	current, _, err := store.UsageWindow("claude", currentStart, now)
	if err != nil {
		status.UsageError = fmt.Sprintf("query current burn window: %v", err)
		return status
	}
	previous, _, err := store.UsageWindow("claude", now.Add(-2*time.Hour), currentStart)
	if err != nil {
		status.UsageError = fmt.Sprintf("query previous burn window: %v", err)
		return status
	}

	localNow := now.In(now.Location())
	currentHour := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), localNow.Hour(), 0, 0, 0, localNow.Location())
	baselineSamples := make([]int64, 0, len(hourly))
	for _, bucket := range hourly {
		if bucket.Start.Before(currentHour) && bucket.Tokens.Total > 0 {
			baselineSamples = append(baselineSamples, bucket.Tokens.Total)
		}
	}
	status.Burn = analytics.ComputeBurnRate(current.Total, previous.Total, baselineSamples)
	status.Forecast = analytics.ComputeForecast(status.Burn, 2*time.Hour)

	attribution, err := store.Attribution("claude", currentStart, now, 5)
	if err != nil {
		status.UsageError = fmt.Sprintf("query burn attribution: %v", err)
		return status
	}
	status.Attribution = attribution

	toolCalls, err := store.ToolCalls("claude", currentStart, now)
	if err != nil {
		status.UsageError = fmt.Sprintf("query tool metadata: %v", err)
		return status
	}
	status.Waste = analytics.AnalyzeWaste(toolCalls, currentStart, now)
	applyConfiguredQuota(&status, store, cfg, now)
	if status.Quota.Available {
		status.Forecast.RunwayAvailable = status.Quota.RunwayAvailable
		status.Forecast.LimitSource = status.Quota.Source
	}
	status.Advisor = analytics.BuildAdvisor(now, status.Burn, status.Forecast, status.Attribution, status.Waste)
	return status
}

func applyConfiguredQuota(status *SystemStatus, store *storage.Store, cfg settings.Settings, now time.Time) {
	if !cfg.Budget.Enabled || cfg.Budget.LimitTokens <= 0 || store == nil {
		return
	}
	start, resetAt := analytics.ConfiguredBudgetWindow(now, cfg.Budget.ResetHour)
	used, _, err := store.UsageWindow("claude", start, now)
	if err != nil {
		if status.UsageError == "" {
			status.UsageError = fmt.Sprintf("query configured budget window: %v", err)
		}
		return
	}
	status.Quota = analytics.ComputeConfiguredQuota(now, start, resetAt, cfg.Budget.Label, cfg.Budget.LimitTokens, used.Total, status.Burn.CurrentPerHour)
}
