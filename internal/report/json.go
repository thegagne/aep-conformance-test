package report

import (
	"encoding/json"
	"io"

	"github.com/thegagne/aep-conformance-test/internal/checks"
)

// jsonResult is the serialized form of a result with string enums.
type jsonResult struct {
	CheckID   string            `json:"check_id"`
	AEP       int               `json:"aep"`
	Title     string            `json:"title"`
	Level     string            `json:"level"`
	Status    string            `json:"status"`
	Severity  string            `json:"severity"`
	Message   string            `json:"message,omitempty"`
	SpecQuote string            `json:"spec_quote,omitempty"`
	Evidence  []checks.Evidence `json:"evidence,omitempty"`
}

// jsonCollection nests a collection's capabilities and results together.
type jsonCollection struct {
	Collection   string                  `json:"collection"`
	Capabilities *CollectionCapabilities `json:"capabilities,omitempty"`
	Results      []jsonResult            `json:"results"`
}

type jsonReport struct {
	API         string           `json:"api"`
	Target      string           `json:"target,omitempty"`
	Conformant  bool             `json:"conformant"`
	Summary     Summary          `json:"summary"`
	APIChecks   []jsonResult     `json:"api_checks,omitempty"`
	Collections []jsonCollection `json:"collections"`
}

// WriteJSON renders the report as machine-readable JSON, with capabilities and
// results nested under a top-level resources array.
func (r *Report) WriteJSON(w io.Writer) error {
	toJSON := func(res checks.Result) jsonResult {
		return jsonResult{
			CheckID:   res.CheckID,
			AEP:       res.AEP,
			Title:     res.Title,
			Level:     res.Level.String(),
			Status:    res.Status.String(),
			Severity:  r.Severity(res).String(),
			Message:   res.Message,
			SpecQuote: res.SpecQuote,
			Evidence:  res.Evidence,
		}
	}

	api, groups := r.Grouped()
	r.sortByAEP(api)
	out := jsonReport{
		API:        r.Title,
		Target:     r.BaseURL,
		Conformant: r.Conformant(),
		Summary:    r.Summary(),
	}
	for _, res := range api {
		out.APIChecks = append(out.APIChecks, toJSON(res))
	}
	for _, g := range groups {
		rr := jsonCollection{Collection: g.Collection, Capabilities: g.Capabilities}
		r.sortByAEP(g.Results)
		for _, res := range g.Results {
			rr.Results = append(rr.Results, toJSON(res))
		}
		out.Collections = append(out.Collections, rr)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
