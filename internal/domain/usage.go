package domain

import "time"

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
	Provider string      `json:"provider,omitempty"`
	Model    string      `json:"model"`
	Calls    int         `json:"calls"`
	Tokens   TokenTotals `json:"tokens"`
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

type HourUsage struct {
	Start  time.Time   `json:"start"`
	Calls  int         `json:"calls"`
	Tokens TokenTotals `json:"tokens"`
}

type SyncStats struct {
	FilesSeen       int   `json:"filesSeen"`
	FilesChanged    int   `json:"filesChanged"`
	BytesIndexed    int64 `json:"bytesIndexed"`
	RecordsObserved int   `json:"recordsObserved"`
	ParseWarnings   int   `json:"parseWarnings"`
	Rebuilt         bool  `json:"rebuilt"`
}

func AddTokens(target *TokenTotals, value TokenTotals) {
	target.Input += value.Input
	target.Output += value.Output
	target.CacheCreation += value.CacheCreation
	target.CacheRead += value.CacheRead
	target.CacheCreation5m += value.CacheCreation5m
	target.CacheCreation1h += value.CacheCreation1h
	target.Total += value.Total
}
