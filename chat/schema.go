package chat

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// JSONSchema represents a JSON Schema for structured output requests.
type JSONSchema struct {
	Type                 string                `json:"type"`
	Properties           map[string]JSONSchema `json:"properties,omitempty"`
	Required             []string              `json:"required,omitempty"`
	Items                *JSONSchema           `json:"items,omitempty"`
	AdditionalProperties *JSONSchema           `json:"additionalProperties,omitempty"`
	Description          string                `json:"description,omitempty"`
}

// NewJSONSchema builds a JSON Schema from any struct value or pointer to struct.
func NewJSONSchema(v any) (*JSONSchema, error) {
	t := reflect.TypeOf(v)
	if t == nil {
		return nil, fmt.Errorf("nil value provided")
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", t.Kind())
	}

	b := schemaBuilder{visiting: map[reflect.Type]bool{}}
	s := b.fromType(t)
	return &s, nil
}

// RawMessage returns the schema serialized as a json.RawMessage.
func (s *JSONSchema) RawMessage() json.RawMessage {
	data, err := json.Marshal(s)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSONSchema: %v", err))
	}
	return data
}

type schemaBuilder struct {
	visiting map[reflect.Type]bool
}

func (b *schemaBuilder) fromType(t reflect.Type) JSONSchema {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Bool:
		return JSONSchema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return JSONSchema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return JSONSchema{Type: "number"}
	case reflect.String:
		return JSONSchema{Type: "string"}
	case reflect.Slice, reflect.Array:
		itemSchema := b.fromType(t.Elem())
		return JSONSchema{Type: "array", Items: &itemSchema}
	case reflect.Map:
		valueSchema := b.fromType(t.Elem())
		return JSONSchema{Type: "object", AdditionalProperties: &valueSchema}
	case reflect.Struct:
		return b.fromStruct(t)
	default:
		return JSONSchema{Type: "string"}
	}
}

func (b *schemaBuilder) fromStruct(t reflect.Type) JSONSchema {
	if b.visiting[t] {
		return JSONSchema{Type: "object"}
	}
	b.visiting[t] = true
	defer delete(b.visiting, t)

	s := JSONSchema{Type: "object", Properties: map[string]JSONSchema{}}

	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name, omitEmpty, skip := parseJSONTag(field)
		if skip {
			continue
		}

		// Handle embedded structs
		if field.Anonymous && name == field.Name {
			embedded := b.fromType(field.Type)
			for propName, propSchema := range embedded.Properties {
				s.Properties[propName] = propSchema
			}
			s.Required = append(s.Required, embedded.Required...)
			continue
		}

		fieldSchema := b.fromType(field.Type)

		// Use the "llm" tag as the description if present.
		if desc := field.Tag.Get("llm"); desc != "" && desc != "ignore" {
			fieldSchema.Description = desc
		}
		if field.Tag.Get("llm") == "ignore" {
			continue
		}

		s.Properties[name] = fieldSchema

		isPtr := field.Type.Kind() == reflect.Ptr
		if !omitEmpty && !isPtr {
			s.Required = append(s.Required, name)
		}
	}

	return s
}

func parseJSONTag(field reflect.StructField) (name string, omitEmpty bool, skip bool) {
	name = field.Name
	tag := field.Tag.Get("json")
	if tag == "" {
		return name, false, false
	}

	parts := strings.Split(tag, ",")
	if parts[0] == "-" {
		return "", false, true
	}
	if parts[0] != "" {
		name = parts[0]
	}

	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			omitEmpty = true
			break
		}
	}
	return name, omitEmpty, false
}
