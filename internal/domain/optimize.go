package domain

import "time"

type OptimizationPolicy struct {
	Enabled                bool    `json:"enabled"`
	Goal                   string  `json:"goal"`
	TargetReductionPercent float64 `json:"targetReductionPercent"`
	HorizonMinutes         int     `json:"horizonMinutes"`
	ExecutionMode          string  `json:"executionMode"`
}

type OptimizationScenario struct {
	CurrentPerHour         int64   `json:"currentPerHour"`
	HorizonMinutes         int     `json:"horizonMinutes"`
	ProjectedTokens        int64   `json:"projectedTokens"`
	TargetReductionPercent float64 `json:"targetReductionPercent"`
	TargetTokens           int64   `json:"targetTokens"`
	TargetSavings          int64   `json:"targetSavings"`
	Feasibility            string  `json:"feasibility"`
}

type OptimizationAction struct {
	ID                   string   `json:"id"`
	Kind                 string   `json:"kind"`
	Priority             string   `json:"priority"`
	Title                string   `json:"title"`
	Summary              string   `json:"summary"`
	Evidence             string   `json:"evidence"`
	Provider             string   `json:"provider,omitempty"`
	Project              string   `json:"project,omitempty"`
	Session              string   `json:"session,omitempty"`
	Confidence           string   `json:"confidence"`
	EstimatedSavingsLow  int64    `json:"estimatedSavingsLow"`
	EstimatedSavingsHigh int64    `json:"estimatedSavingsHigh"`
	NonAdditive          bool     `json:"nonAdditive"`
	Steps                []string `json:"steps"`
}

type OptimizationReport struct {
	GeneratedAt        time.Time             `json:"generatedAt"`
	State              string                `json:"state"`
	Goal               string                `json:"goal"`
	Mode               string                `json:"mode"`
	ExecutionAvailable bool                  `json:"executionAvailable"`
	Scenario           OptimizationScenario  `json:"scenario"`
	Actions            []OptimizationAction  `json:"actions"`
	ProvidersObserved  int                   `json:"providersObserved"`
	ProjectsObserved   int                   `json:"projectsObserved"`
	SessionsObserved   int                   `json:"sessionsObserved"`
	Note               string                `json:"note"`
}
