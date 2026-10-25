package claude

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Detection struct {
	Detected       bool      `json:"detected"`
	BinaryDetected bool      `json:"binaryDetected"`
	DataDetected   bool      `json:"dataDetected"`
	ClaudeHome     string    `json:"claudeHome"`
	ProjectsPath   string    `json:"projectsPath"`
	Projects       int       `json:"projects"`
	Sessions       int       `json:"sessions"`
	LastActivity   time.Time `json:"lastActivity,omitempty"`
}

func Detect() Detection {
	home, err := os.UserHomeDir()
	if err != nil {
		return Detection{}
	}

	return DetectAt(home, exec.LookPath)
}

func DetectAt(home string, lookup func(string) (string, error)) Detection {
	claudeHome := filepath.Join(home, ".claude")
	projectsPath := filepath.Join(claudeHome, "projects")

	result := Detection{
		ClaudeHome:   claudeHome,
		ProjectsPath: projectsPath,
	}

	if lookup != nil {
		if _, err := lookup("claude"); err == nil {
			result.BinaryDetected = true
		}
	}

	if info, err := os.Stat(claudeHome); err == nil && info.IsDir() {
		result.DataDetected = true
	}

	projectEntries, err := os.ReadDir(projectsPath)
	if err == nil {
		for _, project := range projectEntries {
			if !project.IsDir() {
				continue
			}

			result.Projects++

			entries, err := os.ReadDir(filepath.Join(projectsPath, project.Name()))
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".jsonl") {
					continue
				}

				result.Sessions++

				info, err := entry.Info()
				if err == nil && info.ModTime().After(result.LastActivity) {
					result.LastActivity = info.ModTime()
				}
			}
		}
	}

	result.Detected = result.BinaryDetected || result.DataDetected
	return result
}
