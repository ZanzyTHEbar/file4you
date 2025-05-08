package trees

import (
	"fmt"
	"log/slog"

	"gonum.org/v1/gonum/spatial/kdtree"
)

// BuildKDTree constructs the KD-Tree from the DirectoryTree’s nodes.
// IMPORTANT: This K-D Tree implementation currently only indexes DirectoryNode objects
// (i.e., directories), not individual FileNode objects. This is due to the logic
// in collectDirectoryPoints which filters for node.Metadata.NodeType == Directory.
// Searches performed using this K-D Tree will therefore find directories that match
// the search criteria based on their metadata.
func (tree *DirectoryTree) BuildKDTree() {
	// Populate KDTreeData with DirectoryPoints
	tree.KDTreeData = DirectoryPointCollection{}
	tree.collectDirectoryPoints(tree.Root)

	// Create KD-Tree with populated data
	tree.KDTree = kdtree.New(tree.KDTreeData, false)
}

// FIXME: InsertNodeToKDTree inserts a DirectoryNode into the KD-Tree.
// Note: This will only effectively insert directories due to the filtering
// in collectDirectoryPoints, which is the source of data for the tree.
// If a FileNode were passed, its point might be added to KDTreeData but
// would be inconsistent with trees built by BuildKDTree unless
// collectDirectoryPoints is also changed.
// TODO: Consider adding a check here: if node.Metadata.NodeType != Directory, log/return.
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

	// Rebuild the KD-Tree to include the new point.
	// PERFORMANCE NOTE: Rebuilding the entire K-D tree on every single insertion
	// is computationally expensive (O(N log N) where N is total points).
	// This approach is acceptable if insertions are infrequent or batched.
	// For frequent, individual insertions, consider strategies like:
	//  1. Batching insertions and rebuilding the tree periodically.
	//  2. Investigating if the kdtree library or an alternative supports
	//     more efficient incremental updates (though many k-d tree implementations
	//     achieve best balance by periodic full rebuilds after many incremental changes).
	//
	// The comment "(can be optimized if necessary)" in previous versions acknowledged this.
	tree.KDTree = kdtree.New(tree.KDTreeData, false)
}

// RangeSearchKDTree finds all DirectoryNode objects (directories) within a specified
// radius from the query DirectoryPoint.
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

// NearestNeighborSearchKDTree finds the k nearest DirectoryNode objects (directories)
// to the query DirectoryPoint.
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
// IMPORTANT: This function currently only creates DirectoryPoint entries for
// nodes where node.Metadata.NodeType == Directory. This means the K-D Tree
// will only contain points representing directories, not individual files.
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
