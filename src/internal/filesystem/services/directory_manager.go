package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"file4you/internal/filesystem/interfaces"
	"file4you/internal/filesystem/options"
	"file4you/internal/filesystem/types"
	"file4you/internal/trees"
)

// DirectoryManagerService provides high-performance directory management
// leveraging Phase 1 concurrent traversal and indexing systems
type DirectoryManagerService struct {
	traverser     ConcurrentTraverser
	directoryTree *trees.DirectoryTree
	mu            sync.RWMutex
	metrics       *DirectoryMetrics
}

// ConcurrentTraverser interface for dependency injection
type ConcurrentTraverser interface {
	TraverseDirectory(rootPath string, recursive bool, maxDepth int, handler TraversalHandler) (*trees.DirectoryNode, error)
	Cleanup()
}

// TraversalHandler defines the interface for handling traversal callbacks (legacy)
type TraversalHandler interface {
	GetDesktopCleanerIgnore(dir string) (IgnoreChecker, error)
	HandleDirectory(node *trees.DirectoryNode) error
	HandleFile(node *trees.FileNode) error
}

// IgnoreChecker interface for file ignore patterns
type IgnoreChecker interface {
	MatchesPath(path string) bool
}

// DirectoryMetrics tracks performance metrics for directory operations
type DirectoryMetrics struct {
	TotalTraversals  int64
	TotalFiles       int64
	TotalDirectories int64
	AverageTime      time.Duration
	LastOperation    time.Time
	mu               sync.RWMutex
}

// NewDirectoryManagerService creates a new directory management service
func NewDirectoryManagerService(traverser ConcurrentTraverser, dirTree *trees.DirectoryTree) *DirectoryManagerService {
	return &DirectoryManagerService{
		traverser:     traverser,
		directoryTree: dirTree,
		metrics:       &DirectoryMetrics{},
	}
}

// CalculateMaxDepth efficiently calculates directory depth using concurrent traversal
// Optimized replacement for the old filesystem.Walk approach
func (dm *DirectoryManagerService) CalculateMaxDepth(ctx context.Context, rootPath string) (int, error) {
	if rootPath == "" {
		return 0, fmt.Errorf("root path cannot be empty")
	}

	start := time.Now()
	defer func() {
		dm.updateMetrics(start)
	}()

	// Create a temporary traversal handler for depth calculation
	handler := &depthCalculationHandler{
		maxDepth: 0,
		rootPath: rootPath,
	}

	// Use concurrent traversal with unlimited depth to find actual max depth
	opts := options.DefaultTraversalOptions()
	opts.MaxDepth = 0    // Unlimited to find true max depth
	opts.WorkerCount = 2 // Limited workers for depth calculation

	node, err := dm.buildDirectoryTreeWithOptions(ctx, rootPath, opts, handler)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate max depth: %w", err)
	}

	maxDepth := dm.calculateNodeDepth(node, 0)

	slog.Info("Max depth calculated",
		"path", rootPath,
		"depth", maxDepth,
		"duration", time.Since(start))

	return maxDepth, nil
}

// IndexDirectory builds and indexes a directory structure using Phase 1 systems
func (dm *DirectoryManagerService) IndexDirectory(ctx context.Context, rootPath string, opts options.IndexOptions) error {
	start := time.Now()
	defer func() {
		dm.updateMetrics(start)
	}()

	slog.Info("Starting directory indexing",
		"path", rootPath,
		"recursive", opts.Recursive,
		"maxDepth", opts.MaxDepth,
		"workers", opts.WorkerCount)

	// Create traversal options from index options
	traversalOpts := options.TraversalOptions{
		Recursive:     opts.Recursive,
		MaxDepth:      opts.MaxDepth,
		IncludeHidden: opts.IncludeHidden,
		WorkerCount:   opts.WorkerCount,
		BufferSize:    opts.BatchSize,
	}

	// Build directory tree with concurrent traversal
	handler := &indexingHandler{
		ignorePatterns:   opts.IgnorePatterns,
		includeHidden:    opts.IncludeHidden,
		progressCallback: opts.ProgressCallback,
		processedCount:   0,
		totalEstimate:    0,
	}

	node, err := dm.buildDirectoryTreeWithOptions(ctx, rootPath, traversalOpts, handler)
	if err != nil {
		return fmt.Errorf("failed to index directory: %w", err)
	}

	// Update the main directory tree
	dm.mu.Lock()
	if dm.directoryTree == nil {
		dm.directoryTree = trees.NewDirectoryTree(trees.WithRoot(rootPath))
	}
	dm.directoryTree.Root = node
	dm.mu.Unlock()

	// Build indexes if specified
	if err := dm.buildIndexes(ctx, opts.IndexTypes, opts.BatchSize); err != nil {
		return fmt.Errorf("failed to build indexes: %w", err)
	}

	slog.Info("Directory indexing completed",
		"path", rootPath,
		"duration", time.Since(start),
		"processed", handler.processedCount)

	return nil
}

// BuildDirectoryTree creates a directory tree structure using concurrent traversal
func (dm *DirectoryManagerService) BuildDirectoryTree(ctx context.Context, rootPath string, opts options.TraversalOptions) (*trees.DirectoryNode, error) {
	start := time.Now()
	defer func() {
		dm.updateMetrics(start)
	}()

	handler := &treeBuilderHandler{
		filter:        opts.Filter,
		includeHidden: opts.IncludeHidden,
	}

	node, err := dm.buildDirectoryTreeWithOptions(ctx, rootPath, opts, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to build directory tree: %w", err)
	}

	slog.Info("Directory tree built",
		"path", rootPath,
		"duration", time.Since(start))

	return node, nil
}

// AnalyzeDirectory performs comprehensive directory analysis
func (dm *DirectoryManagerService) AnalyzeDirectory(ctx context.Context, rootPath string) (*types.DirectoryAnalysis, error) {
	start := time.Now()

	handler := &analysisHandler{
		analysis: &types.DirectoryAnalysis{
			FileTypes:        make(map[string]int),
			SizeDistribution: make(map[string]int),
			AgeDistribution:  make(map[string]int),
		},
		rootPath: rootPath,
	}

	opts := options.DefaultTraversalOptions()
	opts.WorkerCount = 4 // Use more workers for analysis

	node, err := dm.buildDirectoryTreeWithOptions(ctx, rootPath, opts, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze directory: %w", err)
	}

	// Finalize analysis
	analysis := handler.analysis
	analysis.Duration = time.Since(start)
	analysis.MaxDepth = dm.calculateNodeDepth(node, 0)

	// Sort files by criteria
	dm.sortFilesBySize(node, analysis)
	dm.sortFilesByAge(node, analysis)

	slog.Info("Directory analysis completed",
		"path", rootPath,
		"files", analysis.TotalFiles,
		"directories", analysis.TotalDirectories,
		"size", analysis.TotalSize,
		"duration", analysis.Duration)

	return analysis, nil
}

// BuildDirectoryTreeWithAnalysis creates a directory tree structure and performs analysis in one pass
func (dm *DirectoryManagerService) BuildDirectoryTreeWithAnalysis(ctx context.Context, rootPath string, opts options.TraversalOptions) (*trees.DirectoryNode, *types.DirectoryAnalysis, error) {
	start := time.Now()
	defer func() {
		dm.updateMetrics(start)
	}()

	// Create combined handler that does both analysis and tree building
	handler := &combinedHandler{
		analysis: &types.DirectoryAnalysis{
			FileTypes:        make(map[string]int),
			SizeDistribution: make(map[string]int),
			AgeDistribution:  make(map[string]int),
		},
		rootPath:      rootPath,
		filter:        opts.Filter,
		includeHidden: opts.IncludeHidden,
	}

	node, err := dm.buildDirectoryTreeWithOptions(ctx, rootPath, opts, handler)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build directory tree with analysis: %w", err)
	}

	// Finalize analysis
	analysis := handler.analysis
	analysis.Duration = time.Since(start)
	analysis.MaxDepth = dm.calculateNodeDepth(node, 0)

	slog.Info("Directory tree built with analysis",
		"path", rootPath,
		"files", analysis.TotalFiles,
		"directories", analysis.TotalDirectories,
		"size", analysis.TotalSize,
		"duration", analysis.Duration)

	return node, analysis, nil
}

// buildDirectoryTreeWithOptions is the core method that uses concurrent traversal
func (dm *DirectoryManagerService) buildDirectoryTreeWithOptions(ctx context.Context, rootPath string, opts options.TraversalOptions, handler traversalHandlerAdapter) (*trees.DirectoryNode, error) {
	// Convert our handler to the interface expected by ConcurrentTraverser
	adaptedHandler := &handlerAdapter{
		original: handler,
		ctx:      ctx,
	}

	node, err := dm.traverser.TraverseDirectory(
		rootPath,
		opts.Recursive,
		opts.MaxDepth,
		adaptedHandler,
	)

	if err != nil {
		return nil, fmt.Errorf("concurrent traversal failed: %w", err)
	}

	// Post-process with sorting if required
	if opts.SortBy.Field != "" {
		dm.sortNode(node, opts.SortBy)
	}

	return node, nil
}

// buildIndexes builds the specified index types
func (dm *DirectoryManagerService) buildIndexes(ctx context.Context, indexTypes []string, batchSize int) error {
	if dm.directoryTree == nil {
		return fmt.Errorf("directory tree not available for indexing")
	}

	for _, indexType := range indexTypes {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		switch indexType {
		case "path":
			// Path index is built automatically during tree construction
			slog.Debug("Path index ready", "batchSize", batchSize)
		case "spatial":
			// Spatial index is built automatically during tree construction
			slog.Debug("Spatial index ready", "batchSize", batchSize)
		case "multi":
			// Multi-index is built automatically during tree construction
			slog.Debug("Multi-index ready", "batchSize", batchSize)
		default:
			slog.Warn("Unknown index type", "type", indexType)
		}
	}

	return nil
}

// calculateNodeDepth recursively calculates the maximum depth of a directory tree
func (dm *DirectoryManagerService) calculateNodeDepth(node *trees.DirectoryNode, currentDepth int) int {
	if node == nil {
		return currentDepth
	}

	maxDepth := currentDepth
	for _, child := range node.Children {
		childDepth := dm.calculateNodeDepth(child, currentDepth+1)
		if childDepth > maxDepth {
			maxDepth = childDepth
		}
	}

	return maxDepth
}

// updateMetrics updates performance metrics
func (dm *DirectoryManagerService) updateMetrics(start time.Time) {
	dm.metrics.mu.Lock()
	defer dm.metrics.mu.Unlock()

	dm.metrics.TotalTraversals++
	duration := time.Since(start)

	// Calculate rolling average
	if dm.metrics.TotalTraversals == 1 {
		dm.metrics.AverageTime = duration
	} else {
		dm.metrics.AverageTime = (dm.metrics.AverageTime*time.Duration(dm.metrics.TotalTraversals-1) + duration) / time.Duration(dm.metrics.TotalTraversals)
	}

	dm.metrics.LastOperation = time.Now()
}

// Interface implementations and handlers follow...

// depthCalculationHandler handles depth calculation during traversal
// Handler implementations for different traversal needs
type depthCalculationHandler struct {
	maxDepth int64
	rootPath string
	mu       sync.Mutex
}

func (h *depthCalculationHandler) HandleDirectory(node *trees.DirectoryNode) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Calculate depth from root path
	relPath, _ := filepath.Rel(h.rootPath, node.Path)
	depth := int64(strings.Count(relPath, string(os.PathSeparator)))
	if depth > h.maxDepth {
		h.maxDepth = depth
	}

	return nil
}

func (h *depthCalculationHandler) HandleFile(node *trees.FileNode) error {
	// Files don't affect directory depth calculation
	return nil
}

// indexingHandler handles directory indexing during traversal
type indexingHandler struct {
	ignorePatterns   []string
	includeHidden    bool
	progressCallback func(current, total int)
	processedCount   int64
	totalEstimate    int64
}

func (h *indexingHandler) HandleDirectory(node *trees.DirectoryNode) error {
	dirName := filepath.Base(node.Path)
	if !h.includeHidden && strings.HasPrefix(dirName, ".") {
		return fmt.Errorf("skip hidden directory: %s", dirName)
	}

	// Check ignore patterns
	for _, pattern := range h.ignorePatterns {
		if matched, _ := filepath.Match(pattern, dirName); matched {
			return fmt.Errorf("skip ignored directory: %s", dirName)
		}
	}

	atomic.AddInt64(&h.processedCount, 1)
	if h.progressCallback != nil {
		h.progressCallback(int(atomic.LoadInt64(&h.processedCount)), int(atomic.LoadInt64(&h.totalEstimate)))
	}

	return nil
}

func (h *indexingHandler) HandleFile(node *trees.FileNode) error {
	if !h.includeHidden && strings.HasPrefix(node.Name, ".") {
		return nil // Skip hidden files silently
	}

	// Check ignore patterns
	for _, pattern := range h.ignorePatterns {
		if matched, _ := filepath.Match(pattern, node.Name); matched {
			return nil // Skip ignored files silently
		}
	}

	atomic.AddInt64(&h.processedCount, 1)
	if h.progressCallback != nil {
		h.progressCallback(int(atomic.LoadInt64(&h.processedCount)), int(atomic.LoadInt64(&h.totalEstimate)))
	}

	return nil
}

// treeBuilderHandler handles tree building during traversal
type treeBuilderHandler struct {
	filter        func(*trees.FileNode) bool
	includeHidden bool
}

func (h *treeBuilderHandler) HandleDirectory(node *trees.DirectoryNode) error {
	dirName := filepath.Base(node.Path)
	if !h.includeHidden && strings.HasPrefix(dirName, ".") {
		return fmt.Errorf("skip hidden directory: %s", dirName)
	}

	return nil
}

func (h *treeBuilderHandler) HandleFile(node *trees.FileNode) error {
	if !h.includeHidden && strings.HasPrefix(node.Name, ".") {
		return nil // Skip hidden files silently
	}

	if h.filter != nil && !h.filter(node) {
		return nil // Skip filtered files silently
	}

	return nil
}

// analysisHandler handles directory analysis during traversal
type analysisHandler struct {
	analysis *types.DirectoryAnalysis
	rootPath string
	mu       sync.Mutex
}

func (h *analysisHandler) HandleDirectory(node *trees.DirectoryNode) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.analysis.TotalDirectories++

	// Calculate depth from root path
	relPath, _ := filepath.Rel(h.rootPath, node.Path)
	depth := strings.Count(relPath, string(os.PathSeparator))
	if depth > h.analysis.MaxDepth {
		h.analysis.MaxDepth = depth
	}

	return nil
}

func (h *analysisHandler) HandleFile(node *trees.FileNode) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.analysis.TotalFiles++
	h.analysis.TotalSize += node.Metadata.Size

	// File type analysis
	ext := strings.ToLower(filepath.Ext(node.Name))
	if ext == "" {
		ext = "no_extension"
	}
	h.analysis.FileTypes[ext]++

	// Size distribution
	var sizeCategory string
	switch {
	case node.Metadata.Size < 1024:
		sizeCategory = "small_1kb"
	case node.Metadata.Size < 1024*1024:
		sizeCategory = "medium_1mb"
	case node.Metadata.Size < 1024*1024*100:
		sizeCategory = "large_100mb"
	default:
		sizeCategory = "very_large"
	}
	h.analysis.SizeDistribution[sizeCategory]++

	// Age distribution
	age := time.Since(node.Metadata.ModifiedAt)
	var ageCategory string
	switch {
	case age < 24*time.Hour:
		ageCategory = "today"
	case age < 7*24*time.Hour:
		ageCategory = "week"
	case age < 30*24*time.Hour:
		ageCategory = "month"
	case age < 365*24*time.Hour:
		ageCategory = "year"
	default:
		ageCategory = "older"
	}
	h.analysis.AgeDistribution[ageCategory]++

	// Track largest files
	if len(h.analysis.LargestFiles) < 10 {
		h.analysis.LargestFiles = append(h.analysis.LargestFiles, node)
	} else {
		// Find the smallest of the largest files and replace if current is larger
		minIdx := 0
		for i, f := range h.analysis.LargestFiles {
			if f.Metadata.Size < h.analysis.LargestFiles[minIdx].Metadata.Size {
				minIdx = i
			}
		}
		if node.Metadata.Size > h.analysis.LargestFiles[minIdx].Metadata.Size {
			h.analysis.LargestFiles[minIdx] = node
		}
	}

	return nil
}

// combinedHandler handles both analysis and tree building in one pass
type combinedHandler struct {
	analysis      *types.DirectoryAnalysis
	rootPath      string
	filter        func(*trees.FileNode) bool
	includeHidden bool
	mu            sync.Mutex
}

func (h *combinedHandler) HandleDirectory(node *trees.DirectoryNode) error {
	// Apply directory filtering (skip hidden if needed)
	dirName := filepath.Base(node.Path)
	if !h.includeHidden && strings.HasPrefix(dirName, ".") {
		return fmt.Errorf("skip hidden directory: %s", dirName)
	}

	// Update analysis
	h.mu.Lock()
	defer h.mu.Unlock()
	h.analysis.TotalDirectories++

	// Calculate depth from root path
	relPath, _ := filepath.Rel(h.rootPath, node.Path)
	depth := strings.Count(relPath, string(os.PathSeparator))
	if depth > h.analysis.MaxDepth {
		h.analysis.MaxDepth = depth
	}

	return nil
}

func (h *combinedHandler) HandleFile(node *trees.FileNode) error {
	// Apply file filtering (skip hidden if needed)
	if !h.includeHidden && strings.HasPrefix(node.Name, ".") {
		return nil // Skip hidden files silently
	}

	if h.filter != nil && !h.filter(node) {
		return nil // Skip filtered files silently
	}

	// Update analysis
	h.mu.Lock()
	defer h.mu.Unlock()
	h.analysis.TotalFiles++
	h.analysis.TotalSize += node.Metadata.Size

	// File type analysis
	ext := strings.ToLower(filepath.Ext(node.Name))
	if ext == "" {
		ext = "no_extension"
	}
	h.analysis.FileTypes[ext]++

	// Size distribution
	var sizeCategory string
	switch {
	case node.Metadata.Size < 1024:
		sizeCategory = "small_1kb"
	case node.Metadata.Size < 1024*1024:
		sizeCategory = "medium_1mb"
	case node.Metadata.Size < 1024*1024*100:
		sizeCategory = "large_100mb"
	default:
		sizeCategory = "huge_100mb_plus"
	}
	h.analysis.SizeDistribution[sizeCategory]++

	// Age distribution (simple example)
	modTime := node.Metadata.ModifiedAt
	if !modTime.IsZero() {
		ageDays := int(time.Since(modTime).Hours() / 24)
		var ageCategory string
		switch {
		case ageDays < 7:
			ageCategory = "recent_week"
		case ageDays < 30:
			ageCategory = "recent_month"
		case ageDays < 365:
			ageCategory = "recent_year"
		default:
			ageCategory = "old_year_plus"
		}
		h.analysis.AgeDistribution[ageCategory]++
	}

	return nil
}

// handlerAdapter adapts our handlers to the ConcurrentTraverser interface
type handlerAdapter struct {
	original traversalHandlerAdapter
	ctx      context.Context
}

// traversalHandlerAdapter defines the interface for our traversal handlers
type traversalHandlerAdapter interface {
	HandleDirectory(node *trees.DirectoryNode) error
	HandleFile(node *trees.FileNode) error
}

// Implementation of handler methods...
func (h *handlerAdapter) GetDesktopCleanerIgnore(dir string) (IgnoreChecker, error) {
	// TODO: Return a null ignore checker for now - can be enhanced later
	return &nullIgnoreChecker{}, nil
}

func (h *handlerAdapter) HandleDirectory(node *trees.DirectoryNode) error {
	return h.original.HandleDirectory(node)
}

func (h *handlerAdapter) HandleFile(node *trees.FileNode) error {
	return h.original.HandleFile(node)
}

type nullIgnoreChecker struct{}

func (n *nullIgnoreChecker) MatchesPath(path string) bool {
	return false
}

// Sorting methods for analysis
func (dm *DirectoryManagerService) sortFilesBySize(node *trees.DirectoryNode, analysis *types.DirectoryAnalysis) {
	// TODO: Implementation for sorting files by size for analysis
	// This would traverse the tree and extract largest files
}

func (dm *DirectoryManagerService) sortFilesByAge(node *trees.DirectoryNode, analysis *types.DirectoryAnalysis) {
	// TODO: Implementation for sorting files by age for analysis
	// This would traverse the tree and extract oldest/newest files
}

func (dm *DirectoryManagerService) sortNode(node *trees.DirectoryNode, criteria options.SortCriteria) {
	// TODO: Implementation for sorting directory contents based on criteria
}

// Ensure DirectoryManagerService implements the interface
var _ interfaces.DirectoryService = (*DirectoryManagerService)(nil)
