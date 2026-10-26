package settings

import (
	"os"
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
	want.Guardrails.SpikePercent = 120
	if _, err := Save(path, want); err != nil { t.Fatal(err) }
	got, err := Load(path)
	if err != nil { t.Fatal(err) }
	if got.RefreshSeconds != 9 || !got.Budget.Enabled || got.Budget.LimitTokens != 2_000_000 || got.Budget.ResetHour != 6 || got.Guardrails.SpikePercent != 120 {
		t.Fatalf("unexpected settings: %+v", got)
	}
}

func TestSaveRejectsEnabledZeroBudget(t *testing.T) {
	value := Default()
	value.Budget.Enabled = true
	if _, err := Save(filepath.Join(t.TempDir(), "settings.json"), value); err == nil { t.Fatal("expected enabled zero budget to fail validation") }
}

func TestNormalizeClampsRefreshAndResetHour(t *testing.T) {
	value := Normalize(Settings{RefreshSeconds: 1, Budget: ConfiguredBudget{ResetHour: 99}})
	if value.RefreshSeconds != 2 || value.Budget.ResetHour != 23 { t.Fatalf("unexpected normalization: %+v", value) }
	if !value.Guardrails.Enabled || value.Guardrails.MinimumTokens != 1000 { t.Fatalf("legacy defaults not applied: %+v", value.Guardrails) }
}

func TestLoadMigratesVersionOneGuardrails(t *testing.T){
	path:=filepath.Join(t.TempDir(),"settings.json")
	data:=[]byte(`{"version":1,"refreshSeconds":5,"budget":{"enabled":false,"label":"Configured daily budget","limitTokens":0,"resetHour":0}}`)
	if err:=os.WriteFile(path,data,0o600);err!=nil{t.Fatal(err)}
	got,err:=Load(path);if err!=nil{t.Fatal(err)}
	if got.Version!=2||!got.Guardrails.Enabled||got.Guardrails.ProjectConcentrationPercent!=70{t.Fatalf("unexpected migration: %+v",got)}
}
