package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"path"

	"github.com/dtylman/goai/chat"
)

//go:embed embedded
var embeddedFS embed.FS

// Render renders a prompt template for the given task, style, and name, using the provided parameters.
func Render(task string, style string, role chat.Role, name string, params any) (string, error) {
	path := path.Join("embedded", task, style, fmt.Sprintf("%v_%v.tmpl", role, name))
	log.Printf("Loading prompt template from %v", path)
	content, err := embeddedFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("prompt template not found: %v: %w", path, err)
	}
	return executeTemplate(name, string(content), params)
}

// RenderFromReader reads a prompt template from the given reader and renders it with the provided parameters.
func RenderFromReader(r io.Reader, params any) (string, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		return "", fmt.Errorf("read prompt template: %w", err)
	}
	return executeTemplate("custom", buf.String(), params)
}

func executeTemplate(name, text string, data any) (string, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Parse(text)
	if err != nil {
		return "", fmt.Errorf("parse prompt template %q: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute prompt template %q: %w", name, err)
	}
	return buf.String(), nil
}
