package fs

import (
	"file4you/internal/cli"
	"file4you/internal/filesystem/options"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// NewSearch creates a new search command for finding files
func NewSearch(params *cli.CmdParams) *cobra.Command {
	var (
		targetDir     string
		pattern       string
		namePattern   string
		sizeFilter    string
		typeFilter    string
		maxDepth      int
		showHidden    bool
		caseSensitive bool
		outputFormat  string
	)

	searchCmd := &cobra.Command{
		Use:   "search [pattern]",
		Short: "Search for files and directories",
		Long: `Search provides powerful file and directory searching with support for:
- Name pattern matching (wildcards and regex)
- File type filtering
- Size-based filtering
- Content search (future feature)
- Multiple output formats`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get pattern from args if provided
			if len(args) > 0 {
				pattern = args[0]
			}

			// Use name pattern if no general pattern specified
			if pattern == "" && namePattern != "" {
				pattern = namePattern
			}

			if pattern == "" {
				return fmt.Errorf("search pattern is required (use --name or provide as argument)")
			}

			// Determine target directory
			if targetDir == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
				targetDir = cwd
			}

			// Validate directory exists
			if _, err := os.Stat(targetDir); os.IsNotExist(err) {
				return fmt.Errorf("directory does not exist: %s", targetDir)
			}

			// Convert to absolute path
			absPath, err := filepath.Abs(targetDir)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}
			targetDir = absPath

			params.Interactor.Info(fmt.Sprintf("Searching in: %s", targetDir))
			params.Interactor.Info(fmt.Sprintf("Pattern: %s", pattern))

			// Create search options
			searchOpts := &SearchOptions{
				Pattern:       pattern,
				Directory:     targetDir,
				TypeFilter:    typeFilter,
				SizeFilter:    sizeFilter,
				MaxDepth:      maxDepth,
				IncludeHidden: showHidden,
				CaseSensitive: caseSensitive,
			}

			// Perform search
			results, err := performSearch(params, searchOpts)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			// Display results
			if err := displaySearchResults(params, results, outputFormat); err != nil {
				return fmt.Errorf("failed to display results: %w", err)
			}

			return nil
		},
	}

	// Add flags
	searchCmd.Flags().StringVarP(&targetDir, "directory", "d", "", "Directory to search in (default: current directory)")
	searchCmd.Flags().StringVarP(&namePattern, "name", "n", "", "File name pattern (wildcards supported)")
	searchCmd.Flags().StringVarP(&typeFilter, "type", "t", "", "File type filter (e.g., 'f' for files, 'd' for directories)")
	searchCmd.Flags().StringVarP(&sizeFilter, "size", "s", "", "Size filter (e.g., '+1M', '-100K')")
	searchCmd.Flags().IntVar(&maxDepth, "max-depth", -1, "Maximum search depth (-1 for unlimited)")
	searchCmd.Flags().BoolVar(&showHidden, "hidden", false, "Include hidden files and directories")
	searchCmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Case sensitive pattern matching")
	searchCmd.Flags().StringVarP(&outputFormat, "format", "f", "simple", "Output format (simple, detailed, json)")

	return searchCmd
}

// SearchOptions holds search configuration
type SearchOptions struct {
	Pattern       string
	Directory     string
	TypeFilter    string
	SizeFilter    string
	MaxDepth      int
	IncludeHidden bool
	CaseSensitive bool
}

// SearchResult represents a single search result
type SearchResult struct {
	Path      string
	Name      string
	Type      string
	Size      int64
	IsDir     bool
	Extension string
}

// performSearch executes the file search
func performSearch(params *cli.CmdParams, opts *SearchOptions) ([]*SearchResult, error) {
	var results []*SearchResult

	// Compile pattern as regex if needed
	var pattern *regexp.Regexp
	var err error

	if opts.CaseSensitive {
		pattern, err = regexp.Compile(opts.Pattern)
	} else {
		pattern, err = regexp.Compile("(?i)" + opts.Pattern)
	}

	if err != nil {
		// If regex compilation fails, treat as simple wildcard pattern
		pattern = nil
	}

	// Configure traversal options
	traversalOpts := options.DefaultTraversalOptions()
	traversalOpts.MaxDepth = opts.MaxDepth
	traversalOpts.IncludeHidden = opts.IncludeHidden

	// Walk the directory tree
	err = filepath.Walk(opts.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log error but continue walking
			params.Interactor.Info(fmt.Sprintf("Warning - Error accessing %s: %v", path, err))
			return nil
		}

		// Skip hidden files if not requested
		if !opts.IncludeHidden && strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check depth limit
		if opts.MaxDepth >= 0 {
			relPath, _ := filepath.Rel(opts.Directory, path)
			depth := strings.Count(relPath, string(os.PathSeparator))
			if depth > opts.MaxDepth {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Apply type filter
		if opts.TypeFilter != "" {
			switch opts.TypeFilter {
			case "f", "file":
				if info.IsDir() {
					return nil
				}
			case "d", "dir", "directory":
				if !info.IsDir() {
					return nil
				}
			}
		}

		// Apply pattern matching
		matched := false
		if pattern != nil {
			matched = pattern.MatchString(info.Name()) || pattern.MatchString(path)
		} else {
			// Simple wildcard matching
			matched = matchWildcard(opts.Pattern, info.Name(), opts.CaseSensitive)
		}

		if matched {
			result := &SearchResult{
				Path:      path,
				Name:      info.Name(),
				Size:      info.Size(),
				IsDir:     info.IsDir(),
				Extension: filepath.Ext(info.Name()),
			}

			if info.IsDir() {
				result.Type = "directory"
			} else {
				result.Type = "file"
			}

			results = append(results, result)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return results, nil
}

// displaySearchResults shows search results in specified format
func displaySearchResults(params *cli.CmdParams, results []*SearchResult, format string) error {
	if len(results) == 0 {
		params.Interactor.Info("No files found matching the search criteria")
		return nil
	}

	params.Interactor.Success(fmt.Sprintf("Found %d result(s)", len(results)))

	switch format {
	case "simple":
		for _, result := range results {
			params.Interactor.Output(result.Path)
		}
	case "detailed":
		for _, result := range results {
			typeIndicator := "F"
			if result.IsDir {
				typeIndicator = "D"
			}
			size := ""
			if !result.IsDir {
				size = fmt.Sprintf(" (%s)", formatBytes(result.Size))
			}
			params.Interactor.Output(fmt.Sprintf("[%s] %s%s", typeIndicator, result.Path, size))
		}
	case "json":
		// TODO: Implement JSON output
		params.Interactor.Info("JSON output format will be implemented")
		fallthrough
	default:
		// Default to simple format
		for _, result := range results {
			params.Interactor.Output(result.Path)
		}
	}

	return nil
}

// matchWildcard performs simple wildcard pattern matching
func matchWildcard(pattern, name string, caseSensitive bool) bool {
	if !caseSensitive {
		pattern = strings.ToLower(pattern)
		name = strings.ToLower(name)
	}

	// Simple wildcard matching - convert * to regex
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = strings.ReplaceAll(pattern, "?", ".")
	pattern = "^" + pattern + "$"

	matched, _ := regexp.MatchString(pattern, name)
	return matched
}
