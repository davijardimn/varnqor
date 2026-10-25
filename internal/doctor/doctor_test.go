package doctor

import (
	"testing"

	"github.com/davijardimn/varnqor/internal/domain"
)

func TestSummarizePrefersFailures(t *testing.T) {
	report := Summarize(domain.DoctorReport{Checks: []domain.DoctorCheck{{Status: "pass"}, {Status: "warn"}, {Status: "fail"}}})
	if report.Status != "error" || report.Pass != 1 || report.Warn != 1 || report.Fail != 1 {
		t.Fatalf("unexpected summary: %+v", report)
	}
}

func TestSummarizeReportsWarningWithoutFailure(t *testing.T) {
	report := Summarize(domain.DoctorReport{Checks: []domain.DoctorCheck{{Status: "pass"}, {Status: "warn"}}})
	if report.Status != "warning" || report.Fail != 0 {
		t.Fatalf("unexpected summary: %+v", report)
	}
}
