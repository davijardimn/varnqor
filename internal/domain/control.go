package domain

import "time"

// ControlPolicySnapshot is the exact local policy used for one Control Plane
// evaluation. It is returned with every report so the UI can explain why an
// incident exists instead of hiding the decision behind an opaque score.
type ControlPolicySnapshot struct {
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

type ControlMission struct {
	ID            string      `json:"id"`
	Provider      string      `json:"provider"`
	Project       string      `json:"project"`
	Session       string      `json:"session"`
	State         string      `json:"state"`
	Pressure      string      `json:"pressure"`
	Calls         int         `json:"calls"`
	Tokens        TokenTotals `json:"tokens"`
	SharePercent  float64     `json:"sharePercent"`
	VelocityHour  int64       `json:"velocityHour"`
	FirstActivity time.Time   `json:"firstActivity"`
	LastActivity  time.Time   `json:"lastActivity"`
	AgeMinutes    int         `json:"ageMinutes"`
	Subagent      bool        `json:"subagent"`
}

type ControlIncident struct {
	ID        string  `json:"id"`
	Severity  string  `json:"severity"`
	Kind      string  `json:"kind"`
	Title     string  `json:"title"`
	Detail    string  `json:"detail"`
	Source    string  `json:"source"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	Provider  string  `json:"provider,omitempty"`
	Project   string  `json:"project,omitempty"`
	Session   string  `json:"session,omitempty"`
}

type ControlDecision struct {
	ID                   string   `json:"id"`
	Priority             string   `json:"priority"`
	Kind                 string   `json:"kind"`
	Title                string   `json:"title"`
	Reason               string   `json:"reason"`
	Provider             string   `json:"provider,omitempty"`
	Project              string   `json:"project,omitempty"`
	Session              string   `json:"session,omitempty"`
	Confidence           string   `json:"confidence"`
	EstimatedSavingsLow  int64    `json:"estimatedSavingsLow"`
	EstimatedSavingsHigh int64    `json:"estimatedSavingsHigh"`
	NonAdditive          bool     `json:"nonAdditive"`
	Steps                []string `json:"steps"`
}

type ControlPlaneReport struct {
	GeneratedAt       time.Time             `json:"generatedAt"`
	State             string                `json:"state"`
	RangeMinutes      int                   `json:"rangeMinutes"`
	CurrentPerHour    int64                 `json:"currentPerHour"`
	ProjectedTokens   int64                 `json:"projectedTokens"`
	ActiveProviders   int                   `json:"activeProviders"`
	ActiveMissions    int                   `json:"activeMissions"`
	CoolingMissions   int                   `json:"coolingMissions"`
	HighIncidents     int                   `json:"highIncidents"`
	MediumIncidents   int                   `json:"mediumIncidents"`
	LowIncidents      int                   `json:"lowIncidents"`
	Policy            ControlPolicySnapshot `json:"policy"`
	Missions          []ControlMission      `json:"missions"`
	Incidents         []ControlIncident     `json:"incidents"`
	Decisions         []ControlDecision     `json:"decisions"`
	ExternalExecution bool                  `json:"externalExecution"`
	Note              string                `json:"note"`
}
