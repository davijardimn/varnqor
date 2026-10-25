package domain

type ForecastPoint struct {
	Minutes         int   `json:"minutes"`
	ProjectedTokens int64 `json:"projectedTokens"`
	BaselineTokens  int64 `json:"baselineTokens"`
	EnvelopeLow     int64 `json:"envelopeLow"`
	EnvelopeHigh    int64 `json:"envelopeHigh"`
}

type ForecastReport struct {
	Ready              bool            `json:"ready"`
	Method             string          `json:"method"`
	Pressure           string          `json:"pressure"`
	Confidence         string          `json:"confidence"`
	CurrentPerHour     int64           `json:"currentPerHour"`
	PreviousPerHour    int64           `json:"previousPerHour"`
	BaselinePerHour    int64           `json:"baselinePerHour"`
	PressureRatio      float64         `json:"pressureRatio"`
	Trend              string          `json:"trend"`
	HorizonMinutes     int             `json:"horizonMinutes"`
	ProjectedTokens    int64           `json:"projectedTokens"`
	BaselineProjection int64           `json:"baselineProjection"`
	EnvelopeLow        int64           `json:"envelopeLow"`
	EnvelopeHigh       int64           `json:"envelopeHigh"`
	RunwayAvailable    bool            `json:"runwayAvailable"`
	LimitSource        string          `json:"limitSource"`
	Points             []ForecastPoint `json:"points"`
}
