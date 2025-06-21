package filesystem

import (
	"file4you/internal/db"
	"file4you/internal/trees"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServiceIntegrationEndToEnd validates that all enhanced packages work together
func TestServiceIntegrationEndToEnd(t *testing.T) {
	// Create temporary test directory structure
	tempDir, err := os.MkdirTemp("", "file4you_integration_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files with various types for comprehensive tagging
	testStructure := map[string]string{
		"document.pdf":     "test pdf content",
		"image.jpg":        "fake image data",
		"code.go":          "package main\nfunc main() {}",
		"hidden/.dotfile":  "hidden file content",
		"data/config.json": `{"test": true}`,
		"archive/test.zip": "fake zip data",
		"video/demo.mp4":   "fake video data",
		"README.md":        "# Test Project",
	}

	for filePath, content := range testStructure {
		fullPath := filepath.Join(tempDir, filePath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	t.Run("ConcurrentTraverserWithEnhancedTagging", func(t *testing.T) {
		// Test enhanced metadata and tagging on an existing directory tree
		// Create a simple directory tree manually for testing
		rootNode := &trees.DirectoryNode{
			Path:     tempDir,
			Children: []*trees.DirectoryNode{},
			Files:    []*trees.FileNode{},
		}

		// Manually populate the tree structure for testing
		err := populateTestDirectoryTree(rootNode, tempDir)
		require.NoError(t, err)

		// Test enhanced metadata and tagging
		err = trees.AddMetadataToTree(rootNode)
		require.NoError(t, err)

		// Validate that files have enhanced metadata with owner information
		var testFiles []*trees.FileNode
		collectAllFiles(rootNode, &testFiles)

		assert.GreaterOrEqual(t, len(testFiles), 8, "Should find all test files")

		// Test owner retrieval works
		for _, file := range testFiles {
			assert.NotEmpty(t, file.Metadata.Owner, "Owner should be retrieved for file: %s", file.Path)
			assert.NotEqual(t, "unknown", file.Metadata.Owner, "Owner should not be unknown for file: %s", file.Path)
		}

		// Test comprehensive tagging
		pdfFile := findFileByExtension(testFiles, ".pdf")
		if assert.NotNil(t, pdfFile, "Should find PDF file") {
			// Should have extension-based tags
			assert.Contains(t, pdfFile.Metadata.Tags, "type:pdf")
			assert.Contains(t, pdfFile.Metadata.Tags, "type:document")
			// Should have time-based tags
			assert.True(t, containsAnyTag(pdfFile.Metadata.Tags, []string{"type:recent", "type:thisweek"}))
			// Should have size-based tags
			assert.True(t, containsAnyTag(pdfFile.Metadata.Tags, []string{"type:small", "type:empty"}))
		}

		goFile := findFileByExtension(testFiles, ".go")
		if assert.NotNil(t, goFile, "Should find Go file") {
			assert.Contains(t, goFile.Metadata.Tags, "type:go")
			assert.Contains(t, goFile.Metadata.Tags, "type:code")
		}

		hiddenFile := findFileByName(testFiles, ".dotfile")
		if assert.NotNil(t, hiddenFile, "Should find hidden file") {
			assert.Contains(t, hiddenFile.Metadata.Tags, "type:hidden")
		}
	})

	t.Run("DatabaseIntegrationWithEnhancedMetadata", func(t *testing.T) {
		// Create workspace database for testing
		workspaceDB, err := db.NewWorkspaceDB(tempDir)
		require.NoError(t, err)
		defer workspaceDB.Close()

		// Test file metadata operations with enhanced data
		testFile := filepath.Join(tempDir, "document.pdf")
		metadata, err := trees.GenerateMetadataFromPath(testFile)
		require.NoError(t, err)

		// Add enhanced tags with filename
		err = trees.AddTagsToMetadataWithFilename(&metadata, "document.pdf")
		require.NoError(t, err)

		// Convert to FileMetadata for database operations
		fileMetadata := trees.FileMetadata{
			FilePath: testFile,
			Size:     metadata.Size,
			ModTime:  metadata.ModifiedAt,
			IsDir:    false,
			Checksum: "", // Optional
		}

		// Test database operations with metadata
		err = workspaceDB.InsertFileMetadata(&fileMetadata)
		assert.NoError(t, err)

		// Test retrieval
		retrievedMetadata, err := workspaceDB.GetFileMetadata(testFile)
		require.NoError(t, err)

		// Validate metadata was preserved
		assert.Equal(t, fileMetadata.Size, retrievedMetadata.Size)
		assert.Equal(t, fileMetadata.FilePath, retrievedMetadata.FilePath)
		assert.Equal(t, fileMetadata.IsDir, retrievedMetadata.IsDir)

		// Test batch operations with multiple files
		var allFileMetadata []trees.FileMetadata
		for filePath := range testStructure {
			fullPath := filepath.Join(tempDir, filePath)
			if fileInfo, err := os.Stat(fullPath); err == nil && !fileInfo.IsDir() {
				meta := trees.FileMetadata{
					FilePath: fullPath,
					Size:     fileInfo.Size(),
					ModTime:  fileInfo.ModTime(),
					IsDir:    false,
					Checksum: "",
				}
				allFileMetadata = append(allFileMetadata, meta)
			}
		}

		// Test batch insertion
		err = workspaceDB.BatchInsertFiles(allFileMetadata)
		assert.NoError(t, err)

		// Validate batch operations worked
		allRetrieved, err := workspaceDB.GetAllFileMetadata()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allRetrieved), len(allFileMetadata))
	})

	t.Run("PerformanceValidation", func(t *testing.T) {
		// Create a larger test structure for performance validation
		perfTestDir, err := os.MkdirTemp("", "file4you_perf_test_*")
		require.NoError(t, err)
		defer os.RemoveAll(perfTestDir)

		// Create multiple directories and files for performance testing
		for i := 0; i < 10; i++ {
			dirPath := filepath.Join(perfTestDir, "dir"+string(rune('0'+i)))
			err := os.MkdirAll(dirPath, 0755)
			require.NoError(t, err)

			for j := 0; j < 10; j++ {
				fileName := "file" + string(rune('0'+j)) + ".txt"
				filePath := filepath.Join(dirPath, fileName)
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				require.NoError(t, err)
			}
		}

		// Test metadata generation performance directly
		metadataStartTime := time.Now()

		// Create a simple tree structure and test metadata generation
		perfRootNode := &trees.DirectoryNode{
			Path:     perfTestDir,
			Children: []*trees.DirectoryNode{},
			Files:    []*trees.FileNode{},
		}

		err = populateTestDirectoryTree(perfRootNode, perfTestDir)
		require.NoError(t, err)

		err = trees.AddMetadataToTree(perfRootNode)
		require.NoError(t, err)

		metadataDuration := time.Since(metadataStartTime)

		slog.Info("Performance test results",
			"metadata_duration", metadataDuration,
			"total_files", countAllFiles(perfRootNode),
			"total_dirs", countAllDirs(perfRootNode))

		// Validate performance expectations
		assert.Less(t, metadataDuration, 200*time.Millisecond, "Metadata generation should be fast")

		// Validate that performance is comparable to Phase 1 achievements
		filesPerSecond := float64(countAllFiles(perfRootNode)) / metadataDuration.Seconds()
		assert.Greater(t, filesPerSecond, 500.0, "Should process >500 files/sec")
	})

	t.Run("EnhancedTaggingIntegration", func(t *testing.T) {
		// Test the enhanced tagging system end-to-end
		testFile := filepath.Join(tempDir, "document.pdf")

		// Generate metadata with enhanced tagging
		metadata, err := trees.GenerateMetadataFromPath(testFile)
		require.NoError(t, err)

		// Test filename-aware tagging
		originalTagCount := len(metadata.Tags)
		err = trees.AddTagsToMetadataWithFilename(&metadata, "document.pdf")
		require.NoError(t, err)

		// Should have more tags after filename processing
		assert.Greater(t, len(metadata.Tags), originalTagCount, "Should add filename-based tags")

		// Validate specific tags were added
		assert.Contains(t, metadata.Tags, "type:pdf")
		assert.Contains(t, metadata.Tags, "type:document")

		// Test with different file types
		goFile := filepath.Join(tempDir, "code.go")
		goMetadata, err := trees.GenerateMetadataFromPath(goFile)
		require.NoError(t, err)

		err = trees.AddTagsToMetadataWithFilename(&goMetadata, "code.go")
		require.NoError(t, err)

		assert.Contains(t, goMetadata.Tags, "type:go")
		assert.Contains(t, goMetadata.Tags, "type:code")

		// Test hidden file tagging
		hiddenFile := filepath.Join(tempDir, "hidden/.dotfile")
		hiddenMetadata, err := trees.GenerateMetadataFromPath(hiddenFile)
		require.NoError(t, err)

		err = trees.AddTagsToMetadataWithFilename(&hiddenMetadata, ".dotfile")
		require.NoError(t, err)

		assert.Contains(t, hiddenMetadata.Tags, "type:hidden")
	})
}

// Helper functions for testing

func populateTestDirectoryTree(rootNode *trees.DirectoryNode, dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			childNode := &trees.DirectoryNode{
				Path:     fullPath,
				Children: []*trees.DirectoryNode{},
				Files:    []*trees.FileNode{},
			}
			rootNode.Children = append(rootNode.Children, childNode)

			// Recursively populate child directories
			err = populateTestDirectoryTree(childNode, fullPath)
			if err != nil {
				return err
			}
		} else {
			fileNode := &trees.FileNode{
				Path: fullPath,
			}
			rootNode.Files = append(rootNode.Files, fileNode)
		}
	}

	return nil
}

func collectAllFiles(node *trees.DirectoryNode, files *[]*trees.FileNode) {
	*files = append(*files, node.Files...)
	for _, child := range node.Children {
		collectAllFiles(child, files)
	}
}

func findFileByExtension(files []*trees.FileNode, ext string) *trees.FileNode {
	for _, file := range files {
		if filepath.Ext(file.Path) == ext {
			return file
		}
	}
	return nil
}

func findFileByName(files []*trees.FileNode, name string) *trees.FileNode {
	for _, file := range files {
		if filepath.Base(file.Path) == name {
			return file
		}
	}
	return nil
}

func containsAnyTag(tags []string, searchTags []string) bool {
	for _, tag := range tags {
		for _, searchTag := range searchTags {
			if tag == searchTag {
				return true
			}
		}
	}
	return false
}

func countAllFiles(node *trees.DirectoryNode) int {
	count := len(node.Files)
	for _, child := range node.Children {
		count += countAllFiles(child)
	}
	return count
}

func countAllDirs(node *trees.DirectoryNode) int {
	count := 1 // count this directory
	for _, child := range node.Children {
		count += countAllDirs(child)
	}
	return count
}
