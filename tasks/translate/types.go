package translate

// Character represents a character in the source material.
type Character struct {
	Name        string `json:"name" llm:"The character name"`
	Gender      string `json:"gender" llm:"The character gender"`
	Age         int    `json:"age,omitempty" llm:"The character age"`
	Role        string `json:"role" llm:"The character role in the story"`
	Description string `json:"description,omitempty" llm:"A brief description of the character"`
}

// ProjectContext provides metadata about the work being translated.
type ProjectContext struct {
	Title        string            `json:"title" llm:"The title of the work being translated"`
	Author       string            `json:"author" llm:"The author of the work being translated"`
	Genre        string            `json:"genre" llm:"The genre of the work being translated"`
	Synopsis     string            `json:"synopsis" llm:"A brief synopsis of the work being translated"`
	WritingStyle string            `json:"writing_style" llm:"The writing style of the work being translated"`
	Glossary     map[string]string `json:"glossary" llm:"A glossary of terms for the work being translated"`
	Characters   []Character       `json:"characters" llm:"A list of characters in the work being translated"`
}

// Request represents a single translation request.
type Request struct {
	ProjectContext *ProjectContext `json:"project_context,omitempty" llm:"Metadata about the work being translated, which may be used to inform the translation. This can be omitted if the Task was created with a ProjectContext."`
	// SourceLanguage is the language of the input text.
	SourceLanguage string `json:"source_language" llm:"The language of the input text, e.g. \"English\" or \"Chinese\""`
	// TargetLanguage is the language of the output text.
	TargetLanguage string `json:"target_language" llm:"The language of the output text, e.g. \"English\" or \"Chinese\""`
	// Text is the paragraph to translate.
	Text string `json:"text" llm:"The paragraph to translate"`
	// PreviousSource contains preceding source paragraphs for context.
	PreviousSource []string `json:"previous_source,omitempty" llm:"Preceding source paragraphs for context, ordered from oldest to most recent"`
	// PreviousTarget contains the corresponding previous translations.
	PreviousTarget []string `json:"previous_target,omitempty" llm:"The corresponding previous translations for context, ordered from oldest to most recent"`
	// Style overrides the task-level style for this request.
	Style string `json:"style,omitempty" llm:"The desired writing style for the translation, e.g. \"formal\", \"informal\", \"literary\", etc. Overrides the task-level style if set."`
}

// ProofreadRequest is the input for the proofreader.
type ProofreadRequest struct {
	TranslationReq  *Request `json:"translation_req" llm:"The original translation request containing source text and context"`
	DraftText       string   `json:"draft_text" llm:"The draft translation to be proofread"`
	TranslatorNotes string   `json:"translator_notes,omitempty" llm:"Comments left by the original translator"`
}

// FixRequest is the input when fixing a rejected/broken translation.
type FixRequest struct {
	TranslationReq *Request `json:"translation_req" llm:"The original translation request"`
	DraftText      string   `json:"draft_text" llm:"The flawed translation"`
}

// Result represents the output of a translation.
type Result struct {
	Translation string `json:"translation" llm:"The translated text, without any commentary or notes"`
	Comments    string `json:"comments,omitempty" llm:"Any translation notes, reasoning, or commentary"`
}
