package docs

import (
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {
	generator := NewGenerator()

	mod := &module.Module{
		ID:          "test-module",
		Name:        "Test Module",
		Version:     "1.0.0",
		Description: "A test module",
		Author:      "Test Author",
		License:     "MIT",
		Dependencies: []*module.Dependency{
			{
				Name:    "dep1",
				Version: "1.0.0",
			},
		},
		Variables: []*module.Variable{
			{
				Name:        "var1",
				Type:        "string",
				Description: "Test variable",
				Required:    true,
				Default:     "default",
			},
		},
		Resources: []*module.Resource{
			{
				Type:        "test",
				Provider:    "mock",
				Description: "Test resource",
				Properties: map[string]*module.Property{
					"prop1": {
						Type:        "string",
						Description: "Test property",
						Required:    true,
					},
				},
			},
		},
		Tests: []*module.Test{
			{
				Name:        "test1",
				Description: "Test case 1",
				Variables: map[string]interface{}{
					"var1": "value1",
				},
				ExpectedOutputs: map[string]interface{}{
					"output1": "expected1",
				},
				Assertions: []string{
					"variable var1 equals value1",
					"output output1 equals expected1",
				},
			},
		},
	}

	// Test Markdown generation
	markdown, err := generator.Generate(mod, FormatMarkdown)
	require.NoError(t, err)
	assert.NotEmpty(t, markdown)

	content := string(markdown)
	assert.Contains(t, content, "# Test Module")
	assert.Contains(t, content, "## Overview")
	assert.Contains(t, content, "**ID**: test-module")
	assert.Contains(t, content, "**Version**: 1.0.0")
	assert.Contains(t, content, "**Description**: A test module")
	assert.Contains(t, content, "**Author**: Test Author")
	assert.Contains(t, content, "**License**: MIT")
	assert.Contains(t, content, "dep1 (1.0.0)")
	assert.Contains(t, content, "### var1")
	assert.Contains(t, content, "**Type**: string")
	assert.Contains(t, content, "**Description**: Test variable")
	assert.Contains(t, content, "**Required**: true")
	assert.Contains(t, content, "**Default**: default")
	assert.Contains(t, content, "### test")
	assert.Contains(t, content, "**Provider**: mock")
	assert.Contains(t, content, "**Description**: Test resource")
	assert.Contains(t, content, "**prop1**: Test property")
	assert.Contains(t, content, "### test1")
	assert.Contains(t, content, "Test case 1")
	assert.Contains(t, content, "var1: value1")
	assert.Contains(t, content, "output1: expected1")
	assert.Contains(t, content, "- variable var1 equals value1")
	assert.Contains(t, content, "- output output1 equals expected1")

	// Test HTML generation
	html, err := generator.Generate(mod, FormatHTML)
	require.NoError(t, err)
	assert.NotEmpty(t, html)

	content = string(html)
	assert.Contains(t, content, "<title>Test Module</title>")
	assert.Contains(t, content, "<h1>Test Module</h1>")
	assert.Contains(t, content, "<strong>ID:</strong> test-module")
	assert.Contains(t, content, "<strong>Version:</strong> 1.0.0")
	assert.Contains(t, content, "<strong>Description:</strong> A test module")
	assert.Contains(t, content, "<strong>Author:</strong> Test Author")
	assert.Contains(t, content, "<strong>License:</strong> MIT")
	assert.Contains(t, content, "dep1 (1.0.0)")
	assert.Contains(t, content, "<h3>var1</h3>")
	assert.Contains(t, content, "<strong>Type:</strong> string")
	assert.Contains(t, content, "<strong>Description:</strong> Test variable")
	assert.Contains(t, content, "<strong>Required:</strong> true")
	assert.Contains(t, content, "<strong>Default:</strong> default")
	assert.Contains(t, content, "<h3>test</h3>")
	assert.Contains(t, content, "<strong>Provider:</strong> mock")
	assert.Contains(t, content, "<strong>Description:</strong> Test resource")
	assert.Contains(t, content, "<strong>prop1:</strong> Test property")
	assert.Contains(t, content, "<h3>test1</h3>")
	assert.Contains(t, content, "Test case 1")
	assert.Contains(t, content, "<strong>var1:</strong> value1")
	assert.Contains(t, content, "<strong>output1:</strong> expected1")
	assert.Contains(t, content, "variable var1 equals value1")
	assert.Contains(t, content, "output output1 equals expected1")

	// Test index generation
	modules := []*module.Module{mod}

	// Test Markdown index
	markdownIndex, err := generator.GenerateIndex(modules, FormatMarkdown)
	require.NoError(t, err)
	assert.NotEmpty(t, markdownIndex)

	content = string(markdownIndex)
	assert.Contains(t, content, "# Module Index")
	assert.Contains(t, content, "## Test Module")
	assert.Contains(t, content, "**ID**: test-module")
	assert.Contains(t, content, "**Version**: 1.0.0")
	assert.Contains(t, content, "**Description**: A test module")
	assert.Contains(t, content, "**Author**: Test Author")
	assert.Contains(t, content, "[View Details](test-module.md)")

	// Test HTML index
	htmlIndex, err := generator.GenerateIndex(modules, FormatHTML)
	require.NoError(t, err)
	assert.NotEmpty(t, htmlIndex)

	content = string(htmlIndex)
	assert.Contains(t, content, "<title>Module Index</title>")
	assert.Contains(t, content, "<h1>Module Index</h1>")
	assert.Contains(t, content, "<h2>Test Module</h2>")
	assert.Contains(t, content, "<strong>ID:</strong> test-module")
	assert.Contains(t, content, "<strong>Version:</strong> 1.0.0")
	assert.Contains(t, content, "<strong>Description:</strong> A test module")
	assert.Contains(t, content, "<strong>Author:</strong> Test Author")
	assert.Contains(t, content, "<a href=\"test-module.html\">View Details</a>")

	// Test unsupported format
	_, err = generator.Generate(mod, Format("unsupported"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")

	_, err = generator.GenerateIndex(modules, Format("unsupported"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestFormatTime(t *testing.T) {
	now := time.Now()
	formatted := formatTime(now)
	assert.Equal(t, now.Format("2006-01-02 15:04:05"), formatted)
}
