package checks

import (
	"testing"

	"github.com/thegagne/aep-conformance-test/internal/client"
	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func applyResource() *discovery.Resource {
	schema := schemaWith(map[string]*discovery.Schema{
		"path":   {Type: "string", ReadOnly: true},
		"title":  {Type: "string"},
		"author": {Type: "string"},
	}, "path", "title", "author")
	schema.Required = []string{"title"}
	return &discovery.Resource{Singular: "book", Plural: "books", Schema: schema,
		Methods: map[discovery.Method]*discovery.Endpoint{
			discovery.MethodApply: {HTTPVerb: "PUT"},
			discovery.MethodGet:   {HTTPVerb: "GET"},
		}}
}

func inter(status int, reqBody, respJSON string) *Interaction {
	r := &client.Response{Status: status}
	if respJSON != "" {
		r.JSON = parseJSONObject(respJSON)
	}
	return &Interaction{Resp: r, ReqBody: reqBody}
}

func evalApply(t *testing.T, id string, r *discovery.Resource, ix map[string]*Interaction) Result {
	t.Helper()
	p := &Probe{Interactions: ix}
	for _, c := range All() {
		if c.ID == id {
			return Evaluate(c, &RunContext{Model: &discovery.APIModel{}, Resource: r, Probe: p})
		}
	}
	t.Fatalf("check %q not registered", id)
	return Result{}
}

func TestApplyChecks(t *testing.T) {
	r := applyResource()

	// apply-update-200
	if s := evalApply(t, "apply-update-200", r, map[string]*Interaction{
		Apply: inter(200, "{}", "{}")}).Status; s != Pass {
		t.Errorf("apply-update-200 (200): got %v, want Pass", s)
	}
	if s := evalApply(t, "apply-update-200", r, map[string]*Interaction{
		Apply: inter(201, "{}", "{}")}).Status; s != Fail {
		t.Errorf("apply-update-200 (201): got %v, want Fail", s)
	}

	// apply-consistent-on-get
	if s := evalApply(t, "apply-consistent-on-get", r, map[string]*Interaction{
		Apply:         inter(200, `{"title":"x"}`, "{}"),
		GetAfterApply: inter(200, "", `{"title":"x"}`)}).Status; s != Pass {
		t.Errorf("apply-consistent-on-get (match): got %v, want Pass", s)
	}
	if s := evalApply(t, "apply-consistent-on-get", r, map[string]*Interaction{
		Apply:         inter(200, `{"title":"x"}`, "{}"),
		GetAfterApply: inter(200, "", `{"title":"y"}`)}).Status; s != Fail {
		t.Errorf("apply-consistent-on-get (mismatch): got %v, want Fail", s)
	}

	// apply-preserves-absent-fields: author omitted from the second apply.
	preserveIx := func(afterJSON string) map[string]*Interaction {
		return map[string]*Interaction{
			ApplySetOptional:     inter(200, `{"title":"a","author":"b"}`, ""),
			ApplyDropOptional:    inter(200, `{"title":"a"}`, ""),
			GetAfterApplyPartial: inter(200, "", afterJSON),
		}
	}
	if s := evalApply(t, "apply-preserves-absent-fields", r,
		preserveIx(`{"title":"a","author":"b"}`)).Status; s != Pass {
		t.Errorf("apply-preserves (kept): got %v, want Pass", s)
	}
	if s := evalApply(t, "apply-preserves-absent-fields", r,
		preserveIx(`{"title":"a"}`)).Status; s != Fail {
		t.Errorf("apply-preserves (cleared): got %v, want Fail", s)
	}

	// apply-create-201
	if s := evalApply(t, "apply-create-201", r, map[string]*Interaction{
		ApplyCreate: inter(201, "{}", "{}")}).Status; s != Pass {
		t.Errorf("apply-create-201 (201): got %v, want Pass", s)
	}
	if s := evalApply(t, "apply-create-201", r, map[string]*Interaction{
		ApplyCreate: inter(400, "{}", "")}).Status; s != NotApplicable {
		t.Errorf("apply-create-201 (400): got %v, want N/A", s)
	}
	if s := evalApply(t, "apply-create-201", r, map[string]*Interaction{
		ApplyCreate: inter(200, "{}", "{}")}).Status; s != Fail {
		t.Errorf("apply-create-201 (200): got %v, want Fail", s)
	}

	// apply-missing-required-400
	if s := evalApply(t, "apply-missing-required-400", r, map[string]*Interaction{
		ApplyMissingRequired: inter(400, `{"author":"b"}`, "")}).Status; s != Pass {
		t.Errorf("apply-missing-required-400 (400): got %v, want Pass", s)
	}
	if s := evalApply(t, "apply-missing-required-400", r, map[string]*Interaction{
		ApplyMissingRequired: inter(200, `{"author":"b"}`, "{}")}).Status; s != Fail {
		t.Errorf("apply-missing-required-400 (200): got %v, want Fail", s)
	}

	// apply-fully-populated: 'path' (output-only) present → Pass; absent → Fail.
	if s := evalApply(t, "apply-fully-populated", r, map[string]*Interaction{
		Apply: inter(200, "{}", `{"path":"books/b1","title":"x"}`)}).Status; s != Pass {
		t.Errorf("apply-fully-populated (path present): got %v, want Pass", s)
	}
	if s := evalApply(t, "apply-fully-populated", r, map[string]*Interaction{
		Apply: inter(200, "{}", `{"title":"x"}`)}).Status; s != Fail {
		t.Errorf("apply-fully-populated (path absent): got %v, want Fail", s)
	}
}
