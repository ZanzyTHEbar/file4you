package trees

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"gonum.org/v1/gonum/spatial/kdtree"
)

// IncrementalKDTreeManager manages incremental updates to the KD-Tree to avoid O(N log N) rebuilds
type IncrementalKDTreeManager struct {
	mu                     sync.RWMutex
	rebalanceThreshold     int              // Trigger rebuild after this many insertions
	insertionsSinceRebuild int              // Count of insertions since last rebuild
	lastRebuildTime        time.Time        // Time of last rebuild
	pendingInsertions      []DirectoryPoint // Queue of points awaiting batch insertion
	batchSize              int              // Size of batch insertions
}

// NewIncrementalKDTreeManager creates a new incremental KD-Tree manager
func NewIncrementalKDTreeManager() *IncrementalKDTreeManager {
	return &IncrementalKDTreeManager{
		rebalanceThreshold: 100, // Rebuild after 100 insertions
		batchSize:          10,  // Batch insertions in groups of 10
		lastRebuildTime:    time.Now(),
		pendingInsertions:  make([]DirectoryPoint, 0, 10),
	}
}

// ShouldRebalance determines if the tree should be rebuilt for optimal performance
func (m *IncrementalKDTreeManager) ShouldRebalance() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Rebuild if we've hit the insertion threshold or after a time period
	timeSinceRebuild := time.Since(m.lastRebuildTime)
	return m.insertionsSinceRebuild >= m.rebalanceThreshold ||
		(m.insertionsSinceRebuild > 0 && timeSinceRebuild > 5*time.Minute)
}

// RecordInsertion tracks a new insertion for rebalance decision making
func (m *IncrementalKDTreeManager) RecordInsertion() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.insertionsSinceRebuild++
}

// RecordRebalance resets the insertion counter after a rebuild
func (m *IncrementalKDTreeManager) RecordRebalance() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.insertionsSinceRebuild = 0
	m.lastRebuildTime = time.Now()
}

// AddPendingInsertion queues a point for batch insertion
func (m *IncrementalKDTreeManager) AddPendingInsertion(point DirectoryPoint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pendingInsertions = append(m.pendingInsertions, point)
}

// FlushPendingInsertions returns and clears all pending insertions
func (m *IncrementalKDTreeManager) FlushPendingInsertions() []DirectoryPoint {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]DirectoryPoint, len(m.pendingInsertions))
	copy(result, m.pendingInsertions)
	m.pendingInsertions = m.pendingInsertions[:0] // Clear slice but keep capacity
	return result
}

// ShouldBatchInsert determines if we have enough pending insertions for a batch
func (m *IncrementalKDTreeManager) ShouldBatchInsert() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.pendingInsertions) >= m.batchSize
}

// BuildKDTree constructs the KD-Tree from the DirectoryTree's nodes.
// IMPORTANT: This K-D Tree implementation currently only indexes DirectoryNode objects
// (i.e., directories), not individual FileNode objects. This is due to the logic
// in collectDirectoryPoints which filters for node.Metadata.NodeType == Directory.
// Searches performed using this K-D Tree will therefore find directories that match
// the search criteria based on their metadata.
func (tree *DirectoryTree) BuildKDTree() {
	tree.mu.Lock()
	defer tree.mu.Unlock()

	// Populate KDTreeData with DirectoryPoints
	tree.KDTreeData = DirectoryPointCollection{}
	tree.collectDirectoryPoints(tree.Root)

	// Create KD-Tree with populated data
	tree.KDTree = kdtree.New(tree.KDTreeData, false)

	// Reset the incremental manager since we just did a full rebuild
	tree.kdManager.RecordRebalance()

	slog.Info("KD-Tree built successfully",
		"total_points", len(tree.KDTreeData),
		"rebuild_type", "full")
}

// BuildKDTreeIncremental performs an incremental update to the KD-Tree
// This method processes pending insertions and rebuilds only when necessary
func (tree *DirectoryTree) BuildKDTreeIncremental() {
	tree.mu.Lock()
	defer tree.mu.Unlock()

	// Check if we have pending insertions to process
	pending := tree.kdManager.FlushPendingInsertions()
	if len(pending) == 0 {
		return
	}

	// Add pending points to our data collection
	tree.KDTreeData = append(tree.KDTreeData, pending...)

	// Determine if we need a full rebuild or can continue with current tree
	if tree.kdManager.ShouldRebalance() {
		// Full rebuild for optimal balance
		tree.KDTree = kdtree.New(tree.KDTreeData, false)
		tree.kdManager.RecordRebalance()

		slog.Info("KD-Tree incrementally rebuilt",
			"total_points", len(tree.KDTreeData),
			"rebuild_type", "full_rebalance",
			"pending_processed", len(pending))
	} else {
		// Light rebuild - just recreate with new data but don't reset counters
		tree.KDTree = kdtree.New(tree.KDTreeData, false)

		slog.Debug("KD-Tree incrementally updated",
			"total_points", len(tree.KDTreeData),
			"rebuild_type", "incremental",
			"pending_processed", len(pending))
	}
}

// InsertNodeToKDTreeIncremental inserts a DirectoryNode into the KD-Tree using incremental approach
// This is the new high-performance version that avoids O(N log N) rebuilds
func (tree *DirectoryTree) InsertNodeToKDTreeIncremental(node *DirectoryNode) {
	// Validate that this is a directory node
	if node.Metadata.NodeType != Directory {
		slog.Debug("Skipping non-directory node for KD-Tree insertion", "path", node.Path)
		return
	}

	// Create a DirectoryPoint from node metadata
	metadataPoint, err := node.Metadata.ToKDTreePoint()
	if err != nil {
		slog.Error("Error converting metadata to KDTree point", "error", err, "path", node.Path)
		return
	}

	point := DirectoryPoint{
		Node:     node,
		Metadata: metadataPoint,
	}

	// Add to pending insertions for batch processing
	tree.kdManager.AddPendingInsertion(point)
	tree.kdManager.RecordInsertion()

	// Process batch if we have enough pending insertions
	if tree.kdManager.ShouldBatchInsert() {
		tree.BuildKDTreeIncremental()
	}

	slog.Debug("Directory node queued for incremental KD-Tree insertion",
		"path", node.Path,
		"pending_count", len(tree.kdManager.pendingInsertions))
}

// InsertNodeToKDTree inserts a DirectoryNode into the KD-Tree.
// DEPRECATED: This method performs O(N log N) rebuilds. Use InsertNodeToKDTreeIncremental instead.
// Note: This will only effectively insert directories due to the filtering
// in collectDirectoryPoints, which is the source of data for the tree.
func (tree *DirectoryTree) InsertNodeToKDTree(node *DirectoryNode) {
	slog.Warn("Using deprecated InsertNodeToKDTree - consider switching to InsertNodeToKDTreeIncremental for better performance")

	// Create a DirectoryPoint from node metadata and add it to the collection
	metadataPoint, err := node.Metadata.ToKDTreePoint()
	if err != nil {
		slog.Error(fmt.Sprintf("Error converting metadata to KDTree point: %v", err))
		return
	}

	point := DirectoryPoint{
		Node:     node,
		Metadata: metadataPoint,
	}
	tree.KDTreeData = append(tree.KDTreeData, point)

	// Rebuild the KD-Tree to include the new point.
	// PERFORMANCE NOTE: Rebuilding the entire K-D tree on every single insertion
	// is computationally expensive (O(N log N) where N is total points).
	// This approach is acceptable if insertions are infrequent or batched.
	// For frequent, individual insertions, use InsertNodeToKDTreeIncremental instead.
	tree.KDTree = kdtree.New(tree.KDTreeData, false)
}

// RangeSearchKDTree finds all DirectoryNode objects (directories) within a specified
// radius from the query DirectoryPoint.
func (tree *DirectoryTree) RangeSearchKDTree(query DirectoryPoint, radius float64) []*DirectoryNode {
	tree.mu.Lock()
	defer tree.mu.Unlock()

	// Process any pending incremental updates before searching
	tree.BuildKDTreeIncremental()

	if tree.KDTree == nil {
		slog.Warn("KD-Tree not initialized, returning empty results")
		return []*DirectoryNode{}
	}

	keeper := kdtree.NewDistKeeper(radius * radius) // Using squared distance for radius
	tree.KDTree.NearestSet(keeper, query)

	var results []*DirectoryNode
	for _, item := range keeper.Heap {
		dirPoint := item.Comparable.(DirectoryPoint)
		results = append(results, dirPoint.Node)
	}

	slog.Debug("Range search completed",
		"query_radius", radius,
		"results_count", len(results))

	return results
}

// NearestNeighborSearchKDTree finds the k nearest DirectoryNode objects (directories)
// to the query DirectoryPoint.
func (tree *DirectoryTree) NearestNeighborSearchKDTree(query DirectoryPoint, k int) []*DirectoryNode {
	tree.mu.Lock()
	defer tree.mu.Unlock()

	// Process any pending incremental updates before searching
	tree.BuildKDTreeIncremental()

	if tree.KDTree == nil {
		slog.Warn("KD-Tree not initialized, returning empty results")
		return []*DirectoryNode{}
	}

	keeper := kdtree.NewNKeeper(k)
	tree.KDTree.NearestSet(keeper, query)

	var results []*DirectoryNode
	for _, item := range keeper.Heap {
		dirPoint := item.Comparable.(DirectoryPoint)
		results = append(results, dirPoint.Node)
	}

	slog.Debug("Nearest neighbor search completed",
		"k", k,
		"results_count", len(results))

	return results
}

// collectDirectoryPoints recursively collects DirectoryPoints for KD-Tree construction.
// IMPORTANT: This function currently only creates DirectoryPoint entries for
// nodes where node.Metadata.NodeType == Directory. This means the K-D Tree
// will only contain points representing directories, not individual files.
func (tree *DirectoryTree) collectDirectoryPoints(node *DirectoryNode) {
	if node == nil {
		return
	}

	if err := node.Metadata.Validate(); err != nil {
		return
	}

	if node.Metadata.NodeType != Directory {
		return
	}

	metadataPoint, err := node.Metadata.ToKDTreePoint()
	if err != nil {
		slog.Error(fmt.Sprintf("Error converting metadata to KDTree point: %v", err))
		return
	}

	// Convert node metadata to DirectoryPoint and add to KDTreeData
	point := DirectoryPoint{
		Node:     node,
		Metadata: metadataPoint,
	}
	tree.KDTreeData = append(tree.KDTreeData, point)

	// Recursively add child directories
	for _, child := range node.Children {
		tree.collectDirectoryPoints(child)
	}
}

// FlushPendingKDTreeUpdates forces processing of all pending KD-Tree insertions
// This is useful before performing searches or when shutting down
func (tree *DirectoryTree) FlushPendingKDTreeUpdates() {
	tree.mu.Lock()
	defer tree.mu.Unlock()

	tree.BuildKDTreeIncremental()

	slog.Debug("KD-Tree pending updates flushed",
		"total_points", len(tree.KDTreeData))
}
