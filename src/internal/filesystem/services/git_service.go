package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"file4you/internal/filesystem/interfaces"
)

// Constants for detecting stash conflict messages
const (
	PopStashConflictMsg = "overwritten by merge"
	ConflictMsgFilesEnd = "commit your changes"
	defaultGitTimeout   = 30 * time.Second
)

// GitServiceImpl provides git repository operations following the service-oriented architecture
type GitServiceImpl struct {
	mutex   sync.RWMutex
	timeout time.Duration
}

// NewGitService creates a new git service instance
func NewGitService() interfaces.GitService {
	return &GitServiceImpl{
		timeout: defaultGitTimeout,
	}
}

// runGitCommand is a helper to execute git commands with context, logging, and timeouts
func (gs *GitServiceImpl) runGitCommand(ctx context.Context, repoDir string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", repoDir}, args...)
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)

	slog.Debug("Executing git command",
		"dir", repoDir,
		"args", args,
		"timeout", gs.timeout)

	output, err := cmd.CombinedOutput()

	if err != nil {
		slog.Error("Git command failed",
			"dir", repoDir,
			"args", args,
			"output", string(output),
			"error", err)
	}

	return string(output), err
}

// InitRepository initializes a git repository in the specified directory if it doesn't already exist
func (gs *GitServiceImpl) InitRepository(ctx context.Context, dir string) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
		defer cancel()

		if _, err := gs.runGitCommand(ctxWithTimeout, dir, "init"); err != nil {
			return fmt.Errorf("failed to initialize Git repository: %w", err)
		}

		slog.Info("Initialized new Git repository", "path", dir)
	} else {
		slog.Debug("Git repository already exists", "path", dir)
	}

	return nil
}

// IsGitRepository checks if the given directory is a git repository (convenience method for testing)
func (gs *GitServiceImpl) IsGitRepository(ctx context.Context, dir string) (bool, error) {
	return gs.IsRepository(dir), nil
}

// IsRepository checks if the specified directory is a git repository
func (gs *GitServiceImpl) IsRepository(dir string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), gs.timeout)
	defer cancel()

	if _, err := gs.runGitCommand(ctx, dir, "rev-parse", "--is-inside-work-tree"); err != nil {
		return false
	}
	return true
}

// AddAndCommit adds all changes and commits them with the specified message
func (gs *GitServiceImpl) AddAndCommit(ctx context.Context, repoDir, message string) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	// Add all changes
	if _, err := gs.runGitCommand(ctxWithTimeout, repoDir, "add", "."); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	// Commit changes
	if _, err := gs.runGitCommand(ctxWithTimeout, repoDir, "commit", "-m", message); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	slog.Info("Successfully added and committed changes", "repo", repoDir, "message", message)
	return nil
}

// AddFiles adds specified files to the git index
func (gs *GitServiceImpl) AddFiles(ctx context.Context, repoDir string, paths ...string) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	args := append([]string{"add"}, paths...)
	if _, err := gs.runGitCommand(ctxWithTimeout, repoDir, args...); err != nil {
		return fmt.Errorf("failed to add files %v: %w", paths, err)
	}

	slog.Debug("Added files to git index", "repo", repoDir, "files", paths)
	return nil
}

// CommitChanges commits staged changes with the specified message
func (gs *GitServiceImpl) CommitChanges(ctx context.Context, repoDir, message string) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	if _, err := gs.runGitCommand(ctxWithTimeout, repoDir, "commit", "-m", message); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	slog.Info("Successfully committed changes", "repo", repoDir, "message", message)
	return nil
}

// HasUncommittedChanges checks if there are any uncommitted changes in the repository
func (gs *GitServiceImpl) HasUncommittedChanges(ctx context.Context, repoDir string) (bool, error) {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	output, err := gs.runGitCommand(ctxWithTimeout, repoDir, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	hasChanges := strings.TrimSpace(output) != ""
	slog.Debug("Checked for uncommitted changes", "repo", repoDir, "hasChanges", hasChanges)
	return hasChanges, nil
}

// StashCreate creates a stash with the specified message
func (gs *GitServiceImpl) StashCreate(ctx context.Context, repoDir, message string) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	if _, err := gs.runGitCommand(ctxWithTimeout, repoDir, "stash", "push", "-m", message); err != nil {
		return fmt.Errorf("failed to create stash: %w", err)
	}

	slog.Info("Created git stash", "repo", repoDir, "message", message)
	return nil
}

// StashPop pops the most recent stash
func (gs *GitServiceImpl) StashPop(ctx context.Context, repoDir string, forceOverwrite bool) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	args := []string{"stash", "pop"}
	if forceOverwrite {
		// First, clear any uncommitted changes that might conflict
		if err := gs.clearUncommittedChangesInternal(ctxWithTimeout, repoDir); err != nil {
			return fmt.Errorf("failed to clear uncommitted changes before stash pop: %w", err)
		}
	}

	output, err := gs.runGitCommand(ctxWithTimeout, repoDir, args...)
	if err != nil {
		if strings.Contains(output, PopStashConflictMsg) || strings.Contains(output, ConflictMsgFilesEnd) {
			if forceOverwrite {
				// Try to resolve conflicts by resetting
				if _, resetErr := gs.runGitCommand(ctxWithTimeout, repoDir, "reset", "--hard"); resetErr != nil {
					return fmt.Errorf("failed to resolve stash conflicts: %w", resetErr)
				}
				// Try popping again
				if _, popErr := gs.runGitCommand(ctxWithTimeout, repoDir, "stash", "pop"); popErr != nil {
					return fmt.Errorf("failed to pop stash after conflict resolution: %w", popErr)
				}
			} else {
				return fmt.Errorf("stash pop conflicts detected, use forceOverwrite=true to resolve: %w", err)
			}
		} else {
			return fmt.Errorf("failed to pop stash: %w", err)
		}
	}

	slog.Info("Successfully popped git stash", "repo", repoDir, "forceOverwrite", forceOverwrite)
	return nil
}

// ClearUncommittedChanges clears all uncommitted changes in the repository
func (gs *GitServiceImpl) ClearUncommittedChanges(ctx context.Context, repoDir string) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	return gs.clearUncommittedChangesInternal(ctxWithTimeout, repoDir)
}

// Internal helper for clearing uncommitted changes (assumes mutex is already held)
func (gs *GitServiceImpl) clearUncommittedChangesInternal(ctx context.Context, repoDir string) error {
	// Reset tracked files
	if _, err := gs.runGitCommand(ctx, repoDir, "reset", "--hard"); err != nil {
		return fmt.Errorf("failed to reset tracked files: %w", err)
	}

	// Clean untracked files and directories
	if _, err := gs.runGitCommand(ctx, repoDir, "clean", "-fd"); err != nil {
		return fmt.Errorf("failed to clean untracked files: %w", err)
	}

	slog.Info("Cleared all uncommitted changes", "repo", repoDir)
	return nil
}

// FileHasUncommittedChanges checks if a specific file has uncommitted changes
func (gs *GitServiceImpl) FileHasUncommittedChanges(ctx context.Context, repoDir, path string) (bool, error) {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	output, err := gs.runGitCommand(ctxWithTimeout, repoDir, "status", "--porcelain", path)
	if err != nil {
		return false, fmt.Errorf("failed to check file status: %w", err)
	}

	hasChanges := strings.TrimSpace(output) != ""
	slog.Debug("Checked file for uncommitted changes", "repo", repoDir, "file", path, "hasChanges", hasChanges)
	return hasChanges, nil
}

// CheckoutFile resets a specific file to its committed state
func (gs *GitServiceImpl) CheckoutFile(ctx context.Context, repoDir, path string) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	if _, err := gs.runGitCommand(ctxWithTimeout, repoDir, "checkout", "HEAD", "--", path); err != nil {
		return fmt.Errorf("failed to checkout file %s: %w", path, err)
	}

	slog.Info("Successfully checked out file", "repo", repoDir, "file", path)
	return nil
}

// GetCommitHistory retrieves the commit history for the repository
func (gs *GitServiceImpl) GetCommitHistory(ctx context.Context, repoDir string) ([]string, error) {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	output, err := gs.runGitCommand(ctxWithTimeout, repoDir, "log", "--oneline", "--max-count=20")
	if err != nil {
		return nil, fmt.Errorf("failed to get commit history: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}

	slog.Debug("Retrieved commit history", "repo", repoDir, "commits", len(lines))
	return lines, nil
}

// Rewind rewinds the repository to a specific commit SHA
func (gs *GitServiceImpl) Rewind(ctx context.Context, repoDir, targetSha string) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	// Parse target if it's a step count instead of SHA
	if steps, err := strconv.Atoi(targetSha); err == nil {
		targetSha = fmt.Sprintf("HEAD~%d", steps)
	}

	// Validate the target exists
	if _, err := gs.runGitCommand(ctxWithTimeout, repoDir, "rev-parse", "--verify", targetSha); err != nil {
		return fmt.Errorf("invalid target SHA or steps: %s", targetSha)
	}

	// Check for uncommitted changes
	hasChanges, err := gs.HasUncommittedChanges(ctxWithTimeout, repoDir)
	if err != nil {
		return fmt.Errorf("failed to check for uncommitted changes: %w", err)
	}

	if hasChanges {
		// Stash uncommitted changes before rewinding
		stashMsg := fmt.Sprintf("Auto-stash before rewind to %s", targetSha)
		if err := gs.StashCreate(ctxWithTimeout, repoDir, stashMsg); err != nil {
			return fmt.Errorf("failed to stash changes before rewind: %w", err)
		}
		slog.Info("Stashed uncommitted changes before rewind", "repo", repoDir)
	}

	// Perform the rewind
	if _, err := gs.runGitCommand(ctxWithTimeout, repoDir, "reset", "--hard", targetSha); err != nil {
		return fmt.Errorf("failed to rewind to %s: %w", targetSha, err)
	}

	slog.Info("Successfully rewound repository", "repo", repoDir, "target", targetSha)
	return nil
}

// Ensure GitServiceImpl implements the GitService interface
var _ interfaces.GitService = (*GitServiceImpl)(nil)
