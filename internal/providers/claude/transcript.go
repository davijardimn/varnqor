package claude

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type ToolUseEvent struct {
	ID               string
	Timestamp        string
	Name             string
	Project          string
	Session          string
	Subagent         bool
	FilePath         string
	ReadOffset       int64
	ReadLimit        int64
	InputFingerprint string
	SourceOffset     int64
}

type ToolResultEvent struct {
	ToolUseID    string
	ResultBytes  int64
	IsError      bool
	SourceOffset int64
}

type TranscriptChunk struct {
	Usage       []IndexedUsageCall
	ToolUses    []ToolUseEvent
	ToolResults []ToolResultEvent
	NextOffset  int64
	Warnings    int
}

type toolEnvelope struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type toolBlock struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
	IsError   bool            `json:"is_error"`
}

type fileToolInput struct {
	FilePath     string `json:"file_path"`
	NotebookPath string `json:"notebook_path"`
	Path         string `json:"path"`
	Offset       int64  `json:"offset"`
	Limit        int64  `json:"limit"`
}

// ReadTranscriptChunk extracts usage and privacy-minimised tool metadata in a
// single pass. Prompt text, source-code content, commands and write payloads are
// deliberately not retained.
func ReadTranscriptChunk(file TranscriptFile, offset int64) (TranscriptChunk, error) {
	stream, err := os.Open(file.Path)
	if err != nil {
		return TranscriptChunk{}, err
	}
	defer stream.Close()
	if _, err := stream.Seek(offset, io.SeekStart); err != nil {
		return TranscriptChunk{}, err
	}

	chunk := TranscriptChunk{NextOffset: offset}
	reader := bufio.NewReaderSize(stream, 128*1024)
	cursor := offset
	for {
		lineStart := cursor
		line, readErr := reader.ReadBytes('\n')
		cursor += int64(len(line))
		complete := len(line) > 0 && line[len(line)-1] == '\n'
		if complete {
			chunk.NextOffset = cursor
			lineWarning := false

			if bytes.Contains(line, []byte(`"usage"`)) {
				call, ok, parseErr := parseUsageLine(line, file.Project, file.Session, file.Subagent)
				if parseErr != nil {
					lineWarning = true
				} else if ok {
					chunk.Usage = append(chunk.Usage, IndexedUsageCall{
						Key:          usageKey(call, file.Path, lineStart),
						SourceOffset: lineStart,
						Call:         call,
					})
				}
			}

			if bytes.Contains(line, []byte(`"tool_use"`)) || bytes.Contains(line, []byte(`"tool_result"`)) {
				uses, results, parseErr := parseToolLine(line, file, lineStart)
				if parseErr != nil {
					lineWarning = true
				} else {
					chunk.ToolUses = append(chunk.ToolUses, uses...)
					chunk.ToolResults = append(chunk.ToolResults, results...)
				}
			}

			if lineWarning {
				chunk.Warnings++
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return chunk, fmt.Errorf("read %s: %w", file.Path, readErr)
		}
	}
	return chunk, nil
}

func parseToolLine(line []byte, file TranscriptFile, sourceOffset int64) ([]ToolUseEvent, []ToolResultEvent, error) {
	var envelope toolEnvelope
	if err := json.Unmarshal(line, &envelope); err != nil {
		return nil, nil, err
	}
	blocks, err := decodeToolBlocks(envelope.Message.Content)
	if err != nil {
		return nil, nil, err
	}

	uses := make([]ToolUseEvent, 0)
	results := make([]ToolResultEvent, 0)
	for _, block := range blocks {
		switch block.Type {
		case "tool_use":
			if block.ID == "" || block.Name == "" {
				continue
			}
			if _, err := parseTimestamp(envelope.Timestamp); err != nil {
				return nil, nil, err
			}
			var input fileToolInput
			if len(block.Input) > 0 && string(block.Input) != "null" {
				_ = json.Unmarshal(block.Input, &input)
			}
			filePath := input.FilePath
			if filePath == "" {
				filePath = input.NotebookPath
			}
			if filePath == "" {
				filePath = input.Path
			}
			uses = append(uses, ToolUseEvent{
				ID:               block.ID,
				Timestamp:        envelope.Timestamp,
				Name:             block.Name,
				Project:          file.Project,
				Session:          file.Session,
				Subagent:         file.Subagent,
				FilePath:         filePath,
				ReadOffset:       input.Offset,
				ReadLimit:        input.Limit,
				InputFingerprint: fingerprintToolInput(block.Input),
				SourceOffset:     sourceOffset,
			})
		case "tool_result":
			if block.ToolUseID == "" {
				continue
			}
			results = append(results, ToolResultEvent{
				ToolUseID:    block.ToolUseID,
				ResultBytes:  payloadBytes(block.Content),
				IsError:      block.IsError,
				SourceOffset: sourceOffset,
			})
		}
	}
	return uses, results, nil
}

func decodeToolBlocks(raw json.RawMessage) ([]toolBlock, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	if raw[0] == '[' {
		var blocks []toolBlock
		if err := json.Unmarshal(raw, &blocks); err != nil {
			return nil, err
		}
		return blocks, nil
	}
	if raw[0] == '{' {
		var block toolBlock
		if err := json.Unmarshal(raw, &block); err != nil {
			return nil, err
		}
		return []toolBlock{block}, nil
	}
	return nil, nil
}

func fingerprintToolInput(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return ""
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:])
}

func payloadBytes(raw json.RawMessage) int64 {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return 0
	}
	if raw[0] == '"' {
		var value string
		if json.Unmarshal(raw, &value) == nil {
			return int64(len([]byte(value)))
		}
	}
	return int64(len(raw))
}
