package main

import (
	"context"
	"fmt"
	"time"

	"github.com/davijardimn/varnqor/internal/analytics"
	"github.com/davijardimn/varnqor/internal/domain"
	"github.com/davijardimn/varnqor/internal/indexer"
	"github.com/davijardimn/varnqor/internal/providers/claude"
	"github.com/davijardimn/varnqor/internal/storage"
)

type App struct {
	ctx      context.Context
	store    *storage.Store
	storeErr error
}

type SystemStatus struct {
	Claude       claude.Detection         `json:"claude"`
	Usage        domain.UsageOverview     `json:"usage"`
	Hourly       []domain.HourUsage       `json:"hourly"`
	Burn         analytics.BurnRate       `json:"burn"`
	Attribution  domain.AttributionReport `json:"attribution"`
	Waste        domain.WasteReport       `json:"waste"`
	Sync         domain.SyncStats         `json:"sync"`
	StorageReady bool                     `json:"storageReady"`
	UsageError   string                   `json:"usageError,omitempty"`
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

func (a *App) GetSystemStatus() SystemStatus {
	detection := claude.Detect()
	status := SystemStatus{Claude: detection, StorageReady: a.store != nil && a.storeErr == nil}

	if a.storeErr != nil {
		status.UsageError = fmt.Sprintf("open local usage store: %v", a.storeErr)
		return status
	}
	if a.store == nil {
		status.UsageError = "local usage store is not initialized"
		return status
	}
	if !detection.DataDetected {
		return status
	}

	syncStats, err := indexer.SyncClaude(a.store, detection.ProjectsPath)
	if err != nil {
		status.UsageError = err.Error()
		return status
	}
	status.Sync = syncStats

	now := time.Now()
	usage, hourly, err := a.store.DailyUsage("claude", now)
	if err != nil {
		status.UsageError = fmt.Sprintf("query local usage: %v", err)
		return status
	}
	files, warnings, err := a.store.SourceSummary("claude")
	if err != nil {
		status.UsageError = fmt.Sprintf("query source summary: %v", err)
		return status
	}
	usage.FilesScanned = files
	usage.ParseWarnings = warnings
	status.Usage = usage
	status.Hourly = hourly

	currentStart := now.Add(-time.Hour)
	current, _, err := a.store.UsageWindow("claude", currentStart, now)
	if err != nil {
		status.UsageError = fmt.Sprintf("query current burn window: %v", err)
		return status
	}
	previous, _, err := a.store.UsageWindow("claude", now.Add(-2*time.Hour), currentStart)
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

	attribution, err := a.store.Attribution("claude", currentStart, now, 5)
	if err != nil {
		status.UsageError = fmt.Sprintf("query burn attribution: %v", err)
		return status
	}
	status.Attribution = attribution

	toolCalls, err := a.store.ToolCalls("claude", currentStart, now)
	if err != nil {
		status.UsageError = fmt.Sprintf("query tool metadata: %v", err)
		return status
	}
	status.Waste = analytics.AnalyzeWaste(toolCalls, currentStart, now)
	return status
}
