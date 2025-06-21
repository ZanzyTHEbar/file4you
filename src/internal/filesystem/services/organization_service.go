package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"file4you/internal/filesystem/interfaces"
	"file4you/internal/filesystem/options"
	"file4you/internal/filesystem/types"
	"file4you/internal/trees"
)

// OrganizationService handles file organization logic with rule-based and future AI categorization
type OrganizationService struct {
	conflictMgr   interfaces.ConflictResolver
	fileOps       interfaces.FileOperations
	directoryMgr  interfaces.DirectoryService
	eventHandlers []func(types.Event)
	mu            sync.RWMutex
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(
	conflictMgr interfaces.ConflictResolver,
	fileOps interfaces.FileOperations,
	directoryMgr interfaces.DirectoryService,
) *OrganizationService {
	return &OrganizationService{
		conflictMgr:   conflictMgr,
		fileOps:       fileOps,
		directoryMgr:  directoryMgr,
		eventHandlers: make([]func(types.Event), 0),
	}
}

// OrganizeDirectory performs comprehensive directory organization using AI-powered categorization
func (ors *OrganizationService) OrganizeDirectory(ctx context.Context, sourcePath, targetPath string, opts options.OrganizationOptions) (*types.OrganizationResult, error) {
	start := time.Now()

	slog.Info("Starting directory organization",
		"source", sourcePath,
		"target", targetPath,
		"aiEnabled", opts.UseAI,
		"dryRun", opts.DryRun)

	result := &types.OrganizationResult{
		StartTime:      start,
		SourcePath:     sourcePath,
		TargetPath:     targetPath,
		DryRun:         opts.DryRun,
		ProcessedFiles: make([]types.FileOperation, 0),
		Conflicts:      make([]types.ConflictInfo, 0),
		Events:         make([]types.Event, 0),
	}

	// Emit start event
	ors.emitEvent(types.Event{
		Type:      types.EventOperationStart,
		Timestamp: start,
		Path:      sourcePath,
		Operation: "organize_directory",
		Success:   true,
		Metadata: map[string]interface{}{
			"target_path": targetPath,
			"dry_run":     opts.DryRun,
			"use_ai":      opts.UseAI,
		},
	})

	// Analyze source directory and build tree in one pass
	traversalOpts := options.TraversalOptions{
		Recursive:     true,
		MaxDepth:      opts.MaxDepth,
		IncludeHidden: opts.IncludeHidden,
		WorkerCount:   opts.WorkerCount,
		BufferSize:    opts.BatchSize,
	}

	rootNode, analysis, err := ors.directoryMgr.BuildDirectoryTreeWithAnalysis(ctx, sourcePath, traversalOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze and build directory tree: %w", err)
	}

	result.SourceAnalysis = analysis

	// Process files concurrently
	if err := ors.processDirectoryTree(ctx, rootNode, targetPath, opts, result); err != nil {
		return nil, fmt.Errorf("failed to process directory tree: %w", err)
	}

	// Finalize result
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = true

	// Emit completion event
	ors.emitEvent(types.Event{
		Type:      types.EventOperationEnd,
		Timestamp: result.EndTime,
		Path:      sourcePath,
		Operation: "organize_directory",
		Success:   true,
		Metadata: map[string]interface{}{
			"processed_files": len(result.ProcessedFiles),
			"conflicts":       len(result.Conflicts),
			"duration":        result.Duration.String(),
		},
	})

	slog.Info("Directory organization completed",
		"duration", result.Duration,
		"processed", len(result.ProcessedFiles),
		"conflicts", len(result.Conflicts))

	return result, nil
}

// processDirectoryTree recursively processes the directory tree for organization
func (ors *OrganizationService) processDirectoryTree(ctx context.Context, node *trees.DirectoryNode, targetBase string, opts options.OrganizationOptions, result *types.OrganizationResult) error {
	// Create worker pool for concurrent file processing
	semaphore := make(chan struct{}, opts.WorkerCount)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Process files in current directory
	for _, file := range node.Files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case semaphore <- struct{}{}:
			wg.Add(1)
			go func(f *trees.FileNode) {
				defer func() {
					<-semaphore
					wg.Done()
				}()

				if err := ors.processFile(ctx, f, targetBase, opts, &mu, result); err != nil {
					slog.Error("Failed to process file", "file", f.Path, "error", err)
				}
			}(file)
		}
	}

	// Process subdirectories recursively
	for _, child := range node.Children {
		if err := ors.processDirectoryTree(ctx, child, targetBase, opts, result); err != nil {
			return err
		}
	}

	wg.Wait()
	return nil
}

// processFile handles individual file organization
func (ors *OrganizationService) processFile(ctx context.Context, file *trees.FileNode, targetBase string, opts options.OrganizationOptions, mu *sync.Mutex, result *types.OrganizationResult) error {
	// Determine target category/folder
	targetCategory, err := ors.determineTargetCategory(ctx, file, opts)
	if err != nil {
		return fmt.Errorf("failed to determine target category for %s: %w", file.Path, err)
	}

	// Skip if no category determined
	if targetCategory == "" {
		return nil
	}

	// Construct target path
	targetDir := filepath.Join(targetBase, targetCategory)
	targetPath := filepath.Join(targetDir, file.Name)

	// Check for conflicts
	if _, err := ors.Stat(targetPath); err == nil {
		conflict := types.ConflictInfo{
			SourcePath:   file.Path,
			TargetPath:   targetPath,
			ConflictType: types.ConflictFileExists,
			Resolution:   string(opts.ConflictResolution),
		}

		// Resolve conflict
		resolvedPath, err := ors.conflictMgr.ResolveConflict(ctx, file.Path, targetPath, opts.ConflictResolution)
		if err != nil {
			return fmt.Errorf("failed to resolve conflict for %s: %w", file.Path, err)
		}

		targetPath = resolvedPath
		conflict.ResolvedPath = resolvedPath

		mu.Lock()
		result.Conflicts = append(result.Conflicts, conflict)
		mu.Unlock()
	}

	// Create target directory if needed
	if !opts.DryRun {
		if err := ors.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create target directory %s: %w", filepath.Dir(targetPath), err)
		}
	}

	// Perform file operation
	operation := types.FileOperation{
		Type:       types.FileOpMove,
		SourcePath: file.Path,
		TargetPath: targetPath,
		Category:   targetCategory,
		Timestamp:  time.Now(),
		DryRun:     opts.DryRun,
	}

	if !opts.DryRun {
		// Use the file operations service
		moveOpts := options.CopyOptions{
			PreservePerms:  true,
			PreserveTimes:  true,
			RemoveSource:   !opts.CopyInsteadOfMove,
			FollowSymlinks: false,
		}

		if opts.CopyInsteadOfMove {
			operation.Type = types.FileOpCopy
			if err := ors.fileOps.CopyFile(ctx, file.Path, targetPath, moveOpts); err != nil {
				operation.Success = false
				operation.Error = err.Error()
				return fmt.Errorf("failed to copy file %s to %s: %w", file.Path, targetPath, err)
			}
		} else {
			if err := ors.fileOps.MoveFile(ctx, file.Path, targetPath, moveOpts); err != nil {
				operation.Success = false
				operation.Error = err.Error()
				return fmt.Errorf("failed to move file %s to %s: %w", file.Path, targetPath, err)
			}
		}
	}

	operation.Success = true

	// Record operation
	mu.Lock()
	result.ProcessedFiles = append(result.ProcessedFiles, operation)
	mu.Unlock()

	// Emit file operation event
	eventType := types.EventFileMoved
	if opts.CopyInsteadOfMove {
		eventType = types.EventFileCopied
	}

	ors.emitEvent(types.Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Path:      file.Path,
		Operation: string(operation.Type),
		Success:   operation.Success,
		Metadata: map[string]interface{}{
			"target_path": targetPath,
			"category":    targetCategory,
			"dry_run":     opts.DryRun,
		},
	})

	return nil
}

// determineTargetCategory determines the target category for a file using rule-based logic
func (ors *OrganizationService) determineTargetCategory(ctx context.Context, file *trees.FileNode, opts options.OrganizationOptions) (string, error) {
	// For now, use rule-based categorization
	// AI categorization can be added later when available
	if opts.UseAI {
		slog.Debug("AI categorization requested but not available, falling back to rule-based")
	}

	// Apply rule-based categorization
	return ors.categorizeByRules(file, opts.CategoryRules)
}

// categorizeWithAI is a placeholder for future AI categorization
func (ors *OrganizationService) categorizeWithAI(ctx context.Context, file *trees.FileNode, prompt string) (string, error) {
	// Placeholder for future AI integration
	return "", fmt.Errorf("AI categorization not yet implemented")
}

// categorizeByRules applies rule-based categorization
func (ors *OrganizationService) categorizeByRules(file *trees.FileNode, rules map[string][]string) (string, error) {
	ext := strings.ToLower(file.Extension)

	// Check custom rules first
	for category, extensions := range rules {
		for _, ruleExt := range extensions {
			if ext == strings.ToLower(ruleExt) {
				return category, nil
			}
		}
	}

	// Default categorization by extension
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".svg":
		return "Images", nil
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm":
		return "Videos", nil
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a":
		return "Audio", nil
	case ".pdf", ".doc", ".docx", ".txt", ".rtf", ".odt":
		return "Documents", nil
	case ".xls", ".xlsx", ".csv", ".ods":
		return "Spreadsheets", nil
	case ".ppt", ".pptx", ".odp":
		return "Presentations", nil
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2":
		return "Archives", nil
	case ".exe", ".msi", ".dmg", ".pkg", ".deb", ".rpm":
		return "Applications", nil
	case ".js", ".html", ".css", ".py", ".go", ".java", ".cpp", ".c":
		return "Code", nil
	default:
		return "Other", nil
	}
}

// PreviewOrganization generates a preview of organization operations without executing them
func (ors *OrganizationService) PreviewOrganization(ctx context.Context, opts options.OrganizationOptions) (*types.OrganizationPreview, error) {
	if opts.SourceDir == "" || opts.TargetDir == "" {
		return nil, fmt.Errorf("source and target directories must be specified")
	}

	// Set dry run for preview
	opts.DryRun = true

	result, err := ors.OrganizeDirectory(ctx, opts.SourceDir, opts.TargetDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate organization preview: %w", err)
	}

	// Convert FileOperation to OrganizationOperation
	operations := make([]*types.OrganizationOperation, len(result.ProcessedFiles))
	for i, op := range result.ProcessedFiles {
		operations[i] = &types.OrganizationOperation{
			Type:       types.OperationType(op.Type),
			SourcePath: op.SourcePath,
			TargetPath: op.TargetPath,
			Reason:     op.Category,
			Priority:   1,
		}
	}

	// Convert ConflictInfo to pointers
	conflicts := make([]*types.ConflictInfo, len(result.Conflicts))
	for i := range result.Conflicts {
		conflicts[i] = &result.Conflicts[i]
	}

	return &types.OrganizationPreview{
		SourcePath:    opts.SourceDir,
		TargetPath:    opts.TargetDir,
		TotalFiles:    len(result.ProcessedFiles),
		TotalSize:     result.SourceAnalysis.TotalSize,
		Operations:    operations,
		Conflicts:     conflicts,
		EstimatedTime: ors.estimateProcessingTime(result.ProcessedFiles),
		CreatedAt:     time.Now(),
	}, nil
}

// extractCategories extracts unique categories from operations
func (ors *OrganizationService) extractCategories(operations []types.FileOperation) map[string]int {
	categories := make(map[string]int)
	for _, op := range operations {
		categories[op.Category]++
	}
	return categories
}

// estimateProcessingTime estimates the time needed for processing
func (ors *OrganizationService) estimateProcessingTime(operations []types.FileOperation) time.Duration {
	// Simple estimation: 10ms per file
	return time.Duration(len(operations)) * 10 * time.Millisecond
}

// RegisterEventHandler registers an event handler for organization events
func (ors *OrganizationService) RegisterEventHandler(handler func(types.Event)) {
	ors.mu.Lock()
	defer ors.mu.Unlock()
	ors.eventHandlers = append(ors.eventHandlers, handler)
}

// emitEvent emits an event to all registered handlers
func (ors *OrganizationService) emitEvent(event types.Event) {
	ors.mu.RLock()
	handlers := make([]func(types.Event), len(ors.eventHandlers))
	copy(handlers, ors.eventHandlers)
	ors.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event) // Emit asynchronously
	}
}

// DetermineTargetPath determines the target path for a file based on organization options
func (ors *OrganizationService) DetermineTargetPath(ctx context.Context, fileNode *trees.FileNode, opts options.OrganizationOptions) (string, bool, error) {
	// Determine target category
	category, err := ors.determineTargetCategory(ctx, fileNode, opts)
	if err != nil {
		return "", false, fmt.Errorf("failed to determine category: %w", err)
	}

	// Skip if no category determined
	if category == "" {
		return "", false, nil
	}

	// Construct target path
	targetDir := filepath.Join(opts.TargetDir, category)
	targetPath := filepath.Join(targetDir, fileNode.Name)

	return targetPath, true, nil
}

// ExecuteOrganization executes a previewed organization operation
func (ors *OrganizationService) ExecuteOrganization(ctx context.Context, preview *types.OrganizationPreview, opts options.OrganizationOptions) error {
	if preview == nil {
		return fmt.Errorf("preview cannot be nil")
	}

	// Ensure dry run is disabled for actual execution
	opts.DryRun = false

	slog.Info("Executing organization operations",
		"total_operations", len(preview.Operations),
		"source", preview.SourcePath,
		"target", preview.TargetPath)

	// Execute each operation
	for _, operation := range preview.Operations {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Convert OrganizationOperation to file operation
		moveOpts := options.CopyOptions{
			PreservePerms:  true,
			PreserveTimes:  true,
			RemoveSource:   operation.Type == types.OpMove,
			FollowSymlinks: false,
		}

		// Create target directory if needed
		if err := ors.MkdirAll(filepath.Dir(operation.TargetPath), 0755); err != nil {
			return fmt.Errorf("failed to create target directory %s: %w", filepath.Dir(operation.TargetPath), err)
		}

		// Execute operation
		switch operation.Type {
		case types.OpCopy:
			if err := ors.fileOps.CopyFile(ctx, operation.SourcePath, operation.TargetPath, moveOpts); err != nil {
				return fmt.Errorf("failed to copy file %s to %s: %w", operation.SourcePath, operation.TargetPath, err)
			}
		case types.OpMove:
			if err := ors.fileOps.MoveFile(ctx, operation.SourcePath, operation.TargetPath, moveOpts); err != nil {
				return fmt.Errorf("failed to move file %s to %s: %w", operation.SourcePath, operation.TargetPath, err)
			}
		case types.OpSkip:
			slog.Debug("Skipping file", "path", operation.SourcePath)
			continue
		default:
			slog.Warn("Unknown operation type", "type", operation.Type, "path", operation.SourcePath)
		}

		slog.Debug("Executed operation",
			"type", operation.Type,
			"source", operation.SourcePath,
			"target", operation.TargetPath)
	}

	slog.Info("Organization execution completed",
		"operations", len(preview.Operations))

	return nil
}

// OrganizeFiles performs file organization based on the provided options
func (ors *OrganizationService) OrganizeFiles(ctx context.Context, opts options.OrganizationOptions) error {
	if opts.SourceDir == "" || opts.TargetDir == "" {
		return fmt.Errorf("source and target directories must be specified")
	}

	_, err := ors.OrganizeDirectory(ctx, opts.SourceDir, opts.TargetDir, opts)
	return err
}

// Utility methods for file system operations
func (ors *OrganizationService) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (ors *OrganizationService) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Ensure OrganizationService implements the interface
var _ interfaces.OrganizationService = (*OrganizationService)(nil)
