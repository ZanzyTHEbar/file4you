package filesystem

import (
	"context"
	"file4you/internal/filesystem/services"
	"file4you/internal/trees"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/conc"
	"github.com/sourcegraph/conc/pool"
)

// ConcurrentTraverser implements high-performance concurrent directory traversal
// using the conc package for robust worker pool and job management
type ConcurrentTraverser struct {
	maxWorkers    int
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
	processedDirs map[string]bool // Track processed directories to avoid duplicates
	pool          *pool.ContextPool
}

// TraversalResult contains the result of processing a directory
type TraversalResult struct {
	Node     *trees.DirectoryNode
	Children []*trees.DirectoryNode
	Files    []*trees.FileNode
	Error    error
	Path     string
}

// TraversalStats tracks performance metrics during traversal
type TraversalStats struct {
	DirsProcessed  int64
	FilesProcessed int64
	ErrorsFound    int64
	StartTime      int64
	EndTime        int64
}

// NewConcurrentTraverser creates a new concurrent directory traverser
// with optimal worker count based on available CPU cores using conc.Pool
func NewConcurrentTraverser(ctx context.Context) *ConcurrentTraverser {
	// Optimal worker count: CPU cores * 2 for I/O bound operations
	maxWorkers := runtime.NumCPU() * 2
	if maxWorkers < 4 {
		maxWorkers = 4 // Minimum workers for responsiveness
	}
	if maxWorkers > 32 {
		maxWorkers = 32 // Maximum to prevent resource exhaustion
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)

	return &ConcurrentTraverser{
		maxWorkers:    maxWorkers,
		ctx:           ctxWithCancel,
		cancel:        cancel,
		processedDirs: make(map[string]bool),
		pool:          pool.New().WithMaxGoroutines(maxWorkers).WithContext(ctxWithCancel),
	}
}

// TraverseDirectory performs concurrent directory traversal using conc.Pool
// for optimal performance and resource management
func (ct *ConcurrentTraverser) TraverseDirectory(rootPath string, recursive bool, maxDepth int, handler services.TraversalHandler) (*trees.DirectoryNode, error) {
	// Initialize root node
	rootNode := trees.NewDirectoryNode(rootPath, nil)

	// Track performance metrics with atomic operations
	stats := &TraversalStats{
		StartTime: getCurrentTime(),
	}

	// Use conc.WaitGroup for better error handling and coordination
	wg := conc.NewWaitGroup()

	// Process directories level by level using a BFS approach with conc.Pool
	currentLevel := []*trees.DirectoryNode{rootNode}

	for depth := 0; (maxDepth == -1 || depth <= maxDepth) && len(currentLevel) > 0; depth++ {
		if !recursive && depth > 0 {
			break
		}

		nextLevel := make([]*trees.DirectoryNode, 0)
		var nextLevelMu sync.Mutex

		// Process all directories at the current level concurrently
		for _, dirNode := range currentLevel {
			dirNode := dirNode // Capture loop variable
			wg.Go(func() {
				result := ct.processDirectoryNode(ct.ctx, dirNode, depth, maxDepth, handler)

				// Update statistics atomically
				if result.Error == nil {
					atomic.AddInt64(&stats.DirsProcessed, 1)
					atomic.AddInt64(&stats.FilesProcessed, int64(len(result.Files)))
				} else {
					atomic.AddInt64(&stats.ErrorsFound, 1)
					slog.Error(fmt.Sprintf("Error processing directory %s: %v", result.Path, result.Error))
				}

				// Add child directories to next level if within depth limits
				if recursive && (maxDepth == -1 || depth < maxDepth) && result.Error == nil {
					nextLevelMu.Lock()
					nextLevel = append(nextLevel, result.Children...)
					nextLevelMu.Unlock()
				}
			})
		}

		// Wait for all directories at this level to complete
		wg.Wait()

		// Move to the next level
		currentLevel = nextLevel
	}

	stats.EndTime = getCurrentTime()
	ct.logPerformanceStats(stats)

	return rootNode, nil
}

// processDirectoryNode processes a single directory node with optimized I/O operations
func (ct *ConcurrentTraverser) processDirectoryNode(ctx context.Context, dirNode *trees.DirectoryNode, depth, maxDepth int, handler services.TraversalHandler) TraversalResult {
	result := TraversalResult{
		Node: dirNode,
		Path: dirNode.Path,
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		result.Error = ctx.Err()
		return result
	default:
	}

	// Check depth limits (-1 means unlimited)
	if maxDepth != -1 && depth > maxDepth {
		slog.Debug(fmt.Sprintf("Max depth %d reached at %s", maxDepth, dirNode.Path))
		return result
	}

	// Check if already processed (prevent duplicates)
	ct.mu.RLock()
	if ct.processedDirs[dirNode.Path] {
		ct.mu.RUnlock()
		return result
	}
	ct.mu.RUnlock()

	// Mark as processed
	ct.mu.Lock()
	ct.processedDirs[dirNode.Path] = true
	ct.mu.Unlock()

	// Read directory entries with optimized I/O
	entries, err := os.ReadDir(dirNode.Path)
	if err != nil {
		result.Error = fmt.Errorf("failed to read directory %s: %w", dirNode.Path, err)
		return result
	}

	// Get ignore patterns
	ignored, err := handler.GetDesktopCleanerIgnore(dirNode.Path)
	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to get ignore patterns for %s: %v", dirNode.Path, err))
	}

	// Process entries with optimized allocation
	children := make([]*trees.DirectoryNode, 0, len(entries)/4) // Estimate 25% are directories
	files := make([]*trees.FileNode, 0, len(entries))

	for _, entry := range entries {
		childPath := filepath.Join(dirNode.Path, entry.Name())

		// Skip ignored files and directories
		if ignored != nil && ignored.MatchesPath(childPath) {
			slog.Debug(fmt.Sprintf("Ignoring file %s", childPath))
			continue
		}

		if entry.IsDir() {
			childDir := trees.NewDirectoryNode(childPath, dirNode)
			children = append(children, childDir)
			dirNode.Children = append(dirNode.Children, childDir)

			// Call handler for directory
			if err := handler.HandleDirectory(childDir); err != nil {
				slog.Warn(fmt.Sprintf("Handler error for directory %s: %v", childPath, err))
			}
		} else {
			entryInfo, err := entry.Info()
			if err != nil {
				slog.Warn(fmt.Sprintf("Error getting file info for %s: %v", entry.Name(), err))
				continue
			}

			childFile := &trees.FileNode{
				Path:      childPath,
				Name:      entry.Name(),
				Extension: strings.ToLower(filepath.Ext(entry.Name())),
				Metadata:  trees.NewMetadata(entryInfo),
			}
			files = append(files, childFile)
			dirNode.AddFile(childFile)

			// Call handler for file
			if err := handler.HandleFile(childFile); err != nil {
				slog.Warn(fmt.Sprintf("Handler error for file %s: %v", childPath, err))
			}
		}
	}

	result.Children = children
	result.Files = files
	return result
}

// TraverseDirectoryWithPool performs concurrent directory traversal using conc.Pool directly
// This is an alternative implementation that leverages the pool for fine-grained control
func (ct *ConcurrentTraverser) TraverseDirectoryWithPool(rootPath string, recursive bool, maxDepth int, handler services.TraversalHandler) (*trees.DirectoryNode, error) {
	// Initialize root node
	rootNode := trees.NewDirectoryNode(rootPath, nil)

	// Track performance metrics with atomic operations
	stats := &TraversalStats{
		StartTime: getCurrentTime(),
	}

	// Create a work queue for directories to process
	type dirWork struct {
		node     *trees.DirectoryNode
		depth    int
		maxDepth int
	}

	workQueue := make(chan dirWork, ct.maxWorkers*4)
	var pendingWork int64 = 1 // Start with root directory

	// Add initial work
	workQueue <- dirWork{node: rootNode, depth: 0, maxDepth: maxDepth}

	// Process work using the pool
	for atomic.LoadInt64(&pendingWork) > 0 {
		select {
		case <-ct.ctx.Done():
			return nil, ct.ctx.Err()
		case work := <-workQueue:
			atomic.AddInt64(&pendingWork, -1)

			ct.pool.Go(func(ctx context.Context) error {
				result := ct.processDirectoryNode(ctx, work.node, work.depth, work.maxDepth, handler)

				// Update statistics atomically
				if result.Error == nil {
					atomic.AddInt64(&stats.DirsProcessed, 1)
					atomic.AddInt64(&stats.FilesProcessed, int64(len(result.Files)))

					// Add child directories to work queue if recursive and within depth
					if recursive && work.depth < work.maxDepth {
						for _, child := range result.Children {
							atomic.AddInt64(&pendingWork, 1)
							select {
							case workQueue <- dirWork{node: child, depth: work.depth + 1, maxDepth: work.maxDepth}:
								// Work submitted successfully
							case <-ctx.Done():
								atomic.AddInt64(&pendingWork, -1)
								return ctx.Err()
							}
						}
					}
				} else {
					atomic.AddInt64(&stats.ErrorsFound, 1)
					slog.Error(fmt.Sprintf("Error processing directory %s: %v", result.Path, result.Error))
				}

				return nil
			})
		}
	}

	// Wait for all work to complete
	ct.pool.Wait()

	stats.EndTime = getCurrentTime()
	ct.logPerformanceStats(stats)

	return rootNode, nil
}

// logPerformanceStats logs traversal performance metrics
func (ct *ConcurrentTraverser) logPerformanceStats(stats *TraversalStats) {
	duration := stats.EndTime - stats.StartTime
	dirsProcessed := atomic.LoadInt64(&stats.DirsProcessed)
	filesProcessed := atomic.LoadInt64(&stats.FilesProcessed)
	errors := atomic.LoadInt64(&stats.ErrorsFound)

	if duration > 0 {
		dirsPerSec := float64(dirsProcessed) / float64(duration) * 1000 // Convert to per second
		filesPerSec := float64(filesProcessed) / float64(duration) * 1000

		slog.Info(fmt.Sprintf("Traversal completed: %d dirs, %d files in %dms (%.1f dirs/sec, %.1f files/sec, %d errors)",
			dirsProcessed, filesProcessed, duration, dirsPerSec, filesPerSec, errors))
	}
}

// Cleanup releases resources used by the traverser
func (ct *ConcurrentTraverser) Cleanup() {
	if ct.cancel != nil {
		ct.cancel()
	}
	if ct.pool != nil {
		ct.pool.Wait() // Ensure all goroutines complete
	}
}

// getCurrentTime returns current time in milliseconds for performance tracking
func getCurrentTime() int64 {
	return time.Now().UnixMilli()
}
