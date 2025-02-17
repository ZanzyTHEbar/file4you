package deskfs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Constants for detecting stash conflict messages
const (
	PopStashConflictMsg = "overwritten by merge"
	ConflictMsgFilesEnd = "commit your changes"
	defaultGitTimeout   = 30 * time.Second
)

// runGitCommand is a helper to execute git commands with context, logging, and timeouts.
func (dfs *DesktopFS) runGitCommand(ctx context.Context, repoDir string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", repoDir}, args...)
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Initialize a Git repository in the specified directory if it doesn't already exist.
func (dfs *DesktopFS) InitGitRepo(directory string) error {
	gitDir := filepath.Join(directory, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
		defer cancel()
		if _, err := dfs.runGitCommand(ctx, directory, "init"); err != nil {
			return fmt.Errorf("failed to initialize Git repository: %w", err)
		}
		log.Info().Msgf("Initialized new Git repository at %s", directory)
	}
	return nil
}

// Check if the specified directory is a Git repository.
func (dfs *DesktopFS) IsGitRepo(dir string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	defer cancel()
	if _, err := dfs.runGitCommand(ctx, dir, "rev-parse", "--is-inside-work-tree"); err != nil {
		return false
	}
	return true
}

// Stage all changes and commit them with the specified message.
func (dfs *DesktopFS) GitAddAndCommit(dir, message string) error {
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	if err := dfs.GitAdd(dir, "."); err != nil {
		return fmt.Errorf("error adding files to Git repository in dir %s: %w", dir, err)
	}

	if err := dfs.GitCommit(dir, message); err != nil {
		return fmt.Errorf("error committing files in Git repository in dir %s: %w", dir, err)
	}

	log.Info().Msgf("Committed changes to Git with message: %s", message)
	return nil
}

// Stage specified paths in the repository.
func (dfs *DesktopFS) GitAdd(repoDir, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	defer cancel()
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	if output, err := dfs.runGitCommand(ctx, repoDir, "add", path); err != nil {
		return fmt.Errorf("error adding files to Git repository at %s: %w | Output: %s", repoDir, err, output)
	}
	return nil
}

// Commit with a specified commit message.
func (dfs *DesktopFS) GitCommit(repoDir, commitMsg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	defer cancel()
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	if output, err := dfs.runGitCommand(ctx, repoDir, "commit", "-m", commitMsg, "--allow-empty"); err != nil {
		return fmt.Errorf("error committing files in Git repository at %s: %w | Output: %s", repoDir, err, output)
	}
	return nil
}

// Check for uncommitted changes in the repository.
func (dfs *DesktopFS) CheckUncommittedChanges(repoDir string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	defer cancel()
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	output, err := dfs.runGitCommand(ctx, repoDir, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("error checking for uncommitted changes: %w | Output: %s", err, output)
	}
	return strings.TrimSpace(output) != "", nil
}

// Stash all uncommitted changes with a specified message.
func (dfs *DesktopFS) GitStashCreate(repoDir, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	defer cancel()
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	if output, err := dfs.runGitCommand(ctx, repoDir, "stash", "push", "--include-untracked", "-m", message); err != nil {
		return fmt.Errorf("error creating Git stash: %w | Output: %s", err, output)
	}
	return nil
}

// Pop the latest stash entry, resolving conflicts if specified.
func (dfs *DesktopFS) GitStashPop(repoDir string, forceOverwrite bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	defer cancel()
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	output, err := dfs.runGitCommand(ctx, repoDir, "stash", "pop")
	if err != nil && strings.Contains(output, PopStashConflictMsg) {
		log.Info().Msg("Conflicts detected while popping stash.")
		if forceOverwrite {
			conflictFiles := parseConflictFiles(output)
			for _, file := range conflictFiles {
				if resetErr := dfs.GitCheckoutFile(repoDir, file); resetErr != nil {
					return fmt.Errorf("error resolving conflict for file %s: %w", file, resetErr)
				}
			}
			// Drop the stash after resolution and log any error
			if dropOut, dropErr := dfs.runGitCommand(context.Background(), repoDir, "stash", "drop"); dropErr != nil {
				log.Error().Err(dropErr).Msgf("Error dropping stash: %s", dropOut)
			}
			return nil
		}
		return fmt.Errorf("conflict encountered popping git stash: %s", output)
	} else if err != nil {
		return fmt.Errorf("error popping git stash: %w | Output: %s", err, output)
	}
	return nil
}

// Clears uncommitted changes, including untracked files.
func (dfs *DesktopFS) GitClearUncommittedChanges(repoDir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	defer cancel()
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	if output, err := dfs.runGitCommand(ctx, repoDir, "reset", "--hard"); err != nil {
		return fmt.Errorf("error resetting changes: %w | Output: %s", err, output)
	}

	if output, err := dfs.runGitCommand(ctx, repoDir, "clean", "-d", "-f"); err != nil {
		return fmt.Errorf("error cleaning untracked files: %w | Output: %s", err, output)
	}
	return nil
}

// Check if a specific file has uncommitted changes.
func (dfs *DesktopFS) GitFileHasUncommittedChanges(repoDir, path string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	defer cancel()
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	output, err := dfs.runGitCommand(ctx, repoDir, "status", "--porcelain", path)
	if err != nil {
		return false, fmt.Errorf("error checking uncommitted changes for file %s: %w | Output: %s", path, err, output)
	}
	return strings.TrimSpace(output) != "", nil
}

// Check out a file to discard local changes.
func (dfs *DesktopFS) GitCheckoutFile(repoDir, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	defer cancel()
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	if output, err := dfs.runGitCommand(ctx, repoDir, "checkout", path); err != nil {
		return fmt.Errorf("error checking out file %s: %w | Output: %s", path, err, output)
	}
	return nil
}

// Retrieve commit history for the repository.
func (dfs *DesktopFS) GetCommitHistory(repoDir string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	defer cancel()
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	output, err := dfs.runGitCommand(ctx, repoDir, "rev-list", "--all")
	if err != nil {
		return nil, fmt.Errorf("error retrieving commit history: %w | Output: %s", err, output)
	}
	return strings.Split(strings.TrimSpace(output), "\n"), nil
}

// Rewind the repository to a specified commit.
func (dfs *DesktopFS) GitRewind(repoDir, targetSha string) error {
	dfs.gitMutex.Lock()
	defer dfs.gitMutex.Unlock()

	commits, err := dfs.GetCommitHistory(repoDir)
	if err != nil {
		return fmt.Errorf("error retrieving commit history: %w", err)
	}

	var targetCommit string
	if isSHA(targetSha) {
		targetCommit = targetSha
	} else {
		steps, err := strconv.Atoi(targetSha)
		if err != nil || steps < 0 || steps >= len(commits) {
			return fmt.Errorf("invalid target commit: %s", targetSha)
		}
		targetCommit = commits[steps]
	}

	if output, err := dfs.runGitCommand(context.Background(), repoDir, "checkout", targetCommit); err != nil {
		return fmt.Errorf("error rewinding to commit %s: %w | Output: %s", targetCommit, err, output)
	}
	return nil
}

func (dfs *DesktopFS) handleUncommittedChanges(dir string, params *FilePathParams) error {
	if params.GitEnabled {
		hasUncommitted, err := dfs.CheckUncommittedChanges(dir)
		if err != nil {
			return fmt.Errorf("error checking uncommitted changes: %w", err)
		}

		if hasUncommitted {
			fmt.Println("There are uncommitted changes. Would you like to stash them? (y/n)")
			var input string
			fmt.Scanln(&input)
			if strings.ToLower(input) == "y" {
				if err := dfs.GitStashCreate(dir, "Auto-stash before organizing"); err != nil {
					return fmt.Errorf("error creating git stash: %w", err)
				}
				fmt.Println("Changes stashed successfully.")
			}
		}
	}
	return nil
}

func (dfs *DesktopFS) clearChangesIfNeeded(dir string, params *FilePathParams) error {
	if params.GitEnabled {
		fmt.Println("Would you like to clear all uncommitted changes before organizing? (y/n)")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) == "y" {
			if err := dfs.GitClearUncommittedChanges(dir); err != nil {
				return fmt.Errorf("error clearing uncommitted changes: %w", err)
			}
			fmt.Println("All uncommitted changes have been cleared.")
		}
	}
	return nil
}

// Helper function to parse conflict files from Git output.
func parseConflictFiles(gitOutput string) []string {
	var conflictFiles []string
	lines := strings.Split(gitOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, PopStashConflictMsg) {
			conflictFiles = append(conflictFiles, strings.TrimSpace(line))
		} else if strings.Contains(line, ConflictMsgFilesEnd) {
			break
		}
	}
	return conflictFiles
}

// Utility to validate if a string is a valid SHA-1 hash.
func isSHA(input string) bool {
	matched, _ := regexp.MatchString("^[a-f0-9]{40}$", input)
	return matched
}
