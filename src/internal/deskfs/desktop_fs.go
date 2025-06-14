package deskfs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"file4you/internal/config"
	"file4you/internal/db"
	"file4you/internal/deskfs/interfaces"
	"file4you/internal/deskfs/options"
	"file4you/internal/deskfs/services"
	"file4you/internal/deskfs/types"
	"file4you/internal/deskfs/utils"
	"file4you/internal/filesystem/trees"
	"file4you/internal/ui"
)

// DesktopFileSystem is the main filesystem manager for the file4you application.
// It provides a modern, service-oriented interface for file organization and management.
type DesktopFileSystem struct {
	// Core services
	directoryService    interfaces.DirectoryService
	fileOperations      interfaces.FileOperations
	organizationService interfaces.OrganizationService
	conflictResolver    interfaces.ConflictResolver

	// Utilities
	pathUtils   *utils.PathUtils
	fileUtils   *utils.FileUtils
	depthUtils  *utils.DepthUtils
	safetyUtils *utils.SafetyUtils

	// System components
	workspaceManager *WorkspaceManager
	config           *config.File4YouConfig
	terminal         ui.Interactor

	// Metadata
	homeDir  string
	cwd      string
	cacheDir string
}

// NewDesktopFileSystem creates a new modern filesystem manager
func NewDesktopFileSystem(interactor ui.Interactor, centralDB db.ICentralDBProvider) (*DesktopFileSystem, error) {
	// Get system directories
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Determine cache directory
	cacheDir := config.AppConfig.File4You.CacheDir
	if cacheDir == "" {
		cacheDir = fmt.Sprintf("%s/.file4you/.cache", home)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}

	// Create utilities
	pathUtils := utils.NewPathUtils()
	fileUtils := utils.NewFileUtils()
	depthUtils := utils.NewDepthUtils()
	safetyUtils := utils.NewSafetyUtils()

	// Create workspace manager
	workspaceManager := NewWorkspaceManager(centralDB, nil) // TODO: Add assert handler

	// Create services in correct order
	conflictResolver := services.NewConflictResolverService()
	fileOperations := services.NewFileOperationsService(conflictResolver, cacheDir)
	// For now, pass nil for ConcurrentTraverser - will be implemented later
	directoryService := services.NewDirectoryManagerService(nil, centralDB.GetDirectoryTree())
	organizationService := services.NewOrganizationService(conflictResolver, fileOperations, directoryService)

	return &DesktopFileSystem{
		directoryService:    directoryService,
		fileOperations:      fileOperations,
		organizationService: organizationService,
		conflictResolver:    conflictResolver,
		pathUtils:           pathUtils,
		fileUtils:           fileUtils,
		depthUtils:          depthUtils,
		safetyUtils:         safetyUtils,
		workspaceManager:    workspaceManager,
		config:              &config.AppConfig.File4You,
		terminal:            interactor,
		homeDir:             home,
		cwd:                 cwd,
		cacheDir:            cacheDir,
	}, nil
}

// High-level API methods

// OrganizeDirectory organizes files in a directory using modern service architecture
func (dfs *DesktopFileSystem) OrganizeDirectory(ctx context.Context, sourceDir, targetDir string, opts options.OrganizationOptions) (*types.OrganizationResult, error) {
	start := time.Now()
	slog.Info("Starting directory organization",
		"source", sourceDir,
		"target", targetDir,
		"dryRun", opts.DryRun)

	// Validate paths
	if err := dfs.safetyUtils.ValidateOperationSafety(sourceDir, targetDir, true); err != nil {
		return nil, fmt.Errorf("operation safety validation failed: %w", err)
	}

	// Set defaults if not provided
	if opts.WorkerCount == 0 {
		opts.WorkerCount = 4
	}
	if opts.BatchSize == 0 {
		opts.BatchSize = 100
	}

	opts.SourceDir = sourceDir
	opts.TargetDir = targetDir

	// Use the organization service
	if err := dfs.organizationService.OrganizeFiles(ctx, opts); err != nil {
		return nil, fmt.Errorf("organization failed: %w", err)
	}

	// For now, return a basic result (the actual result would come from a different method)
	result := &types.OrganizationResult{
		StartTime:      start,
		EndTime:        time.Now(),
		Duration:       time.Since(start),
		SourcePath:     sourceDir,
		TargetPath:     targetDir,
		Success:        true,
		DryRun:         opts.DryRun,
		ProcessedFiles: make([]types.FileOperation, 0), // TODO: Get from service
		Conflicts:      make([]types.ConflictInfo, 0),  // TODO: Get from service
		Events:         make([]types.Event, 0),         // TODO: Get from service
	}

	slog.Info("Directory organization completed",
		"duration", time.Since(start),
		"processed", len(result.ProcessedFiles),
		"conflicts", len(result.Conflicts))

	return result, nil
}

// PreviewOrganization generates a preview of organization operations
func (dfs *DesktopFileSystem) PreviewOrganization(ctx context.Context, opts options.OrganizationOptions) (*types.OrganizationPreview, error) {
	if opts.SourceDir == "" || opts.TargetDir == "" {
		return nil, fmt.Errorf("source and target directories must be specified")
	}

	return dfs.organizationService.PreviewOrganization(ctx, opts)
}

// ExecuteOrganization executes a previewed organization
func (dfs *DesktopFileSystem) ExecuteOrganization(ctx context.Context, preview *types.OrganizationPreview, opts options.OrganizationOptions) error {
	return dfs.organizationService.ExecuteOrganization(ctx, preview, opts)
}

// IndexDirectory indexes a directory structure for fast operations
func (dfs *DesktopFileSystem) IndexDirectory(ctx context.Context, rootPath string, opts options.IndexOptions) error {
	slog.Info("Starting directory indexing", "path", rootPath)

	// Validate path
	if err := dfs.pathUtils.ValidatePath(rootPath); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Use directory service for indexing
	return dfs.directoryService.IndexDirectory(ctx, rootPath, opts)
}

// AnalyzeDirectory performs comprehensive directory analysis
func (dfs *DesktopFileSystem) AnalyzeDirectory(ctx context.Context, rootPath string) (*types.DirectoryAnalysis, error) {
	slog.Info("Starting directory analysis", "path", rootPath)

	return dfs.directoryService.AnalyzeDirectory(ctx, rootPath)
}

// File operation methods

// CopyFile copies a single file with advanced options
func (dfs *DesktopFileSystem) CopyFile(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	return dfs.fileOperations.CopyFile(ctx, srcPath, dstPath, opts)
}

// MoveFile moves a single file with advanced options
func (dfs *DesktopFileSystem) MoveFile(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	return dfs.fileOperations.MoveFile(ctx, srcPath, dstPath, opts)
}

// DeleteFile deletes a single file
func (dfs *DesktopFileSystem) DeleteFile(ctx context.Context, path string) error {
	return dfs.fileOperations.DeleteFile(ctx, path)
}

// CopyDirectory copies a directory recursively
func (dfs *DesktopFileSystem) CopyDirectory(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	return dfs.fileOperations.CopyDirectory(ctx, srcPath, dstPath, opts)
}

// MoveDirectory moves a directory recursively
func (dfs *DesktopFileSystem) MoveDirectory(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	return dfs.fileOperations.MoveDirectory(ctx, srcPath, dstPath, opts)
}

// DeleteDirectory deletes a directory recursively
func (dfs *DesktopFileSystem) DeleteDirectory(ctx context.Context, path string, recursive bool) error {
	return dfs.fileOperations.DeleteDirectory(ctx, path, recursive)
}

// Conflict resolution methods

// ResolveConflict resolves a file conflict using specified strategy
func (dfs *DesktopFileSystem) ResolveConflict(ctx context.Context, srcPath, dstPath string, strategy options.ConflictStrategy) (string, error) {
	return dfs.conflictResolver.ResolveConflict(ctx, srcPath, dstPath, strategy)
}

// DetectConflict checks for conflicts between source and destination
func (dfs *DesktopFileSystem) DetectConflict(ctx context.Context, srcPath, dstPath string) (*types.ConflictInfo, error) {
	return dfs.conflictResolver.DetectConflict(ctx, srcPath, dstPath)
}

// Utility methods

// CalculateMaxDepth calculates the maximum depth of a directory tree
func (dfs *DesktopFileSystem) CalculateMaxDepth(rootPath string) (int, error) {
	return dfs.depthUtils.CalculateMaxDepthInDirectory(rootPath)
}

// GetFileType determines the type of a file
func (dfs *DesktopFileSystem) GetFileType(path string) string {
	return dfs.fileUtils.GetFileType(path)
}

// ValidatePath validates that a path is safe and accessible
func (dfs *DesktopFileSystem) ValidatePath(path string) error {
	return dfs.pathUtils.ValidatePath(path)
}

// GetDirectoryTree returns the current directory tree
func (dfs *DesktopFileSystem) GetDirectoryTree() *trees.DirectoryTree {
	return dfs.workspaceManager.centralDB.GetDirectoryTree()
}

// Legacy compatibility methods (to be phased out)

// EnhancedOrganize provides backward compatibility with the old interface
func (dfs *DesktopFileSystem) EnhancedOrganize(cfg *config.File4YouConfig, params *FilePathParams) error {
	slog.Warn("Using deprecated EnhancedOrganize method, please migrate to OrganizeDirectory")

	// Convert old parameters to new options
	opts := options.OrganizationOptions{
		SourceDir:          params.SourceDir,
		TargetDir:          params.TargetDir,
		DryRun:             params.DryRun,
		ConflictResolution: convertConflictResolution(params.ConflictResolution),
		CopyInsteadOfMove:  params.CopyFiles,
		RemoveAfter:        params.RemoveAfter,
		GitEnabled:         params.GitEnabled,
		Recursive:          params.Recursive,
		MaxDepth:           params.MaxDepth,
		WorkerCount:        4,
		BatchSize:          100,
		CategoryRules:      make(map[string][]string),
	}

	ctx := context.Background()
	_, err := dfs.OrganizeDirectory(ctx, params.SourceDir, params.TargetDir, opts)
	return err
}

// convertConflictResolution converts old conflict resolution to new type
func convertConflictResolution(old ConflictResolutionType) options.ConflictStrategy {
	switch old {
	case Overwrite:
		return options.ConflictOverwrite
	case Skip:
		return options.ConflictSkip
	case RenameSuffix:
		return options.ConflictRename
	default:
		return options.ConflictRename
	}
}

// Getters for backward compatibility

// GetHomeDir returns the home directory
func (dfs *DesktopFileSystem) GetHomeDir() string {
	return dfs.homeDir
}

// GetCwd returns the current working directory
func (dfs *DesktopFileSystem) GetCwd() string {
	return dfs.cwd
}

// GetCacheDir returns the cache directory
func (dfs *DesktopFileSystem) GetCacheDir() string {
	return dfs.cacheDir
}

// GetWorkspaceManager returns the workspace manager
func (dfs *DesktopFileSystem) GetWorkspaceManager() *WorkspaceManager {
	return dfs.workspaceManager
}

// GetConfig returns the configuration
func (dfs *DesktopFileSystem) GetConfig() *config.File4YouConfig {
	return dfs.config
}

// GetTerminal returns the terminal interactor
func (dfs *DesktopFileSystem) GetTerminal() ui.Interactor {
	return dfs.terminal
}
