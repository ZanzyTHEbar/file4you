package db

import (
	"file4you/internal/trees"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCentralDBProviderIntegration tests the actual CentralDBProvider implementation
func TestCentralDBProviderIntegration(t *testing.T) {
	// Create a temporary directory for test database
	tempDir, err := os.MkdirTemp("", "file4you_test_central_db_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Override the default database path for testing
	testDBPath := filepath.Join(tempDir, "test_central.db")

	// Connect to the test database
	db, err := ConnectToDB(testDBPath)
	require.NoError(t, err)
	defer db.Close()

	provider := &CentralDBProvider{db: db}
	err = provider.init()
	require.NoError(t, err)

	t.Run("AddWorkspace", func(t *testing.T) {
		workspace, err := provider.AddWorkspace("/test/path", "test config")
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, workspace.ID)
		assert.Equal(t, "/test/path", workspace.RootPath)
		assert.Equal(t, "test config", workspace.Config)
		assert.False(t, workspace.Timestamp.IsZero())
	})

	t.Run("GetWorkspace", func(t *testing.T) {
		// First add a workspace
		workspace, err := provider.AddWorkspace("/test/path2", "test config 2")
		require.NoError(t, err)

		// Then retrieve it
		retrieved, err := provider.GetWorkspace(workspace.ID)
		require.NoError(t, err)
		assert.Equal(t, workspace.ID, retrieved.ID)
		assert.Equal(t, workspace.RootPath, retrieved.RootPath)
		assert.Equal(t, workspace.Config, retrieved.Config)
	})

	t.Run("ListWorkspaces", func(t *testing.T) {
		// Add multiple workspaces
		workspace1, err := provider.AddWorkspace("/test/list1", "config1")
		require.NoError(t, err)
		workspace2, err := provider.AddWorkspace("/test/list2", "config2")
		require.NoError(t, err)

		// List all workspaces
		workspaces, err := provider.ListWorkspaces()
		require.NoError(t, err)

		// Should have at least our 2 workspaces (plus any from previous tests)
		assert.GreaterOrEqual(t, len(workspaces), 2)

		// Find our workspaces in the list
		found1, found2 := false, false
		for _, ws := range workspaces {
			if ws.ID == workspace1.ID {
				found1 = true
			}
			if ws.ID == workspace2.ID {
				found2 = true
			}
		}
		assert.True(t, found1, "workspace1 should be in the list")
		assert.True(t, found2, "workspace2 should be in the list")
	})

	t.Run("WorkspaceExists", func(t *testing.T) {
		workspace, err := provider.AddWorkspace("/test/exists", "config")
		require.NoError(t, err)

		exists, err := provider.WorkspaceExists(workspace.ID)
		require.NoError(t, err)
		assert.True(t, exists)

		// Test with non-existent workspace
		nonExistentID := uuid.New()
		exists, err = provider.WorkspaceExists(nonExistentID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("SnapshotOperations", func(t *testing.T) {
		// Test inserting a snapshot
		snapshot := &Snapshot{
			ID:             uuid.New(),
			TakenAt:        time.Now(),
			DirectoryState: []byte("test directory state"),
		}

		snapshotID, err := provider.InsertSnapshot(snapshot)
		require.NoError(t, err)
		assert.Equal(t, snapshot.ID, snapshotID)

		// Test getting the snapshot
		retrieved, err := provider.GetSnapshot(snapshotID)
		require.NoError(t, err)
		assert.Equal(t, snapshot.ID, retrieved.ID)
		assert.Equal(t, snapshot.DirectoryState, retrieved.DirectoryState)
		// Allow some time tolerance for timestamp comparison
		assert.WithinDuration(t, snapshot.TakenAt, retrieved.TakenAt, time.Second)

		// Test getting latest snapshot
		latest, err := provider.GetLatestSnapshot()
		require.NoError(t, err)
		assert.Equal(t, snapshot.ID, latest.ID)

		// Test getting all snapshots
		all, err := provider.GetAllSnapshots()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 1)

		// Find our snapshot in the list
		found := false
		for _, s := range all {
			if s.ID == snapshot.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "our snapshot should be in the list")
	})

	t.Run("UpdateWorkspaceConfig", func(t *testing.T) {
		workspace, err := provider.AddWorkspace("/test/update", "original config")
		require.NoError(t, err)

		success, err := provider.UpdateWorkspaceConfig(workspace.ID, "updated config")
		require.NoError(t, err)
		assert.True(t, success)

		// Verify the update
		config, err := provider.GetWorkspaceConfig(workspace.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated config", config)
	})

	t.Run("DeleteWorkspace", func(t *testing.T) {
		workspace, err := provider.AddWorkspace("/test/delete", "delete me")
		require.NoError(t, err)

		// Verify it exists
		exists, err := provider.WorkspaceExists(workspace.ID)
		require.NoError(t, err)
		assert.True(t, exists)

		// Delete it
		err = provider.DeleteWorkspace(workspace.ID)
		require.NoError(t, err)

		// Verify it's gone
		exists, err = provider.WorkspaceExists(workspace.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Backup", func(t *testing.T) {
		backupPath, err := provider.Backup()
		require.NoError(t, err)
		assert.NotEmpty(t, backupPath)

		// Verify backup file exists
		_, err = os.Stat(backupPath)
		assert.NoError(t, err)
	})
}

// TestWorkspaceDBIntegration tests the WorkspaceDB implementation
func TestWorkspaceDBIntegration(t *testing.T) {
	// Create a temporary directory for test workspace
	tempDir, err := os.MkdirTemp("", "file4you_test_workspace_db_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create workspace database
	workspaceDB, err := NewWorkspaceDB(tempDir)
	require.NoError(t, err)
	defer workspaceDB.Close()

	t.Run("DatabaseInitialization", func(t *testing.T) {
		// Verify tables were created by attempting to query them
		var count int
		err := workspaceDB.db.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
		assert.NoError(t, err, "files table should exist")

		err = workspaceDB.db.QueryRow("SELECT COUNT(*) FROM history").Scan(&count)
		assert.NoError(t, err, "history table should exist")
	})

	t.Run("BatchInsertFiles", func(t *testing.T) {
		files := []trees.FileMetadata{
			{
				FilePath: "/test/file1.txt",
				Size:     100,
				ModTime:  time.Now(),
			},
			{
				FilePath: "/test/file2.txt",
				Size:     200,
				ModTime:  time.Now(),
			},
		}

		err := workspaceDB.BatchInsertFiles(files)
		assert.NoError(t, err)

		// Verify files were inserted
		rows, err := workspaceDB.db.Query("SELECT COUNT(*) FROM files")
		require.NoError(t, err)
		defer rows.Close()

		var count int
		require.True(t, rows.Next())
		err = rows.Scan(&count)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2)
	})

	t.Run("BatchInsertHistory", func(t *testing.T) {
		events := []HistoryEvent{
			{
				EventType: "file_created",
				EventJSON: `{"path": "/test/file1.txt"}`,
			},
			{
				EventType: "file_modified",
				EventJSON: `{"path": "/test/file2.txt"}`,
			},
		}

		err := workspaceDB.BatchInsertHistory(events)
		assert.NoError(t, err)

		// Verify events were inserted
		rows, err := workspaceDB.db.Query("SELECT COUNT(*) FROM history")
		require.NoError(t, err)
		defer rows.Close()

		var count int
		require.True(t, rows.Next())
		err = rows.Scan(&count)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2)
	})
}
