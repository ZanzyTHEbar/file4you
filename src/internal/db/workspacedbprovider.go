package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

// WorkspaceDB handles data storage for a specific workspace.
type WorkspaceDB struct {
	db *sql.DB
}

// NewWorkspaceDBProvider opens or initializes a workspace-specific database.
func NewWorkspaceDB(rootPath string) (*WorkspaceDB, error) {
	dbPath := filepath.Join(rootPath, "workspace.db")
	db, err := ConnectToDB(dbPath)
	if err != nil {
		return nil, err
	}

	provider := &WorkspaceDB{db: db}
	if err := provider.init(); err != nil {
		return nil, err
	}
	return provider, nil
}

// init sets up tables for the workspace database.
func (w *WorkspaceDB) init() error {
	createTables := []string{
		`CREATE TABLE IF NOT EXISTS files (id TEXT PRIMARY KEY, workspace_id TEXT, path TEXT, metadata BLOB)`,
		//`CREATE TABLE IF NOT EXISTS vectors (file_id TEXT PRIMARY KEY, vector BLOB)`,
		`CREATE TABLE IF NOT EXISTS history (id TEXT PRIMARY KEY, event_type TEXT, event_json TEXT)`,
	}
	for _, query := range createTables {
		if _, err := w.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (w *WorkspaceDB) GetWorkspace() (*Workspace, error) {
	var workspace Workspace
	err := w.db.QueryRow("SELECT * FROM workspaces").Scan(&workspace.ID, &workspace.RootPath, &workspace.Config)
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

// Close closes the workspace-specific database connection.
func (w *WorkspaceDB) Close() error {
	return w.db.Close()
}

func (w *WorkspaceDB) GetHistory() ([]string, error) {
	rows, err := w.db.Query("SELECT * FROM history")
	if err != nil {
		return nil, err
	}

	var history []string
	for rows.Next() {
		var event string
		if err := rows.Scan(&event); err != nil {
			return nil, err
		}
		history = append(history, event)
	}

	return history, nil
}
func (w *WorkspaceDB) SetHistory([]string) error {
	_, err := w.db.Exec("INSERT INTO history (event) VALUES (?)", "event")
	if err != nil {
		return err
	}

	return err
}

// Utility function to load a workspace database by ID.
func LoadWorkspaceDBProvider(central *CentralDBProvider, workspaceID uuid.UUID) (*WorkspaceDB, error) {
	rootPath, err := central.GetWorkspacePath(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("could not find workspace with ID %d: %v", workspaceID, err)
	}
	return NewWorkspaceDB(rootPath)
}

/* // Example function: AddFileMetadata adds file metadata in a workspace-specific database.
func (w *WorkspaceDB) AddFileMetadata(path string, metadata Metadata) error {
	metadataBlob, err := serializeMetadata(metadata)
	if err != nil {
		return err
	}
	_, err = w.db.Exec("INSERT INTO files (path, metadata) VALUES (?, ?)", path, metadataBlob)
	return err
}

// UpdateFileMetadata updates the metadata for a given file in the workspace
func (w *WorkspaceDB) UpdateFileMetadata(workspaceID int, path string, metadata Metadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata into JSON: %w", err)
	}

	_, err = w.db.Exec("UPDATE file_metadata SET metadata_json = ? WHERE workspace_id = ? AND path = ?", string(metadataJSON), workspaceID, path)
	if err != nil {
		return fmt.Errorf("failed to update file metadata: %w", err)
	}

	return nil
}

// TODO: Read Turso libs on this - StoreVector stores a vector embedding for a specific file
func (w *WorkspaceDB) StoreVector(fileID int, vector []float64) error {
	vectorBlob, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("failed to marshal vector into blob: %w", err)
	}

	_, err = w.db.Exec("INSERT INTO file_vectors (file_id, vector) VALUES (?, ?)", fileID, vectorBlob)
	if err != nil {
		return fmt.Errorf("failed to insert file vector: %w", err)
	}

	return nil
}

// Function to add a historical event to the workspace database.
func (w *WorkspaceDB) AddHistoryEvent(eventType string, eventJSON string) error {
	_, err := w.db.Exec("INSERT INTO history (event_type, event_json) VALUES (?, ?)", eventType, eventJSON)
	return err
} */

/* // serializeMetadata serializes metadata for storage.
func serializeMetadata(metadata Metadata) ([]byte, error) {
	// Implement actual serialization logic here
	return []byte{}, nil
} */
