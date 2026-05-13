package ocr

// Segment represents a piece of text extracted via OCR.
type Segment struct {
	// Text is the raw OCR text.
	Text string `json:"text" llm:"The raw OCR text"`
	// FontSize is the detected font size, if available.
	FontSize float64 `json:"font_size,omitempty" llm:"The detected font size"`
}

// Request represents a request to clean OCR text.
type Request struct {
	// Page is the page number being processed.
	Page int
	// Segments contains the raw OCR text segments to clean.
	Segments []Segment
}

// Paragraph represents a single cleaned paragraph.
type Paragraph struct {
	ID   string `json:"id" llm:"A unique identifier for this paragraph"`
	Text string `json:"text" llm:"The cleaned paragraph text"`
}

// Result represents cleaned and structured OCR output.
type Result struct {
	Header    string      `json:"header" llm:"The page header text, if any"`
	Body      []Paragraph `json:"body" llm:"The cleaned body paragraphs"`
	Footer    string      `json:"footer" llm:"The page footer text, if any"`
	Footnotes string      `json:"footnotes" llm:"Footnote text found on the page"`
	Comments  string      `json:"comments" llm:"Any notes or observations about the OCR cleanup"`
}

// ProjectContext provides metadata about the document being processed.
type ProjectContext struct {
	Title    string
	Author   string
	Genre    string
	Synopsis string
}
