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
	want.Language = "pt-BR"
	want.Theme = "light"
	want.Budget.Enabled = true
	want.Budget.LimitTokens = 2_000_000
	want.Budget.ResetHour = 6
	want.Guardrails.SpikePercent = 120
	want.Optimization.Goal = "economy"
	want.Optimization.TargetReductionPercent = 30
	want.Optimization.HorizonMinutes = 360
	want.Optimization.ExecutionMode = "manual"
	if _, err := Save(path, want); err != nil { t.Fatal(err) }
	got, err := Load(path)
	if err != nil { t.Fatal(err) }
	if got.RefreshSeconds != 9 || got.Language != "pt-BR" || got.Theme != "light" || !got.Budget.Enabled || got.Budget.LimitTokens != 2_000_000 || got.Budget.ResetHour != 6 || got.Guardrails.SpikePercent != 120 || got.Optimization.Goal != "economy" || got.Optimization.TargetReductionPercent != 30 || got.Optimization.HorizonMinutes != 360 || got.Optimization.ExecutionMode != "manual" {
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
	if !value.Guardrails.Enabled || value.Guardrails.MinimumTokens != 1000 { t.Fatalf("legacy guardrail defaults not applied: %+v", value.Guardrails) }
	if value.Language != "system" || value.Theme != "system" { t.Fatalf("v3 appearance defaults not applied: %+v", value) }
	if !value.Optimization.Enabled || value.Optimization.Goal != "balanced" || value.Optimization.TargetReductionPercent != 20 { t.Fatalf("v3 optimizer defaults not applied: %+v", value.Optimization) }
}

func TestLoadMigratesVersionOneThroughV3(t *testing.T){
	path:=filepath.Join(t.TempDir(),"settings.json")
	data:=[]byte(`{"version":1,"refreshSeconds":5,"budget":{"enabled":false,"label":"Configured daily budget","limitTokens":0,"resetHour":0}}`)
	if err:=os.WriteFile(path,data,0o600);err!=nil{t.Fatal(err)}
	got,err:=Load(path);if err!=nil{t.Fatal(err)}
	if got.Version!=3||!got.Guardrails.Enabled||got.Guardrails.ProjectConcentrationPercent!=70||got.Language!="system"||got.Theme!="system"||got.Optimization.Goal!="balanced"{t.Fatalf("unexpected migration: %+v",got)}
}

func TestLoadMigratesVersionTwoAppearanceAndOptimizer(t *testing.T){
	path:=filepath.Join(t.TempDir(),"settings.json")
	data:=[]byte(`{"version":2,"refreshSeconds":5,"budget":{"enabled":false,"label":"Configured daily budget","limitTokens":0,"resetHour":0},"guardrails":{"enabled":true,"spikePercent":90,"projectConcentrationPercent":75,"sessionConcentrationPercent":65,"cacheReadMinimumPercent":10,"minimumTokens":2000}}`)
	if err:=os.WriteFile(path,data,0o600);err!=nil{t.Fatal(err)}
	got,err:=Load(path);if err!=nil{t.Fatal(err)}
	if got.Version!=3||got.Guardrails.SpikePercent!=90||got.Language!="system"||got.Theme!="system"||got.Optimization.HorizonMinutes!=120{t.Fatalf("unexpected v2 migration: %+v",got)}
}

func TestValidateRejectsUnknownLocaleThemeAndOptimization(t *testing.T){
	value:=Default();value.Language="xx";if Validate(value)==nil{t.Fatal("expected unknown language rejection")}
	value=Default();value.Theme="sepia";if Validate(value)==nil{t.Fatal("expected unknown theme rejection")}
	value=Default();value.Optimization.Goal="magic";if Validate(value)==nil{t.Fatal("expected unknown optimization goal rejection")}
	value=Default();value.Optimization.TargetReductionPercent=90;if Validate(value)==nil{t.Fatal("expected aggressive reduction rejection")}
}
