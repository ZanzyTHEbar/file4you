package trees

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector provides metrics collection functionality for tree operations
// Added mu for locking during OperationCounts updates.
type MetricsCollector struct {
	mu      sync.Mutex
	metrics atomic.Value // stores *TreeMetrics
	started time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		started: time.Now(),
	}
	mc.metrics.Store(&TreeMetrics{
		OperationCounts: make(map[string]int64),
	})
	return mc
}

// IncrementOperation safely increments operation count using mutex locking
func (mc *MetricsCollector) IncrementOperation(op string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	metrics := mc.metrics.Load().(*TreeMetrics)
	metrics.OperationCounts[op]++
	mc.metrics.Store(metrics)
}

// computeTreeMetrics recursively computes metrics starting from the given directory node.
func computeTreeMetrics(node *DirectoryNode, depth int, metrics *TreeMetrics) {
	if node == nil {
		return
	}

	// Include the directory node
	metrics.TotalNodes++
	metrics.TotalSize += node.Metadata.Size
	if depth > metrics.MaxDepth {
		metrics.MaxDepth = depth
	}

	// Process files as part of metrics
	for _, file := range node.Files {
		metrics.TotalNodes++
		metrics.TotalSize += file.Metadata.Size
	}

	// Process child directories
	for _, child := range node.Children {
		computeTreeMetrics(child, depth+1, metrics)
	}
}

// UpdateMetrics updates tree metrics based on the current state of the DirectoryTree.
func (mc *MetricsCollector) UpdateMetrics(ctx context.Context, tree *DirectoryTree) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create a new metrics snapshot
	newMetrics := &TreeMetrics{
		OperationCounts: make(map[string]int64),
		LastUpdated:     time.Now(),
	}

	// Compute metrics by traversing the directory tree
	computeTreeMetrics(tree.Root, 0, newMetrics)

	// Update processing time
	newMetrics.ProcessingTime = time.Since(mc.started) // assuming mc.started is tree start time

	// Store the new metrics
	mc.metrics.Store(newMetrics)

	return nil
}
