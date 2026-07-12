package checks

import (
	"strings"
)

// AEP-122 resource-path structure checks, derived statically from the discovered
// path patterns (e.g. "publishers/{publisher_id}/books/{book_id}").

func init() {
	// AEP-122: the path of the parent resource must be a prefix of the path of
	// the collection (and therefore of the child's canonical path).
	Register(Check{
		ID: "parent-path-is-prefix", AEP: 122, Level: MUST, Static: true,
		Title:      "Parent resource path is a prefix of the child path",
		Applicable: func(rc *RunContext) bool { return len(rc.Resource.Parents) > 0 },
		Run: func(rc *RunContext) Result {
			child := rc.Resource.CanonicalPattern()
			parent := rc.Model.ResourceBySingular(rc.Resource.Parents[0])
			if parent == nil {
				return na("parent resource not found in model")
			}
			pp := parent.CanonicalPattern()
			if pp == "" || child == "" {
				return na("missing path pattern")
			}
			if strings.HasPrefix(child, pp+"/") {
				return pass(pp + " ⊑ " + child)
			}
			return quote(failf("parent path %q is not a prefix of child path %q", pp, child),
				"The path of the parent resource must be a prefix to the path of the collection.")
		},
	})

	// AEP-122: resource-path components must alternate between collection
	// identifiers (literal segments) and resource IDs (variable segments).
	Register(Check{
		ID: "path-segments-alternate", AEP: 122, Level: MUST, Static: true,
		Title:      "Path alternates collection-id / resource-id segments",
		Applicable: func(rc *RunContext) bool { return rc.Resource.CanonicalPattern() != "" },
		Run: func(rc *RunContext) Result {
			pat := rc.Resource.CanonicalPattern()
			segs := strings.Split(strings.Trim(pat, "/"), "/")
			for i, s := range segs {
				isVar := strings.HasPrefix(s, "{")
				// Even segments must be literal collection ids; odd segments must
				// be variables. (Singletons legitimately end on a literal, which
				// this still allows — they simply stop at an even index.)
				if i%2 == 0 && isVar {
					return quote(failf("segment %d of %q is a variable; expected a collection identifier", i, pat),
						"Resource path components must alternate between collection identifiers and resource IDs.")
				}
				if i%2 == 1 && !isVar {
					return quote(failf("segment %d of %q is a literal; expected a resource ID variable", i, pat),
						"Resource path components must alternate between collection identifiers and resource IDs.")
				}
			}
			return pass("segments alternate: " + pat)
		},
	})

	// AEP-122: a resource path must be unique within an API — no two resources
	// may share the same canonical path pattern.
	Register(Check{
		ID: "resource-paths-unique", AEP: 122, Level: MUST, Static: true, PerAPI: true,
		Title: "Resource path patterns are unique across the API",
		Run: func(rc *RunContext) Result {
			seen := map[string]string{}
			for _, r := range rc.Model.Resources {
				pat := r.CanonicalPattern()
				if pat == "" {
					continue
				}
				if prev, dup := seen[pat]; dup {
					return quote(failf("resources %q and %q share the path pattern %q", prev, r.Singular, pat),
						"A resource path must be unique with an API, referring to a single resource.")
				}
				seen[pat] = r.Singular
			}
			return pass("all resource path patterns are unique")
		},
	})
}
