package checks

import (
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func init() {
	// AEP-101: OpenAPI document version.
	Register(Check{
		ID: "openapi-version-3-1", AEP: 101, Level: MUST, Static: true, PerAPI: true,
		Title: "OpenAPI document declares version 3.1.x",
		Run: func(rc *RunContext) Result {
			v := rc.Model.OpenAPIVersion
			if strings.HasPrefix(v, "3.1") {
				return pass("openapi " + v)
			}
			return quote(failf("openapi version is %q, want 3.1.x", v),
				"AEP-compliant APIs should document their HTTP APIs using version 3.1.x of the OpenAPI Specification.")
		},
	})

	// AEP-122: every resource carries a read-only path field.
	Register(Check{
		ID: "path-field-present", AEP: 122, Level: MUST, Static: true,
		Title: "Resource defines a path field",
		Run: func(rc *RunContext) Result {
			if rc.Resource.Schema.HasProp("path") {
				return pass("path field present")
			}
			return quote(fail("resource schema has no 'path' field"),
				"Every resource must include a path field containing the unique resource identifier.")
		},
	})
	Register(Check{
		ID: "path-field-readonly", AEP: 122, Level: MUST, Static: true,
		Title:      "path field is read-only (output only)",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Schema.HasProp("path") },
		Run: func(rc *RunContext) Result {
			if rc.Resource.Schema.Prop("path").ReadOnly {
				return pass("path is readOnly")
			}
			return quote(fail("'path' field is not marked readOnly"),
				"The path field is annotated readOnly: true (OpenAPI) / OUTPUT_ONLY (protobuf).")
		},
	})

	// AEP-127 / standard methods: HTTP verb mapping per method.
	verbChecks := []struct {
		id     string
		method discovery.Method
		verb   string
		level  Level
	}{
		{"get-verb-get", discovery.MethodGet, "GET", MUST},
		{"list-verb-get", discovery.MethodList, "GET", MUST},
		{"create-verb-post", discovery.MethodCreate, "POST", MUST},
		{"update-verb-patch", discovery.MethodUpdate, "PATCH", SHOULD},
		{"apply-verb-put", discovery.MethodApply, "PUT", MUST},
		{"delete-verb-delete", discovery.MethodDelete, "DELETE", MUST},
	}
	for _, vc := range verbChecks {
		Register(Check{
			ID: vc.id, AEP: 127, Level: vc.level, Static: true,
			Title:      string(vc.method) + " uses HTTP " + vc.verb,
			Applicable: func(rc *RunContext) bool { return rc.endpoint(vc.method) != nil },
			Run: func(rc *RunContext) Result {
				ep := rc.endpoint(vc.method)
				if ep.HTTPVerb == vc.verb {
					return pass(vc.verb + " " + ep.Path)
				}
				return failf("%s method uses %s, want %s", vc.method, ep.HTTPVerb, vc.verb)
			},
		})
	}
}
