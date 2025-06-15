package db

import (
	"database/sql"
	"file4you/internal/trees" // Added this import
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockWorkspaceDBProvider is an in-memory mock for WorkspaceDBProvider
type MockWorkspaceDBProvider struct {
	mu       sync.Mutex
	metadata map[string]*trees.FileMetadata
}

func NewMockWorkspaceDBProvider() *MockWorkspaceDBProvider {
	return &MockWorkspaceDBProvider{
		metadata: make(map[string]*trees.FileMetadata),
	}
}

func (m *MockWorkspaceDBProvider) Connect(dsn string) (*sql.DB, error) {
	return nil, nil // Not a real DB connection
}

func (m *MockWorkspaceDBProvider) Close() error {
	return nil
}

func (m *MockWorkspaceDBProvider) InitSchema() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata = make(map[string]*trees.FileMetadata) // Clear data on init
	return nil
}

func (m *MockWorkspaceDBProvider) InsertFileMetadata(meta *trees.FileMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.metadata[meta.FilePath]; exists {
		return fmt.Errorf("metadata for %s already exists", meta.FilePath)
	}
	m.metadata[meta.FilePath] = meta
	return nil
}

func (m *MockWorkspaceDBProvider) GetFileMetadata(filePath string) (*trees.FileMetadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, exists := m.metadata[filePath]
	if !exists {
		return nil, sql.ErrNoRows // Simulate database behavior for not found
	}
	return meta, nil
}

func (m *MockWorkspaceDBProvider) GetAllFileMetadata() ([]*trees.FileMetadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var allMeta []*trees.FileMetadata
	for _, meta := range m.metadata {
		allMeta = append(allMeta, meta)
	}
	return allMeta, nil
}

func (m *MockWorkspaceDBProvider) DeleteFileMetadata(filePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.metadata[filePath]; !exists {
		return fmt.Errorf("metadata for %s not found", filePath)
	}
	delete(m.metadata, filePath)
	return nil
}

func (m *MockWorkspaceDBProvider) UpdateFileMetadata(meta *trees.FileMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.metadata[meta.FilePath]; !exists {
		// For an update, the item should typically exist.
		// Depending on desired semantics (upsert vs strict update), this can change.
		// sql.ErrNoRows might be more idiomatic if it's a strict update.
		return fmt.Errorf("metadata for %s not found, cannot update", meta.FilePath)
	}
	m.metadata[meta.FilePath] = meta
	return nil
}

// MockCentralDBProvider is an in-memory mock for CentralDBProvider
type MockCentralDBProvider struct {
	mu            sync.Mutex
	snapshots     map[uuid.UUID]*Snapshot
	workspaces    map[uuid.UUID]*Workspace
	DirectoryTree *trees.DirectoryTree // Added for compatibility with tests
}

func NewMockCentralDBProvider() *MockCentralDBProvider {
	return &MockCentralDBProvider{
		snapshots:     make(map[uuid.UUID]*Snapshot),
		workspaces:    make(map[uuid.UUID]*Workspace),
		DirectoryTree: nil, // Will be initialized during tests that need it
	}
}

func (m *MockCentralDBProvider) Connect(dsn string) (*sql.DB, error) {
	return nil, nil // Not a real DB connection
}

func (m *MockCentralDBProvider) Close() error {
	return nil
}

func (m *MockCentralDBProvider) InitSchema() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots = make(map[uuid.UUID]*Snapshot) // Re-initialize with correct type
	m.workspaces = make(map[uuid.UUID]*Workspace)
	return nil
}

func (m *MockCentralDBProvider) InsertSnapshot(snapshot *Snapshot) (uuid.UUID, error) { // Return uuid.UUID
	m.mu.Lock()
	defer m.mu.Unlock()
	if snapshot.ID == uuid.Nil {
		snapshot.ID = uuid.New() // Generate new UUID if not provided
	}
	if _, exists := m.snapshots[snapshot.ID]; exists {
		return uuid.Nil, fmt.Errorf("snapshot with ID %s already exists", snapshot.ID)
	}
	m.snapshots[snapshot.ID] = snapshot
	return snapshot.ID, nil
}

func (m *MockCentralDBProvider) GetSnapshot(id uuid.UUID) (*Snapshot, error) { // Parameter is uuid.UUID
	m.mu.Lock()
	defer m.mu.Unlock()
	snap, exists := m.snapshots[id]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return snap, nil
}

func (m *MockCentralDBProvider) GetLatestSnapshot() (*Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.snapshots) == 0 {
		return nil, sql.ErrNoRows
	}
	var latestSnap *Snapshot
	var latestTime time.Time
	first := true
	for _, snap := range m.snapshots {
		if first || snap.TakenAt.After(latestTime) { // Use TakenAt from the correct Snapshot struct
			latestSnap = snap
			latestTime = snap.TakenAt // Use TakenAt
			first = false
		}
	}
	return latestSnap, nil
}

func (m *MockCentralDBProvider) GetAllSnapshots() ([]Snapshot, error) { // Return type is []Snapshot
	m.mu.Lock()
	defer m.mu.Unlock()
	var allSnaps []Snapshot // Changed to []Snapshot
	for _, snap := range m.snapshots {
		allSnaps = append(allSnaps, *snap) // Dereference snap before appending
	}
	return allSnaps, nil
}

// Workspace-related methods
func (m *MockCentralDBProvider) AddWorkspace(rootPath, config string) (*Workspace, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if a workspace with this path already exists
	for _, ws := range m.workspaces {
		if ws.RootPath == rootPath {
			return nil, fmt.Errorf("workspace for path %s already exists", rootPath)
		}
	}

	workspace := &Workspace{
		ID:        uuid.New(),
		RootPath:  rootPath,
		Config:    config,
		Timestamp: time.Now(),
	}

	m.workspaces[workspace.ID] = workspace
	return workspace, nil
}

func (m *MockCentralDBProvider) UpdateWorkspaceConfig(workspaceID uuid.UUID, config string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	workspace, exists := m.workspaces[workspaceID]
	if !exists {
		return false, fmt.Errorf("workspace with ID %s not found", workspaceID)
	}

	workspace.Config = config
	workspace.Timestamp = time.Now()
	return true, nil
}

func (m *MockCentralDBProvider) GetWorkspace(id uuid.UUID) (*Workspace, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	workspace, exists := m.workspaces[id]
	if !exists {
		return nil, fmt.Errorf("workspace with ID %s not found", id)
	}
	return workspace, nil
}

func (m *MockCentralDBProvider) GetWorkspacePath(workspaceID uuid.UUID) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	workspace, exists := m.workspaces[workspaceID]
	if !exists {
		return "", fmt.Errorf("workspace with ID %s not found", workspaceID)
	}
	return workspace.RootPath, nil
}

func (m *MockCentralDBProvider) GetWorkspaceID(rootPath string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, ws := range m.workspaces {
		if ws.RootPath == rootPath {
			// TODO: This is a bit awkward - the real implementation might return an int ID,
			// but we have UUID. For mock purposes, return 1 to indicate success.
			return 1, nil
		}
	}
	return 0, fmt.Errorf("no workspace found for path %s", rootPath)
}

func (m *MockCentralDBProvider) GetWorkspaceConfig(workspaceID uuid.UUID) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	workspace, exists := m.workspaces[workspaceID]
	if !exists {
		return "", fmt.Errorf("workspace with ID %s not found", workspaceID)
	}
	return workspace.Config, nil
}

func (m *MockCentralDBProvider) SetWorkspaceConfig(workspaceID uuid.UUID, config string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	workspace, exists := m.workspaces[workspaceID]
	if !exists {
		return fmt.Errorf("workspace with ID %s not found", workspaceID)
	}
	workspace.Config = config
	return nil
}

func (m *MockCentralDBProvider) DeleteWorkspace(workspaceID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.workspaces[workspaceID]; !exists {
		return fmt.Errorf("workspace with ID %s not found", workspaceID)
	}
	delete(m.workspaces, workspaceID)
	return nil
}

func (m *MockCentralDBProvider) ListWorkspaces() ([]Workspace, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	workspaces := make([]Workspace, 0, len(m.workspaces))
	for _, ws := range m.workspaces {
		workspaces = append(workspaces, *ws)
	}
	return workspaces, nil
}

func (m *MockCentralDBProvider) WorkspaceExists(workspaceID uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.workspaces[workspaceID]
	return exists, nil
}

func (m *MockCentralDBProvider) GetDirectoryTree() *trees.DirectoryTree {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.DirectoryTree
}

func (m *MockCentralDBProvider) SetDirectoryTree(tree *trees.DirectoryTree) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DirectoryTree = tree
}

// Backup implements ICentralDBProvider.Backup
func (m *MockCentralDBProvider) Backup() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Just return a mock path for testing purposes
	return "/mock/path/to/backup/central.db", nil
}

// Compile-time check to ensure mocks implement the interfaces.
var _ WorkspaceDBProvider = (*MockWorkspaceDBProvider)(nil)
var _ ICentralDBProvider = (*MockCentralDBProvider)(nil)
