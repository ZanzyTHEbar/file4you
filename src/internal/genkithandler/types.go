// Package genkithandler provides integration with the genkit AI platform.
package genkithandler

import (
	"fmt"

	"github.com/google/uuid"
)

// BackupToolInput defines the input for the backup tool.
// For now, it's empty, implying a full backup. It can be extended with options.
type BackupToolInput struct {
	// Example: BackupTarget string (e.g., "deskfs_only", "centraldb_only", "all")
}

// BackupToolOutput defines the output structure for the backup tool.
// This is what an LLM would receive as a structured response if it invoked this tool.
type BackupToolOutput struct {
	DeskFSBackupPath    string `json:"deskFSBackupPath,omitempty"`
	CentralDBBackupPath string `json:"centralDBBackupPath,omitempty"`
	Message             string `json:"message,omitempty"`
}

func (b BackupToolOutput) Error() string {
	// Only return a non-empty string if the message indicates an error or failure
	if b.Message == "" ||
		(b.Message != "" && !(containsSuccess(b.Message))) {
		return fmt.Sprintf("DeskFSBackupPath: %s, CentralDBBackupPath: %s, Message: %s", b.DeskFSBackupPath, b.CentralDBBackupPath, b.Message)
	}
	return ""
}

// containsSuccess checks if the message indicates a successful backup
func containsSuccess(msg string) bool {
	return contains(msg, "success") || contains(msg, "succeed")
}

// contains is a helper for case-insensitive substring search
func contains(s, substr string) bool {
	return len(s) >= len(substr) && // quick check
		(len(s) > 0 && len(substr) > 0 &&
			(stringContainsFold(s, substr)))
}

// stringContainsFold is a case-insensitive substring search
func stringContainsFold(s, substr string) bool {
	s, substr = toLower(s), toLower(substr)
	return len(substr) > 0 && len(s) > 0 && (indexOf(s, substr) >= 0)
}

// toLower returns a lower-case version of the string
func toLower(s string) string {
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	n, m := len(s), len(substr)
	for i := 0; i <= n-m; i++ {
		if s[i:i+m] == substr {
			return i
		}
	}
	return -1
}

// SessionID generates a unique session identifier.
func SessionID() string {
	return uuid.New().String()
}

// AI-powered file organization types

// FileOrganizationResult represents the result of AI-powered file organization
type FileOrganizationResult struct {
	SuggestedActions []OrganizationAction `json:"suggested_actions"`
	Confidence       float64              `json:"confidence"`
	Reasoning        string               `json:"reasoning"`
}

// OrganizationAction represents a specific file organization action
type OrganizationAction struct {
	Action     string   `json:"action"` // "move", "rename", "create_folder", "tag"
	FileName   string   `json:"file_name"`
	SourcePath string   `json:"source_path"`
	TargetPath string   `json:"target_path"`
	NewName    string   `json:"new_name,omitempty"`
	FolderName string   `json:"folder_name,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Confidence float64  `json:"confidence"`
	Reasoning  string   `json:"reasoning"`
}

// FileCategorization represents the result of AI-powered file categorization
type FileCategorization struct {
	Category    string   `json:"category"`
	SubCategory string   `json:"sub_category,omitempty"`
	Tags        []string `json:"tags"`
	Confidence  float64  `json:"confidence"`
	Reasoning   string   `json:"reasoning"`
}

// DuplicateDetectionResult represents the result of AI-powered duplicate detection
type DuplicateDetectionResult struct {
	DuplicateGroups []DuplicateGroup `json:"duplicate_groups"`
	Confidence      float64          `json:"confidence"`
	Reasoning       string           `json:"reasoning"`
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Files      []string `json:"files"`
	Similarity float64  `json:"similarity"`
	Type       string   `json:"type"` // "exact", "content_similar", "name_similar"
	Reasoning  string   `json:"reasoning"`
}
