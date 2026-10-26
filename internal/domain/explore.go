package domain

import "time"

type ExploreFilter struct {
	Provider     string `json:"provider"`
	Project      string `json:"project"`
	Model        string `json:"model"`
	RangeMinutes int    `json:"rangeMinutes"`
	Limit        int    `json:"limit"`
}

type WindowSummary struct {
	Start          time.Time   `json:"start"`
	End            time.Time   `json:"end"`
	Calls          int         `json:"calls"`
	Projects       int         `json:"projects"`
	Sessions       int         `json:"sessions"`
	Tokens         TokenTotals `json:"tokens"`
	CacheReadShare float64     `json:"cacheReadShare"`
}

type DimensionActivity struct {
	Provider     string      `json:"provider,omitempty"`
	Name         string      `json:"name"`
	Calls        int         `json:"calls"`
	Sessions     int         `json:"sessions"`
	Tokens       TokenTotals `json:"tokens"`
	SharePercent float64     `json:"sharePercent"`
	LastActivity time.Time   `json:"lastActivity,omitempty"`
}

type SessionActivity struct {
	Provider      string      `json:"provider"`
	Project       string      `json:"project"`
	Session       string      `json:"session"`
	Calls         int         `json:"calls"`
	Tokens        TokenTotals `json:"tokens"`
	SharePercent  float64     `json:"sharePercent"`
	FirstActivity time.Time   `json:"firstActivity"`
	LastActivity  time.Time   `json:"lastActivity"`
	Subagent      bool        `json:"subagent"`
}

type CallActivity struct {
	Provider  string      `json:"provider"`
	Project   string      `json:"project"`
	Session   string      `json:"session"`
	Model     string      `json:"model"`
	Timestamp time.Time   `json:"timestamp"`
	Tokens    TokenTotals `json:"tokens"`
	Subagent  bool        `json:"subagent"`
}

type TimelineBucket struct {
	Start  time.Time   `json:"start"`
	Calls  int         `json:"calls"`
	Tokens TokenTotals `json:"tokens"`
}

type CompareReport struct {
	TokensChangePercent float64 `json:"tokensChangePercent"`
	CallsChangePercent  float64 `json:"callsChangePercent"`
	CacheChangePoints   float64 `json:"cacheChangePoints"`
	TokensComparable    bool    `json:"tokensComparable"`
	CallsComparable     bool    `json:"callsComparable"`
	CacheComparable     bool    `json:"cacheComparable"`
	Direction           string  `json:"direction"`
}

type ExploreFacets struct {
	Providers []string `json:"providers"`
	Projects  []string `json:"projects"`
	Models    []string `json:"models"`
}

type PulseSignal struct {
	ID        string  `json:"id"`
	Severity  string  `json:"severity"`
	Kind      string  `json:"kind"`
	Title     string  `json:"title"`
	Detail    string  `json:"detail"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	Provider  string  `json:"provider,omitempty"`
	Project   string  `json:"project,omitempty"`
	Session   string  `json:"session,omitempty"`
}

type ExploreReport struct {
	GeneratedAt   time.Time           `json:"generatedAt"`
	Filter        ExploreFilter       `json:"filter"`
	BucketMinutes int                 `json:"bucketMinutes"`
	Current       WindowSummary       `json:"current"`
	Previous      WindowSummary       `json:"previous"`
	Comparison    CompareReport       `json:"comparison"`
	Timeline      []TimelineBucket    `json:"timeline"`
	Providers     []DimensionActivity `json:"providers"`
	Projects      []DimensionActivity `json:"projects"`
	Models        []DimensionActivity `json:"models"`
	Sessions      []SessionActivity   `json:"sessions"`
	Recent        []CallActivity      `json:"recent"`
	Facets        ExploreFacets       `json:"facets"`
	Signals       []PulseSignal       `json:"signals"`
}
