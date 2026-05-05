package chat_test

import (
	"context"
	"testing"

	"github.com/dtylman/goai/chat"
)

type mockClient struct {
	response string
	err      error
	lastReq  *chat.Request
}

func (m *mockClient) Chat(_ context.Context, req *chat.Request) (*chat.Response, error) {
	m.lastReq = req
	if m.err != nil {
		return nil, m.err
	}
	return &chat.Response{Content: m.response}, nil
}

func TestChatInto_Success(t *testing.T) {
	type Result struct {
		Title string `json:"title"`
		Score int    `json:"score"`
	}

	mock := &mockClient{response: `{"title":"hello","score":42}`}
	var result Result

	resp, err := chat.ChatInto(context.Background(), mock, &chat.Request{
		Messages: []chat.Message{{Role: chat.RoleUser, Content: "test"}},
	}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != `{"title":"hello","score":42}` {
		t.Errorf("unexpected content: %s", resp.Content)
	}
	if result.Title != "hello" || result.Score != 42 {
		t.Errorf("unexpected result: %+v", result)
	}
	if mock.lastReq.Schema == nil {
		t.Fatal("expected schema to be set on request")
	}
	if mock.lastReq.Schema.Type != "object" {
		t.Errorf("expected schema type=object, got %s", mock.lastReq.Schema.Type)
	}
}

func TestChatInto_InvalidJSON(t *testing.T) {
	type Result struct {
		Name string `json:"name"`
	}

	mock := &mockClient{response: `not json`}
	var result Result

	resp, err := chat.ChatInto(context.Background(), mock, &chat.Request{
		Messages: []chat.Message{{Role: chat.RoleUser, Content: "test"}},
	}, &result)
	if err == nil {
		t.Fatal("expected decode error")
	}
	if resp == nil || resp.Content != "not json" {
		t.Errorf("expected raw response to be returned on decode error")
	}
}

func TestChatInto_ProviderError(t *testing.T) {
	type Result struct {
		Name string `json:"name"`
	}

	mock := &mockClient{err: context.DeadlineExceeded}
	var result Result

	_, err := chat.ChatInto(context.Background(), mock, &chat.Request{
		Messages: []chat.Message{{Role: chat.RoleUser, Content: "test"}},
	}, &result)
	if err == nil {
		t.Fatal("expected provider error")
	}
}
