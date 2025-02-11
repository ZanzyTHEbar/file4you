package deskfs

import (
	"context"
	"desktop-cleaner/internal"
	"desktop-cleaner/internal/db"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ZanzyTHEbar/assert-lib"
	"github.com/google/uuid"
)

type WorkspaceManager struct {
	centralDB     *db.CentralDBProvider
	AssertHandler *assert.AssertHandler
}

func NewWorkspaceManager(centralDB *db.CentralDBProvider, assertHandler *assert.AssertHandler) *WorkspaceManager {
	return &WorkspaceManager{
		centralDB:     centralDB,
		AssertHandler: assertHandler,
	}
}

func createWorkspacePath(rootPath string) string {
	return filepath.Join(rootPath, internal.DefaultWorkspaceDotDir)
}

// CreateWorkspace creates a new workspace, adding it to the central DB and initializing its own DB.
func (wm *WorkspaceManager) CreateWorkspace(rootPath, config string) (uuid.UUID, error) {
	slog.Debug(fmt.Sprintf("Creating workspace at path: %s\n", rootPath))

	rootPath = createWorkspacePath(rootPath)

	slog.Debug(fmt.Sprintf("Workspace path: %s\n", rootPath))
	//mkdirall to check if the directory exists, if not create it
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		if err := os.MkdirAll(rootPath, 0755); err != nil {
			slog.Info(fmt.Sprintf("Path %s: %v", rootPath, err))
			errMsg := fmt.Sprintf("Error creating directory at %s", rootPath)
			ConfigAssertHandler.NoError(context.Background(), err, errMsg, slog.Error)
		}
	}

	// create the ignore file
	ignoreFilePath := filepath.Join(rootPath, ".desktop_cleaner_ignore")

	if _, err := os.Stat(ignoreFilePath); os.IsNotExist(err) {
		if _, err := os.Create(ignoreFilePath); err != nil {
			slog.Info(fmt.Sprintf("Path %s: %v", ignoreFilePath, err))
			errMsg := fmt.Sprintf("Error creating ignore file at %s", ignoreFilePath)
			ConfigAssertHandler.NoError(context.Background(), err, errMsg, slog.Error)
		}

		// Add the `.git` folder to the ignore file
		ignoreFile, err := os.OpenFile(ignoreFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			slog.Info(fmt.Sprintf("Path %s: %v", ignoreFilePath, err))
			errMsg := fmt.Sprintf("Error opening ignore file at %s", ignoreFilePath)
			ConfigAssertHandler.NoError(context.Background(), err, errMsg, slog.Error)
		}

		// Write the `.git` folder to the ignore file
		if _, err := ignoreFile.WriteString(".git\n"); err != nil {
			slog.Info(fmt.Sprintf("Path %s: %v", ignoreFilePath, err))
			errMsg := fmt.Sprintf("Error writing to ignore file at %s", ignoreFilePath)
			ConfigAssertHandler.NoError(context.Background(), err, errMsg, slog.Error)
		}

		defer ignoreFile.Close()
	}

	// Initialize workspace-specific database
	workspaceDB, err := db.NewWorkspaceDB(rootPath)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to initialize workspace database: %v", err)
	}
	defer workspaceDB.Close()

	workspace, err := wm.centralDB.AddWorkspace(rootPath, config)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add workspace: %v", err)
	}

	workspaceID := workspace.ID

	slog.Debug(fmt.Sprintf("Workspace created with ID: %s at path: %s\n", workspaceID, rootPath))
	return workspaceID, nil
}

func (wm *WorkspaceManager) GetWorkspace(workspaceID uuid.UUID) (*db.Workspace, error) {
	workspace, err := wm.centralDB.GetWorkspace(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %v", err)
	}
	return workspace, nil
}

// UpdateWorkspace updates the configuration of an existing workspace by ID.
func (wm *WorkspaceManager) UpdateWorkspace(workspaceID uuid.UUID, newConfig string) error {
	_, err := wm.centralDB.UpdateWorkspaceConfig(workspaceID, newConfig)
	if err != nil {
		return fmt.Errorf("failed to update workspace configuration: %v", err)
	}
	slog.Debug(fmt.Sprintf("Workspace with ID %d updated.\n", workspaceID))
	return nil
}

// DeleteWorkspace deletes a workspace from the central DB and removes its specific database file.
func (wm *WorkspaceManager) DeleteWorkspace(workspaceID uuid.UUID) error {
	// Get the root path of the workspace to delete
	rootPath, err := wm.centralDB.GetWorkspacePath(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to find workspace: %v", err)
	}

	// Delete the workspace entry from the central database
	err = wm.centralDB.DeleteWorkspace(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to delete workspace from central DB: %v", err)
	}

	// Remove the workspace database file

	// Stat the workspace DB file, and if it doesn't exist, return
	rootPath = createWorkspacePath(rootPath)
	workspaceDBPath := filepath.Join(rootPath, "workspace.db")

	if _, err := os.Stat(workspaceDBPath); os.IsNotExist(err) {
		return nil
	}

	if err := os.Remove(workspaceDBPath); err != nil {
		return fmt.Errorf("failed to delete workspace DB file: %v", err)
	}
	slog.Debug(fmt.Sprintf("Workspace with ID %d deleted.\n", workspaceID))
	return nil
}

func (wm *WorkspaceManager) ListWorkspaces() ([]db.Workspace, error) {
	workspaces, err := wm.centralDB.ListWorkspaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %v", err)
	}
	return workspaces, nil
}

/* // InitWorkspace initializes the current directory as a workspace
func Init() error {
	// Initialize the database
	dbAssertHandler := assert.NewAssertHandler()

	workspace := db.NewWorkspace()

	// Create workspace entry
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	_, err = workspaceDB.DB.Exec("INSERT OR IGNORE INTO workspaces (root_path) VALUES (?)", pwd)
	if err != nil {
		return fmt.Errorf("failed to create workspace in database: %w", err)
	}

	// Index the filesystem and create metadata entries
	fs := deskfs.NewDesktopFS(nil)
	directoryTree, err := deskfs.NewDirectoryTree(pwd)
	if err != nil {
		return fmt.Errorf("failed to create directory tree: %w", err)
	}
	fs.DirectoryTree = directoryTree

	// Add metadata and relationships to directory tree
	if err := fs.AddMetadataToTree(directoryTree.Root); err != nil {
		return fmt.Errorf("failed to add metadata: %w", err)
	}
	fs.AddRelationships(fs.DirectoryTree.Root)

	// Store metadata in database
	for path, metadata := range fs.FlattenMetadata(fs.DirectoryTree.Root) {
		_, err := workspaceDB.DB.Exec(
			"INSERT INTO file_metadata (workspace_id, path, metadata_json) VALUES ((SELECT id FROM workspaces WHERE root_path = ?), ?, ?)",
			pwd, path, metadata,
		)
		if err != nil {
			return fmt.Errorf("failed to insert metadata for %s: %w", path, err)
		}
	}

	return nil
} */

/* // AddWorkspace adds a new workspace to the database
func (db *SQLiteWorkspaceDB) AddWorkspace(rootPath string, config string) (int, error) {
	result, err := db.DB.Exec("INSERT OR IGNORE INTO workspaces (root_path, config) VALUES (?, ?)", rootPath, config)
	if err != nil {
		return 0, fmt.Errorf("failed to insert workspace: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	return int(id), nil
}

// AddFileMetadata adds file metadata for a given workspace
func (db *SQLiteWorkspaceDB) AddFileMetadata(workspaceID uuid.UUID, path string, metadata deskfs.Metadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata into JSON: %w", err)
	}

	_, err = db.DB.Exec("INSERT INTO file_metadata (workspace_id, path, metadata_json) VALUES (?, ?, ?)", workspaceID, path, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to insert file metadata: %w", err)
	}

	return nil
}

// UpdateFileMetadata updates the metadata for a given file in the workspace
func (db *SQLiteWorkspaceDB) UpdateFileMetadata(workspaceID uuid.UUID, path string, metadata deskfs.Metadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata into JSON: %w", err)
	}

	_, err = db.DB.Exec("UPDATE file_metadata SET metadata_json = ? WHERE workspace_id = ? AND path = ?", string(metadataJSON), workspaceID, path)
	if err != nil {
		return fmt.Errorf("failed to update file metadata: %w", err)
	}

	return nil
}

// TODO: Read Turso libs on this - StoreVector stores a vector embedding for a specific file
func (db *SQLiteWorkspaceDB) StoreVector(fileID int, vector []float64) error {
	vectorBlob, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("failed to marshal vector into blob: %w", err)
	}

	_, err = db.DB.Exec("INSERT INTO file_vectors (file_id, vector) VALUES (?, ?)", fileID, vectorBlob)
	if err != nil {
		return fmt.Errorf("failed to insert file vector: %w", err)
	}

	return nil
}

// AddHistoryEvent adds a historical event to track workspace changes
func (db *SQLiteWorkspaceDB) AddHistoryEvent(workspaceID uuid.UUID, eventType string, eventJSON string) error {
	_, err := db.DB.Exec("INSERT INTO history (workspace_id, event_type, event_json) VALUES (?, ?, ?)", workspaceID, eventType, eventJSON)
	if err != nil {
		return fmt.Errorf("failed to insert history event: %w", err)
	}

	return nil
}

// GetWorkspaceID gets the workspace ID by root path
func (db *SQLiteWorkspaceDB) GetWorkspaceID(rootPath string) (int, error) {
	var workspaceID uuid.UUID
	err := db.DB.QueryRow("SELECT id FROM workspaces WHERE root_path = ?", rootPath).Scan(&workspaceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get workspace ID: %w", err)
	}

	return workspaceID, nil
}

// Close closes the database connection
func (db *SQLiteWorkspaceDB) Close() error {
	return db.DB.Close()
}

func RebuildWorkspace(dbPath string) error {
	// Initialize the database
	workspaceDB := db.NewWorkspaceDB()

	// Fetch workspace information
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	var workspaceID uuid.UUID
	err = workspaceDB.DB.QueryRow("SELECT id FROM workspaces WHERE root_path = ?", pwd).Scan(&workspaceID)
	if err != nil {
		return fmt.Errorf("workspace not found in database: %w", err)
	}

	// Get the current metadata and compare with the stored version
	// Update the metadata and history accordingly
	// ...

	return nil
} */
