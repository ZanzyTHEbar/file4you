package db

import (
	"database/sql"
	"file4you/internal/filesystem/trees"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/tursodatabase/go-libsql"
)

// TODO: Fix logging so that it is optional and outputs to the default unix log directory, stdout|stderr, or the dfault config directory

// TODO: Implement generic database validation methods

// CentralDBProvider tracks the locations of all workspaces.
type CentralDBProvider struct {
	db            *sql.DB
	DirectoryTree *trees.DirectoryTree
}

const centralDBFileName = "central.db"

// NewCentralDBProvider opens or initializes the central database at the binary location.
func NewCentralDBProvider() (*CentralDBProvider, error) {
	homeDir, err := os.UserHomeDir()
	if (err != nil) {
		return nil, fmt.Errorf("could not get user home directory: %v", err)
	}

	// Construct config path
	configPath := filepath.Join(homeDir, ".config", "desktop_cleaner")

	// Ensure the config directory exists
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return nil, fmt.Errorf("could not create config directory: %v", err)
	}

	dbPath := filepath.Join(configPath, centralDBFileName)

	slog.Info("Central database path:", "path", dbPath)

	db, err := ConnectToDB(dbPath)
	if err != nil {
		return nil, err
	}

	provider := &CentralDBProvider{db: db}
	if err := provider.init(); err != nil {
		return nil, err
	}
	return provider, nil
}

// init sets up the central database tables.
func (c *CentralDBProvider) init() error {
	_, err := c.db.Exec(`CREATE TABLE IF NOT EXISTS workspaces (
		id TEXT PRIMARY KEY UNIQUE,
		root_path TEXT,
		config TEXT,
		time_stamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

// AddWorkspace adds a new workspace to the central database and returns its ID.
func (c *CentralDBProvider) AddWorkspace(rootPath, config string) (*Workspace, error) {
	slog.Debug(fmt.Sprintf("Adding workspace with root path %s\n", rootPath))

	c.db.Query("BEGIN TRANSACTION")

	// Create Workspace instance
	workspace := Workspace{
		ID:        uuid.New(),
		RootPath:  rootPath,
		Config:    config,
		Timestamp: time.Now(),
	}
	// Create a new workspace entry in the database
	result, err := c.db.Exec("INSERT INTO workspaces (id, root_path, config) VALUES (?, ?, ?)", workspace.ID, workspace.RootPath, workspace.Config)
	if err != nil {
		c.db.Query("ROLLBACK")
		return nil, fmt.Errorf("failed to insert workspace: %v", err)
	}

	// Check the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.db.Query("ROLLBACK")
		return nil, fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected != 1 {
		c.db.Query("ROLLBACK")
		return nil, fmt.Errorf("expected 1 row affected, got %d", rowsAffected)
	}

	// Commit the transaction
	c.db.Query("END TRANSACTION")

	slog.Debug(fmt.Sprintf("Successfully created Workspace"))

	return &workspace, nil
}

func (c *CentralDBProvider) UpdateWorkspaceConfig(workspaceID uuid.UUID, config string) (bool, error) {
	_, err := c.db.Exec("UPDATE workspaces SET config = ? WHERE id = ?", config, workspaceID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *CentralDBProvider) GetWorkspace(id uuid.UUID) (*Workspace, error) {
	var workspace Workspace
	err := c.db.QueryRow("SELECT * FROM workspaces WHERE id = ?", id).Scan(&workspace.ID, &workspace.RootPath, &workspace.Config)
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

// GetWorkspacePath retrieves the root path of a workspace by its ID.
func (c *CentralDBProvider) GetWorkspacePath(workspaceID uuid.UUID) (string, error) {
	var rootPath string
	err := c.db.QueryRow("SELECT root_path FROM workspaces WHERE id = ?", workspaceID).Scan(&rootPath)
	return rootPath, err
}

func (c *CentralDBProvider) GetWorkspaceID(rootPath string) (int, error) {
	var id int
	err := c.db.QueryRow("SELECT id FROM workspaces WHERE root_path = ?", rootPath).Scan(&id)
	return id, err
}

func (c *CentralDBProvider) GetWorkspaceConfig(workspaceID uuid.UUID) (string, error) {
	var config string
	err := c.db.QueryRow("SELECT config FROM workspaces WHERE id = ?", workspaceID).Scan(&config)
	return config, err
}

func (c *CentralDBProvider) SetWorkspaceConfig(workspaceID uuid.UUID, config string) error {
	_, err := c.db.Exec("UPDATE workspaces SET config = ? WHERE id = ?", config, workspaceID)
	return err
}

func (c *CentralDBProvider) DeleteWorkspace(workspaceID uuid.UUID) error {
	_, err := c.db.Exec("DELETE FROM workspaces WHERE id = ?", workspaceID)
	return err
}

func (c *CentralDBProvider) ListWorkspaces() ([]Workspace, error) {
	rows, err := c.db.Query("SELECT id, root_path, config, time_stamp FROM workspaces ORDER BY time_stamp ASC;")
	if err != nil {
		return nil, fmt.Errorf("failed to query workspaces: %v", err)
	}
	defer rows.Close()

	var workspaces []Workspace

	for rows.Next() {
		var workspace Workspace
		// Scan directly into the Workspace struct fields
		if err := rows.Scan(&workspace.ID, &workspace.RootPath, &workspace.Config, &workspace.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %v", err)
		}
		workspaces = append(workspaces, workspace)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return workspaces, nil
}

func (c *CentralDBProvider) WorkspaceExists(workspaceID uuid.UUID) (bool, error) {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = ?)", workspaceID).Scan(&exists)
	return exists, err
}

// Close closes the central database connection.
func (c *CentralDBProvider) Close() error {
	return c.db.Close()
}

// GetDirectoryTree returns the directory tree
func (c *CentralDBProvider) GetDirectoryTree() *trees.DirectoryTree {
	return c.DirectoryTree
}

// SetDirectoryTree sets the directory tree
func (c *CentralDBProvider) SetDirectoryTree(tree *trees.DirectoryTree) {
	c.DirectoryTree = tree
}

// Connect implements ICentralDBProvider.Connect
func (c *CentralDBProvider) Connect(dsn string) (*sql.DB, error) {
	var err error
	c.db, err = ConnectToDB(dsn)
	return c.db, err
}

// InitSchema implements ICentralDBProvider.InitSchema
func (c *CentralDBProvider) InitSchema() error {
	return c.init()
}

// Backup creates a backup of the central database.
// It returns the path to the backup file and any error that occurred during the process.
func (c *CentralDBProvider) Backup() (string, error) {
	if c.db == nil {
		return "", fmt.Errorf("cannot backup: database connection is nil")
	}
	
	// Create backup directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %v", err)
	}
	
	backupDir := filepath.Join(homeDir, ".config", "desktop_cleaner", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("could not create backup directory: %v", err)
	}
	
	// Generate unique backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("central_backup_%s.db", timestamp))
	
	// Execute the backup using SQL VACUUM INTO command
	// This is specific to SQLite and creates a copy of the database
	_, err = c.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath))
	if err != nil {
		return "", fmt.Errorf("backup failed: %v", err)
	}
	
	slog.Info("Database backup created successfully", "path", backupPath)
	return backupPath, nil
}

// GetAllSnapshots retrieves all snapshots from the central database
func (c *CentralDBProvider) GetAllSnapshots() ([]Snapshot, error) {
	rows, err := c.db.Query("SELECT id, taken_at, directory_state FROM snapshots ORDER BY taken_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []Snapshot
	for rows.Next() {
		var snapshot Snapshot
		var id string
		var takenAt string
		var directoryState []byte

		if err := rows.Scan(&id, &takenAt, &directoryState); err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}

		snapshot.ID, err = uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("failed to parse snapshot ID: %w", err)
		}

		snapshot.TakenAt, err = time.Parse(time.RFC3339, takenAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse snapshot timestamp: %w", err)
		}

		snapshot.DirectoryState = directoryState
		snapshots = append(snapshots, snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during snapshot iteration: %w", err)
	}

	return snapshots, nil
}

// GetLatestSnapshot retrieves the most recent snapshot from the central database
func (c *CentralDBProvider) GetLatestSnapshot() (*Snapshot, error) {
	row := c.db.QueryRow("SELECT id, taken_at, directory_state FROM snapshots ORDER BY taken_at DESC LIMIT 1")
	
	var snapshot Snapshot
	var id string
	var takenAt string
	var directoryState []byte
	
	err := row.Scan(&id, &takenAt, &directoryState)
	if err != nil {
		return nil, fmt.Errorf("failed to scan snapshot: %w", err)
	}
	
	snapshot.ID, err = uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot ID: %w", err)
	}
	
	snapshot.TakenAt, err = time.Parse(time.RFC3339, takenAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot timestamp: %w", err)
	}
	
	snapshot.DirectoryState = directoryState
	return &snapshot, nil
}

// GetSnapshot retrieves a specific snapshot by ID from the central database
func (c *CentralDBProvider) GetSnapshot(id uuid.UUID) (*Snapshot, error) {
	row := c.db.QueryRow("SELECT id, taken_at, directory_state FROM snapshots WHERE id = $1", id.String())
	
	var snapshot Snapshot
	var idStr string
	var takenAt string
	var directoryState []byte
	
	err := row.Scan(&idStr, &takenAt, &directoryState)
	if err != nil {
		return nil, fmt.Errorf("failed to scan snapshot: %w", err)
	}
	
	snapshot.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot ID: %w", err)
	}
	
	snapshot.TakenAt, err = time.Parse(time.RFC3339, takenAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot timestamp: %w", err)
	}
	
	snapshot.DirectoryState = directoryState
	return &snapshot, nil
}

// InsertSnapshot adds a new snapshot to the central database
func (c *CentralDBProvider) InsertSnapshot(snapshot *Snapshot) (uuid.UUID, error) {
	if snapshot.ID == uuid.Nil {
		snapshot.ID = uuid.New()
	}
	
	if snapshot.TakenAt.IsZero() {
		snapshot.TakenAt = time.Now()
	}
	
	_, err := c.db.Exec(
		"INSERT INTO snapshots (id, taken_at, directory_state) VALUES ($1, $2, $3)",
		snapshot.ID.String(),
		snapshot.TakenAt.Format(time.RFC3339),
		snapshot.DirectoryState,
	)
	
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert snapshot: %w", err)
	}
	
	return snapshot.ID, nil
}
