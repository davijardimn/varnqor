package analytics

import (
	"path"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

const repeatedReadWindow = 15 * time.Minute

type readState struct {
	Last       time.Time
	Generation int
}

type readAccumulator struct {
	Project       string
	Session       string
	FilePath      string
	ReadOffset    int64
	ReadLimit     int64
	Reads         int
	Repeats       int
	RepeatedBytes int64
}

// AnalyzeRepeatedFileReads detects exact-range Read calls repeated within a
// short window without an observed mutation to the same file in the same
// transcript. It is deliberately conservative: the signal is evidence of
// possible repeated context, not proof that the read was unnecessary.
func AnalyzeRepeatedFileReads(calls []domain.ToolCallObservation, start, end time.Time) domain.WasteReport {
	report := domain.WasteReport{
		Start:      start,
		End:        end,
		ToolCalls:  len(calls),
		Confidence: "insufficient",
		Findings:   []domain.WasteFinding{},
	}
	if len(calls) == 0 {
		return report
	}

	ordered := append([]domain.ToolCallObservation(nil), calls...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Timestamp.Before(ordered[j].Timestamp)
	})

	generations := make(map[string]int)
	states := make(map[string]readState)
	accumulators := make(map[string]*readAccumulator)

	for _, call := range ordered {
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
		acc, ok := accumulators[key]
		if !ok {
			acc = &readAccumulator{
				Project:    call.Project,
				Session:    call.Session,
				FilePath:   call.FilePath,
				ReadOffset: call.ReadOffset,
				ReadLimit:  call.ReadLimit,
			}
			accumulators[key] = acc
		}
		acc.Reads++

		generation := generations[fileScope]
		if previous, exists := states[key]; exists && previous.Generation == generation {
			delta := call.Timestamp.Sub(previous.Last)
			if delta >= 0 && delta <= repeatedReadWindow {
				acc.Repeats++
				if call.ResultBytes > 0 {
					acc.RepeatedBytes += call.ResultBytes
				}
			}
		}
		states[key] = readState{Last: call.Timestamp, Generation: generation}
	}

	for _, acc := range accumulators {
		if acc.Repeats == 0 {
			continue
		}
		estimatedTokens := bytesToEstimatedTokens(acc.RepeatedBytes)
		confidence := "low"
		if acc.Repeats >= 2 && acc.RepeatedBytes > 0 {
			confidence = "medium"
		}
		finding := domain.WasteFinding{
			RuleID:          "repeated-file-read",
			Title:           "Repeated file read",
			Severity:        wasteSeverity(estimatedTokens),
			Confidence:      confidence,
			Project:         acc.Project,
			Session:         acc.Session,
			FilePath:        acc.FilePath,
			ReadOffset:      acc.ReadOffset,
			ReadLimit:       acc.ReadLimit,
			Reads:           acc.Reads,
			Repeats:         acc.Repeats,
			RepeatedBytes:   acc.RepeatedBytes,
			EstimatedTokens: estimatedTokens,
			Explanation:     "The same Read target and range was requested again within 15 minutes without an observed write/edit to that file in the same transcript.",
		}
		report.Findings = append(report.Findings, finding)
		report.RepeatedReads += acc.Repeats
		report.EstimatedRepeatedTokens += estimatedTokens
		if confidence == "medium" {
			report.Confidence = "medium"
		} else if report.Confidence == "insufficient" {
			report.Confidence = "low"
		}
	}

	sort.SliceStable(report.Findings, func(i, j int) bool {
		if report.Findings[i].EstimatedTokens == report.Findings[j].EstimatedTokens {
			return report.Findings[i].Repeats > report.Findings[j].Repeats
		}
		return report.Findings[i].EstimatedTokens > report.Findings[j].EstimatedTokens
	})
	return report
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
