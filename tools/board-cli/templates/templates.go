package templates

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed *.md
var templateFS embed.FS

type TemplateData struct {
	Title       string
	Description string
}

func RenderReadme(elemType string, data TemplateData) (string, error) {
	name := fmt.Sprintf("%s_readme.md", elemType)
	return render(name, data)
}

func RenderProgress(elemType string) (string, error) {
	name := fmt.Sprintf("%s_progress.md", elemType)
	content, err := templateFS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("reading template %s: %w", name, err)
	}
	return string(content), nil
}

func render(name string, data TemplateData) (string, error) {
	content, err := templateFS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("reading template %s: %w", name, err)
	}
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template %s: %w", name, err)
	}
	return buf.String(), nil
}
