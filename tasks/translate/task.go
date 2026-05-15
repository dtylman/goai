package translate

import (
	"context"
	"fmt"

	"github.com/dtylman/goai/chat"
	"github.com/dtylman/goai/prompts"
)

// Task orchestrates translation workflows.
type Task struct {
	client        chat.Client
	Project       *ProjectContext
	Style         string
	AutoProofread bool
}

// New creates a new translation Task with the given client and options.
func New(client chat.Client, style string, project *ProjectContext) *Task {
	t := &Task{
		client:        client,
		Style:         style,
		Project:       project,
		AutoProofread: true,
	}
	return t
}

// Translate translates a single paragraph.
func (t *Task) Translate(ctx context.Context, req *Request) (*Result, error) {
	result, err := t.doTranslate(ctx, req)
	if err != nil {
		return nil, err
	}

	if t.AutoProofread {
		result, err = t.doProofread(ctx, req, result.Translation)
		if err != nil {
			return nil, fmt.Errorf("proofread: %w", err)
		}
	}

	return result, nil
}

// Proofread proofreads an existing translation.
func (t *Task) Proofread(ctx context.Context, req *Request, translation string) (*Result, error) {
	return t.doProofread(ctx, req, translation)
}

// Fix re-translates a paragraph that was flagged as poor quality.
func (t *Task) Fix(ctx context.Context, req *Request, badTranslation string) (*Result, error) {
	return t.doFix(ctx, req, badTranslation)
}

func (t *Task) doTranslate(ctx context.Context, req *Request) (*Result, error) {
	if req.Text == "" {
		return &Result{Translation: ""}, nil
	}
	if req.Style == "" {
		req.Style = t.Style
	}
	systemPrompt, err := prompts.Render("translate", req.Style, chat.RoleSystem, "translate", req)
	if err != nil {
		return nil, err
	}
	userPrompt, err := prompts.Render("translate", req.Style, chat.RoleUser, "translate", req)
	if err != nil {
		return nil, err
	}

	chatReq := &chat.Request{
		Messages: []chat.Message{
			{Role: chat.RoleSystem, Content: systemPrompt},
			{Role: chat.RoleUser, Content: userPrompt},
		},
	}

	var result Result
	resp, err := chat.ChatInto(ctx, t.client, chatReq, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to translate: %w, %v:", err, resp.Content)
	}
	return &result, nil
}

func (t *Task) doProofread(ctx context.Context, tr *Request, translation string) (*Result, error) {
	req := &ProofreadRequest{
		TranslationReq: tr,
		DraftText:      translation,
	}

	systemPrompt, err := prompts.Render("translate", tr.Style, chat.RoleSystem, "proofread", req)
	if err != nil {
		return nil, err
	}

	userPrompt, err := prompts.Render("translate", tr.Style, chat.RoleUser, "proofread", req)
	if err != nil {
		return nil, err
	}

	chatReq := &chat.Request{
		Messages: []chat.Message{
			{Role: chat.RoleSystem, Content: systemPrompt},
			{Role: chat.RoleUser, Content: userPrompt},
		},
	}

	var result Result
	resp, err := chat.ChatInto(ctx, t.client, chatReq, &result)
	if err != nil {
		return nil, fmt.Errorf("proofread: %w, %v", err, resp.Content)
	}

	return &result, nil
}

func (t *Task) doFix(ctx context.Context, req *Request, badTranslation string) (*Result, error) {
	fixReq := &FixRequest{
		TranslationReq: req,
		DraftText:      badTranslation,
	}

	systemPrompt, err := prompts.Render("translate", "default", chat.RoleSystem, "fix", fixReq)
	if err != nil {
		return nil, err
	}
	userPrompt, err := prompts.Render("translate", "default", chat.RoleUser, "fix", fixReq)
	if err != nil {
		return nil, err
	}

	chatReq := &chat.Request{
		Messages: []chat.Message{
			{Role: chat.RoleSystem, Content: systemPrompt},
			{Role: chat.RoleUser, Content: userPrompt},
		},
	}

	var result Result
	resp, err := chat.ChatInto(ctx, t.client, chatReq, &result)
	if err != nil {
		return nil, fmt.Errorf("fix: %w, %v", err, resp.Content)
	}

	return &result, nil
}
