package genkithandler

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"file4you/internal/filesystem/trees"
)

// Prompt represents a loaded prompt template.
type Prompt struct {
	Name     string
	Content  string
	Template *template.Template
}

// PromptContext contains data that can be used in prompt templates
type PromptContext struct {
	Files         []FileInfo        `json:"files"`
	DirectoryPath string            `json:"directory_path"`
	FileCount     int               `json:"file_count"`
	FileTypes     map[string]int    `json:"file_types"`
	TotalSize     int64             `json:"total_size"`
	Metadata      map[string]string `json:"metadata"`
	UserContext   string            `json:"user_context,omitempty"`
}

// FileInfo represents file information for prompt context
type FileInfo struct {
	Name      string                 `json:"name"`
	Path      string                 `json:"path"`
	Size      int64                  `json:"size"`
	Extension string                 `json:"extension"`
	IsDir     bool                   `json:"is_dir"`
	ModTime   string                 `json:"mod_time"`
	Checksum  string                 `json:"checksum,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// LoadPrompts loads all .prompt files from the given directory.
// It returns a map of prompt name (filename without extension) to Prompt struct.
func LoadPrompts(promptsDir string) (map[string]Prompt, error) {
	loadedPrompts := make(map[string]Prompt)

	if promptsDir == "" {
		slog.Warn("Prompts directory is not configured. No prompts will be loaded.")
		return loadedPrompts, nil
	}

	slog.Info("Loading prompts from directory", "directory", promptsDir)

	entries, err := os.ReadDir(promptsDir)
	if err != nil {
		// If the directory doesn't exist, it might not be an error if prompts are optional.
		// For now, we'll log a warning and return an empty map.
		// If prompts are critical, this should return an error.
		slog.Warn("Failed to read prompts directory. No prompts will be loaded.", "directory", promptsDir, "error", err)
		return loadedPrompts, nil // Or return nil, fmt.Errorf("failed to read prompts directory %s: %w", promptsDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".prompt") {
			promptName := strings.TrimSuffix(entry.Name(), ".prompt")
			filePath := filepath.Join(promptsDir, entry.Name())

			contentBytes, err := os.ReadFile(filePath)
			if err != nil {
				slog.Error("Failed to read prompt file", "path", filePath, "error", err)
				continue // Skip this prompt
			}

			content := string(contentBytes)

			// Create template for dynamic prompt generation
			tmpl, err := template.New(promptName).Parse(content)
			if err != nil {
				slog.Warn("Failed to parse prompt template, using as static content",
					"name", promptName, "error", err)
				tmpl = nil
			}

			prompt := Prompt{
				Name:     promptName,
				Content:  content,
				Template: tmpl,
			}
			loadedPrompts[promptName] = prompt
			slog.Debug("Loaded prompt", "name", promptName, "path", filePath, "has_template", tmpl != nil)
		}
	}

	if len(loadedPrompts) == 0 {
		slog.Info("No .prompt files found or loaded.", "directory", promptsDir)
	} else {
		slog.Info(fmt.Sprintf("Successfully loaded %d prompts.", len(loadedPrompts)), "directory", promptsDir)
	}

	return loadedPrompts, nil
}

// GetPrompt retrieves a loaded prompt by name.
func GetPrompt(name string, prompts map[string]Prompt) (Prompt, bool) {
	p, ok := prompts[name]
	return p, ok
}

// RenderPrompt renders a prompt template with the given context
func RenderPrompt(prompt Prompt, context PromptContext) (string, error) {
	if prompt.Template == nil {
		// Static prompt, return as-is
		return prompt.Content, nil
	}

	var buf bytes.Buffer
	err := prompt.Template.Execute(&buf, context)
	if err != nil {
		return "", fmt.Errorf("failed to render prompt template %s: %w", prompt.Name, err)
	}

	return buf.String(), nil
}

// CreateFileOrganizationContext creates a prompt context from file metadata
func CreateFileOrganizationContext(files []trees.FileMetadata, directoryPath string, userContext string) PromptContext {
	context := PromptContext{
		Files:         make([]FileInfo, 0, len(files)),
		DirectoryPath: directoryPath,
		FileCount:     len(files),
		FileTypes:     make(map[string]int),
		TotalSize:     0,
		Metadata:      make(map[string]string),
		UserContext:   userContext,
	}

	for _, file := range files {
		// Extract file name from path
		fileName := filepath.Base(file.FilePath)

		// Extract extension
		extension := filepath.Ext(fileName)
		if extension != "" && len(extension) > 1 {
			extension = extension[1:] // Remove the dot
		}

		fileInfo := FileInfo{
			Name:      fileName,
			Path:      file.FilePath,
			Size:      file.Size,
			Extension: extension,
			IsDir:     file.IsDir,
			ModTime:   file.ModTime.Format("2006-01-02 15:04:05"),
			Checksum:  file.Checksum,
			Metadata:  make(map[string]interface{}),
		}

		context.Files = append(context.Files, fileInfo)
		context.TotalSize += file.Size

		// Count file types by extension
		if !file.IsDir && extension != "" {
			context.FileTypes[extension]++
		} else if file.IsDir {
			context.FileTypes["directory"]++
		} else {
			context.FileTypes["unknown"]++
		}
	}

	return context
}
