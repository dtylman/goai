package ocr

// Option configures an OCR Task.
type Option func(*Task)

// WithProjectContext sets document-level metadata for OCR cleaning.
func WithProjectContext(ctx *ProjectContext) Option {
	return func(t *Task) {
		t.project = ctx
	}
}

// WithSystemPrompt overrides the default system prompt template.
func WithSystemPrompt(tmpl string) Option {
	return func(t *Task) {
		t.promptOverrides["clean/system"] = tmpl
	}
}

// WithUserPrompt overrides the default user prompt template.
func WithUserPrompt(tmpl string) Option {
	return func(t *Task) {
		t.promptOverrides["clean/user"] = tmpl
	}
}
