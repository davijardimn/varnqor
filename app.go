package main

import (
	"context"
	"fmt"
	"time"

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
	Claude       claude.Detection     `json:"claude"`
	Usage        domain.UsageOverview `json:"usage"`
	Hourly       []domain.HourUsage   `json:"hourly"`
	Sync         domain.SyncStats     `json:"sync"`
	StorageReady bool                 `json:"storageReady"`
	UsageError   string               `json:"usageError,omitempty"`
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

	usage, hourly, err := a.store.DailyUsage("claude", time.Now())
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
	return status
}
