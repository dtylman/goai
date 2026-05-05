package ocr

// Segment represents a piece of text extracted via OCR.
type Segment struct {
	// Text is the raw OCR text.
	Text string `json:"text"`
	// FontSize is the detected font size, if available.
	FontSize float64 `json:"font_size,omitempty"`
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
	ID   string `json:"id"`
	Text string `json:"text"`
}

// Result represents cleaned and structured OCR output.
type Result struct {
	Header    string      `json:"header"`
	Body      []Paragraph `json:"body"`
	Footer    string      `json:"footer"`
	Footnotes string      `json:"footnotes"`
	Comments  string      `json:"comments"`
}

// ProjectContext provides metadata about the document being processed.
type ProjectContext struct {
	Title    string
	Author   string
	Genre    string
	Synopsis string
}
