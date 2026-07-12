// Package checks defines the conformance-check framework and the registry of
// checks. Each check is a small registered value (metadata + a closure) so the
// catalog scales without bespoke wiring in the runner. Static checks read only
// the discovered model; dynamic checks read a Probe of live interactions.
package checks

import (
	"fmt"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// Level is the normative strength of a requirement.
type Level int

const (
	MUST Level = iota
	MUSTNOT
	SHOULD
	SHOULDNOT
	MAY
)

func (l Level) String() string {
	switch l {
	case MUST:
		return "MUST"
	case MUSTNOT:
		return "MUST NOT"
	case SHOULD:
		return "SHOULD"
	case SHOULDNOT:
		return "SHOULD NOT"
	case MAY:
		return "MAY"
	}
	return "?"
}

// Required reports whether a failure of this level breaks conformance.
func (l Level) Required() bool { return l == MUST || l == MUSTNOT }

// Status is the outcome of a check.
type Status int

const (
	Pass Status = iota
	Fail
	Warn
	NotApplicable
	Skipped
)

func (s Status) String() string {
	switch s {
	case Pass:
		return "PASS"
	case Fail:
		return "FAIL"
	case Warn:
		return "WARN"
	case NotApplicable:
		return "N/A"
	case Skipped:
		return "SKIP"
	}
	return "?"
}

// Evidence captures one request/response pair supporting a result.
type Evidence struct {
	Method   string `json:"method,omitempty"`
	URL      string `json:"url,omitempty"`
	Request  string `json:"request,omitempty"`
	Status   int    `json:"status,omitempty"`
	Response string `json:"response,omitempty"`
}

// Result is the outcome of running a check for a resource.
type Result struct {
	CheckID   string     `json:"check_id"`
	AEP       int        `json:"aep"`
	Title     string     `json:"title"`
	Level     Level      `json:"level"`
	Status    Status     `json:"status"`
	Resource  string     `json:"resource,omitempty"`
	Message   string     `json:"message,omitempty"`
	SpecQuote string     `json:"spec_quote,omitempty"`
	Evidence  []Evidence `json:"evidence,omitempty"`
}

// RunContext is passed to every check.
type RunContext struct {
	Model    *discovery.APIModel
	Resource *discovery.Resource
	// Probe holds captured live interactions; nil when running static-only
	// (no base URL), in which case dynamic checks report Skipped.
	Probe *Probe
}

// Check is one registered conformance check.
type Check struct {
	ID    string
	AEP   int
	Title string
	Level Level
	// Static checks need only the model; dynamic checks need a Probe.
	Static bool
	// PerAPI checks run once against the whole model (Resource is nil) rather
	// than once per resource — e.g. the OpenAPI version and API-name checks.
	PerAPI bool
	// Applicable gates the check; when it returns false the check reports
	// NotApplicable (documenting the optional area rather than failing).
	Applicable func(rc *RunContext) bool
	// Run performs the assertion and returns a bare result; the framework fills
	// in metadata (ID, AEP, level, resource) afterward.
	Run func(rc *RunContext) Result
}

var registry []Check

// Register adds a check to the global registry. Intended for use from init().
func Register(c Check) {
	if c.ID == "" {
		panic("check with empty ID")
	}
	registry = append(registry, c)
}

// All returns every registered check.
func All() []Check { return registry }

// Evaluate runs a single check within a context, applying the Static/Applicable
// gates and stamping metadata onto the returned result.
func Evaluate(c Check, rc *RunContext) Result {
	stamp := func(r Result) Result {
		r.CheckID = c.ID
		r.AEP = c.AEP
		r.Title = c.Title
		r.Level = c.Level
		if rc.Resource != nil {
			r.Resource = rc.Resource.Singular
		}
		return r
	}
	if !c.Static && rc.Probe == nil {
		return stamp(Result{Status: Skipped, Message: "no live server; dynamic check skipped"})
	}
	if c.Applicable != nil && !c.Applicable(rc) {
		return stamp(Result{Status: NotApplicable, Message: "feature not present in spec"})
	}
	return stamp(c.Run(rc))
}

// Helpers for building results inside a check.

func pass(msg string, ev ...Evidence) Result {
	return Result{Status: Pass, Message: msg, Evidence: ev}
}

func fail(msg string, ev ...Evidence) Result {
	return Result{Status: Fail, Message: msg, Evidence: ev}
}

func na(msg string) Result { return Result{Status: NotApplicable, Message: msg} }

func quote(r Result, q string) Result { r.SpecQuote = q; return r }

func failf(format string, a ...any) Result {
	return Result{Status: Fail, Message: fmt.Sprintf(format, a...)}
}
