package db

import (
	"database/sql"
	"encoding/json"
	"file4you/internal/trees"
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

// Connect implements WorkspaceDBProvider.Connect
func (w *WorkspaceDB) Connect(dsn string) (*sql.DB, error) {
	var err error
	w.db, err = ConnectToDB(dsn)
	return w.db, err
}

// InitSchema implements WorkspaceDBProvider.InitSchema
func (w *WorkspaceDB) InitSchema() error {
	return w.init()
}

// InsertFileMetadata implements WorkspaceDBProvider.InsertFileMetadata
func (w *WorkspaceDB) InsertFileMetadata(meta *trees.FileMetadata) error {
	metadataJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	fileID := uuid.New().String()
	_, err = w.db.Exec("INSERT INTO files (id, workspace_id, path, metadata) VALUES (?, ?, ?, ?)",
		fileID, "", meta.FilePath, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert file metadata: %w", err)
	}
	return nil
}

// GetFileMetadata implements WorkspaceDBProvider.GetFileMetadata
func (w *WorkspaceDB) GetFileMetadata(filePath string) (*trees.FileMetadata, error) {
	var metadataJSON []byte
	err := w.db.QueryRow("SELECT metadata FROM files WHERE path = ?", filePath).Scan(&metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	var metadata trees.FileMetadata
	err = json.Unmarshal(metadataJSON, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	return &metadata, nil
}

// GetAllFileMetadata implements WorkspaceDBProvider.GetAllFileMetadata
func (w *WorkspaceDB) GetAllFileMetadata() ([]*trees.FileMetadata, error) {
	rows, err := w.db.Query("SELECT metadata FROM files")
	if err != nil {
		return nil, fmt.Errorf("failed to query file metadata: %w", err)
	}
	defer rows.Close()

	var metadataList []*trees.FileMetadata
	for rows.Next() {
		var metadataJSON []byte
		if err := rows.Scan(&metadataJSON); err != nil {
			return nil, fmt.Errorf("failed to scan metadata: %w", err)
		}

		var metadata trees.FileMetadata
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		metadataList = append(metadataList, &metadata)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return metadataList, nil
}

// DeleteFileMetadata implements WorkspaceDBProvider.DeleteFileMetadata
func (w *WorkspaceDB) DeleteFileMetadata(filePath string) error {
	_, err := w.db.Exec("DELETE FROM files WHERE path = ?", filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}
	return nil
}

// UpdateFileMetadata implements WorkspaceDBProvider.UpdateFileMetadata
func (w *WorkspaceDB) UpdateFileMetadata(meta *trees.FileMetadata) error {
	metadataJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = w.db.Exec("UPDATE files SET metadata = ? WHERE path = ?", metadataJSON, meta.FilePath)
	if err != nil {
		return fmt.Errorf("failed to update file metadata: %w", err)
	}
	return nil
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

// Ensure WorkspaceDB implements WorkspaceDBProvider interface
var _ WorkspaceDBProvider = (*WorkspaceDB)(nil)
