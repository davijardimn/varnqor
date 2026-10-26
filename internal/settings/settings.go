package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const currentVersion = 4

type ConfiguredBudget struct {
	Enabled     bool   `json:"enabled"`
	Label       string `json:"label"`
	LimitTokens int64  `json:"limitTokens"`
	ResetHour   int    `json:"resetHour"`
}

type GuardrailSettings struct {
	Enabled                     bool    `json:"enabled"`
	SpikePercent                float64 `json:"spikePercent"`
	ProjectConcentrationPercent float64 `json:"projectConcentrationPercent"`
	SessionConcentrationPercent float64 `json:"sessionConcentrationPercent"`
	CacheReadMinimumPercent     float64 `json:"cacheReadMinimumPercent"`
	MinimumTokens               int64   `json:"minimumTokens"`
}

type OptimizationSettings struct {
	Enabled                bool    `json:"enabled"`
	Goal                   string  `json:"goal"`
	TargetReductionPercent float64 `json:"targetReductionPercent"`
	HorizonMinutes         int     `json:"horizonMinutes"`
	ExecutionMode          string  `json:"executionMode"`
}

// ControlSettings define the local decision policy used by the Varnqor 4
// Control Plane. These thresholds never alter an external agent directly;
// they decide when Varnqor should surface an incident or intervention.
type ControlSettings struct {
	Enabled                  bool    `json:"enabled"`
	Profile                  string  `json:"profile"`
	BurnCeilingRatio         float64 `json:"burnCeilingRatio"`
	SessionDominancePercent  float64 `json:"sessionDominancePercent"`
	ProjectDominancePercent  float64 `json:"projectDominancePercent"`
	CacheFloorPercent        float64 `json:"cacheFloorPercent"`
	BudgetWarningPercent     float64 `json:"budgetWarningPercent"`
	MissionActiveMinutes     int     `json:"missionActiveMinutes"`
	MissionCoolingMinutes    int     `json:"missionCoolingMinutes"`
	DecisionHorizonMinutes   int     `json:"decisionHorizonMinutes"`
	MinimumEvaluationTokens int64   `json:"minimumEvaluationTokens"`
}

type Settings struct {
	Version        int                  `json:"version"`
	RefreshSeconds int                  `json:"refreshSeconds"`
	Language       string               `json:"language"`
	Theme          string               `json:"theme"`
	Budget         ConfiguredBudget     `json:"budget"`
	Guardrails     GuardrailSettings    `json:"guardrails"`
	Optimization   OptimizationSettings `json:"optimization"`
	Control        ControlSettings      `json:"control"`
}

func defaultGuardrails() GuardrailSettings {
	return GuardrailSettings{
		Enabled:                     true,
		SpikePercent:                75,
		ProjectConcentrationPercent: 70,
		SessionConcentrationPercent: 60,
		CacheReadMinimumPercent:     15,
		MinimumTokens:               1000,
	}
}

func defaultOptimization() OptimizationSettings {
	return OptimizationSettings{
		Enabled:                true,
		Goal:                   "balanced",
		TargetReductionPercent: 20,
		HorizonMinutes:         120,
		ExecutionMode:          "preview",
	}
}

func defaultControl() ControlSettings {
	return ControlSettings{
		Enabled:                  true,
		Profile:                  "balanced",
		BurnCeilingRatio:         1.70,
		SessionDominancePercent:  62,
		ProjectDominancePercent:  72,
		CacheFloorPercent:        15,
		BudgetWarningPercent:     80,
		MissionActiveMinutes:     12,
		MissionCoolingMinutes:    45,
		DecisionHorizonMinutes:   120,
		MinimumEvaluationTokens: 1000,
	}
}

func Default() Settings {
	return Settings{
		Version:        currentVersion,
		RefreshSeconds: 5,
		Language:       "system",
		// V4 intentionally returns to the V2 visual language as its default.
		Theme: "dark",
		Budget: ConfiguredBudget{
			Label:     "Configured daily budget",
			ResetHour: 0,
		},
		Guardrails:   defaultGuardrails(),
		Optimization: defaultOptimization(),
		Control:      defaultControl(),
	}
}

func Normalize(value Settings) Settings {
	previousVersion := value.Version
	value.Version = currentVersion
	if value.RefreshSeconds < 2 { value.RefreshSeconds = 2 }
	if value.RefreshSeconds > 60 { value.RefreshSeconds = 60 }
	if value.Budget.ResetHour < 0 { value.Budget.ResetHour = 0 }
	if value.Budget.ResetHour > 23 { value.Budget.ResetHour = 23 }
	if value.Budget.Label == "" { value.Budget.Label = "Configured daily budget" }
	if value.Budget.LimitTokens < 0 { value.Budget.LimitTokens = 0 }

	if previousVersion < 2 { value.Guardrails = defaultGuardrails() }
	if value.Guardrails.SpikePercent <= 0 { value.Guardrails.SpikePercent = 75 }
	if value.Guardrails.ProjectConcentrationPercent <= 0 { value.Guardrails.ProjectConcentrationPercent = 70 }
	if value.Guardrails.SessionConcentrationPercent <= 0 { value.Guardrails.SessionConcentrationPercent = 60 }
	if value.Guardrails.CacheReadMinimumPercent < 0 { value.Guardrails.CacheReadMinimumPercent = 15 }
	if value.Guardrails.MinimumTokens <= 0 { value.Guardrails.MinimumTokens = 1000 }

	if previousVersion < 3 {
		value.Language = "system"
		value.Theme = "system"
		value.Optimization = defaultOptimization()
	}
	value.Language = normalizeLanguage(value.Language)
	value.Theme = normalizeTheme(value.Theme)
	value.Optimization.Goal = strings.ToLower(strings.TrimSpace(value.Optimization.Goal))
	value.Optimization.ExecutionMode = strings.ToLower(strings.TrimSpace(value.Optimization.ExecutionMode))
	if value.Optimization.Goal == "" { value.Optimization.Goal = "balanced" }
	if value.Optimization.ExecutionMode == "" { value.Optimization.ExecutionMode = "preview" }
	if value.Optimization.TargetReductionPercent <= 0 { value.Optimization.TargetReductionPercent = 20 }
	if value.Optimization.HorizonMinutes <= 0 { value.Optimization.HorizonMinutes = 120 }
	if value.Optimization.HorizonMinutes < 30 { value.Optimization.HorizonMinutes = 30 }
	if value.Optimization.HorizonMinutes > 1440 { value.Optimization.HorizonMinutes = 1440 }

	if previousVersion < 4 {
		value.Control = defaultControl()
		// V3's system/light warm palette was the wrong visual direction. Preserve
		// an explicit light choice, but migrate system/empty installs to V2 dark.
		if value.Theme == "system" || value.Theme == "" { value.Theme = "dark" }
	}
	value.Control.Profile = strings.ToLower(strings.TrimSpace(value.Control.Profile))
	if value.Control.Profile == "" { value.Control.Profile = "balanced" }
	if value.Control.BurnCeilingRatio <= 0 { value.Control.BurnCeilingRatio = 1.70 }
	if value.Control.SessionDominancePercent <= 0 { value.Control.SessionDominancePercent = 62 }
	if value.Control.ProjectDominancePercent <= 0 { value.Control.ProjectDominancePercent = 72 }
	if value.Control.CacheFloorPercent < 0 { value.Control.CacheFloorPercent = 15 }
	if value.Control.BudgetWarningPercent <= 0 { value.Control.BudgetWarningPercent = 80 }
	if value.Control.MissionActiveMinutes <= 0 { value.Control.MissionActiveMinutes = 12 }
	if value.Control.MissionCoolingMinutes <= 0 { value.Control.MissionCoolingMinutes = 45 }
	if value.Control.MissionCoolingMinutes < value.Control.MissionActiveMinutes { value.Control.MissionCoolingMinutes = value.Control.MissionActiveMinutes }
	if value.Control.DecisionHorizonMinutes <= 0 { value.Control.DecisionHorizonMinutes = 120 }
	if value.Control.MinimumEvaluationTokens <= 0 { value.Control.MinimumEvaluationTokens = 1000 }
	return value
}

func normalizeLanguage(value string) string {
	value = strings.TrimSpace(value)
	switch strings.ToLower(value) {
	case "", "system", "auto": return "system"
	case "en", "en-us", "en-gb": return "en"
	case "pt", "pt-br", "pt_br": return "pt-BR"
	case "es", "es-es", "es-419": return "es"
	case "de", "de-de": return "de"
	case "fr", "fr-fr": return "fr"
	default: return value
	}
}

func normalizeTheme(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "auto" { return "system" }
	return value
}

func Validate(value Settings) error {
	if value.RefreshSeconds < 2 || value.RefreshSeconds > 60 { return errors.New("refreshSeconds must be between 2 and 60") }
	if !oneOf(value.Language, "system", "en", "pt-BR", "es", "de", "fr") { return errors.New("language must be system, en, pt-BR, es, de, or fr") }
	if !oneOf(value.Theme, "system", "light", "dark") { return errors.New("theme must be system, light, or dark") }
	if value.Budget.ResetHour < 0 || value.Budget.ResetHour > 23 { return errors.New("budget.resetHour must be between 0 and 23") }
	if value.Budget.Enabled && value.Budget.LimitTokens <= 0 { return errors.New("budget.limitTokens must be greater than zero when the budget is enabled") }
	if value.Guardrails.SpikePercent < 10 || value.Guardrails.SpikePercent > 1000 { return errors.New("guardrails.spikePercent must be between 10 and 1000") }
	if value.Guardrails.ProjectConcentrationPercent < 10 || value.Guardrails.ProjectConcentrationPercent > 100 { return errors.New("guardrails.projectConcentrationPercent must be between 10 and 100") }
	if value.Guardrails.SessionConcentrationPercent < 10 || value.Guardrails.SessionConcentrationPercent > 100 { return errors.New("guardrails.sessionConcentrationPercent must be between 10 and 100") }
	if value.Guardrails.CacheReadMinimumPercent < 0 || value.Guardrails.CacheReadMinimumPercent > 100 { return errors.New("guardrails.cacheReadMinimumPercent must be between 0 and 100") }
	if value.Guardrails.MinimumTokens < 1 { return errors.New("guardrails.minimumTokens must be greater than zero") }
	if !oneOf(value.Optimization.Goal, "balanced", "economy", "throughput") { return errors.New("optimization.goal must be balanced, economy, or throughput") }
	if value.Optimization.TargetReductionPercent < 5 || value.Optimization.TargetReductionPercent > 60 { return errors.New("optimization.targetReductionPercent must be between 5 and 60") }
	if value.Optimization.HorizonMinutes < 30 || value.Optimization.HorizonMinutes > 1440 { return errors.New("optimization.horizonMinutes must be between 30 and 1440") }
	if !oneOf(value.Optimization.ExecutionMode, "preview", "manual") { return errors.New("optimization.executionMode must be preview or manual") }
	if !oneOf(value.Control.Profile, "balanced", "economy", "throughput") { return errors.New("control.profile must be balanced, economy, or throughput") }
	if value.Control.BurnCeilingRatio < 1 || value.Control.BurnCeilingRatio > 10 { return errors.New("control.burnCeilingRatio must be between 1 and 10") }
	if value.Control.SessionDominancePercent < 20 || value.Control.SessionDominancePercent > 100 { return errors.New("control.sessionDominancePercent must be between 20 and 100") }
	if value.Control.ProjectDominancePercent < 20 || value.Control.ProjectDominancePercent > 100 { return errors.New("control.projectDominancePercent must be between 20 and 100") }
	if value.Control.CacheFloorPercent < 0 || value.Control.CacheFloorPercent > 100 { return errors.New("control.cacheFloorPercent must be between 0 and 100") }
	if value.Control.BudgetWarningPercent < 10 || value.Control.BudgetWarningPercent > 100 { return errors.New("control.budgetWarningPercent must be between 10 and 100") }
	if value.Control.MissionActiveMinutes < 1 || value.Control.MissionActiveMinutes > 120 { return errors.New("control.missionActiveMinutes must be between 1 and 120") }
	if value.Control.MissionCoolingMinutes < value.Control.MissionActiveMinutes || value.Control.MissionCoolingMinutes > 720 { return errors.New("control.missionCoolingMinutes must be between active minutes and 720") }
	if value.Control.DecisionHorizonMinutes < 30 || value.Control.DecisionHorizonMinutes > 1440 { return errors.New("control.decisionHorizonMinutes must be between 30 and 1440") }
	if value.Control.MinimumEvaluationTokens < 1 { return errors.New("control.minimumEvaluationTokens must be greater than zero") }
	return nil
}

func oneOf(value string, options ...string) bool {
	for _, option := range options { if value == option { return true } }
	return false
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil { return "", fmt.Errorf("resolve config directory: %w", err) }
	return filepath.Join(configDir, "varnqor", "settings.json"), nil
}

func LoadDefault() (Settings, error) {
	path, err := DefaultPath()
	if err != nil { return Default(), err }
	return Load(path)
}

func Load(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) { return Default(), nil }
	if err != nil { return Default(), fmt.Errorf("read settings: %w", err) }
	var value Settings
	if err := json.Unmarshal(data, &value); err != nil { return Default(), fmt.Errorf("decode settings: %w", err) }
	value = Normalize(value)
	if err := Validate(value); err != nil { return Default(), fmt.Errorf("validate settings: %w", err) }
	return value, nil
}

func SaveDefault(value Settings) (Settings, error) {
	path, err := DefaultPath()
	if err != nil { return Default(), err }
	return Save(path, value)
}

func Save(path string, value Settings) (Settings, error) {
	value = Normalize(value)
	if err := Validate(value); err != nil { return value, err }
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil { return value, fmt.Errorf("encode settings: %w", err) }
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil { return value, fmt.Errorf("create settings directory: %w", err) }
	temp, err := os.CreateTemp(filepath.Dir(path), "settings-*.tmp")
	if err != nil { return value, fmt.Errorf("create settings temp file: %w", err) }
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil { _ = temp.Close(); return value, fmt.Errorf("secure settings temp file: %w", err) }
	if _, err := temp.Write(data); err != nil { _ = temp.Close(); return value, fmt.Errorf("write settings: %w", err) }
	if err := temp.Sync(); err != nil { _ = temp.Close(); return value, fmt.Errorf("sync settings: %w", err) }
	if err := temp.Close(); err != nil { return value, fmt.Errorf("close settings: %w", err) }
	if err := os.Rename(tempPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) { return value, fmt.Errorf("replace settings: %w", err) }
		if err := os.Rename(tempPath, path); err != nil { return value, fmt.Errorf("replace settings: %w", err) }
	}
	return value, nil
}
