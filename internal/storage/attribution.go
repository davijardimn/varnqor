package storage

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

// Attribution explains a usage window without reading prompt or response text.
// Every dimension is derived from metadata already present in the local usage index.
func (s *Store) Attribution(provider string, start, end time.Time, limit int) (domain.AttributionReport, error) {
	if limit <= 0 {
		limit = 5
	}

	tokens, calls, err := s.UsageWindow(provider, start, end)
	if err != nil {
		return domain.AttributionReport{}, err
	}
	report := domain.AttributionReport{Start: start, End: end, Calls: calls, Tokens: tokens}
	from, to := start.UnixMilli(), end.UnixMilli()

	if report.Projects, err = s.projectAttribution(provider, from, to, limit, tokens.Total); err != nil {
		return domain.AttributionReport{}, fmt.Errorf("project attribution: %w", err)
	}
	if report.Sessions, err = s.sessionAttribution(provider, from, to, limit, tokens.Total); err != nil {
		return domain.AttributionReport{}, fmt.Errorf("session attribution: %w", err)
	}
	if report.Models, err = s.modelAttribution(provider, from, to, limit, tokens.Total); err != nil {
		return domain.AttributionReport{}, fmt.Errorf("model attribution: %w", err)
	}
	if report.Agents, err = s.agentAttribution(provider, from, to, tokens.Total); err != nil {
		return domain.AttributionReport{}, fmt.Errorf("agent attribution: %w", err)
	}
	return report, nil
}

func (s *Store) projectAttribution(provider string, from, to int64, limit int, total int64) ([]domain.AttributionItem, error) {
	rows, err := s.db.Query(`SELECT
		CASE WHEN project = '' THEN 'unknown' ELSE project END,
		COUNT(*),
		COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0),
		COALESCE(SUM(cache_creation_tokens), 0), COALESCE(SUM(cache_read_tokens), 0),
		COALESCE(SUM(cache_creation_5m_tokens), 0), COALESCE(SUM(cache_creation_1h_tokens), 0),
		COALESCE(SUM(total_tokens), 0)
	FROM usage_calls
	WHERE provider = ? AND timestamp_ms >= ? AND timestamp_ms < ?
	GROUP BY CASE WHEN project = '' THEN 'unknown' ELSE project END
	ORDER BY SUM(total_tokens) DESC, COUNT(*) DESC
	LIMIT ?`, provider, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.AttributionItem, 0)
	for rows.Next() {
		var item domain.AttributionItem
		if err := rows.Scan(&item.Key, &item.Calls,
			&item.Tokens.Input, &item.Tokens.Output, &item.Tokens.CacheCreation, &item.Tokens.CacheRead,
			&item.Tokens.CacheCreation5m, &item.Tokens.CacheCreation1h, &item.Tokens.Total); err != nil {
			return nil, err
		}
		item.Label = item.Key
		item.Project = item.Key
		item.SharePercent = share(item.Tokens.Total, total)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) sessionAttribution(provider string, from, to int64, limit int, total int64) ([]domain.AttributionItem, error) {
	rows, err := s.db.Query(`SELECT
		CASE WHEN project = '' THEN 'unknown' ELSE project END,
		CASE WHEN session = '' THEN 'unknown' ELSE session END,
		COUNT(*),
		COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0),
		COALESCE(SUM(cache_creation_tokens), 0), COALESCE(SUM(cache_read_tokens), 0),
		COALESCE(SUM(cache_creation_5m_tokens), 0), COALESCE(SUM(cache_creation_1h_tokens), 0),
		COALESCE(SUM(total_tokens), 0)
	FROM usage_calls
	WHERE provider = ? AND timestamp_ms >= ? AND timestamp_ms < ?
	GROUP BY project, session
	ORDER BY SUM(total_tokens) DESC, COUNT(*) DESC
	LIMIT ?`, provider, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.AttributionItem, 0)
	for rows.Next() {
		var item domain.AttributionItem
		if err := rows.Scan(&item.Project, &item.Session, &item.Calls,
			&item.Tokens.Input, &item.Tokens.Output, &item.Tokens.CacheCreation, &item.Tokens.CacheRead,
			&item.Tokens.CacheCreation5m, &item.Tokens.CacheCreation1h, &item.Tokens.Total); err != nil {
			return nil, err
		}
		item.Key = item.Project + ":" + item.Session
		item.Label = item.Session
		item.SharePercent = share(item.Tokens.Total, total)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) modelAttribution(provider string, from, to int64, limit int, total int64) ([]domain.AttributionItem, error) {
	rows, err := s.db.Query(`SELECT
		CASE WHEN model = '' THEN 'unknown' ELSE model END,
		COUNT(*),
		COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0),
		COALESCE(SUM(cache_creation_tokens), 0), COALESCE(SUM(cache_read_tokens), 0),
		COALESCE(SUM(cache_creation_5m_tokens), 0), COALESCE(SUM(cache_creation_1h_tokens), 0),
		COALESCE(SUM(total_tokens), 0)
	FROM usage_calls
	WHERE provider = ? AND timestamp_ms >= ? AND timestamp_ms < ?
	GROUP BY CASE WHEN model = '' THEN 'unknown' ELSE model END
	ORDER BY SUM(total_tokens) DESC, COUNT(*) DESC
	LIMIT ?`, provider, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSimpleAttribution(rows, total)
}

func (s *Store) agentAttribution(provider string, from, to int64, total int64) ([]domain.AttributionItem, error) {
	rows, err := s.db.Query(`SELECT
		CASE WHEN subagent = 1 THEN 'subagents' ELSE 'main' END,
		COUNT(*),
		COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0),
		COALESCE(SUM(cache_creation_tokens), 0), COALESCE(SUM(cache_read_tokens), 0),
		COALESCE(SUM(cache_creation_5m_tokens), 0), COALESCE(SUM(cache_creation_1h_tokens), 0),
		COALESCE(SUM(total_tokens), 0)
	FROM usage_calls
	WHERE provider = ? AND timestamp_ms >= ? AND timestamp_ms < ?
	GROUP BY CASE WHEN subagent = 1 THEN 'subagents' ELSE 'main' END
	ORDER BY SUM(total_tokens) DESC`, provider, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := scanSimpleAttribution(rows, total)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Key == "main" {
			items[i].Label = "Main agent"
		} else if items[i].Key == "subagents" {
			items[i].Label = "Subagents"
		}
	}
	return items, nil
}

func scanSimpleAttribution(rows *sql.Rows, total int64) ([]domain.AttributionItem, error) {
	items := make([]domain.AttributionItem, 0)
	for rows.Next() {
		var item domain.AttributionItem
		if err := rows.Scan(&item.Key, &item.Calls,
			&item.Tokens.Input, &item.Tokens.Output, &item.Tokens.CacheCreation, &item.Tokens.CacheRead,
			&item.Tokens.CacheCreation5m, &item.Tokens.CacheCreation1h, &item.Tokens.Total); err != nil {
			return nil, err
		}
		item.Label = item.Key
		item.SharePercent = share(item.Tokens.Total, total)
		items = append(items, item)
	}
	return items, rows.Err()
}

func share(value, total int64) float64 {
	if value <= 0 || total <= 0 {
		return 0
	}
	return math.Round((float64(value)/float64(total))*1000) / 10
}
