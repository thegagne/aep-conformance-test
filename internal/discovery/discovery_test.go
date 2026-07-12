package discovery

import (
	"os"
	"testing"
)

func loadTestModel(t *testing.T) *APIModel {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/bookstore_openapi.json")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	m, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return m
}

func TestParseVersionAndResources(t *testing.T) {
	m := loadTestModel(t)
	if m.OpenAPIVersion != "3.1.0" {
		t.Errorf("openapi version = %q, want 3.1.0", m.OpenAPIVersion)
	}
	want := map[string]bool{"publisher": true, "book": true, "book-edition": true, "isbn": true, "store": true, "item": true}
	got := map[string]bool{}
	for _, r := range m.Resources {
		got[r.Singular] = true
	}
	for s := range want {
		if !got[s] {
			t.Errorf("missing resource %q", s)
		}
	}
}

func TestResourceHierarchyAndDepth(t *testing.T) {
	m := loadTestModel(t)
	byName := map[string]*Resource{}
	for _, r := range m.Resources {
		byName[r.Singular] = r
	}
	book := byName["book"]
	if book == nil {
		t.Fatal("no book resource")
	}
	if book.Depth != 1 {
		t.Errorf("book depth = %d, want 1", book.Depth)
	}
	if len(book.Parents) != 1 || book.Parents[0] != "publisher" {
		t.Errorf("book parents = %v, want [publisher]", book.Parents)
	}
	if ed := byName["book-edition"]; ed == nil || ed.Depth != 2 {
		t.Errorf("book-edition depth wrong: %+v", ed)
	}
	// Resources must be ordered ancestors-first.
	for i := 1; i < len(m.Resources); i++ {
		if m.Resources[i-1].Depth > m.Resources[i].Depth {
			t.Errorf("resources not ordered by depth at %d", i)
		}
	}
}

func TestEndpointClassification(t *testing.T) {
	m := loadTestModel(t)
	var book *Resource
	for _, r := range m.Resources {
		if r.Singular == "book" {
			book = r
		}
	}
	if book == nil {
		t.Fatal("no book")
	}
	cases := []struct {
		method Method
		verb   string
		path   string
	}{
		{MethodGet, "GET", "/publishers/{publisher_id}/books/{book_id}"},
		{MethodList, "GET", "/publishers/{publisher_id}/books"},
		{MethodCreate, "POST", "/publishers/{publisher_id}/books"},
		{MethodUpdate, "PATCH", "/publishers/{publisher_id}/books/{book_id}"},
		{MethodApply, "PUT", "/publishers/{publisher_id}/books/{book_id}"},
		{MethodDelete, "DELETE", "/publishers/{publisher_id}/books/{book_id}"},
	}
	for _, c := range cases {
		ep := book.Method(c.method)
		if ep == nil {
			t.Errorf("book missing method %s", c.method)
			continue
		}
		if ep.HTTPVerb != c.verb || ep.Path != c.path {
			t.Errorf("%s: got %s %s, want %s %s", c.method, ep.HTTPVerb, ep.Path, c.verb, c.path)
		}
	}
	// :archive custom method.
	if len(book.Custom) != 1 || book.Custom[0].CustomVerb != "archive" {
		t.Errorf("book custom methods = %+v, want one archive", book.Custom)
	}
}

func TestFeatureDetection(t *testing.T) {
	m := loadTestModel(t)
	var book *Resource
	for _, r := range m.Resources {
		if r.Singular == "book" {
			book = r
		}
	}
	if !book.Features.Pagination {
		t.Error("book should have pagination (max_page_size/page_token)")
	}
	if !book.Features.UserSettableID {
		t.Error("book should have user-settable id (id query param on create)")
	}
	if !book.Features.Unreachable {
		t.Error("book list response has unreachable field")
	}
	if book.Features.Filter {
		t.Error("book should not have filter in this spec")
	}
}

func TestAPINameFromResourceType(t *testing.T) {
	m := loadTestModel(t)
	if m.APIName != "bookstore.example.com" {
		t.Errorf("APIName = %q, want bookstore.example.com", m.APIName)
	}
}

func TestPathFieldPresentAndReadOnly(t *testing.T) {
	m := loadTestModel(t)
	for _, r := range m.Resources {
		p := r.Schema.Prop("path")
		if p == nil {
			t.Errorf("resource %q missing path field", r.Singular)
			continue
		}
		if !p.ReadOnly {
			t.Errorf("resource %q path field not readOnly", r.Singular)
		}
	}
}
