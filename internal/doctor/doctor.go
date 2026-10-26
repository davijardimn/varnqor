package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davijardimn/varnqor/internal/domain"
	"github.com/davijardimn/varnqor/internal/providers/claude"
	"github.com/davijardimn/varnqor/internal/providers/codex"
	"github.com/davijardimn/varnqor/internal/providers/gemini"
	"github.com/davijardimn/varnqor/internal/settings"
	"github.com/davijardimn/varnqor/internal/storage"
)

func Run(store *storage.Store, storeErr error, version string) domain.DoctorReport {
	report := domain.DoctorReport{Status: "healthy", Version: version, Checks: []domain.DoctorCheck{}}
	c, x, g := claude.Detect(), codex.Detect(), gemini.Detect()
	detected := 0
	if c.Detected {
		detected++
		detail := fmt.Sprintf("%d sessions observed", c.Sessions)
		if c.DataDetected { report.Checks = append(report.Checks, check("provider-claude", "Claude Code", "pass", detail)) } else { report.Checks = append(report.Checks, check("provider-claude", "Claude Code", "pass", "binary available; no local sessions yet")) }
	}
	if x.Detected {
		detected++
		detail := fmt.Sprintf("%d JSONL sessions observed", x.Sessions)
		if x.UnsupportedFiles > 0 { detail += fmt.Sprintf("; %d compressed session files not indexed", x.UnsupportedFiles) }
		report.Checks = append(report.Checks, check("provider-codex", "Codex", "pass", detail))
	}
	if g.Detected {
		detected++
		report.Checks = append(report.Checks, check("provider-gemini", "Gemini CLI", "pass", fmt.Sprintf("%d local chat sessions observed", g.Sessions)))
	}
	if detected == 0 { report.Checks = append(report.Checks, check("providers", "AI agent providers", "warn", "no supported local provider was detected")) } else { report.Checks = append(report.Checks, check("providers", "AI agent providers", "pass", fmt.Sprintf("%d supported provider(s) detected", detected))) }
	if storeErr != nil { report.Checks = append(report.Checks, check("sqlite", "Local SQLite store", "fail", storeErr.Error())) } else if store == nil { report.Checks = append(report.Checks, check("sqlite", "Local SQLite store", "fail", "store is not initialized")) } else { report.Checks = append(report.Checks, check("sqlite", "Local SQLite store", "pass", store.Path())) }
	if _, err := settings.LoadDefault(); err != nil { report.Checks = append(report.Checks, check("settings", "Settings file", "fail", err.Error())) } else { path, _ := settings.DefaultPath(); report.Checks = append(report.Checks, check("settings", "Settings file", "pass", path)) }
	if err := configWriteCheck(); err != nil { report.Checks = append(report.Checks, check("config-write", "Config directory writable", "fail", err.Error())) } else { report.Checks = append(report.Checks, check("config-write", "Config directory writable", "pass", "temporary write succeeded")) }
	if store != nil && storeErr == nil { if err := store.PrepareV1(); err != nil { report.Checks = append(report.Checks, check("data-model", "Resource data model", "fail", err.Error())) } else { report.Checks = append(report.Checks, check("data-model", "Resource data model", "pass", "resource index hygiene is current")) } }
	return Summarize(report)
}

func Summarize(report domain.DoctorReport) domain.DoctorReport { report.Pass, report.Warn, report.Fail = 0, 0, 0; for _, item := range report.Checks { switch item.Status { case "fail": report.Fail++; case "warn": report.Warn++; default: report.Pass++ } }; switch { case report.Fail > 0: report.Status = "error"; case report.Warn > 0: report.Status = "warning"; default: report.Status = "healthy" }; return report }
func check(id, label, status, detail string) domain.DoctorCheck { return domain.DoctorCheck{ID:id, Label:label, Status:status, Detail:detail} }
func configWriteCheck() error { path, err := settings.DefaultPath(); if err != nil { return err }; dir := filepath.Dir(path); if err := os.MkdirAll(dir, 0o700); err != nil { return err }; file, err := os.CreateTemp(dir, ".varnqor-doctor-*"); if err != nil { return err }; name := file.Name(); if err := file.Close(); err != nil { _ = os.Remove(name); return err }; return os.Remove(name) }
