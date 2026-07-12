package runner

import (
	"os"
	"testing"

	"github.com/thegagne/aep-conformance-test/internal/checks"
	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// TestStaticRunAgainstFixture runs the suite with no live client (static checks
// only) against the aepc bookstore spec and asserts known outcomes, locking
// check behavior without needing a running server.
func TestStaticRunAgainstFixture(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/bookstore_openapi.json")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	model, err := discovery.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	rn := New(model, nil) // nil client => static only
	results := rn.Run(nil)

	byKey := map[string]checks.Status{}
	for _, r := range results {
		byKey[r.CheckID+"/"+r.Resource] = r.Status
	}

	// Structural checks that must pass on the fixture.
	assertStatus(t, byKey, "path-field-present/book", checks.Pass)
	assertStatus(t, byKey, "path-field-readonly/book", checks.Pass)
	assertStatus(t, byKey, "openapi-version-3-1/", checks.Pass)

	// aepc emits List{singular}; AEP-130 requires List{plural} -> a real failure.
	assertStatus(t, byKey, "operation-id-list-plural/book", checks.Fail)

	// Dynamic checks must be skipped when there is no live server.
	assertStatus(t, byKey, "get-200-on-existing/book", checks.Skipped)
	assertStatus(t, byKey, "create-duplicate-already-exists/book", checks.Skipped)

	// Feature-gated static check: the fixture's book List has an unreachable field.
	assertStatus(t, byKey, "unreachable-field-present/book", checks.Pass)
}

func assertStatus(t *testing.T, m map[string]checks.Status, key string, want checks.Status) {
	t.Helper()
	got, ok := m[key]
	if !ok {
		t.Errorf("no result for %q", key)
		return
	}
	if got != want {
		t.Errorf("%s: status = %v, want %v", key, got, want)
	}
}
