package ocr

import (
	"context"
	"fmt"

	"github.com/dtylman/goai/chat"
)

// Task orchestrates OCR text cleanup.
type Task struct {
	client          chat.Client
	project         *ProjectContext
	promptOverrides map[string]string
}

// New creates a new OCR cleanup Task with the given client and options.
func New(client chat.Client, opts ...Option) *Task {
	t := &Task{
		client:          client,
		promptOverrides: make(map[string]string),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Clean processes raw OCR segments and returns structured, cleaned text.
func (t *Task) Clean(ctx context.Context, req *Request) (*Result, error) {
	data := &promptData{
		Page:           req.Page,
		Segments:       req.Segments,
		ProjectContext: t.project,
	}

	systemPrompt, err := t.renderPrompt("system", data)
	if err != nil {
		return nil, fmt.Errorf("render system prompt: %w", err)
	}
	userPrompt, err := t.renderPrompt("user", data)
	if err != nil {
		return nil, fmt.Errorf("render user prompt: %w", err)
	}

	chatReq := &chat.Request{
		Messages: []chat.Message{
			{Role: chat.RoleSystem, Content: systemPrompt},
			{Role: chat.RoleUser, Content: userPrompt},
		},
	}

	var result Result
	if _, err := chat.ChatInto(ctx, t.client, chatReq, &result); err != nil {
		return nil, fmt.Errorf("clean page %d: %w", req.Page, err)
	}

	return &result, nil
}
