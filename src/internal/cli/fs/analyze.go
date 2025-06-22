package fs

import (
	"file4you/internal/cli"
	"file4you/internal/filesystem/types"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// NewAnalyze creates a new analyze command for directory analysis
func NewAnalyze(params *cli.CmdParams) *cobra.Command {
	var (
		targetDir  string
		maxDepth   int
		showHidden bool
		detailed   bool
		outputPath string
	)

	analyzeCmd := &cobra.Command{
		Use:   "analyze [directory]",
		Short: "Analyze directory structure and file statistics",
		Long: `Analyze provides detailed statistics about directory contents including:
- File count and size distribution
- File type breakdown
- Directory structure analysis
- Duplicate detection
- Organization recommendations`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine target directory
			if len(args) > 0 {
				targetDir = args[0]
			} else if targetDir == "" {
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

			params.Interactor.Info(fmt.Sprintf("Analyzing directory: %s", targetDir))

			// Analyze directory
			analysis, err := params.Filesystem.AnalyzeDirectory(cmd.Context(), targetDir)
			if err != nil {
				return fmt.Errorf("failed to analyze directory: %w", err)
			}

			// Display analysis results
			if err := displayAnalysisResults(params, analysis, detailed); err != nil {
				return fmt.Errorf("failed to display results: %w", err)
			}

			// Save detailed report if requested
			if outputPath != "" {
				if err := saveAnalysisReport(params, analysis, outputPath, detailed); err != nil {
					return fmt.Errorf("failed to save report: %w", err)
				}
				params.Interactor.Success(fmt.Sprintf("Detailed report saved to: %s", outputPath))
			}

			return nil
		},
	}

	// Add flags
	analyzeCmd.Flags().StringVarP(&targetDir, "directory", "d", "", "Directory to analyze (default: current directory)")
	analyzeCmd.Flags().IntVar(&maxDepth, "max-depth", -1, "Maximum depth for traversal (-1 for unlimited)")
	analyzeCmd.Flags().BoolVar(&showHidden, "show-hidden", false, "Include hidden files and directories")
	analyzeCmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed analysis including file listings")
	analyzeCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Save detailed report to file")

	return analyzeCmd
}

// displayAnalysisResults shows the analysis results in the terminal
func displayAnalysisResults(params *cli.CmdParams, analysis *types.DirectoryAnalysis, detailed bool) error {
	params.Interactor.Success("Directory analysis completed successfully")

	params.Interactor.Output("Analysis Results:")
	params.Interactor.Output("================")

	// Basic statistics
	params.Interactor.Info(fmt.Sprintf("Total Files: %d", analysis.TotalFiles))
	params.Interactor.Info(fmt.Sprintf("Total Directories: %d", analysis.TotalDirectories))
	params.Interactor.Info(fmt.Sprintf("Total Size: %s", formatBytes(analysis.TotalSize)))
	params.Interactor.Info(fmt.Sprintf("Max Depth: %d", analysis.MaxDepth))
	params.Interactor.Info(fmt.Sprintf("Analysis Duration: %v", analysis.Duration))

	// File types breakdown
	if len(analysis.FileTypes) > 0 {
		params.Interactor.Output("\nFile Types:")
		for fileType, count := range analysis.FileTypes {
			params.Interactor.Info(fmt.Sprintf("  %s: %d files", fileType, count))
		}
	}

	// Size distribution
	if len(analysis.SizeDistribution) > 0 {
		params.Interactor.Output("\nSize Distribution:")
		for sizeRange, count := range analysis.SizeDistribution {
			params.Interactor.Info(fmt.Sprintf("  %s: %d files", sizeRange, count))
		}
	}

	// Show largest files if detailed
	if detailed && len(analysis.LargestFiles) > 0 {
		params.Interactor.Output("\nLargest Files:")
		for i, file := range analysis.LargestFiles {
			if i >= 10 { // Limit to top 10
				break
			}
			params.Interactor.Info(fmt.Sprintf("  %s (%s)", file.Path, formatBytes(file.Metadata.Size)))
		}
	}

	return nil
}

// saveAnalysisReport saves the analysis results to a file
func saveAnalysisReport(params *cli.CmdParams, analysis *types.DirectoryAnalysis, outputPath string, detailed bool) error {
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	reportContent := generateAnalysisReport(analysis, detailed)

	if err := os.WriteFile(outputPath, []byte(reportContent), 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

// generateAnalysisReport creates a formatted analysis report
func generateAnalysisReport(analysis *types.DirectoryAnalysis, detailed bool) string {
	var report strings.Builder

	report.WriteString("File4You Directory Analysis Report\n")
	report.WriteString("==================================\n\n")

	// Basic statistics
	report.WriteString(fmt.Sprintf("Total Files: %d\n", analysis.TotalFiles))
	report.WriteString(fmt.Sprintf("Total Directories: %d\n", analysis.TotalDirectories))
	report.WriteString(fmt.Sprintf("Total Size: %s\n", formatBytes(analysis.TotalSize)))
	report.WriteString(fmt.Sprintf("Max Depth: %d\n", analysis.MaxDepth))
	report.WriteString(fmt.Sprintf("Analysis Duration: %v\n\n", analysis.Duration))

	// File types
	if len(analysis.FileTypes) > 0 {
		report.WriteString("File Types:\n")
		for fileType, count := range analysis.FileTypes {
			report.WriteString(fmt.Sprintf("  %s: %d files\n", fileType, count))
		}
		report.WriteString("\n")
	}

	// Size distribution
	if len(analysis.SizeDistribution) > 0 {
		report.WriteString("Size Distribution:\n")
		for sizeRange, count := range analysis.SizeDistribution {
			report.WriteString(fmt.Sprintf("  %s: %d files\n", sizeRange, count))
		}
		report.WriteString("\n")
	}

	// Detailed information
	if detailed {
		if len(analysis.LargestFiles) > 0 {
			report.WriteString("Largest Files:\n")
			for i, file := range analysis.LargestFiles {
				if i >= 20 { // Limit to top 20 in detailed report
					break
				}
				report.WriteString(fmt.Sprintf("  %s (%s)\n", file.Path, formatBytes(file.Metadata.Size)))
			}
			report.WriteString("\n")
		}

		if len(analysis.OldestFiles) > 0 {
			report.WriteString("Oldest Files:\n")
			for i, file := range analysis.OldestFiles {
				if i >= 10 {
					break
				}
				report.WriteString(fmt.Sprintf("  %s (%v)\n", file.Path, file.Metadata.ModifiedAt))
			}
			report.WriteString("\n")
		}

		if len(analysis.NewestFiles) > 0 {
			report.WriteString("Newest Files:\n")
			for i, file := range analysis.NewestFiles {
				if i >= 10 {
					break
				}
				report.WriteString(fmt.Sprintf("  %s (%v)\n", file.Path, file.Metadata.ModifiedAt))
			}
		}
	}

	return report.String()
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
