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

		calls, nextOffset, warnings, err := claude.ReadUsageChunk(file, offset)
		if err != nil {
			return stats, fmt.Errorf("index %s: %w", file.Path, err)
		}
		checkpoint, err := fileCheckpoint(file.Path, nextOffset)
		if err != nil {
			return stats, fmt.Errorf("checkpoint %s: %w", file.Path, err)
		}

		records := make([]storage.UsageRecord, 0, len(calls))
		for _, indexed := range calls {
			records = append(records, storage.UsageRecord{
				Key:          indexed.Key,
				Provider:     claudeProvider,
				SourcePath:   file.Path,
				SourceOffset: indexed.SourceOffset,
				Call:         indexed.Call,
			})
		}

		if err := store.ApplyFile(storage.SourceState{
			Provider:      claudeProvider,
			Path:          file.Path,
			Size:          file.Size,
			ModTimeNS:     file.ModTime.UnixNano(),
			Offset:        nextOffset,
			Checkpoint:    checkpoint,
			ParseWarnings: state.ParseWarnings + warnings,
		}, records); err != nil {
			return stats, fmt.Errorf("persist %s: %w", file.Path, err)
		}

		stats.FilesChanged++
		stats.BytesIndexed += nextOffset - offset
		stats.RecordsObserved += len(calls)
		stats.ParseWarnings += warnings
	}

	return stats, nil
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
