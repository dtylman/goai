package translate

// Request represents a single translation request.
type Request struct {
	// SourceLanguage is the language of the input text.
	SourceLanguage string
	// TargetLanguage is the language of the output text.
	TargetLanguage string
	// Text is the paragraph to translate.
	Text string
	// PreviousSource contains preceding source paragraphs for context.
	PreviousSource []string
	// PreviousTarget contains the corresponding previous translations.
	PreviousTarget []string
	// Style overrides the task-level style for this request.
	Style string
}

// Result represents the output of a translation.
type Result struct {
	// Text is the translated paragraph.
	Text string
}

// Character represents a character in the source material.
type Character struct {
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Role   string `json:"role"`
}

// ProjectContext provides metadata about the work being translated.
type ProjectContext struct {
	Title      string
	Author     string
	Genre      string
	Synopsis   string
	Characters []Character
}
