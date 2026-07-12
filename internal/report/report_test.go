package report

import (
	"testing"

	"github.com/thegagne/aep-conformance-test/internal/checks"
)

func TestSeverityMapping(t *testing.T) {
	r := &Report{}
	cases := []struct {
		status checks.Status
		level  checks.Level
		want   Severity
	}{
		{checks.Pass, checks.MUST, SevPass},
		{checks.Fail, checks.MUST, SevFail},
		{checks.Fail, checks.MUSTNOT, SevFail},
		{checks.Fail, checks.SHOULD, SevWarn},
		{checks.Fail, checks.MAY, SevWarn},
		{checks.NotApplicable, checks.MAY, SevNA},
		{checks.Skipped, checks.MUST, SevSkip},
	}
	for _, c := range cases {
		got := r.Severity(checks.Result{Status: c.status, Level: c.level})
		if got != c.want {
			t.Errorf("severity(%v,%v) = %v, want %v", c.status, c.level, got, c.want)
		}
	}
}

func TestStrictPromotesShouldToFail(t *testing.T) {
	r := &Report{Strict: true}
	if got := r.Severity(checks.Result{Status: checks.Fail, Level: checks.SHOULD}); got != SevFail {
		t.Errorf("strict SHOULD fail = %v, want FAIL", got)
	}
}

func TestConformantIgnoresWarnings(t *testing.T) {
	r := &Report{Results: []checks.Result{
		{Status: checks.Fail, Level: checks.SHOULD}, // warn, not fatal
		{Status: checks.Pass, Level: checks.MUST},
	}}
	if !r.Conformant() {
		t.Error("a SHOULD failure should not break conformance")
	}
	r.Results = append(r.Results, checks.Result{Status: checks.Fail, Level: checks.MUST})
	if r.Conformant() {
		t.Error("a MUST failure must break conformance")
	}
}
