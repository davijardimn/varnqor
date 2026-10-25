package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/davijardimn/varnqor/internal/domain"
	"github.com/davijardimn/varnqor/internal/providers/claude"
	"github.com/davijardimn/varnqor/internal/storage"
)

const claudeProvider = "claude"
const checkpointBytes int64 = 512

func SyncClaude(store *storage.Store, projectsPath string) (domain.SyncStats, error) {
	files, err := claude.DiscoverTranscripts(projectsPath)
	if err != nil {
		return domain.SyncStats{}, err
	}

	stats := domain.SyncStats{FilesSeen: len(files)}
	toolBackfill, err := store.EnsureToolIndex(claudeProvider)
	if err != nil {
		return stats, fmt.Errorf("prepare Claude tool index: %w", err)
	}
	if toolBackfill {
		stats.Rebuilt = true
	}

	states, err := store.SourceStates(claudeProvider)
	if err != nil {
		return stats, err
	}

	stateByPath := make(map[string]storage.SourceState, len(states))
	fileByPath := make(map[string]claude.TranscriptFile, len(files))
	for _, state := range states {
		stateByPath[state.Path] = state
	}
	for _, file := range files {
		fileByPath[file.Path] = file
	}

	rebuild, err := indexNeedsRebuild(states, fileByPath)
	if err != nil {
		return stats, err
	}
	if rebuild {
		if err := store.ResetToolProvider(claudeProvider); err != nil {
			return stats, fmt.Errorf("reset Claude tool index: %w", err)
		}
		if err := store.ResetProvider(claudeProvider); err != nil {
			return stats, fmt.Errorf("reset Claude index: %w", err)
		}
		stats.Rebuilt = true
		stateByPath = make(map[string]storage.SourceState)
	}

	for _, file := range files {
		state, exists := stateByPath[file.Path]
		offset := int64(0)
		if exists {
			offset = state.Offset
			if state.Offset == file.Size && state.Size == file.Size && state.ModTimeNS == file.ModTime.UnixNano() {
				continue
			}
		}

		chunk, err := claude.ReadTranscriptChunk(file, offset)
		if err != nil {
			return stats, fmt.Errorf("index %s: %w", file.Path, err)
		}
		checkpoint, err := fileCheckpoint(file.Path, chunk.NextOffset)
		if err != nil {
			return stats, fmt.Errorf("checkpoint %s: %w", file.Path, err)
		}

		usageRecords := make([]storage.UsageRecord, 0, len(chunk.Usage))
		for _, indexed := range chunk.Usage {
			usageRecords = append(usageRecords, storage.UsageRecord{
				Key:          indexed.Key,
				Provider:     claudeProvider,
				SourcePath:   file.Path,
				SourceOffset: indexed.SourceOffset,
				Call:         indexed.Call,
			})
		}

		toolUses := make([]storage.ToolUseRecord, 0, len(chunk.ToolUses))
		for _, event := range chunk.ToolUses {
			timestamp, err := timeFromClaude(event.Timestamp)
			if err != nil {
				return stats, fmt.Errorf("parse tool timestamp in %s: %w", file.Path, err)
			}
			toolUses = append(toolUses, storage.ToolUseRecord{
				Provider:     claudeProvider,
				ToolUseID:    event.ID,
				Timestamp:    timestamp,
				ToolName:     event.Name,
				Project:      event.Project,
				Session:      event.Session,
				Subagent:     event.Subagent,
				FilePath:     event.FilePath,
				ReadOffset:   event.ReadOffset,
				ReadLimit:    event.ReadLimit,
				SourcePath:   file.Path,
				SourceOffset: event.SourceOffset,
			})
		}
		toolResults := make([]storage.ToolResultRecord, 0, len(chunk.ToolResults))
		for _, event := range chunk.ToolResults {
			toolResults = append(toolResults, storage.ToolResultRecord{
				Provider:     claudeProvider,
				ToolUseID:    event.ToolUseID,
				ResultBytes:  event.ResultBytes,
				IsError:      event.IsError,
				SourcePath:   file.Path,
				SourceOffset: event.SourceOffset,
			})
		}
		if err := store.ApplyToolEvents(toolUses, toolResults); err != nil {
			return stats, fmt.Errorf("persist tool metadata for %s: %w", file.Path, err)
		}

		if err := store.ApplyFile(storage.SourceState{
			Provider:      claudeProvider,
			Path:          file.Path,
			Size:          file.Size,
			ModTimeNS:     file.ModTime.UnixNano(),
			Offset:        chunk.NextOffset,
			Checkpoint:    checkpoint,
			ParseWarnings: state.ParseWarnings + chunk.Warnings,
		}, usageRecords); err != nil {
			return stats, fmt.Errorf("persist %s: %w", file.Path, err)
		}

		stats.FilesChanged++
		stats.BytesIndexed += chunk.NextOffset - offset
		stats.RecordsObserved += len(chunk.Usage)
		stats.ParseWarnings += chunk.Warnings
	}

	return stats, nil
}

func timeFromClaude(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid timestamp %q", value)
}

func indexNeedsRebuild(states []storage.SourceState, fileByPath map[string]claude.TranscriptFile) (bool, error) {
	if len(states) == 0 {
		return false, nil
	}

	for _, state := range states {
		file, exists := fileByPath[state.Path]
		if !exists {
			return true, nil
		}
		if file.Size < state.Size || file.Size < state.Offset {
			return true, nil
		}
		if state.Offset > 0 && state.Checkpoint != "" {
			checkpoint, err := fileCheckpoint(file.Path, state.Offset)
			if err != nil {
				return false, err
			}
			if checkpoint != state.Checkpoint {
				return true, nil
			}
		}
		if file.Size == state.Size && state.Offset == state.Size && file.ModTime.UnixNano() != state.ModTimeNS {
			return true, nil
		}
	}

	return false, nil
}

func fileCheckpoint(path string, offset int64) (string, error) {
	if offset <= 0 {
		return "", nil
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	start := offset - checkpointBytes
	if start < 0 {
		start = 0
	}
	length := offset - start
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return "", err
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(file, buf); err != nil {
		return "", err
	}
	digest := sha256.Sum256(buf)
	return hex.EncodeToString(digest[:]), nil
}
