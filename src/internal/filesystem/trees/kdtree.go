package trees

import (
	"fmt"
	"log/slog"

	"gonum.org/v1/gonum/spatial/kdtree"
)

// BuildKDTree constructs the KD-Tree from the DirectoryTree’s nodes.
func (tree *DirectoryTree) BuildKDTree() {
	// Populate KDTreeData with DirectoryPoints
	tree.KDTreeData = DirectoryPointCollection{}
	tree.collectDirectoryPoints(tree.Root)

	// Create KD-Tree with populated data
	tree.KDTree = kdtree.New(tree.KDTreeData, false)
}

// InsertNodeToKDTree inserts a DirectoryNode into the KD-Tree.
func (tree *DirectoryTree) InsertNodeToKDTree(node *DirectoryNode) {
	// Create a DirectoryPoint from node metadata and add it to the collection

	metadataPoint, err := node.Metadata.ToKDTreePoint()
	if err != nil {
		slog.Error(fmt.Sprintf("Error converting metadata to KDTree point: %v", err))
		return
	}

	point := DirectoryPoint{
		Node:     node,
		Metadata: metadataPoint,
	}
	tree.KDTreeData = append(tree.KDTreeData, point)

	// Rebuild the KD-Tree to include the new point (can be optimized if necessary)
	tree.KDTree = kdtree.New(tree.KDTreeData, false)
}

// RangeSearchKDTree finds all nodes within a specified radius from the query point.
func (tree *DirectoryTree) RangeSearchKDTree(query DirectoryPoint, radius float64) []*DirectoryNode {
	keeper := kdtree.NewDistKeeper(radius * radius) // Using squared distance for radius
	tree.KDTree.NearestSet(keeper, query)

	var results []*DirectoryNode
	for _, item := range keeper.Heap {
		dirPoint := item.Comparable.(DirectoryPoint)
		results = append(results, dirPoint.Node)
	}
	return results
}

// NearestNeighborSearchKDTree finds the k nearest neighbors to the query point.
func (tree *DirectoryTree) NearestNeighborSearchKDTree(query DirectoryPoint, k int) []*DirectoryNode {
	keeper := kdtree.NewNKeeper(k)
	tree.KDTree.NearestSet(keeper, query)

	var results []*DirectoryNode
	for _, item := range keeper.Heap {
		dirPoint := item.Comparable.(DirectoryPoint)
		results = append(results, dirPoint.Node)
	}
	return results
}

// collectDirectoryPoints recursively collects DirectoryPoints for KD-Tree construction.
func (tree *DirectoryTree) collectDirectoryPoints(node *DirectoryNode) {
	if node == nil {
		return
	}

	if err := node.Metadata.Validate(); err != nil {
		return
	}

	if node.Metadata.NodeType != Directory {
		return
	}

	metadataPoint, err := node.Metadata.ToKDTreePoint()
	if err != nil {
		slog.Error(fmt.Sprintf("Error converting metadata to KDTree point: %v", err))
		return
	}

	// Convert node metadata to DirectoryPoint and add to KDTreeData
	point := DirectoryPoint{
		Node:     node,
		Metadata: metadataPoint,
	}
	tree.KDTreeData = append(tree.KDTreeData, point)

	// Recursively add child directories
	for _, child := range node.Children {
		tree.collectDirectoryPoints(child)
	}
}
