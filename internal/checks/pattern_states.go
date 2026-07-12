package checks

import (
	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// Conditional pattern checks for soft delete (AEP-164) and resource states
// (AEP-216). Each is gated so APIs that don't opt into the pattern report
// NotApplicable rather than failing.

func init() {
	// AEP-164: when soft delete is offered, List exposes a boolean show_deleted
	// query parameter to opt into returning soft-deleted resources.
	Register(Check{
		ID: "soft-delete-show-deleted-bool", AEP: 164, Level: SHOULD, Static: true,
		Title:      "List's show_deleted is a boolean",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.ShowDeleted },
		Run: func(rc *RunContext) Result {
			list := rc.endpoint(discovery.MethodList)
			if list == nil {
				return na("no List method")
			}
			q := list.QueryParam("show_deleted")
			if q == nil {
				return na("no show_deleted parameter")
			}
			if q.Schema != nil && q.Schema.Type != "" && q.Schema.Type != "boolean" {
				return quote(failf("show_deleted is typed %q, want boolean", q.Schema.Type),
					"APIs that support soft delete should include a bool show_deleted field on the List request.")
			}
			return pass("show_deleted is boolean")
		},
	})

	// AEP-216: a resource's state field, when present, must be an enumeration.
	Register(Check{
		ID: "state-field-is-enum", AEP: 216, Level: SHOULD, Static: true,
		Title:      "state field is an enum",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Schema.HasProp("state") },
		Run: func(rc *RunContext) Result {
			s := rc.Resource.Schema.Prop("state")
			if len(s.Enum) > 0 {
				return pass("state is an enum")
			}
			return quote(fail("'state' field declares no enum values"),
				"The state field should be defined as an enum.")
		},
	})

	// AEP-216: the state field is server-owned and must be output only.
	Register(Check{
		ID: "state-field-output-only", AEP: 216, Level: SHOULD, Static: true,
		Title:      "state field is output only",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Schema.HasProp("state") },
		Run: func(rc *RunContext) Result {
			if rc.Resource.Schema.Prop("state").ReadOnly {
				return pass("state is readOnly")
			}
			return quote(fail("'state' field is not marked readOnly"),
				"The state field should be output only, as it is owned by the service.")
		},
	})
}
