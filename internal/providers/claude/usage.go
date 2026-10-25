package claude

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TokenTotals struct {
	Input           int64 `json:"input"`
	Output          int64 `json:"output"`
	CacheCreation   int64 `json:"cacheCreation"`
	CacheRead       int64 `json:"cacheRead"`
	CacheCreation5m int64 `json:"cacheCreation5m"`
	CacheCreation1h int64 `json:"cacheCreation1h"`
	Total           int64 `json:"total"`
}

type UsageCall struct {
	MessageID  string      `json:"messageId"`
	RequestID  string      `json:"requestId"`
	Timestamp  time.Time   `json:"timestamp"`
	Model      string      `json:"model"`
	Project    string      `json:"project"`
	Session    string      `json:"session"`
	Subagent   bool        `json:"subagent"`
	StopReason string      `json:"stopReason"`
	Tokens     TokenTotals `json:"tokens"`
}

type ModelUsage struct {
	Model  string      `json:"model"`
	Calls  int         `json:"calls"`
	Tokens TokenTotals `json:"tokens"`
}

type UsageOverview struct {
	Date           string       `json:"date"`
	Calls          int          `json:"calls"`
	Sessions       int          `json:"sessions"`
	Projects       int          `json:"projects"`
	FilesScanned   int          `json:"filesScanned"`
	ParseWarnings  int          `json:"parseWarnings"`
	CacheReadShare float64      `json:"cacheReadShare"`
	FirstActivity  time.Time    `json:"firstActivity,omitempty"`
	LastActivity   time.Time    `json:"lastActivity,omitempty"`
	Tokens         TokenTotals  `json:"tokens"`
	Models         []ModelUsage `json:"models"`
}

type scanStats struct {
	FilesScanned  int
	ParseWarnings int
}

type transcriptRecord struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestId"`
	Message   struct {
		ID         string  `json:"id"`
		Model      string  `json:"model"`
		StopReason *string `json:"stop_reason"`
		Usage      *struct {
			InputTokens         int64 `json:"input_tokens"`
			OutputTokens        int64 `json:"output_tokens"`
			CacheCreationTokens int64 `json:"cache_creation_input_tokens"`
			CacheReadTokens     int64 `json:"cache_read_input_tokens"`
			CacheCreation       struct {
				Ephemeral5m int64 `json:"ephemeral_5m_input_tokens"`
				Ephemeral1h int64 `json:"ephemeral_1h_input_tokens"`
			} `json:"cache_creation"`
		} `json:"usage"`
	} `json:"message"`
}

func ScanUsageToday(projectsPath string, now time.Time) (UsageOverview, error) {
	calls, stats, err := scanUsage(projectsPath)
	if err != nil {
		return UsageOverview{}, err
	}

	return summarizeDay(calls, stats, now), nil
}

func scanUsage(projectsPath string) ([]UsageCall, scanStats, error) {
	callsByKey := make(map[string]UsageCall)
	stats := scanStats{}

	if _, err := os.Stat(projectsPath); err != nil {
		if os.IsNotExist(err) {
			return nil, stats, nil
		}
		return nil, stats, err
	}

	err := filepath.WalkDir(projectsPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".jsonl") {
			return nil
		}

		stats.FilesScanned++
		project, session, subagent := usageLocation(projectsPath, path)
		fileCalls, warnings, err := readUsageFile(path, project, session, subagent)
		stats.ParseWarnings += warnings
		if err != nil {
			return err
		}

		for index, call := range fileCalls {
			key := usageKey(call, path, index)
			if current, ok := callsByKey[key]; !ok || betterUsageCall(call, current) {
				callsByKey[key] = call
			}
		}

		return nil
	})
	if err != nil {
		return nil, stats, err
	}

	calls := make([]UsageCall, 0, len(callsByKey))
	for _, call := range callsByKey {
		calls = append(calls, call)
	}
	sort.Slice(calls, func(i, j int) bool {
		return calls[i].Timestamp.Before(calls[j].Timestamp)
	})

	return calls, stats, nil
}

func readUsageFile(path, project, session string, subagent bool) ([]UsageCall, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 128*1024)
	calls := make([]UsageCall, 0)
	warnings := 0

	for lineNumber := 1; ; lineNumber++ {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 && bytes.Contains(line, []byte(`"usage"`)) {
			call, ok, parseErr := parseUsageLine(line, project, session, subagent)
			if parseErr != nil {
				warnings++
			} else if ok {
				calls = append(calls, call)
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return nil, warnings, fmt.Errorf("read %s line %d: %w", path, lineNumber, readErr)
		}
	}

	return calls, warnings, nil
}

func parseUsageLine(line []byte, project, session string, subagent bool) (UsageCall, bool, error) {
	var record transcriptRecord
	if err := json.Unmarshal(line, &record); err != nil {
		return UsageCall{}, false, err
	}
	if record.Type != "assistant" || record.Message.Usage == nil {
		return UsageCall{}, false, nil
	}

	timestamp, err := parseTimestamp(record.Timestamp)
	if err != nil {
		return UsageCall{}, false, err
	}

	usage := record.Message.Usage
	cacheCreation := usage.CacheCreationTokens
	cacheSplit := usage.CacheCreation.Ephemeral5m + usage.CacheCreation.Ephemeral1h
	if cacheCreation == 0 && cacheSplit > 0 {
		cacheCreation = cacheSplit
	}

	tokens := TokenTotals{
		Input:           usage.InputTokens,
		Output:          usage.OutputTokens,
		CacheCreation:   cacheCreation,
		CacheRead:       usage.CacheReadTokens,
		CacheCreation5m: usage.CacheCreation.Ephemeral5m,
		CacheCreation1h: usage.CacheCreation.Ephemeral1h,
	}
	tokens.Total = tokens.Input + tokens.Output + tokens.CacheCreation + tokens.CacheRead

	stopReason := ""
	if record.Message.StopReason != nil {
		stopReason = *record.Message.StopReason
	}

	return UsageCall{
		MessageID:  record.Message.ID,
		RequestID:  record.RequestID,
		Timestamp:  timestamp,
		Model:      record.Message.Model,
		Project:    project,
		Session:    session,
		Subagent:   subagent,
		StopReason: stopReason,
		Tokens:     tokens,
	}, true, nil
}

func parseTimestamp(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("missing timestamp")
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid timestamp %q", value)
}

func usageLocation(projectsPath, path string) (project, session string, subagent bool) {
	relative, err := filepath.Rel(projectsPath, path)
	if err != nil {
		return "unknown", strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), false
	}

	parts := strings.Split(relative, string(filepath.Separator))
	if len(parts) > 0 {
		project = parts[0]
	}
	if len(parts) == 2 {
		session = strings.TrimSuffix(parts[1], filepath.Ext(parts[1]))
	} else if len(parts) >= 3 {
		session = parts[1]
	}

	for _, part := range parts {
		if part == "subagents" {
			subagent = true
			break
		}
	}
	if strings.HasPrefix(filepath.Base(path), "agent-") {
		subagent = true
	}
	if project == "" {
		project = "unknown"
	}
	if session == "" {
		session = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	return project, session, subagent
}

func usageKey(call UsageCall, path string, index int) string {
	if call.MessageID != "" {
		if call.RequestID != "" {
			return call.MessageID + ":" + call.RequestID
		}
		return call.MessageID
	}
	if call.RequestID != "" {
		return "request:" + call.RequestID
	}
	return fmt.Sprintf("%s:%d", path, index)
}

func betterUsageCall(candidate, current UsageCall) bool {
	candidateFinal := candidate.StopReason != ""
	currentFinal := current.StopReason != ""
	if candidateFinal != currentFinal {
		return candidateFinal
	}
	if candidate.Tokens.Output != current.Tokens.Output {
		return candidate.Tokens.Output > current.Tokens.Output
	}
	if candidate.Tokens.Total != current.Tokens.Total {
		return candidate.Tokens.Total > current.Tokens.Total
	}
	return candidate.Timestamp.After(current.Timestamp)
}

func summarizeDay(calls []UsageCall, stats scanStats, now time.Time) UsageOverview {
	location := now.Location()
	day := now.In(location).Format("2006-01-02")
	overview := UsageOverview{
		Date:          day,
		FilesScanned:  stats.FilesScanned,
		ParseWarnings: stats.ParseWarnings,
	}

	sessions := make(map[string]struct{})
	projects := make(map[string]struct{})
	models := make(map[string]*ModelUsage)

	for _, call := range calls {
		if call.Timestamp.IsZero() || call.Timestamp.In(location).Format("2006-01-02") != day {
			continue
		}

		overview.Calls++
		addTokens(&overview.Tokens, call.Tokens)

		if overview.FirstActivity.IsZero() || call.Timestamp.Before(overview.FirstActivity) {
			overview.FirstActivity = call.Timestamp
		}
		if overview.LastActivity.IsZero() || call.Timestamp.After(overview.LastActivity) {
			overview.LastActivity = call.Timestamp
		}

		projects[call.Project] = struct{}{}
		sessions[call.Project+":"+call.Session] = struct{}{}

		model := call.Model
		if model == "" {
			model = "unknown"
		}
		breakdown, ok := models[model]
		if !ok {
			breakdown = &ModelUsage{Model: model}
			models[model] = breakdown
		}
		breakdown.Calls++
		addTokens(&breakdown.Tokens, call.Tokens)
	}

	overview.Projects = len(projects)
	overview.Sessions = len(sessions)
	promptTokens := overview.Tokens.Input + overview.Tokens.CacheCreation + overview.Tokens.CacheRead
	if promptTokens > 0 {
		overview.CacheReadShare = (float64(overview.Tokens.CacheRead) / float64(promptTokens)) * 100
	}

	overview.Models = make([]ModelUsage, 0, len(models))
	for _, breakdown := range models {
		overview.Models = append(overview.Models, *breakdown)
	}
	sort.Slice(overview.Models, func(i, j int) bool {
		if overview.Models[i].Tokens.Total == overview.Models[j].Tokens.Total {
			return overview.Models[i].Model < overview.Models[j].Model
		}
		return overview.Models[i].Tokens.Total > overview.Models[j].Tokens.Total
	})

	return overview
}

func addTokens(target *TokenTotals, value TokenTotals) {
	target.Input += value.Input
	target.Output += value.Output
	target.CacheCreation += value.CacheCreation
	target.CacheRead += value.CacheRead
	target.CacheCreation5m += value.CacheCreation5m
	target.CacheCreation1h += value.CacheCreation1h
	target.Total += value.Total
}
