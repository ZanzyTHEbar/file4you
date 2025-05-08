package db

import (
	"database/sql"
	"file4you/internal/filesystem/trees" // Added this import
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid" // Added for snapshot tests
)

func TestMockWorkspaceDBProvider_InsertFileMetadata(t *testing.T) {
	tests := []struct {
		name         string
		initialMeta  map[string]*trees.FileMetadata
		metaToInsert *trees.FileMetadata
		wantErr      bool
		wantMeta     map[string]*trees.FileMetadata
	}{
		{
			name:         "insert new metadata",
			initialMeta:  map[string]*trees.FileMetadata{},
			metaToInsert: &trees.FileMetadata{FilePath: "/test/file1.txt", Size: 100, ModTime: time.Now(), IsDir: false},
			wantErr:      false,
			wantMeta: map[string]*trees.FileMetadata{
				"/test/file1.txt": {FilePath: "/test/file1.txt", Size: 100, ModTime: time.Now(), IsDir: false},
			},
		},
		{
			name: "insert duplicate metadata",
			initialMeta: map[string]*trees.FileMetadata{
				"/test/file1.txt": {FilePath: "/test/file1.txt", Size: 100, ModTime: time.Now(), IsDir: false},
			},
			metaToInsert: &trees.FileMetadata{FilePath: "/test/file1.txt", Size: 200, ModTime: time.Now(), IsDir: false},
			wantErr:      true,
			wantMeta: map[string]*trees.FileMetadata{
				"/test/file1.txt": {FilePath: "/test/file1.txt", Size: 100, ModTime: time.Now(), IsDir: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockWorkspaceDBProvider()
			// Create deep copies of the test metadata to avoid pointer comparison issues
			for k, v := range tt.initialMeta {
				clone := *v // Make a copy of the struct
				mockDB.metadata[k] = &clone
			}

			err := mockDB.InsertFileMetadata(tt.metaToInsert)
			if (err != nil) != tt.wantErr {
				t.Errorf("MockWorkspaceDBProvider.InsertFileMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare field values rather than pointers
			// First check if maps have the same keys
			if len(mockDB.metadata) != len(tt.wantMeta) {
				t.Errorf("MockWorkspaceDBProvider.InsertFileMetadata() metadata has %d entries, want %d",
					len(mockDB.metadata), len(tt.wantMeta))
				return
			}

			// Then check if values match by their content rather than pointer equality
			for k, wantV := range tt.wantMeta {
				gotV, exists := mockDB.metadata[k]
				if !exists {
					t.Errorf("MockWorkspaceDBProvider.InsertFileMetadata() missing key %s in result", k)
					continue
				}

				// Compare fields that matter (ignore ModTime as it can vary)
				if gotV.FilePath != wantV.FilePath || gotV.Size != wantV.Size || gotV.IsDir != wantV.IsDir {
					t.Errorf("MockWorkspaceDBProvider.InsertFileMetadata() metadata[%s] = %+v, want %+v",
						k, gotV, wantV)
				}
			}
		})
	}
}

func TestMockWorkspaceDBProvider_GetFileMetadata(t *testing.T) {
	now := time.Now()
	initialData := map[string]*trees.FileMetadata{
		"/test/file1.txt": {FilePath: "/test/file1.txt", Size: 100, ModTime: now, IsDir: false, Checksum: "abc"},
		"/test/dir1":      {FilePath: "/test/dir1", Size: 0, ModTime: now, IsDir: true, Checksum: ""},
	}

	tests := []struct {
		name     string
		filePath string
		wantMeta *trees.FileMetadata
		wantErr  error // Using error type for more specific checks like sql.ErrNoRows
	}{
		{
			name:     "get existing file metadata",
			filePath: "/test/file1.txt",
			wantMeta: &trees.FileMetadata{FilePath: "/test/file1.txt", Size: 100, ModTime: now, IsDir: false, Checksum: "abc"},
			wantErr:  nil,
		},
		{
			name:     "get non-existing metadata",
			filePath: "/test/nonexistent.txt",
			wantMeta: nil,
			wantErr:  sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockWorkspaceDBProvider()
			mockDB.metadata = initialData // Pre-populate

			gotMeta, err := mockDB.GetFileMetadata(tt.filePath)

			if err != tt.wantErr {
				t.Errorf("MockWorkspaceDBProvider.GetFileMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMeta, tt.wantMeta) {
				t.Errorf("MockWorkspaceDBProvider.GetFileMetadata() gotMeta = %v, want %v", gotMeta, tt.wantMeta)
			}
		})
	}
}

func TestMockWorkspaceDBProvider_GetAllFileMetadata(t *testing.T) {
	now := time.Now()
	initialData := map[string]*trees.FileMetadata{
		"/test/file1.txt": {FilePath: "/test/file1.txt", Size: 100, ModTime: now, IsDir: false},
		"/test/dir1":      {FilePath: "/test/dir1", Size: 0, ModTime: now, IsDir: true},
	}

	tests := []struct {
		name        string
		initialMeta map[string]*trees.FileMetadata
		wantMetas   []*trees.FileMetadata // Order might not be guaranteed by map iteration
		wantErr     bool
	}{
		{
			name:        "get all metadata from populated db",
			initialMeta: initialData,
			wantMetas: []*trees.FileMetadata{
				{FilePath: "/test/file1.txt", Size: 100, ModTime: now, IsDir: false},
				{FilePath: "/test/dir1", Size: 0, ModTime: now, IsDir: true},
			}, // Note: Order depends on map iteration. Test should account for this.
			wantErr: false,
		},
		{
			name:        "get all metadata from empty db",
			initialMeta: map[string]*trees.FileMetadata{},
			wantMetas:   []*trees.FileMetadata{},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockWorkspaceDBProvider()
			mockDB.metadata = tt.initialMeta

			gotMetas, err := mockDB.GetAllFileMetadata()
			if (err != nil) != tt.wantErr {
				t.Errorf("MockWorkspaceDBProvider.GetAllFileMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare slices, accounting for order insensitivity if necessary
			if len(gotMetas) != len(tt.wantMetas) {
				t.Errorf("MockWorkspaceDBProvider.GetAllFileMetadata() len(gotMetas) = %d, want %d", len(gotMetas), len(tt.wantMetas))
			}
			// For a more robust check, convert to maps or sort before DeepEqual if order is not guaranteed.
			// For simplicity, this example assumes order or small enough N to manually verify.
			// A better way for slices where order doesn't matter:
			foundMap := make(map[string]*trees.FileMetadata)
			for _, m := range gotMetas {
				foundMap[m.FilePath] = m
			}
			wantMap := make(map[string]*trees.FileMetadata)
			for _, m := range tt.wantMetas {
				wantMap[m.FilePath] = m
			}
			if !reflect.DeepEqual(foundMap, wantMap) {
				t.Errorf("MockWorkspaceDBProvider.GetAllFileMetadata() gotMetas = %v, want %v (compared as maps)", foundMap, wantMap)
			}
		})
	}
}

func TestMockWorkspaceDBProvider_DeleteFileMetadata(t *testing.T) {
	now := time.Now()
	initialData := map[string]*trees.FileMetadata{
		"/test/file1.txt": {FilePath: "/test/file1.txt", Size: 100, ModTime: now},
		"/test/file2.txt": {FilePath: "/test/file2.txt", Size: 200, ModTime: now},
	}

	tests := []struct {
		name        string
		filePath    string
		wantErr     bool
		wantMetaMap map[string]*trees.FileMetadata // Expected state of the internal map after deletion
	}{
		{
			name:     "delete existing metadata",
			filePath: "/test/file1.txt",
			wantErr:  false,
			wantMetaMap: map[string]*trees.FileMetadata{
				"/test/file2.txt": {FilePath: "/test/file2.txt", Size: 200, ModTime: now},
			},
		},
		{
			name:        "delete non-existing metadata",
			filePath:    "/test/nonexistent.txt",
			wantErr:     true,
			wantMetaMap: initialData, // Map should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockWorkspaceDBProvider()
			// Deep copy initialData to avoid modification across test cases
			currentMeta := make(map[string]*trees.FileMetadata)
			for k, v := range initialData {
				clone := *v // shallow clone is enough as FileMetadata fields are simple
				currentMeta[k] = &clone
			}
			mockDB.metadata = currentMeta

			err := mockDB.DeleteFileMetadata(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("MockWorkspaceDBProvider.DeleteFileMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(mockDB.metadata, tt.wantMetaMap) {
				t.Errorf("MockWorkspaceDBProvider.DeleteFileMetadata() metadata map = %v, want %v", mockDB.metadata, tt.wantMetaMap)
			}
		})
	}
}

func TestMockWorkspaceDBProvider_UpdateFileMetadata(t *testing.T) {
	now := time.Now()
	later := now.Add(1 * time.Hour)

	initialData := map[string]*trees.FileMetadata{
		"/test/file1.txt": {FilePath: "/test/file1.txt", Size: 100, ModTime: now, Checksum: "abc"},
	}

	tests := []struct {
		name         string
		metaToUpdate *trees.FileMetadata
		wantErr      bool
		wantMeta     *trees.FileMetadata // Expected state of the specific metadata item after update
	}{
		{
			name:         "update existing metadata",
			metaToUpdate: &trees.FileMetadata{FilePath: "/test/file1.txt", Size: 150, ModTime: later, Checksum: "def"},
			wantErr:      false,
			wantMeta:     &trees.FileMetadata{FilePath: "/test/file1.txt", Size: 150, ModTime: later, Checksum: "def"},
		},
		{
			name:         "update non-existing metadata",
			metaToUpdate: &trees.FileMetadata{FilePath: "/test/nonexistent.txt", Size: 200, ModTime: later},
			wantErr:      true,
			wantMeta:     nil, // No specific metadata item to check, as it shouldn't be added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockWorkspaceDBProvider()
			// Deep copy initialData
			currentMeta := make(map[string]*trees.FileMetadata)
			for k, v := range initialData {
				clone := *v
				currentMeta[k] = &clone
			}
			mockDB.metadata = currentMeta

			err := mockDB.UpdateFileMetadata(tt.metaToUpdate)
			if (err != nil) != tt.wantErr {
				t.Errorf("MockWorkspaceDBProvider.UpdateFileMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				gotUpdatedMeta, _ := mockDB.GetFileMetadata(tt.metaToUpdate.FilePath)
				if !reflect.DeepEqual(gotUpdatedMeta, tt.wantMeta) {
					t.Errorf("MockWorkspaceDBProvider.UpdateFileMetadata() metadata = %v, want %v", gotUpdatedMeta, tt.wantMeta)
				}
			} else {
				// Ensure non-existing item was not added
				if tt.metaToUpdate.FilePath == "/test/nonexistent.txt" {
					if _, exists := mockDB.metadata[tt.metaToUpdate.FilePath]; exists {
						t.Errorf("MockWorkspaceDBProvider.UpdateFileMetadata() an item was added for a non-existing filepath on error case")
					}
				}
			}
		})
	}
}

func TestMockCentralDBProvider_InsertSnapshot(t *testing.T) {
	newUUID1 := uuid.New()
	newUUID2 := uuid.New()

	tests := []struct {
		name         string
		initialSnaps map[uuid.UUID]*Snapshot
		snapToInsert *Snapshot
		wantID       uuid.UUID // Changed to uuid.UUID
		wantErr      bool
		wantSnaps    map[uuid.UUID]*Snapshot // Key changed to uuid.UUID
	}{
		{
			name:         "insert first snapshot",
			initialSnaps: make(map[uuid.UUID]*Snapshot),
			snapToInsert: &Snapshot{ID: newUUID1, TakenAt: time.Now(), DirectoryState: []byte("state1")},
			wantID:       newUUID1,
			wantErr:      false,
			wantSnaps: map[uuid.UUID]*Snapshot{
				newUUID1: {ID: newUUID1, TakenAt: time.Now(), DirectoryState: []byte("state1")},
			},
		},
		{
			name: "insert another snapshot",
			initialSnaps: map[uuid.UUID]*Snapshot{
				newUUID1: {ID: newUUID1, TakenAt: time.Now().Add(-time.Hour), DirectoryState: []byte("state_prev")},
			},
			snapToInsert: &Snapshot{ID: newUUID2, TakenAt: time.Now(), DirectoryState: []byte("state_new")},
			wantID:       newUUID2,
			wantErr:      false,
			wantSnaps: map[uuid.UUID]*Snapshot{
				newUUID1: {ID: newUUID1, TakenAt: time.Now().Add(-time.Hour), DirectoryState: []byte("state_prev")},
				newUUID2: {ID: newUUID2, TakenAt: time.Now(), DirectoryState: []byte("state_new")},
			},
		},
		{
			name: "insert snapshot with existing ID",
			initialSnaps: map[uuid.UUID]*Snapshot{
				newUUID1: {ID: newUUID1, TakenAt: time.Now(), DirectoryState: []byte("state1")},
			},
			snapToInsert: &Snapshot{ID: newUUID1, TakenAt: time.Now().Add(time.Minute), DirectoryState: []byte("state1_updated")},
			wantID:       uuid.Nil, // Expecting error, so ID might be nil or unchanged
			wantErr:      true,
			wantSnaps: map[uuid.UUID]*Snapshot{
				newUUID1: {ID: newUUID1, TakenAt: time.Now(), DirectoryState: []byte("state1")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockCentralDBProvider()
			// Deep copy initialSnaps to avoid modification across test cases
			mockDB.snapshots = make(map[uuid.UUID]*Snapshot)
			for id, snap := range tt.initialSnaps {
				clone := *snap // Create a shallow copy
				mockDB.snapshots[id] = &clone
			}

			// For wantSnaps, ensure TakenAt is aligned for comparison if it's time.Now()
			// This is a common issue in tests involving time.Now().
			// We will align the `TakenAt` field of the `wantSnaps` with the `snapToInsert` before comparison.
			if tt.wantSnaps[tt.snapToInsert.ID] != nil {
				tt.wantSnaps[tt.snapToInsert.ID].TakenAt = tt.snapToInsert.TakenAt
			}
			// Also align for pre-existing ones if their time is based on time.Now() in the test definition
			if tt.initialSnaps[newUUID1] != nil && tt.wantSnaps[newUUID1] != nil && tt.name == "insert another snapshot" {
				tt.wantSnaps[newUUID1].TakenAt = tt.initialSnaps[newUUID1].TakenAt
			}
			if tt.initialSnaps[newUUID1] != nil && tt.wantSnaps[newUUID1] != nil && tt.name == "insert snapshot with existing ID" {
				tt.wantSnaps[newUUID1].TakenAt = tt.initialSnaps[newUUID1].TakenAt
			}

			gotID, err := mockDB.InsertSnapshot(tt.snapToInsert)

			if (err != nil) != tt.wantErr {
				t.Errorf("MockCentralDBProvider.InsertSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("MockCentralDBProvider.InsertSnapshot() gotID = %v, want %v", gotID, tt.wantID)
			}

			// If no error was expected, verify the content of the snapshots map
			if !tt.wantErr {
				// Align TakenAt for the main inserted item for comparison
				if actualInserted, ok := mockDB.snapshots[gotID]; ok {
					if expected, okE := tt.wantSnaps[gotID]; okE {
						expected.TakenAt = actualInserted.TakenAt
					}
				}
			}

			if !reflect.DeepEqual(mockDB.snapshots, tt.wantSnaps) {
				// For better debugging of DeepEqual failures with complex structs:
				for kWant, vWant := range tt.wantSnaps {
					vGot, ok := mockDB.snapshots[kWant]
					if !ok {
						t.Errorf("Missing key in mockDB.snapshots: %s", kWant)
						continue
					}
					if !reflect.DeepEqual(vGot, vWant) {
						t.Errorf("Mismatch for key %s:\nGot: %+v\nWant: %+v", kWant, vGot, vWant)
					}
				}
				for kGot := range mockDB.snapshots {
					if _, ok := tt.wantSnaps[kGot]; !ok {
						t.Errorf("Extra key in mockDB.snapshots: %s", kGot)
					}
				}
				t.Errorf("MockCentralDBProvider.InsertSnapshot() snapshots map not as expected.") // General error if specific not caught
			}
		})
	}
}

func TestMockCentralDBProvider_GetSnapshot(t *testing.T) {
	uuid1 := uuid.New()
	uuid2 := uuid.New()
	ts1 := time.Now().Add(-time.Hour).Truncate(time.Second)
	ts2 := time.Now().Truncate(time.Second)

	initialSnaps := map[uuid.UUID]*Snapshot{
		uuid1: {ID: uuid1, TakenAt: ts1, DirectoryState: []byte("state1")},
		uuid2: {ID: uuid2, TakenAt: ts2, DirectoryState: []byte("state2")},
	}

	tests := []struct {
		name     string
		idToGet  uuid.UUID // Changed to uuid.UUID
		wantSnap *Snapshot
		wantErr  error
	}{
		{
			name:     "get existing snapshot",
			idToGet:  uuid1,
			wantSnap: &Snapshot{ID: uuid1, TakenAt: ts1, DirectoryState: []byte("state1")},
			wantErr:  nil,
		},
		{
			name:     "get another existing snapshot",
			idToGet:  uuid2,
			wantSnap: &Snapshot{ID: uuid2, TakenAt: ts2, DirectoryState: []byte("state2")},
			wantErr:  nil,
		},
		{
			name:     "get non-existing snapshot",
			idToGet:  uuid.New(), // A new, different UUID
			wantSnap: nil,
			wantErr:  sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockCentralDBProvider()
			mockDB.snapshots = initialSnaps // Pre-populate

			gotSnap, err := mockDB.GetSnapshot(tt.idToGet)
			if err != tt.wantErr {
				t.Errorf("MockCentralDBProvider.GetSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSnap, tt.wantSnap) {
				t.Errorf("MockCentralDBProvider.GetSnapshot() gotSnap = %+v, want %+v", gotSnap, tt.wantSnap)
			}
		})
	}
}

func TestMockCentralDBProvider_GetLatestSnapshot(t *testing.T) {
	uuid1, uuid2, uuid3 := uuid.New(), uuid.New(), uuid.New()
	ts1 := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
	ts2 := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
	ts3 := time.Now().Truncate(time.Second) // Latest

	snap1 := &Snapshot{ID: uuid1, TakenAt: ts1, DirectoryState: []byte("state_old")}
	snap2 := &Snapshot{ID: uuid2, TakenAt: ts2, DirectoryState: []byte("state_mid")}
	snap3 := &Snapshot{ID: uuid3, TakenAt: ts3, DirectoryState: []byte("state_new")}

	tests := []struct {
		name         string
		initialSnaps map[uuid.UUID]*Snapshot // Key changed to uuid.UUID
		wantSnap     *Snapshot
		wantErr      error
	}{
		{
			name: "get latest from multiple snapshots",
			initialSnaps: map[uuid.UUID]*Snapshot{
				uuid1: snap1, // Older
				uuid2: snap2, // Middle
				uuid3: snap3, // Newest
			},
			wantSnap: snap3,
			wantErr:  nil,
		},
		{
			name: "get latest when only one snapshot",
			initialSnaps: map[uuid.UUID]*Snapshot{
				uuid1: snap1,
			},
			wantSnap: snap1,
			wantErr:  nil,
		},
		{
			name:         "get latest from empty snapshots",
			initialSnaps: make(map[uuid.UUID]*Snapshot),
			wantSnap:     nil,
			wantErr:      sql.ErrNoRows,
		},
		{
			name: "get latest with mixed ID and time order",
			initialSnaps: map[uuid.UUID]*Snapshot{
				uuid3: snap1, // ID uuid3, but oldest time
				uuid1: snap2, // ID uuid1, but middle time
				uuid2: snap3, // ID uuid2, but newest time
			},
			wantSnap: snap3, // snap3 is the latest by time (TakenAt)
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockCentralDBProvider()
			mockDB.snapshots = tt.initialSnaps

			gotSnap, err := mockDB.GetLatestSnapshot()
			if err != tt.wantErr {
				t.Errorf("MockCentralDBProvider.GetLatestSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSnap, tt.wantSnap) {
				t.Errorf("MockCentralDBProvider.GetLatestSnapshot() gotSnap = %+v, want %+v", gotSnap, tt.wantSnap)
			}
		})
	}
}

func TestMockCentralDBProvider_GetAllSnapshots(t *testing.T) {
	uuid1, uuid2 := uuid.New(), uuid.New()
	ts1 := time.Now().Add(-time.Hour).Truncate(time.Second)
	ts2 := time.Now().Truncate(time.Second)

	snap1 := &Snapshot{ID: uuid1, TakenAt: ts1, DirectoryState: []byte("state1")}
	snap2 := &Snapshot{ID: uuid2, TakenAt: ts2, DirectoryState: []byte("state2")}

	tests := []struct {
		name         string
		initialSnaps map[uuid.UUID]*Snapshot // Key changed to uuid.UUID
		wantSnaps    []Snapshot              // Changed to []Snapshot to match mock return type
		wantErr      bool
	}{
		{
			name: "get all from populated snapshots",
			initialSnaps: map[uuid.UUID]*Snapshot{
				uuid1: snap1,
				uuid2: snap2,
			},
			// Order in slice might vary from map iteration, so compare content carefully
			wantSnaps: []Snapshot{*snap1, *snap2},
			wantErr:   false,
		},
		{
			name:         "get all from empty snapshots",
			initialSnaps: make(map[uuid.UUID]*Snapshot),
			wantSnaps:    []Snapshot{},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockCentralDBProvider()
			mockDB.snapshots = tt.initialSnaps

			gotSnaps, err := mockDB.GetAllSnapshots()
			if (err != nil) != tt.wantErr {
				t.Errorf("MockCentralDBProvider.GetAllSnapshots() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(gotSnaps) != len(tt.wantSnaps) {
				t.Errorf("MockCentralDBProvider.GetAllSnapshots() len = %d, want %d", len(gotSnaps), len(tt.wantSnaps))
				return // Avoid panic on reflect.DeepEqual if lengths differ
			}

			// For robust comparison of slices where order doesn't matter and items are structs:
			// Convert both to maps keyed by a unique identifier (e.g., ID).
			gotSnapsMap := make(map[uuid.UUID]Snapshot)
			for _, s := range gotSnaps {
				gotSnapsMap[s.ID] = s
			}
			wantSnapsMap := make(map[uuid.UUID]Snapshot)
			for _, s := range tt.wantSnaps {
				wantSnapsMap[s.ID] = s
			}

			if !reflect.DeepEqual(gotSnapsMap, wantSnapsMap) {
				t.Errorf("MockCentralDBProvider.GetAllSnapshots() got = %v, want %v (compared as maps)", gotSnapsMap, wantSnapsMap)
			}
		})
	}
}
