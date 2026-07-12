package checks

import (
	"testing"

	"github.com/thegagne/aep-conformance-test/internal/client"
	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// evalInModel runs the check with the given ID against a specific resource
// inside a multi-resource model (needed for parent/PerAPI checks).
func evalInModel(t *testing.T, id string, m *discovery.APIModel, r *discovery.Resource) Result {
	t.Helper()
	for _, c := range All() {
		if c.ID == id {
			return Evaluate(c, &RunContext{Model: m, Resource: r})
		}
	}
	t.Fatalf("check %q not registered", id)
	return Result{}
}

func schemaWith(props map[string]*discovery.Schema, order ...string) *discovery.Schema {
	return &discovery.Schema{Properties: &discovery.OrderedSchemas{Order: order, Map: props}}
}

func TestHierarchyChecks(t *testing.T) {
	publisher := &discovery.Resource{Singular: "publisher", Plural: "publishers",
		Patterns: []string{"publishers/{publisher_id}"}}
	book := &discovery.Resource{Singular: "book", Plural: "books", Parents: []string{"publisher"},
		Patterns: []string{"publishers/{publisher_id}/books/{book_id}"}}
	orphan := &discovery.Resource{Singular: "book", Plural: "books", Parents: []string{"publisher"},
		Patterns: []string{"stores/{store_id}/books/{book_id}"}}
	m := &discovery.APIModel{Resources: []*discovery.Resource{publisher, book}}

	if s := evalInModel(t, "parent-path-is-prefix", m, book).Status; s != Pass {
		t.Errorf("parent-path-is-prefix (valid): got %v, want Pass", s)
	}
	if s := evalInModel(t, "parent-path-is-prefix", m, orphan).Status; s != Fail {
		t.Errorf("parent-path-is-prefix (mismatch): got %v, want Fail", s)
	}

	// Alternation: good pattern passes, a literal in a variable slot fails.
	if s := evalInModel(t, "path-segments-alternate", m, book).Status; s != Pass {
		t.Errorf("path-segments-alternate (valid): got %v, want Pass", s)
	}
	bad := &discovery.Resource{Singular: "book", Plural: "books",
		Patterns: []string{"publishers/books/{book_id}"}}
	if s := evalInModel(t, "path-segments-alternate", m, bad).Status; s != Fail {
		t.Errorf("path-segments-alternate (bad): got %v, want Fail", s)
	}

	// Uniqueness (PerAPI): two resources sharing a pattern fails.
	dupA := &discovery.Resource{Singular: "a", Patterns: []string{"widgets/{widget_id}"}}
	dupB := &discovery.Resource{Singular: "b", Patterns: []string{"widgets/{widget_id}"}}
	dupModel := &discovery.APIModel{Resources: []*discovery.Resource{dupA, dupB}}
	if s := evalInModel(t, "resource-paths-unique", dupModel, dupA).Status; s != Fail {
		t.Errorf("resource-paths-unique (dup): got %v, want Fail", s)
	}
	if s := evalInModel(t, "resource-paths-unique", m, publisher).Status; s != Pass {
		t.Errorf("resource-paths-unique (unique): got %v, want Pass", s)
	}
}

func TestStateAndSoftDeleteChecks(t *testing.T) {
	// state field: enum + readOnly passes both checks.
	good := &discovery.Resource{Singular: "book", Plural: "books",
		Schema: schemaWith(map[string]*discovery.Schema{
			"state": {Type: "string", Enum: []string{"ACTIVE", "DELETED"}, ReadOnly: true},
		}, "state")}
	if s := evalByID(t, "state-field-is-enum", good).Status; s != Pass {
		t.Errorf("state-field-is-enum (valid): got %v, want Pass", s)
	}
	if s := evalByID(t, "state-field-output-only", good).Status; s != Pass {
		t.Errorf("state-field-output-only (valid): got %v, want Pass", s)
	}

	// state field without enum / not readOnly fails.
	bad := &discovery.Resource{Singular: "book", Plural: "books",
		Schema: schemaWith(map[string]*discovery.Schema{
			"state": {Type: "string"},
		}, "state")}
	if s := evalByID(t, "state-field-is-enum", bad).Status; s != Fail {
		t.Errorf("state-field-is-enum (no enum): got %v, want Fail", s)
	}
	if s := evalByID(t, "state-field-output-only", bad).Status; s != Fail {
		t.Errorf("state-field-output-only (writable): got %v, want Fail", s)
	}

	// Resource with no state field: checks are N/A.
	none := &discovery.Resource{Singular: "book", Plural: "books",
		Schema: schemaWith(map[string]*discovery.Schema{"title": {Type: "string"}}, "title")}
	if s := evalByID(t, "state-field-is-enum", none).Status; s != NotApplicable {
		t.Errorf("state-field-is-enum (absent): got %v, want N/A", s)
	}

	// show_deleted typed as string fails; boolean passes.
	softBad := &discovery.Resource{Singular: "book", Plural: "books",
		Features: discovery.FeatureFlags{ShowDeleted: true},
		Methods: map[discovery.Method]*discovery.Endpoint{
			discovery.MethodList: {HTTPVerb: "GET", QueryParams: []discovery.Param{
				{Name: "show_deleted", In: "query", Schema: &discovery.Schema{Type: "string"}}}},
		}}
	if s := evalByID(t, "soft-delete-show-deleted-bool", softBad).Status; s != Fail {
		t.Errorf("soft-delete-show-deleted-bool (string): got %v, want Fail", s)
	}
	softOK := &discovery.Resource{Singular: "book", Plural: "books",
		Features: discovery.FeatureFlags{ShowDeleted: true},
		Methods: map[discovery.Method]*discovery.Endpoint{
			discovery.MethodList: {HTTPVerb: "GET", QueryParams: []discovery.Param{
				{Name: "show_deleted", In: "query", Schema: &discovery.Schema{Type: "boolean"}}}},
		}}
	if s := evalByID(t, "soft-delete-show-deleted-bool", softOK).Status; s != Pass {
		t.Errorf("soft-delete-show-deleted-bool (bool): got %v, want Pass", s)
	}
}

func TestAcrossCollectionsCheck(t *testing.T) {
	book := &discovery.Resource{Singular: "book", Plural: "books", Parents: []string{"publisher"},
		Methods: map[discovery.Method]*discovery.Endpoint{discovery.MethodList: {HTTPVerb: "GET"}}}

	// Build a probe carrying a single wildcard-list interaction with the given
	// status and body.
	probeWith := func(status int, body string) *Probe {
		resp := &client.Response{Status: status}
		if len(body) > 0 && body[0] == '{' {
			resp.JSON = parseJSONObject(body)
		}
		return &Probe{Interactions: map[string]*Interaction{
			ListAcrossCollections: {Name: ListAcrossCollections, Resp: resp},
		}}
	}
	eval := func(p *Probe) Result {
		for _, c := range All() {
			if c.ID == "across-collections-real-parent-ids" {
				return Evaluate(c, &RunContext{Model: &discovery.APIModel{}, Resource: book, Probe: p})
			}
		}
		t.Fatal("check not registered")
		return Result{}
	}

	// Server rejects the wildcard → feature unsupported → N/A.
	if s := eval(probeWith(400, "")).Status; s != NotApplicable {
		t.Errorf("rejected wildcard: got %v, want N/A", s)
	}
	// Honored, canonical parent ids → Pass.
	ok := `{"results":[{"path":"publishers/p1/books/b1"},{"path":"publishers/p2/books/b2"}]}`
	if s := eval(probeWith(200, ok)).Status; s != Pass {
		t.Errorf("canonical ids: got %v, want Pass", s)
	}
	// Honored but a result still echoes '-' → Fail (the AEP-159 MUST).
	bad := `{"results":[{"path":"publishers/-/books/b1"}]}`
	if s := eval(probeWith(200, bad)).Status; s != Fail {
		t.Errorf("wildcard echoed in path: got %v, want Fail", s)
	}
	// Honored but empty → nothing to verify → N/A.
	if s := eval(probeWith(200, `{"results":[]}`)).Status; s != NotApplicable {
		t.Errorf("empty results: got %v, want N/A", s)
	}
}
