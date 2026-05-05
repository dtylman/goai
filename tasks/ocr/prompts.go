package ocr

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed embedded
var embeddedFS embed.FS

// promptData holds all template parameters for prompt rendering.
type promptData struct {
	Page           int
	Segments       []Segment
	ProjectContext *ProjectContext
}

// renderPrompt renders a prompt template by role (system or user).
func (t *Task) renderPrompt(role string, data *promptData) (string, error) {
	key := "clean/" + role
	if override, ok := t.promptOverrides[key]; ok {
		return executeTemplate(key, override, data)
	}

	path := "embedded/clean_" + role + ".tmpl"
	content, err := embeddedFS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return executeTemplate(key, string(content), data)
}

func executeTemplate(name, text string, data *promptData) (string, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
