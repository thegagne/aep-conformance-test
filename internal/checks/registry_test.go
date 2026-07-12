package checks

import "testing"

// TestRegistryIDsUnique enforces the plan's self-consistency guarantee: every
// check id maps to exactly one registry entry, so coverage stays auditable.
func TestRegistryIDsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range All() {
		if c.ID == "" {
			t.Error("registered check with empty ID")
		}
		if seen[c.ID] {
			t.Errorf("duplicate check id %q", c.ID)
		}
		seen[c.ID] = true
		if c.Run == nil {
			t.Errorf("check %q has nil Run", c.ID)
		}
		if c.AEP == 0 {
			t.Errorf("check %q has no AEP number", c.ID)
		}
	}
	if len(seen) == 0 {
		t.Fatal("no checks registered")
	}
}
