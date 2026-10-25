package domain

import "time"

// ToolCallObservation is the privacy-minimised metadata Varnqor needs for
// tool-level analysis. SourcePath identifies the transcript internally and is
// never serialized to the desktop UI. InputFingerprint is a SHA-256 digest of
// canonical tool input; the original tool input is not retained.
type ToolCallObservation struct {
	ToolUseID        string    `json:"toolUseId"`
	Timestamp        time.Time `json:"timestamp"`
	Name             string    `json:"name"`
	Project          string    `json:"project"`
	Session          string    `json:"session"`
	Subagent         bool      `json:"subagent"`
	FilePath         string    `json:"filePath,omitempty"`
	ReadOffset       int64     `json:"readOffset,omitempty"`
	ReadLimit        int64     `json:"readLimit,omitempty"`
	InputFingerprint string    `json:"-"`
	ResultBytes      int64     `json:"resultBytes,omitempty"`
	IsError          bool      `json:"isError"`
	SourcePath       string    `json:"-"`
}

type WasteFinding struct {
	RuleID          string `json:"ruleId"`
	Title           string `json:"title"`
	Severity        string `json:"severity"`
	Confidence      string `json:"confidence"`
	Project         string `json:"project"`
	Session         string `json:"session"`
	ToolName        string `json:"toolName,omitempty"`
	FilePath        string `json:"filePath,omitempty"`
	Target          string `json:"target,omitempty"`
	ReadOffset      int64  `json:"readOffset,omitempty"`
	ReadLimit       int64  `json:"readLimit,omitempty"`
	Reads           int    `json:"reads,omitempty"`
	Occurrences     int    `json:"occurrences,omitempty"`
	Repeats         int    `json:"repeats,omitempty"`
	Failures        int    `json:"failures,omitempty"`
	RepeatedBytes   int64  `json:"repeatedBytes,omitempty"`
	ResultBytes     int64  `json:"resultBytes,omitempty"`
	EstimatedTokens int64  `json:"estimatedTokens"`
	Explanation     string `json:"explanation"`
}

type WasteRuleSummary struct {
	RuleID          string `json:"ruleId"`
	Label           string `json:"label"`
	Findings        int    `json:"findings"`
	Signals         int    `json:"signals"`
	EstimatedTokens int64  `json:"estimatedTokens"`
}

type WasteReport struct {
	Start                   time.Time          `json:"start"`
	End                     time.Time          `json:"end"`
	ToolCalls               int                `json:"toolCalls"`
	RepeatedReads           int                `json:"repeatedReads"`
	RepeatedSearches        int                `json:"repeatedSearches"`
	OversizedResponses      int                `json:"oversizedResponses"`
	RepeatedFailures        int                `json:"repeatedFailures"`
	EstimatedRepeatedTokens int64              `json:"estimatedRepeatedTokens"`
	EstimatedSignalTokens   int64              `json:"estimatedSignalTokens"`
	Confidence              string             `json:"confidence"`
	Rules                    []WasteRuleSummary `json:"rules"`
	Findings                 []WasteFinding     `json:"findings"`
}
