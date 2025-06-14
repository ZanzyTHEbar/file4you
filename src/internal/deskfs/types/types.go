package types

import (
	"file4you/internal/filesystem/trees"
	"time"
)

// Event represents a filesystem event with metadata
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Path      string                 `json:"path"`
	Operation string                 `json:"operation"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EventType defines the types of filesystem events
type EventType string

const (
	EventFileCreated    EventType = "file_created"
	EventFileModified   EventType = "file_modified"
	EventFileDeleted    EventType = "file_deleted"
	EventFileCopied     EventType = "file_copied"
	EventFileMoved      EventType = "file_moved"
	EventDirCreated     EventType = "dir_created"
	EventDirDeleted     EventType = "dir_deleted"
	EventIndexUpdated   EventType = "index_updated"
	EventOperationStart EventType = "operation_start"
	EventOperationEnd   EventType = "operation_end"
)

// DirectoryAnalysis contains analysis results for a directory
type DirectoryAnalysis struct {
	TotalFiles       int               `json:"total_files"`
	TotalDirectories int               `json:"total_directories"`
	TotalSize        int64             `json:"total_size"`
	MaxDepth         int               `json:"max_depth"`
	FileTypes        map[string]int    `json:"file_types"`
	SizeDistribution map[string]int    `json:"size_distribution"`
	AgeDistribution  map[string]int    `json:"age_distribution"`
	LargestFiles     []*trees.FileNode `json:"largest_files"`
	OldestFiles      []*trees.FileNode `json:"oldest_files"`
	NewestFiles      []*trees.FileNode `json:"newest_files"`
	Duration         time.Duration     `json:"duration"`
}

// OrganizationPreview contains a preview of organization operations
type OrganizationPreview struct {
	SourcePath    string                   `json:"source_path"`
	TargetPath    string                   `json:"target_path"`
	TotalFiles    int                      `json:"total_files"`
	TotalSize     int64                    `json:"total_size"`
	Operations    []*OrganizationOperation `json:"operations"`
	Conflicts     []*ConflictInfo          `json:"conflicts"`
	EstimatedTime time.Duration            `json:"estimated_time"`
	CreatedAt     time.Time                `json:"created_at"`
}

// OrganizationOperation represents a single file organization operation
type OrganizationOperation struct {
	Type         OperationType   `json:"type"`
	SourcePath   string          `json:"source_path"`
	TargetPath   string          `json:"target_path"`
	FileInfo     *trees.FileNode `json:"file_info"`
	Reason       string          `json:"reason"`
	Priority     int             `json:"priority"`
	Dependencies []string        `json:"dependencies,omitempty"`
}

// OperationType defines the types of organization operations
type OperationType string

const (
	OpCopy   OperationType = "copy"
	OpMove   OperationType = "move"
	OpDelete OperationType = "delete"
	OpRename OperationType = "rename"
	OpSkip   OperationType = "skip"
)

// ConflictInfo contains information about file conflicts
type ConflictInfo struct {
	SourcePath   string                 `json:"source_path"`
	TargetPath   string                 `json:"target_path"`
	ConflictType ConflictType           `json:"conflict_type"`
	SourceInfo   *trees.FileNode        `json:"source_info"`
	TargetInfo   *trees.FileNode        `json:"target_info,omitempty"`
	Resolution   string                 `json:"resolution,omitempty"`
	ResolvedPath string                 `json:"resolved_path,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ConflictType defines the types of file conflicts
type ConflictType string

const (
	ConflictFileExists       ConflictType = "file_exists"
	ConflictDirectoryExists  ConflictType = "directory_exists"
	ConflictPermissionDenied ConflictType = "permission_denied"
	ConflictDiskSpace        ConflictType = "disk_space"
	ConflictCrossDevice      ConflictType = "cross_device"
)

// BackupInfo contains information about a backup
type BackupInfo struct {
	ID          string                 `json:"id"`
	Path        string                 `json:"path"`
	CreatedAt   time.Time              `json:"created_at"`
	Size        int64                  `json:"size"`
	Compressed  bool                   `json:"compressed"`
	Encrypted   bool                   `json:"encrypted"`
	Checksum    string                 `json:"checksum"`
	Metadata    map[string]interface{} `json:"metadata"`
	Description string                 `json:"description,omitempty"`
}

// OperationResult contains the result of a filesystem operation
type OperationResult struct {
	Success        bool                   `json:"success"`
	Error          error                  `json:"error,omitempty"`
	ProcessedFiles int                    `json:"processed_files"`
	ProcessedDirs  int                    `json:"processed_dirs"`
	SkippedFiles   int                    `json:"skipped_files"`
	Conflicts      []*ConflictInfo        `json:"conflicts,omitempty"`
	Duration       time.Duration          `json:"duration"`
	Events         []Event                `json:"events,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ProgressInfo contains progress information for long-running operations
type ProgressInfo struct {
	Current       int           `json:"current"`
	Total         int           `json:"total"`
	Percentage    float64       `json:"percentage"`
	ElapsedTime   time.Duration `json:"elapsed_time"`
	EstimatedTime time.Duration `json:"estimated_time"`
	Rate          float64       `json:"rate"` // items per second
	Phase         string        `json:"phase"`
}

// PerformanceMetrics contains performance metrics for operations
type PerformanceMetrics struct {
	OperationType     string        `json:"operation_type"`
	TotalOperations   int64         `json:"total_operations"`
	SuccessfulOps     int64         `json:"successful_ops"`
	FailedOps         int64         `json:"failed_ops"`
	AverageTime       time.Duration `json:"average_time"`
	MedianTime        time.Duration `json:"median_time"`
	TotalTime         time.Duration `json:"total_time"`
	ThroughputPerSec  float64       `json:"throughput_per_sec"`
	ConcurrentWorkers int           `json:"concurrent_workers"`
	MemoryUsage       int64         `json:"memory_usage"`
}

// OrganizationResult contains the complete result of an organization operation
type OrganizationResult struct {
	StartTime      time.Time          `json:"start_time"`
	EndTime        time.Time          `json:"end_time"`
	Duration       time.Duration      `json:"duration"`
	SourcePath     string             `json:"source_path"`
	TargetPath     string             `json:"target_path"`
	Success        bool               `json:"success"`
	DryRun         bool               `json:"dry_run"`
	ProcessedFiles []FileOperation    `json:"processed_files"`
	Conflicts      []ConflictInfo     `json:"conflicts"`
	Events         []Event            `json:"events"`
	SourceAnalysis *DirectoryAnalysis `json:"source_analysis,omitempty"`
	Error          string             `json:"error,omitempty"`
}

// FileOperation represents a file operation performed during organization
type FileOperation struct {
	Type       FileOpType    `json:"type"`
	SourcePath string        `json:"source_path"`
	TargetPath string        `json:"target_path"`
	Category   string        `json:"category"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
	DryRun     bool          `json:"dry_run"`
	Size       int64         `json:"size"`
	Duration   time.Duration `json:"duration"`
}

// FileOpType defines the types of file operations
type FileOpType string

const (
	FileOpCopy   FileOpType = "copy"
	FileOpMove   FileOpType = "move"
	FileOpDelete FileOpType = "delete"
	FileOpRename FileOpType = "rename"
	FileOpSkip   FileOpType = "skip"
)

// OrganizationSummary provides a summary of organization operations
type OrganizationSummary struct {
	TotalFiles     int            `json:"total_files"`
	TotalConflicts int            `json:"total_conflicts"`
	Categories     map[string]int `json:"categories"`
	EstimatedTime  time.Duration  `json:"estimated_time"`
}
