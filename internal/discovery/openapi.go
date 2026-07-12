package discovery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Document is the raw OpenAPI 3.1 document, decoded just enough to build the
// APIModel. Only the fields the tester needs are modeled.
type Document struct {
	OpenAPI    string        `json:"openapi"`
	Info       Info          `json:"info"`
	Servers    []Server      `json:"servers"`
	Paths      *OrderedPaths `json:"paths"`
	Components struct {
		Schemas *OrderedSchemas `json:"schemas"`
	} `json:"components"`
}

// Info is the OpenAPI info object (subset).
type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// Server is an OpenAPI server entry (subset).
type Server struct {
	URL string `json:"url"`
}

// PathItem holds the operations declared for a single path template.
type PathItem struct {
	Get    *Operation `json:"get"`
	Post   *Operation `json:"post"`
	Put    *Operation `json:"put"`
	Patch  *Operation `json:"patch"`
	Delete *Operation `json:"delete"`
}

// operations returns the declared (verb, operation) pairs in a stable order.
func (p *PathItem) operations() []struct {
	Verb string
	Op   *Operation
} {
	var out []struct {
		Verb string
		Op   *Operation
	}
	add := func(v string, op *Operation) {
		if op != nil {
			out = append(out, struct {
				Verb string
				Op   *Operation
			}{v, op})
		}
	}
	add("GET", p.Get)
	add("POST", p.Post)
	add("PUT", p.Put)
	add("PATCH", p.Patch)
	add("DELETE", p.Delete)
	return out
}

// Operation is an OpenAPI operation object (subset).
type Operation struct {
	OperationID string                 `json:"operationId"`
	Parameters  []Parameter            `json:"parameters"`
	RequestBody *RequestBody           `json:"requestBody"`
	Responses   map[string]ResponseObj `json:"responses"`
}

// Parameter is an OpenAPI parameter object (subset).
type Parameter struct {
	Name     string  `json:"name"`
	In       string  `json:"in"`
	Required bool    `json:"required"`
	Schema   *Schema `json:"schema"`
}

// RequestBody is an OpenAPI requestBody object (subset).
type RequestBody struct {
	Required bool                 `json:"required"`
	Content  map[string]MediaType `json:"content"`
}

// ResponseObj is an OpenAPI response object (subset).
type ResponseObj struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content"`
}

// MediaType is an OpenAPI media type object (subset).
type MediaType struct {
	Schema *Schema `json:"schema"`
}

// Load reads and parses an OpenAPI document from a local file path or an HTTP(S)
// URL, returning the built APIModel. Optional headers are applied when the
// source is fetched over HTTP (e.g. auth for a protected spec endpoint).
func Load(source string, headers ...map[string]string) (*APIModel, error) {
	var h map[string]string
	if len(headers) > 0 {
		h = headers[0]
	}
	raw, err := readSource(source, h)
	if err != nil {
		return nil, err
	}
	return Parse(raw)
}

// LoadFromBaseURL fetches <baseURL>/openapi.json.
func LoadFromBaseURL(baseURL string, headers ...map[string]string) (*APIModel, error) {
	return Load(strings.TrimRight(baseURL, "/")+"/openapi.json", headers...)
}

func readSource(source string, headers map[string]string) ([]byte, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest(http.MethodGet, source, nil)
		if err != nil {
			return nil, fmt.Errorf("fetch %s: %w", source, err)
		}
		req.Header.Set("Accept", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch %s: %w", source, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch %s: status %d", source, resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	b, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", source, err)
	}
	return b, nil
}

// Parse decodes a raw OpenAPI JSON document and builds the APIModel.
func Parse(raw []byte) (*APIModel, error) {
	var doc Document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi: %w", err)
	}
	if doc.OpenAPI == "" {
		return nil, fmt.Errorf("not an openapi document: missing 'openapi' version field")
	}
	m := &APIModel{
		OpenAPIVersion: doc.OpenAPI,
		Title:          doc.Info.Title,
		Doc:            &doc,
		Components:     map[string]*Schema{},
	}
	if doc.Components.Schemas != nil {
		for _, name := range doc.Components.Schemas.Order {
			m.Components[name] = doc.Components.Schemas.Map[name]
		}
	}
	for _, s := range doc.Servers {
		m.Servers = append(m.Servers, s.URL)
	}
	build(m, &doc)
	m.APIName = deriveAPIName(m)
	return m, nil
}

// resolve follows a local $ref (one level) to its component schema. It returns
// the input unchanged if there is nothing to resolve.
func (m *APIModel) resolve(s *Schema) *Schema {
	if s == nil || s.Ref == "" {
		return s
	}
	name := strings.TrimPrefix(s.Ref, "#/components/schemas/")
	if c, ok := m.Components[name]; ok {
		return c
	}
	return s
}

// deriveAPIName prefers the API-name prefix of a resource type
// (e.g. "bookstore.example.com/book" -> "bookstore.example.com"), which is the
// canonical AEP API name, falling back to the document title.
func deriveAPIName(m *APIModel) string {
	for _, r := range m.Resources {
		if i := strings.LastIndex(r.Type, "/"); i > 0 {
			return r.Type[:i]
		}
	}
	return strings.ToLower(strings.TrimSpace(m.Doc.Info.Title))
}

// OrderedSchemas decodes a JSON object of named schemas while preserving key
// order (needed for the AEP-122 "path field first" check and stable output).
type OrderedSchemas struct {
	Order []string
	Map   map[string]*Schema
}

// UnmarshalJSON reads the object keys in document order.
func (o *OrderedSchemas) UnmarshalJSON(b []byte) error {
	o.Map = map[string]*Schema{}
	o.Order = nil
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if tok == nil {
		return nil // null
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return fmt.Errorf("expected object, got %v", tok)
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key := keyTok.(string)
		var s Schema
		if err := dec.Decode(&s); err != nil {
			return err
		}
		if _, exists := o.Map[key]; !exists {
			o.Order = append(o.Order, key)
		}
		sc := s
		o.Map[key] = &sc
	}
	_, err = dec.Token() // closing }
	return err
}

// OrderedPaths decodes the OpenAPI paths object preserving declaration order.
type OrderedPaths struct {
	Order []string
	Map   map[string]*PathItem
}

// UnmarshalJSON reads path keys in document order.
func (o *OrderedPaths) UnmarshalJSON(b []byte) error {
	o.Map = map[string]*PathItem{}
	o.Order = nil
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if tok == nil {
		return nil
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return fmt.Errorf("expected object, got %v", tok)
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key := keyTok.(string)
		var pi PathItem
		if err := dec.Decode(&pi); err != nil {
			return err
		}
		if _, exists := o.Map[key]; !exists {
			o.Order = append(o.Order, key)
		}
		p := pi
		o.Map[key] = &p
	}
	_, err = dec.Token()
	return err
}
