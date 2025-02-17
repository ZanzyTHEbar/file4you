package llm

import (
	"bytes"
	"text/template"
	"time"
)

// FileData holds information about a file for prompt templating.
type FileData struct {
	Filename       string
	FileType       string
	ModDate        string
	ContentPreview string
}

// BuildPrompt builds a prompt from a template using file metadata.
// TODO: The template should be moved to a file for better maintainability.
func BuildPrompt(data FileData) (string, error) {
	const tmplStr = `
You are an expert file organizer.
Given the file metadata below:

Filename: {{.Filename}}
File Type: {{.FileType}}
Modification Date: {{.ModDate}}
Content Preview: {{.ContentPreview}}

Decide the optimal destination folder and suggest a new file name based on the rule:
"Vacation footage goes to the 'Vacation' folder organized by date (YYYY-MM)."

Return your decision as a JSON object with keys "destinationFolder" and "newFileName".
`
	tmpl, err := template.New("filePrompt").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Truncate returns a preview of the content (max n characters).
func Truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

// PrepareFileData prepares FileData from raw file metadata.
func PrepareFileData(filename, fileType, content string, modDate time.Time) FileData {
	return FileData{
		Filename:       filename,
		FileType:       fileType,
		ModDate:        modDate.Format("2006-01-02"),
		ContentPreview: Truncate(content, 200),
	}
}