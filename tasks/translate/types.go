package translate

import (
	"fmt"
	"strings"
)

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

// chatResponse is the structured JSON response expected from the model.
// The Comments field gives the model a dedicated place for reasoning,
// notes, and translation commentary so they don't leak into the text.
type chatResponse struct {
	Translation string `json:"translation" llm:"The translated text"`
	Comments    string `json:"comments,omitempty" llm:"Any translation notes, reasoning, or commentary"`
}

// Character represents a character in the source material.
type Character struct {
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	Age         int    `json:"age,omitempty"`
	Role        string `json:"role"`
	Description string `json:"description,omitempty"`
}

// ProjectContext provides metadata about the work being translated.
type ProjectContext struct {
	Title        string
	Author       string
	Genre        string
	Synopsis     string
	WritingStyle string
	Glossary     map[string]string
	Characters   []Character
}

// GlossaryFormatted returns the glossary as a prompt-ready string.
// Returns "" when the glossary is empty.
func (p *ProjectContext) GlossaryFormatted() string {
	if p == nil || len(p.Glossary) == 0 {
		return ""
	}
	var buf strings.Builder
	for term, trans := range p.Glossary {
		fmt.Fprintf(&buf, "  \"%s\" → \"%s\"\n", term, trans)
	}
	return buf.String()
}
