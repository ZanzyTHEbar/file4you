package trees

import (
	"fmt"
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
)

// GenerateTags generates tags based on the metadata of a file or directory
func GenerateTags(metadata Metadata) ([]string, error) {
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

	// TODO: Add custom logic to generate other tags, e.g., by extension or modification time, user generated, llm generated, etc.

	return tags, nil
}

// AddTagsToMetadata adds tags to a Metadata struct
func AddTagsToMetadata(metadata *Metadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	tags, err := GenerateTags(*metadata)
	if err != nil {
		return fmt.Errorf("failed to generate tags: %w", err)
	}

	metadata.Tags = tags
	return nil
}
