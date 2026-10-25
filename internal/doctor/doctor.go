package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davijardimn/varnqor/internal/domain"
	"github.com/davijardimn/varnqor/internal/providers/claude"
	"github.com/davijardimn/varnqor/internal/settings"
	"github.com/davijardimn/varnqor/internal/storage"
)

func Run(store *storage.Store, storeErr error, version string) domain.DoctorReport {
	report := domain.DoctorReport{Status: "healthy", Version: version, Checks: []domain.DoctorCheck{}}
	detection := claude.Detect()
	if detection.BinaryDetected {
		report.Checks = append(report.Checks, check("claude-binary", "Claude Code binary", "pass", "claude executable is available on PATH"))
	} else if detection.DataDetected {
		report.Checks = append(report.Checks, check("claude-binary", "Claude Code binary", "warn", "local Claude data exists, but the claude executable was not found on PATH"))
	} else {
		report.Checks = append(report.Checks, check("claude-binary", "Claude Code binary", "warn", "claude executable was not found on PATH"))
	}
	if detection.DataDetected {
		report.Checks = append(report.Checks, check("claude-data", "Claude local data", "pass", detection.ClaudeHome))
	} else {
		report.Checks = append(report.Checks, check("claude-data", "Claude local data", "warn", detection.ClaudeHome+" was not found"))
	}
	if info, err := os.Stat(detection.ProjectsPath); err == nil && info.IsDir() {
		report.Checks = append(report.Checks, check("claude-projects", "Claude projects store", "pass", fmt.Sprintf("%d project stores detected", detection.Projects)))
	} else {
		report.Checks = append(report.Checks, check("claude-projects", "Claude projects store", "warn", detection.ProjectsPath+" is not available"))
	}
	if storeErr != nil {
		report.Checks = append(report.Checks, check("sqlite", "Local SQLite store", "fail", storeErr.Error()))
	} else if store == nil {
		report.Checks = append(report.Checks, check("sqlite", "Local SQLite store", "fail", "store is not initialized"))
	} else {
		report.Checks = append(report.Checks, check("sqlite", "Local SQLite store", "pass", store.Path()))
	}
	if _, err := settings.LoadDefault(); err != nil {
		report.Checks = append(report.Checks, check("settings", "Settings file", "fail", err.Error()))
	} else {
		path, _ := settings.DefaultPath()
		report.Checks = append(report.Checks, check("settings", "Settings file", "pass", path))
	}
	if err := configWriteCheck(); err != nil {
		report.Checks = append(report.Checks, check("config-write", "Config directory writable", "fail", err.Error()))
	} else {
		report.Checks = append(report.Checks, check("config-write", "Config directory writable", "pass", "temporary write succeeded"))
	}
	return Summarize(report)
}

func Summarize(report domain.DoctorReport) domain.DoctorReport {
	report.Pass, report.Warn, report.Fail = 0, 0, 0
	for _, item := range report.Checks {
		switch item.Status {
		case "fail":
			report.Fail++
		case "warn":
			report.Warn++
		default:
			report.Pass++
		}
	}
	switch {
	case report.Fail > 0:
		report.Status = "error"
	case report.Warn > 0:
		report.Status = "warning"
	default:
		report.Status = "healthy"
	}
	return report
}

func check(id, label, status, detail string) domain.DoctorCheck {
	return domain.DoctorCheck{ID: id, Label: label, Status: status, Detail: detail}
}

func configWriteCheck() error {
	path, err := settings.DefaultPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".varnqor-doctor-*")
	if err != nil {
		return err
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	return os.Remove(name)
}
