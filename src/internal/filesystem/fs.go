package filesystem

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"file4you/internal/config"
	"file4you/internal/db"
	"file4you/internal/filesystem/interfaces"
	"file4you/internal/filesystem/options"
	"file4you/internal/filesystem/services"
	"file4you/internal/filesystem/types"
	"file4you/internal/filesystem/utils"
	"file4you/internal/trees"
	"file4you/internal/ui"

	"github.com/ZanzyTHEbar/assert-lib"
	ignore "github.com/sabhiram/go-gitignore"
)

// FileSystem is the main filesystem manager for the file4you application.
// It provides a modern, service-oriented interface for file organization and management.
type FileSystem struct {
	// Core services
	directoryService    interfaces.DirectoryService
	fileOperations      interfaces.FileOperations
	organizationService interfaces.OrganizationService
	conflictResolver    interfaces.ConflictResolver
	gitService          interfaces.GitService

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

// New creates a new modern filesystem manager
func New(interactor ui.Interactor, centralDB db.ICentralDBProvider) (*FileSystem, error) {
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

	// Create assert handler for workspace manager
	assertHandler := assert.NewAssertHandler()

	// Create workspace manager
	workspaceManager := NewWorkspaceManager(centralDB, assertHandler)

	// Create context for concurrent operations
	ctx := context.Background()

	// Create concurrent traverser for high-performance directory operations
	traverser := NewConcurrentTraverser(ctx)

	// Create services in correct order
	conflictResolver := services.NewConflictResolverService()
	fileOperations := services.NewFileOperationsService(conflictResolver, cacheDir)
	gitService := services.NewGitService()
	directoryService := services.NewDirectoryManagerService(traverser, centralDB.GetDirectoryTree())
	organizationService := services.NewOrganizationService(conflictResolver, fileOperations, directoryService)

	return &FileSystem{
		directoryService:    directoryService,
		fileOperations:      fileOperations,
		organizationService: organizationService,
		conflictResolver:    conflictResolver,
		gitService:          gitService,
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
func (dfs *FileSystem) OrganizeDirectory(ctx context.Context, sourceDir, targetDir string, opts options.OrganizationOptions) (*types.OrganizationResult, error) {
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
	result, err := dfs.organizationService.OrganizeDirectory(ctx, sourceDir, targetDir, opts)
	if err != nil {
		return nil, fmt.Errorf("organization failed: %w", err)
	}

	slog.Info("Directory organization completed",
		"duration", result.Duration,
		"processed", len(result.ProcessedFiles),
		"conflicts", len(result.Conflicts))

	return result, nil
}

// PreviewOrganization generates a preview of organization operations
func (dfs *FileSystem) PreviewOrganization(ctx context.Context, opts options.OrganizationOptions) (*types.OrganizationPreview, error) {
	if opts.SourceDir == "" || opts.TargetDir == "" {
		return nil, fmt.Errorf("source and target directories must be specified")
	}

	return dfs.organizationService.PreviewOrganization(ctx, opts)
}

// ExecuteOrganization executes a previewed organization
func (dfs *FileSystem) ExecuteOrganization(ctx context.Context, preview *types.OrganizationPreview, opts options.OrganizationOptions) error {
	return dfs.organizationService.ExecuteOrganization(ctx, preview, opts)
}

// IndexDirectory indexes a directory structure for fast operations
func (dfs *FileSystem) IndexDirectory(ctx context.Context, rootPath string, opts options.IndexOptions) error {
	slog.Info("Starting directory indexing", "path", rootPath)

	// Validate path
	if err := dfs.pathUtils.ValidatePath(rootPath); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Use directory service for indexing
	return dfs.directoryService.IndexDirectory(ctx, rootPath, opts)
}

// AnalyzeDirectory performs comprehensive directory analysis
func (dfs *FileSystem) AnalyzeDirectory(ctx context.Context, rootPath string) (*types.DirectoryAnalysis, error) {
	slog.Info("Starting directory analysis", "path", rootPath)

	return dfs.directoryService.AnalyzeDirectory(ctx, rootPath)
}

// File operation methods

// CopyFile copies a single file with advanced options
func (dfs *FileSystem) CopyFile(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	return dfs.fileOperations.CopyFile(ctx, srcPath, dstPath, opts)
}

// MoveFile moves a single file with advanced options
func (dfs *FileSystem) MoveFile(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	return dfs.fileOperations.MoveFile(ctx, srcPath, dstPath, opts)
}

// DeleteFile deletes a single file
func (dfs *FileSystem) DeleteFile(ctx context.Context, path string) error {
	return dfs.fileOperations.DeleteFile(ctx, path)
}

// CopyDirectory copies a directory recursively
func (dfs *FileSystem) CopyDirectory(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	return dfs.fileOperations.CopyDirectory(ctx, srcPath, dstPath, opts)
}

// MoveDirectory moves a directory recursively
func (dfs *FileSystem) MoveDirectory(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	return dfs.fileOperations.MoveDirectory(ctx, srcPath, dstPath, opts)
}

// DeleteDirectory deletes a directory recursively
func (dfs *FileSystem) DeleteDirectory(ctx context.Context, path string, recursive bool) error {
	return dfs.fileOperations.DeleteDirectory(ctx, path, recursive)
}

// Conflict resolution methods

// ResolveConflict resolves a file conflict using specified strategy
func (dfs *FileSystem) ResolveConflict(ctx context.Context, srcPath, dstPath string, strategy options.ConflictStrategy) (string, error) {
	return dfs.conflictResolver.ResolveConflict(ctx, srcPath, dstPath, strategy)
}

// DetectConflict checks for conflicts between source and destination
func (dfs *FileSystem) DetectConflict(ctx context.Context, srcPath, dstPath string) (*types.ConflictInfo, error) {
	return dfs.conflictResolver.DetectConflict(ctx, srcPath, dstPath)
}

// Git service methods for repository management

func (dfs *FileSystem) IsGitRepo(dir string) bool {
	return dfs.gitService.IsRepository(dir)
}

func (dfs *FileSystem) InitGitRepo(dir string) error {
	ctx := context.Background()
	return dfs.gitService.InitRepository(ctx, dir)
}

func (dfs *FileSystem) GitRewind(dir string, stepsOrSha string) error {
	ctx := context.Background()
	return dfs.gitService.Rewind(ctx, dir, stepsOrSha)
}

func (dfs *FileSystem) GitAddAndCommit(dir, message string) error {
	ctx := context.Background()
	return dfs.gitService.AddAndCommit(ctx, dir, message)
}

func (dfs *FileSystem) GitHasUncommittedChanges(dir string) (bool, error) {
	ctx := context.Background()
	return dfs.gitService.HasUncommittedChanges(ctx, dir)
}

func (dfs *FileSystem) GitStashCreate(dir, message string) error {
	ctx := context.Background()
	return dfs.gitService.StashCreate(ctx, dir, message)
}

func (dfs *FileSystem) GitStashPop(dir string, forceOverwrite bool) error {
	ctx := context.Background()
	return dfs.gitService.StashPop(ctx, dir, forceOverwrite)
}

// Utility methods

// CalculateMaxDepth calculates the maximum depth of a directory tree
func (dfs *FileSystem) CalculateMaxDepth(rootPath string) (int, error) {
	return dfs.depthUtils.CalculateMaxDepthInDirectory(rootPath)
}

// GetFileType determines the type of a file
func (dfs *FileSystem) GetFileType(path string) string {
	return dfs.fileUtils.GetFileType(path)
}

// ValidatePath validates that a path is safe and accessible
func (dfs *FileSystem) ValidatePath(path string) error {
	return dfs.pathUtils.ValidatePath(path)
}

// GetDirectoryTree returns the current directory tree
func (dfs *FileSystem) GetDirectoryTree() *trees.DirectoryTree {
	return dfs.workspaceManager.centralDB.GetDirectoryTree()
}

// GetDesktopCleanerIgnore loads ignore patterns for file organization
func (dfs *FileSystem) GetDesktopCleanerIgnore(dir string) (services.IgnoreChecker, error) {
	ignorePath := filepath.Join(dir, ".file4you-ignore")

	if _, err := os.Stat(ignorePath); err == nil {
		ignored, err := ignore.CompileIgnoreFile(ignorePath)
		if err != nil {
			return nil, fmt.Errorf("error reading .file4you-ignore file: %w", err)
		}
		return ignored, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error checking for .file4you-ignore file: %w", err)
	}

	return nil, nil
}

// Public API methods for CLI and external access

// GetWorkspaceManager returns the workspace manager
func (dfs *FileSystem) GetWorkspaceManager() *WorkspaceManager {
	return dfs.workspaceManager
}

// GetCwd returns the current working directory
func (dfs *FileSystem) GetCwd() string {
	return dfs.cwd
}

// GetConfig returns the configuration
func (dfs *FileSystem) GetConfig() *config.File4YouConfig {
	return dfs.config
}

// GetGitService returns the git service for git operations
func (dfs *FileSystem) GetGitService() interfaces.GitService {
	return dfs.gitService
}

// Service accessor methods

// GetDirectoryService returns the directory service instance
func (dfs *FileSystem) GetDirectoryService() interfaces.DirectoryService {
	return dfs.directoryService
}

// GetFileOperations returns the file operations service instance
func (dfs *FileSystem) GetFileOperations() interfaces.FileOperations {
	return dfs.fileOperations
}

// GetOrganizationService returns the organization service instance
func (dfs *FileSystem) GetOrganizationService() interfaces.OrganizationService {
	return dfs.organizationService
}

// GetConflictResolver returns the conflict resolver service instance
func (dfs *FileSystem) GetConflictResolver() interfaces.ConflictResolver {
	return dfs.conflictResolver
}

// OrganizeWithOptions organizes files using the new options system
func (dfs *FileSystem) OrganizeWithOptions(ctx context.Context, opts options.OrganizationOptions) error {
	return dfs.organizationService.OrganizeFiles(ctx, opts)
}

// Legacy compatibility methods - TODO: Remove after CLI migration

// EnhancedOrganize provides legacy compatibility for the CLI organizer
func (dfs *FileSystem) EnhancedOrganize(cfg *config.File4YouConfig, params *options.FilePathParams) error {
	ctx := context.Background()
	opts := params.ToOrganizationOptions()
	opts.Config = cfg
	return dfs.OrganizeWithOptions(ctx, opts)
}

// InstanceConfig returns the instance configuration for legacy compatibility
func (dfs *FileSystem) InstanceConfig() *config.File4YouConfig {
	return dfs.config
}
