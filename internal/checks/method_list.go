package checks

import (
	"fmt"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func init() {
	// List returns 200.
	Register(Check{
		ID: "list-200", AEP: 132, Level: MUST,
		Title:      "List returns 200",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodList) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, List)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 200 {
				return pass("200 OK")
			}
			return fail(fmt.Sprintf("list returned %d, expected 200", i.Resp.Status), i.Evidence())
		},
	})

	// The resources array must be named "results" and be unwrapped.
	Register(Check{
		ID: "list-results-array-named-results", AEP: 132, Level: MUST,
		Title:      "List response array is named 'results'",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodList) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, List)
			if skip != nil {
				return *skip
			}
			if i.Resp.JSON == nil {
				return fail("list response is not a JSON object", i.Evidence())
			}
			if _, ok := i.Resp.JSON["results"].([]any); ok {
				return pass("results array present")
			}
			return quote(fail("response has no 'results' array (checked keys: "+jsonKeys(i.Resp.JSON)+")", i.Evidence()),
				"The array of resources must be named results and contain resources with no additional wrapping.")
		},
	})

	// The List 200 response schema must declare next_page_token (static).
	Register(Check{
		ID: "list-next-page-token-present", AEP: 132, Level: MUST, Static: true,
		Title: "List response schema declares next_page_token",
		Applicable: func(rc *RunContext) bool {
			ep := rc.endpoint(discovery.MethodList)
			return ep != nil && rc.Resource.Features.Pagination
		},
		Run: func(rc *RunContext) Result {
			ep := rc.endpoint(discovery.MethodList)
			resp := ep.Responses["200"]
			if resp != nil && resp.Schema != nil &&
				(resp.Schema.HasProp("next_page_token") || resp.Schema.HasProp("nextPageToken")) {
				return pass("next_page_token declared")
			}
			return quote(fail("List 200 response schema lacks next_page_token"),
				"The string nextPageToken field must be included in the list response schema.")
		},
	})

	// List controls (pagination, filter, ordering) must be query parameters, not
	// path variables. (Path variables legitimately span the parent hierarchy for
	// nested resources, so their count is not itself constrained.)
	Register(Check{
		ID: "list-controls-in-query", AEP: 132, Level: MUST, Static: true,
		Title:      "List controls are query parameters",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodList) != nil },
		Run: func(rc *RunContext) Result {
			ep := rc.endpoint(discovery.MethodList)
			controls := []string{"max_page_size", "page_size", "page_token", "filter", "order_by", "skip", "show_deleted", "read_mask"}
			for _, p := range ep.PathParams {
				for _, ctl := range controls {
					if p.Name == ctl {
						return quote(failf("list control %q is a path variable; must be a query parameter", ctl),
							"All remaining parameters must map to URI query parameters.")
					}
				}
			}
			return pass("list controls are query parameters")
		},
	})
}

func jsonKeys(m map[string]any) string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	return fmt.Sprintf("%v", ks)
}
