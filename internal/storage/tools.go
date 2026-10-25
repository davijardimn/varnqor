package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

const toolIndexVersion = "1"

type ToolUseRecord struct {
	Provider     string
	ToolUseID    string
	Timestamp    time.Time
	ToolName     string
	Project      string
	Session      string
	Subagent     bool
	FilePath     string
	ReadOffset   int64
	ReadLimit    int64
	SourcePath   string
	SourceOffset int64
}

type ToolResultRecord struct {
	Provider     string
	ToolUseID    string
	ResultBytes  int64
	IsError      bool
	SourcePath   string
	SourceOffset int64
}

// EnsureToolIndex creates the privacy-minimised tool-call index. When the tool
// index format changes, existing transcript offsets are rewound so the normal
// incremental scanner can backfill metadata without duplicating usage calls.
func (s *Store) EnsureToolIndex(provider string) (bool, error) {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS tool_calls (
			provider TEXT NOT NULL,
			tool_use_id TEXT NOT NULL,
			timestamp_ms INTEGER NOT NULL,
			tool_name TEXT NOT NULL,
			project TEXT NOT NULL,
			session TEXT NOT NULL,
			subagent INTEGER NOT NULL,
			file_path TEXT NOT NULL,
			read_offset INTEGER NOT NULL DEFAULT 0,
			read_limit INTEGER NOT NULL DEFAULT 0,
			result_bytes INTEGER NOT NULL DEFAULT 0,
			is_error INTEGER NOT NULL DEFAULT 0,
			source_path TEXT NOT NULL,
			source_offset INTEGER NOT NULL,
			PRIMARY KEY (provider, tool_use_id)
		)`,
		`CREATE INDEX IF NOT EXISTS tool_calls_provider_time
			ON tool_calls(provider, timestamp_ms)`,
		`CREATE INDEX IF NOT EXISTS tool_calls_provider_session
			ON tool_calls(provider, project, session, timestamp_ms)`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return false, fmt.Errorf("create tool index: %w", err)
		}
	}

	key := "tool_index_version:" + provider
	var version string
	err := s.db.QueryRow(`SELECT value FROM metadata WHERE key = ?`, key).Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("read tool index version: %w", err)
	}
	if version == toolIndexVersion {
		return false, nil
	}

	var sources int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM source_files WHERE provider = ?`, provider).Scan(&sources); err != nil {
		return false, fmt.Errorf("count indexed sources: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM tool_calls WHERE provider = ?`, provider); err != nil {
		return false, fmt.Errorf("clear tool index: %w", err)
	}
	if _, err := tx.Exec(`UPDATE source_files
		SET size = 0, mod_time_ns = 0, offset = 0, checkpoint = '', parse_warnings = 0
		WHERE provider = ?`, provider); err != nil {
		return false, fmt.Errorf("rewind source offsets: %w", err)
	}
	if _, err := tx.Exec(`INSERT INTO metadata(key, value) VALUES(?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, toolIndexVersion); err != nil {
		return false, fmt.Errorf("write tool index version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return sources > 0, nil
}

func (s *Store) ApplyToolEvents(uses []ToolUseRecord, results []ToolResultRecord) error {
	if len(uses) == 0 && len(results) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const upsertUse = `INSERT INTO tool_calls (
		provider, tool_use_id, timestamp_ms, tool_name, project, session, subagent,
		file_path, read_offset, read_limit, result_bytes, is_error, source_path, source_offset
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?)
	ON CONFLICT(provider, tool_use_id) DO UPDATE SET
		timestamp_ms = excluded.timestamp_ms,
		tool_name = excluded.tool_name,
		project = excluded.project,
		session = excluded.session,
		subagent = excluded.subagent,
		file_path = excluded.file_path,
		read_offset = excluded.read_offset,
		read_limit = excluded.read_limit,
		source_path = excluded.source_path,
		source_offset = excluded.source_offset`

	for _, record := range uses {
		if record.Provider == "" || record.ToolUseID == "" {
			continue
		}
		if _, err := tx.Exec(upsertUse,
			record.Provider, record.ToolUseID, record.Timestamp.UnixMilli(), record.ToolName,
			record.Project, record.Session, boolInt(record.Subagent), record.FilePath,
			record.ReadOffset, record.ReadLimit, record.SourcePath, record.SourceOffset,
		); err != nil {
			return fmt.Errorf("upsert tool use: %w", err)
		}
	}

	for _, result := range results {
		if result.Provider == "" || result.ToolUseID == "" {
			continue
		}
		if _, err := tx.Exec(`UPDATE tool_calls
			SET result_bytes = ?, is_error = ?
			WHERE provider = ? AND tool_use_id = ?`,
			result.ResultBytes, boolInt(result.IsError), result.Provider, result.ToolUseID,
		); err != nil {
			return fmt.Errorf("attach tool result metadata: %w", err)
		}
	}

	return tx.Commit()
}

func (s *Store) ResetToolProvider(provider string) error {
	_, err := s.db.Exec(`DELETE FROM tool_calls WHERE provider = ?`, provider)
	return err
}

func (s *Store) ToolCalls(provider string, start, end time.Time) ([]domain.ToolCallObservation, error) {
	rows, err := s.db.Query(`SELECT
		tool_use_id, timestamp_ms, tool_name, project, session, subagent,
		file_path, read_offset, read_limit, result_bytes, is_error, source_path
	FROM tool_calls
	WHERE provider = ? AND timestamp_ms >= ? AND timestamp_ms < ?
	ORDER BY timestamp_ms ASC, source_offset ASC`, provider, start.UnixMilli(), end.UnixMilli())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	calls := make([]domain.ToolCallObservation, 0)
	for rows.Next() {
		var call domain.ToolCallObservation
		var timestampMS int64
		var subagent, isError int
		if err := rows.Scan(
			&call.ToolUseID, &timestampMS, &call.Name, &call.Project, &call.Session, &subagent,
			&call.FilePath, &call.ReadOffset, &call.ReadLimit, &call.ResultBytes, &isError, &call.SourcePath,
		); err != nil {
			return nil, err
		}
		call.Timestamp = time.UnixMilli(timestampMS)
		call.Subagent = subagent != 0
		call.IsError = isError != 0
		calls = append(calls, call)
	}
	return calls, rows.Err()
}
