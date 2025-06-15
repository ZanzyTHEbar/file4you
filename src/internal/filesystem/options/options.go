package options

import (
	"file4you/internal/config"
	"file4you/internal/filesystem/types"
	"file4you/internal/trees"
	"time"
)

// ConflictStrategy defines how to handle file conflicts
type ConflictStrategy string

const (
	ConflictOverwrite ConflictStrategy = "overwrite"
	ConflictSkip      ConflictStrategy = "skip"
	ConflictRename    ConflictStrategy = "rename"
	ConflictPrompt    ConflictStrategy = "prompt"
)

// CopyOptions configures file and directory copy operations
type CopyOptions struct {
	Recursive      bool             // Copy directories recursively
	RemoveSource   bool             // Remove source after copy (acts like move)
	DryRun         bool             // Preview operations without executing
	Conflict       ConflictStrategy // How to handle file conflicts
	PreservePerms  bool             // Preserve file permissions
	PreserveTimes  bool             // Preserve modification times
	FollowSymlinks bool             // Follow symbolic links
	MaxDepth       int              // Maximum recursion depth (0 = unlimited)
}

// MoveOptions configures file and directory move operations
type MoveOptions struct {
	DryRun         bool             // Preview operations without executing
	Conflict       ConflictStrategy // How to handle file conflicts
	PreservePerms  bool             // Preserve file permissions
	FallbackToCopy bool             // Use copy+delete for cross-device moves
	MaxRetries     int              // Retry attempts for transient failures
}

// DeleteOptions configures file and directory deletion operations
type DeleteOptions struct {
	Recursive    bool // Delete directories recursively
	Force        bool // Force deletion of read-only files
	DryRun       bool // Preview operations without executing
	MoveToTrash  bool // Move to trash instead of permanent deletion
	SecureDelete bool // Use secure deletion methods
}

// IndexOptions configures directory indexing operations
type IndexOptions struct {
	Recursive        bool                     // Index directories recursively
	MaxDepth         int                      // Maximum recursion depth
	UpdateExisting   bool                     // Update existing index entries
	IncludeHidden    bool                     // Include hidden files and directories
	IgnorePatterns   []string                 // Patterns to ignore (gitignore style)
	IndexTypes       []string                 // Types of indexes to build
	BatchSize        int                      // Batch size for database operations
	WorkerCount      int                      // Number of concurrent workers
	ProgressCallback func(current, total int) // Progress reporting
}

// TraversalOptions configures directory traversal operations
type TraversalOptions struct {
	Recursive      bool                       // Traverse directories recursively
	MaxDepth       int                        // Maximum recursion depth
	FollowSymlinks bool                       // Follow symbolic links
	IncludeHidden  bool                       // Include hidden files and directories
	Filter         func(*trees.FileNode) bool // File filter function
	SortBy         SortCriteria               // How to sort results
	WorkerCount    int                        // Number of concurrent workers
	BufferSize     int                        // Channel buffer size
}

// SortCriteria defines sorting options for directory traversal
type SortCriteria struct {
	Field     SortField // Field to sort by
	Direction SortDirection
}

type SortField string

const (
	SortByName    SortField = "name"
	SortBySize    SortField = "size"
	SortByModTime SortField = "modtime"
	SortByPath    SortField = "path"
)

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// OrganizationOptions configures file organization operations
type OrganizationOptions struct {
	SourceDir          string                   // Source directory to organize
	TargetDir          string                   // Target directory for organized files
	Config             *config.File4YouConfig   // Configuration for organization rules
	DryRun             bool                     // Preview operations without executing
	ConflictResolution ConflictStrategy         // How to handle file conflicts
	CopyInsteadOfMove  bool                     // Copy instead of move
	RemoveAfter        bool                     // Remove source after copy
	GitEnabled         bool                     // Enable git operations
	Timeout            time.Duration            // Operation timeout
	Recursive          bool                     // Process directories recursively
	MaxDepth           int                      // Maximum recursion depth
	WorkerCount        int                      // Number of concurrent workers
	BatchSize          int                      // Batch size for operations
	IncludeHidden      bool                     // Include hidden files
	UseAI              bool                     // Use AI for categorization
	AIPrompt           string                   // Custom AI prompt for categorization
	CategoryRules      map[string][]string      // Custom categorization rules
	ProgressCallback   func(current, total int) // Progress reporting
	EventCallback      func(types.Event)        // Event notification
}

// BackupOptions configures backup operations
type BackupOptions struct {
	IncludeConfig     bool     // Include configuration files
	IncludeWorkspaces bool     // Include workspace data
	IncludeCache      bool     // Include cache data
	Compression       bool     // Compress backup files
	Encryption        bool     // Encrypt backup files
	ExcludePatterns   []string // Patterns to exclude from backup
	DestinationPath   string   // Custom backup destination
}

// RestoreOptions configures restore operations
type RestoreOptions struct {
	OverwriteExisting bool     // Overwrite existing files
	RestoreConfig     bool     // Restore configuration files
	RestoreWorkspaces bool     // Restore workspace data
	RestoreCache      bool     // Restore cache data
	TargetPath        string   // Custom restore target
	IncludePatterns   []string // Patterns to include in restore
}

// DefaultCopyOptions returns sensible defaults for copy operations
func DefaultCopyOptions() CopyOptions {
	return CopyOptions{
		Recursive:      true,
		RemoveSource:   false,
		DryRun:         false,
		Conflict:       ConflictRename,
		PreservePerms:  true,
		PreserveTimes:  true,
		FollowSymlinks: false,
		MaxDepth:       0, // Unlimited
	}
}

// DefaultMoveOptions returns sensible defaults for move operations
func DefaultMoveOptions() MoveOptions {
	return MoveOptions{
		DryRun:         false,
		Conflict:       ConflictRename,
		PreservePerms:  true,
		FallbackToCopy: true,
		MaxRetries:     3,
	}
}

// DefaultTraversalOptions returns sensible defaults for traversal operations
func DefaultTraversalOptions() TraversalOptions {
	return TraversalOptions{
		Recursive:      true,
		MaxDepth:       0, // Unlimited
		FollowSymlinks: false,
		IncludeHidden:  false,
		SortBy: SortCriteria{
			Field:     SortByName,
			Direction: SortAsc,
		},
		WorkerCount: 4,
		BufferSize:  100,
	}
}

// DefaultOrganizationOptions returns sensible defaults for organization operations
func DefaultOrganizationOptions() OrganizationOptions {
	return OrganizationOptions{
		DryRun:             false,
		ConflictResolution: ConflictRename,
		CopyInsteadOfMove:  false,
		RemoveAfter:        false,
		GitEnabled:         false,
		Timeout:            10 * time.Minute,
		Recursive:          true,
		MaxDepth:           0, // Unlimited
		WorkerCount:        4,
		BatchSize:          100,
		IncludeHidden:      false,
		UseAI:              false,
		CategoryRules:      make(map[string][]string),
	}
}

// Legacy compatibility types - TODO: Remove after CLI migration

// FilePathParams provides legacy compatibility for CLI
type FilePathParams struct {
	SourceDir   string
	TargetDir   string
	Recursive   bool
	DryRun      bool
	MaxDepth    int
	GitEnabled  bool
	CopyFiles   bool
	RemoveAfter bool
}

// NewFilePathParams creates a new FilePathParams with defaults
func NewFilePathParams() *FilePathParams {
	return &FilePathParams{
		MaxDepth: -1,
	}
}

// ToOrganizationOptions converts legacy params to new options
func (fp *FilePathParams) ToOrganizationOptions() OrganizationOptions {
	return OrganizationOptions{
		SourceDir:         fp.SourceDir,
		TargetDir:         fp.TargetDir,
		Recursive:         fp.Recursive,
		DryRun:            fp.DryRun,
		MaxDepth:          fp.MaxDepth,
		GitEnabled:        fp.GitEnabled,
		CopyInsteadOfMove: fp.CopyFiles,
		RemoveAfter:       fp.RemoveAfter,
		WorkerCount:       4,
		BatchSize:         100,
	}
}
