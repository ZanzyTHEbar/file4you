package trees

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"gonum.org/v1/gonum/spatial/kdtree"
)

// Metadata holds additional information for each node in the DirectoryTree
type Metadata struct {
	Size        int64       `json:"size"`
	ModifiedAt  time.Time   `json:"modified_at"`
	CreatedAt   time.Time   `json:"created_at"`
	NodeType    NodeType    `json:"node_type"` // Should be an enum
	Permissions os.FileMode `json:"permissions"`
	Owner       string      `json:"owner"`
	Tags        []string    `json:"tags"`
}

type NodeType int

const (
	Directory NodeType = iota
	File
)

func NewMetadata(fileinfo os.FileInfo) Metadata {
	// Get file permissions and modification time
	permissions := fileinfo.Mode()
	modifiedAt := fileinfo.ModTime()

	// For Linux, creation time is not typically available. Use zero time or alternative method if needed.
	createdAt := time.Time{}

	// Set NodeType based on file type
	var nodeType NodeType
	if fileinfo.IsDir() {
		nodeType = Directory
	} else {
		nodeType = File
	}

	// Create metadata struct
	return Metadata{
		Size:        fileinfo.Size(),
		ModifiedAt:  modifiedAt,
		CreatedAt:   createdAt,
		NodeType:    nodeType,
		Permissions: permissions,
		Owner:       getFileOwner(fileinfo), // Implement owner retrieval for Linux
		Tags:        []string{},
	}
}

// ToKDTreePoint converts Metadata attributes into a k-dimensional point (slice of float64) for KD-Tree usage.
func (m *Metadata) ToKDTreePoint() (kdtree.Point, error) {
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("invalid metadata: %w", err)
	}

	return kdtree.Point{
		float64(m.Size),
		float64(m.ModifiedAt.Unix()),
		float64(m.CreatedAt.Unix()),
		float64(m.Permissions.Perm()),
	}, nil
}

// Add validation method
func (m *Metadata) Validate() error {
	if m.Size < 0 {
		return fmt.Errorf("size cannot be negative")
	}
	if m.ModifiedAt.IsZero() {
		return fmt.Errorf("modified time cannot be zero")
	}
	if m.NodeType != File && m.NodeType != Directory {
		return fmt.Errorf("invalid node type: %s", m.NodeType.String())
	}
	return nil
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
	// Add tags to metadata with filename-aware tagging
	dirName := filepath.Base(node.Path)
	err = AddTagsToMetadataWithFilename(&metadata, dirName)
	if err != nil {
		return err
	}
	node.Metadata = metadata

	// Add metadata to all files within the directory
	for _, fileNode := range node.Files {
		fileMetadata, err := GenerateMetadataFromPath(fileNode.Path)
		if err != nil {
			return err
		}
		// Add enhanced tags to file metadata with filename
		fileName := filepath.Base(fileNode.Path)
		err = AddTagsToMetadataWithFilename(&fileMetadata, fileName)
		if err != nil {
			return err
		}
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

// Convert NodeType to String
func (n NodeType) String() string {
	switch n {
	case Directory:
		return "directory"
	case File:
		return "file"
	default:
		return "unknown"
	}
}

// Map string to NodeType
func StringToNodeType(s string) NodeType {
	switch s {
	case "directory":
		return Directory
	case "file":
		return File
	default:
		return -1
	}
}

// getFileOwner retrieves the owner name for a file on Unix-like systems
func getFileOwner(fileinfo os.FileInfo) string {
	// Try to get the owner name from file system info
	if stat, ok := fileinfo.Sys().(*syscall.Stat_t); ok {
		if u, err := user.LookupId(strconv.Itoa(int(stat.Uid))); err == nil {
			return u.Username
		}
		// If lookup fails, return the UID as a string
		return strconv.Itoa(int(stat.Uid))
	}

	// Fallback if we can't get system info
	return "unknown"
}
