package trees

import (
	"fmt"
	"log/slog"
	"strings"
)

// FileTypeNode represents a folder and associated file types
type FileTypeNode struct {
	Name       string
	Extensions []string        // File extensions associated with this folder
	Parent     *FileTypeNode   // Reference to the parent node, added here
	Children   []*FileTypeNode // Sub-categories or sub-folders for nested types
}

type FileTypeTree struct {
	Root *FileTypeNode
}

// NewFileTypeNode initializes a FileTypeNode
func NewFileTypeNode(name string) *FileTypeNode {
	return &FileTypeNode{
		Name:       name,
		Extensions: []string{},
		Children:   []*FileTypeNode{},
	}
}

// NewFileTypeTree initializes a FileTypeTree with a root node
func NewFileTypeTree() *FileTypeTree {
	return &FileTypeTree{
		Root: NewFileTypeNode("root"), // Root node representing the top-level file organization
	}
}

// AllowsExtension checks if this node allows files with the given extension.
func (n *FileTypeNode) AllowsExtension(ext string) bool {
	for _, allowedExt := range n.Extensions {
		if allowedExt == ext {
			return true
		}
	}
	return false
}

func (n *FileTypeNode) FindExtension(ext string) bool {
	if n.AllowsExtension(ext) {
		return true
	}

	for _, child := range n.Children {
		if child.FindExtension(ext) {
			return true
		}
	}

	return false
}

// AllowsExtension checks if this node allows files with the given extension.
func (n *FileTypeTree) AllowsExtension(ext string) bool {
	return n.Root.AllowsExtension(ext)
}

// FindOrCreatePath traverses the tree to find or create a directory path.
func (tree *FileTypeTree) FindOrCreatePath(path []string) *FileTypeNode {
	current := tree.Root

	for _, dir := range path {
		if dir == "" {
			continue
		}

		// Look for the directory among the current node's children
		var next *FileTypeNode
		for _, child := range current.Children {
			if child.Name == dir {
				next = child
				break
			}
		}

		// If the directory does not exist, create it and set the parent
		if next == nil {
			next = current.AddChild(dir)
			next.Parent = current
		}

		// Move to the next level in the tree
		current = next
	}

	return current
}

func (filetype *FileTypeNode) String() string {
	return filetype.Name
}

func (filetype *FileTypeNode) IsRoot() bool {
	return filetype.Name == "root"
}

func (filetype *FileTypeNode) IsLeaf() bool {
	return len(filetype.Children) == 0
}

func (filetype *FileTypeNode) Flatten() []string {
	var filetypes []string
	flattenFileTypeNode(filetype, filetype.Name, &filetypes)
	return filetypes
}

func flattenFileTypeNode(node *FileTypeNode, currentFileType string, filetypes *[]string) {
	if len(node.Extensions) > 0 {
		*filetypes = append(*filetypes, currentFileType)
	}

	for _, child := range node.Children {
		flattenFileTypeNode(child, currentFileType+"/"+child.Name, filetypes)
	}
}

// AddChild adds a new child folder or sub-category to this node
func (filetype *FileTypeNode) AddChild(name string) *FileTypeNode {
	child := NewFileTypeNode(name)
	filetype.Children = append(filetype.Children, child)
	return child
}

// AddExtensions adds file extensions to the current node
func (filetype *FileTypeNode) AddExtensions(extensions []string) {
	filetype.Extensions = append(filetype.Extensions, extensions...)
}

// PopulateFileTypes builds the file type tree based on a set of rules
// Example input: map[string][]string{"Docs/Reports": {".docx", ".pdf"}, "Photos": {".jpg", ".png"}}
func (tree *FileTypeTree) PopulateFileTypes(fileTypeRules map[string][]string) {
	for path, extensions := range fileTypeRules {
		tree.addDirectPath(path, extensions)
		slog.Debug(fmt.Sprintf("Added path: %s with extensions: %v", path, extensions))
	}
}

// addDirectPath creates a final node in FileTypeTree with the given path and associates extensions with it.
func (tree *FileTypeTree) addDirectPath(path string, extensions []string) {
	// Split the path into directories, keeping it as a single direct path
	dirs := strings.Split(path, "/")
	current := tree.Root

	// Traverse or create each level until the final directory in the path
	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		// Check the case of the directory name
		slog.Debug(fmt.Sprintf("Adding directory: %s", dir))

		// Look for an existing child node with the same name
		var next *FileTypeNode
		for _, child := range current.Children {
			if child.Name == dir {
				next = child
				break
			}
		}

		// If not found, create a new node and link it as a child
		if next == nil {
			next = current.AddChild(dir)
			next.Parent = current
		}
		current = next
	}

	// Attach extensions at the last directory level
	current.AddExtensions(extensions)
}
