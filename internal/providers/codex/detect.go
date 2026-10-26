package codex

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Detection struct {
	Detected         bool      `json:"detected"`
	BinaryDetected   bool      `json:"binaryDetected"`
	DataDetected     bool      `json:"dataDetected"`
	Home             string    `json:"home"`
	SessionsPath     string    `json:"sessionsPath"`
	ArchivedPath     string    `json:"archivedPath"`
	Sessions         int       `json:"sessions"`
	UnsupportedFiles int       `json:"unsupportedFiles"`
	LastActivity     time.Time `json:"lastActivity,omitempty"`
}

func Detect() Detection {
	home, err := os.UserHomeDir()
	if err != nil { return Detection{} }
	return DetectAt(home, exec.LookPath)
}

func DetectAt(userHome string, lookup func(string) (string, error)) Detection {
	root := filepath.Join(userHome, ".codex")
	result := Detection{Home: root, SessionsPath: filepath.Join(root, "sessions"), ArchivedPath: filepath.Join(root, "archived_sessions")}
	if lookup != nil { if _, err := lookup("codex"); err == nil { result.BinaryDetected = true } }
	for _, base := range []string{result.SessionsPath, result.ArchivedPath} {
		_ = filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() { return nil }
			name := strings.ToLower(entry.Name())
			switch {
			case strings.HasSuffix(name, ".jsonl"): result.Sessions++; result.DataDetected = true
			case strings.HasSuffix(name, ".jsonl.zst"): result.UnsupportedFiles++; result.DataDetected = true
			default: return nil
			}
			if info, infoErr := entry.Info(); infoErr == nil && info.ModTime().After(result.LastActivity) { result.LastActivity = info.ModTime() }
			return nil
		})
	}
	result.Detected = result.BinaryDetected || result.DataDetected
	return result
}
