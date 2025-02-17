package deskfs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"file4you/internal/terminal"

	"github.com/stretchr/testify/assert"
)

// TODO: Setup mock filesystem & Database for testing
// TODO: Test workspaces feature

func loadTestConfig(configPath string) *DeskFSConfig {
	// Call NewConfig with the provided path (can be nil if no path is specified)
	config := NewIntermediateConfig(configPath)

	deskfsConfig := NewDeskFSConfig()

	// Build FileTypeTree
	deskfsConfig = deskfsConfig.BuildFileTypeTree(config)

	return deskfsConfig
}

// Helper to create a temporary directory structure for tests
func setupTestDir(t *testing.T, structure map[string]string) (string, func()) {
	dir, err := os.MkdirTemp("", "desktop_cleaner_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create directories and files as defined in structure map
	for path, content := range structure {
		fullPath := filepath.Join(dir, path)
		if filepath.Ext(path) == "" {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				t.Fatalf("failed to create directory %s: %v", fullPath, err)
			}
		} else {
			parentDir := filepath.Dir(fullPath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				t.Fatalf("failed to create parent directory %s: %v", parentDir, err)
			}
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				t.Fatalf("failed to create file %s: %v", fullPath, err)
			}
		}
	}

	return dir, func() { os.RemoveAll(dir) }
}

func createTestConfigFile(t *testing.T, content string) (string, func()) {
	tmpFile, err := os.CreateTemp("", "desktop_cleaner_config_*.toml")
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp config file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp config file: %v", err)
	}
	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
}

func TestNewConfig(t *testing.T) {
	t.Run("loads from current working directory", func(t *testing.T) {
		dir, cleanup := setupTestDir(t, map[string]string{
			".desktop_cleaner.toml": "file_types = { \"docs\" = [\".docx\"] }",
		})
		defer cleanup()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(dir)

		config := loadTestConfig("")
		// Verify the existence of .docx extension in the FileTypeTree
		found := config.FileTypeTree.Root.FindExtension(".docx")
		assert.True(t, found, "Expected to find '.docx' extension")
	})

	t.Run("loads from optional config path", func(t *testing.T) {
		configContent := "file_types = { \"pics\" = [\".jpg\", \".png\"] }"
		configPath, cleanup := createTestConfigFile(t, configContent)
		defer cleanup()

		config := loadTestConfig(configPath)
		foundJPG := config.FileTypeTree.Root.FindExtension(".jpg")
		foundPNG := config.FileTypeTree.Root.FindExtension(".png")
		assert.True(t, foundJPG, "Expected to find '.jpg' extension")
		assert.True(t, foundPNG, "Expected to find '.png' extension")
	})

	t.Run("creates default config if no config found", func(t *testing.T) {
		config := loadTestConfig("")
		found := config.FileTypeTree.Root.FindExtension(".md")
		assert.True(t, found, "Expected to find '.md' extension in default config")
	})
}

func TestBuildTreeAndCache(t *testing.T) {
	term := terminal.NewTerminal()
	dfs := NewDesktopFS(term)

	dir, cleanup := setupTestDir(t, map[string]string{
		"docs/report.docx": "",
		"pics/photo.jpg":   "",
		"scripts/setup.sh": "",
	})
	defer cleanup()

	newDirTree, err := NewDirectoryTree(dir)
	assert.NoError(t, err)

	dfs.DirectoryTree = newDirTree

	err = dfs.buildTreeAndCache(dir, true, 10)
	assert.NoError(t, err)

	// Check that each expected path is in the cache
	reportDocPath := filepath.Join(dir, "docs", "report.docx")
	photoPath := filepath.Join(dir, "pics", "photo.jpg")
	setupShPath := filepath.Join(dir, "scripts", "setup.sh")

	_, reportExists := dfs.DirectoryTree.Cache[reportDocPath]
	_, photoExists := dfs.DirectoryTree.Cache[photoPath]
	_, setupExists := dfs.DirectoryTree.Cache[setupShPath]

	assert.True(t, reportExists, "Expected report.docx to be in the cache")
	assert.True(t, photoExists, "Expected photo.jpg to be in the cache")
	assert.True(t, setupExists, "Expected setup.sh to be in the cache")
}

func TestPopulateFileTypes(t *testing.T) {
	tree := NewFileTypeTree()
	rules := map[string][]string{
		"docs/Reports":  {".docx", ".pdf"},
		"pics/Photos":   {".jpg", ".png"},
		"scripts/Setup": {".sh"},
	}

	tree.PopulateFileTypes(rules)

	reportNode := tree.FindOrCreatePath([]string{"docs", "Reports"})
	assert.True(t, reportNode.AllowsExtension(".docx"))

	photoNode := tree.FindOrCreatePath([]string{"pics", "Photos"})
	assert.True(t, photoNode.AllowsExtension(".jpg"))

	setupNode := tree.FindOrCreatePath([]string{"scripts", "Setup"})
	assert.True(t, setupNode.AllowsExtension(".sh"))
}

func TestEnhancedOrganize(t *testing.T) {
	// Add cleanup for test resources
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})

	// Add test cases for error conditions
	t.Run("handles concurrent file operations", func(t *testing.T) {
		// Test concurrent file operations
	})

	t.Run("handles file system errors", func(t *testing.T) {
		// Test file system errors
	})

	term := terminal.NewTerminal()
	dfs := NewDesktopFS(term)

	dir, cleanup := setupTestDir(t, map[string]string{
		"source/report.docx":           "",
		"source/photo.jpg":             "",
		"source/setup.sh":              "",
		"target/.desktop_cleaner.toml": `file_types = { "docs/Reports" = [".docx"], "pics/Photos" = [".jpg"], "scripts/Setup" = [".sh"] }`,
	})
	defer cleanup()

	configFile := filepath.Join(dir, "target/.desktop_cleaner.toml")
	dfs.InitConfig(configFile)

	params := &FilePathParams{
		SourceDir:   filepath.Join(dir, "source"),
		TargetDir:   filepath.Join(dir, "target"),
		Recursive:   true,
		CopyFiles:   false,
		RemoveAfter: false,
	}

	fmt.Printf("Expecting organized file paths:\n")
	fmt.Printf("  - %s\n", filepath.Join(dir, "target/docs/Reports/report.docx"))
	fmt.Printf("  - %s\n", filepath.Join(dir, "target/pics/Photos/photo.jpg"))
	fmt.Printf("  - %s\n", filepath.Join(dir, "target/scripts/Setup/setup.sh"))

	// Run EnhancedOrganize and capture any errors
	err := dfs.EnhancedOrganize(dfs.InstanceConfig, params)
	assert.Nil(t, err)

	// Check for organized files in expected locations
	expectedFiles := map[string]string{
		"report.docx": filepath.Join(dir, "target/docs/Reports/report.docx"),
		"photo.jpg":   filepath.Join(dir, "target/pics/Photos/photo.jpg"),
		"setup.sh":    filepath.Join(dir, "target/scripts/Setup/setup.sh"),
	}

	for name, path := range expectedFiles {
		fmt.Printf("Checking organized file %s at %s\n", name, path)
		assert.True(t, pathExists(path), fmt.Sprintf("Expected file %s at %s", name, path))
		assert.FileExists(t, path)
	}
}

func TestEnhancedOrganize_NonexistentDirs(t *testing.T) {
	dfs := initDeskFS(t)
	params := &FilePathParams{
		SourceDir: "/nonexistent/source",
		TargetDir: "/nonexistent/target",
		Recursive: true,
		DryRun:    true,
	}
	err := dfs.EnhancedOrganize(dfs.InstanceConfig, params)
	assert.Error(t, err, "Expected error for nonexistent directories")
}

//func TestConfigValidation_DuplicateExtensions(t *testing.T) {
//	cfg := &IntermediateConfig{
//		FileTypes: map[string][]string{
//			"docs": {".txt", ".doc"},
//			"text": {".txt"},
//		},
//	}
//	err := cfg.validateConfig()
//	assert.Error(t, err, "Expected error for duplicate extensions")
//}

func initDeskFS(t *testing.T) *DesktopFS {
	term := terminal.NewTerminal()
	dfs := NewDesktopFS(term)

	dir, cleanup := setupTestDir(t, map[string]string{
		"source/report.docx":           "",
		"source/photo.jpg":             "",
		"source/setup.sh":              "",
		"target/.desktop_cleaner.toml": `file_types = { "docs/Reports" = [".docx"], "pics/Photos" = [".jpg"], "scripts/Setup" = [".sh"] }`,
	})
	defer cleanup()

	configFile := filepath.Join(dir, "target/.desktop_cleaner.toml")
	dfs.InitConfig(configFile)

	return dfs
}

// Helper function to check if a file exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
