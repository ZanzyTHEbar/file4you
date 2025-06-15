package trees

import (
	"context"
	"io"
	"time"
)

// TreeNode represents any node in a tree structure
type TreeNode interface {
	GetPath() string
	GetName() string
	GetMetadata() *Metadata
	IsDirectory() bool
}

// TreeWalker defines operations for traversing tree structures
type TreeWalker interface {
	Walk(ctx context.Context) error
	Filter(predicate func(TreeNode) bool) []TreeNode
	ForEach(fn func(TreeNode) error) error
}

// MetricsCollector defines operations for collecting tree metrics
type MetricsCollectorI interface {
	CollectMetrics(ctx context.Context) (*TreeMetrics, error)
	ResetMetrics()
}

// TreeMetrics holds statistical information about the tree
type TreeMetrics struct {
	TotalNodes      int64
	TotalSize       int64
	MaxDepth        int
	LastUpdated     time.Time
	ProcessingTime  time.Duration
	OperationCounts map[string]int64
}

// Cleanable defines cleanup operations
type Cleanable interface {
	Cleanup() error
	io.Closer
}
