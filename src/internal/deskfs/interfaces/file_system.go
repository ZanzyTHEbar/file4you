package interfaces

import (
	"context"
	"file4you/internal/deskfs/options"
	"file4you/internal/deskfs/types"
	"file4you/internal/filesystem/trees"
)

// FileSystemManager defines the core file system operations interface
type FileSystemManager interface {
	// Copy operations
	Copy(ctx context.Context, src *trees.DirectoryNode, dst string, opts options.CopyOptions) error
	CopyFile(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error

	// Move operations
	Move(ctx context.Context, src *trees.DirectoryNode, dst string, opts options.MoveOptions) error
	MoveFile(ctx context.Context, srcPath, dstPath string, opts options.MoveOptions) error

	// Delete operations
	Delete(ctx context.Context, path string, opts options.DeleteOptions) error
	MoveToTrash(ctx context.Context, node *trees.DirectoryNode) error
}

// DirectoryService defines directory management and traversal operations
type DirectoryService interface {
	// Index and build directory structures
	IndexDirectory(ctx context.Context, rootPath string, opts options.IndexOptions) error
	BuildDirectoryTree(ctx context.Context, rootPath string, opts options.TraversalOptions) (*trees.DirectoryNode, error)

	// Directory analysis
	CalculateMaxDepth(ctx context.Context, rootPath string) (int, error)
	AnalyzeDirectory(ctx context.Context, rootPath string) (*types.DirectoryAnalysis, error)
}

// OrganizationService defines file organization and workflow operations
type OrganizationService interface {
	// Core organization
	OrganizeFiles(ctx context.Context, opts options.OrganizationOptions) error
	DetermineTargetPath(ctx context.Context, fileNode *trees.FileNode, opts options.OrganizationOptions) (string, bool, error)

	// Workflow operations
	PreviewOrganization(ctx context.Context, opts options.OrganizationOptions) (*types.OrganizationPreview, error)
	ExecuteOrganization(ctx context.Context, preview *types.OrganizationPreview, opts options.OrganizationOptions) error
}

// ConflictResolver defines file conflict resolution strategies
type ConflictResolver interface {
	// Conflict detection and resolution
	ResolveConflict(ctx context.Context, srcPath, dstPath string, strategy options.ConflictStrategy) (string, error)
	DetectConflict(ctx context.Context, srcPath, dstPath string) (*types.ConflictInfo, error)
	GenerateUniqueFilename(path string) string
}

// BackupManager defines backup and restore operations
type BackupManager interface {
	CreateBackup(ctx context.Context, opts options.BackupOptions) (*types.BackupInfo, error)
	RestoreBackup(ctx context.Context, backupPath string, opts options.RestoreOptions) error
	ListBackups(ctx context.Context) ([]*types.BackupInfo, error)
}

// FileOperations defines file and directory operations
type FileOperations interface {
	// File operations
	CopyFile(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error
	MoveFile(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error
	DeleteFile(ctx context.Context, path string) error

	// Directory operations
	CopyDirectory(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error
	MoveDirectory(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error
	DeleteDirectory(ctx context.Context, path string, recursive bool) error

	// Batch operations
	CopyBatch(ctx context.Context, operations []types.OrganizationOperation, opts options.CopyOptions) (*types.OperationResult, error)
	MoveBatch(ctx context.Context, operations []types.OrganizationOperation, opts options.CopyOptions) (*types.OperationResult, error)

	// Utility operations
	ValidatePath(path string) error
	CalculateChecksum(path string) (string, error)
	GetFileInfo(path string) (*trees.FileNode, error)
}

// GitService defines git repository operations
type GitService interface {
	// Repository management
	InitRepository(ctx context.Context, dir string) error
	IsRepository(dir string) bool
	IsGitRepository(ctx context.Context, dir string) (bool, error) // Add convenience method for tests

	// File operations
	AddFiles(ctx context.Context, repoDir string, paths ...string) error
	CommitChanges(ctx context.Context, repoDir, message string) error
	AddAndCommit(ctx context.Context, repoDir, message string) error

	// Status and inspection
	HasUncommittedChanges(ctx context.Context, repoDir string) (bool, error)
	FileHasUncommittedChanges(ctx context.Context, repoDir, path string) (bool, error)
	GetCommitHistory(ctx context.Context, repoDir string) ([]string, error)

	// Stash operations
	StashCreate(ctx context.Context, repoDir, message string) error
	StashPop(ctx context.Context, repoDir string, forceOverwrite bool) error

	// Reset and rewind operations
	CheckoutFile(ctx context.Context, repoDir, path string) error
	ClearUncommittedChanges(ctx context.Context, repoDir string) error
	Rewind(ctx context.Context, repoDir, targetSha string) error
}
