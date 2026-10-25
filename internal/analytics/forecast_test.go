package analytics

import (
	"testing"
	"time"
)

func TestComputeForecastProjectsCurrentPace(t *testing.T) {
	burn := BurnRate{
		Ready: true, CurrentPerHour: 1500, PreviousPerHour: 1000, BaselinePerHour: 1000,
		Ratio: 1.5, Trend: "accelerating", Level: "high", Confidence: "medium", BaselineSamples: 4,
	}

	result := ComputeForecast(burn, 2*time.Hour)
	if !result.Ready || result.ProjectedTokens != 3000 || result.BaselineProjection != 2000 {
		t.Fatalf("unexpected projection: %+v", result)
	}
	if result.Pressure != "high" || result.PressureRatio != 1.5 {
		t.Fatalf("unexpected pressure: %+v", result)
	}
	if result.EnvelopeLow != 2000 || result.EnvelopeHigh != 3000 {
		t.Fatalf("unexpected envelope: %+v", result)
	}
	if len(result.Points) != 3 || result.Points[0].Minutes != 30 || result.Points[0].ProjectedTokens != 750 {
		t.Fatalf("unexpected points: %+v", result.Points)
	}
	if result.RunwayAvailable || result.LimitSource != "unavailable" {
		t.Fatalf("runway must remain unavailable without a trusted limit: %+v", result)
	}
}

func TestComputeForecastProjectsWhilePressureIsLearning(t *testing.T) {
	burn := BurnRate{CurrentPerHour: 600, PreviousPerHour: 400, Trend: "accelerating", Level: "learning", Confidence: "insufficient"}
	result := ComputeForecast(burn, 2*time.Hour)
	if !result.Ready || result.ProjectedTokens != 1200 {
		t.Fatalf("expected observed persistence forecast: %+v", result)
	}
	if result.Pressure != "learning" || result.Confidence != "low" {
		t.Fatalf("expected learning/low confidence: %+v", result)
	}
}

func TestComputeForecastHandlesIdleCurrentPace(t *testing.T) {
	burn := BurnRate{Ready: true, CurrentPerHour: 0, PreviousPerHour: 900, BaselinePerHour: 1000, Ratio: 0, Trend: "cooling", Level: "low", Confidence: "medium"}
	result := ComputeForecast(burn, 2*time.Hour)
	if result.ProjectedTokens != 0 || result.Pressure != "low" {
		t.Fatalf("unexpected idle forecast: %+v", result)
	}
	if result.EnvelopeLow != 0 || result.EnvelopeHigh != 2000 {
		t.Fatalf("expected observed envelope to preserve idle current pace: %+v", result)
	}
}
