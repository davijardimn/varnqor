package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
	_ "modernc.org/sqlite"
)

const schemaVersion = 1

type Store struct {
	db   *sql.DB
	path string
}

type SourceState struct {
	Provider      string
	Path          string
	Size          int64
	ModTimeNS     int64
	Offset        int64
	Checkpoint    string
	ParseWarnings int
}

type UsageRecord struct {
	Key          string
	Provider     string
	SourcePath   string
	SourceOffset int64
	Call         domain.UsageCall
}

func OpenDefault() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve config directory: %w", err)
	}

	dir := filepath.Join(configDir, "varnqor")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create Varnqor data directory: %w", err)
	}
	return Open(filepath.Join(dir, "varnqor.db"))
}

func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db, path: path}
	if err := store.configure(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Path() string { return s.path }

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) configure() error {
	for _, statement := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := s.db.Exec(statement); err != nil {
			return fmt.Errorf("configure sqlite (%s): %w", statement, err)
		}
	}
	return nil
}

func (s *Store) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS metadata (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS source_files (
			provider TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER NOT NULL,
			mod_time_ns INTEGER NOT NULL,
			offset INTEGER NOT NULL,
			checkpoint TEXT NOT NULL,
			parse_warnings INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (provider, path)
		)`,
		`CREATE TABLE IF NOT EXISTS usage_calls (
			dedupe_key TEXT NOT NULL,
			provider TEXT NOT NULL,
			message_id TEXT NOT NULL,
			request_id TEXT NOT NULL,
			timestamp_ms INTEGER NOT NULL,
			model TEXT NOT NULL,
			project TEXT NOT NULL,
			session TEXT NOT NULL,
			subagent INTEGER NOT NULL,
			stop_reason TEXT NOT NULL,
			is_final INTEGER NOT NULL,
			input_tokens INTEGER NOT NULL,
			output_tokens INTEGER NOT NULL,
			cache_creation_tokens INTEGER NOT NULL,
			cache_read_tokens INTEGER NOT NULL,
			cache_creation_5m_tokens INTEGER NOT NULL,
			cache_creation_1h_tokens INTEGER NOT NULL,
			total_tokens INTEGER NOT NULL,
			source_path TEXT NOT NULL,
			source_offset INTEGER NOT NULL,
			PRIMARY KEY (provider, dedupe_key)
		)`,
		`CREATE INDEX IF NOT EXISTS usage_calls_provider_time
			ON usage_calls(provider, timestamp_ms)`,
		`CREATE INDEX IF NOT EXISTS usage_calls_provider_project
			ON usage_calls(provider, project, session)`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return fmt.Errorf("migrate sqlite: %w", err)
		}
	}
	_, err := s.db.Exec(`INSERT INTO metadata(key, value) VALUES('schema_version', ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`, fmt.Sprint(schemaVersion))
	if err != nil {
		return fmt.Errorf("write schema version: %w", err)
	}
	return nil
}

func (s *Store) SourceStates(provider string) ([]SourceState, error) {
	rows, err := s.db.Query(`SELECT provider, path, size, mod_time_ns, offset, checkpoint, parse_warnings
		FROM source_files WHERE provider = ?`, provider)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := make([]SourceState, 0)
	for rows.Next() {
		var state SourceState
		if err := rows.Scan(&state.Provider, &state.Path, &state.Size, &state.ModTimeNS, &state.Offset, &state.Checkpoint, &state.ParseWarnings); err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	return states, rows.Err()
}

func (s *Store) ResetProvider(provider string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM usage_calls WHERE provider = ?`, provider); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM source_files WHERE provider = ?`, provider); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ApplyFile(state SourceState, records []UsageRecord) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const upsert = `INSERT INTO usage_calls (
		dedupe_key, provider, message_id, request_id, timestamp_ms, model, project, session,
		subagent, stop_reason, is_final, input_tokens, output_tokens, cache_creation_tokens,
		cache_read_tokens, cache_creation_5m_tokens, cache_creation_1h_tokens, total_tokens,
		source_path, source_offset
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(provider, dedupe_key) DO UPDATE SET
		provider=excluded.provider,
		message_id=excluded.message_id,
		request_id=excluded.request_id,
		timestamp_ms=excluded.timestamp_ms,
		model=excluded.model,
		project=excluded.project,
		session=excluded.session,
		subagent=excluded.subagent,
		stop_reason=excluded.stop_reason,
		is_final=excluded.is_final,
		input_tokens=excluded.input_tokens,
		output_tokens=excluded.output_tokens,
		cache_creation_tokens=excluded.cache_creation_tokens,
		cache_read_tokens=excluded.cache_read_tokens,
		cache_creation_5m_tokens=excluded.cache_creation_5m_tokens,
		cache_creation_1h_tokens=excluded.cache_creation_1h_tokens,
		total_tokens=excluded.total_tokens,
		source_path=excluded.source_path,
		source_offset=excluded.source_offset
	WHERE excluded.is_final > usage_calls.is_final
		OR (excluded.is_final = usage_calls.is_final AND excluded.output_tokens > usage_calls.output_tokens)
		OR (excluded.is_final = usage_calls.is_final AND excluded.output_tokens = usage_calls.output_tokens AND excluded.total_tokens > usage_calls.total_tokens)
		OR (excluded.is_final = usage_calls.is_final AND excluded.output_tokens = usage_calls.output_tokens AND excluded.total_tokens = usage_calls.total_tokens AND excluded.timestamp_ms > usage_calls.timestamp_ms)`

	for _, record := range records {
		call := record.Call
		if _, err := tx.Exec(upsert,
			record.Key, record.Provider, call.MessageID, call.RequestID, call.Timestamp.UnixMilli(), call.Model,
			call.Project, call.Session, boolInt(call.Subagent), call.StopReason, boolInt(call.StopReason != ""),
			call.Tokens.Input, call.Tokens.Output, call.Tokens.CacheCreation, call.Tokens.CacheRead,
			call.Tokens.CacheCreation5m, call.Tokens.CacheCreation1h, call.Tokens.Total,
			record.SourcePath, record.SourceOffset,
		); err != nil {
			return fmt.Errorf("upsert usage record: %w", err)
		}
	}

	_, err = tx.Exec(`INSERT INTO source_files(provider, path, size, mod_time_ns, offset, checkpoint, parse_warnings)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(provider, path) DO UPDATE SET
			size=excluded.size,
			mod_time_ns=excluded.mod_time_ns,
			offset=excluded.offset,
			checkpoint=excluded.checkpoint,
			parse_warnings=excluded.parse_warnings`,
		state.Provider, state.Path, state.Size, state.ModTimeNS, state.Offset, state.Checkpoint, state.ParseWarnings)
	if err != nil {
		return fmt.Errorf("update source state: %w", err)
	}

	return tx.Commit()
}

func (s *Store) SourceSummary(provider string) (int, int, error) {
	var files, warnings int
	err := s.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(parse_warnings), 0) FROM source_files WHERE provider = ?`, provider).Scan(&files, &warnings)
	return files, warnings, err
}

func (s *Store) DailyUsage(provider string, now time.Time) (domain.UsageOverview, []domain.HourUsage, error) {
	location := now.Location()
	localNow := now.In(location)
	start := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location)
	end := start.AddDate(0, 0, 1)
	from, to := start.UnixMilli(), end.UnixMilli()

	overview := domain.UsageOverview{Date: start.Format("2006-01-02")}
	var firstMS, lastMS sql.NullInt64
	err := s.db.QueryRow(`SELECT
		COUNT(*),
		COUNT(DISTINCT project),
		COUNT(DISTINCT project || char(31) || session),
		COALESCE(SUM(input_tokens), 0),
		COALESCE(SUM(output_tokens), 0),
		COALESCE(SUM(cache_creation_tokens), 0),
		COALESCE(SUM(cache_read_tokens), 0),
		COALESCE(SUM(cache_creation_5m_tokens), 0),
		COALESCE(SUM(cache_creation_1h_tokens), 0),
		COALESCE(SUM(total_tokens), 0),
		MIN(timestamp_ms), MAX(timestamp_ms)
	FROM usage_calls
	WHERE provider = ? AND timestamp_ms >= ? AND timestamp_ms < ?`, provider, from, to).Scan(
		&overview.Calls, &overview.Projects, &overview.Sessions,
		&overview.Tokens.Input, &overview.Tokens.Output, &overview.Tokens.CacheCreation,
		&overview.Tokens.CacheRead, &overview.Tokens.CacheCreation5m, &overview.Tokens.CacheCreation1h,
		&overview.Tokens.Total, &firstMS, &lastMS,
	)
	if err != nil {
		return domain.UsageOverview{}, nil, err
	}
	if firstMS.Valid {
		overview.FirstActivity = time.UnixMilli(firstMS.Int64)
	}
	if lastMS.Valid {
		overview.LastActivity = time.UnixMilli(lastMS.Int64)
	}
	promptTokens := overview.Tokens.Input + overview.Tokens.CacheCreation + overview.Tokens.CacheRead
	if promptTokens > 0 {
		overview.CacheReadShare = float64(overview.Tokens.CacheRead) / float64(promptTokens) * 100
	}

	models, err := s.modelUsage(provider, from, to)
	if err != nil {
		return domain.UsageOverview{}, nil, err
	}
	overview.Models = models

	hourly, err := s.hourlyUsage(provider, start, end)
	if err != nil {
		return domain.UsageOverview{}, nil, err
	}
	return overview, hourly, nil
}

func (s *Store) modelUsage(provider string, from, to int64) ([]domain.ModelUsage, error) {
	rows, err := s.db.Query(`SELECT
		CASE WHEN model = '' THEN 'unknown' ELSE model END,
		COUNT(*),
		COALESCE(SUM(input_tokens), 0),
		COALESCE(SUM(output_tokens), 0),
		COALESCE(SUM(cache_creation_tokens), 0),
		COALESCE(SUM(cache_read_tokens), 0),
		COALESCE(SUM(cache_creation_5m_tokens), 0),
		COALESCE(SUM(cache_creation_1h_tokens), 0),
		COALESCE(SUM(total_tokens), 0)
	FROM usage_calls
	WHERE provider = ? AND timestamp_ms >= ? AND timestamp_ms < ?
	GROUP BY CASE WHEN model = '' THEN 'unknown' ELSE model END
	ORDER BY SUM(total_tokens) DESC`, provider, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	models := make([]domain.ModelUsage, 0)
	for rows.Next() {
		var model domain.ModelUsage
		if err := rows.Scan(&model.Model, &model.Calls,
			&model.Tokens.Input, &model.Tokens.Output, &model.Tokens.CacheCreation,
			&model.Tokens.CacheRead, &model.Tokens.CacheCreation5m, &model.Tokens.CacheCreation1h,
			&model.Tokens.Total); err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func (s *Store) hourlyUsage(provider string, start, end time.Time) ([]domain.HourUsage, error) {
	buckets := make(map[int64]*domain.HourUsage)
	for cursor := start; cursor.Before(end); cursor = cursor.Add(time.Hour) {
		bucketStart := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), cursor.Hour(), 0, 0, 0, start.Location())
		key := bucketStart.UnixMilli()
		if _, exists := buckets[key]; !exists {
			buckets[key] = &domain.HourUsage{Start: bucketStart}
		}
	}

	rows, err := s.db.Query(`SELECT timestamp_ms, input_tokens, output_tokens, cache_creation_tokens,
		cache_read_tokens, cache_creation_5m_tokens, cache_creation_1h_tokens, total_tokens
	FROM usage_calls
	WHERE provider = ? AND timestamp_ms >= ? AND timestamp_ms < ?`, provider, start.UnixMilli(), end.UnixMilli())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var timestampMS int64
		var tokens domain.TokenTotals
		if err := rows.Scan(&timestampMS, &tokens.Input, &tokens.Output, &tokens.CacheCreation,
			&tokens.CacheRead, &tokens.CacheCreation5m, &tokens.CacheCreation1h, &tokens.Total); err != nil {
			return nil, err
		}
		local := time.UnixMilli(timestampMS).In(start.Location())
		bucketStart := time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), 0, 0, 0, start.Location())
		bucket, ok := buckets[bucketStart.UnixMilli()]
		if !ok {
			bucket = &domain.HourUsage{Start: bucketStart}
			buckets[bucketStart.UnixMilli()] = bucket
		}
		bucket.Calls++
		domain.AddTokens(&bucket.Tokens, tokens)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	hourly := make([]domain.HourUsage, 0, len(buckets))
	for _, bucket := range buckets {
		hourly = append(hourly, *bucket)
	}
	sort.Slice(hourly, func(i, j int) bool { return hourly[i].Start.Before(hourly[j].Start) })
	return hourly, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
