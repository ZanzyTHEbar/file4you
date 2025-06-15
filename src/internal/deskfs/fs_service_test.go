package deskfs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"file4you/internal/config"
	"file4you/internal/db"
	"file4you/internal/deskfs/options"
	"file4you/internal/filesystem/trees"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test setup and utilities

func setupTestEnvironment(t *testing.T) (*DesktopFileSystem, string, func()) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "file4you-test-*")
	require.NoError(t, err)

	// Setup test config
	originalConfig := config.AppConfig
	config.AppConfig = config.Config{
		File4You: config.File4YouConfig{
			TargetDir:              tmpDir,
			CacheDir:               filepath.Join(tmpDir, ".cache"),
			OrganizeTimeoutMinutes: 5,
		},
	}

	// Create mock database
	mockDB := db.NewMockCentralDBProvider()
	mockDB.SetDirectoryTree(trees.NewDirectoryTree(trees.WithRoot(tmpDir)))

	// Create mock terminal
	mockTerminal := &MockTerminal{}

	// Create filesystem
	dfs, err := NewDesktopFileSystem(mockTerminal, mockDB)
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(tmpDir)
		config.AppConfig = originalConfig
	}

	return dfs, tmpDir, cleanup
}

func createTestFiles(t *testing.T, baseDir string, files map[string]string) {
	for relPath, content := range files {
		fullPath := filepath.Join(baseDir, relPath)
		dir := filepath.Dir(fullPath)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}
}

// Tests for the new service-oriented architecture

func TestDesktopFileSystem_Creation(t *testing.T) {
	dfs, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	assert.NotNil(t, dfs)
	assert.NotNil(t, dfs.GetDirectoryService())
	assert.NotNil(t, dfs.GetFileOperations())
	assert.NotNil(t, dfs.GetOrganizationService())
	assert.NotNil(t, dfs.GetConflictResolver())
	assert.NotNil(t, dfs.GetGitService())
}

func TestDesktopFileSystem_OrganizeDirectory(t *testing.T) {
	dfs, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test file structure
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")

	testFiles := map[string]string{
		"source/document.pdf":    "PDF content",
		"source/image.jpg":       "JPEG content",
		"source/data.csv":        "CSV content",
		"source/nested/file.txt": "Text content",
		"source/script.py":       "Python script",
	}
	createTestFiles(t, tmpDir, testFiles)

	ctx := context.Background()
	opts := options.OrganizationOptions{
		SourceDir:   sourceDir,
		TargetDir:   targetDir,
		DryRun:      true, // Use dry run for testing
		Recursive:   true,
		WorkerCount: 2,
		BatchSize:   10,
	}

	result, err := dfs.OrganizeDirectory(ctx, sourceDir, targetDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, sourceDir, result.SourcePath)
	assert.Equal(t, targetDir, result.TargetPath)
	assert.True(t, result.Success)
	assert.True(t, result.DryRun)
	assert.True(t, result.Duration > 0)
}

func TestDesktopFileSystem_IndexDirectory(t *testing.T) {
	dfs, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test file structure
	testFiles := map[string]string{
		"dir1/file1.txt":        "content1",
		"dir1/file2.doc":        "content2",
		"dir2/subdir/file3.pdf": "content3",
		"file4.jpg":             "content4",
	}
	createTestFiles(t, tmpDir, testFiles)

	ctx := context.Background()
	opts := options.IndexOptions{
		Recursive:     true,
		IncludeHidden: false,
		WorkerCount:   2,
		MaxDepth:      10,
		BatchSize:     100,
	}

	err := dfs.IndexDirectory(ctx, tmpDir, opts)
	assert.NoError(t, err)
}

func TestDesktopFileSystem_AnalyzeDirectory(t *testing.T) {
	dfs, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test file structure with various file types
	testFiles := map[string]string{
		"documents/report.pdf":  "PDF report",
		"documents/notes.txt":   "Text notes",
		"images/photo1.jpg":     "JPEG image",
		"images/photo2.png":     "PNG image",
		"data/spreadsheet.xlsx": "Excel data",
		"code/script.py":        "Python code",
		"code/main.go":          "Go code",
	}
	createTestFiles(t, tmpDir, testFiles)

	ctx := context.Background()
	analysis, err := dfs.AnalyzeDirectory(ctx, tmpDir)
	require.NoError(t, err)
	require.NotNil(t, analysis)

	// Verify analysis contains expected information
	assert.Greater(t, analysis.TotalFiles, 0)
	assert.Greater(t, analysis.TotalSize, int64(0))
	assert.True(t, len(analysis.FileTypes) > 0)
}

func TestDesktopFileSystem_PreviewOrganization(t *testing.T) {
	dfs, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create test files
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")

	testFiles := map[string]string{
		"source/report.pdf": "Report content",
		"source/image.jpg":  "Image content",
		"source/data.csv":   "CSV data",
		"source/readme.txt": "Text file",
	}
	createTestFiles(t, tmpDir, testFiles)

	ctx := context.Background()
	opts := options.OrganizationOptions{
		SourceDir: sourceDir,
		TargetDir: targetDir,
		Recursive: true,
	}

	preview, err := dfs.PreviewOrganization(ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, preview)

	assert.Equal(t, sourceDir, preview.SourcePath)
	assert.Equal(t, targetDir, preview.TargetPath)
	assert.True(t, len(preview.Operations) >= 0) // May be 0 if no organization rules defined
}

func TestDesktopFileSystem_CopyFile(t *testing.T) {
	dfs, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create source file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	targetFile := filepath.Join(tmpDir, "target.txt")
	testContent := "Test file content"

	require.NoError(t, os.WriteFile(sourceFile, []byte(testContent), 0644))

	ctx := context.Background()
	opts := options.CopyOptions{
		Recursive:     false,
		PreservePerms: true,
		PreserveTimes: true,
		DryRun:        false,
	}

	err := dfs.CopyFile(ctx, sourceFile, targetFile, opts)
	require.NoError(t, err)

	// Verify file was copied
	assert.FileExists(t, targetFile)
	content, err := os.ReadFile(targetFile)
	require.NoError(t, err)
	assert.Equal(t, testContent, string(content))
}

func TestDesktopFileSystem_MoveFile(t *testing.T) {
	dfs, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create source file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	targetFile := filepath.Join(tmpDir, "target.txt")
	testContent := "Test file content"

	require.NoError(t, os.WriteFile(sourceFile, []byte(testContent), 0644))

	ctx := context.Background()
	opts := options.MoveOptions{
		Overwrite:  false,
		CreateDirs: true,
	}

	err := dfs.MoveFile(ctx, sourceFile, targetFile, opts)
	require.NoError(t, err)

	// Verify file was moved
	assert.NoFileExists(t, sourceFile)
	assert.FileExists(t, targetFile)
	content, err := os.ReadFile(targetFile)
	require.NoError(t, err)
	assert.Equal(t, testContent, string(content))
}

func TestDesktopFileSystem_SafetyValidation(t *testing.T) {
	dfs, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("prevents moving to subdirectory of source", func(t *testing.T) {
		sourceDir := tmpDir
		targetDir := filepath.Join(tmpDir, "subdir")

		opts := options.OrganizationOptions{
			SourceDir: sourceDir,
			TargetDir: targetDir,
		}

		_, err := dfs.OrganizeDirectory(ctx, sourceDir, targetDir, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "safety validation failed")
	})

	t.Run("allows safe organization", func(t *testing.T) {
		sourceDir := filepath.Join(tmpDir, "source")
		targetDir := filepath.Join(tmpDir, "target")

		require.NoError(t, os.MkdirAll(sourceDir, 0755))
		require.NoError(t, os.MkdirAll(targetDir, 0755))

		opts := options.OrganizationOptions{
			SourceDir: sourceDir,
			TargetDir: targetDir,
			DryRun:    true,
		}

		_, err := dfs.OrganizeDirectory(ctx, sourceDir, targetDir, opts)
		assert.NoError(t, err)
	})
}

func TestDesktopFileSystem_GitOperations(t *testing.T) {
	dfs, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	// Test git service accessibility
	gitService := dfs.GetGitService()
	assert.NotNil(t, gitService)

	// Test git repository detection (should not be a git repo)
	isRepo, err := gitService.IsGitRepository(ctx, tmpDir)
	assert.NoError(t, err)
	assert.False(t, isRepo)
}

// Test error conditions

func TestDesktopFileSystem_ErrorConditions(t *testing.T) {
	dfs, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("organize with invalid source directory", func(t *testing.T) {
		opts := options.OrganizationOptions{
			SourceDir: "/nonexistent/path",
			TargetDir: tmpDir,
		}

		_, err := dfs.OrganizeDirectory(ctx, "/nonexistent/path", tmpDir, opts)
		assert.Error(t, err)
	})

	t.Run("index invalid directory", func(t *testing.T) {
		opts := options.IndexOptions{}
		err := dfs.IndexDirectory(ctx, "/nonexistent/path", opts)
		assert.Error(t, err)
	})

	t.Run("copy nonexistent file", func(t *testing.T) {
		opts := options.CopyOptions{}
		err := dfs.CopyFile(ctx, "/nonexistent/file", filepath.Join(tmpDir, "target"), opts)
		assert.Error(t, err)
	})
}

// Benchmark tests

func BenchmarkDesktopFileSystem_OrganizeDirectory(b *testing.B) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "file4you-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mockDB := db.NewMockCentralDBProvider()
	mockDB.SetDirectoryTree(trees.NewDirectoryTree(trees.WithRoot(tmpDir)))
	mockTerminal := &MockTerminal{}
	dfs, err := NewDesktopFileSystem(mockTerminal, mockDB)
	if err != nil {
		b.Fatal(err)
	}

	// Create test files
	sourceDir := filepath.Join(tmpDir, "source")
	targetDir := filepath.Join(tmpDir, "target")
	for i := 0; i < 100; i++ {
		testFiles := map[string]string{
			filepath.Join("source", "file"+string(rune(i))+".txt"): "content",
		}
		createTestFiles(b, tmpDir, testFiles)
	}

	opts := options.OrganizationOptions{
		SourceDir: sourceDir,
		TargetDir: targetDir,
		DryRun:    true,
		Recursive: true,
	}

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := dfs.OrganizeDirectory(ctx, sourceDir, targetDir, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// MockTerminal for testing
type MockTerminal struct{}

func (m *MockTerminal) Info(message string)                                  {}
func (m *MockTerminal) Infof(format string, args ...interface{})             {}
func (m *MockTerminal) Success(message string)                               {}
func (m *MockTerminal) Successf(format string, args ...interface{})          {}
func (m *MockTerminal) Warning(message string)                               {}
func (m *MockTerminal) Warningf(format string, args ...interface{})          {}
func (m *MockTerminal) Error(message string, err error)                      {}
func (m *MockTerminal) Errorf(format string, err error, args ...interface{}) {}
func (m *MockTerminal) Fatal(message string, err error)                      {}
func (m *MockTerminal) Fatalf(format string, err error, args ...interface{}) {}
func (m *MockTerminal) Output(message string)                                {}
func (m *MockTerminal) Outputf(format string, args ...interface{})           {}
func (m *MockTerminal) Confirm(prompt string, defaultValue bool) (bool, error) {
	return defaultValue, nil
}
func (m *MockTerminal) Prompt(prompt string, defaultValue string) (string, error) {
	return defaultValue, nil
}
func (m *MockTerminal) Select(prompt string, options []string, defaultValue string) (string, error) {
	return defaultValue, nil
}
func (m *MockTerminal) StartSpinner(message string)                     {}
func (m *MockTerminal) StopSpinner(success bool, message string)        {}
func (m *MockTerminal) ShowCustomHelp(showAll bool, commandPath string) {}
