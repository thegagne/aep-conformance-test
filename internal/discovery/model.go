// Package discovery parses an AEP-annotated OpenAPI 3.1 document into an
// APIModel: the resolved, resource-oriented view that every conformance check
// is driven from. Field names, method endpoints, and optional-feature flags all
// come from the spec rather than being hard-coded, so the checks work whether an
// API serializes snake_case or camelCase.
package discovery

import (
	"slices"
	"strings"
)

// Method is a standard AEP method kind.
type Method string

const (
	MethodGet    Method = "get"
	MethodList   Method = "list"
	MethodCreate Method = "create"
	MethodUpdate Method = "update"
	MethodApply  Method = "apply"
	MethodDelete Method = "delete"
)

// APIModel is the parsed, resolved view of an AEP API's OpenAPI document.
type APIModel struct {
	OpenAPIVersion string
	Title          string
	// APIName is a best-effort API name (from the first server host or info),
	// used by AEP-102 naming checks.
	APIName   string
	Servers   []string
	Resources []*Resource
	// Components holds every component schema by name, for $ref resolution and
	// static checks that need the full field set.
	Components map[string]*Schema
	// Doc is the raw parsed document, retained for static checks that need
	// document-level fidelity (e.g. operationId scanning).
	Doc *Document
}

// Resource is a single AEP resource type discovered from an x-aep-resource
// annotation, together with the endpoints that operate on it.
type Resource struct {
	Singular string
	Plural   string
	Type     string   // x-aep-resource.type, e.g. bookstore.example.com/book
	Parents  []string // singular names of parent resources
	Patterns []string // path patterns, e.g. publishers/{publisher_id}/books/{book_id}
	Schema   *Schema  // the resource body schema

	Methods map[Method]*Endpoint // standard methods present on this resource
	Custom  []*Endpoint          // custom (:verb) methods

	Depth    int // parent-chain depth; 0 == top-level
	Features FeatureFlags
}

// Method returns the endpoint for a standard method, or nil if absent.
func (r *Resource) Method(m Method) *Endpoint { return r.Methods[m] }

// ResourceBySingular finds a resource by its singular name (nil if absent).
func (m *APIModel) ResourceBySingular(singular string) *Resource {
	for _, r := range m.Resources {
		if r.Singular == singular {
			return r
		}
	}
	return nil
}

// CanonicalPattern returns the resource's canonical (with-id) path pattern.
func (r *Resource) CanonicalPattern() string { return lastPattern(r) }

// CollectionID returns the final collection identifier segment of the resource
// (e.g. "books" for publishers/{publisher_id}/books/{book_id}).
func (r *Resource) CollectionID() string {
	p := collectionPattern(r)
	segs := strings.Split(p, "/")
	if len(segs) == 0 {
		return ""
	}
	return segs[len(segs)-1]
}

// FeatureFlags records which optional AEP design patterns a resource's spec
// opts into. Conditional checks are gated on these so absent features report
// NotApplicable rather than failing.
type FeatureFlags struct {
	Pagination      bool // page_token / max_page_size on List
	Skip            bool // skip on List
	Filter          bool // filter on List
	OrderBy         bool // order_by on List
	ReadMask        bool // read_mask / view on Get/List
	FieldMask       bool // update_mask on Update
	ShowDeleted     bool // show_deleted on List (soft delete)
	Force           bool // force on Delete (cascading)
	IdempotencyKey  bool // idempotency_key on a mutating method
	Revisions       bool // a revisions subcollection exists
	ExpireTime      bool // expire_time field on the resource
	Unreachable     bool // unreachable field on List response
	UserSettableID  bool // id query param on Create
	AllowMissing    bool // allow_missing on Update
	OverwriteDelete bool // overwrite_soft_deleted on Create
	LRO             bool // long-running operations (202 response or x-aep-long-running-operation)
}

// Endpoint is a single HTTP operation.
type Endpoint struct {
	Method      Method
	Custom      bool
	CustomVerb  string // for custom methods, the verb after ':'
	HTTPVerb    string // GET, POST, PATCH, PUT, DELETE (upper-case)
	Path        string // OpenAPI path template
	OperationID string

	PathParams  []Param
	QueryParams []Param

	RequestBodySchema   *Schema
	RequestBodyRequired bool
	// RequestContentTypes lists the media types declared for the request body
	// (e.g. application/json, application/merge-patch+json), in document order.
	RequestContentTypes []string

	// Responses maps HTTP status code (as written in the spec) to the response.
	Responses map[string]*Response
}

// QueryParam returns the named query parameter, or nil.
func (e *Endpoint) QueryParam(name string) *Param {
	for i := range e.QueryParams {
		if e.QueryParams[i].Name == name {
			return &e.QueryParams[i]
		}
	}
	return nil
}

// Param is a path or query parameter.
type Param struct {
	Name     string
	In       string // path | query | header
	Required bool
	Schema   *Schema
}

// Response is a single response entry with its resolved JSON body schema.
type Response struct {
	Status      string
	Description string
	Schema      *Schema // resolved application/json schema, if any
}

// Schema is a minimal JSON Schema view carrying the AEP vendor extensions and
// the field properties (in document order, so path-first can be checked).
type Schema struct {
	Ref         string
	Type        string
	Format      string
	ReadOnly    bool
	Description string
	Pattern     string
	Enum        []string
	MaxLength   *int
	MaxItems    *int
	Required    []string

	// Properties preserves the document order of object properties.
	Properties *OrderedSchemas
	Items      *Schema

	XAEPResource *XAEPResource
	XAEPField    *XAEPField
	XAEPLRO      map[string]any
}

// Prop returns a property schema by name (nil if absent).
func (s *Schema) Prop(name string) *Schema {
	if s == nil || s.Properties == nil {
		return nil
	}
	return s.Properties.Map[name]
}

// HasProp reports whether the schema declares the named property.
func (s *Schema) HasProp(name string) bool { return s.Prop(name) != nil }

// IsRequired reports whether name is in the schema's required list.
func (s *Schema) IsRequired(name string) bool {
	if s == nil {
		return false
	}
	return slices.Contains(s.Required, name)
}

// XAEPResource mirrors the x-aep-resource vendor extension.
type XAEPResource struct {
	Singular string   `json:"singular"`
	Plural   string   `json:"plural"`
	Patterns []string `json:"patterns"`
	Parents  []string `json:"parents"`
	Type     string   `json:"type"`
}

// XAEPField mirrors the x-aep-field vendor extension.
type XAEPField struct {
	FieldNumber int `json:"field_number"`
}
