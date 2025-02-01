package docs

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/module"
)

// Format represents the documentation output format
type Format string

const (
	// FormatMarkdown generates documentation in Markdown format
	FormatMarkdown Format = "markdown"
	// FormatHTML generates documentation in HTML format
	FormatHTML Format = "html"
)

// Generator generates documentation for modules
type Generator interface {
	// Generate creates documentation for a module
	Generate(mod *module.Module, format Format) ([]byte, error)

	// GenerateIndex creates an index of all modules
	GenerateIndex(modules []*module.Module, format Format) ([]byte, error)
}

// DefaultGenerator is the default implementation of Generator
type DefaultGenerator struct {
	moduleTemplates map[Format]*template.Template
	indexTemplates  map[Format]*template.Template
}

// NewGenerator creates a new documentation generator
func NewGenerator() Generator {
	return &DefaultGenerator{
		moduleTemplates: make(map[Format]*template.Template),
		indexTemplates:  make(map[Format]*template.Template),
	}
}

// Generate creates documentation for a module
func (g *DefaultGenerator) Generate(mod *module.Module, format Format) ([]byte, error) {
	tmpl, err := g.getTemplate(format)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	data := struct {
		Module    *module.Module
		Generated time.Time
	}{
		Module:    mod,
		Generated: time.Now(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateIndex creates an index of all modules
func (g *DefaultGenerator) GenerateIndex(modules []*module.Module, format Format) ([]byte, error) {
	tmpl, err := g.getIndexTemplate(format)
	if err != nil {
		return nil, fmt.Errorf("failed to get index template: %w", err)
	}

	data := struct {
		Modules   []*module.Module
		Generated time.Time
	}{
		Modules:   modules,
		Generated: time.Now(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// getTemplate returns the template for the specified format
func (g *DefaultGenerator) getTemplate(format Format) (*template.Template, error) {
	if tmpl, ok := g.moduleTemplates[format]; ok {
		return tmpl, nil
	}

	var content string
	switch format {
	case FormatMarkdown:
		content = markdownTemplate
	case FormatHTML:
		content = htmlTemplate
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	tmpl, err := template.New("module").Funcs(template.FuncMap{
		"join":       strings.Join,
		"formatTime": formatTime,
	}).Parse(content)
	if err != nil {
		return nil, err
	}

	g.moduleTemplates[format] = tmpl
	return tmpl, nil
}

// getIndexTemplate returns the index template for the specified format
func (g *DefaultGenerator) getIndexTemplate(format Format) (*template.Template, error) {
	if tmpl, ok := g.indexTemplates[format]; ok {
		return tmpl, nil
	}

	var content string
	switch format {
	case FormatMarkdown:
		content = markdownIndexTemplate
	case FormatHTML:
		content = htmlIndexTemplate
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	tmpl, err := template.New("index").Funcs(template.FuncMap{
		"join":       strings.Join,
		"formatTime": formatTime,
	}).Parse(content)
	if err != nil {
		return nil, err
	}

	g.indexTemplates[format] = tmpl
	return tmpl, nil
}

// formatTime formats a time value for display
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
