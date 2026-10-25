package settings

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	want := Default()
	want.RefreshSeconds = 9
	want.Budget.Enabled = true
	want.Budget.LimitTokens = 2_000_000
	want.Budget.ResetHour = 6
	if _, err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.RefreshSeconds != 9 || !got.Budget.Enabled || got.Budget.LimitTokens != 2_000_000 || got.Budget.ResetHour != 6 {
		t.Fatalf("unexpected settings: %+v", got)
	}
}

func TestSaveRejectsEnabledZeroBudget(t *testing.T) {
	value := Default()
	value.Budget.Enabled = true
	if _, err := Save(filepath.Join(t.TempDir(), "settings.json"), value); err == nil {
		t.Fatal("expected enabled zero budget to fail validation")
	}
}

func TestNormalizeClampsRefreshAndResetHour(t *testing.T) {
	value := Normalize(Settings{RefreshSeconds: 1, Budget: ConfiguredBudget{ResetHour: 99}})
	if value.RefreshSeconds != 2 || value.Budget.ResetHour != 23 {
		t.Fatalf("unexpected normalization: %+v", value)
	}
}
