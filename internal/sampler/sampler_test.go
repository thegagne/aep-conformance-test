package sampler

import (
	"regexp"
	"testing"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

var idPattern = regexp.MustCompile(`^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$`)

func TestNextIDMatchesAEP122(t *testing.T) {
	s := New(nil)
	seen := map[string]bool{}
	for _, name := range []string{"book", "BookEdition", "publisher", "a"} {
		id := s.NextID(name)
		if !idPattern.MatchString(id) {
			t.Errorf("id %q does not match AEP-122 pattern", id)
		}
		if seen[id] {
			t.Errorf("duplicate id %q", id)
		}
		seen[id] = true
	}
}

func schema() *discovery.Schema {
	props := &discovery.OrderedSchemas{
		Order: []string{"path", "title", "count", "published", "tags"},
		Map: map[string]*discovery.Schema{
			"path":      {Type: "string", ReadOnly: true},
			"title":     {Type: "string"},
			"count":     {Type: "integer"},
			"published": {Type: "boolean"},
			"tags":      {Type: "array", Items: &discovery.Schema{Type: "string"}},
		},
	}
	return &discovery.Schema{Type: "object", Properties: props, Required: []string{"title", "count", "published"}}
}

func TestBodySkipsReadOnlyAndFillsRequired(t *testing.T) {
	s := New(nil)
	body := s.Body(schema(), false)
	if _, ok := body["path"]; ok {
		t.Error("body must not include read-only 'path'")
	}
	for _, req := range []string{"title", "count", "published"} {
		if _, ok := body[req]; !ok {
			t.Errorf("required field %q missing from body", req)
		}
	}
	if _, ok := body["tags"]; ok {
		t.Error("optional 'tags' should be omitted when includeOptional=false")
	}
	if body["count"] != 1 || body["published"] != true {
		t.Errorf("unexpected values: %+v", body)
	}
}

func TestBodyOverrides(t *testing.T) {
	s := New(map[string]any{"title": "custom"})
	body := s.Body(schema(), false)
	if body["title"] != "custom" {
		t.Errorf("override not applied: %v", body["title"])
	}
}
