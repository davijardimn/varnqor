package domain

type DoctorCheck struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Status  string `json:"status"`
	Detail  string `json:"detail"`
}

type DoctorReport struct {
	Status  string        `json:"status"`
	Version string        `json:"version"`
	Pass    int           `json:"pass"`
	Warn    int           `json:"warn"`
	Fail    int           `json:"fail"`
	Checks  []DoctorCheck `json:"checks"`
}
