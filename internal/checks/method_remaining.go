package checks

import (
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func init() {
	// AEP-122: data returned from the API uses the canonical resource path — the
	// server-assigned path must match the resource's pattern (literal collection
	// segments intact, id segments filled with real, non-wildcard values).
	Register(Check{
		ID: "path-matches-canonical-pattern", AEP: 122, Level: MUST,
		Title:      "Returned path matches the canonical pattern",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodGet) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Get)
			if skip != nil {
				return *skip
			}
			path := ""
			if i.Resp.JSON != nil {
				path, _ = i.Resp.JSON["path"].(string)
			}
			if path == "" {
				path = rc.Probe.CreatedPath
			}
			if path == "" {
				return Result{Status: Skipped, Message: "no resource path to check"}
			}
			patSegs := strings.Split(rc.Resource.CanonicalPattern(), "/")
			gotSegs := strings.Split(strings.Trim(path, "/"), "/")
			if len(patSegs) != len(gotSegs) {
				return quote(failf("returned path %q does not match pattern %q (segment count)", path, rc.Resource.CanonicalPattern()),
					"All data returned from the API must use the canonical resource path.")
			}
			for idx, ps := range patSegs {
				gs := gotSegs[idx]
				if isPatternVar(ps) {
					if gs == "" || gs == "-" {
						return quote(failf("returned path %q has an empty/wildcard id segment", path),
							"All data returned from the API must use the canonical resource path.")
					}
					continue
				}
				if gs != ps {
					return quote(failf("returned path %q has segment %q where pattern expects literal %q", path, gs, ps),
						"All data returned from the API must use the canonical resource path.")
				}
			}
			return pass("returned path matches canonical pattern: " + path)
		},
	})

	// AEP-134: the Update response should be the fully-populated resource.
	Register(Check{
		ID: "update-fully-populated", AEP: 134, Level: SHOULD,
		Title:      "Update returns a fully-populated resource",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodUpdate) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Update)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status >= 300 {
				return Result{Status: Skipped, Message: "update did not succeed"}
			}
			if i.Resp.JSON == nil {
				return quote(fail("update response is not a JSON resource", i.Evidence()),
					"The response should include the fully-populated resource.")
			}
			var missing []string
			walkTop(rc.Resource.Schema, func(name string, s *discovery.Schema) {
				if s.ReadOnly {
					if _, ok := i.Resp.JSON[name]; !ok {
						missing = append(missing, name)
					}
				}
			})
			if len(missing) > 0 {
				return quote(failf("declared output-only field(s) absent from Update response: %s", strings.Join(missing, ", ")),
					"The response should include the fully-populated resource.")
			}
			return pass("update response fully populated")
		},
	})
}
