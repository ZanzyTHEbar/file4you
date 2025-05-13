// Package genkithandler provides integration with the genkit AI platform.
package genkithandler

import (
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
	SuccessMessage      string `json:"successMessage"`
}

// SessionID generates a unique session identifier.
func SessionID() string {
	return uuid.New().String()
}
