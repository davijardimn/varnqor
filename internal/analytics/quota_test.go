package analytics

import (
	"math"
	"testing"
	"time"
)

func TestConfiguredBudgetWindowUsesMostRecentReset(t *testing.T) {
	loc := time.FixedZone("test", -3*60*60)
	now := time.Date(2026, 10, 25, 5, 0, 0, 0, loc)
	start, end := ConfiguredBudgetWindow(now, 6)
	if start.Day() != 24 || start.Hour() != 6 || end.Day() != 25 || end.Hour() != 6 {
		t.Fatalf("unexpected window: %v -> %v", start, end)
	}
}

func TestComputeConfiguredQuotaProducesRunway(t *testing.T) {
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	start := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	reset := start.Add(24 * time.Hour)
	result := ComputeConfiguredQuota(now, start, reset, "Daily", 1_000_000, 600_000, 200_000)
	if !result.Available || result.Remaining != 400_000 || !result.RunwayAvailable {
		t.Fatalf("unexpected quota: %+v", result)
	}
	if math.Abs(result.RunwayHours-2) > 0.001 {
		t.Fatalf("expected 2h runway, got %.2f", result.RunwayHours)
	}
	if !result.WillExceedBeforeReset {
		t.Fatal("expected current pace to exceed budget before reset")
	}
}

func TestComputeConfiguredQuotaDoesNotInventRunwayWhenIdle(t *testing.T) {
	now := time.Now()
	result := ComputeConfiguredQuota(now, now.Add(-time.Hour), now.Add(time.Hour), "Daily", 1000, 100, 0)
	if result.RunwayAvailable {
		t.Fatalf("runway should be unavailable with zero pace: %+v", result)
	}
}
