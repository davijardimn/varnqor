package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/davijardimn/varnqor/internal/analytics"
	"github.com/davijardimn/varnqor/internal/domain"
	"github.com/davijardimn/varnqor/internal/settings"
)

func (a *App) GetExplore(filter domain.ExploreFilter) (domain.ExploreReport, error) {
	if a.storeErr != nil { return domain.ExploreReport{}, a.storeErr }
	if a.store == nil { return domain.ExploreReport{}, errors.New("local usage store is not initialized") }
	report, err := a.store.Explore(filter, time.Now())
	if err != nil { return domain.ExploreReport{}, err }
	cfg, cfgErr := settings.LoadDefault()
	if cfgErr == nil {
		g := cfg.Guardrails
		report.Signals = analytics.BuildPulse(report, analytics.PulsePolicy{Enabled:g.Enabled,SpikePercent:g.SpikePercent,ProjectConcentrationPercent:g.ProjectConcentrationPercent,SessionConcentrationPercent:g.SessionConcentrationPercent,CacheReadMinimumPercent:g.CacheReadMinimumPercent,MinimumTokens:g.MinimumTokens})
	}
	return report, nil
}

func (a *App) ExportExplore(filter domain.ExploreFilter, format string) (string, error) {
	report, err := a.GetExplore(filter)
	if err != nil { return "", err }
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil { return "", err }
		return string(data), nil
	case "csv":
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		_ = writer.Write([]string{"timestamp","provider","project","session","model","input_tokens","output_tokens","cache_creation_tokens","cache_read_tokens","total_tokens","subagent"})
		for _, item := range report.Recent {
			_ = writer.Write([]string{item.Timestamp.Format(time.RFC3339),item.Provider,item.Project,item.Session,item.Model,strconv.FormatInt(item.Tokens.Input,10),strconv.FormatInt(item.Tokens.Output,10),strconv.FormatInt(item.Tokens.CacheCreation,10),strconv.FormatInt(item.Tokens.CacheRead,10),strconv.FormatInt(item.Tokens.Total,10),strconv.FormatBool(item.Subagent)})
		}
		writer.Flush()
		if err := writer.Error(); err != nil { return "", fmt.Errorf("encode CSV export: %w", err) }
		return buf.String(), nil
	default:
		return "", fmt.Errorf("unsupported export format %q", format)
	}
}
