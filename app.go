package main

import (
	"context"

	"github.com/davijardimn/varnqor/internal/providers/claude"
)

type App struct {
	ctx context.Context
}

type SystemStatus struct {
	Claude claude.Detection `json:"claude"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetSystemStatus() SystemStatus {
	return SystemStatus{
		Claude: claude.Detect(),
	}
}
