package checks

import (
	"regexp"
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// collectionIDRE matches a valid collection identifier: begins with a
// lower-case letter, then lower-case ASCII letters, digits, and hyphens.
var collectionIDRE = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func init() {
	// AEP-122: collection identifiers must be kebab-case — begin with a
	// lower-case letter and contain only lower-case ASCII letters, digits,
	// and hyphens.
	Register(Check{
		ID: "collection-id-format", AEP: 122, Level: MUST, Static: true,
		Title:      "Collection identifier is lower-case kebab-case",
		Applicable: func(rc *RunContext) bool { return rc.Resource.CollectionID() != "" },
		Run: func(rc *RunContext) Result {
			id := rc.Resource.CollectionID()
			if collectionIDRE.MatchString(id) {
				return pass("collection id: " + id)
			}
			return quote(failf("collection identifier %q is not lower-case kebab-case", id),
				"Collection identifiers must begin with a lower-cased letter and contain only lower-case ASCII letters, numbers, and hyphens.")
		},
	})

	// AEP-122: collection identifiers must be plural. Reported conservatively:
	// a failure requires a strong signal (the identifier equals the resource's
	// own singular name), so irregular plurals never produce a false positive.
	Register(Check{
		ID: "collection-id-plural", AEP: 122, Level: MUST, Static: true,
		Title:      "Collection identifier is plural",
		Applicable: func(rc *RunContext) bool { return rc.Resource.CollectionID() != "" },
		Run: func(rc *RunContext) Result {
			id := strings.ToLower(rc.Resource.CollectionID())
			singular := strings.ToLower(rc.Resource.Singular)
			plural := strings.ToLower(rc.Resource.Plural)
			if plural != "" && id == plural {
				return pass("collection id matches declared plural: " + id)
			}
			if singular != "" && id == singular {
				return quote(failf("collection identifier %q equals the singular; it must be plural", id),
					"Collection identifiers must be plural.")
			}
			return pass("collection id: " + id)
		},
	})

	// AEP-122: all ID fields must be strings — i.e. every path parameter that
	// stands in for a resource ID is typed as string.
	Register(Check{
		ID: "id-fields-are-strings", AEP: 122, Level: MUST, Static: true,
		Title: "Resource ID path parameters are strings",
		Run: func(rc *RunContext) Result {
			for _, ep := range allEndpoints(rc.Resource) {
				for _, p := range ep.PathParams {
					if p.Schema == nil || p.Schema.Type == "" {
						continue // untyped in the spec; nothing to assert
					}
					if p.Schema.Type != "string" {
						return quote(failf("path parameter %q on %s is typed %q, want string", p.Name, ep.Path, p.Schema.Type),
							"All ID fields must be strings.")
					}
				}
			}
			return pass("all ID path parameters are strings")
		},
	})
}

// allEndpoints returns every standard and custom endpoint on a resource.
func allEndpoints(r *discovery.Resource) []*discovery.Endpoint {
	var eps []*discovery.Endpoint
	for _, ep := range r.Methods {
		eps = append(eps, ep)
	}
	return append(eps, r.Custom...)
}
