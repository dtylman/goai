package ocr_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/dtylman/goai/chat"
	"github.com/dtylman/goai/tasks/ocr"
)

type mockClient struct {
	lastMessages []chat.Message
	lastSchema   *chat.JSONSchema
	response     string
}

func (m *mockClient) Chat(_ context.Context, req *chat.Request) (*chat.Response, error) {
	m.lastMessages = req.Messages
	m.lastSchema = req.Schema
	return &chat.Response{Content: m.response}, nil
}

func TestClean_Basic(t *testing.T) {
	result := &ocr.Result{
		Header: "Chapter 1",
		Body: []ocr.Paragraph{
			{ID: "p1", Text: "It was a dark and stormy night."},
			{ID: "p2", Text: "The wind howled through the trees."},
		},
		Footer:    "— 42 —",
		Footnotes: "",
		Comments:  "",
	}
	respJSON, _ := json.Marshal(result)

	mock := &mockClient{response: string(respJSON)}
	task := ocr.New(chat.SingleClient(mock))

	got, err := task.Clean(context.Background(), &ocr.Request{
		Page: 42,
		Segments: []ocr.Segment{
			{Text: "Chapter 1", FontSize: 18},
			{Text: "It was a dark and stormy night.", FontSize: 12},
			{Text: "The wind howled through the trees.", FontSize: 12},
			{Text: "— 42 —", FontSize: 10},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Header != "Chapter 1" {
		t.Errorf("header = %q, want %q", got.Header, "Chapter 1")
	}
	if len(got.Body) != 2 {
		t.Fatalf("body len = %d, want 2", len(got.Body))
	}
	if got.Body[0].Text != "It was a dark and stormy night." {
		t.Errorf("body[0] = %q", got.Body[0].Text)
	}
	if got.Footer != "— 42 —" {
		t.Errorf("footer = %q", got.Footer)
	}
}

func TestClean_UsesSchema(t *testing.T) {
	result := &ocr.Result{Body: []ocr.Paragraph{{ID: "p1", Text: "hello"}}}
	respJSON, _ := json.Marshal(result)

	mock := &mockClient{response: string(respJSON)}
	task := ocr.New(chat.SingleClient(mock))

	_, err := task.Clean(context.Background(), &ocr.Request{
		Page:     1,
		Segments: []ocr.Segment{{Text: "hello", FontSize: 12}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastSchema == nil {
		t.Fatal("expected schema to be set on request")
	}
}

func TestClean_WithProjectContext(t *testing.T) {
	result := &ocr.Result{Body: []ocr.Paragraph{{ID: "p1", Text: "text"}}}
	respJSON, _ := json.Marshal(result)

	mock := &mockClient{response: string(respJSON)}
	task := ocr.New(chat.SingleClient(mock), ocr.WithProjectContext(&ocr.ProjectContext{
		Title:  "War and Peace",
		Author: "Tolstoy",
		Genre:  "novel",
	}))

	_, err := task.Clean(context.Background(), &ocr.Request{
		Page:     1,
		Segments: []ocr.Segment{{Text: "some text", FontSize: 12}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sys := mock.lastMessages[0].Content
	if !strings.Contains(sys, "War and Peace") {
		t.Errorf("system prompt missing title: %s", sys)
	}
	if !strings.Contains(sys, "Tolstoy") {
		t.Errorf("system prompt missing author: %s", sys)
	}
	if !strings.Contains(sys, "novel") {
		t.Errorf("system prompt missing genre: %s", sys)
	}
}

func TestClean_UserPromptContainsSegments(t *testing.T) {
	result := &ocr.Result{Body: []ocr.Paragraph{{ID: "p1", Text: "text"}}}
	respJSON, _ := json.Marshal(result)

	mock := &mockClient{response: string(respJSON)}
	task := ocr.New(chat.SingleClient(mock))

	_, err := task.Clean(context.Background(), &ocr.Request{
		Page: 7,
		Segments: []ocr.Segment{
			{Text: "First segment", FontSize: 14},
			{Text: "Second segment", FontSize: 12},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	user := mock.lastMessages[1].Content
	if !strings.Contains(user, "page 7") {
		t.Errorf("user prompt missing page number: %s", user)
	}
	if !strings.Contains(user, "First segment") {
		t.Errorf("user prompt missing first segment: %s", user)
	}
	if !strings.Contains(user, "Second segment") {
		t.Errorf("user prompt missing second segment: %s", user)
	}
	if !strings.Contains(user, "14.0") {
		t.Errorf("user prompt missing font size: %s", user)
	}
}

func TestClean_WithPromptOverride(t *testing.T) {
	result := &ocr.Result{Body: []ocr.Paragraph{{ID: "p1", Text: "text"}}}
	respJSON, _ := json.Marshal(result)

	mock := &mockClient{response: string(respJSON)}
	task := ocr.New(chat.SingleClient(mock),
		ocr.WithSystemPrompt("Custom system for page {{.Page}}"),
		ocr.WithUserPrompt("Custom user: {{range .Segments}}{{.Text}} {{end}}"),
	)

	_, err := task.Clean(context.Background(), &ocr.Request{
		Page:     3,
		Segments: []ocr.Segment{{Text: "hello", FontSize: 12}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sys := mock.lastMessages[0].Content
	if sys != "Custom system for page 3" {
		t.Errorf("system prompt = %q", sys)
	}
	user := mock.lastMessages[1].Content
	if user != "Custom user: hello " {
		t.Errorf("user prompt = %q", user)
	}
}

func TestClean_MultiModelRouting(t *testing.T) {
	result := &ocr.Result{Body: []ocr.Paragraph{{ID: "p1", Text: "cleaned"}}}
	respJSON, _ := json.Marshal(result)

	cleanMock := &mockClient{response: string(respJSON)}
	defaultMock := &mockClient{response: string(respJSON)}

	router := chat.Map(map[string]chat.Client{
		"clean": cleanMock,
	}, defaultMock)

	task := ocr.New(router)
	_, err := task.Clean(context.Background(), &ocr.Request{
		Page:     1,
		Segments: []ocr.Segment{{Text: "test", FontSize: 12}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleanMock.lastMessages == nil {
		t.Error("expected clean mock to be called")
	}
	if defaultMock.lastMessages != nil {
		t.Error("expected default mock to NOT be called")
	}
}
