package deskfs

import (
	"context"
	"errors"
	"file4you/internal/db"
	"file4you/internal/filesystem/trees"
	"file4you/internal/terminal"
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
	"github.com/rs/zerolog/log"

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
	NamesOnly          bool
	ForceSkipIgnore    bool
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

type DesktopFS struct {
	HomeDir          string
	Cwd              string
	CacheDir         string
	HomeDCDir        string
	WorkspaceManager *WorkspaceManager
	InstanceConfig   *DeskFSConfig
	term             *terminal.Terminal
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

func NewDesktopFS(term *terminal.Terminal, centralDB *db.CentralDBProvider) *DesktopFS {
	var err error
	cwd, err := os.Getwd()
	if err != nil {
		term.OutputErrorAndExit("Error getting current working directory: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		term.OutputErrorAndExit("Couldn't find home directory: %v", err)
	}

	homeDCDir := findDesktopCleaner(cwd)
	cacheDir := filepath.Join(homeDCDir, ".cache")

	assertHAndler := assert.NewAssertHandler()

	return &DesktopFS{
		HomeDir:          home,
		Cwd:              cwd,
		CacheDir:         cacheDir,
		HomeDCDir:        homeDCDir,
		WorkspaceManager: NewWorkspaceManager(centralDB, assertHAndler),
		term:             term,
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

func (dfs *DesktopFS) IndexDirectory(cfg *DeskFSConfig, params *FilePathParams) error {
	// Calculate the maximum depth of SourceDir
	maxDepth, err := CalculateMaxDepth(params.SourceDir)
	if err != nil {
		return fmt.Errorf("failed to calculate max depth: %w", err)
	}

	if err := dfs.buildTreeAndCache(params.SourceDir, params.Recursive, maxDepth); err != nil {
		return fmt.Errorf("failed to build directory tree: %w", err)
	}

	return nil
}

// Move or copy files based on the configuration
func (dfs *DesktopFS) EnhancedOrganize(cfg *DeskFSConfig, params *FilePathParams) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is canceled after function exits

	dfs.IndexDirectory(cfg, params)

	var wg sync.WaitGroup
	var once sync.Once
	errCh := make(chan error, len(dfs.WorkspaceManager.centralDB.DirectoryTree.Root.Files))

	// Add timeout
	go func() {
		select {
		case <-time.After(10 * time.Minute):
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	// Traverse and organize files based on config
	dfs.traverseAndOrganize(ctx, cancel, dfs.WorkspaceManager.centralDB.DirectoryTree.Root, cfg, params, &wg, errCh)

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

func (dfs *DesktopFS) InitConfig(optionalConfigPath string) {
	// Call NewConfig with the provided path (can be nil if no path is specified)
	config := NewIntermediateConfig(optionalConfigPath)
	slog.Debug(fmt.Sprintf("Loading configuration from path: %v\n", config))

	deskfsConfig := NewDeskFSConfig()

	// Build FileTypeTree
	deskfsConfig = deskfsConfig.BuildFileTypeTree(config)

	// Set the loaded configuration for this instance
	dfs.InstanceConfig = deskfsConfig
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
			log.Info().Msgf("Dry run: would create directory %s", dst)
		}

		// Copy each child directory
		for _, childDir := range node.Children {
			childDst := filepath.Join(dst, childDir.Path)
			if dryrun {
				log.Info().Msgf("Dry run: moving directory %s to %s", childDir.Path, childDst)
				// Continue processing other children
				continue
			}
			if err := dfs.Copy(childDir, childDst, recursive, remove, dryrun); err != nil {
				return err
			}
		}

		// Copy each file in the directory
		for _, fileNode := range node.Files {
			fileDst := filepath.Join(dst, fileNode.Path)
			if dryrun {
				log.Info().Msgf("Dry run: moving file %s to %s", fileNode.Path, fileDst)
				continue
			}
			if err := dfs.copyFile(fileNode, fileDst, remove, dryrun); err != nil {
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

// Helper function for copying a file
func (dfs *DesktopFS) copyFile(fileNode *trees.FileNode, dst string, remove bool, dryrun bool) error {

	if dryrun {
		slog.Info(fmt.Sprintf("Dry run: moving %s to %s\n", fileNode.Path, dst))
		return nil
	}

	srcFile, err := os.Open(fileNode.Path)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", fileNode.Path, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file %s to %s: %w", fileNode.Path, dst, err)
	}

	// Optionally remove the original file after copying
	if remove {
		if err := os.Remove(fileNode.Path); err != nil {
			return fmt.Errorf("failed to remove original file %s after copy: %w", fileNode.Path, err)
		}
	}
	return nil
}

// Move attempts to move a file or directory from src to dst.
// If a cross-device link error occurs, it falls back to copying and deleting the original.
func (dfs *DesktopFS) Move(node *trees.DirectoryNode, dst string, recursive bool, dryrun bool) error {

	if dryrun {
		log.Info().Msgf("Dry run: moving %s to %s\n", node.Path, dst)
		return nil
	}

	// Try renaming (moving) the directory node directly
	if err := os.Rename(node.Path, dst); err != nil {
		// If we encounter a cross-device link error, fall back to copy and delete
		if linkErr, ok := err.(*os.LinkError); ok && linkErr.Err == syscall.EXDEV {
			log.Warn().Msgf("Cross-device error detected: falling back to copy for %s\n", node.Path)
			if err := dfs.Copy(node, dst, recursive, true, dryrun); err != nil {
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

// buildTreeAndCache recursively builds a directory tree and populates a cache
func (dfs *DesktopFS) buildTreeAndCache(rootPath string, recursive bool, maxDepth int) error {
	// Add deferred cleanup
	defer func() {
		if dfs.WorkspaceManager.centralDB.DirectoryTree != nil {
			dfs.WorkspaceManager.centralDB.DirectoryTree.Cleanup()
		}
	}()

	// Initialize the DirectoryTree and Cache
	if dfs.WorkspaceManager.centralDB.DirectoryTree == nil {
		newDirectoryTree := trees.NewDirectoryTree(trees.WithRoot(rootPath))
		dfs.WorkspaceManager.centralDB.DirectoryTree = newDirectoryTree
	}

	//if dfs.WorkspaceManager.centralDB.DirectoryTree.Cache == nil {
	//	dfs.WorkspaceManager.centralDB.DirectoryTree.Cache = make(map[string]*trees.DirectoryNode)
	//}

	return dfs.buildTreeNodes(dfs.WorkspaceManager.centralDB.DirectoryTree.Root, recursive, maxDepth, 0)
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
		//var child *trees.DirectoryNode

		// Skip ignored files and directories
		if ignored != nil && ignored.MatchesPath(childPath) {
			slog.Info(fmt.Sprintf("Ignoring file %s\n", childPath))
			continue
		}

		if entry.IsDir() {
			childDir := trees.NewDirectoryNode(childPath, node)
			node.Children = append(node.Children, childDir)
			//dfs.WorkspaceManager.centralDB.DirectoryTree.SafeCacheSet(childPath, childDir)

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
			//dfs.WorkspaceManager.centralDB.DirectoryTree.SafeCacheSet(childPath, child)
		}
	}

	return nil
}

// traverseAndOrganize traverses the tree and organizes files based on the configuration
// uses goroutines for concurrent processing with mutexes for thread safety
func (dfs *DesktopFS) traverseAndOrganize(ctx context.Context, cancel context.CancelFunc, node *trees.DirectoryNode, cfg *DeskFSConfig, params *FilePathParams, wg *sync.WaitGroup, errCh chan error) {
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

			// Determine target folder without locking (no shared variable access)
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
			// Check for conflict without holding a lock
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
				fileErr = dfs.copyFile(fileNode, destPath, params.RemoveAfter, params.DryRun)
			} else {
				// For moving, construct a DirectoryNode with file path
				dummyNode := &trees.DirectoryNode{Path: fileNode.Path}
				fileErr = dfs.Move(dummyNode, destPath, false, params.DryRun)
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
func (dfs *DesktopFS) determineTargetFolder(ctx context.Context, fileNode *trees.FileNode, cfg *DeskFSConfig) (string, bool) {
	ext := fileNode.Extension

	path, found := dfs.findFolderForExtension(ctx, cfg.FileTypeTree.Root, ext)
	if found {
		slog.Info(fmt.Sprintf("File %s with extension %s mapped to path: %s\n", fileNode.Name, ext, path))
	} else {
		slog.Info(fmt.Sprintf("No mapping found for file %s with extension %s\n", fileNode.Name, ext))
	}
	return path, found
}

// Helper recursive function to search for the appropriate folder in the FileTypeTree.
func (dfs *DesktopFS) findFolderForExtension(ctx context.Context, node *trees.FileTypeNode, ext string) (string, bool) {
	// Traverse the tree to find a matching extension in the nodes
	if node.AllowsExtension(ext) {
		return buildPathFromNode(ctx, node), true
	}

	// Continue to search for extensions in children
	for _, child := range node.Children {
		if path, found := dfs.findFolderForExtension(ctx, child, ext); found {
			return path, true
		}
	}

	return "", false
}

// buildPathFromNode constructs the path from the root to the given node.
func buildPathFromNode(ctx context.Context, node *trees.FileTypeNode) string {
	// If this is the root node, start from its children
	if node.IsRoot() && len(node.Children) >= 1 {
		// Start from the first child to avoid adding "root" to the path
		node = node.Children[0]
	}

	assertHandler := assert.NewAssertHandler()
	assertHandler.SetExitFunc(func(int) {
		slog.Error("[Path Assertion Error]: assertion failure")
	})

	// Ensure that the node has a valid name
	if node.Name == "" {
		assertHandler.Never(ctx, fmt.Sprintf("Node has an invalid or empty name: %v", node), slog.Error)
	}

	pathSegments := []string{node.Name}
	for current := node.Parent; current != nil; current = current.Parent {
		assertHandler.Assert(ctx, current.Name != "", "Invalid node name detected", slog.Error)
		if current.IsRoot() {
			break // Skip "root" in the path
		}
		pathSegments = append([]string{current.Name}, pathSegments...)
	}
	assertHandler.Assert(ctx, node.IsRoot() || node.Parent != nil, "Root Node should not have a parent", slog.Error)

	finalPath := filepath.Join(pathSegments...)
	slog.Debug(fmt.Sprintf("Final constructed path (with case preserved): %s\n", finalPath))

	assertHandler.Assert(ctx, finalPath != "", "Constructed path should not be empty", slog.Error)

	return finalPath
}

func findDesktopCleaner(baseDir string) string {
	var dir string
	const devEnv = "development"
	const prodEnv = "production"
	const folderName = ".file4you"
	const env = "DESKTOP_CLEANER_ENV"

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
