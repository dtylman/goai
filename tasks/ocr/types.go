package ocr

// ProjectContext provides metadata about the document being processed.
type ProjectContext struct {
	Title    string `json:"title" llm:"The title of the document being processed"`
	Author   string `json:"author" llm:"The author of the document"`
	Genre    string `json:"genre" llm:"The genre of the document"`
	Synopsis string `json:"synopsis" llm:"A brief summary of the document"`
}

// Request represents a request to clean OCR text.
type Request struct {
	Text           string          `json:"text" llm:"The raw OCR text to be cleaned"`
	ProjectContext *ProjectContext `json:"project_context,omitempty" llm:"Metadata about the document being processed"`
}

// Response represents cleaned and structured OCR output.
type Response struct {
	Header    string `json:"header" llm:"The page header text, if any"`
	Body      string `json:"body" llm:"The cleaned text"`
	Footer    string `json:"footer" llm:"The page footer text, if any"`
	Footnotes string `json:"footnotes" llm:"Footnote text found on the page"`
	Comments  string `json:"comments" llm:"Any notes or observations about the OCR cleanup"`
}
