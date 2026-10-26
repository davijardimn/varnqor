package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const currentVersion = 2

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

type Settings struct {
	Version        int               `json:"version"`
	RefreshSeconds int               `json:"refreshSeconds"`
	Budget         ConfiguredBudget  `json:"budget"`
	Guardrails     GuardrailSettings `json:"guardrails"`
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

func Default() Settings {
	return Settings{
		Version:        currentVersion,
		RefreshSeconds: 5,
		Budget: ConfiguredBudget{
			Label:     "Configured daily budget",
			ResetHour: 0,
		},
		Guardrails: defaultGuardrails(),
	}
}

func Normalize(value Settings) Settings {
	legacy := value.Version < currentVersion
	value.Version = currentVersion
	if value.RefreshSeconds < 2 {
		value.RefreshSeconds = 2
	}
	if value.RefreshSeconds > 60 {
		value.RefreshSeconds = 60
	}
	if value.Budget.ResetHour < 0 {
		value.Budget.ResetHour = 0
	}
	if value.Budget.ResetHour > 23 {
		value.Budget.ResetHour = 23
	}
	if value.Budget.Label == "" {
		value.Budget.Label = "Configured daily budget"
	}
	if value.Budget.LimitTokens < 0 {
		value.Budget.LimitTokens = 0
	}
	if legacy {
		value.Guardrails = defaultGuardrails()
	}
	if value.Guardrails.SpikePercent <= 0 {
		value.Guardrails.SpikePercent = 75
	}
	if value.Guardrails.ProjectConcentrationPercent <= 0 {
		value.Guardrails.ProjectConcentrationPercent = 70
	}
	if value.Guardrails.SessionConcentrationPercent <= 0 {
		value.Guardrails.SessionConcentrationPercent = 60
	}
	if value.Guardrails.CacheReadMinimumPercent <= 0 {
		value.Guardrails.CacheReadMinimumPercent = 15
	}
	if value.Guardrails.MinimumTokens <= 0 {
		value.Guardrails.MinimumTokens = 1000
	}
	return value
}

func Validate(value Settings) error {
	if value.RefreshSeconds < 2 || value.RefreshSeconds > 60 {
		return errors.New("refreshSeconds must be between 2 and 60")
	}
	if value.Budget.ResetHour < 0 || value.Budget.ResetHour > 23 {
		return errors.New("budget.resetHour must be between 0 and 23")
	}
	if value.Budget.Enabled && value.Budget.LimitTokens <= 0 {
		return errors.New("budget.limitTokens must be greater than zero when the budget is enabled")
	}
	if value.Guardrails.SpikePercent < 10 || value.Guardrails.SpikePercent > 1000 {
		return errors.New("guardrails.spikePercent must be between 10 and 1000")
	}
	if value.Guardrails.ProjectConcentrationPercent < 10 || value.Guardrails.ProjectConcentrationPercent > 100 {
		return errors.New("guardrails.projectConcentrationPercent must be between 10 and 100")
	}
	if value.Guardrails.SessionConcentrationPercent < 10 || value.Guardrails.SessionConcentrationPercent > 100 {
		return errors.New("guardrails.sessionConcentrationPercent must be between 10 and 100")
	}
	if value.Guardrails.CacheReadMinimumPercent < 0 || value.Guardrails.CacheReadMinimumPercent > 100 {
		return errors.New("guardrails.cacheReadMinimumPercent must be between 0 and 100")
	}
	if value.Guardrails.MinimumTokens < 1 {
		return errors.New("guardrails.minimumTokens must be greater than zero")
	}
	return nil
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config directory: %w", err)
	}
	return filepath.Join(configDir, "varnqor", "settings.json"), nil
}

func LoadDefault() (Settings, error) {
	path, err := DefaultPath()
	if err != nil {
		return Default(), err
	}
	return Load(path)
}

func Load(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Default(), fmt.Errorf("read settings: %w", err)
	}
	var value Settings
	if err := json.Unmarshal(data, &value); err != nil {
		return Default(), fmt.Errorf("decode settings: %w", err)
	}
	value = Normalize(value)
	if err := Validate(value); err != nil {
		return Default(), fmt.Errorf("validate settings: %w", err)
	}
	return value, nil
}

func SaveDefault(value Settings) (Settings, error) {
	path, err := DefaultPath()
	if err != nil {
		return Default(), err
	}
	return Save(path, value)
}

func Save(path string, value Settings) (Settings, error) {
	value = Normalize(value)
	if err := Validate(value); err != nil {
		return value, err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return value, fmt.Errorf("encode settings: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return value, fmt.Errorf("create settings directory: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(path), "settings-*.tmp")
	if err != nil {
		return value, fmt.Errorf("create settings temp file: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return value, fmt.Errorf("secure settings temp file: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return value, fmt.Errorf("write settings: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return value, fmt.Errorf("sync settings: %w", err)
	}
	if err := temp.Close(); err != nil {
		return value, fmt.Errorf("close settings: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return value, fmt.Errorf("replace settings: %w", err)
		}
		if err := os.Rename(tempPath, path); err != nil {
			return value, fmt.Errorf("replace settings: %w", err)
		}
	}
	return value, nil
}
