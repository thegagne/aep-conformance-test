package report

import (
	"os"
	"testing"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

func TestCapabilitiesReflectModel(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/bookstore_openapi.json")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	m, err := discovery.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	caps := Capabilities(m.Resources)

	var book *CollectionCapabilities
	for i := range caps {
		if caps[i].Singular == "book" {
			book = &caps[i]
		}
	}
	if book == nil {
		t.Fatal("no capabilities for book")
	}
	// The collection is addressed by its plural name.
	if book.Collection != "books" {
		t.Errorf("collection = %q, want books", book.Collection)
	}

	feat := map[string]bool{}
	for _, f := range book.Features {
		feat[f.Name] = f.Implemented
	}
	// The fixture's book has pagination + an unreachable field, but no filtering.
	if !feat["pagination"] {
		t.Error("book should report pagination implemented")
	}
	if !feat["unreachable"] {
		t.Error("book should report unreachable implemented")
	}
	if feat["filtering"] {
		t.Error("book should report filtering NOT implemented in this fixture")
	}
	// book exposes an :archive custom method.
	if len(book.Custom) != 1 || book.Custom[0] != ":archive" {
		t.Errorf("book custom = %v, want [:archive]", book.Custom)
	}
}
