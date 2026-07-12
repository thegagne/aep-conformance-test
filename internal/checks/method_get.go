package checks

import (
	"fmt"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func init() {
	// Get on an existing resource returns 200.
	Register(Check{
		ID: "get-200-on-existing", AEP: 131, Level: MUST,
		Title:      "Get returns 200 for an existing resource",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodGet) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Get)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 200 {
				return pass("200 OK")
			}
			return fail(fmt.Sprintf("get returned %d, expected 200", i.Resp.Status), i.Evidence())
		},
	})

	// Get returns the same resource that was created (path matches).
	Register(Check{
		ID: "get-body-matches-created", AEP: 131, Level: MUST,
		Title:      "Get returns the created resource",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodGet) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Get)
			if skip != nil {
				return *skip
			}
			if i.Resp.JSON == nil {
				return fail("get response is not a JSON object", i.Evidence())
			}
			got, _ := i.Resp.JSON["path"].(string)
			if trimSlash(got) == trimSlash(rc.Probe.CreatedPath) {
				return pass("path matches created resource")
			}
			return failf("get path %q != created path %q", got, rc.Probe.CreatedPath)
		},
	})

	// Get on a missing resource returns 404.
	Register(Check{
		ID: "get-404-on-missing", AEP: 131, Level: MUST,
		Title:      "Get returns 404 for a missing resource",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodGet) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, GetMissing)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 404 {
				return pass("404 on missing")
			}
			return quote(fail(fmt.Sprintf("get missing returned %d, expected 404", i.Resp.Status), i.Evidence()),
				"If the user does have proper permission, but the requested resource does not exist, the service must reply with an HTTP 404 error.")
		},
	})

	// Get is safe: repeating it yields the same body (no side effects).
	Register(Check{
		ID: "get-safe-no-side-effects", AEP: 131, Level: MUST,
		Title:      "Get is safe and idempotent",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodGet) != nil },
		Run: func(rc *RunContext) Result {
			a, skip := need(rc.Probe, Get)
			if skip != nil {
				return *skip
			}
			b, skip := need(rc.Probe, GetRepeat)
			if skip != nil {
				return *skip
			}
			if a.Resp.Body == b.Resp.Body {
				return pass("repeated GET is stable")
			}
			return quote(fail("two identical GETs returned different bodies"),
				"The operation must not have side effects and must be safe.")
		},
	})
}

func trimSlash(s string) string {
	for len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	return s
}
