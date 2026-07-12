package checks

import (
	"encoding/json"
	"fmt"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// failStatus builds a Fail result for an unexpected HTTP status with evidence.
func failStatus(what string, got, want int, ev Evidence) Result {
	return Result{
		Status:   Fail,
		Message:  fmt.Sprintf("%s returned %d, expected %d", what, got, want),
		Evidence: []Evidence{ev},
	}
}

// parseJSONObject decodes a JSON object string into a map (nil on failure).
func parseJSONObject(s string) map[string]any {
	var m map[string]any
	if json.Unmarshal([]byte(s), &m) != nil {
		return nil
	}
	return m
}

// need returns the interaction if it produced a response, else a Skipped result
// so dependent checks don't fail when a prerequisite step never ran.
func need(p *Probe, name string) (*Interaction, *Result) {
	i := p.I(name)
	if i == nil || i.Resp == nil {
		s := Result{Status: Skipped, Message: "prerequisite '" + name + "' did not run"}
		return nil, &s
	}
	return i, nil
}

// resourceEndpoint fetches a standard-method endpoint for the context resource.
func (rc *RunContext) endpoint(m discovery.Method) *discovery.Endpoint {
	if rc.Resource == nil {
		return nil
	}
	return rc.Resource.Method(m)
}
