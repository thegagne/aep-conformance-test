// Package sampler builds minimal, schema-valid request bodies from a resource's
// JSON Schema. It fills required fields with deterministic values (deterministic
// so failures are reproducible), honoring enums, formats, and patterns where it
// can, and skips read-only / output-only fields.
package sampler

import (
	"fmt"
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// Sampler generates bodies and unique resource IDs for a run.
type Sampler struct {
	// Overrides supplies values for fields the generator can't satisfy
	// (tight patterns, uniqueness), keyed by field name.
	Overrides map[string]any
	counter   int
}

// New returns a Sampler with optional per-field overrides.
func New(overrides map[string]any) *Sampler {
	return &Sampler{Overrides: overrides}
}

// idPrefix is used for generated resource IDs, kept short and AEP-122 compliant.
const idPrefix = "aepconf"

// NextID returns a fresh resource ID matching ^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$.
func (s *Sampler) NextID(singular string) string {
	s.counter++
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r + 32
		default:
			return '-'
		}
	}, singular)
	id := fmt.Sprintf("%s-%s-%d", idPrefix, clean, s.counter)
	if len(id) > 63 {
		id = id[:63]
	}
	return strings.Trim(id, "-")
}

// Body builds a request body for the given resource schema, filling required
// fields (and, when includeOptional, optional fields too).
func (s *Sampler) Body(schema *discovery.Schema, includeOptional bool) map[string]any {
	if schema == nil {
		return map[string]any{}
	}
	out := map[string]any{}
	if schema.Properties == nil {
		return out
	}
	for _, name := range schema.Properties.Order {
		prop := schema.Properties.Map[name]
		if prop == nil || prop.ReadOnly {
			continue // never send server-assigned / output-only fields
		}
		required := schema.IsRequired(name)
		if !required && !includeOptional {
			continue
		}
		if v, ok := s.Overrides[name]; ok {
			out[name] = v
			continue
		}
		out[name] = s.value(name, prop)
	}
	return out
}

// value produces a deterministic sample for one field schema.
func (s *Sampler) value(name string, sc *discovery.Schema) any {
	if len(sc.Enum) > 0 {
		return sc.Enum[0]
	}
	switch sc.Type {
	case "string":
		return s.stringValue(name, sc)
	case "integer":
		return 1
	case "number":
		return 1
	case "boolean":
		return true
	case "array":
		if sc.Items != nil {
			return []any{s.value(name, sc.Items)}
		}
		return []any{}
	case "object":
		return s.Body(sc, false)
	default:
		// Untyped: default to a string.
		return s.stringValue(name, sc)
	}
}

func (s *Sampler) stringValue(name string, sc *discovery.Schema) string {
	switch sc.Format {
	case "date-time":
		return "2024-01-01T00:00:00Z"
	case "date":
		return "2024-01-01"
	case "byte":
		return "YWVwY29uZg==" // base64("aepconf")
	case "uuid":
		return "00000000-0000-4000-8000-000000000000"
	}
	// A conservative value: lowercase alnum, satisfies most simple patterns.
	v := strings.ToLower(name)
	if v == "" {
		v = "sample"
	}
	return v
}
