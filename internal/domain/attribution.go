package domain

import "time"

type AttributionItem struct {
	Key          string      `json:"key"`
	Label        string      `json:"label"`
	Provider     string      `json:"provider,omitempty"`
	Project      string      `json:"project,omitempty"`
	Session      string      `json:"session,omitempty"`
	Calls        int         `json:"calls"`
	Tokens       TokenTotals `json:"tokens"`
	SharePercent float64     `json:"sharePercent"`
}

type AttributionReport struct {
	Start    time.Time         `json:"start"`
	End      time.Time         `json:"end"`
	Calls    int               `json:"calls"`
	Tokens   TokenTotals       `json:"tokens"`
	Projects []AttributionItem `json:"projects"`
	Sessions []AttributionItem `json:"sessions"`
	Models   []AttributionItem `json:"models"`
	Agents   []AttributionItem `json:"agents"`
}
