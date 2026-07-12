package checks

import (
	"fmt"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func init() {
	// Create returns the created resource (2xx + JSON object body).
	Register(Check{
		ID: "create-returns-resource", AEP: 133, Level: MUST,
		Title:      "Create returns the created resource",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodCreate) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Create)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status >= 300 {
				return fail(fmt.Sprintf("create returned %d", i.Resp.Status), i.Evidence())
			}
			if i.Resp.JSON == nil {
				return fail("create response body is not a JSON object", i.Evidence())
			}
			return pass("create returned resource", i.Evidence())
		},
	})

	// Create should return 201 Created.
	Register(Check{
		ID: "create-201-created", AEP: 133, Level: SHOULD,
		Title:      "Create responds 201 Created",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodCreate) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Create)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 201 {
				return pass("201 Created")
			}
			return fail(fmt.Sprintf("create returned %d, expected 201", i.Resp.Status), i.Evidence())
		},
	})

	// Create populates the server-assigned path on the returned resource.
	Register(Check{
		ID: "create-populates-path", AEP: 133, Level: MUST,
		Title:      "Create response has a populated path",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodCreate) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Create)
			if skip != nil {
				return *skip
			}
			if p, ok := i.Resp.JSON["path"].(string); ok && p != "" {
				return pass("path = " + p)
			}
			return quote(fail("created resource has no populated 'path'", i.Evidence()),
				"Create methods return the created resource with the server-assigned path.")
		},
	})

	// Create echoes the input fields that were sent.
	Register(Check{
		ID: "create-echoes-input", AEP: 133, Level: MUST,
		Title:      "Create echoes provided input fields",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodCreate) != nil },
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, Create)
			if skip != nil {
				return *skip
			}
			if i.Resp.JSON == nil {
				return fail("no JSON body to compare", i.Evidence())
			}
			for k, sent := range rc.Probe.SentBody {
				got, ok := i.Resp.JSON[k]
				if !ok {
					return quote(failf("field %q sent on create is absent from the response", k),
						"The response must include any fields that were provided unless they are input only.")
				}
				if !jsonEqualish(sent, got) {
					return failf("field %q: sent %v, response has %v", k, sent, got)
				}
			}
			return pass("all sent fields echoed")
		},
	})

	// Duplicate create with the same id must conflict (ALREADY_EXISTS / 409).
	Register(Check{
		ID: "create-duplicate-already-exists", AEP: 133, Level: MUST,
		Title: "Duplicate Create returns 409 ALREADY_EXISTS",
		Applicable: func(rc *RunContext) bool {
			return rc.endpoint(discovery.MethodCreate) != nil && rc.Resource.Features.UserSettableID
		},
		Run: func(rc *RunContext) Result {
			i, skip := need(rc.Probe, DuplicateCreate)
			if skip != nil {
				return *skip
			}
			if i.Resp.Status == 409 {
				return pass("409 on duplicate id")
			}
			return quote(fail(fmt.Sprintf("duplicate create returned %d, expected 409", i.Resp.Status), i.Evidence()),
				"If a user tries to create a resource with an ID that would result in a duplicate resource path, the service must error with ALREADY_EXISTS.")
		},
	})
}

// jsonEqualish compares a sent value against a received JSON value, tolerating
// numeric type differences (int sent vs float64 decoded).
func jsonEqualish(a, b any) bool {
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		return af == bf
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}
