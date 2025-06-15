package db

import (
	"database/sql"
	"file4you/internal/trees"

	"github.com/google/uuid"
)

// WorkspaceDBProvider is the interface for workspace database operations
type WorkspaceDBProvider interface {
	Connect(dsn string) (*sql.DB, error)
	Close() error
	InitSchema() error
	InsertFileMetadata(meta *trees.FileMetadata) error
	GetFileMetadata(filePath string) (*trees.FileMetadata, error)
	GetAllFileMetadata() ([]*trees.FileMetadata, error)
	DeleteFileMetadata(filePath string) error
	UpdateFileMetadata(meta *trees.FileMetadata) error
}

// ICentralDBProvider is the interface for central database operations (using I prefix to avoid naming conflict)
type ICentralDBProvider interface {
	Connect(dsn string) (*sql.DB, error)
	Close() error
	InitSchema() error
	// Snapshot methods
	InsertSnapshot(snapshot *Snapshot) (uuid.UUID, error)
	GetSnapshot(id uuid.UUID) (*Snapshot, error)
	GetLatestSnapshot() (*Snapshot, error)
	GetAllSnapshots() ([]Snapshot, error)
	// Workspace methods
	AddWorkspace(rootPath, config string) (*Workspace, error)
	UpdateWorkspaceConfig(workspaceID uuid.UUID, config string) (bool, error)
	GetWorkspace(id uuid.UUID) (*Workspace, error)
	GetWorkspacePath(workspaceID uuid.UUID) (string, error)
	GetWorkspaceID(rootPath string) (int, error)
	GetWorkspaceConfig(workspaceID uuid.UUID) (string, error)
	SetWorkspaceConfig(workspaceID uuid.UUID, config string) error
	DeleteWorkspace(workspaceID uuid.UUID) error
	ListWorkspaces() ([]Workspace, error)
	WorkspaceExists(workspaceID uuid.UUID) (bool, error)
	// DirectoryTree methods
	GetDirectoryTree() *trees.DirectoryTree
	SetDirectoryTree(tree *trees.DirectoryTree)
	// Backup method
	Backup() (string, error)
}
