package analytics

import (
	"path"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

const (
	repeatedReadWindow          = 15 * time.Minute
	repeatedSearchWindow        = 10 * time.Minute
	repeatedFailureWindow       = 10 * time.Minute
	oversizedResponseThreshold  = int64(32 * 1024)
)

type eventState struct {
	Last       time.Time
	Generation int
}

type repeatedAccumulator struct {
	Project       string
	Session       string
	ToolName      string
	FilePath      string
	ReadOffset    int64
	ReadLimit     int64
	Occurrences   int
	Repeats       int
	RepeatedBytes int64
}

// AnalyzeWaste runs independent, explainable rules against tool metadata from
// one analysis window. EstimatedSignalTokens is intentionally not described as
// "wasted tokens": rules can overlap and all token conversions are heuristic.
func AnalyzeWaste(calls []domain.ToolCallObservation, start, end time.Time) domain.WasteReport {
	report := newWasteReport(calls, start, end)
	if len(calls) == 0 {
		return report
	}

	ordered := append([]domain.ToolCallObservation(nil), calls...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Timestamp.Before(ordered[j].Timestamp)
	})

	readFindings, readSignals, readTokens := repeatedFileReadFindings(ordered)
	searchFindings, searchSignals, searchTokens := repeatedSearchFindings(ordered)
	oversizedFindings, oversizedSignals, oversizedTokens := oversizedResponseFindings(ordered)
	failureFindings, failureSignals, failureTokens := repeatedFailureFindings(ordered)

	report.RepeatedReads = readSignals
	report.RepeatedSearches = searchSignals
	report.OversizedResponses = oversizedSignals
	report.RepeatedFailures = failureSignals
	// Kept for compatibility with v0.1 consumers. It contains only repeated-read estimates.
	report.EstimatedRepeatedTokens = readTokens
	report.EstimatedSignalTokens = readTokens + searchTokens + oversizedTokens + failureTokens
	report.Findings = append(report.Findings, readFindings...)
	report.Findings = append(report.Findings, searchFindings...)
	report.Findings = append(report.Findings, oversizedFindings...)
	report.Findings = append(report.Findings, failureFindings...)
	report.Confidence = findingsConfidence(report.Findings)
	report.Rules = []domain.WasteRuleSummary{
		ruleSummary("repeated-file-read", "Repeated reads", readFindings, readSignals),
		ruleSummary("repeated-search", "Repeated search", searchFindings, searchSignals),
		ruleSummary("oversized-tool-response", "Large responses", oversizedFindings, oversizedSignals),
		ruleSummary("repeated-failed-operation", "Repeated failures", failureFindings, failureSignals),
	}

	sort.SliceStable(report.Findings, func(i, j int) bool {
		if report.Findings[i].EstimatedTokens != report.Findings[j].EstimatedTokens {
			return report.Findings[i].EstimatedTokens > report.Findings[j].EstimatedTokens
		}
		if severityRank(report.Findings[i].Severity) != severityRank(report.Findings[j].Severity) {
			return severityRank(report.Findings[i].Severity) > severityRank(report.Findings[j].Severity)
		}
		return report.Findings[i].Title < report.Findings[j].Title
	})
	return report
}

// AnalyzeRepeatedFileReads preserves the v0.1 API for tests and downstream
// callers that only want the first rule.
func AnalyzeRepeatedFileReads(calls []domain.ToolCallObservation, start, end time.Time) domain.WasteReport {
	report := newWasteReport(calls, start, end)
	ordered := append([]domain.ToolCallObservation(nil), calls...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Timestamp.Before(ordered[j].Timestamp) })
	findings, signals, tokens := repeatedFileReadFindings(ordered)
	report.RepeatedReads = signals
	report.EstimatedRepeatedTokens = tokens
	report.EstimatedSignalTokens = tokens
	report.Findings = findings
	report.Confidence = findingsConfidence(findings)
	report.Rules = []domain.WasteRuleSummary{ruleSummary("repeated-file-read", "Repeated reads", findings, signals)}
	return report
}

func newWasteReport(calls []domain.ToolCallObservation, start, end time.Time) domain.WasteReport {
	return domain.WasteReport{
		Start:      start,
		End:        end,
		ToolCalls:  len(calls),
		Confidence: "insufficient",
		Rules:      []domain.WasteRuleSummary{},
		Findings:   []domain.WasteFinding{},
	}
}

func repeatedFileReadFindings(calls []domain.ToolCallObservation) ([]domain.WasteFinding, int, int64) {
	generations := make(map[string]int)
	states := make(map[string]eventState)
	accumulators := make(map[string]*repeatedAccumulator)

	for _, call := range calls {
		filePath := normalizeFilePath(call.FilePath)
		if filePath == "" || filePath == "." {
			continue
		}
		scope := transcriptScope(call)
		fileScope := scope + "\x1f" + filePath
		if isFileMutation(call.Name) {
			generations[fileScope]++
			continue
		}
		if !strings.EqualFold(call.Name, "Read") || call.IsError {
			continue
		}

		key := fileScope + "\x1f" + int64Key(call.ReadOffset) + "\x1f" + int64Key(call.ReadLimit)
		acc := accumulators[key]
		if acc == nil {
			acc = &repeatedAccumulator{Project: call.Project, Session: call.Session, ToolName: call.Name, FilePath: call.FilePath, ReadOffset: call.ReadOffset, ReadLimit: call.ReadLimit}
			accumulators[key] = acc
		}
		acc.Occurrences++
		generation := generations[fileScope]
		if previous, exists := states[key]; exists && previous.Generation == generation {
			delta := call.Timestamp.Sub(previous.Last)
			if delta >= 0 && delta <= repeatedReadWindow {
				acc.Repeats++
				acc.RepeatedBytes += positive(call.ResultBytes)
			}
		}
		states[key] = eventState{Last: call.Timestamp, Generation: generation}
	}

	findings := make([]domain.WasteFinding, 0)
	totalSignals, totalTokens := 0, int64(0)
	for _, acc := range accumulators {
		if acc.Repeats == 0 {
			continue
		}
		estimated := bytesToEstimatedTokens(acc.RepeatedBytes)
		confidence := repeatConfidence(acc.Repeats, acc.RepeatedBytes)
		findings = append(findings, domain.WasteFinding{
			RuleID: "repeated-file-read", Title: "Repeated file read", Severity: wasteSeverity(estimated), Confidence: confidence,
			Project: acc.Project, Session: acc.Session, ToolName: acc.ToolName, FilePath: acc.FilePath, Target: acc.FilePath,
			ReadOffset: acc.ReadOffset, ReadLimit: acc.ReadLimit, Reads: acc.Occurrences, Occurrences: acc.Occurrences,
			Repeats: acc.Repeats, RepeatedBytes: acc.RepeatedBytes, EstimatedTokens: estimated,
			Explanation: "The same Read target and range was requested again within 15 minutes without an observed write/edit to that file in the same transcript.",
		})
		totalSignals += acc.Repeats
		totalTokens += estimated
	}
	return findings, totalSignals, totalTokens
}

func repeatedSearchFindings(calls []domain.ToolCallObservation) ([]domain.WasteFinding, int, int64) {
	states := make(map[string]time.Time)
	accumulators := make(map[string]*repeatedAccumulator)
	for _, call := range calls {
		if !isSearchTool(call.Name) || call.InputFingerprint == "" {
			continue
		}
		key := transcriptScope(call) + "\x1f" + strings.ToLower(call.Name) + "\x1f" + call.InputFingerprint
		if call.IsError {
			delete(states, key)
			continue
		}
		acc := accumulators[key]
		if acc == nil {
			acc = &repeatedAccumulator{Project: call.Project, Session: call.Session, ToolName: call.Name, FilePath: call.FilePath}
			accumulators[key] = acc
		}
		acc.Occurrences++
		if previous, exists := states[key]; exists {
			delta := call.Timestamp.Sub(previous)
			if delta >= 0 && delta <= repeatedSearchWindow {
				acc.Repeats++
				acc.RepeatedBytes += positive(call.ResultBytes)
			}
		}
		states[key] = call.Timestamp
	}

	findings := make([]domain.WasteFinding, 0)
	totalSignals, totalTokens := 0, int64(0)
	for _, acc := range accumulators {
		if acc.Repeats == 0 {
			continue
		}
		estimated := bytesToEstimatedTokens(acc.RepeatedBytes)
		target := acc.FilePath
		if target == "" {
			target = "workspace scope"
		}
		findings = append(findings, domain.WasteFinding{
			RuleID: "repeated-search", Title: "Repeated search", Severity: wasteSeverity(estimated), Confidence: repeatConfidence(acc.Repeats, acc.RepeatedBytes),
			Project: acc.Project, Session: acc.Session, ToolName: acc.ToolName, FilePath: acc.FilePath, Target: target,
			Occurrences: acc.Occurrences, Repeats: acc.Repeats, RepeatedBytes: acc.RepeatedBytes, EstimatedTokens: estimated,
			Explanation: "The same Grep/Glob input fingerprint was executed again within 10 minutes in the same transcript. The raw query is not stored.",
		})
		totalSignals += acc.Repeats
		totalTokens += estimated
	}
	return findings, totalSignals, totalTokens
}

func oversizedResponseFindings(calls []domain.ToolCallObservation) ([]domain.WasteFinding, int, int64) {
	findings := make([]domain.WasteFinding, 0)
	totalTokens := int64(0)
	for _, call := range calls {
		if call.IsError || call.ResultBytes <= oversizedResponseThreshold {
			continue
		}
		excessBytes := call.ResultBytes - oversizedResponseThreshold
		estimated := bytesToEstimatedTokens(excessBytes)
		confidence := "low"
		if call.ResultBytes >= 128*1024 {
			confidence = "medium"
		}
		target := call.FilePath
		if target == "" {
			target = call.Name
		}
		findings = append(findings, domain.WasteFinding{
			RuleID: "oversized-tool-response", Title: "Large tool response", Severity: wasteSeverity(estimated), Confidence: confidence,
			Project: call.Project, Session: call.Session, ToolName: call.Name, FilePath: call.FilePath, Target: target,
			Occurrences: 1, ResultBytes: call.ResultBytes, EstimatedTokens: estimated,
			Explanation: "This successful tool result exceeded 32 KiB. The estimate counts only context above that threshold; a large result can still be fully legitimate.",
		})
		totalTokens += estimated
	}
	return findings, len(findings), totalTokens
}

func repeatedFailureFindings(calls []domain.ToolCallObservation) ([]domain.WasteFinding, int, int64) {
	states := make(map[string]time.Time)
	accumulators := make(map[string]*repeatedAccumulator)
	for _, call := range calls {
		if call.InputFingerprint == "" {
			continue
		}
		key := transcriptScope(call) + "\x1f" + strings.ToLower(call.Name) + "\x1f" + call.InputFingerprint
		if !call.IsError {
			delete(states, key)
			continue
		}
		acc := accumulators[key]
		if acc == nil {
			acc = &repeatedAccumulator{Project: call.Project, Session: call.Session, ToolName: call.Name, FilePath: call.FilePath}
			accumulators[key] = acc
		}
		acc.Occurrences++
		if previous, exists := states[key]; exists {
			delta := call.Timestamp.Sub(previous)
			if delta >= 0 && delta <= repeatedFailureWindow {
				acc.Repeats++
				acc.RepeatedBytes += positive(call.ResultBytes)
			}
		}
		states[key] = call.Timestamp
	}

	findings := make([]domain.WasteFinding, 0)
	totalSignals, totalTokens := 0, int64(0)
	for _, acc := range accumulators {
		if acc.Repeats == 0 {
			continue
		}
		estimated := bytesToEstimatedTokens(acc.RepeatedBytes)
		target := acc.FilePath
		if target == "" {
			target = acc.ToolName
		}
		findings = append(findings, domain.WasteFinding{
			RuleID: "repeated-failed-operation", Title: "Repeated failed operation", Severity: wasteSeverity(estimated), Confidence: repeatConfidence(acc.Repeats, acc.RepeatedBytes),
			Project: acc.Project, Session: acc.Session, ToolName: acc.ToolName, FilePath: acc.FilePath, Target: target,
			Occurrences: acc.Occurrences, Failures: acc.Occurrences, Repeats: acc.Repeats, RepeatedBytes: acc.RepeatedBytes, EstimatedTokens: estimated,
			Explanation: "The same tool and input fingerprint failed again within 10 minutes in the same transcript. A successful identical call resets the sequence.",
		})
		totalSignals += acc.Repeats
		totalTokens += estimated
	}
	return findings, totalSignals, totalTokens
}

func ruleSummary(ruleID, label string, findings []domain.WasteFinding, signals int) domain.WasteRuleSummary {
	summary := domain.WasteRuleSummary{RuleID: ruleID, Label: label, Findings: len(findings), Signals: signals}
	for _, finding := range findings {
		summary.EstimatedTokens += finding.EstimatedTokens
	}
	return summary
}

func findingsConfidence(findings []domain.WasteFinding) string {
	confidence := "insufficient"
	for _, finding := range findings {
		if finding.Confidence == "medium" {
			return "medium"
		}
		if finding.Confidence == "low" {
			confidence = "low"
		}
	}
	return confidence
}

func repeatConfidence(repeats int, resultBytes int64) string {
	if repeats >= 2 && resultBytes > 0 {
		return "medium"
	}
	return "low"
}

func transcriptScope(call domain.ToolCallObservation) string {
	if call.SourcePath != "" {
		return call.SourcePath
	}
	return call.Project + "\x1f" + call.Session + "\x1f" + boolKey(call.Subagent)
}

func normalizeFilePath(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	if value == "" {
		return ""
	}
	value = path.Clean(value)
	if runtime.GOOS == "windows" {
		value = strings.ToLower(value)
	}
	return value
}

func isFileMutation(name string) bool {
	switch strings.ToLower(name) {
	case "edit", "write", "multiedit", "notebookedit":
		return true
	default:
		return false
	}
}

func isSearchTool(name string) bool {
	switch strings.ToLower(name) {
	case "grep", "glob":
		return true
	default:
		return false
	}
}

func bytesToEstimatedTokens(value int64) int64 {
	if value <= 0 {
		return 0
	}
	return (value + 3) / 4
}

func wasteSeverity(tokens int64) string {
	switch {
	case tokens >= 25_000:
		return "high"
	case tokens >= 5_000:
		return "medium"
	default:
		return "low"
	}
}

func severityRank(value string) int {
	switch value {
	case "high":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}

func positive(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func int64Key(value int64) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var buf [20]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func boolKey(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
