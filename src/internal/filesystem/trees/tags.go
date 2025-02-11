package trees

import "strings"

// GenerateTags generates tags based on the metadata of a file or directory
func GenerateTags(metadata Metadata) []string {
	tags := []string{}

	// Tag based on NodeType
	if metadata.NodeType == "directory" {
		tags = append(tags, "folder")
	} else if metadata.NodeType == "file" {
		tags = append(tags, "file")
	}

	// Tag based on file size
	if metadata.Size > 1e6 {
		tags = append(tags, "large")
	} else if metadata.Size > 1e3 {
		tags = append(tags, "medium")
	} else {
		tags = append(tags, "small")
	}

	// Tag based on permissions
	if metadata.Permissions&0200 != 0 {
		tags = append(tags, "writable")
	}
	if metadata.Permissions&0400 != 0 {
		tags = append(tags, "readable")
	}

	if metadata.NodeType == "file" && strings.HasSuffix(metadata.Permissions.String(), ".txt") {
		tags = append(tags, "text-file")
	}

	// TODO: Add custom logic to generate other tags, e.g., by extension or modification time, user generated, llm generated, etc.

	return tags
}

// AddTagsToMetadata adds tags to a Metadata struct
func AddTagsToMetadata(metadata *Metadata) {
	metadata.Tags = GenerateTags(*metadata)
}
