package trees

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiIndexIntegration(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"BasicMultiIndexOperations", testMultiIndexBasicOperations},
		{"QueryPerformanceComparison", testMultiIndexPerformanceComparison},
		{"ConcurrentMultiIndexAccess", testMultiIndexConcurrentAccess},
		{"DirectoryTreeIntegration", testDirectoryTreeIntegration},
		{"IndexValidationAndStats", testIndexValidationAndStats},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testMultiIndexBasicOperations(t *testing.T) {
	mi := NewMultiIndex()

	// Create test nodes with various characteristics
	testNodes := []*DirectoryNode{
		createTestDirectoryNode("/home/user/documents", 1024*1024, "2024-01-15"),     // 1MB
		createTestDirectoryNode("/home/user/pictures", 512*1024*1024, "2024-02-20"),  // 512MB
		createTestDirectoryNode("/home/user/videos", 2*1024*1024*1024, "2024-03-10"), // 2GB
		createTestDirectoryNode("/var/log/system", 10*1024, "2024-06-01"),            // 10KB
		createTestDirectoryNode("/usr/local/bin", 5*1024*1024, "2024-05-15"),         // 5MB
	}

	// Add files to directories for category testing
	addTestFilesToNode(testNodes[0], []string{"document.pdf", "report.docx", "notes.txt"})
	addTestFilesToNode(testNodes[1], []string{"photo1.jpg", "image.png", "vacation.gif"})
	addTestFilesToNode(testNodes[2], []string{"movie.mp4", "video.avi", "clip.mov"})
	addTestFilesToNode(testNodes[3], []string{"error.log", "access.log", "debug.txt"})
	addTestFilesToNode(testNodes[4], []string{"binary", "script.sh", "program.exe"})

	// Insert all nodes
	for _, node := range testNodes {
		err := mi.Insert(node)
		require.NoError(t, err, "Should insert node successfully: %s", node.Path)
	}

	// Test path queries
	node, found := mi.QueryByPath("/home/user/documents")
	assert.True(t, found, "Should find exact path")
	assert.Equal(t, "/home/user/documents", node.Path)

	// Test path prefix queries
	homeNodes := mi.QueryByPathPrefix("/home/user")
	assert.Len(t, homeNodes, 3, "Should find 3 nodes under /home/user")

	// Test size range queries
	largeNodes := mi.QueryBySizeRange(100*1024*1024, 3*1024*1024*1024) // 100MB - 3GB
	assert.Len(t, largeNodes, 2, "Should find 2 large nodes (pictures and videos)")

	// Test time range queries
	start, _ := time.Parse("2006-01-02", "2024-02-01")
	end, _ := time.Parse("2006-01-02", "2024-04-01")
	recentNodes := mi.QueryByTimeRange(start, end)
	assert.Len(t, recentNodes, 2, "Should find 2 nodes in time range")

	// Test extension queries
	imageNodes := mi.QueryByExtension(".jpg")
	assert.Len(t, imageNodes, 1, "Should find 1 directory with .jpg files")
	// Test category queries
	documentNodes := mi.QueryByCategory("documents")
	assert.Len(t, documentNodes, 2, "Should find 2 directories with document files (.pdf/.docx/.txt)")

	videoNodes := mi.QueryByCategory("videos")
	assert.Len(t, videoNodes, 1, "Should find 1 directory with video files")
}

func testMultiIndexPerformanceComparison(t *testing.T) {
	const numNodes = 1000

	// Setup multi-index
	mi := NewMultiIndex()

	// Create test dataset
	nodes := make([]*DirectoryNode, numNodes)
	for i := 0; i < numNodes; i++ {
		path := fmt.Sprintf("/test/category%d/item%d", i/100, i%100)
		nodes[i] = createTestDirectoryNode(path, int64(i*1024), "2024-01-01")

		// Add some files for testing
		if i%10 == 0 {
			addTestFilesToNode(nodes[i], []string{"file.txt", "data.json"})
		}
	}

	// Benchmark insertions
	insertStart := time.Now()
	for _, node := range nodes {
		err := mi.Insert(node)
		require.NoError(t, err)
	}
	insertDuration := time.Since(insertStart)

	// Benchmark path lookups
	lookupStart := time.Now()
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("/test/category%d/item%d", i/10, i%10)
		_, found := mi.QueryByPath(path)
		assert.True(t, found, "Should find path: %s", path)
	}
	lookupDuration := time.Since(lookupStart)

	// Benchmark prefix queries
	prefixStart := time.Now()
	for i := 0; i < 10; i++ {
		prefix := fmt.Sprintf("/test/category%d", i)
		results := mi.QueryByPathPrefix(prefix)
		assert.True(t, len(results) > 0, "Should find results for prefix: %s", prefix)
	}
	prefixDuration := time.Since(prefixStart)

	// Log performance metrics
	stats := mi.GetStats()
	t.Logf("Performance Results:")
	t.Logf("  Insertions: %d nodes in %v (%.2f ops/sec)",
		numNodes, insertDuration, float64(numNodes)/insertDuration.Seconds())
	t.Logf("  Path Lookups: 100 queries in %v (%.2f ops/sec)",
		lookupDuration, 100.0/lookupDuration.Seconds())
	t.Logf("  Prefix Queries: 10 queries in %v (%.2f ops/sec)",
		prefixDuration, 10.0/prefixDuration.Seconds())
	t.Logf("  Total Operations: %d", stats.TotalOperations)
	t.Logf("  Average Query Time: %v", stats.AverageQueryTime)

	// Performance assertions
	assert.Less(t, insertDuration.Milliseconds(), int64(1000), "Insertions should complete within 1 second")
	assert.Less(t, lookupDuration.Milliseconds(), int64(100), "Lookups should complete within 100ms")
	assert.Less(t, prefixDuration.Milliseconds(), int64(100), "Prefix queries should complete within 100ms")
}

func testMultiIndexConcurrentAccess(t *testing.T) {
	mi := NewMultiIndex()
	const numGoroutines = 10
	const operationsPerGoroutine = 100

	// Concurrent insertions
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				path := fmt.Sprintf("/worker%d/item%d", workerID, j)
				node := createTestDirectoryNode(path, int64(j*1024), "2024-01-01")

				err := mi.Insert(node)
				assert.NoError(t, err, "Concurrent insert should succeed")
			}
		}(i)
	}

	// Wait for all insertions
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Concurrent queries
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				path := fmt.Sprintf("/worker%d/item%d", workerID, j)
				node, found := mi.QueryByPath(path)
				assert.True(t, found, "Concurrent query should find node")
				assert.Equal(t, path, node.Path, "Should return correct node")
			}
		}(i)
	}

	// Wait for all queries
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final state
	expectedTotal := int64(numGoroutines * operationsPerGoroutine)
	assert.Equal(t, expectedTotal, mi.Size(), "Should handle concurrent operations correctly")
}

func testDirectoryTreeIntegration(t *testing.T) {
	// Create DirectoryTree with multi-index integration
	dt := NewDirectoryTree(WithRoot("/test"))

	// Add nodes through DirectoryTree (should update all indexes)
	testPaths := []string{
		"/test/documents/work",
		"/test/documents/personal",
		"/test/pictures/vacation",
		"/test/videos/movies",
		"/test/downloads/software",
	}

	for _, path := range testPaths {
		node := createTestDirectoryNode(path, 1024*1024, "2024-01-01")
		err := dt.AddNode(node)
		require.NoError(t, err, "Should add node to DirectoryTree: %s", path)
	}

	// Test DirectoryTree query methods
	node, found := dt.FindByPath("/test/documents/work")
	assert.True(t, found, "DirectoryTree should find path")
	assert.Equal(t, "/test/documents/work", node.Path)

	// Test prefix queries
	docNodes := dt.FindByPathPrefix("/test/documents")
	assert.Len(t, docNodes, 2, "Should find 2 document nodes")

	// Test size range queries
	mediumNodes := dt.FindBySizeRange(512*1024, 2*1024*1024) // 512KB - 2MB
	assert.Len(t, mediumNodes, 5, "Should find all medium-sized nodes")

	// Test time range queries
	start, _ := time.Parse("2006-01-02", "2023-12-01")
	end, _ := time.Parse("2006-01-02", "2024-02-01")
	recentNodes := dt.FindByTimeRange(start, end)
	assert.Len(t, recentNodes, 5, "Should find all recent nodes")

	// Test index statistics
	stats, err := dt.GetIndexStats()
	require.NoError(t, err, "Should get index statistics")
	assert.Contains(t, stats, "path_index", "Should contain path index stats")
	assert.Contains(t, stats, "multi_index", "Should contain multi-index stats")

	// Test index validation
	errors := dt.ValidateIndexes()
	assert.Empty(t, errors, "Index validation should pass")
}

func testIndexValidationAndStats(t *testing.T) {
	mi := NewMultiIndex()

	// Add test data
	for i := 0; i < 50; i++ {
		path := fmt.Sprintf("/test/item%d", i)
		node := createTestDirectoryNode(path, int64(i*1024), "2024-01-01")
		err := mi.Insert(node)
		require.NoError(t, err)
	}

	// Perform various queries to populate statistics
	for i := 0; i < 10; i++ {
		mi.QueryByPath(fmt.Sprintf("/test/item%d", i))
		mi.QueryByPathPrefix("/test")
		mi.QueryBySizeRange(0, 100*1024)
	}

	// Check statistics
	stats := mi.GetStats()
	assert.Equal(t, int64(80), stats.TotalOperations, "Should track all operations (50 inserts + 30 queries)")
	assert.Equal(t, int64(20), stats.PathQueries, "Should track path queries (10 direct + 10 prefix)")
	assert.Equal(t, int64(10), stats.SizeQueries, "Should track size queries")
	assert.Greater(t, stats.AverageQueryTime, time.Duration(0), "Should track query times")

	// Validate indexes
	errors := mi.Validate()
	assert.Empty(t, errors, "Multi-index validation should pass")

	// Test clear functionality
	mi.Clear()
	assert.Equal(t, int64(0), mi.Size(), "Should be empty after clear")

	clearedStats := mi.GetStats()
	assert.Equal(t, int64(0), clearedStats.TotalOperations, "Should reset statistics")
}

// Helper functions for testing

func createTestDirectoryNode(path string, size int64, dateStr string) *DirectoryNode {
	modTime, _ := time.Parse("2006-01-02", dateStr)

	node := NewDirectoryNode(path, nil)
	node.Metadata = Metadata{
		Size:       size,
		ModifiedAt: modTime,
		CreatedAt:  modTime,
		NodeType:   Directory,
	}

	return node
}

func addTestFilesToNode(node *DirectoryNode, filenames []string) {
	for _, filename := range filenames {
		ext := ""
		if dotIndex := strings.LastIndex(filename, "."); dotIndex != -1 {
			ext = filename[dotIndex:]
		}

		file := &FileNode{
			Path:      node.Path + "/" + filename,
			Name:      filename,
			Extension: ext,
			Metadata: Metadata{
				Size:     1024,
				NodeType: File,
			},
		}

		node.AddFile(file)
	}
}

// Benchmark tests for performance validation

func BenchmarkMultiIndexInsert(b *testing.B) {
	mi := NewMultiIndex()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/benchmark/item%d", i)
		node := createTestDirectoryNode(path, int64(i*1024), "2024-01-01")
		mi.Insert(node)
	}
}

func BenchmarkMultiIndexPathQuery(b *testing.B) {
	mi := NewMultiIndex()

	// Pre-populate
	for i := 0; i < 10000; i++ {
		path := fmt.Sprintf("/benchmark/item%d", i)
		node := createTestDirectoryNode(path, int64(i*1024), "2024-01-01")
		mi.Insert(node)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/benchmark/item%d", i%10000)
		mi.QueryByPath(path)
	}
}

func BenchmarkMultiIndexPrefixQuery(b *testing.B) {
	mi := NewMultiIndex()

	// Pre-populate with hierarchical structure
	for i := 0; i < 1000; i++ {
		for j := 0; j < 10; j++ {
			path := fmt.Sprintf("/benchmark/category%d/item%d", i, j)
			node := createTestDirectoryNode(path, int64(j*1024), "2024-01-01")
			mi.Insert(node)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prefix := fmt.Sprintf("/benchmark/category%d", i%1000)
		mi.QueryByPathPrefix(prefix)
	}
}

func BenchmarkDirectoryTreeIntegration(b *testing.B) {
	dt := NewDirectoryTree(WithRoot("/benchmark"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/benchmark/item%d", i)
		node := createTestDirectoryNode(path, int64(i*1024), "2024-01-01")
		dt.AddNode(node)

		// Perform a query every 10 insertions
		if i%10 == 0 {
			dt.FindByPath(path)
		}
	}
}
