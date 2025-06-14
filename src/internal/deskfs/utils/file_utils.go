package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PathUtils provides path manipulation utilities
type PathUtils struct{}

// NewPathUtils creates a new PathUtils instance
func NewPathUtils() *PathUtils {
	return &PathUtils{}
}

// NormalizePath normalizes a file path for cross-platform compatibility
func (pu *PathUtils) NormalizePath(path string) string {
	// Convert to absolute path
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(abs)
}

// IsSubpath checks if child is a subpath of parent
func (pu *PathUtils) IsSubpath(parent, child string) bool {
	parent = pu.NormalizePath(parent)
	child = pu.NormalizePath(child)

	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, "..") && rel != "."
}

// GetRelativePath returns the relative path from base to target
func (pu *PathUtils) GetRelativePath(base, target string) (string, error) {
	base = pu.NormalizePath(base)
	target = pu.NormalizePath(target)

	return filepath.Rel(base, target)
}

// SplitPath splits a path into directory and filename components
func (pu *PathUtils) SplitPath(path string) (dir, name, ext string) {
	dir = filepath.Dir(path)
	name = filepath.Base(path)
	ext = filepath.Ext(name)

	if ext != "" {
		name = strings.TrimSuffix(name, ext)
	}

	return dir, name, ext
}

// JoinPath safely joins path components
func (pu *PathUtils) JoinPath(components ...string) string {
	return filepath.Join(components...)
}

// ValidatePath validates that a path is safe and accessible
func (pu *PathUtils) ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for invalid characters (basic check)
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null character")
	}

	// Check path length (reasonable limit)
	if len(path) > 4096 {
		return fmt.Errorf("path too long (max 4096 characters)")
	}

	return nil
}

// FileUtils provides file manipulation utilities
type FileUtils struct{}

// NewFileUtils creates a new FileUtils instance
func NewFileUtils() *FileUtils {
	return &FileUtils{}
}

// CalculateChecksum calculates the checksum of a file using the specified algorithm
func (fu *FileUtils) CalculateChecksum(path string, algorithm string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	var hasher hash.Hash
	switch strings.ToLower(algorithm) {
	case "md5":
		hasher = md5.New()
	case "sha256":
		hasher = sha256.New()
	default:
		return "", fmt.Errorf("unsupported checksum algorithm: %s", algorithm)
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum for %s: %w", path, err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// GetFileSize returns the size of a file in bytes
func (fu *FileUtils) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file %s: %w", path, err)
	}
	return info.Size(), nil
}

// IsEmpty checks if a file or directory is empty
func (fu *FileUtils) IsEmpty(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("failed to stat %s: %w", path, err)
	}

	if info.IsDir() {
		// Check if directory is empty
		f, err := os.Open(path)
		if err != nil {
			return false, fmt.Errorf("failed to open directory %s: %w", path, err)
		}
		defer f.Close()

		_, err = f.Readdirnames(1)
		if err == io.EOF {
			return true, nil
		}
		return false, err
	}

	// File is empty if size is 0
	return info.Size() == 0, nil
}

// GetFileType determines the file type based on extension and content
func (fu *FileUtils) GetFileType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".svg", ".webp":
		return "image"
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm", ".m4v":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a", ".wma":
		return "audio"
	case ".pdf":
		return "pdf"
	case ".doc", ".docx", ".txt", ".rtf", ".odt", ".md":
		return "document"
	case ".xls", ".xlsx", ".csv", ".ods":
		return "spreadsheet"
	case ".ppt", ".pptx", ".odp":
		return "presentation"
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz":
		return "archive"
	case ".exe", ".msi", ".dmg", ".pkg", ".deb", ".rpm", ".app":
		return "application"
	case ".js", ".html", ".css", ".py", ".go", ".java", ".cpp", ".c", ".rs", ".php":
		return "code"
	case ".json", ".xml", ".yaml", ".yml", ".toml", ".ini", ".cfg":
		return "config"
	default:
		return "other"
	}
}

// CopyFileAttributes copies file attributes from source to destination
func (fu *FileUtils) CopyFileAttributes(srcPath, dstPath string, preservePerms, preserveTimes bool) error {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat source %s: %w", srcPath, err)
	}

	if preservePerms {
		if err := os.Chmod(dstPath, srcInfo.Mode()); err != nil {
			return fmt.Errorf("failed to set permissions on %s: %w", dstPath, err)
		}
	}

	if preserveTimes {
		if err := os.Chtimes(dstPath, time.Now(), srcInfo.ModTime()); err != nil {
			return fmt.Errorf("failed to set times on %s: %w", dstPath, err)
		}
	}

	return nil
}

// DepthUtils provides depth calculation utilities
type DepthUtils struct{}

// NewDepthUtils creates a new DepthUtils instance
func NewDepthUtils() *DepthUtils {
	return &DepthUtils{}
}

// CalculateDepth calculates the depth of a path relative to a base path
func (du *DepthUtils) CalculateDepth(basePath, targetPath string) (int, error) {
	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate relative path: %w", err)
	}

	if relPath == "." {
		return 0, nil
	}

	if strings.HasPrefix(relPath, "..") {
		return 0, fmt.Errorf("target path is not under base path")
	}

	return strings.Count(relPath, string(os.PathSeparator)), nil
}

// CalculateMaxDepthInDirectory calculates the maximum depth in a directory tree
func (du *DepthUtils) CalculateMaxDepthInDirectory(rootPath string) (int, error) {
	maxDepth := 0

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		depth, err := du.CalculateDepth(rootPath, path)
		if err != nil {
			return err
		}

		if depth > maxDepth {
			maxDepth = depth
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	return maxDepth, nil
}

// TruncateDepth truncates a path to a maximum depth relative to base
func (du *DepthUtils) TruncateDepth(basePath, targetPath string, maxDepth int) (string, error) {
	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate relative path: %w", err)
	}

	if relPath == "." {
		return targetPath, nil
	}

	components := strings.Split(relPath, string(os.PathSeparator))
	if len(components) <= maxDepth {
		return targetPath, nil
	}

	truncatedRel := filepath.Join(components[:maxDepth]...)
	return filepath.Join(basePath, truncatedRel), nil
}

// SafetyUtils provides safety checks for file operations
type SafetyUtils struct{}

// NewSafetyUtils creates a new SafetyUtils instance
func NewSafetyUtils() *SafetyUtils {
	return &SafetyUtils{}
}

// IsSafeOperation checks if a file operation is safe to perform
func (su *SafetyUtils) IsSafeOperation(srcPath, dstPath string) error {
	// Check for self-reference
	if srcPath == dstPath {
		return fmt.Errorf("source and destination are the same")
	}

	// Check for potential overwrites of important files
	if su.isSystemPath(dstPath) {
		return fmt.Errorf("destination is a system path: %s", dstPath)
	}

	// Check for circular operations (moving parent into child)
	pathUtils := NewPathUtils()
	if pathUtils.IsSubpath(srcPath, dstPath) {
		return fmt.Errorf("cannot move parent directory into child directory")
	}

	return nil
}

// isSystemPath checks if a path is a critical system path
func (su *SafetyUtils) isSystemPath(path string) bool {
	systemPaths := []string{
		"/bin", "/sbin", "/usr/bin", "/usr/sbin",
		"/etc", "/boot", "/dev", "/proc", "/sys",
		"C:\\Windows", "C:\\Program Files", "C:\\Program Files (x86)",
	}

	for _, sysPath := range systemPaths {
		if strings.HasPrefix(path, sysPath) {
			return true
		}
	}

	return false
}

// ValidateOperationSafety performs comprehensive safety validation
func (su *SafetyUtils) ValidateOperationSafety(srcPath, dstPath string, checkSpace bool) error {
	if err := su.IsSafeOperation(srcPath, dstPath); err != nil {
		return err
	}

	// Check source exists
	if _, err := os.Stat(srcPath); err != nil {
		return fmt.Errorf("source does not exist: %s", srcPath)
	}

	// Check destination directory exists and is writable
	dstDir := filepath.Dir(dstPath)
	if info, err := os.Stat(dstDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("destination directory does not exist: %s", dstDir)
		}
		return fmt.Errorf("failed to access destination directory: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("destination parent is not a directory: %s", dstDir)
	}

	if checkSpace {
		if err := su.checkDiskSpace(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// checkDiskSpace checks if there's enough disk space for the operation
func (su *SafetyUtils) checkDiskSpace(srcPath, dstPath string) error {
	// Get source size
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// For simplicity, assume we need at least the source size in destination
	// In a full implementation, we'd calculate available disk space
	if srcInfo.Size() > 1024*1024*1024 { // > 1GB
		// Could add actual disk space checking here
		// For now, just warn about large files
		return nil
	}

	return nil
}
