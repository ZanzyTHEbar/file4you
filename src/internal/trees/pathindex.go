package trees

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/armon/go-radix"
)

// PathIndexStats tracks performance metrics for the path index
type PathIndexStats struct {
	TotalNodes       int64
	PathLookups      int64
	PrefixLookups    int64
	Insertions       int64
	Deletions        int64
	AveragePathDepth float64
	mu               sync.RWMutex
}

// PatriciaPathIndex provides ultra-fast O(k) path lookups using a compressed trie (patricia tree)
// where k is the length of the path being searched, not the number of nodes in the tree
type PatriciaPathIndex struct {
	tree  *radix.Tree               // Core patricia tree for path storage
	mu    sync.RWMutex              // Read-write mutex for concurrent access
	stats *PathIndexStats           // Performance tracking
	nodes map[string]*DirectoryNode // Direct path -> node mapping for verification
}

// NewPatriciaPathIndex creates a new patricia tree-based path index
func NewPatriciaPathIndex() *PatriciaPathIndex {
	return &PatriciaPathIndex{
		tree:  radix.New(),
		stats: &PathIndexStats{},
		nodes: make(map[string]*DirectoryNode),
	}
}

// Insert adds a directory node to the path index with O(k) complexity
// where k is the length of the path
func (idx *PatriciaPathIndex) Insert(node *DirectoryNode) error {
	if node == nil {
		return fmt.Errorf("invalid input: node cannot be nil")
	}

	path := idx.normalizePath(node.Path)

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Insert into patricia tree
	_, updated := idx.tree.Insert(path, node)

	// Also maintain direct mapping for verification
	idx.nodes[path] = node

	// Update statistics
	idx.stats.mu.Lock()
	if !updated {
		idx.stats.TotalNodes++
	}
	idx.stats.Insertions++
	idx.updateAverageDepth()
	idx.stats.mu.Unlock()

	slog.Debug("Path index insertion completed",
		"path", path,
		"was_update", updated,
		"total_nodes", idx.stats.TotalNodes)

	return nil
}

// Lookup finds a directory node by its exact path with O(k) complexity
func (idx *PatriciaPathIndex) Lookup(path string) (*DirectoryNode, bool) {
	normalizedPath := idx.normalizePath(path)

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Lookup in patricia tree
	value, found := idx.tree.Get(normalizedPath)

	// Update statistics
	idx.stats.mu.Lock()
	idx.stats.PathLookups++
	idx.stats.mu.Unlock()

	if !found {
		slog.Debug("Path lookup miss", "path", normalizedPath)
		return nil, false
	}

	node := value.(*DirectoryNode)
	slog.Debug("Path lookup hit",
		"path", normalizedPath,
		"node_found", node != nil)

	return node, true
}

// PrefixLookup finds all directory nodes whose paths start with the given prefix
// This enables efficient directory tree navigation and autocomplete functionality
func (idx *PatriciaPathIndex) PrefixLookup(prefix string) []*DirectoryNode {
	normalizedPrefix := idx.normalizePath(prefix)

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var results []*DirectoryNode

	// Walk all keys with the given prefix
	idx.tree.WalkPrefix(normalizedPrefix, func(key string, value interface{}) bool {
		if node, ok := value.(*DirectoryNode); ok {
			results = append(results, node)
		}
		return false // Continue walking
	})

	// Update statistics
	idx.stats.mu.Lock()
	idx.stats.PrefixLookups++
	idx.stats.mu.Unlock()

	slog.Debug("Prefix lookup completed",
		"prefix", normalizedPrefix,
		"results_count", len(results))

	return results
}

// Remove deletes a directory node from the path index
func (idx *PatriciaPathIndex) Remove(path string) bool {
	normalizedPath := idx.normalizePath(path)

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Remove from patricia tree
	_, deleted := idx.tree.Delete(normalizedPath)

	// Remove from direct mapping
	if deleted {
		delete(idx.nodes, normalizedPath)
	}

	// Update statistics
	idx.stats.mu.Lock()
	if deleted {
		idx.stats.TotalNodes--
	}
	idx.stats.Deletions++
	idx.updateAverageDepth()
	idx.stats.mu.Unlock()

	slog.Debug("Path index removal completed",
		"path", normalizedPath,
		"was_deleted", deleted,
		"total_nodes", idx.stats.TotalNodes)

	return deleted
}

// GetChildren returns all direct children of a given directory path
// This is more efficient than walking the tree manually
func (idx *PatriciaPathIndex) GetChildren(parentPath string) []*DirectoryNode {
	normalizedParent := idx.normalizePath(parentPath)
	if !strings.HasSuffix(normalizedParent, "/") {
		normalizedParent += "/"
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var children []*DirectoryNode

	// Walk paths that start with parent path
	idx.tree.WalkPrefix(normalizedParent, func(key string, value interface{}) bool {
		// Skip the parent itself
		if key == strings.TrimSuffix(normalizedParent, "/") {
			return false
		}

		// Only include direct children (no additional slashes after parent)
		remaining := strings.TrimPrefix(key, normalizedParent)
		if !strings.Contains(remaining, "/") && remaining != "" {
			if node, ok := value.(*DirectoryNode); ok {
				children = append(children, node)
			}
		}
		return false // Continue walking
	})

	slog.Debug("Get children completed",
		"parent_path", normalizedParent,
		"children_count", len(children))

	return children
}

// Size returns the total number of nodes in the path index
func (idx *PatriciaPathIndex) Size() int64 {
	idx.stats.mu.RLock()
	defer idx.stats.mu.RUnlock()
	return idx.stats.TotalNodes
}

// GetStats returns a copy of the current path index statistics
func (idx *PatriciaPathIndex) GetStats() PathIndexStats {
	idx.stats.mu.RLock()
	defer idx.stats.mu.RUnlock()

	return PathIndexStats{
		TotalNodes:       idx.stats.TotalNodes,
		PathLookups:      idx.stats.PathLookups,
		PrefixLookups:    idx.stats.PrefixLookups,
		Insertions:       idx.stats.Insertions,
		Deletions:        idx.stats.Deletions,
		AveragePathDepth: idx.stats.AveragePathDepth,
	}
}

// Clear removes all entries from the path index
func (idx *PatriciaPathIndex) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.tree = radix.New()
	idx.nodes = make(map[string]*DirectoryNode)

	idx.stats.mu.Lock()
	idx.stats.TotalNodes = 0
	idx.stats.AveragePathDepth = 0
	idx.stats.mu.Unlock()

	slog.Info("Path index cleared")
}

// Validate performs integrity checking between the patricia tree and direct mapping
func (idx *PatriciaPathIndex) Validate() []error {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var errors []error

	// Count nodes in patricia tree
	treeNodeCount := 0
	idx.tree.Walk(func(key string, value interface{}) bool {
		treeNodeCount++

		// Verify the direct mapping exists
		if _, exists := idx.nodes[key]; !exists {
			errors = append(errors, fmt.Errorf("patricia_tree_mapping_missing: path exists in patricia tree but missing from direct mapping: %s", key))
		}

		// Verify the value is a valid DirectoryNode
		if _, ok := value.(*DirectoryNode); !ok {
			errors = append(errors, fmt.Errorf("invalid_node_type: invalid node type in patricia tree: %s", key))
		}

		return false // Continue walking
	})

	// Verify counts match
	if treeNodeCount != len(idx.nodes) {
		errors = append(errors, fmt.Errorf("count_mismatch: patricia tree and direct mapping have different counts"))
	}

	// Verify statistics
	if idx.stats.TotalNodes != int64(treeNodeCount) {
		errors = append(errors, fmt.Errorf("stats_mismatch: statistics don't match actual counts"))
	}

	if len(errors) > 0 {
		slog.Warn("Path index validation found issues", "error_count", len(errors))
	} else {
		slog.Debug("Path index validation passed", "total_nodes", treeNodeCount)
	}

	return errors
}

// normalizePath ensures consistent path formatting for the index
func (idx *PatriciaPathIndex) normalizePath(path string) string {
	// First replace backslashes with forward slashes (for Windows paths)
	normalized := strings.ReplaceAll(path, "\\", "/")

	// Then clean the path to resolve . and .. elements
	normalized = filepath.ToSlash(filepath.Clean(normalized))

	// Remove trailing slash unless it's the root
	if len(normalized) > 1 && strings.HasSuffix(normalized, "/") {
		normalized = strings.TrimSuffix(normalized, "/")
	}

	return normalized
}

// updateAverageDepth recalculates the average path depth (called with stats mutex held)
func (idx *PatriciaPathIndex) updateAverageDepth() {
	if idx.stats.TotalNodes == 0 {
		idx.stats.AveragePathDepth = 0
		return
	}

	totalDepth := 0
	for path := range idx.nodes {
		depth := strings.Count(path, "/")
		if path != "/" { // Root has depth 0, everything else adds 1
			depth++
		}
		totalDepth += depth
	}

	idx.stats.AveragePathDepth = float64(totalDepth) / float64(idx.stats.TotalNodes)
}

// WalkPaths executes a function for each path in the index
// The function receives the path and associated DirectoryNode
func (idx *PatriciaPathIndex) WalkPaths(fn func(path string, node *DirectoryNode) bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	idx.tree.Walk(func(key string, value interface{}) bool {
		if node, ok := value.(*DirectoryNode); ok {
			return fn(key, node)
		}
		return false // Continue if type assertion fails
	})
}

// GetLongestCommonPrefix finds the longest common path prefix among all indexed paths
func (idx *PatriciaPathIndex) GetLongestCommonPrefix() string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if idx.stats.TotalNodes == 0 {
		return ""
	}

	var longestPrefix string
	first := true

	idx.tree.Walk(func(key string, value interface{}) bool {
		if first {
			longestPrefix = key
			first = false
		} else {
			longestPrefix = idx.commonPrefix(longestPrefix, key)
		}
		return len(longestPrefix) == 0 // Stop if no common prefix
	})

	return longestPrefix
}

// commonPrefix finds the common prefix between two paths
func (idx *PatriciaPathIndex) commonPrefix(path1, path2 string) string {
	minLen := len(path1)
	if len(path2) < minLen {
		minLen = len(path2)
	}

	for i := 0; i < minLen; i++ {
		if path1[i] != path2[i] {
			return path1[:i]
		}
	}

	return path1[:minLen]
}
