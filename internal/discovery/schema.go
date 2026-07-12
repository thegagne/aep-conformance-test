package discovery

import (
	"encoding/json"
	"fmt"
)

// schemaJSON is the wire shape of a schema. type and enum need lenient decoding
// (OpenAPI 3.1 allows "type" to be a string or an array, and enum values may be
// non-strings), so Schema.UnmarshalJSON post-processes them.
type schemaJSON struct {
	Ref         string          `json:"$ref"`
	Type        json.RawMessage `json:"type"`
	Format      string          `json:"format"`
	ReadOnly    bool            `json:"readOnly"`
	Description string          `json:"description"`
	Pattern     string          `json:"pattern"`
	Enum        []any           `json:"enum"`
	MaxLength   *int            `json:"maxLength"`
	MaxItems    *int            `json:"maxItems"`
	Required    []string        `json:"required"`
	Properties  *OrderedSchemas `json:"properties"`
	Items       *Schema         `json:"items"`

	XAEPResource *XAEPResource  `json:"x-aep-resource"`
	XAEPField    *XAEPField     `json:"x-aep-field"`
	XAEPLRO      map[string]any `json:"x-aep-long-running-operation"`
}

// UnmarshalJSON decodes a schema, tolerating 3.1's string-or-array "type".
func (s *Schema) UnmarshalJSON(b []byte) error {
	var j schemaJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	s.Ref = j.Ref
	s.Type = decodeType(j.Type)
	s.Format = j.Format
	s.ReadOnly = j.ReadOnly
	s.Description = j.Description
	s.Pattern = j.Pattern
	s.MaxLength = j.MaxLength
	s.MaxItems = j.MaxItems
	s.Required = j.Required
	s.Properties = j.Properties
	s.Items = j.Items
	s.XAEPResource = j.XAEPResource
	s.XAEPField = j.XAEPField
	s.XAEPLRO = j.XAEPLRO
	for _, e := range j.Enum {
		s.Enum = append(s.Enum, fmt.Sprintf("%v", e))
	}
	return nil
}

// decodeType extracts the primary (non-null) type from a 3.1 type value that
// may be a bare string or an array such as ["string","null"].
func decodeType(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return single
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err == nil {
		for _, t := range many {
			if t != "null" {
				return t
			}
		}
	}
	return ""
}
