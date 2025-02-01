package docs

const markdownTemplate = `# {{ .Module.Name }}

## Overview

- **ID**: {{ .Module.ID }}
- **Version**: {{ .Module.Version }}
- **Description**: {{ .Module.Description }}
- **Author**: {{ .Module.Author }}
- **License**: {{ .Module.License }}

## Dependencies

{{ range .Module.Dependencies }}
- {{ .Name }} ({{ .Version }})
{{ end }}

## Variables

{{ range .Module.Variables }}
### {{ .Name }}

- **Type**: {{ .Type }}
- **Description**: {{ .Description }}
- **Required**: {{ .Required }}
{{ if .Default }}
- **Default**: {{ .Default }}
{{ end }}
{{ end }}

## Resources

{{ range .Module.Resources }}
### {{ .Type }}

- **Provider**: {{ .Provider }}
- **Description**: {{ .Description }}

#### Properties
{{ range $name, $prop := .Properties }}
- **{{ $name }}**: {{ $prop.Description }}
{{ end }}
{{ end }}

## Tests

{{ range .Module.Tests }}
### {{ .Name }}

{{ .Description }}

#### Variables
{{ range $name, $value := .Variables }}
- {{ $name }}: {{ $value }}
{{ end }}

#### Expected Outputs
{{ range $name, $value := .ExpectedOutputs }}
- {{ $name }}: {{ $value }}
{{ end }}

#### Assertions
{{ range .Assertions }}
- {{ . }}
{{ end }}
{{ end }}

---
Generated on {{ formatTime .Generated }}`

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>{{ .Module.Name }}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        h1, h2, h3 {
            color: #333;
        }
        .metadata {
            background: #f5f5f5;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
        .section {
            margin: 30px 0;
        }
        .resource {
            border: 1px solid #ddd;
            padding: 15px;
            margin: 10px 0;
            border-radius: 5px;
        }
        .test {
            background: #f9f9f9;
            padding: 15px;
            margin: 10px 0;
            border-radius: 5px;
        }
        .footer {
            margin-top: 50px;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <h1>{{ .Module.Name }}</h1>

    <div class="metadata">
        <p><strong>ID:</strong> {{ .Module.ID }}</p>
        <p><strong>Version:</strong> {{ .Module.Version }}</p>
        <p><strong>Description:</strong> {{ .Module.Description }}</p>
        <p><strong>Author:</strong> {{ .Module.Author }}</p>
        <p><strong>License:</strong> {{ .Module.License }}</p>
    </div>

    <div class="section">
        <h2>Dependencies</h2>
        <ul>
        {{ range .Module.Dependencies }}
            <li>{{ .Name }} ({{ .Version }})</li>
        {{ end }}
        </ul>
    </div>

    <div class="section">
        <h2>Variables</h2>
        {{ range .Module.Variables }}
        <div class="resource">
            <h3>{{ .Name }}</h3>
            <p><strong>Type:</strong> {{ .Type }}</p>
            <p><strong>Description:</strong> {{ .Description }}</p>
            <p><strong>Required:</strong> {{ .Required }}</p>
            {{ if .Default }}
            <p><strong>Default:</strong> {{ .Default }}</p>
            {{ end }}
        </div>
        {{ end }}
    </div>

    <div class="section">
        <h2>Resources</h2>
        {{ range .Module.Resources }}
        <div class="resource">
            <h3>{{ .Type }}</h3>
            <p><strong>Provider:</strong> {{ .Provider }}</p>
            <p><strong>Description:</strong> {{ .Description }}</p>
            
            <h4>Properties</h4>
            <ul>
            {{ range $name, $prop := .Properties }}
                <li><strong>{{ $name }}:</strong> {{ $prop.Description }}</li>
            {{ end }}
            </ul>
        </div>
        {{ end }}
    </div>

    <div class="section">
        <h2>Tests</h2>
        {{ range .Module.Tests }}
        <div class="test">
            <h3>{{ .Name }}</h3>
            <p>{{ .Description }}</p>

            <h4>Variables</h4>
            <ul>
            {{ range $name, $value := .Variables }}
                <li><strong>{{ $name }}:</strong> {{ $value }}</li>
            {{ end }}
            </ul>

            <h4>Expected Outputs</h4>
            <ul>
            {{ range $name, $value := .ExpectedOutputs }}
                <li><strong>{{ $name }}:</strong> {{ $value }}</li>
            {{ end }}
            </ul>

            <h4>Assertions</h4>
            <ul>
            {{ range .Assertions }}
                <li>{{ . }}</li>
            {{ end }}
            </ul>
        </div>
        {{ end }}
    </div>

    <div class="footer">
        Generated on {{ formatTime .Generated }}
    </div>
</body>
</html>`

const markdownIndexTemplate = `# Module Index

Generated on {{ formatTime .Generated }}

{{ range .Modules }}
## {{ .Name }}

- **ID**: {{ .ID }}
- **Version**: {{ .Version }}
- **Description**: {{ .Description }}
- **Author**: {{ .Author }}

[View Details]({{ .ID }}.md)

---
{{ end }}`

const htmlIndexTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Module Index</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        h1, h2 {
            color: #333;
        }
        .module {
            border: 1px solid #ddd;
            padding: 15px;
            margin: 10px 0;
            border-radius: 5px;
        }
        .footer {
            margin-top: 50px;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <h1>Module Index</h1>

    {{ range .Modules }}
    <div class="module">
        <h2>{{ .Name }}</h2>
        <p><strong>ID:</strong> {{ .ID }}</p>
        <p><strong>Version:</strong> {{ .Version }}</p>
        <p><strong>Description:</strong> {{ .Description }}</p>
        <p><strong>Author:</strong> {{ .Author }}</p>
        <p><a href="{{ .ID }}.html">View Details</a></p>
    </div>
    {{ end }}

    <div class="footer">
        Generated on {{ formatTime .Generated }}
    </div>
</body>
</html>`
