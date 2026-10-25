package main

import (
	"context"
	"time"

	"github.com/davijardimn/varnqor/internal/providers/claude"
)

type App struct {
	ctx context.Context
}

type SystemStatus struct {
	Claude     claude.Detection     `json:"claude"`
	Usage      claude.UsageOverview `json:"usage"`
	UsageError string               `json:"usageError,omitempty"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetSystemStatus() SystemStatus {
	detection := claude.Detect()
	status := SystemStatus{Claude: detection}

	if !detection.DataDetected {
		return status
	}

	usage, err := claude.ScanUsageToday(detection.ProjectsPath, time.Now())
	if err != nil {
		status.UsageError = err.Error()
		return status
	}
	status.Usage = usage

	return status
}
