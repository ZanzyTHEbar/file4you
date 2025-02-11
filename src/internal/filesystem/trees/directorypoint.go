package trees

import (
	"math"

	"gonum.org/v1/gonum/spatial/kdtree"
)

type DirectoryPoint struct {
	Node     *DirectoryNode
	Metadata kdtree.Point
}

// Compare performs axis comparisons for KD-Tree.
func (d DirectoryPoint) Compare(comparable kdtree.Comparable, dim kdtree.Dim) float64 {
	other := comparable.(DirectoryPoint)
	return d.Metadata[dim] - other.Metadata[dim]
}

// Dims returns the number of dimensions in the metadata point.
func (d DirectoryPoint) Dims() int {
	return len(d.Metadata)
}

// Distance calculates the Euclidean distance between two DirectoryPoints.
func (d DirectoryPoint) Distance(c kdtree.Comparable) float64 {
	other := c.(DirectoryPoint)
	dist := 0.0
	for i := 0; i < d.Dims(); i++ {
		delta := d.Metadata[i] - other.Metadata[i]
		dist += delta * delta
	}
	return math.Sqrt(dist)
}
