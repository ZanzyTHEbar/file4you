package db

import (
	"database/sql"
	"encoding/json"
	"file4you/internal/filesystem/trees"
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

// BatchInsertFiles efficiently inserts multiple file metadata records in a single transaction
func (w *WorkspaceDB) BatchInsertFiles(files []trees.FileMetadata) error {
	if len(files) == 0 {
		return nil
	}

	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO files (id, workspace_id, path, metadata) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, file := range files {
		metadataJSON, err := json.Marshal(file)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for file %s: %w", file.FilePath, err)
		}

		fileID := uuid.New().String()
		_, err = stmt.Exec(fileID, "", file.FilePath, metadataJSON)
		if err != nil {
			return fmt.Errorf("failed to insert file %s (batch item %d): %w", file.FilePath, i, err)
		}
	}

	return tx.Commit()
}

// BatchUpdateFiles efficiently updates multiple file metadata records
func (w *WorkspaceDB) BatchUpdateFiles(updates map[string]trees.FileMetadata) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE files SET metadata = ? WHERE path = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for path, metadata := range updates {
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for file %s: %w", path, err)
		}

		_, err = stmt.Exec(metadataJSON, path)
		if err != nil {
			return fmt.Errorf("failed to update file %s: %w", path, err)
		}
	}

	return tx.Commit()
}

// BatchDeleteFiles efficiently removes multiple file records by their paths
func (w *WorkspaceDB) BatchDeleteFiles(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM files WHERE path = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, path := range paths {
		_, err = stmt.Exec(path)
		if err != nil {
			return fmt.Errorf("failed to delete file %s (batch item %d): %w", path, i, err)
		}
	}

	return tx.Commit()
}

// BatchInsertHistory efficiently inserts multiple history events
func (w *WorkspaceDB) BatchInsertHistory(events []HistoryEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO history (id, event_type, event_json) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, event := range events {
		eventID := uuid.New().String()
		_, err = stmt.Exec(eventID, event.EventType, event.EventJSON)
		if err != nil {
			return fmt.Errorf("failed to insert history event %d: %w", i, err)
		}
	}

	return tx.Commit()
}

// HistoryEvent represents a historical event in the workspace
type HistoryEvent struct {
	EventType string
	EventJSON string
}
