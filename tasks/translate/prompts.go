package translate

// promptData holds all template parameters for prompt rendering.
type promptData struct {
	SourceLang      string
	TargetLang      string
	Text            string
	Translation     string
	PreviousContext string
	ProjectContext  *ProjectContext
}
