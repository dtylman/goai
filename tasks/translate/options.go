package translate

// Option configures a translation Task.
type Option func(*Task)

// WithProjectContext sets project-level metadata for all translations.
func WithProjectContext(ctx *ProjectContext) Option {
	return func(t *Task) {
		t.project = ctx
	}
}

// WithStyle sets the default prompt style (e.g., "strict", "literary", "academic").
func WithStyle(style string) Option {
	return func(t *Task) {
		t.style = style
	}
}

// WithAutoProofread enables automatic proofreading after translation.
func WithAutoProofread(enabled bool) Option {
	return func(t *Task) {
		t.autoProofread = enabled
	}
}

// WithSystemPrompt overrides the system prompt template for a given step.
func WithSystemPrompt(step string, tmpl string) Option {
	return func(t *Task) {
		t.promptOverrides[step+"/system"] = tmpl
	}
}

// WithUserPrompt overrides the user prompt template for a given step.
func WithUserPrompt(step string, tmpl string) Option {
	return func(t *Task) {
		t.promptOverrides[step+"/user"] = tmpl
	}
}
