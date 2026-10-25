package domain

import "time"

type AdvisorRecommendation struct {
	ID              string `json:"id"`
	Priority        string `json:"priority"`
	Category        string `json:"category"`
	Title           string `json:"title"`
	Evidence        string `json:"evidence"`
	Recommendation  string `json:"recommendation"`
	EstimatedTokens int64  `json:"estimatedTokens,omitempty"`
	RuleID          string `json:"ruleId,omitempty"`
}

type AdvisorReport struct {
	GeneratedAt     time.Time               `json:"generatedAt"`
	Mode            string                  `json:"mode"`
	High            int                     `json:"high"`
	Medium          int                     `json:"medium"`
	Low             int                     `json:"low"`
	Recommendations []AdvisorRecommendation `json:"recommendations"`
}
