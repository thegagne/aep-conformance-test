package report

import (
	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// Capability is one implemented-or-not aspect of a collection.
type Capability struct {
	Name        string `json:"name"`
	Implemented bool   `json:"implemented"`
}

// CollectionCapabilities is the implemented/not-implemented picture for a
// collection: which standard methods and optional AEP features its spec
// declares. The collection/singular identifiers are carried by the enclosing
// group (they name it once), so they are not serialized here.
type CollectionCapabilities struct {
	Collection string       `json:"-"` // plural, e.g. "books"
	Singular   string       `json:"-"` // e.g. "book"; matches check results
	Methods    []Capability `json:"methods"`
	Custom     []string     `json:"custom_methods,omitempty"`
	Features   []Capability `json:"features"`
}

// stdMethodOrder is the display order for standard methods.
var stdMethodOrder = []struct {
	label  string
	method discovery.Method
}{
	{"Get", discovery.MethodGet},
	{"List", discovery.MethodList},
	{"Create", discovery.MethodCreate},
	{"Update", discovery.MethodUpdate},
	{"Apply", discovery.MethodApply},
	{"Delete", discovery.MethodDelete},
}

// Capabilities computes the capability matrix for the given collections.
func Capabilities(resources []*discovery.Resource) []CollectionCapabilities {
	var out []CollectionCapabilities
	for _, r := range resources {
		rc := CollectionCapabilities{Collection: r.Plural, Singular: r.Singular}
		for _, m := range stdMethodOrder {
			rc.Methods = append(rc.Methods, Capability{m.label, r.Method(m.method) != nil})
		}
		for _, c := range r.Custom {
			rc.Custom = append(rc.Custom, ":"+c.CustomVerb)
		}
		f := r.Features
		rc.Features = []Capability{
			{"pagination", f.Pagination},
			{"filtering", f.Filter},
			{"ordering", f.OrderBy},
			{"skip", f.Skip},
			{"read_mask", f.ReadMask},
			{"field_mask", f.FieldMask},
			{"soft_delete", f.ShowDeleted},
			{"cascade_force", f.Force},
			{"idempotency", f.IdempotencyKey},
			{"revisions", f.Revisions},
			{"expiration", f.ExpireTime},
			{"unreachable", f.Unreachable},
			{"user_settable_id", f.UserSettableID},
			{"lro", f.LRO},
		}
		out = append(out, rc)
	}
	return out
}
