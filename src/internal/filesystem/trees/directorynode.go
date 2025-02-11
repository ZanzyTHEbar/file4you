package trees

import (
	"encoding/json"

	"github.com/google/uuid"
)

type NodeType int

const (
	Directory NodeType = iota
	File
)

// FileNode represents a file with metadata
type FileNode struct {
	ID        uuid.UUID `json:"id"`
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Extension string    `json:"extension"`
	Metadata  Metadata  `json:"metadata"`
}

type DirectoryNode struct {
	ID       string           `json:"id"` // Unique ID for each node
	Path     string           `json:"path"`
	Type     NodeType         `json:"type"`
	Parent   *DirectoryNode   `json:"-"` // Omit Parent from JSON output, use ID instead
	Children []*DirectoryNode `json:"-"` // Omit Children from JSON output, use IDs instead
	Files    []*FileNode      `json:"files"`
	Metadata Metadata         `json:"metadata"`
}

type directoryNodeJSON struct {
	ID          string   `json:"id"`
	Path        string   `json:"path"`
	Type        NodeType `json:"type"`
	ParentID    *string  `json:"parent_id,omitempty"`
	ChildrenIDs []string `json:"children_ids,omitempty"`
	FileIDs     []string `json:"files"`
	Metadata    Metadata `json:"metadata"`
}

var tempRelationshipMap = make(map[string]*directoryNodeJSON)

// NewDirectoryNode creates a new DirectoryNode
func NewDirectoryNode(path string, parent *DirectoryNode) *DirectoryNode {
	var nodeType NodeType
	if parent == nil {
		nodeType = Directory
	} else {
		nodeType = parent.Type
	}

	return &DirectoryNode{
		ID:       uuid.NewString(),
		Path:     path,
		Type:     nodeType,
		Parent:   parent,
		Children: []*DirectoryNode{},
		Files:    []*FileNode{},
		Metadata: Metadata{},
	}
}

// AddChildDirectory adds a child directory to the current directory
func (directorynode *DirectoryNode) AddChildDirectory(path string) *DirectoryNode {
	child := NewDirectoryNode(path, directorynode)
	directorynode.Children = append(directorynode.Children, child)
	return child
}

// AddFile adds a file to the current directory
func (directorynode *DirectoryNode) AddFile(file *FileNode) *DirectoryNode {
	directorynode.Files = append(directorynode.Files, file)
	return directorynode
}

func (directorynode *DirectoryNode) String() string {
	return directorynode.Path
}

func (directorynode *DirectoryNode) IsRoot() bool {
	return directorynode.Parent == nil
}

func (directorynode *DirectoryNode) IsLeaf() bool {
	return len(directorynode.Children) == 0
}

func (directorynode *DirectoryNode) IsFile() bool {
	return directorynode.Type == File
}

func (directorynode *DirectoryNode) IsDir() bool {
	return directorynode.Type == Directory
}

func (directorynode *DirectoryNode) MarshalJSON() ([]byte, error) {
	// Define an alias to avoid recursion and customize JSON structure, creates a copy thus breaking the reference.
	type Alias DirectoryNode
	var parentID *string
	if directorynode.Parent != nil {
		parentID = &directorynode.Parent.ID
	}

	// Collect child IDs
	childIDs := make([]string, len(directorynode.Children))
	for i, child := range directorynode.Children {
		childIDs[i] = child.ID
	}

	fileIDs := make([]string, len(directorynode.Files))
	for i, file := range directorynode.Files {
		fileIDs[i] = file.ID.String()
	}

	// Marshal custom JSON structure
	return json.Marshal(&struct {
		*Alias
		ParentID    *string  `json:"parent_id,omitempty"`
		ChildrenIDs []string `json:"children_ids,omitempty"`
		FileIDs     []string `json:"file_ids,omitempty"`
	}{
		Alias:       (*Alias)(directorynode),
		ParentID:    parentID,
		ChildrenIDs: childIDs,
		FileIDs:     fileIDs,
	})
}

func (directorynode *DirectoryNode) UnMarshalJSON(data []byte) error {
	var aux directoryNodeJSON

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	directorynode.ID = aux.ID
	directorynode.Path = aux.Path
	directorynode.Type = aux.Type
	directorynode.Metadata = aux.Metadata

	tempRelationshipMap[directorynode.ID] = &aux

	return nil
}

// RebuildGraph builds relationships between nodes based on parent and children IDs
func RebuildGraph(nodes []*DirectoryNode) {
	nodeMap := make(map[string]*DirectoryNode)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	for id, temp := range tempRelationshipMap {
		node := nodeMap[id]

		if temp.ParentID != nil {
			if parent, ok := nodeMap[*temp.ParentID]; ok {
				node.Parent = parent
				parent.Children = append(parent.Children, node)
			}
		}

		for _, childID := range temp.ChildrenIDs {
			if child, ok := nodeMap[childID]; ok {
				node.Children = append(node.Children, child)
				child.Parent = node
			}
		}
	}

	// Clear the temporary map to free memory
	tempRelationshipMap = nil
}
