package domain

import "time"

type QuotaStatus struct {
	Available             bool      `json:"available"`
	Source                string    `json:"source"`
	Metric                string    `json:"metric"`
	Label                 string    `json:"label"`
	Limit                  int64     `json:"limit"`
	Used                   int64     `json:"used"`
	Remaining              int64     `json:"remaining"`
	UtilizationPercent     float64   `json:"utilizationPercent"`
	WindowStart            time.Time `json:"windowStart,omitempty"`
	ResetAt                time.Time `json:"resetAt,omitempty"`
	CurrentPerHour         int64     `json:"currentPerHour"`
	RunwayAvailable        bool      `json:"runwayAvailable"`
	RunwayHours            float64   `json:"runwayHours"`
	ProjectedAtReset       int64     `json:"projectedAtReset"`
	WillExceedBeforeReset  bool      `json:"willExceedBeforeReset"`
	Confidence             string    `json:"confidence"`
}
