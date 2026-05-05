package translate

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed embedded
var embeddedFS embed.FS

// promptData holds all template parameters for prompt rendering.
type promptData struct {
	SourceLang      string
	TargetLang      string
	Text            string
	Translation     string
	PreviousContext string
	ProjectContext  *ProjectContext
}

// renderPrompt renders a prompt template by step and role.
// It checks overrides first, then falls back to embedded defaults.
func (t *Task) renderPrompt(step, role string, data *promptData) (string, error) {
	// Check caller override
	key := step + "/" + role
	if override, ok := t.promptOverrides[key]; ok {
		return executeTemplate(key, override, data)
	}

	// Load from embedded FS
	style := t.style
	if style == "" {
		style = "strict"
	}
	path := fmt.Sprintf("embedded/%s/%s_%s.tmpl", style, step, role)
	content, err := embeddedFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("prompt template not found: %s (style=%s): %w", key, style, err)
	}

	return executeTemplate(key, string(content), data)
}

func executeTemplate(name, text string, data *promptData) (string, error) {
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

// formatPreviousContext formats previous source/target pairs for inclusion in prompts.
func formatPreviousContext(sourceLang, targetLang string, prevSource, prevTarget []string) string {
	if len(prevSource) == 0 {
		return ""
	}
	var buf bytes.Buffer
	for i := range prevSource {
		fmt.Fprintf(&buf, "[%s]: %s\n", sourceLang, prevSource[i])
		if i < len(prevTarget) && prevTarget[i] != "" {
			fmt.Fprintf(&buf, "[%s]: %s\n", targetLang, prevTarget[i])
		}
		buf.WriteString("\n")
	}
	return buf.String()
}
