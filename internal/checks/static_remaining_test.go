package checks

import (
	"testing"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func TestRemainingStaticChecks(t *testing.T) {
	book := func(pattern string) *discovery.Resource {
		return &discovery.Resource{Singular: "book", Plural: "books", Patterns: []string{pattern}}
	}

	// get-path-param-id-form
	if s := evalByID(t, "get-path-param-id-form", book("publishers/{publisher_id}/books/{book_id}")).Status; s != Pass {
		t.Errorf("id-form ({book_id}): got %v, want Pass", s)
	}
	if s := evalByID(t, "get-path-param-id-form", book("publishers/{publisher_id}/books/{id}")).Status; s != Fail {
		t.Errorf("id-form ({id}): got %v, want Fail", s)
	}

	// segment-charset-ascii
	if s := evalByID(t, "segment-charset-ascii", book("publishers/{publisher_id}/books/{book_id}")).Status; s != Pass {
		t.Errorf("charset (clean): got %v, want Pass", s)
	}
	if s := evalByID(t, "segment-charset-ascii", book("Publishers/{book_id}")).Status; s != Fail {
		t.Errorf("charset (uppercase): got %v, want Fail", s)
	}

	// create-no-required-query
	mkCreate := func(q ...discovery.Param) *discovery.Resource {
		return &discovery.Resource{Singular: "book", Plural: "books",
			Methods: map[discovery.Method]*discovery.Endpoint{
				discovery.MethodCreate: {HTTPVerb: "POST", QueryParams: q}}}
	}
	if s := evalByID(t, "create-no-required-query", mkCreate(discovery.Param{Name: "id", Required: false})).Status; s != Pass {
		t.Errorf("create-no-required-query (optional id): got %v, want Pass", s)
	}
	if s := evalByID(t, "create-no-required-query", mkCreate(discovery.Param{Name: "foo", Required: true})).Status; s != Fail {
		t.Errorf("create-no-required-query (required foo): got %v, want Fail", s)
	}

	// update-merge-patch-mime
	mkUpdate := func(cts ...string) *discovery.Resource {
		return &discovery.Resource{Singular: "book", Plural: "books",
			Methods: map[discovery.Method]*discovery.Endpoint{
				discovery.MethodUpdate: {HTTPVerb: "PATCH", RequestContentTypes: cts}}}
	}
	if s := evalByID(t, "update-merge-patch-mime", mkUpdate("application/merge-patch+json")).Status; s != Pass {
		t.Errorf("merge-patch (present): got %v, want Pass", s)
	}
	if s := evalByID(t, "update-merge-patch-mime", mkUpdate("application/json")).Status; s != Fail {
		t.Errorf("merge-patch (absent): got %v, want Fail", s)
	}

	// order-by-is-string
	mkOrderBy := func(typ string) *discovery.Resource {
		return &discovery.Resource{Singular: "book", Plural: "books",
			Features: discovery.FeatureFlags{OrderBy: true},
			Methods: map[discovery.Method]*discovery.Endpoint{
				discovery.MethodList: {HTTPVerb: "GET", QueryParams: []discovery.Param{
					{Name: "order_by", Schema: &discovery.Schema{Type: typ}}}}}}
	}
	if s := evalByID(t, "order-by-is-string", mkOrderBy("string")).Status; s != Pass {
		t.Errorf("order_by (string): got %v, want Pass", s)
	}
	if s := evalByID(t, "order-by-is-string", mkOrderBy("integer")).Status; s != Fail {
		t.Errorf("order_by (integer): got %v, want Fail", s)
	}

	// total-size-is-integer
	mkTotal := func(typ string) *discovery.Resource {
		return &discovery.Resource{Singular: "book", Plural: "books",
			Methods: map[discovery.Method]*discovery.Endpoint{
				discovery.MethodList: {HTTPVerb: "GET", Responses: map[string]*discovery.Response{
					"200": {Schema: schemaWith(map[string]*discovery.Schema{
						"results":    {Type: "array"},
						"total_size": {Type: typ},
					}, "results", "total_size")}}}}}
	}
	if s := evalByID(t, "total-size-is-integer", mkTotal("integer")).Status; s != Pass {
		t.Errorf("total_size (integer): got %v, want Pass", s)
	}
	if s := evalByID(t, "total-size-is-integer", mkTotal("string")).Status; s != Fail {
		t.Errorf("total_size (string): got %v, want Fail", s)
	}

	// user-settable-id-not-uuid
	mkID := func(format string) *discovery.Resource {
		return &discovery.Resource{Singular: "book", Plural: "books",
			Features: discovery.FeatureFlags{UserSettableID: true},
			Methods: map[discovery.Method]*discovery.Endpoint{
				discovery.MethodCreate: {HTTPVerb: "POST", QueryParams: []discovery.Param{
					{Name: "id", Schema: &discovery.Schema{Type: "string", Format: format}}}}}}
	}
	if s := evalByID(t, "user-settable-id-not-uuid", mkID("")).Status; s != Pass {
		t.Errorf("id-not-uuid (plain): got %v, want Pass", s)
	}
	if s := evalByID(t, "user-settable-id-not-uuid", mkID("uuid")).Status; s != Fail {
		t.Errorf("id-not-uuid (uuid): got %v, want Fail", s)
	}
}

func TestRemainingDynamicChecks(t *testing.T) {
	pattern := "publishers/{publisher_id}/books/{book_id}"
	res := &discovery.Resource{Singular: "book", Plural: "books", Patterns: []string{pattern},
		Schema: schemaWith(map[string]*discovery.Schema{
			"path":  {Type: "string", ReadOnly: true},
			"title": {Type: "string"},
		}, "path", "title"),
		Methods: map[discovery.Method]*discovery.Endpoint{
			discovery.MethodGet:    {HTTPVerb: "GET"},
			discovery.MethodUpdate: {HTTPVerb: "PATCH"},
		}}

	// path-matches-canonical-pattern
	pathProbe := func(p string) Result {
		pr := &Probe{Interactions: map[string]*Interaction{Get: inter(200, "", `{"path":"`+p+`"}`)}}
		return Evaluate(findCheck(t, "path-matches-canonical-pattern"), &RunContext{Model: &discovery.APIModel{}, Resource: res, Probe: pr})
	}
	if s := pathProbe("publishers/p1/books/b1").Status; s != Pass {
		t.Errorf("path-matches (canonical): got %v, want Pass", s)
	}
	if s := pathProbe("publishers/p1/chapters/b1").Status; s != Fail {
		t.Errorf("path-matches (wrong collection): got %v, want Fail", s)
	}
	if s := pathProbe("publishers/-/books/b1").Status; s != Fail {
		t.Errorf("path-matches (wildcard segment): got %v, want Fail", s)
	}

	// update-fully-populated
	updProbe := func(body string) Result {
		pr := &Probe{Interactions: map[string]*Interaction{Update: inter(200, "{}", body)}}
		return Evaluate(findCheck(t, "update-fully-populated"), &RunContext{Model: &discovery.APIModel{}, Resource: res, Probe: pr})
	}
	if s := updProbe(`{"path":"publishers/p1/books/b1","title":"t"}`).Status; s != Pass {
		t.Errorf("update-fully-populated (path present): got %v, want Pass", s)
	}
	if s := updProbe(`{"title":"t"}`).Status; s != Fail {
		t.Errorf("update-fully-populated (path absent): got %v, want Fail", s)
	}
}

func findCheck(t *testing.T, id string) Check {
	t.Helper()
	for _, c := range All() {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("check %q not registered", id)
	return Check{}
}
