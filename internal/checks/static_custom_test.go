package checks

import (
	"testing"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// evalByID runs the single registered check with the given ID against a
// one-resource model and returns its result.
func evalByID(t *testing.T, id string, r *discovery.Resource) Result {
	t.Helper()
	m := &discovery.APIModel{Resources: []*discovery.Resource{r}}
	for _, c := range All() {
		if c.ID == id {
			return Evaluate(c, &RunContext{Model: m, Resource: r})
		}
	}
	t.Fatalf("check %q not registered", id)
	return Result{}
}

// resourceWithCustom builds a minimal resource carrying one custom endpoint.
func resourceWithCustom(ep *discovery.Endpoint) *discovery.Resource {
	return &discovery.Resource{
		Singular: "book", Plural: "books",
		Patterns: []string{"publishers/{publisher_id}/books/{book_id}"},
		Methods:  map[discovery.Method]*discovery.Endpoint{},
		Custom:   []*discovery.Endpoint{ep},
	}
}

func TestCustomMethodChecks(t *testing.T) {
	good := &discovery.Endpoint{
		HTTPVerb: "POST", Custom: true, CustomVerb: "archive",
		Path: "/publishers/{publisher_id}/books/{book_id}:archive", OperationID: ":ArchiveBook",
	}
	cases := []struct {
		name, id string
		ep       *discovery.Endpoint
		want     Status
	}{
		{"valid custom method passes", "custom-verb-matches-name", good, Pass},
		{"verb/name mismatch fails", "custom-verb-matches-name",
			&discovery.Endpoint{HTTPVerb: "POST", CustomVerb: "archive", Path: "/b:archive", OperationID: "MoveBook"}, Fail},
		{"kebab verb fails snake-case", "custom-verb-snake-case",
			&discovery.Endpoint{HTTPVerb: "POST", CustomVerb: "sign-off", Path: "/b:sign-off", OperationID: "SignOffBook"}, Fail},
		{"preposition in name fails", "custom-name-no-prepositions",
			&discovery.Endpoint{HTTPVerb: "POST", CustomVerb: "checkout", Path: "/b:checkout", OperationID: "CheckOutForBook"}, Fail},
		{"GET with body fails", "custom-no-body-on-get-delete",
			&discovery.Endpoint{HTTPVerb: "GET", CustomVerb: "peek", Path: "/b:peek", OperationID: "PeekBook",
				RequestBodySchema: &discovery.Schema{Type: "object"}}, Fail},
		{"PATCH custom method warns", "custom-not-patch-delete",
			&discovery.Endpoint{HTTPVerb: "PATCH", CustomVerb: "tweak", Path: "/b:tweak", OperationID: "TweakBook"}, Fail},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := evalByID(t, tc.id, resourceWithCustom(tc.ep)).Status
			if got != tc.want {
				t.Errorf("%s: got %v, want %v", tc.id, got, tc.want)
			}
		})
	}
}

func TestCollectionIDAndIDFieldChecks(t *testing.T) {
	// A resource whose final collection identifier is "books" (plural, kebab).
	ok := &discovery.Resource{
		Singular: "book", Plural: "books",
		Patterns: []string{"publishers/{publisher_id}/books/{book_id}"},
		Methods: map[discovery.Method]*discovery.Endpoint{
			discovery.MethodGet: {HTTPVerb: "GET", Path: "/publishers/{publisher_id}/books/{book_id}",
				PathParams: []discovery.Param{{Name: "book_id", In: "path", Schema: &discovery.Schema{Type: "string"}}}},
		},
	}
	if s := evalByID(t, "collection-id-format", ok).Status; s != Pass {
		t.Errorf("collection-id-format on valid resource: got %v, want Pass", s)
	}
	if s := evalByID(t, "collection-id-plural", ok).Status; s != Pass {
		t.Errorf("collection-id-plural on valid resource: got %v, want Pass", s)
	}
	if s := evalByID(t, "id-fields-are-strings", ok).Status; s != Pass {
		t.Errorf("id-fields-are-strings on valid resource: got %v, want Pass", s)
	}

	// Non-plural collection identifier ("book"): must fail the plural check.
	nonPlural := &discovery.Resource{
		Singular: "book", Plural: "books",
		Patterns: []string{"publishers/{publisher_id}/book/{book_id}"},
	}
	if s := evalByID(t, "collection-id-plural", nonPlural).Status; s != Fail {
		t.Errorf("collection-id-plural on singular id: got %v, want Fail", s)
	}

	// Integer ID path param: must fail id-fields-are-strings.
	intID := &discovery.Resource{
		Singular: "book", Plural: "books",
		Patterns: []string{"publishers/{publisher_id}/books/{book_id}"},
		Methods: map[discovery.Method]*discovery.Endpoint{
			discovery.MethodGet: {HTTPVerb: "GET", Path: "/books/{book_id}",
				PathParams: []discovery.Param{{Name: "book_id", In: "path", Schema: &discovery.Schema{Type: "integer"}}}},
		},
	}
	if s := evalByID(t, "id-fields-are-strings", intID).Status; s != Fail {
		t.Errorf("id-fields-are-strings on integer id: got %v, want Fail", s)
	}
}
