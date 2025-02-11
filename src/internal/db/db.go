package db

import (
	"time"

	"github.com/google/uuid"
)

type OperationType int

const (
	OperationTypeCreate OperationType = iota
	OperationTypeUpdate
	OperationTypeDelete
	OperationTypeMove
)

type Workspace struct {
	ID        uuid.UUID
	RootPath  string
	Config    string
	Timestamp time.Time
}

type OperationHistory struct {
	ID          uuid.UUID
	NodeID      uuid.UUID
	Operation   OperationType
	NewPath     string
	TimeStamp   time.Time
	PerformedBy string
}

// Example usage:
//func main() {
//	// Initialize central database
//	centralDB, err := NewCentralDBProvider()
//	if err != nil {
//		log.Fatal("Failed to initialize central database:", err)
//	}
//	defer centralDB.Close()
//
//	// Example usage: Add a new workspace
//	workspaceID, err := centralDB.AddWorkspace("/path/to/workspace", "config_data")
//	if err != nil {
//		log.Fatal("Failed to add workspace:", err)
//	}
//	fmt.Println("Added workspace with ID:", workspaceID)
//
//	// Load workspace database by ID
//	workspaceDB, err := LoadWorkspaceDBProvider(centralDB, workspaceID)
//	if err != nil {
//		log.Fatal("Failed to load workspace database:", err)
//	}
//	defer workspaceDB.Close()
//
//	// Example operation in workspace database
//	err = workspaceDB.AddFileMetadata("/path/to/file", deskfs.Metadata{})
//	if err != nil {
//		log.Fatal("Failed to add file metadata:", err)
//	}
//}
