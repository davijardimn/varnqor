package analytics

import (
	"math"
	"testing"
)

func TestComputeBurnRateUsesMedianActiveBaseline(t *testing.T) {
	result := ComputeBurnRate(1500, 1000, []int64{800, 1000, 1200, 10000})

	if !result.Ready {
		t.Fatal("expected burn rate to be ready")
	}
	if result.BaselinePerHour != 1100 {
		t.Fatalf("expected median baseline 1100, got %d", result.BaselinePerHour)
	}
	if math.Abs(result.Ratio-1.36) > 0.001 {
		t.Fatalf("expected ratio 1.36, got %.2f", result.Ratio)
	}
	if result.Level != "elevated" {
		t.Fatalf("expected elevated level, got %s", result.Level)
	}
	if result.Trend != "accelerating" || math.Abs(result.ChangePercent-50) > 0.001 {
		t.Fatalf("unexpected trend: %+v", result)
	}
	if result.Confidence != "medium" {
		t.Fatalf("expected medium confidence, got %s", result.Confidence)
	}
}

func TestComputeBurnRateIgnoresIdleHours(t *testing.T) {
	result := ComputeBurnRate(900, 900, []int64{0, 0, 800, 1000})
	if !result.Ready {
		t.Fatal("expected two active samples to be enough")
	}
	if result.BaselinePerHour != 900 || result.Ratio != 1 {
		t.Fatalf("unexpected baseline result: %+v", result)
	}
	if result.Level != "normal" || result.Trend != "steady" {
		t.Fatalf("unexpected classification: %+v", result)
	}
	if result.Confidence != "low" {
		t.Fatalf("expected low confidence, got %s", result.Confidence)
	}
}

func TestComputeBurnRateLearnsBeforeClaimingPressure(t *testing.T) {
	result := ComputeBurnRate(4000, 0, []int64{2000})
	if result.Ready {
		t.Fatal("expected burn rate to remain in learning state")
	}
	if result.Level != "learning" || result.Confidence != "insufficient" {
		t.Fatalf("unexpected learning state: %+v", result)
	}
	if result.Trend != "new-spike" {
		t.Fatalf("expected new-spike trend, got %s", result.Trend)
	}
}
