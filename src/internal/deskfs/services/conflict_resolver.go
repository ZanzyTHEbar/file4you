package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"file4you/internal/deskfs/interfaces"
	"file4you/internal/deskfs/options"
	"file4you/internal/deskfs/types"
	"file4you/internal/filesystem/trees"
)

// ConflictResolverService handles file conflict resolution with various strategies
type ConflictResolverService struct {
	// TODO: Future: Add hooks for user interaction prompts
}

// NewConflictResolverService creates a new conflict resolver service
func NewConflictResolverService() *ConflictResolverService {
	return &ConflictResolverService{}
}

// ResolveConflict resolves a file conflict using the specified strategy
func (cr *ConflictResolverService) ResolveConflict(ctx context.Context, srcPath, dstPath string, strategy options.ConflictStrategy) (string, error) {
	switch strategy {
	case options.ConflictOverwrite:
		return cr.resolveByOverwrite(ctx, srcPath, dstPath)
	case options.ConflictSkip:
		return cr.resolveBySkip(ctx, srcPath, dstPath)
	case options.ConflictRename:
		return cr.resolveByRename(ctx, srcPath, dstPath)
	case options.ConflictPrompt:
		return cr.resolveByPrompt(ctx, srcPath, dstPath)
	default:
		return "", fmt.Errorf("unknown conflict strategy: %s", strategy)
	}
}

// DetectConflict checks if a conflict exists between source and destination
func (cr *ConflictResolverService) DetectConflict(ctx context.Context, srcPath, dstPath string) (*types.ConflictInfo, error) {
	// Check if destination exists
	dstInfo, err := os.Stat(dstPath)
	if os.IsNotExist(err) {
		// No conflict - destination doesn't exist
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat destination %s: %w", dstPath, err)
	}

	// Get source info
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat source %s: %w", srcPath, err)
	}

	// Determine conflict type
	var conflictType types.ConflictType
	if dstInfo.IsDir() && srcInfo.IsDir() {
		conflictType = types.ConflictDirectoryExists
	} else if !dstInfo.IsDir() && !srcInfo.IsDir() {
		conflictType = types.ConflictFileExists
	} else {
		// Mixed file/directory conflict - treat as file exists
		conflictType = types.ConflictFileExists
	}

	// Create file nodes for conflict info
	srcNode := &trees.FileNode{
		Path: srcPath,
		Name: filepath.Base(srcPath),
		Metadata: trees.Metadata{
			Size:       srcInfo.Size(),
			ModifiedAt: srcInfo.ModTime(),
			NodeType:   trees.File,
		},
	}

	dstNode := &trees.FileNode{
		Path: dstPath,
		Name: filepath.Base(dstPath),
		Metadata: trees.Metadata{
			Size:       dstInfo.Size(),
			ModifiedAt: dstInfo.ModTime(),
			NodeType:   trees.File,
		},
	}

	if srcInfo.IsDir() {
		srcNode.Metadata.NodeType = trees.Directory
	}
	if dstInfo.IsDir() {
		dstNode.Metadata.NodeType = trees.Directory
	}

	return &types.ConflictInfo{
		SourcePath:   srcPath,
		TargetPath:   dstPath,
		ConflictType: conflictType,
		SourceInfo:   srcNode,
		TargetInfo:   dstNode,
	}, nil
}

// GenerateUniqueFilename generates a unique filename by appending a suffix
func (cr *ConflictResolverService) GenerateUniqueFilename(path string) string {
	dir := filepath.Dir(path)
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	baseName := strings.TrimSuffix(name, ext)

	counter := 1
	for {
		var newName string
		if ext != "" {
			newName = fmt.Sprintf("%s_%d%s", baseName, counter, ext)
		} else {
			newName = fmt.Sprintf("%s_%d", baseName, counter)
		}

		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++

		// Prevent infinite loops
		if counter > 9999 {
			return filepath.Join(dir, fmt.Sprintf("%s_%d_%d%s", baseName, counter, time.Now().Unix(), ext))
		}
	}
}

// resolveByOverwrite handles overwrite strategy
func (cr *ConflictResolverService) resolveByOverwrite(ctx context.Context, srcPath, dstPath string) (string, error) {
	// Check if destination exists and is writable
	if info, err := os.Stat(dstPath); err == nil {
		if info.Mode().Perm()&0200 == 0 {
			return "", fmt.Errorf("destination file %s is read-only", dstPath)
		}
	}

	// Return original destination path to allow overwrite
	return dstPath, nil
}

// resolveBySkip handles skip strategy
func (cr *ConflictResolverService) resolveBySkip(ctx context.Context, srcPath, dstPath string) (string, error) {
	// Return empty string to indicate skipping
	return "", nil
}

// resolveByRename handles rename strategy
func (cr *ConflictResolverService) resolveByRename(ctx context.Context, srcPath, dstPath string) (string, error) {
	return cr.GenerateUniqueFilename(dstPath), nil
}

// resolveByPrompt handles interactive prompt strategy
func (cr *ConflictResolverService) resolveByPrompt(ctx context.Context, srcPath, dstPath string) (string, error) {
	// For now, fall back to rename strategy
	// TODO: Implement interactive prompting in Phase 2
	return cr.resolveByRename(ctx, srcPath, dstPath)
}

// GetConflictResolutionOptions returns available resolution options for a conflict
func (cr *ConflictResolverService) GetConflictResolutionOptions(conflict *types.ConflictInfo) []options.ConflictStrategy {
	switch conflict.ConflictType {
	case types.ConflictFileExists:
		return []options.ConflictStrategy{
			options.ConflictOverwrite,
			options.ConflictSkip,
			options.ConflictRename,
			options.ConflictPrompt,
		}
	case types.ConflictDirectoryExists:
		return []options.ConflictStrategy{
			options.ConflictSkip,
			options.ConflictRename,
			options.ConflictPrompt,
		}
	case types.ConflictPermissionDenied:
		return []options.ConflictStrategy{
			options.ConflictSkip,
			options.ConflictPrompt,
		}
	default:
		return []options.ConflictStrategy{
			options.ConflictSkip,
		}
	}
}

// BatchResolveConflicts resolves multiple conflicts efficiently
func (cr *ConflictResolverService) BatchResolveConflicts(ctx context.Context, conflicts []types.ConflictInfo, strategy options.ConflictStrategy) (map[string]string, error) {
	resolutions := make(map[string]string)

	for _, conflict := range conflicts {
		select {
		case <-ctx.Done():
			return resolutions, ctx.Err()
		default:
		}

		resolved, err := cr.ResolveConflict(ctx, conflict.SourcePath, conflict.TargetPath, strategy)
		if err != nil {
			return resolutions, fmt.Errorf("failed to resolve conflict for %s: %w", conflict.SourcePath, err)
		}

		if resolved != "" {
			resolutions[conflict.SourcePath] = resolved
		}
	}

	return resolutions, nil
}

// ValidateResolution validates that a conflict resolution is valid
func (cr *ConflictResolverService) ValidateResolution(srcPath, resolvedPath string, strategy options.ConflictStrategy) error {
	switch strategy {
	case options.ConflictOverwrite:
		if srcPath == resolvedPath {
			return fmt.Errorf("overwrite resolution must use destination path")
		}
	case options.ConflictSkip:
		if resolvedPath != "" {
			return fmt.Errorf("skip resolution must return empty path")
		}
	case options.ConflictRename:
		if srcPath == resolvedPath {
			return fmt.Errorf("rename resolution must generate different path")
		}
		if filepath.Dir(srcPath) != filepath.Dir(resolvedPath) {
			return fmt.Errorf("rename resolution must be in same directory")
		}
	}

	return nil
}

// Ensure ConflictResolverService implements the interface
var _ interfaces.ConflictResolver = (*ConflictResolverService)(nil)
