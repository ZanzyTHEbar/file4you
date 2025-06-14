package deskfs

import (
	"context"
	"file4you/internal/filesystem/trees"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// ConcurrentTraverser implements high-performance concurrent directory traversal
// using worker pools and bounded goroutines for optimal resource utilization
type ConcurrentTraverser struct {
	maxWorkers    int
	jobQueue      chan DirectoryJob
	resultQueue   chan TraversalResult
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.RWMutex
	processedDirs map[string]bool // Track processed directories to avoid duplicates
}

// DirectoryJob represents a single directory traversal task
type DirectoryJob struct {
	Path      string
	Depth     int
	Parent    *trees.DirectoryNode
	MaxDepth  int
	Recursive bool
	Node      *trees.DirectoryNode
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
	mu             sync.RWMutex
}

// NewConcurrentTraverser creates a new concurrent directory traverser
// with optimal worker count based on available CPU cores
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
		jobQueue:      make(chan DirectoryJob, maxWorkers*4), // Buffer for smooth operation
		resultQueue:   make(chan TraversalResult, maxWorkers*4),
		ctx:           ctxWithCancel,
		cancel:        cancel,
		processedDirs: make(map[string]bool),
	}
}

// TraverseDirectory performs concurrent directory traversal with optimal performance
func (ct *ConcurrentTraverser) TraverseDirectory(rootPath string, recursive bool, maxDepth int, dfs *DesktopFS) (*trees.DirectoryNode, error) {
	// Initialize root node
	rootNode := trees.NewDirectoryNode(rootPath, nil)

	// Track performance metrics
	stats := &TraversalStats{
		StartTime: getCurrentTime(),
	}

	// Start error group for coordinated worker management
	g, ctx := errgroup.WithContext(ct.ctx)

	// Start worker pool
	for i := 0; i < ct.maxWorkers; i++ {
		workerID := i
		g.Go(func() error {
			return ct.worker(ctx, workerID, dfs, stats)
		})
	}

	// Start result collector
	var collectorErr error
	g.Go(func() error {
		collectorErr = ct.resultCollector(ctx, stats)
		return collectorErr
	})

	// Submit initial job
	initialJob := DirectoryJob{
		Path:      rootPath,
		Depth:     0,
		Parent:    nil,
		MaxDepth:  maxDepth,
		Recursive: recursive,
		Node:      rootNode,
	}

	select {
	case ct.jobQueue <- initialJob:
		// Job submitted successfully
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Signal completion and wait for all workers
	close(ct.jobQueue)

	if err := g.Wait(); err != nil && collectorErr == nil {
		return nil, fmt.Errorf("traversal failed: %w", err)
	}

	close(ct.resultQueue)

	stats.EndTime = getCurrentTime()
	ct.logPerformanceStats(stats)

	return rootNode, collectorErr
}

// worker processes directory jobs concurrently with proper error handling
func (ct *ConcurrentTraverser) worker(ctx context.Context, workerID int, dfs *DesktopFS, stats *TraversalStats) error {
	slog.Debug(fmt.Sprintf("Worker %d started", workerID))
	defer slog.Debug(fmt.Sprintf("Worker %d stopped", workerID))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case job, ok := <-ct.jobQueue:
			if !ok {
				return nil // Channel closed, worker should exit
			}

			result := ct.processDirectory(ctx, job, dfs)

			// Update statistics
			if result.Error == nil {
				stats.mu.Lock()
				stats.DirsProcessed++
				stats.FilesProcessed += int64(len(result.Files))
				stats.mu.Unlock()
			} else {
				stats.mu.Lock()
				stats.ErrorsFound++
				stats.mu.Unlock()
			}

			select {
			case ct.resultQueue <- result:
				// Result sent successfully
			case <-ctx.Done():
				return ctx.Err()
			}

			// Submit child directory jobs if recursive
			if job.Recursive && result.Error == nil {
				for _, childNode := range result.Children {
					if job.Depth < job.MaxDepth {
						childJob := DirectoryJob{
							Path:      childNode.Path,
							Depth:     job.Depth + 1,
							Parent:    job.Node,
							MaxDepth:  job.MaxDepth,
							Recursive: job.Recursive,
							Node:      childNode,
						}

						select {
						case ct.jobQueue <- childJob:
							// Child job submitted
						case <-ctx.Done():
							return ctx.Err()
						}
					}
				}
			}
		}
	}
}

// processDirectory processes a single directory with optimized I/O operations
func (ct *ConcurrentTraverser) processDirectory(ctx context.Context, job DirectoryJob, dfs *DesktopFS) TraversalResult {
	result := TraversalResult{
		Node: job.Node,
		Path: job.Path,
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		result.Error = ctx.Err()
		return result
	default:
	}

	// Check depth limits
	if job.Depth > job.MaxDepth {
		slog.Debug(fmt.Sprintf("Max depth %d reached at %s", job.MaxDepth, job.Path))
		return result
	}

	// Check if already processed (prevent duplicates)
	ct.mu.RLock()
	if ct.processedDirs[job.Path] {
		ct.mu.RUnlock()
		return result
	}
	ct.mu.RUnlock()

	// Mark as processed
	ct.mu.Lock()
	ct.processedDirs[job.Path] = true
	ct.mu.Unlock()

	// Read directory entries with optimized I/O
	entries, err := os.ReadDir(job.Path)
	if err != nil {
		result.Error = fmt.Errorf("failed to read directory %s: %w", job.Path, err)
		return result
	}

	// Get ignore patterns
	ignored, err := dfs.GetDesktopCleanerIgnore(job.Path)
	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to get ignore patterns for %s: %v", job.Path, err))
	}

	// Process entries with optimized allocation
	children := make([]*trees.DirectoryNode, 0, len(entries)/4) // Estimate 25% are directories
	files := make([]*trees.FileNode, 0, len(entries))

	for _, entry := range entries {
		childPath := filepath.Join(job.Path, entry.Name())

		// Skip ignored files and directories
		if ignored != nil && ignored.MatchesPath(childPath) {
			slog.Debug(fmt.Sprintf("Ignoring file %s", childPath))
			continue
		}

		if entry.IsDir() {
			childDir := trees.NewDirectoryNode(childPath, job.Node)
			children = append(children, childDir)
			job.Node.Children = append(job.Node.Children, childDir)
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
			job.Node.AddFile(childFile)
		}
	}

	result.Children = children
	result.Files = files
	return result
}

// resultCollector aggregates results from workers
func (ct *ConcurrentTraverser) resultCollector(ctx context.Context, stats *TraversalStats) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case result, ok := <-ct.resultQueue:
			if !ok {
				return nil // Channel closed
			}

			if result.Error != nil {
				slog.Error(fmt.Sprintf("Error processing directory %s: %v", result.Path, result.Error))
				// Continue processing other directories despite errors
			}
		}
	}
}

// logPerformanceStats logs traversal performance metrics
func (ct *ConcurrentTraverser) logPerformanceStats(stats *TraversalStats) {
	stats.mu.RLock()
	duration := stats.EndTime - stats.StartTime
	dirsProcessed := stats.DirsProcessed
	filesProcessed := stats.FilesProcessed
	errors := stats.ErrorsFound
	stats.mu.RUnlock()

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
}

// getCurrentTime returns current time in milliseconds for performance tracking
func getCurrentTime() int64 {
	return time.Now().UnixMilli()
}
