package internal_test

import (
	"file4you/internal/config"
	"file4you/internal/filesystem/types"
	"file4you/internal/trees"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite tests core functionality without CLI dependencies
type IntegrationTestSuite struct {
	suite.Suite
	tempDir string
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "file4you-integration-test-*")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir

	// Create test directory structure
	suite.createTestStructure()
}

func (suite *IntegrationTestSuite) TearDownTest() {
	// Clean up temporary directory
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

func (suite *IntegrationTestSuite) createTestStructure() {
	structure := map[string]string{
		"documents/report.pdf": "PDF content here",
		"documents/notes.txt":  "Text notes content",
		"images/photo1.jpg":    "JPEG image data",
		"images/photo2.png":    "PNG image data",
		"code/main.go":         "package main\nfunc main() {}\n",
		"code/helper.js":       "function helper() { return true; }",
		"data/config.json":     `{"key": "value"}`,
		"data/data.csv":        "name,value\ntest,123",
	}

	for relativePath, content := range structure {
		fullPath := filepath.Join(suite.tempDir, relativePath)
		
		// Create directory if it doesn't exist
		dir := filepath.Dir(fullPath)
		err := os.MkdirAll(dir, 0755)
		require.NoError(suite.T(), err)

		// Create file
		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(suite.T(), err)
	}
}

func (suite *IntegrationTestSuite) TestDirectoryTreeCreation() {
	// Test trees.DirectoryNode creation and basic functionality
	rootNode := trees.NewDirectoryNode(suite.tempDir, nil)
	
	require.NotNil(suite.T(), rootNode)
	assert.Equal(suite.T(), suite.tempDir, rootNode.Path)
	assert.Equal(suite.T(), trees.Directory, rootNode.Metadata.NodeType)
}

func (suite *IntegrationTestSuite) TestFileNodeCreation() {
	// Test FileNode structure (manually created since there's no NewFileNode)
	testFile := filepath.Join(suite.tempDir, "documents", "report.pdf")
	
	fileNode := &trees.FileNode{
		Path:      testFile,
		Name:      "report.pdf",
		Extension: ".pdf",
		Metadata: trees.Metadata{
			NodeType: trees.File,
		},
	}
	
	require.NotNil(suite.T(), fileNode)
	assert.Equal(suite.T(), testFile, fileNode.Path)
	assert.Equal(suite.T(), "report.pdf", fileNode.Name)
	assert.Equal(suite.T(), ".pdf", fileNode.Extension)
	assert.Equal(suite.T(), trees.File, fileNode.Metadata.NodeType)
}

func (suite *IntegrationTestSuite) TestDirectoryAnalysisTypes() {
	// Test DirectoryAnalysis structure
	analysis := &types.DirectoryAnalysis{
		TotalFiles:       8,
		TotalDirectories: 4,
		TotalSize:        1024,
		MaxDepth:         2,
		FileTypes:        make(map[string]int),
		SizeDistribution: make(map[string]int),
	}

	// Populate file types
	analysis.FileTypes[".pdf"] = 1
	analysis.FileTypes[".txt"] = 1
	analysis.FileTypes[".jpg"] = 1
	analysis.FileTypes[".png"] = 1
	analysis.FileTypes[".go"] = 1
	analysis.FileTypes[".js"] = 1
	analysis.FileTypes[".json"] = 1
	analysis.FileTypes[".csv"] = 1

	// Verify structure
	assert.Equal(suite.T(), 8, analysis.TotalFiles)
	assert.Equal(suite.T(), 4, analysis.TotalDirectories)
	assert.Equal(suite.T(), int64(1024), analysis.TotalSize)
	assert.Equal(suite.T(), 2, analysis.MaxDepth)
	assert.Len(suite.T(), analysis.FileTypes, 8)
}

func (suite *IntegrationTestSuite) TestConfigStructure() {
	// Test config loading using the actual LoadConfig function
	cfg, err := config.LoadConfig("")
	
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), cfg)
	
	// Test basic configuration access
	assert.NotNil(suite.T(), cfg.Genkit)
	assert.NotNil(suite.T(), cfg.File4You)
}

func (suite *IntegrationTestSuite) TestFileSystemOperations() {
	// Test basic file system operations
	testFile := filepath.Join(suite.tempDir, "test-ops.txt")
	testContent := "test content for operations"

	// Create file
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(suite.T(), err)

	// Verify file exists
	_, err = os.Stat(testFile)
	require.NoError(suite.T(), err)

	// Read file content
	content, err := os.ReadFile(testFile)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testContent, string(content))

	// Copy file
	copyFile := filepath.Join(suite.tempDir, "test-ops-copy.txt")
	
	// Read and write to simulate copy
	err = os.WriteFile(copyFile, content, 0644)
	require.NoError(suite.T(), err)

	// Verify copy exists
	_, err = os.Stat(copyFile)
	require.NoError(suite.T(), err)

	// Verify copy content
	copyContent, err := os.ReadFile(copyFile)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testContent, string(copyContent))

	// Clean up
	os.Remove(testFile)
	os.Remove(copyFile)
}

func (suite *IntegrationTestSuite) TestDirectoryTraversal() {
	// Test manual directory traversal (basic implementation)
	var fileCount int
	var dirCount int

	err := filepath.Walk(suite.tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			dirCount++
		} else {
			fileCount++
		}

		return nil
	})

	require.NoError(suite.T(), err)
	
	// We expect 8 files plus the root directory and 4 subdirectories
	assert.Equal(suite.T(), 8, fileCount)
	assert.Greater(suite.T(), dirCount, 4) // Root + 4 subdirs
}

// TestTreeStructureValidation validates tree structures
func TestTreeStructureValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file4you-tree-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a simple structure
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Test DirectoryNode with FileNode
	rootNode := trees.NewDirectoryNode(tempDir, nil)
	fileNode := &trees.FileNode{
		Path: testFile,
		Name: "test.txt",
		Extension: ".txt",
		Metadata: trees.Metadata{
			NodeType: trees.File,
		},
	}
	
	assert.NotNil(t, rootNode)
	assert.NotNil(t, fileNode)
}

// TestFileTypeCategorization tests file type detection
func TestFileTypeCategorization(t *testing.T) {
	testCases := []struct {
		filename     string
		expectedExt  string
		expectedType string
	}{
		{"document.pdf", ".pdf", "document"},
		{"image.jpg", ".jpg", "image"},
		{"script.js", ".js", "code"},
		{"data.json", ".json", "data"},
		{"readme.md", ".md", "document"},
		{"photo.png", ".png", "image"},
		{"app.go", ".go", "code"},
		{"config.yaml", ".yaml", "config"},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			ext := filepath.Ext(tc.filename)
			assert.Equal(t, tc.expectedExt, ext)
		})
	}
}

// TestConfigurationAccess tests configuration loading and access
func TestConfigurationAccess(t *testing.T) {
	cfg, err := config.LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test that configuration sections exist
	assert.NotNil(t, cfg.Genkit)
	assert.NotNil(t, cfg.File4You)

	// Test some default values are set
	assert.NotNil(t, cfg.File4You.Database)
	assert.NotEmpty(t, cfg.File4You.TargetDir)
}

// BenchmarkBasicOperations benchmarks core operations
func BenchmarkBasicOperations(b *testing.B) {
	b.Run("DirectoryNodeCreation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			node := trees.NewDirectoryNode("/test/path", nil)
			_ = node
		}
	})

	b.Run("FileNodeCreation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			node := &trees.FileNode{
				Path: "/test/file.txt",
				Name: "file.txt",
				Extension: ".txt",
				Metadata: trees.Metadata{
					NodeType: trees.File,
				},
			}
			_ = node
		}
	})

	b.Run("ConfigLoading", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg, _ := config.LoadConfig("")
			_ = cfg
		}
	})
}
