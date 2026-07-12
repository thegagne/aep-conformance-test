package checks

import (
	"fmt"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func init() {
	// Update returns 2xx and the resource path is unchanged.
	Register(Check{
		ID: "update-path-unchanged", AEP: 134, Level: MUST,
		Title:      "Update does not change the resource path",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodUpdate) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Update)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status >= 300 {
				return fail(fmt.Sprintf("update returned %d", i.Resp.Status), i.Evidence())
			}
			if i.Resp.JSON != nil {
				if p, ok := i.Resp.JSON["path"].(string); ok && trimSlash(p) != trimSlash(rc.Probe.CreatedPath) {
					return failf("update changed path: %q != %q", p, rc.Probe.CreatedPath)
				}
			}
			return pass("path unchanged after update")
		},
	})

	// Update must have strong consistency: a subsequent Get reflects the change.
	Register(Check{
		ID: "update-strong-consistency", AEP: 134, Level: MUST,
		Title: "Update is strongly consistent (reflected in Update response)",
		Applicable: func(rc *RunContext) bool {
			return rc.endpoint(discovery.MethodUpdate) != nil
		},
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Update)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status >= 300 || i.Resp.JSON == nil {
				return fail("update did not return the updated resource", i.Evidence())
			}
			// The runner sent exactly one changed field; confirm it echoes back.
			var reqBody map[string]any
			if len(i.ReqBody) > 0 {
				reqBody = parseJSONObject(i.ReqBody)
			}
			for k, want := range reqBody {
				if k == "update_mask" {
					continue
				}
				if got, ok := i.Resp.JSON[k]; !ok || !jsonEqualish(want, got) {
					return quote(failf("updated field %q not reflected (sent %v, got %v)", k, want, i.Resp.JSON[k]),
						"The response must include any fields that were sent and included in the update mask unless they are input only.")
				}
			}
			return pass("update reflected in response")
		},
	})

	// Apply (PUT) returns the resource.
	Register(Check{
		ID: "apply-response-is-resource", AEP: 137, Level: MUST,
		Title:      "Apply returns the resource",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodApply) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Apply)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status >= 300 {
				return fail(fmt.Sprintf("apply returned %d", i.Resp.Status), i.Evidence())
			}
			if i.Resp.JSON == nil {
				return fail("apply response is not a JSON object", i.Evidence())
			}
			return pass("apply returned resource")
		},
	})

	// Delete returns 204 No Content (or 200 with the resource for soft delete).
	Register(Check{
		ID: "delete-204-empty", AEP: 135, Level: SHOULD,
		Title:      "Delete responds 204 No Content",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodDelete) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Delete)
			if skip != nil {
				return *skip
			}
			switch {
			case i.Resp.Status == 204:
				return pass("204 No Content")
			case i.Resp.Status == 200:
				return pass("200 OK (soft delete returns resource)")
			default:
				return fail(fmt.Sprintf("delete returned %d, expected 204", i.Resp.Status), i.Evidence())
			}
		},
	})

	// Delete must have strong consistency: Get afterward returns 404.
	Register(Check{
		ID: "delete-strong-consistency-404", AEP: 135, Level: MUST,
		Title: "Get after Delete returns 404 (steady state)",
		Applicable: func(rc *RunContext) bool {
			return rc.endpoint(discovery.MethodDelete) != nil && rc.endpoint(discovery.MethodGet) != nil
		},
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, GetAfterDelete)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 404 {
				return pass("404 after delete")
			}
			return quote(fail(fmt.Sprintf("get after delete returned %d, expected 404", i.Resp.Status), i.Evidence()),
				"The completion of a delete method must mean that reading resource state returns a consistent 404 Not found response.")
		},
	})

	// Delete on a missing resource returns 404.
	Register(Check{
		ID: "delete-404-on-missing", AEP: 135, Level: MUST,
		Title:      "Delete returns 404 for a missing resource",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodDelete) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, DeleteMissing)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 404 {
				return pass("404 on missing delete")
			}
			return quote(fail(fmt.Sprintf("delete missing returned %d, expected 404", i.Resp.Status), i.Evidence()),
				"If the user does have proper permission, but the requested resource does not exist, the service must error with 404 Not found.")
		},
	})

	// AEP-134: updating a resource that doesn't exist returns NOT_FOUND (404).
	Register(Check{
		ID: "update-404-on-missing", AEP: 134, Level: MUST,
		Title:      "Update returns 404 for a missing resource",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodUpdate) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, UpdateMissing)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 404 {
				return pass("404 on missing update")
			}
			return quote(fail(fmt.Sprintf("update missing returned %d, expected 404", i.Resp.Status), i.Evidence()),
				"If the user does have proper permission but resource doesn't exist, service must error with NOT_FOUND.")
		},
	})

	// AEP-135: a DELETE that carries a request body must ignore the body and must
	// not error because of it.
	Register(Check{
		ID: "delete-body-ignored", AEP: 135, Level: MUST,
		Title:      "Delete ignores a request body (does not error)",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodDelete) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, DeleteWithBody)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status < 300 {
				return pass(fmt.Sprintf("delete-with-body succeeded (%d)", i.Resp.Status))
			}
			return quote(fail(fmt.Sprintf("delete with a body returned %d; the body must be ignored, not cause an error", i.Resp.Status), i.Evidence()),
				"If a delete request contains a body, the body must be ignored, and must not cause an error.")
		},
	})
}
