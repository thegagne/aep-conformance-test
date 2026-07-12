package checks

import (
	"fmt"
	"slices"
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// AEP-159 (reading across collections) has no static fingerprint in the OpenAPI
// document — the '-' wildcard is a runtime value passed into an ordinary parent
// path parameter, disclosed only in prose. So this check is behavioral: the
// runner issues a List with the immediate parent replaced by '-'. If the server
// rejects it, the optional feature is unsupported (N/A). If it is honored, the
// returned resources must carry canonical parent ids rather than the wildcard.

func init() {
	Register(Check{
		ID: "across-collections-real-parent-ids", AEP: 159, Level: MUST,
		Title: "Wildcard '-' list returns canonical parent ids",
		Applicable: func(rc *RunContext) bool {
			return len(rc.Resource.Parents) > 0 && rc.endpoint(discovery.MethodList) != nil
		},
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, ListAcrossCollections)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status < 200 || i.Resp.Status >= 300 {
				return na(fmt.Sprintf("reading across collections not supported (wildcard list returned %d)", i.Resp.Status))
			}
			results, ok := i.Resp.JSON["results"].([]any)
			if !ok || len(results) == 0 {
				return na("wildcard list returned no results to verify")
			}
			for _, item := range results {
				obj, ok := item.(map[string]any)
				if !ok {
					continue
				}
				path, _ := obj["path"].(string)
				if hasWildcardSegment(path) {
					return quote(failf("wildcard list result path %q still contains '-' instead of a real parent id", path),
						"The resources provided in the response must use the canonical name of the resource, with the actual parent collection identifiers (instead of `-`).")
				}
			}
			return pass(fmt.Sprintf("wildcard list returned %d resources with canonical parent ids", len(results)), i.Evidence())
		},
	})
}

// hasWildcardSegment reports whether any '/'-delimited segment of path is the
// bare '-' wildcard.
func hasWildcardSegment(path string) bool {
	return slices.Contains(strings.Split(path, "/"), "-")
}
