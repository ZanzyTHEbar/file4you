package trees

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatriciaPathIndex(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"BasicInsertAndLookup", testPathIndexBasicInsertAndLookup},
		{"PrefixLookup", testPathIndexPrefixLookup},
		{"GetChildren", testPathIndexGetChildren},
		{"RemoveNode", testPathIndexRemoveNode},
		{"NormalizePath", testPathIndexNormalizePath},
		{"Statistics", testPathIndexStatistics},
		{"ConcurrentAccess", testPathIndexConcurrentAccess},
		{"Validation", testPathIndexValidation},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testPathIndexBasicInsertAndLookup(t *testing.T) {
	idx := NewPatriciaPathIndex()

	// Test data
	paths := []string{
		"/home/user/documents",
		"/home/user/downloads",
		"/home/user/pictures",
		"/var/log/system",
		"/usr/local/bin",
	}

	nodes := make([]*DirectoryNode, len(paths))
	for i, path := range paths {
		nodes[i] = NewDirectoryNode(path, nil)
		err := idx.Insert(nodes[i])
		require.NoError(t, err, "Insert should succeed for path: %s", path)
	}

	// Test exact lookups
	for i, path := range paths {
		foundNode, exists := idx.Lookup(path)
		assert.True(t, exists, "Path should exist: %s", path)
		assert.Equal(t, nodes[i], foundNode, "Should return correct node for path: %s", path)
		assert.Equal(t, path, foundNode.Path, "Node path should match: %s", path)
	}

	// Test non-existent paths
	nonExistent := []string{
		"/home/user/videos",
		"/nonexistent",
		"",
	}

	for _, path := range nonExistent {
		foundNode, exists := idx.Lookup(path)
		assert.False(t, exists, "Non-existent path should not be found: %s", path)
		assert.Nil(t, foundNode, "Should return nil for non-existent path: %s", path)
	}

	// Verify size
	assert.Equal(t, int64(len(paths)), idx.Size(), "Size should match number of inserted nodes")
}

func testPathIndexPrefixLookup(t *testing.T) {
	idx := NewPatriciaPathIndex()

	// Test data with common prefixes
	testData := map[string]*DirectoryNode{
		"/home/user/documents":          NewDirectoryNode("/home/user/documents", nil),
		"/home/user/documents/work":     NewDirectoryNode("/home/user/documents/work", nil),
		"/home/user/documents/personal": NewDirectoryNode("/home/user/documents/personal", nil),
		"/home/user/downloads":          NewDirectoryNode("/home/user/downloads", nil),
		"/home/admin/config":            NewDirectoryNode("/home/admin/config", nil),
		"/var/log/app":                  NewDirectoryNode("/var/log/app", nil),
		"/var/log/system":               NewDirectoryNode("/var/log/system", nil),
	}

	// Insert all nodes
	for path, node := range testData {
		err := idx.Insert(node)
		require.NoError(t, err, "Insert should succeed for: %s", path)
	}

	// Test prefix searches
	testCases := []struct {
		prefix   string
		expected []string
	}{
		{
			prefix:   "/home/user/documents",
			expected: []string{"/home/user/documents", "/home/user/documents/work", "/home/user/documents/personal"},
		},
		{
			prefix:   "/home/user",
			expected: []string{"/home/user/documents", "/home/user/documents/work", "/home/user/documents/personal", "/home/user/downloads"},
		},
		{
			prefix:   "/var/log",
			expected: []string{"/var/log/app", "/var/log/system"},
		},
		{
			prefix:   "/home",
			expected: []string{"/home/user/documents", "/home/user/documents/work", "/home/user/documents/personal", "/home/user/downloads", "/home/admin/config"},
		},
		{
			prefix:   "/nonexistent",
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		results := idx.PrefixLookup(tc.prefix)
		assert.Len(t, results, len(tc.expected), "Should find correct number of results for prefix: %s", tc.prefix)

		resultPaths := make([]string, len(results))
		for i, node := range results {
			resultPaths[i] = node.Path
		}

		for _, expectedPath := range tc.expected {
			assert.Contains(t, resultPaths, expectedPath, "Should contain path: %s for prefix: %s", expectedPath, tc.prefix)
		}
	}
}

func testPathIndexGetChildren(t *testing.T) {
	idx := NewPatriciaPathIndex()

	// Build a tree structure
	testPaths := []string{
		"/home",
		"/home/user",
		"/home/admin",
		"/home/user/documents",
		"/home/user/downloads",
		"/home/admin/config",
		"/var",
		"/var/log",
		"/var/cache",
	}

	for _, path := range testPaths {
		node := NewDirectoryNode(path, nil)
		err := idx.Insert(node)
		require.NoError(t, err)
	}

	// Test getting children
	testCases := []struct {
		parent   string
		expected []string
	}{
		{
			parent:   "/home",
			expected: []string{"/home/user", "/home/admin"},
		},
		{
			parent:   "/home/user",
			expected: []string{"/home/user/documents", "/home/user/downloads"},
		},
		{
			parent:   "/var",
			expected: []string{"/var/log", "/var/cache"},
		},
		{
			parent:   "/home/user/documents", // Leaf node
			expected: []string{},
		},
		{
			parent:   "/nonexistent",
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		children := idx.GetChildren(tc.parent)
		assert.Len(t, children, len(tc.expected), "Should find correct number of children for: %s", tc.parent)

		childPaths := make([]string, len(children))
		for i, child := range children {
			childPaths[i] = child.Path
		}

		for _, expectedPath := range tc.expected {
			assert.Contains(t, childPaths, expectedPath, "Should contain child: %s for parent: %s", expectedPath, tc.parent)
		}
	}
}

func testPathIndexRemoveNode(t *testing.T) {
	idx := NewPatriciaPathIndex()

	// Insert test nodes
	paths := []string{
		"/home/user/documents",
		"/home/user/downloads",
		"/var/log/system",
	}

	for _, path := range paths {
		node := NewDirectoryNode(path, nil)
		err := idx.Insert(node)
		require.NoError(t, err)
	}

	initialSize := idx.Size()

	// Remove existing node
	removed := idx.Remove("/home/user/documents")
	assert.True(t, removed, "Should successfully remove existing node")
	assert.Equal(t, initialSize-1, idx.Size(), "Size should decrease after removal")

	// Verify node is gone
	_, exists := idx.Lookup("/home/user/documents")
	assert.False(t, exists, "Removed node should not be found")

	// Remove non-existent node
	removed = idx.Remove("/nonexistent")
	assert.False(t, removed, "Should return false for non-existent node")
	assert.Equal(t, initialSize-1, idx.Size(), "Size should not change for non-existent removal")

	// Verify other nodes still exist
	for _, path := range []string{"/home/user/downloads", "/var/log/system"} {
		_, exists := idx.Lookup(path)
		assert.True(t, exists, "Other nodes should still exist: %s", path)
	}
}

func testPathIndexNormalizePath(t *testing.T) {
	idx := NewPatriciaPathIndex()

	testCases := []struct {
		input    string
		expected string
	}{
		{"/home/user", "/home/user"},
		{"/home/user/", "/home/user"},
		{"/home//user", "/home/user"},
		{"/home/./user", "/home/user"},
		{"/", "/"},
		{"", "."},
		{"/home/../var", "/var"},
		{"C:\\Users\\test", "C:/Users/test"}, // Windows path
	}

	for _, tc := range testCases {
		normalized := idx.normalizePath(tc.input)
		assert.Equal(t, tc.expected, normalized, "Path normalization failed for: %s", tc.input)
	}
}

func testPathIndexStatistics(t *testing.T) {
	idx := NewPatriciaPathIndex()

	// Insert nodes and perform operations
	paths := []string{
		"/home/user/documents",
		"/home/user/downloads",
		"/var/log/system",
	}

	for _, path := range paths {
		node := NewDirectoryNode(path, nil)
		err := idx.Insert(node)
		require.NoError(t, err)
	}

	// Perform lookups
	for _, path := range paths {
		idx.Lookup(path)
	}

	// Perform prefix lookups
	idx.PrefixLookup("/home")
	idx.PrefixLookup("/var")

	// Remove a node
	idx.Remove("/var/log/system")

	stats := idx.GetStats()

	assert.Equal(t, int64(2), stats.TotalNodes, "Should track total nodes correctly")
	assert.Equal(t, int64(3), stats.PathLookups, "Should track path lookups")
	assert.Equal(t, int64(2), stats.PrefixLookups, "Should track prefix lookups")
	assert.Equal(t, int64(3), stats.Insertions, "Should track insertions")
	assert.Equal(t, int64(1), stats.Deletions, "Should track deletions")
	assert.Greater(t, stats.AveragePathDepth, 0.0, "Should calculate average path depth")
}

func testPathIndexConcurrentAccess(t *testing.T) {
	idx := NewPatriciaPathIndex()

	// Concurrent insertions
	const numGoroutines = 10
	const pathsPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	// Concurrent insertions
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			for j := 0; j < pathsPerGoroutine; j++ {
				path := fmt.Sprintf("/worker%d/path%d", workerID, j)
				node := NewDirectoryNode(path, nil)
				err := idx.Insert(node)
				assert.NoError(t, err, "Concurrent insert should succeed")
			}
		}(i)
	}

	// Wait for all insertions to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify total count
	expectedTotal := int64(numGoroutines * pathsPerGoroutine)
	assert.Equal(t, expectedTotal, idx.Size(), "Should handle concurrent insertions correctly")

	// Concurrent lookups
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			for j := 0; j < pathsPerGoroutine; j++ {
				path := fmt.Sprintf("/worker%d/path%d", workerID, j)
				node, exists := idx.Lookup(path)
				assert.True(t, exists, "Concurrent lookup should find node")
				assert.Equal(t, path, node.Path, "Should return correct node")
			}
		}(i)
	}

	// Wait for all lookups to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func testPathIndexValidation(t *testing.T) {
	idx := NewPatriciaPathIndex()

	// Insert test data
	paths := []string{
		"/home/user",
		"/var/log",
		"/usr/bin",
	}

	for _, path := range paths {
		node := NewDirectoryNode(path, nil)
		err := idx.Insert(node)
		require.NoError(t, err)
	}

	// Validation should pass
	errors := idx.Validate()
	assert.Empty(t, errors, "Validation should pass for healthy index")

	// Test validation after clearing
	idx.Clear()
	errors = idx.Validate()
	assert.Empty(t, errors, "Validation should pass for empty index")

	// Verify clear worked
	assert.Equal(t, int64(0), idx.Size(), "Index should be empty after clear")
}

// Benchmark tests for performance validation

func BenchmarkPathIndexInsert(b *testing.B) {
	idx := NewPatriciaPathIndex()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/benchmark/path/%d", i)
		node := NewDirectoryNode(path, nil)
		idx.Insert(node)
	}
}

func BenchmarkPathIndexLookup(b *testing.B) {
	idx := NewPatriciaPathIndex()

	// Pre-populate index
	for i := 0; i < 10000; i++ {
		path := fmt.Sprintf("/benchmark/path/%d", i)
		node := NewDirectoryNode(path, nil)
		idx.Insert(node)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/benchmark/path/%d", i%10000)
		idx.Lookup(path)
	}
}

func BenchmarkPathIndexPrefixLookup(b *testing.B) {
	idx := NewPatriciaPathIndex()

	// Pre-populate index with hierarchical structure
	for i := 0; i < 1000; i++ {
		for j := 0; j < 10; j++ {
			path := fmt.Sprintf("/benchmark/category%d/item%d", i, j)
			node := NewDirectoryNode(path, nil)
			idx.Insert(node)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prefix := fmt.Sprintf("/benchmark/category%d", i%1000)
		idx.PrefixLookup(prefix)
	}
}
