package trees

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Add constants for tag categories
const (
	TagSizeSmall  = "small"
	TagSizeMedium = "medium"
	TagSizeLarge  = "large"

	sizeThresholdSmall  = 1e3
	sizeThresholdMedium = 1e6

	TagTypeFile      = "file"
	TagTypeDirectory = "folder"

	TagPermReadable = "readable"
	TagPermWritable = "writable"

	TagPrefix = "type:" // Prefix for categorized tags
)

// GenerateTags generates tags based on the metadata of a file or directory
func GenerateTags(metadata Metadata, filename string) ([]string, error) {
	if err := metadata.Validate(); err != nil {
		return nil, fmt.Errorf("invalid metadata: %w", err)
	}

	tags := []string{}

	// Tag based on NodeType
	if metadata.NodeType.String() == "directory" {
		tags = append(tags, TagTypeDirectory)
	} else if metadata.NodeType.String() == "file" {
		tags = append(tags, TagTypeFile)
	}

	// Tag based on file size using defined thresholds
	if metadata.Size > sizeThresholdMedium {
		tags = append(tags, TagSizeLarge)
	} else if metadata.Size > sizeThresholdSmall {
		tags = append(tags, TagSizeMedium)
	} else {
		tags = append(tags, TagSizeSmall)
	}

	// Tag based on permissions
	if metadata.Permissions&0200 != 0 {
		tags = append(tags, TagPermWritable)
	}
	if metadata.Permissions&0400 != 0 {
		tags = append(tags, TagPermReadable)
	}

	// Removed erroneous file type tagging that was checking permissions string for '.txt'

	// Generate tags based on file extension
	if metadata.NodeType == File && filename != "" {
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != "" {
			// Remove the dot from extension
			ext = ext[1:]
			tags = append(tags, TagPrefix+ext)

			// Add common file type categories
			switch ext {
			case "txt", "md", "doc", "docx", "rtf":
				tags = append(tags, TagPrefix+"document")
			case "pdf":
				tags = append(tags, TagPrefix+"pdf")
				tags = append(tags, TagPrefix+"document")
			case "jpg", "jpeg", "png", "gif", "bmp", "svg", "webp":
				tags = append(tags, TagPrefix+"image")
			case "mp4", "avi", "mkv", "mov", "wmv", "flv", "webm":
				tags = append(tags, TagPrefix+"video")
			case "mp3", "wav", "flac", "aac", "ogg", "m4a":
				tags = append(tags, TagPrefix+"audio")
			case "zip", "rar", "7z", "tar", "gz", "bz2":
				tags = append(tags, TagPrefix+"archive")
			case "js", "ts", "py", "go", "cpp", "c", "java", "rs", "php", "rb":
				tags = append(tags, TagPrefix+"code")
			case "html", "css", "scss", "less":
				tags = append(tags, TagPrefix+"web")
			case "json", "xml", "yaml", "yml", "toml", "csv":
				tags = append(tags, TagPrefix+"data")
			}
		}
	}

	// Generate tags based on modification time
	now := time.Now()
	modAge := now.Sub(metadata.ModifiedAt)

	switch {
	case modAge < 24*time.Hour:
		tags = append(tags, TagPrefix+"recent")
	case modAge < 7*24*time.Hour:
		tags = append(tags, TagPrefix+"thisweek")
	case modAge < 30*24*time.Hour:
		tags = append(tags, TagPrefix+"thismonth")
	case modAge < 365*24*time.Hour:
		tags = append(tags, TagPrefix+"thisyear")
	default:
		tags = append(tags, TagPrefix+"old")
	}

	// Generate tags based on file size
	sizeGB := float64(metadata.Size) / (1024 * 1024 * 1024)
	sizeMB := float64(metadata.Size) / (1024 * 1024)

	switch {
	case metadata.Size == 0:
		tags = append(tags, TagPrefix+"empty")
	case sizeMB < 1:
		tags = append(tags, TagPrefix+"small")
	case sizeMB < 100:
		tags = append(tags, TagPrefix+"medium")
	case sizeGB < 1:
		tags = append(tags, TagPrefix+"large")
	default:
		tags = append(tags, TagPrefix+"huge")
	}

	// Add hidden file tag
	if metadata.NodeType == File && filename != "" && strings.HasPrefix(filename, ".") {
		tags = append(tags, TagPrefix+"hidden")
	}

	return tags, nil
}

// AddTagsToMetadata adds tags to a Metadata struct
func AddTagsToMetadata(metadata *Metadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	tags, err := GenerateTags(*metadata, "")
	if err != nil {
		return fmt.Errorf("failed to generate tags: %w", err)
	}

	metadata.Tags = removeDuplicates(tags)
	return nil
}

// AddTagsToMetadataWithFilename adds tags to a Metadata struct with filename context
func AddTagsToMetadataWithFilename(metadata *Metadata, filename string) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	tags, err := GenerateTags(*metadata, filename)
	if err != nil {
		return fmt.Errorf("failed to generate tags: %w", err)
	}

	metadata.Tags = removeDuplicates(tags)
	return nil
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(tags []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, tag := range tags {
		if !keys[tag] {
			keys[tag] = true
			result = append(result, tag)
		}
	}

	return result
}
