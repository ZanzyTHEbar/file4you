package trees

import (
	"os"
	"time"

	"gonum.org/v1/gonum/spatial/kdtree"
)

// Metadata holds additional information for each node in the DirectoryTree
type Metadata struct {
	Size        int64       // Size of the file or directory
	ModifiedAt  time.Time   // Last modified time
	CreatedAt   time.Time   // Creation time (if available)
	NodeType    string      // "file" or "directory"
	Permissions os.FileMode // File permissions
	Owner       string      // Owner of the file (if available)
	Tags        []string    // Tags associated with the file or directory
}

func NewMetadata(fileinfo os.FileInfo) Metadata {
	// Get file permissions and modification time
	permissions := fileinfo.Mode()
	modifiedAt := fileinfo.ModTime()

	// Placeholder for createdAt, as it may not be available on all systems
	createdAt := time.Time{}

	// Set NodeType to "file" or "directory"
	nodeType := "file"
	if fileinfo.IsDir() {
		nodeType = "directory"
	}

	// Create metadata struct
	return Metadata{
		Size:        fileinfo.Size(),
		ModifiedAt:  modifiedAt,
		CreatedAt:   createdAt,
		NodeType:    nodeType,
		Permissions: permissions,
		Owner:       "unknown",  // TODO: Owner retrieval can be implemented based on platform
		Tags:        []string{}, // Initialize with an empty list of tags
	}
}

// ToKDTreePoint converts Metadata attributes into a k-dimensional point (slice of float64) for KD-Tree usage.
func (m *Metadata) ToKDTreePoint() kdtree.Point {
	// Convert metadata attributes like size, modification time (Unix timestamp), permissions, etc., to float64
	return kdtree.Point{
		float64(m.Size),
		float64(m.ModifiedAt.Unix()),
		float64(m.CreatedAt.Unix()),
		float64(m.Permissions.Perm()),
	}
}

// GenerateMetadata generates metadata for a given file or directory node
func GenerateMetadataFromPath(nodePath string) (Metadata, error) {
	fileInfo, err := os.Stat(nodePath)
	if err != nil {
		return Metadata{}, err
	}

	metadata := NewMetadata(fileInfo)

	return metadata, nil
}

// AddMetadataToTree recursively traverses the DirectoryTree and adds metadata to each node
func AddMetadataToTree(node *DirectoryNode) error {
	// Generate metadata for the current directory node
	metadata, err := GenerateMetadataFromPath(node.Path)
	if err != nil {
		return err
	}
	// Add tags to metadata
	AddTagsToMetadata(&metadata)
	node.Metadata = metadata

	// Add metadata to all files within the directory
	for _, fileNode := range node.Files {
		fileMetadata, err := GenerateMetadataFromPath(fileNode.Path)
		if err != nil {
			return err
		}
		// Add tags to file metadata
		AddTagsToMetadata(&fileMetadata)
		fileNode.Metadata = fileMetadata
	}

	// Recursively add metadata to child directories
	for _, childDir := range node.Children {
		if err := AddMetadataToTree(childDir); err != nil {
			return err
		}
	}

	return nil
}

// FlattenMetadata flattens metadata into a map that can be used for LLM input
func FlattenMetadata(node *DirectoryNode) map[string]interface{} {
	flatMetadata := make(map[string]interface{})

	// Add directory node metadata
	flatMetadata[node.Path] = node.Metadata

	// Add files metadata
	for _, fileNode := range node.Files {
		flatMetadata[fileNode.Path] = fileNode.Metadata
	}

	// Recursively add child directory metadata
	for _, childDir := range node.Children {
		childMetadata := FlattenMetadata(childDir)
		for key, value := range childMetadata {
			flatMetadata[key] = value
		}
	}

	return flatMetadata
}

// collectAllNodes collects all nodes (both directories and files) from the given DirectoryNode
func collectAllNodes(node *DirectoryNode) []*DirectoryNode {
	var nodes []*DirectoryNode
	nodes = append(nodes, node)
	for _, child := range node.Children {
		nodes = append(nodes, collectAllNodes(child)...)
	}
	return nodes
}
