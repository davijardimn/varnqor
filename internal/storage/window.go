package storage

import (
	"time"

	"github.com/davijardimn/varnqor/internal/domain"
)

func (s *Store) UsageWindow(provider string, start, end time.Time) (domain.TokenTotals, int, error) {
	var tokens domain.TokenTotals
	var calls int

	err := s.db.QueryRow(`SELECT
		COUNT(*),
		COALESCE(SUM(input_tokens), 0),
		COALESCE(SUM(output_tokens), 0),
		COALESCE(SUM(cache_creation_tokens), 0),
		COALESCE(SUM(cache_read_tokens), 0),
		COALESCE(SUM(cache_creation_5m_tokens), 0),
		COALESCE(SUM(cache_creation_1h_tokens), 0),
		COALESCE(SUM(total_tokens), 0)
	FROM usage_calls
	WHERE provider = ? AND timestamp_ms >= ? AND timestamp_ms < ?`,
		provider, start.UnixMilli(), end.UnixMilli()).Scan(
		&calls,
		&tokens.Input,
		&tokens.Output,
		&tokens.CacheCreation,
		&tokens.CacheRead,
		&tokens.CacheCreation5m,
		&tokens.CacheCreation1h,
		&tokens.Total,
	)
	return tokens, calls, err
}
