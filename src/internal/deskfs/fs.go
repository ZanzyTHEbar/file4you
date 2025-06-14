package deskfs

import (
	"context"
	"errors"
	"file4you/internal/config"
	"file4you/internal/db"
	"file4you/internal/filesystem/trees"

	// "file4you/internal/terminal" // No longer used
	"file4you/internal/ui"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ZanzyTHEbar/assert-lib"

	ignore "github.com/sabhiram/go-gitignore"
)

type ConflictResolutionType string

const (
	Overwrite    ConflictResolutionType = "overwrite"
	Skip         ConflictResolutionType = "skip"
	RenameSuffix ConflictResolutionType = "rename"
)

type FilePathParams struct {
	RemoveAfter        bool
	Recursive          bool
	MaxDepth           int
	GitEnabled         bool
	CopyFiles          bool
	SourceDir          string
	TargetDir          string
	DryRun             bool
	ConflictResolution ConflictResolutionType // "overwrite", "skip", or "rename"
}

// Validate checks if the FilePathParams are valid
func (p *FilePathParams) Validate() error {
	if p.SourceDir == "" || p.TargetDir == "" {
		return fmt.Errorf("source and target directories must be specified")
	}
	if p.MaxDepth < 0 {
		return fmt.Errorf("max depth cannot be negative")
	}
	return nil
}

// DesktopFS is the main filesystem manager for the file4you application.
// It provides functionality for organizing and managing files across workspaces.
type DesktopFS struct {
	HomeDir          string
	Cwd              string
	CacheDir         string
	HomeDCDir        string
	WorkspaceManager *WorkspaceManager
	InstanceConfig   *config.File4YouConfig
	term             ui.Interactor
	gitMutex         sync.Mutex
}

// NewFilePathParams initializes FilePathParams with sensible defaults.
func NewFilePathParams() *FilePathParams {
	return &FilePathParams{
		SourceDir:          "",
		TargetDir:          "",
		Recursive:          true,
		CopyFiles:          false,
		RemoveAfter:        false,
		DryRun:             false,
		ConflictResolution: "rename",
	}
}

func NewDesktopFS(interactor ui.Interactor, centralDB db.ICentralDBProvider) *DesktopFS {
	var err error
	cwd, err := os.Getwd()
	if err != nil {
		interactor.Error("Error getting current working directory", err)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		interactor.Error("Couldn't find home directory", err)
		os.Exit(1)
	}

	// Use AppConfig for cache directory
	cacheDir := config.AppConfig.File4You.CacheDir
	if cacheDir == "" { // Fallback if not set by config
		homeDCDir := findDesktopCleaner(cwd)
		if homeDCDir != "" && homeDCDir != cwd {
			cacheDir = filepath.Join(homeDCDir, ".cache")
		} else {
			cacheDir = filepath.Join(home, ".file4you", ".cache")
		}
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		interactor.Error(fmt.Sprintf("Error creating cache directory %s", cacheDir), err)
		os.Exit(1)
	}

	assertHHandler := assert.NewAssertHandler()

	return &DesktopFS{
		HomeDir:  home,
		Cwd:      cwd,
		CacheDir: cacheDir,
		// HomeDCDir will be determined by config or other logic if still needed
		WorkspaceManager: NewWorkspaceManager(centralDB, assertHHandler),
		InstanceConfig:   &config.AppConfig.File4You, // Use loaded global config
		term:             interactor,                 // Assign interactor
	}
}

// CalculateMaxDepth calculates the maximum depth of the directory structure in `sourceDir`.
func CalculateMaxDepth(sourceDir string) (int, error) {
	if sourceDir == "" {
		return 0, fmt.Errorf("source directory path cannot be empty")
	}

	// Initialize the maximum depth counter
	maxDepth := 0

	// Walk through the directory structure of sourceDir
	err := filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate depth relative to sourceDir
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Calculate the depth by counting separators in the relative path
		if relPath != "." { // Skip the root itself
			depth := strings.Count(relPath, string(os.PathSeparator)) + 1
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("error calculating max depth: %w", err)
	}

	return maxDepth, nil
}

func (dfs *DesktopFS) IndexDirectory(cfg *config.File4YouConfig, params *FilePathParams) error { // Changed to use new config type
	var actualMaxDepth int
	var err error

	if params.MaxDepth > 0 {
		actualMaxDepth = params.MaxDepth
		slog.Debug(fmt.Sprintf("Using user-defined MaxDepth: %d", actualMaxDepth))
	} else {
		slog.Debug("Calculating MaxDepth from source directory structure.")
		actualMaxDepth, err = CalculateMaxDepth(params.SourceDir)
		if err != nil {
			return fmt.Errorf("failed to calculate max depth: %w", err)
		}
		slog.Debug(fmt.Sprintf("Calculated MaxDepth: %d", actualMaxDepth))
	}

	if err := dfs.buildTreeAndCache(params.SourceDir, params.Recursive, actualMaxDepth); err != nil {
		return fmt.Errorf("failed to build directory tree: %w", err)
	}

	return nil
}

// Move or copy files based on the configuration
func (dfs *DesktopFS) EnhancedOrganize(cfg *config.File4YouConfig, params *FilePathParams) error { // Changed to use new config type
	// Validate that source and target directories exist
	if _, err := os.Stat(params.SourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", params.SourceDir)
	}

	if _, err := os.Stat(params.TargetDir); os.IsNotExist(err) {
		return fmt.Errorf("target directory does not exist: %s", params.TargetDir)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is canceled after function exits

	dfs.IndexDirectory(cfg, params)

	var wg sync.WaitGroup
	var once sync.Once
	dirTree := dfs.WorkspaceManager.centralDB.GetDirectoryTree()

	// Check if the DirectoryTree or its Root is nil, and initialize if needed
	if dirTree == nil || dirTree.Root == nil {
		// Initialize the DirectoryTree with the source directory if it doesn't exist or has no root
		dirTree = trees.NewDirectoryTree(trees.WithRoot(params.SourceDir))
		dfs.WorkspaceManager.centralDB.SetDirectoryTree(dirTree)
	}

	var errChSize int
	if dirTree.Root != nil && dirTree.Root.Files != nil {
		errChSize = len(dirTree.Root.Files)
	} else {
		errChSize = 10 // Default capacity if Files is nil
	}
	errCh := make(chan error, errChSize)

	// Determine timeout duration from config, with a fallback default
	timeoutMinutes := 10 // Default fallback
	if cfg != nil && cfg.OrganizeTimeoutMinutes > 0 {
		timeoutMinutes = cfg.OrganizeTimeoutMinutes
	}
	slog.Debug(fmt.Sprintf("EnhancedOrganize timeout set to %d minutes", timeoutMinutes))

	// Add timeout
	go func() {
		select {
		case <-time.After(time.Duration(timeoutMinutes) * time.Minute):
			slog.Warn(fmt.Sprintf("EnhancedOrganize timed out after %d minutes", timeoutMinutes))
			cancel()
		case <-ctx.Done():
			return
		}
	}() // Traverse and organize files based on config
	if dirTree != nil && dirTree.Root != nil {
		dfs.traverseAndOrganize(ctx, cancel, dirTree.Root, cfg, params, &wg, errCh)
	} else {
		// Log the error but don't stop execution
		slog.Error("Cannot traverse directory tree: tree or root is nil")
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		once.Do(func() { close(errCh) })
	}()

	if err, ok := <-errCh; ok {
		cancel() // Cancel ongoing operations
		return fmt.Errorf("failed to organize files: %w", err)
	}

	// Commit changes if Git is enabled
	if params.GitEnabled {
		if err := dfs.GitAddAndCommit(dfs.Cwd, fmt.Sprintf("Organized files for %s", dfs.Cwd)); err != nil {
			return fmt.Errorf("failed to commit to git: %w", err)
		}

		// Pop the stash if any changes were stashed before organizing
		if err := dfs.GitStashPop(dfs.Cwd, true); err != nil {
			return fmt.Errorf("error popping git stash after organizing: %w", err)
		}
	}

	return nil
}

func (dfs *DesktopFS) InitConfig(optionalConfigPath string, interactor ui.Interactor) {
	// Configuration is now loaded globally via config.LoadConfig()
	// We just need to ensure dfs.InstanceConfig points to the loaded File4You specific part.
	dfs.InstanceConfig = &config.AppConfig.File4You

	// The old logic for loading/creating intermediate config and building FileTypeTree
	// is removed as file type mapping is no longer the primary way of organization.
	// If any specific initialization based on config is still needed here, it would be added.
	// For now, we assume the global AppConfig is sufficient.
	slog.Debug("DesktopFS using globally loaded configuration.")
}

func (dfs *DesktopFS) GetDesktopCleanerIgnore(dir string) (*ignore.GitIgnore, error) {
	ignorePath := filepath.Join(dir, ".file4you-ignore")

	if _, err := os.Stat(ignorePath); err == nil {
		ignored, err := ignore.CompileIgnoreFile(ignorePath)

		if err != nil {
			return nil, fmt.Errorf("error reading .file4you-ignore file: %s", err)
		}

		return ignored, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error checking for .file4you-ignore file: %s", err)
	}

	return nil, nil
}

// Copy copies a file or directory to the destination path.
// It uses recursion for directories if the recursive flag is enabled.
func (dfs *DesktopFS) Copy(node *trees.DirectoryNode, dst string, recursive bool, remove bool, dryrun bool) error {
	if len(node.Children) > 0 || len(node.Files) > 0 { // Check if node is a directory
		if !recursive {
			return fmt.Errorf("source is a directory, use recursive flag to copy directories")
		}

		// Ensure destination directory exists
		if !dryrun {
			if err := os.MkdirAll(dst, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dst, err)
			}
		} else {
			slog.Info(fmt.Sprintf("Dry run: would create directory %s", dst))
		}

		// Copy each child directory
		for _, childDir := range node.Children {
			childDst := filepath.Join(dst, childDir.Path)
			if dryrun {
				slog.Info(fmt.Sprintf("Dry run: moving directory %s to %s", childDir.Path, childDst))
				// Continue processing other children
				continue
			}
			if err := dfs.Copy(childDir, childDst, recursive, remove, dryrun); err != nil {
				return err
			}
		}

		// Copy each file in the directory
		for _, fileNode := range node.Files {
			fileSrcPath := fileNode.Path                                    // Source path of the file
			fileDstPath := filepath.Join(dst, filepath.Base(fileNode.Path)) // Destination path for the file
			if dryrun {
				slog.Info(fmt.Sprintf("Dry run: copying file %s to %s", fileSrcPath, fileDstPath))
				continue
			}
			if err := dfs.copyFile(fileSrcPath, fileDstPath, remove, dryrun); err != nil { // Pass remove from parent Copy
				return err
			}
		}

		// Optionally remove the original directory after copying
		if remove && !dryrun {
			return os.RemoveAll(node.Path)
		}
		return nil
	}
	return fmt.Errorf("node has no files or directories to copy")
}

// Helper function for copying a single file
func (dfs *DesktopFS) copyFile(srcPath string, dstPath string, removeOriginal bool, dryRun bool) error {
	if dryRun {
		operation := "copying"
		if removeOriginal {
			operation = "moving (by copy-then-delete)"
		}
		slog.Info(fmt.Sprintf("Dry run: %s file %s to %s", operation, srcPath, dstPath))
		return nil
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file %s to %s: %w", srcPath, dstPath, err)
	}

	// Optionally remove the original file after copying
	if removeOriginal {
		// Ensure files are closed before attempting removal
		dstFile.Close()
		srcFile.Close() // Must be closed before os.Remove can succeed on some OS (e.g. Windows)
		if err := os.Remove(srcPath); err != nil {
			return fmt.Errorf("failed to remove original file %s after copy: %w", srcPath, err)
		}
	}
	return nil
}

// moveFile attempts to move a single file from srcPath to dstPath.
// If a cross-device link error occurs, it falls back to copying and then deleting the original file.
func (dfs *DesktopFS) moveFile(srcPath, dstPath string, dryRun bool) error {
	if dryRun {
		slog.Info(fmt.Sprintf("Dry run: moving file %s to %s", srcPath, dstPath))
		return nil
	}

	// Try renaming (moving) the file directly
	if err := os.Rename(srcPath, dstPath); err != nil {
		// If we encounter a cross-device link error, fall back to copy and delete
		if linkErr, ok := err.(*os.LinkError); ok && linkErr.Err == syscall.EXDEV {
			slog.Warn(fmt.Sprintf("Cross-device error detected for file %s: falling back to copy and delete", srcPath))

			// Fallback: Copy the file
			srcFile, openErr := os.Open(srcPath)
			if openErr != nil {
				return fmt.Errorf("failed to open source file %s for copy-fallback: %w", srcPath, openErr)
			}
			defer srcFile.Close()

			dstFile, createErr := os.Create(dstPath)
			if createErr != nil {
				return fmt.Errorf("failed to create destination file %s for copy-fallback: %w", dstPath, createErr)
			}
			defer dstFile.Close()

			if _, copyErr := io.Copy(dstFile, srcFile); copyErr != nil {
				return fmt.Errorf("failed to copy file %s to %s during copy-fallback: %w", srcPath, dstPath, copyErr)
			}
			// Close files before attempting to remove source
			dstFile.Close()
			srcFile.Close()

			// Fallback: Delete the original file
			if removeErr := os.Remove(srcPath); removeErr != nil {
				return fmt.Errorf("failed to remove original file %s after copy-fallback: %w", srcPath, removeErr)
			}
			return nil
		}
		// Not a cross-device error, or another type of LinkError
		return fmt.Errorf("failed to move file %s to %s: %w", srcPath, dstPath, err)
	}
	return nil
}

// Move attempts to move a file or directory from src to dst.
// If a cross-device link error occurs, it falls back to copying and deleting the original.
// This function is primarily for directories. For single files, use moveFile.
func (dfs *DesktopFS) Move(node *trees.DirectoryNode, dst string, recursive bool, dryRun bool) error {

	if dryRun {
		slog.Info(fmt.Sprintf("Dry run: moving %s to %s", node.Path, dst))
		return nil
	}

	// Try renaming (moving) the directory node directly
	if err := os.Rename(node.Path, dst); err != nil {
		// If we encounter a cross-device link error, fall back to copy and delete
		if linkErr, ok := err.(*os.LinkError); ok && linkErr.Err == syscall.EXDEV {
			slog.Warn(fmt.Sprintf("Cross-device error detected: falling back to copy for %s", node.Path))
			if err := dfs.Copy(node, dst, recursive, true, dryRun); err != nil { // Corrected dryrun to dryRun
				return fmt.Errorf("failed to copy file for cross-device move: %w", err)
			}
			return nil
		} else {
			return fmt.Errorf("failed to move directory: %w", err)
		}
	}
	return nil
}

// MoveToTrash moves a file or directory to the  trash (cache) directory
func (dfs *DesktopFS) MoveToTrash(node *trees.DirectoryNode) error {
	dst := filepath.Join(dfs.CacheDir, filepath.Base(node.Path))
	return os.Rename(node.Path, dst)
}

// buildTreeAndCache builds a directory tree using concurrent traversal for optimal performance
func (dfs *DesktopFS) buildTreeAndCache(rootPath string, recursive bool, maxDepth int) error {
	// Add deferred cleanup
	defer func() {
		dirTree := dfs.WorkspaceManager.centralDB.GetDirectoryTree()
		if dirTree != nil {
			dirTree.Cleanup()
		}
	}()

	// Initialize the DirectoryTree if it doesn't exist
	if dfs.WorkspaceManager.centralDB.GetDirectoryTree() == nil {
		newDirectoryTree := trees.NewDirectoryTree(trees.WithRoot(rootPath))
		dfs.WorkspaceManager.centralDB.SetDirectoryTree(newDirectoryTree)
	}

	// Use concurrent traverser for high-performance directory scanning
	ctx := context.Background()
	traverser := NewConcurrentTraverser(ctx)
	defer traverser.Cleanup()

	slog.Info(fmt.Sprintf("Starting concurrent traversal of %s (recursive: %v, maxDepth: %d)", rootPath, recursive, maxDepth))

	rootNode, err := traverser.TraverseDirectory(rootPath, recursive, maxDepth, dfs)
	if err != nil {
		return fmt.Errorf("concurrent traversal failed: %w", err)
	}

	// Update the directory tree with the traversed root
	dirTree := dfs.WorkspaceManager.centralDB.GetDirectoryTree()
	dirTree.Root = rootNode

	slog.Info("Concurrent traversal completed successfully")
	return nil
}

// Recursive helper to populate the directory tree with DirectoryNode entries
func (dfs *DesktopFS) buildTreeNodes(node *trees.DirectoryNode, recursive bool, maxDepth int, currentDepth int) error {
	// Check if the current depth exceeds the maxDepth
	if currentDepth > maxDepth {
		slog.Warn(fmt.Sprintf("Max depth of %d reached at %s. Skipping deeper levels.\n", maxDepth, node.Path))
		return nil
	}

	entries, err := os.ReadDir(node.Path)
	if err != nil {
		return err
	}

	ignored, err := dfs.GetDesktopCleanerIgnore(node.Path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		childPath := filepath.Join(node.Path, entry.Name())

		// Skip ignored files and directories
		if ignored != nil && ignored.MatchesPath(childPath) {
			slog.Info(fmt.Sprintf("Ignoring file %s\n", childPath))
			continue
		}

		if entry.IsDir() {
			childDir := trees.NewDirectoryNode(childPath, node)
			node.Children = append(node.Children, childDir)

			if !recursive {
				continue
			}

			if err := dfs.buildTreeNodes(childDir, recursive, maxDepth, currentDepth+1); err != nil {
				return err
			}
		} else {
			entryInfo, err := entry.Info()
			if err != nil {
				slog.Warn(fmt.Sprintf("Error getting file info for %s: %v", entry.Name(), err))
			}

			childFile := &trees.FileNode{
				Path:      childPath,
				Name:      entry.Name(),
				Extension: strings.ToLower(filepath.Ext(entry.Name())),
				Metadata:  trees.NewMetadata(entryInfo),
			}
			_ = node.AddFile(childFile)
		}
	}

	return nil
}

// traverseAndOrganize traverses the tree and organizes files based on the configuration
// uses goroutines for concurrent processing with mutexes for thread safety
func (dfs *DesktopFS) traverseAndOrganize(ctx context.Context, cancel context.CancelFunc, node *trees.DirectoryNode, cfg *config.File4YouConfig, params *FilePathParams, wg *sync.WaitGroup, errCh chan error) { // Changed to use new config type
	// Process files concurrently without a coarse-grained mutex
	for _, fileNode := range node.Files {
		wg.Add(1)
		go func(fileNode *trees.FileNode) {
			defer wg.Done()

			// Check for cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			targetDir, found := dfs.determineTargetFolder(ctx, fileNode, cfg)
			if !found {
				slog.Warn(fmt.Sprintf("Skipping file %s as no target path found", fileNode.Name))
				return
			}

			// Construct destination paths
			destDir := filepath.Join(params.TargetDir, targetDir)
			// Ensure target directory exists
			if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
				select {
				case errCh <- fmt.Errorf("failed to create target directory %s: %w", destDir, err):
					cancel()
				default:
				}
				return
			}

			destPath := filepath.Join(destDir, filepath.Base(fileNode.Path))
			if _, err := os.Stat(destPath); err == nil {
				switch params.ConflictResolution {
				case Overwrite:
					slog.Info(fmt.Sprintf("Overwriting existing file: %s", destPath))
				case Skip:
					slog.Info(fmt.Sprintf("Skipping file to avoid conflict: %s", destPath))
					return
				case RenameSuffix:
					destPath = generateUniqueFilename(destPath)
					slog.Info(fmt.Sprintf("Renaming file to avoid conflict: %s", destPath))
				default:
					slog.Info(fmt.Sprintf("Unknown conflict resolution type: %s, skipping file %s", params.ConflictResolution, fileNode.Path))
					return
				}
			}

			slog.Info(fmt.Sprintf("Moving file %s to %s", fileNode.Path, destPath))
			// Copy or move the file based on params
			var fileErr error
			if params.CopyFiles {
				// If CopyFiles is true, it implies we want a copy.
				// params.RemoveAfter will determine if the original is deleted after copy.
				fileErr = dfs.copyFile(fileNode.Path, destPath, params.RemoveAfter, params.DryRun)
			} else {
				// If CopyFiles is false, it implies a direct move.
				fileErr = dfs.moveFile(fileNode.Path, destPath, params.DryRun)
			}

			if fileErr != nil {
				slog.Error(fmt.Sprintf("Error moving file %s: %v", fileNode.Path, fileErr))
				select {
				case errCh <- fmt.Errorf("file operation failed for %s: %w", fileNode.Path, fileErr):
					cancel()
				default:
				}
				return
			}
		}(fileNode)
	}

	// Process child directories recursively
	for _, childDir := range node.Children {
		if params.Recursive {
			dfs.traverseAndOrganize(ctx, cancel, childDir, cfg, params, wg, errCh)
		}
	}
}

// determineTargetFolder traverses the FileTypeTree in DeskFSConfig to find the appropriate folder
// based on the file's extension. It returns the path to the target folder if a match is found.
func (dfs *DesktopFS) determineTargetFolder(ctx context.Context, fileNode *trees.FileNode, cfg *config.File4YouConfig) (string, bool) {
	ext := strings.ToLower(filepath.Ext(fileNode.Name))
	// The FileTypeTree is no longer part of the config as per the consolidation.
	// The logic for determining the target folder will need to be revised
	// based on the new agent-driven approach.
	// For now, returning a default/placeholder or an error.
	// This part of the code needs to be re-evaluated based on how agents decide file placement.
	slog.Warn(fmt.Sprintf("determineTargetFolder: FileTypeTree logic removed. File extension '%s' for '%s' cannot be automatically mapped. Agent intervention required.", ext, fileNode.Path))
	// Placeholder: return the source directory, indicating no specific rule was found.
	// In a real scenario, this might involve querying an agent or using other metadata.
	return filepath.Dir(fileNode.Path), false // Returning original path and false, as no rule applied
}

// buildPathFromNode constructs the path from the root to the given node.
// If the given node is the root itself, an empty path is returned.
func buildPathFromNode(ctx context.Context, node *trees.FileTypeNode) string {
	if node == nil || node.IsRoot() {
		slog.Debug("buildPathFromNode: node is root or nil, returning empty path")
		return "" // Represents the base of the target directory
	}

	assertHandler := assert.NewAssertHandler()
	assertHandler.SetExitFunc(func(int) {
		slog.Error("[Path Assertion Error]: assertion failure in buildPathFromNode")
	})

	// Ensure that the node has a valid name (it's not root, so it should have one if part of a path)
	if node.Name == "" {
		// This case should ideally not be reached if nodes forming paths always have names.
		assertHandler.Never(ctx, fmt.Sprintf("buildPathFromNode: node has an invalid or empty name: %v", node), slog.Error)
		slog.Warn(fmt.Sprintf("buildPathFromNode: encountered node with empty name: %v. Path construction might be incorrect.", node))
		// Depending on desired behavior, could return error or continue carefully.
		// For now, let it proceed, path will just miss this segment's name.
	}

	pathSegments := []string{}
	current := node
	// Iterate upwards from the current node until we reach the root or nil parent
	for current != nil && !current.IsRoot() {
		// current.Name should not be empty if it's part of a valid path structure.
		// The root node (named "root") is skipped by !current.IsRoot().
		if current.Name == "" {
			// This indicates a malformed tree structure if a non-root node in the path has no name.
			assertHandler.Never(ctx, fmt.Sprintf("buildPathFromNode: encountered node with empty name in path hierarchy: %v", current), slog.Error)
			slog.Warn(fmt.Sprintf("buildPathFromNode: node in path hierarchy has empty name: %v. Skipping in path.", current))
		} else {
			pathSegments = append([]string{current.Name}, pathSegments...)
		}
		current = current.Parent
	}

	finalPath := filepath.Join(pathSegments...)
	slog.Debug(fmt.Sprintf("Final constructed path (with case preserved): %s (from node: %s)", finalPath, node.Name))
	// The assertion `finalPath != ""` might be too strict if an empty path is valid (e.g., for extensions on root,
	// but that case is handled by the initial `if node.IsRoot()` check returning `""`).
	// If `node` is not root, `finalPath` should ideally not be empty if `node.Name` is not empty.
	if !node.IsRoot() && node.Name != "" && finalPath == "" {
		// This could happen if node.Name was the only segment and it was empty, or pathSegments ended up empty.
		slog.Warn(fmt.Sprintf("buildPathFromNode: constructed empty path for non-root node: %s", node.Name))
	}

	return finalPath
}

func findDesktopCleaner(baseDir string) string {
	var dir string
	const devEnv = "development"
	const prodEnv = "production"
	const folderName = ".file4you"
	const env = "FILE4YOU_ENV"

	envValue, envSet := os.LookupEnv(env)

	if !envSet {
		return ""
	}

	dir = filepath.Join(baseDir, folderName+"-"+envValue)
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return baseDir
	}

	return dir
}

func generateUniqueFilename(path string) string {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := filepath.Base(path[:len(path)-len(ext)])

	// Iterate to find an available filename
	for i := 1; ; i++ {
		newPath := filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}

// Backup creates a backup of the DesktopFS configuration files and workspace data.
// It returns the path to the backup directory and any error that occurred.
func (dfs *DesktopFS) Backup() (string, error) {
	// Create a backup directory with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(dfs.HomeDCDir, "backups", fmt.Sprintf("deskfs_backup_%s", timestamp))

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("could not create backup directory: %v", err)
	}

	// Backup the workspace configurations
	if dfs.WorkspaceManager != nil && dfs.WorkspaceManager.centralDB != nil {
		// Backup workspace configurations
		workspaces, err := dfs.WorkspaceManager.centralDB.ListWorkspaces()
		if err == nil {
			for _, workspace := range workspaces {
				config, err := dfs.WorkspaceManager.centralDB.GetWorkspaceConfig(workspace.ID)
				if err == nil {
					configPath := filepath.Join(backupDir, fmt.Sprintf("workspace_%s.json", workspace.ID))
					if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
						slog.Error("Failed to backup workspace config", "workspace", workspace.ID.String(), "error", err)
					} else {
						slog.Info("Workspace config backed up successfully", "workspace", workspace.ID.String(), "path", configPath)
					}
				}
			}
		}
	}

	// Backup key file paths and metadata
	infoFile := filepath.Join(backupDir, "deskfs_info.txt")
	info := fmt.Sprintf("DesktopFS Backup\nTimestamp: %s\n"+
		"Home Directory: %s\n"+
		"Current Working Directory: %s\n"+
		"Cache Directory: %s\n"+
		"Home DC Directory: %s\n",
		timestamp, dfs.HomeDir, dfs.Cwd, dfs.CacheDir, dfs.HomeDCDir)

	if err := os.WriteFile(infoFile, []byte(info), 0644); err != nil {
		slog.Error("Failed to write backup info file", "error", err)
	}

	slog.Info("DesktopFS backup completed", "path", backupDir)
	return backupDir, nil
}
