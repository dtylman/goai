package chat_test

import (
	"encoding/json"
	"testing"

	"github.com/dtylman/goai/chat"
)

func TestNewJSONSchema_SimpleStruct(t *testing.T) {
	type Simple struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Score float64
	}

	schema, err := chat.NewJSONSchema(&Simple{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Type != "object" {
		t.Fatalf("expected object, got %s", schema.Type)
	}
	if len(schema.Properties) != 3 {
		t.Fatalf("expected 3 properties, got %d", len(schema.Properties))
	}
	if schema.Properties["name"].Type != "string" {
		t.Errorf("expected name to be string, got %s", schema.Properties["name"].Type)
	}
	if schema.Properties["age"].Type != "integer" {
		t.Errorf("expected age to be integer, got %s", schema.Properties["age"].Type)
	}
	if schema.Properties["Score"].Type != "number" {
		t.Errorf("expected Score to be number, got %s", schema.Properties["Score"].Type)
	}
	// All non-pointer, non-omitempty fields should be required
	assertContains(t, schema.Required, "name")
	assertContains(t, schema.Required, "age")
	assertContains(t, schema.Required, "Score")
}

func TestNewJSONSchema_OmitEmpty(t *testing.T) {
	type WithOmit struct {
		Required string `json:"required"`
		Optional string `json:"optional,omitempty"`
	}

	schema, err := chat.NewJSONSchema(&WithOmit{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, schema.Required, "required")
	assertNotContains(t, schema.Required, "optional")
}

func TestNewJSONSchema_PointerFields(t *testing.T) {
	type WithPtr struct {
		Name     string  `json:"name"`
		Nickname *string `json:"nickname"`
	}

	schema, err := chat.NewJSONSchema(&WithPtr{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, schema.Required, "name")
	assertNotContains(t, schema.Required, "nickname")
	if schema.Properties["nickname"].Type != "string" {
		t.Errorf("expected nickname to be string, got %s", schema.Properties["nickname"].Type)
	}
}

func TestNewJSONSchema_NestedStruct(t *testing.T) {
	type Address struct {
		City    string `json:"city"`
		Country string `json:"country"`
	}
	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	schema, err := chat.NewJSONSchema(&Person{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	addr := schema.Properties["address"]
	if addr.Type != "object" {
		t.Fatalf("expected address to be object, got %s", addr.Type)
	}
	if addr.Properties["city"].Type != "string" {
		t.Errorf("expected city to be string, got %s", addr.Properties["city"].Type)
	}
}

func TestNewJSONSchema_Slice(t *testing.T) {
	type WithSlice struct {
		Tags []string `json:"tags"`
	}

	schema, err := chat.NewJSONSchema(&WithSlice{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags := schema.Properties["tags"]
	if tags.Type != "array" {
		t.Fatalf("expected array, got %s", tags.Type)
	}
	if tags.Items == nil || tags.Items.Type != "string" {
		t.Errorf("expected items to be string")
	}
}

func TestNewJSONSchema_Map(t *testing.T) {
	type WithMap struct {
		Metadata map[string]int `json:"metadata"`
	}

	schema, err := chat.NewJSONSchema(&WithMap{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	meta := schema.Properties["metadata"]
	if meta.Type != "object" {
		t.Fatalf("expected object, got %s", meta.Type)
	}
	if meta.AdditionalProperties == nil || meta.AdditionalProperties.Type != "integer" {
		t.Errorf("expected additionalProperties to be integer")
	}
}

func TestNewJSONSchema_SkipDash(t *testing.T) {
	type WithSkip struct {
		Visible string `json:"visible"`
		Hidden  string `json:"-"`
	}

	schema, err := chat.NewJSONSchema(&WithSkip{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Properties) != 1 {
		t.Fatalf("expected 1 property, got %d", len(schema.Properties))
	}
	if _, ok := schema.Properties["visible"]; !ok {
		t.Error("expected visible property")
	}
}

func TestNewJSONSchema_LLMTag(t *testing.T) {
	type WithDesc struct {
		Name string `json:"name" llm:"The full name of the person"`
	}

	schema, err := chat.NewJSONSchema(&WithDesc{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Properties["name"].Description != "The full name of the person" {
		t.Errorf("expected description from llm tag, got %q", schema.Properties["name"].Description)
	}
}

func TestNewJSONSchema_LLMIgnore(t *testing.T) {
	type WithIgnore struct {
		Name     string `json:"name"`
		Internal string `json:"internal" llm:"ignore"`
	}

	schema, err := chat.NewJSONSchema(&WithIgnore{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schema.Properties) != 1 {
		t.Fatalf("expected 1 property, got %d", len(schema.Properties))
	}
}

func TestNewJSONSchema_EmbeddedStruct(t *testing.T) {
	type Base struct {
		ID string `json:"id"`
	}
	type Extended struct {
		Base
		Name string `json:"name"`
	}

	schema, err := chat.NewJSONSchema(&Extended{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := schema.Properties["id"]; !ok {
		t.Error("expected embedded id property")
	}
	if _, ok := schema.Properties["name"]; !ok {
		t.Error("expected name property")
	}
}

func TestNewJSONSchema_NilReturnsError(t *testing.T) {
	_, err := chat.NewJSONSchema(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestNewJSONSchema_NonStructReturnsError(t *testing.T) {
	s := "hello"
	_, err := chat.NewJSONSchema(&s)
	if err == nil {
		t.Fatal("expected error for non-struct input")
	}
}

func TestNewJSONSchema_RawMessage(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	schema, err := chat.NewJSONSchema(&Simple{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	raw := schema.RawMessage()
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("RawMessage produced invalid JSON: %v", err)
	}
	if parsed["type"] != "object" {
		t.Errorf("expected type=object in raw JSON")
	}
}

func TestNewJSONSchema_RecursiveStruct(t *testing.T) {
	type Node struct {
		Value    string `json:"value"`
		Children []Node `json:"children"`
	}

	schema, err := chat.NewJSONSchema(&Node{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	children := schema.Properties["children"]
	if children.Type != "array" {
		t.Fatalf("expected array, got %s", children.Type)
	}
	// The recursive item should resolve to object (cycle guard)
	if children.Items == nil || children.Items.Type != "object" {
		t.Errorf("expected recursive items to be object")
	}
}

func assertContains(t *testing.T, slice []string, item string) {
	t.Helper()
	for _, s := range slice {
		if s == item {
			return
		}
	}
	t.Errorf("expected %v to contain %q", slice, item)
}

func assertNotContains(t *testing.T, slice []string, item string) {
	t.Helper()
	for _, s := range slice {
		if s == item {
			t.Errorf("expected %v to NOT contain %q", slice, item)
			return
		}
	}
}
