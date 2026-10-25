package domain

import "time"

// ToolCallObservation is the privacy-minimised metadata Varnqor needs for
// tool-level analysis. SourcePath identifies the transcript internally and is
// never serialized to the desktop UI.
type ToolCallObservation struct {
	ToolUseID   string    `json:"toolUseId"`
	Timestamp   time.Time `json:"timestamp"`
	Name        string    `json:"name"`
	Project     string    `json:"project"`
	Session     string    `json:"session"`
	Subagent    bool      `json:"subagent"`
	FilePath    string    `json:"filePath,omitempty"`
	ReadOffset  int64     `json:"readOffset,omitempty"`
	ReadLimit   int64     `json:"readLimit,omitempty"`
	ResultBytes int64     `json:"resultBytes,omitempty"`
	IsError     bool      `json:"isError"`
	SourcePath  string    `json:"-"`
}

type WasteFinding struct {
	RuleID          string `json:"ruleId"`
	Title           string `json:"title"`
	Severity        string `json:"severity"`
	Confidence      string `json:"confidence"`
	Project         string `json:"project"`
	Session         string `json:"session"`
	FilePath        string `json:"filePath"`
	ReadOffset      int64  `json:"readOffset,omitempty"`
	ReadLimit       int64  `json:"readLimit,omitempty"`
	Reads           int    `json:"reads"`
	Repeats         int    `json:"repeats"`
	RepeatedBytes   int64  `json:"repeatedBytes"`
	EstimatedTokens int64  `json:"estimatedTokens"`
	Explanation     string `json:"explanation"`
}

type WasteReport struct {
	Start                   time.Time      `json:"start"`
	End                     time.Time      `json:"end"`
	ToolCalls               int            `json:"toolCalls"`
	RepeatedReads           int            `json:"repeatedReads"`
	EstimatedRepeatedTokens int64          `json:"estimatedRepeatedTokens"`
	Confidence              string         `json:"confidence"`
	Findings                []WasteFinding `json:"findings"`
}
