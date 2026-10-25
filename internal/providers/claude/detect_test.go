package claude

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDetectAtFindsClaudeDataAndSessions(t *testing.T) {
	home := t.TempDir()
	projectA := filepath.Join(home, ".claude", "projects", "-work-api")
	projectB := filepath.Join(home, ".claude", "projects", "-work-web")

	for _, dir := range []string{projectA, projectB} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	sessionA := filepath.Join(projectA, "a.jsonl")
	sessionB := filepath.Join(projectA, "b.jsonl")
	sessionC := filepath.Join(projectB, "c.jsonl")

	for _, path := range []string{sessionA, sessionB, sessionC} {
		if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	now := time.Now().Add(-time.Minute).Truncate(time.Second)
	if err := os.Chtimes(sessionC, now, now); err != nil {
		t.Fatal(err)
	}

	result := DetectAt(home, func(string) (string, error) {
		return "/usr/local/bin/claude", nil
	})

	if !result.Detected || !result.BinaryDetected || !result.DataDetected {
		t.Fatalf("expected Claude to be detected: %+v", result)
	}
	if result.Projects != 2 {
		t.Fatalf("expected 2 projects, got %d", result.Projects)
	}
	if result.Sessions != 3 {
		t.Fatalf("expected 3 sessions, got %d", result.Sessions)
	}
	if result.LastActivity.IsZero() {
		t.Fatal("expected last activity to be populated")
	}
}

func TestDetectAtReturnsCleanStateWhenClaudeIsMissing(t *testing.T) {
	result := DetectAt(t.TempDir(), func(string) (string, error) {
		return "", errors.New("not found")
	})

	if result.Detected {
		t.Fatalf("expected Claude not to be detected: %+v", result)
	}
	if result.Projects != 0 || result.Sessions != 0 {
		t.Fatalf("expected no projects or sessions: %+v", result)
	}
}
