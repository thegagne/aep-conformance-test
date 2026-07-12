// Package report aggregates check results into a conformance report and renders
// it as stdout, JSON, or markdown. Severity is derived from each check's level:
// a failed MUST breaks conformance; a failed SHOULD is a warning (promoted to a
// failure under --strict); MAY and absent features are informational.
package report

import (
	"sort"

	"github.com/thegagne/aep-conformance-test/internal/checks"
)

// Severity is the display/rollup classification of a result.
type Severity int

const (
	SevPass Severity = iota
	SevFail
	SevWarn
	SevNA
	SevSkip
)

func (s Severity) String() string {
	switch s {
	case SevPass:
		return "PASS"
	case SevFail:
		return "FAIL"
	case SevWarn:
		return "WARN"
	case SevNA:
		return "N/A"
	case SevSkip:
		return "SKIP"
	}
	return "?"
}

// Report holds the run results and metadata.
type Report struct {
	Title        string
	BaseURL      string
	Strict       bool
	Results      []checks.Result
	Capabilities []CollectionCapabilities
}

// Severity classifies a single result given the report's strictness.
func (r *Report) Severity(res checks.Result) Severity {
	switch res.Status {
	case checks.Pass:
		return SevPass
	case checks.NotApplicable:
		return SevNA
	case checks.Skipped:
		return SevSkip
	case checks.Fail:
		if res.Level.Required() || r.Strict {
			return SevFail
		}
		return SevWarn
	}
	return SevNA
}

// Summary counts results by severity.
type Summary struct {
	Pass  int `json:"pass"`
	Fail  int `json:"fail"`
	Warn  int `json:"warn"`
	NA    int `json:"na"`
	Skip  int `json:"skip"`
	Total int `json:"total"`
}

// Summary tallies all results.
func (r *Report) Summary() Summary {
	var s Summary
	for _, res := range r.Results {
		s.Total++
		switch r.Severity(res) {
		case SevPass:
			s.Pass++
		case SevFail:
			s.Fail++
		case SevWarn:
			s.Warn++
		case SevNA:
			s.NA++
		case SevSkip:
			s.Skip++
		}
	}
	return s
}

// Conformant reports whether the API passed all required (and, under strict,
// recommended) checks.
func (r *Report) Conformant() bool { return r.Summary().Fail == 0 }

// sevRank orders severities for display so the actionable outcomes come first
// within a group of results sharing an AEP.
func sevRank(s Severity) int {
	switch s {
	case SevFail:
		return 0
	case SevWarn:
		return 1
	case SevPass:
		return 2
	case SevSkip:
		return 3
	case SevNA:
		return 4
	}
	return 5
}

// sortByAEP orders results by AEP number (spec order) first, then by severity so
// failures surface within each AEP, then by check ID for stable output. Every
// renderer sorts through this so ordering is consistent across formats.
func (r *Report) sortByAEP(res []checks.Result) {
	sort.SliceStable(res, func(i, j int) bool {
		if res[i].AEP != res[j].AEP {
			return res[i].AEP < res[j].AEP
		}
		if si, sj := sevRank(r.Severity(res[i])), sevRank(r.Severity(res[j])); si != sj {
			return si < sj
		}
		return res[i].CheckID < res[j].CheckID
	})
}

// CollectionGroup bundles a collection's capabilities with its check results.
// Collection is the plural display name (e.g. "books"); results are matched by
// the resource singular that checks record.
type CollectionGroup struct {
	Collection   string
	Capabilities *CollectionCapabilities
	Results      []checks.Result
}

// Grouped splits results into API-level checks (not tied to a collection) and
// per-collection groups, each carrying that collection's capabilities. Order
// follows first appearance in the results.
func (r *Report) Grouped() (api []checks.Result, groups []CollectionGroup) {
	capBySingular := map[string]CollectionCapabilities{}
	for i := range r.Capabilities {
		capBySingular[r.Capabilities[i].Singular] = r.Capabilities[i]
	}
	idx := map[string]int{}
	for _, res := range r.Results {
		if res.Resource == "" {
			api = append(api, res)
			continue
		}
		i, ok := idx[res.Resource]
		if !ok {
			i = len(groups)
			idx[res.Resource] = i
			g := CollectionGroup{Collection: res.Resource} // fallback to singular
			if c, ok := capBySingular[res.Resource]; ok {
				cc := c
				g.Capabilities = &cc
				g.Collection = c.Collection
			}
			groups = append(groups, g)
		}
		groups[i].Results = append(groups[i].Results, res)
	}
	return api, groups
}
