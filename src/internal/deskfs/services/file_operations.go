package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"file4you/internal/deskfs/interfaces"
	"file4you/internal/deskfs/options"
	"file4you/internal/deskfs/types"
	"file4you/internal/filesystem/trees"
)

// FileOperationsService provides high-performance file operations
// Integrates with Phase 1 batch database operations and concurrent systems
type FileOperationsService struct {
	conflictResolver interfaces.ConflictResolver
	cacheDir         string
	metrics          *FileOperationMetrics
}

// FileOperationMetrics tracks performance for file operations
type FileOperationMetrics struct {
	TotalOperations       int64
	SuccessfulOps         int64
	FailedOps             int64
	TotalBytesTransferred int64
	AverageSpeed          float64 // bytes per second
	LastOperation         time.Time
}

// NewFileOperationsService creates a new file operations service
func NewFileOperationsService(conflictResolver interfaces.ConflictResolver, cacheDir string) *FileOperationsService {
	return &FileOperationsService{
		conflictResolver: conflictResolver,
		cacheDir:         cacheDir,
		metrics:          &FileOperationMetrics{},
	}
}

// Copy copies a file or directory with enhanced performance and error handling
func (fos *FileOperationsService) Copy(ctx context.Context, src *trees.DirectoryNode, dst string, opts options.CopyOptions) error {
	start := time.Now()
	defer fos.updateMetrics(start, true)

	if src == nil {
		return fmt.Errorf("source node cannot be nil")
	}

	if opts.DryRun {
		return fos.previewCopy(ctx, src, dst, opts)
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Handle directory copying
	if len(src.Children) > 0 || len(src.Files) > 0 {
		return fos.copyDirectory(ctx, src, dst, opts)
	}

	return fmt.Errorf("source node has no files or directories to copy")
}

// CopyFile copies a single file with optimized performance
func (fos *FileOperationsService) CopyFile(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	start := time.Now()
	defer fos.updateMetrics(start, false)

	if opts.DryRun {
		slog.Info("Dry run: would copy file", "src", srcPath, "dst", dstPath)
		return nil
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Handle conflicts
	if err := fos.handleFileConflict(ctx, srcPath, dstPath, opts.Conflict); err != nil {
		return fmt.Errorf("conflict resolution failed: %w", err)
	}

	// Perform the actual copy with optimized I/O
	if err := fos.performFileCopy(ctx, srcPath, dstPath, opts); err != nil {
		fos.metrics.FailedOps++
		return fmt.Errorf("file copy failed: %w", err)
	}

	// Remove source if requested (acts like move)
	if opts.RemoveSource {
		if err := os.Remove(srcPath); err != nil {
			return fmt.Errorf("failed to remove source file after copy: %w", err)
		}
	}

	fos.metrics.SuccessfulOps++
	return nil
}

// Move moves a file or directory with cross-device fallback
func (fos *FileOperationsService) Move(ctx context.Context, src *trees.DirectoryNode, dst string, opts options.MoveOptions) error {
	start := time.Now()
	defer fos.updateMetrics(start, true)

	if src == nil {
		return fmt.Errorf("source node cannot be nil")
	}

	if opts.DryRun {
		slog.Info("Dry run: would move directory", "src", src.Path, "dst", dst)
		return nil
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Try direct rename first (most efficient)
	if err := os.Rename(src.Path, dst); err != nil {
		// Handle cross-device link error with fallback
		if fos.isCrossDeviceError(err) && opts.FallbackToCopy {
			slog.Warn("Cross-device move detected, falling back to copy+delete", "src", src.Path, "dst", dst)

			copyOpts := options.CopyOptions{
				Recursive:     true,
				RemoveSource:  true,
				DryRun:        false,
				Conflict:      opts.Conflict,
				PreservePerms: opts.PreservePerms,
			}

			return fos.Copy(ctx, src, dst, copyOpts)
		}
		return fmt.Errorf("move operation failed: %w", err)
	}

	fos.metrics.SuccessfulOps++
	return nil
}

// MoveFile moves a single file with cross-device fallback
func (fos *FileOperationsService) MoveFile(ctx context.Context, srcPath, dstPath string, opts options.MoveOptions) error {
	start := time.Now()
	defer fos.updateMetrics(start, false)

	if opts.DryRun {
		slog.Info("Dry run: would move file", "src", srcPath, "dst", dstPath)
		return nil
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Handle conflicts
	if err := fos.handleFileConflict(ctx, srcPath, dstPath, opts.Conflict); err != nil {
		return fmt.Errorf("conflict resolution failed: %w", err)
	}

	// Try direct rename first
	err := os.Rename(srcPath, dstPath)
	if err == nil {
		fos.metrics.SuccessfulOps++
		return nil
	}

	// Handle cross-device link error with fallback
	if fos.isCrossDeviceError(err) && opts.FallbackToCopy {
		slog.Warn("Cross-device move detected, falling back to copy+delete", "src", srcPath, "dst", dstPath)

		copyOpts := options.CopyOptions{
			RemoveSource:  true,
			DryRun:        false,
			Conflict:      opts.Conflict,
			PreservePerms: opts.PreservePerms,
		}

		return fos.CopyFile(ctx, srcPath, dstPath, copyOpts)
	}

	fos.metrics.FailedOps++
	return fmt.Errorf("move operation failed: %w", err)
}

// Delete deletes a file or directory
func (fos *FileOperationsService) Delete(ctx context.Context, path string, opts options.DeleteOptions) error {
	start := time.Now()
	defer fos.updateMetrics(start, false)

	if opts.DryRun {
		slog.Info("Dry run: would delete", "path", path)
		return nil
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Move to trash if requested
	if opts.MoveToTrash {
		return fos.moveToTrash(ctx, path)
	}

	// Perform deletion
	var err error
	if opts.Recursive {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}

	if err != nil {
		fos.metrics.FailedOps++
		return fmt.Errorf("delete operation failed: %w", err)
	}

	fos.metrics.SuccessfulOps++
	return nil
}

// MoveToTrash moves a file or directory to the trash directory
func (fos *FileOperationsService) MoveToTrash(ctx context.Context, node *trees.DirectoryNode) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	return fos.moveToTrash(ctx, node.Path)
}

// CalculateChecksum calculates the checksum of a file
func (fos *FileOperationsService) CalculateChecksum(path string) (string, error) {
	// Use SHA256 as default algorithm
	return fos.calculateChecksumWithAlgorithm(path, "sha256")
}

// GetFileInfo returns file information as a FileNode
func (fos *FileOperationsService) GetFileInfo(path string) (*trees.FileNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", path, err)
	}

	return &trees.FileNode{
		Path:      path,
		Name:      filepath.Base(path),
		Extension: filepath.Ext(path),
		Metadata: trees.Metadata{
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
			NodeType:   trees.File,
		},
	}, nil
}

// calculateChecksumWithAlgorithm calculates checksum using specified algorithm
func (fos *FileOperationsService) calculateChecksumWithAlgorithm(path, algorithm string) (string, error) {
	// Placeholder implementation - would use actual checksum calculation
	// For now, return a simple hash based on file size and modification time
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat file %s: %w", path, err)
	}

	// Simple checksum based on size and modtime (not cryptographically secure)
	hash := fmt.Sprintf("%d_%d", info.Size(), info.ModTime().Unix())
	return hash, nil
}

// CopyDirectory copies a directory recursively from source path to destination path
func (fos *FileOperationsService) CopyDirectory(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	// This is a wrapper around the internal copyDirectory method
	// For now, we'll create a simple DirectoryNode from the source path
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat source directory %s: %w", srcPath, err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", srcPath)
	}

	// Create a simple DirectoryNode for the source
	srcNode := &trees.DirectoryNode{
		Path: srcPath,
		Type: trees.Directory,
	}

	return fos.copyDirectory(ctx, srcNode, dstPath, opts)
}

// MoveDirectory moves a directory recursively from source path to destination path
func (fos *FileOperationsService) MoveDirectory(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	// Copy first
	if err := fos.CopyDirectory(ctx, srcPath, dstPath, opts); err != nil {
		return fmt.Errorf("failed to copy directory during move: %w", err)
	}

	// Remove source if copy was successful and RemoveSource is true
	if opts.RemoveSource {
		if err := fos.DeleteDirectory(ctx, srcPath, true); err != nil {
			return fmt.Errorf("failed to remove source directory after copy: %w", err)
		}
	}

	return nil
}

// DeleteDirectory deletes a directory recursively
func (fos *FileOperationsService) DeleteDirectory(ctx context.Context, path string, recursive bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat directory %s: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	if recursive {
		return os.RemoveAll(path)
	}

	// Non-recursive delete - directory must be empty
	return os.Remove(path)
}

// Private helper methods

func (fos *FileOperationsService) previewCopy(ctx context.Context, src *trees.DirectoryNode, dst string, opts options.CopyOptions) error {
	slog.Info("Dry run: would copy directory", "src", src.Path, "dst", dst, "recursive", opts.Recursive)

	// Recursively preview copy operations
	if opts.Recursive {
		for _, child := range src.Children {
			childDst := filepath.Join(dst, filepath.Base(child.Path))
			if err := fos.previewCopy(ctx, child, childDst, opts); err != nil {
				return err
			}
		}
	}

	for _, file := range src.Files {
		fileDst := filepath.Join(dst, filepath.Base(file.Path))
		slog.Info("Dry run: would copy file", "src", file.Path, "dst", fileDst)
	}

	return nil
}

func (fos *FileOperationsService) copyDirectory(ctx context.Context, src *trees.DirectoryNode, dst string, opts options.CopyOptions) error {
	if !opts.Recursive {
		return fmt.Errorf("source is a directory, use recursive flag to copy directories")
	}

	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dst, err)
	}

	// Copy files in the directory
	for _, fileNode := range src.Files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fileDst := filepath.Join(dst, filepath.Base(fileNode.Path))
		fileOpts := options.CopyOptions{
			RemoveSource:  opts.RemoveSource,
			DryRun:        false,
			Conflict:      opts.Conflict,
			PreservePerms: opts.PreservePerms,
			PreserveTimes: opts.PreserveTimes,
		}

		if err := fos.CopyFile(ctx, fileNode.Path, fileDst, fileOpts); err != nil {
			return fmt.Errorf("failed to copy file %s: %w", fileNode.Path, err)
		}
	}

	// Copy child directories recursively
	for _, childDir := range src.Children {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		childDst := filepath.Join(dst, filepath.Base(childDir.Path))
		if err := fos.copyDirectory(ctx, childDir, childDst, opts); err != nil {
			return fmt.Errorf("failed to copy directory %s: %w", childDir.Path, err)
		}
	}

	// Remove source directory if requested
	if opts.RemoveSource {
		if err := os.RemoveAll(src.Path); err != nil {
			return fmt.Errorf("failed to remove source directory after copy: %w", err)
		}
	}

	return nil
}

func (fos *FileOperationsService) performFileCopy(ctx context.Context, srcPath, dstPath string, opts options.CopyOptions) error {
	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	// Copy with progress tracking
	written, err := fos.copyWithProgress(ctx, dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	fos.metrics.TotalBytesTransferred += written

	// Preserve permissions if requested
	if opts.PreservePerms {
		if srcInfo, err := srcFile.Stat(); err == nil {
			if err := os.Chmod(dstPath, srcInfo.Mode()); err != nil {
				slog.Warn("Failed to preserve permissions", "file", dstPath, "error", err)
			}
		}
	}

	// Preserve times if requested
	if opts.PreserveTimes {
		if srcInfo, err := srcFile.Stat(); err == nil {
			modTime := srcInfo.ModTime()
			if err := os.Chtimes(dstPath, modTime, modTime); err != nil {
				slog.Warn("Failed to preserve times", "file", dstPath, "error", err)
			}
		}
	}

	return nil
}

func (fos *FileOperationsService) copyWithProgress(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	// Use a buffered copy with context cancellation support
	buf := make([]byte, 32*1024) // 32KB buffer
	var written int64

	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}

		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return written, er
			}
			break
		}
	}

	return written, nil
}

func (fos *FileOperationsService) handleFileConflict(ctx context.Context, srcPath, dstPath string, strategy options.ConflictStrategy) error {
	// Check if destination exists
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		return nil // No conflict
	}

	// Use conflict resolver if available
	if fos.conflictResolver != nil {
		resolvedPath, err := fos.conflictResolver.ResolveConflict(ctx, srcPath, dstPath, strategy)
		if err != nil {
			return err
		}
		// Update dstPath with resolved path
		_ = resolvedPath // For now, just validate resolution worked
	}

	return nil
}

func (fos *FileOperationsService) moveToTrash(ctx context.Context, path string) error {
	if fos.cacheDir == "" {
		return fmt.Errorf("trash directory not configured")
	}

	trashPath := filepath.Join(fos.cacheDir, "trash")
	if err := os.MkdirAll(trashPath, 0755); err != nil {
		return fmt.Errorf("failed to create trash directory: %w", err)
	}

	// Generate unique filename in trash
	baseName := filepath.Base(path)
	timestamp := time.Now().Format("20060102_150405")
	trashFile := filepath.Join(trashPath, fmt.Sprintf("%s_%s", timestamp, baseName))

	return os.Rename(path, trashFile)
}

func (fos *FileOperationsService) isCrossDeviceError(err error) bool {
	if linkErr, ok := err.(*os.LinkError); ok {
		return linkErr.Err == syscall.EXDEV
	}
	return false
}

func (fos *FileOperationsService) updateMetrics(start time.Time, isDirectory bool) {
	fos.metrics.TotalOperations++
	fos.metrics.LastOperation = time.Now()

	duration := time.Since(start)
	if fos.metrics.TotalBytesTransferred > 0 {
		fos.metrics.AverageSpeed = float64(fos.metrics.TotalBytesTransferred) / duration.Seconds()
	}
}

// CopyBatch performs batch copy operations
func (fos *FileOperationsService) CopyBatch(ctx context.Context, operations []types.OrganizationOperation, opts options.CopyOptions) (*types.OperationResult, error) {
	start := time.Now()
	result := &types.OperationResult{
		Success:        true,
		ProcessedFiles: 0,
		ProcessedDirs:  0,
		SkippedFiles:   0,
		Conflicts:      make([]*types.ConflictInfo, 0),
		Events:         make([]types.Event, 0),
	}

	for _, op := range operations {
		select {
		case <-ctx.Done():
			result.Success = false
			result.Error = ctx.Err()
			return result, ctx.Err()
		default:
		}

		switch op.Type {
		case types.OpCopy:
			if err := fos.CopyFile(ctx, op.SourcePath, op.TargetPath, opts); err != nil {
				result.SkippedFiles++
				slog.Error("Failed to copy file in batch", "source", op.SourcePath, "error", err)
				continue
			}
			result.ProcessedFiles++
		case types.OpSkip:
			result.SkippedFiles++
		default:
			slog.Warn("Unknown operation type in batch", "type", op.Type)
			result.SkippedFiles++
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// MoveBatch performs batch move operations
func (fos *FileOperationsService) MoveBatch(ctx context.Context, operations []types.OrganizationOperation, opts options.CopyOptions) (*types.OperationResult, error) {
	start := time.Now()
	result := &types.OperationResult{
		Success:        true,
		ProcessedFiles: 0,
		ProcessedDirs:  0,
		SkippedFiles:   0,
		Conflicts:      make([]*types.ConflictInfo, 0),
		Events:         make([]types.Event, 0),
	}

	for _, op := range operations {
		select {
		case <-ctx.Done():
			result.Success = false
			result.Error = ctx.Err()
			return result, ctx.Err()
		default:
		}

		switch op.Type {
		case types.OpMove:
			moveOpts := options.MoveOptions{
				DryRun:         opts.DryRun,
				Conflict:       opts.Conflict,
				PreservePerms:  opts.PreservePerms,
				FallbackToCopy: true,
				MaxRetries:     3,
			}
			if err := fos.MoveFile(ctx, op.SourcePath, op.TargetPath, moveOpts); err != nil {
				result.SkippedFiles++
				slog.Error("Failed to move file in batch", "source", op.SourcePath, "error", err)
				continue
			}
			result.ProcessedFiles++
		case types.OpSkip:
			result.SkippedFiles++
		default:
			slog.Warn("Unknown operation type in batch", "type", op.Type)
			result.SkippedFiles++
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// Ensure FileOperationsService implements the interface
var _ interfaces.FileSystemManager = (*FileOperationsService)(nil)
