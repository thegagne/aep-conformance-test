package checks

import (
	"slices"
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func init() {
	// AEP-122: every literal path segment (collection identifier) must be
	// ASCII and free of characters that would require URL-escaping.
	Register(Check{
		ID: "segment-charset-ascii", AEP: 122, Level: MUST, Static: true,
		Title:      "Path segments are ASCII and URL-safe",
		Applicable: func(rc *RunContext) bool { return rc.Resource.CanonicalPattern() != "" },
		Run: func(rc *RunContext) Result {
			for _, seg := range strings.Split(rc.Resource.CanonicalPattern(), "/") {
				if seg == "" || isPatternVar(seg) {
					continue // variable id segments are placeholders
				}
				if !collectionIDRE.MatchString(seg) {
					return quote(failf("path segment %q is not lower-case ASCII / URL-safe", seg),
						"Resource paths must not use characters that require URL-escaping, or characters outside of ASCII.")
				}
			}
			return pass("all literal segments are URL-safe")
		},
	})

	// AEP-131: the resource-id path parameter must be named after the resource
	// singular (bare, or with an _id suffix), not a bare {id}.
	Register(Check{
		ID: "get-path-param-id-form", AEP: 131, Level: MUST, Static: true,
		Title:      "Resource-id path parameter is named for the resource",
		Applicable: func(rc *RunContext) bool { return lastPatternVar(rc.Resource.CanonicalPattern()) != "" },
		Run: func(rc *RunContext) Result {
			v := lastPatternVar(rc.Resource.CanonicalPattern())
			norm := strings.ReplaceAll(strings.ToLower(rc.Resource.Singular), "-", "_")
			if strings.TrimSuffix(v, "_id") == norm {
				return pass("id path parameter: {" + v + "}")
			}
			return quote(failf("resource-id path parameter is {%s}, expected {%s} or {%s_id}", v, norm, norm),
				"The path parameter for all resource IDs must be in the form {resource-singular}.")
		},
	})

	// AEP-133: a Create request must not require any query fields (the id is
	// optional and the parent travels in the path).
	Register(Check{
		ID: "create-no-required-query", AEP: 133, Level: MUSTNOT, Static: true,
		Title:      "Create requires no query fields",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodCreate) != nil },
		Run: func(rc *RunContext) Result {
			for _, q := range rc.endpoint(discovery.MethodCreate).QueryParams {
				if q.Required {
					return quote(failf("Create requires query field %q", q.Name),
						"The request message must not contain any other required fields.")
				}
			}
			return pass("no required query fields on Create")
		},
	})

	// AEP-134: Update must accept the application/merge-patch+json media type.
	Register(Check{
		ID: "update-merge-patch-mime", AEP: 134, Level: MUST, Static: true,
		Title:      "Update accepts application/merge-patch+json",
		Applicable: func(rc *RunContext) bool { return rc.endpoint(discovery.MethodUpdate) != nil },
		Run: func(rc *RunContext) Result {
			ct := rc.endpoint(discovery.MethodUpdate).RequestContentTypes
			if slices.Contains(ct, "application/merge-patch+json") {
				return pass("declares application/merge-patch+json")
			}
			return quote(failf("Update request media types are %v; want application/merge-patch+json", ct),
				"The method must support the MIME type application/merge-patch+json.")
		},
	})

	// AEP-132: an order_by field, when offered, must be a string.
	Register(Check{
		ID: "order-by-is-string", AEP: 132, Level: SHOULD, Static: true,
		Title:      "List order_by is a string",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.OrderBy },
		Run: func(rc *RunContext) Result {
			q := rc.endpoint(discovery.MethodList).QueryParam("order_by")
			if q == nil {
				return na("no order_by parameter")
			}
			if q.Schema != nil && q.Schema.Type != "" && q.Schema.Type != "string" {
				return quote(failf("order_by is typed %q, want string", q.Schema.Type),
					"Request message should contain a string order_by field.")
			}
			return pass("order_by is a string")
		},
	})

	// AEP-132: a total_size field, when offered, is an integer.
	Register(Check{
		ID: "total-size-is-integer", AEP: 132, Level: MAY, Static: true,
		Title:      "List total_size is an integer",
		Applicable: func(rc *RunContext) bool { return listTotalField(rc.Resource) != "" },
		Run: func(rc *RunContext) Result {
			name := listTotalField(rc.Resource)
			sc := rc.endpoint(discovery.MethodList).Responses["200"].Schema.Prop(name)
			if sc != nil && sc.Type != "" && sc.Type != "integer" {
				return quote(failf("%s is typed %q, want integer", name, sc.Type),
					"The response struct may include a int32 total_size field.")
			}
			return pass(name + " is an integer")
		},
	})

	// AEP-122: user-settable resource IDs should not be UUIDs.
	Register(Check{
		ID: "user-settable-id-not-uuid", AEP: 122, Level: SHOULDNOT, Static: true,
		Title:      "User-settable id is not a UUID",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Features.UserSettableID },
		Run: func(rc *RunContext) Result {
			q := rc.endpoint(discovery.MethodCreate).QueryParam("id")
			if q == nil || q.Schema == nil {
				return na("no id parameter schema")
			}
			if q.Schema.Format == "uuid" {
				return quote(fail("user-settable id is declared as a UUID (format: uuid)"),
					"User-settable IDs should not be permitted to be a UUID.")
			}
			return pass("user-settable id is not a UUID")
		},
	})
}

// isPatternVar reports whether a path segment is a {variable} placeholder.
func isPatternVar(seg string) bool {
	return strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}")
}

// lastPatternVar returns the name of the final {variable} in a path pattern
// (without braces), or "" when the pattern has no variable.
func lastPatternVar(pattern string) string {
	segs := strings.Split(pattern, "/")
	for i := len(segs) - 1; i >= 0; i-- {
		if isPatternVar(segs[i]) {
			return segs[i][1 : len(segs[i])-1]
		}
	}
	return ""
}

// listTotalField returns the name of the List response's total-count field
// (total_size or total), or "" when the response declares neither.
func listTotalField(r *discovery.Resource) string {
	list := r.Method(discovery.MethodList)
	if list == nil {
		return ""
	}
	resp := list.Responses["200"]
	if resp == nil || resp.Schema == nil {
		return ""
	}
	for _, name := range []string{"total_size", "total"} {
		if resp.Schema.HasProp(name) {
			return name
		}
	}
	return ""
}
