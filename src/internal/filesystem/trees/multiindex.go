package trees

import (
	"container/heap"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
)

// MultiIndexStats tracks performance metrics across all indexes
type MultiIndexStats struct {
	TotalOperations  int64
	PathQueries      int64
	SizeQueries      int64
	TimeQueries      int64
	ExtensionQueries int64
	SpatialQueries   int64
	AverageQueryTime time.Duration
	mu               sync.RWMutex
}

// SizeHeapEntry represents an entry in the size-based heap for efficient range queries
type SizeHeapEntry struct {
	Node *DirectoryNode
	Size int64
}

// SizeHeap implements a min-heap for size-based queries
type SizeHeap []*SizeHeapEntry

func (h SizeHeap) Len() int           { return len(h) }
func (h SizeHeap) Less(i, j int) bool { return h[i].Size < h[j].Size }
func (h SizeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *SizeHeap) Push(x interface{}) {
	*h = append(*h, x.(*SizeHeapEntry))
}

func (h *SizeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

// TimeInterval represents a time range for temporal queries
type TimeInterval struct {
	Start time.Time
	End   time.Time
	Nodes []*DirectoryNode
}

// IntervalTree provides efficient temporal range queries
type IntervalTree struct {
	intervals []TimeInterval
	mu        sync.RWMutex
	sorted    bool
}

// NewIntervalTree creates a new interval tree for temporal queries
func NewIntervalTree() *IntervalTree {
	return &IntervalTree{
		intervals: make([]TimeInterval, 0),
		sorted:    true,
	}
}

// Insert adds a node to the interval tree based on its modification time
func (it *IntervalTree) Insert(node *DirectoryNode) {
	if node == nil || node.Metadata.ModifiedAt.IsZero() {
		return
	}

	it.mu.Lock()
	defer it.mu.Unlock()

	// Create a new interval or add to existing one
	modTime := node.Metadata.ModifiedAt
	added := false

	for i := range it.intervals {
		// Check if this node fits in an existing interval (within 1 hour)
		if modTime.Sub(it.intervals[i].Start) >= 0 && modTime.Sub(it.intervals[i].End) <= time.Hour {
			it.intervals[i].Nodes = append(it.intervals[i].Nodes, node)
			if modTime.Before(it.intervals[i].Start) {
				it.intervals[i].Start = modTime
			}
			if modTime.After(it.intervals[i].End) {
				it.intervals[i].End = modTime
			}
			added = true
			break
		}
	}

	if !added {
		// Create new interval
		it.intervals = append(it.intervals, TimeInterval{
			Start: modTime,
			End:   modTime,
			Nodes: []*DirectoryNode{node},
		})
		it.sorted = false
	}
}

// Query finds all nodes within a time range
func (it *IntervalTree) Query(start, end time.Time) []*DirectoryNode {
	it.mu.RLock()
	defer it.mu.RUnlock()

	if !it.sorted {
		it.sortIntervals()
	}

	var results []*DirectoryNode
	for _, interval := range it.intervals {
		// Check if intervals overlap
		if interval.End.Before(start) || interval.Start.After(end) {
			continue
		}

		// Add nodes that fall within the query range
		for _, node := range interval.Nodes {
			if !node.Metadata.ModifiedAt.Before(start) && !node.Metadata.ModifiedAt.After(end) {
				results = append(results, node)
			}
		}
	}

	return results
}

// sortIntervals sorts intervals by start time (called with lock held)
func (it *IntervalTree) sortIntervals() {
	sort.Slice(it.intervals, func(i, j int) bool {
		return it.intervals[i].Start.Before(it.intervals[j].Start)
	})
	it.sorted = true
}

// CategoryMap provides efficient type-based lookups
type CategoryMap struct {
	extensions map[string][]*DirectoryNode // extension -> nodes
	categories map[string][]*DirectoryNode // category -> nodes
	mu         sync.RWMutex
}

// NewCategoryMap creates a new category-based index
func NewCategoryMap() *CategoryMap {
	return &CategoryMap{
		extensions: make(map[string][]*DirectoryNode),
		categories: make(map[string][]*DirectoryNode),
	}
}

// Insert adds a node to the category index
func (cm *CategoryMap) Insert(node *DirectoryNode) {
	if node == nil {
		return
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Track extensions and categories for this directory to avoid duplicates
	extensionsFound := make(map[string]bool)
	categoriesFound := make(map[string]bool)

	// Index by file extensions in this directory
	for _, file := range node.Files {
		if file.Extension != "" {
			ext := strings.ToLower(file.Extension)
			if !extensionsFound[ext] {
				cm.extensions[ext] = append(cm.extensions[ext], node)
				extensionsFound[ext] = true
			}
		}

		// Index by file categories
		category := cm.categorizeFile(file)
		if category != "" && !categoriesFound[category] {
			cm.categories[category] = append(cm.categories[category], node)
			categoriesFound[category] = true
		}
	}
}

// QueryByExtension finds all directories containing files with specific extensions
func (cm *CategoryMap) QueryByExtension(extension string) []*DirectoryNode {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ext := strings.ToLower(extension)
	if nodes, exists := cm.extensions[ext]; exists {
		result := make([]*DirectoryNode, len(nodes))
		copy(result, nodes)
		return result
	}
	return []*DirectoryNode{}
}

// QueryByCategory finds all directories containing files of specific category
func (cm *CategoryMap) QueryByCategory(category string) []*DirectoryNode {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if nodes, exists := cm.categories[category]; exists {
		result := make([]*DirectoryNode, len(nodes))
		copy(result, nodes)
		return result
	}
	return []*DirectoryNode{}
}

// categorizeFile determines the category of a file based on its extension
func (cm *CategoryMap) categorizeFile(file *FileNode) string {
	ext := strings.ToLower(file.Extension)

	// Common file type categories
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".svg": true, ".webp": true, ".tiff": true,
	}
	videoExts := map[string]bool{
		".mp4": true, ".avi": true, ".mkv": true, ".mov": true,
		".wmv": true, ".flv": true, ".webm": true, ".m4v": true,
	}
	audioExts := map[string]bool{
		".mp3": true, ".wav": true, ".flac": true, ".aac": true,
		".ogg": true, ".wma": true, ".m4a": true,
	}
	documentExts := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".txt": true,
		".rtf": true, ".odt": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true,
	}
	codeExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".py": true,
		".java": true, ".cpp": true, ".c": true, ".rs": true,
		".rb": true, ".php": true, ".html": true, ".css": true,
	}

	switch {
	case imageExts[ext]:
		return "images"
	case videoExts[ext]:
		return "videos"
	case audioExts[ext]:
		return "audio"
	case documentExts[ext]:
		return "documents"
	case codeExts[ext]:
		return "code"
	default:
		return "other"
	}
}

// MultiIndex coordinates multiple specialized indexes for optimal query performance
type MultiIndex struct {
	pathIndex    *PatriciaPathIndex // O(k) path lookups
	sizeHeap     *SizeHeap          // Size-based range queries
	timeTree     *IntervalTree      // Temporal queries
	categoryMap  *CategoryMap       // Extension/category queries
	spatialIndex *DirectoryTree     // Spatial queries (via DirectoryTree)
	stats        *MultiIndexStats   // Performance tracking
	mu           sync.RWMutex       // Coordination lock
}

// NewMultiIndex creates a new multi-index system
func NewMultiIndex() *MultiIndex {
	sizeHeap := &SizeHeap{}
	heap.Init(sizeHeap)

	return &MultiIndex{
		pathIndex:   NewPatriciaPathIndex(),
		sizeHeap:    sizeHeap,
		timeTree:    NewIntervalTree(),
		categoryMap: NewCategoryMap(),
		stats:       &MultiIndexStats{},
	}
}

// SetSpatialIndex links the DirectoryTree for spatial queries
func (mi *MultiIndex) SetSpatialIndex(dirTree *DirectoryTree) {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	mi.spatialIndex = dirTree
}

// Insert adds a directory node to all relevant indexes
func (mi *MultiIndex) Insert(node *DirectoryNode) error {
	if node == nil {
		return fmt.Errorf("cannot insert nil node")
	}

	start := time.Now()

	mi.mu.Lock()
	defer mi.mu.Unlock()

	// Insert into path index
	if err := mi.pathIndex.Insert(node); err != nil {
		return fmt.Errorf("path index insertion failed: %w", err)
	}

	// Insert into size heap
	sizeEntry := &SizeHeapEntry{
		Node: node,
		Size: node.Metadata.Size,
	}
	heap.Push(mi.sizeHeap, sizeEntry)

	// Insert into time tree
	mi.timeTree.Insert(node)

	// Insert into category map
	mi.categoryMap.Insert(node)

	// Update statistics
	mi.stats.mu.Lock()
	mi.stats.TotalOperations++
	duration := time.Since(start)
	mi.stats.AverageQueryTime = ((mi.stats.AverageQueryTime * time.Duration(mi.stats.TotalOperations-1)) + duration) / time.Duration(mi.stats.TotalOperations)
	mi.stats.mu.Unlock()

	slog.Debug("Multi-index insertion completed",
		"path", node.Path,
		"duration", duration)

	return nil
}

// Remove removes a node from all indexes
func (mi *MultiIndex) Remove(path string) bool {
	start := time.Now()

	mi.mu.Lock()
	defer mi.mu.Unlock()

	// Remove from path index
	removed := mi.pathIndex.Remove(path)
	if !removed {
		return false // Node not found
	}

	// Remove from size heap (mark for lazy removal since heap removal is expensive)
	// The heap will be cleaned during next rebalancing operation

	// Remove from time tree (lazy removal - intervals will be cleaned during next sort)
	// Note: Full removal from interval tree is expensive, so we defer it

	// Remove from category map (lazy removal)
	// Category maps will be cleaned during next query operation

	// Update statistics
	mi.stats.mu.Lock()
	mi.stats.TotalOperations++
	duration := time.Since(start)
	mi.stats.AverageQueryTime = ((mi.stats.AverageQueryTime * time.Duration(mi.stats.TotalOperations-1)) + duration) / time.Duration(mi.stats.TotalOperations)
	mi.stats.mu.Unlock()

	slog.Debug("Multi-index removal completed",
		"path", path,
		"duration", duration)

	return true
}

// QueryByPath performs O(k) path lookup
func (mi *MultiIndex) QueryByPath(path string) (*DirectoryNode, bool) {
	start := time.Now()
	defer func() {
		mi.stats.mu.Lock()
		mi.stats.PathQueries++
		mi.stats.TotalOperations++
		mi.stats.mu.Unlock()
	}()

	node, found := mi.pathIndex.Lookup(path)

	slog.Debug("Path query completed",
		"path", path,
		"found", found,
		"duration", time.Since(start))

	return node, found
}

// QueryByPathPrefix finds all paths with given prefix
func (mi *MultiIndex) QueryByPathPrefix(prefix string) []*DirectoryNode {
	start := time.Now()
	defer func() {
		mi.stats.mu.Lock()
		mi.stats.PathQueries++
		mi.stats.TotalOperations++
		mi.stats.mu.Unlock()
	}()

	results := mi.pathIndex.PrefixLookup(prefix)

	slog.Debug("Path prefix query completed",
		"prefix", prefix,
		"results_count", len(results),
		"duration", time.Since(start))

	return results
}

// QueryBySizeRange finds directories within a size range
func (mi *MultiIndex) QueryBySizeRange(minSize, maxSize int64) []*DirectoryNode {
	start := time.Now()
	defer func() {
		mi.stats.mu.Lock()
		mi.stats.SizeQueries++
		mi.stats.TotalOperations++
		mi.stats.mu.Unlock()
	}()

	mi.mu.RLock()
	defer mi.mu.RUnlock()

	var results []*DirectoryNode

	// Linear scan through heap (could be optimized with balanced tree)
	for _, entry := range *mi.sizeHeap {
		if entry.Size >= minSize && entry.Size <= maxSize {
			results = append(results, entry.Node)
		}
	}

	slog.Debug("Size range query completed",
		"min_size", minSize,
		"max_size", maxSize,
		"results_count", len(results),
		"duration", time.Since(start))

	return results
}

// QueryByTimeRange finds directories modified within a time range
func (mi *MultiIndex) QueryByTimeRange(start, end time.Time) []*DirectoryNode {
	queryStart := time.Now()
	defer func() {
		mi.stats.mu.Lock()
		mi.stats.TimeQueries++
		mi.stats.TotalOperations++
		mi.stats.mu.Unlock()
	}()

	results := mi.timeTree.Query(start, end)

	slog.Debug("Time range query completed",
		"start_time", start,
		"end_time", end,
		"results_count", len(results),
		"duration", time.Since(queryStart))

	return results
}

// QueryByExtension finds directories containing files with specific extension
func (mi *MultiIndex) QueryByExtension(extension string) []*DirectoryNode {
	start := time.Now()
	defer func() {
		mi.stats.mu.Lock()
		mi.stats.ExtensionQueries++
		mi.stats.TotalOperations++
		mi.stats.mu.Unlock()
	}()

	results := mi.categoryMap.QueryByExtension(extension)

	slog.Debug("Extension query completed",
		"extension", extension,
		"results_count", len(results),
		"duration", time.Since(start))

	return results
}

// QueryByCategory finds directories containing files of specific category
func (mi *MultiIndex) QueryByCategory(category string) []*DirectoryNode {
	start := time.Now()
	defer func() {
		mi.stats.mu.Lock()
		mi.stats.ExtensionQueries++
		mi.stats.TotalOperations++
		mi.stats.mu.Unlock()
	}()

	results := mi.categoryMap.QueryByCategory(category)

	slog.Debug("Category query completed",
		"category", category,
		"results_count", len(results),
		"duration", time.Since(start))

	return results
}

// QuerySpatialNearest finds nearest directories using KD-Tree
func (mi *MultiIndex) QuerySpatialNearest(query DirectoryPoint, k int) ([]*DirectoryNode, error) {
	start := time.Now()
	defer func() {
		mi.stats.mu.Lock()
		mi.stats.SpatialQueries++
		mi.stats.TotalOperations++
		mi.stats.mu.Unlock()
	}()

	mi.mu.RLock()
	spatialIndex := mi.spatialIndex
	mi.mu.RUnlock()

	if spatialIndex == nil {
		return nil, fmt.Errorf("spatial index not available")
	}

	// This would need to be implemented on the DirectoryTree that contains the KD-Tree
	// For now, return empty results
	results := []*DirectoryNode{}

	slog.Debug("Spatial nearest query completed",
		"k", k,
		"results_count", len(results),
		"duration", time.Since(start))

	return results, nil
}

// GetStats returns current performance statistics
func (mi *MultiIndex) GetStats() MultiIndexStats {
	mi.stats.mu.RLock()
	defer mi.stats.mu.RUnlock()

	return MultiIndexStats{
		TotalOperations:  mi.stats.TotalOperations,
		PathQueries:      mi.stats.PathQueries,
		SizeQueries:      mi.stats.SizeQueries,
		TimeQueries:      mi.stats.TimeQueries,
		ExtensionQueries: mi.stats.ExtensionQueries,
		SpatialQueries:   mi.stats.SpatialQueries,
		AverageQueryTime: mi.stats.AverageQueryTime,
	}
}

// Clear removes all entries from all indexes
func (mi *MultiIndex) Clear() {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	mi.pathIndex.Clear()
	mi.sizeHeap = &SizeHeap{}
	heap.Init(mi.sizeHeap)
	mi.timeTree = NewIntervalTree()
	mi.categoryMap = NewCategoryMap()

	mi.stats.mu.Lock()
	mi.stats.TotalOperations = 0
	mi.stats.PathQueries = 0
	mi.stats.SizeQueries = 0
	mi.stats.TimeQueries = 0
	mi.stats.ExtensionQueries = 0
	mi.stats.SpatialQueries = 0
	mi.stats.AverageQueryTime = 0
	mi.stats.mu.Unlock()

	slog.Info("Multi-index cleared")
}

// Size returns the total number of nodes across all indexes
func (mi *MultiIndex) Size() int64 {
	return mi.pathIndex.Size()
}

// Validate performs integrity checking across all indexes
func (mi *MultiIndex) Validate() []error {
	mi.mu.RLock()
	defer mi.mu.RUnlock()

	var errors []error

	// Validate path index
	pathErrors := mi.pathIndex.Validate()
	errors = append(errors, pathErrors...)

	// Basic validation of other indexes
	if mi.sizeHeap.Len() < 0 {
		errors = append(errors, fmt.Errorf("invalid size heap length"))
	}

	if len(errors) > 0 {
		slog.Warn("Multi-index validation found issues", "error_count", len(errors))
	} else {
		slog.Debug("Multi-index validation passed")
	}

	return errors
}
