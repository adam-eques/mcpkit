package tools

import "encoding/json"

// Prop describes a single JSON Schema property. It covers the small subset of
// draft 2020-12 that tool inputs need, keeping tool definitions declarative and
// dependency-free.
type Prop struct {
	Type        string          `json:"type,omitempty"`
	Description string          `json:"description,omitempty"`
	Enum        []string        `json:"enum,omitempty"`
	Items       *Prop           `json:"items,omitempty"`
	Properties  map[string]Prop `json:"properties,omitempty"`
	Default     any             `json:"default,omitempty"`
	Minimum     *float64        `json:"minimum,omitempty"`
	Maximum     *float64        `json:"maximum,omitempty"`
}

// Object builds an object schema from named properties and a required list.
func Object(props map[string]Prop, required ...string) json.RawMessage {
	schema := map[string]any{"type": "object"}
	if props == nil {
		props = map[string]Prop{}
	}
	schema["properties"] = props
	if len(required) > 0 {
		schema["required"] = required
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		// The inputs are always encodable; fall back to a permissive schema.
		return json.RawMessage(`{"type":"object"}`)
	}
	return raw
}

// Str is a shorthand for a string property with a description.
func Str(desc string) Prop { return Prop{Type: "string", Description: desc} }

// Num is a shorthand for a number property with a description.
func Num(desc string) Prop { return Prop{Type: "number", Description: desc} }

// Int is a shorthand for an integer property with a description.
func Int(desc string) Prop { return Prop{Type: "integer", Description: desc} }

// Bool is a shorthand for a boolean property with a description.
func Bool(desc string) Prop { return Prop{Type: "boolean", Description: desc} }

// Ptr returns a pointer to v, convenient for the Minimum/Maximum fields.
func Ptr[T any](v T) *T { return &v }
