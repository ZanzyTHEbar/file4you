package genkithandler

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Prompt represents a loaded prompt template.
type Prompt struct {
	Name    string
	Content string
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

			prompt := Prompt{
				Name:    promptName,
				Content: string(contentBytes),
			}
			loadedPrompts[promptName] = prompt
			slog.Debug("Loaded prompt", "name", promptName, "path", filePath)
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
// This would typically be a method on a struct that holds the loaded prompts.
// For now, it's a placeholder if we were to pass the map around.
func GetPrompt(name string, prompts map[string]Prompt) (Prompt, bool) {
	p, ok := prompts[name]
	return p, ok
}
