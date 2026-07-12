package checks

import (
	"regexp"
	"strings"
)

// AEP-136 custom-method requirements. These are static: they read the custom
// endpoints discovered on a resource (the ':verb' operations). Every check is
// gated on the resource actually declaring at least one custom method, so APIs
// without custom methods report NotApplicable rather than failing.

var (
	customVerbRE = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	// Prepositions that must not appear in a custom-method name (AEP-136).
	prepositions = map[string]bool{
		"with": true, "for": true, "to": true, "from": true, "by": true,
		"at": true, "in": true, "on": true, "of": true, "into": true,
		"onto": true, "over": true, "under": true, "across": true, "per": true,
	}
)

func hasCustom(rc *RunContext) bool { return len(rc.Resource.Custom) > 0 }

func init() {
	// AEP-136: the HTTP URI must use ':' followed by the custom verb, and the
	// verb in the URI must match the verb in the RPC name.
	Register(Check{
		ID: "custom-verb-in-uri", AEP: 136, Level: MUST, Static: true,
		Title:      "Custom methods place the verb after a ':' in the URI",
		Applicable: hasCustom,
		Run: func(rc *RunContext) Result {
			for _, ep := range rc.Resource.Custom {
				if ep.CustomVerb == "" || !strings.Contains(ep.Path, ":"+ep.CustomVerb) {
					return quote(failf("custom method %q has no ':verb' suffix in URI %q", ep.OperationID, ep.Path),
						"The HTTP URI must use a ':' character followed by the custom verb.")
				}
			}
			return pass("all custom methods use ':verb' URIs")
		},
	})

	// AEP-136: if word separators are necessary in the verb, snake_case must be
	// used (not kebab-case or camelCase).
	Register(Check{
		ID: "custom-verb-snake-case", AEP: 136, Level: MUST, Static: true,
		Title:      "Custom verbs use snake_case",
		Applicable: hasCustom,
		Run: func(rc *RunContext) Result {
			for _, ep := range rc.Resource.Custom {
				if !customVerbRE.MatchString(ep.CustomVerb) {
					return quote(failf("custom verb %q is not snake_case", ep.CustomVerb),
						"If word separators are necessary, snake_case must be used.")
				}
			}
			return pass("custom verbs are snake_case")
		},
	})

	// AEP-136: the verb in the URI must match the verb in the RPC name — the
	// operationId, lower-cased with separators stripped, must begin with the verb.
	Register(Check{
		ID: "custom-verb-matches-name", AEP: 136, Level: MUST, Static: true,
		Title:      "Custom URI verb matches the operationId verb",
		Applicable: hasCustom,
		Run: func(rc *RunContext) Result {
			for _, ep := range rc.Resource.Custom {
				if ep.OperationID == "" {
					continue // nothing to compare against
				}
				// Compare on alphanumerics only: operationIds in these specs may
				// carry a leading ':' and/or separators (e.g. ":ArchiveBook").
				opNorm := alnumLower(ep.OperationID)
				verbNorm := alnumLower(ep.CustomVerb)
				if !strings.HasPrefix(opNorm, verbNorm) {
					return quote(failf("operationId %q does not begin with URI verb %q", ep.OperationID, ep.CustomVerb),
						"The verb in the URI must match the verb in the name of the RPC.")
				}
			}
			return pass("URI verbs match operationId verbs")
		},
	})

	// AEP-136: the RPC name must not contain prepositions.
	Register(Check{
		ID: "custom-name-no-prepositions", AEP: 136, Level: MUST, Static: true,
		Title:      "Custom-method names contain no prepositions",
		Applicable: hasCustom,
		Run: func(rc *RunContext) Result {
			for _, ep := range rc.Resource.Custom {
				for _, w := range splitWords(ep.OperationID) {
					if prepositions[strings.ToLower(w)] {
						return quote(failf("custom method %q contains preposition %q", ep.OperationID, w),
							"The name of the RPC must not contain prepositions.")
					}
				}
			}
			return pass("no prepositions in custom-method names")
		},
	})

	// AEP-136: custom methods using GET or DELETE must omit the request body.
	Register(Check{
		ID: "custom-no-body-on-get-delete", AEP: 136, Level: MUST, Static: true,
		Title:      "GET/DELETE custom methods declare no request body",
		Applicable: hasCustom,
		Run: func(rc *RunContext) Result {
			for _, ep := range rc.Resource.Custom {
				if (ep.HTTPVerb == "GET" || ep.HTTPVerb == "DELETE") && ep.RequestBodySchema != nil {
					return quote(failf("custom method %q (%s) declares a request body", ep.OperationID, ep.HTTPVerb),
						"If using GET or DELETE, the body clause must be absent.")
				}
			}
			return pass("no body on GET/DELETE custom methods")
		},
	})

	// AEP-136: custom methods should not use PATCH or DELETE.
	Register(Check{
		ID: "custom-not-patch-delete", AEP: 136, Level: SHOULDNOT, Static: true,
		Title:      "Custom methods avoid PATCH and DELETE",
		Applicable: hasCustom,
		Run: func(rc *RunContext) Result {
			for _, ep := range rc.Resource.Custom {
				if ep.HTTPVerb == "PATCH" || ep.HTTPVerb == "DELETE" {
					return quote(failf("custom method %q uses %s", ep.OperationID, ep.HTTPVerb),
						"Custom methods should not use PATCH or DELETE.")
				}
			}
			return pass("custom methods use POST/GET/PUT")
		},
	})
}

// splitWords breaks an identifier (camelCase, snake_case, or kebab-case) into
// its component words.
func splitWords(id string) []string {
	// Normalize separators to spaces, then break camelCase boundaries.
	spaced := strings.NewReplacer("_", " ", "-", " ").Replace(id)
	var b strings.Builder
	var prev rune
	for i, r := range spaced {
		if i > 0 && isUpper(r) && !isUpper(prev) && prev != ' ' {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
		prev = r
	}
	return strings.Fields(b.String())
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }

// alnumLower keeps only ASCII letters and digits, lower-cased.
func alnumLower(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		}
	}
	return b.String()
}
