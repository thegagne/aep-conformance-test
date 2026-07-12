package checks

import (
	"slices"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func init() {
	// AEP-127: GET and DELETE operations must not declare a request body.
	for _, m := range []discovery.Method{discovery.MethodGet, discovery.MethodList, discovery.MethodDelete} {
		method := m
		Register(Check{
			ID: "no-body-" + string(method), AEP: 127, Level: MUST, Static: true,
			Title:      string(method) + " declares no request body",
			Applicable: func(rc *RunContext) bool { return rc.endpoint(method) != nil },
			Run: func(rc *RunContext) Result {
				ep := rc.endpoint(method)
				if ep.RequestBodySchema == nil {
					return pass("no request body")
				}
				return quote(fail(string(method)+" declares a request body"),
					"RPCs must not define a body at all for RPCs that use the GET or DELETE HTTP verbs.")
			},
		})
	}

	// AEP-131/133/135: Get/Create/Delete must not require extra query fields
	// beyond the resource path / parent.
	for _, m := range []discovery.Method{discovery.MethodGet, discovery.MethodDelete} {
		method := m
		Register(Check{
			ID: "no-extra-required-query-" + string(method), AEP: 131, Level: MUSTNOT, Static: true,
			Title:      string(method) + " requires no extra query fields",
			Applicable: func(rc *RunContext) bool { return rc.endpoint(method) != nil },
			Run: func(rc *RunContext) Result {
				ep := rc.endpoint(method)
				for _, q := range ep.QueryParams {
					if q.Required {
						return quote(failf("%s requires query field %q", method, q.Name),
							"The request must not require any fields in the query string beyond the resource path.")
					}
				}
				return pass("no required query fields")
			},
		})
	}

	// AEP-135: cascading delete of a resource with children requires a force flag.
	Register(Check{
		ID: "delete-cascade-requires-force", AEP: 135, Level: MUST, Static: true,
		Title: "Delete of a parent resource offers a force flag",
		Applicable: func(rc *RunContext) bool {
			return rc.endpoint(discovery.MethodDelete) != nil && hasChildren(rc)
		},
		Run: func(rc *RunContext) Result {
			if rc.Resource.Features.Force {
				return pass("force flag present")
			}
			return quote(fail("resource has children but Delete has no 'force' flag"),
				"If an API allows deletion of a resource that may have child resources, the API must provide a bool force field.")
		},
	})
}

// hasChildren reports whether any other resource names this one as a parent.
func hasChildren(rc *RunContext) bool {
	for _, other := range rc.Model.Resources {
		if slices.Contains(other.Parents, rc.Resource.Singular) {
			return true
		}
	}
	return false
}
