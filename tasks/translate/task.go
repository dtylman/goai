package translate

import (
	"context"
	"fmt"

	"github.com/dtylman/goai/chat"
)

// Task orchestrates translation workflows.
type Task struct {
	client          chat.Client
	project         *ProjectContext
	style           string
	autoProofread   bool
	promptOverrides map[string]string
}

// New creates a new translation Task with the given client and options.
func New(client chat.Client, opts ...Option) *Task {
	t := &Task{
		client:          client,
		style:           "strict",
		promptOverrides: make(map[string]string),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Translate translates a single paragraph.
func (t *Task) Translate(ctx context.Context, req *Request) (*Result, error) {
	result, err := t.doTranslate(ctx, req)
	if err != nil {
		return nil, err
	}

	if t.autoProofread {
		result, err = t.doProofread(ctx, req, result.Text)
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
	style := req.Style
	if style != "" {
		// Temporarily override style for this request
		origStyle := t.style
		t.style = style
		defer func() { t.style = origStyle }()
	}

	data := &promptData{
		SourceLang:      req.SourceLanguage,
		TargetLang:      req.TargetLanguage,
		Text:            req.Text,
		PreviousContext: formatPreviousContext(req.SourceLanguage, req.TargetLanguage, req.PreviousSource, req.PreviousTarget),
		ProjectContext:  t.project,
	}

	systemPrompt, err := t.renderPrompt("translate", "system", data)
	if err != nil {
		return nil, err
	}
	userPrompt, err := t.renderPrompt("translate", "user", data)
	if err != nil {
		return nil, err
	}

	chatReq := &chat.Request{
		Messages: []chat.Message{
			{Role: chat.RoleSystem, Content: systemPrompt},
			{Role: chat.RoleUser, Content: userPrompt},
		},
	}

	var cr chatResponse
	if _, err := chat.ChatInto(ctx, t.client, chatReq, &cr); err != nil {
		return nil, fmt.Errorf("translate: %w", err)
	}

	return &Result{Text: cr.Translation}, nil
}

func (t *Task) doProofread(ctx context.Context, req *Request, translation string) (*Result, error) {
	data := &promptData{
		SourceLang:     req.SourceLanguage,
		TargetLang:     req.TargetLanguage,
		Text:           req.Text,
		Translation:    translation,
		ProjectContext: t.project,
	}

	systemPrompt, err := t.renderPrompt("proofread", "system", data)
	if err != nil {
		return nil, err
	}
	userPrompt, err := t.renderPrompt("proofread", "user", data)
	if err != nil {
		return nil, err
	}

	chatReq := &chat.Request{
		Messages: []chat.Message{
			{Role: chat.RoleSystem, Content: systemPrompt},
			{Role: chat.RoleUser, Content: userPrompt},
		},
	}

	var cr chatResponse
	if _, err := chat.ChatInto(ctx, t.client, chatReq, &cr); err != nil {
		return nil, fmt.Errorf("proofread: %w", err)
	}

	return &Result{Text: cr.Translation}, nil
}

func (t *Task) doFix(ctx context.Context, req *Request, badTranslation string) (*Result, error) {
	data := &promptData{
		SourceLang:     req.SourceLanguage,
		TargetLang:     req.TargetLanguage,
		Text:           req.Text,
		Translation:    badTranslation,
		ProjectContext: t.project,
	}

	systemPrompt, err := t.renderPrompt("fix", "system", data)
	if err != nil {
		return nil, err
	}
	userPrompt, err := t.renderPrompt("fix", "user", data)
	if err != nil {
		return nil, err
	}

	chatReq := &chat.Request{
		Messages: []chat.Message{
			{Role: chat.RoleSystem, Content: systemPrompt},
			{Role: chat.RoleUser, Content: userPrompt},
		},
	}

	var cr chatResponse
	if _, err := chat.ChatInto(ctx, t.client, chatReq, &cr); err != nil {
		return nil, fmt.Errorf("fix: %w", err)
	}

	return &Result{Text: cr.Translation}, nil
}
