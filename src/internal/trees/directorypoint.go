package trees

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/spatial/kdtree"
)

type DirectoryPoint struct {
	Node     *DirectoryNode `json:"node"`
	Metadata kdtree.Point   `json:"metadata"`
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

// Validate checks if the DirectoryPoint is valid
func (d DirectoryPoint) Validate() error {
	if d.Node == nil {
		return fmt.Errorf("directory node cannot be nil")
	}
	if len(d.Metadata) == 0 {
		return fmt.Errorf("metadata cannot be empty")
	}
	return nil
}

// Distance calculates the Euclidean distance between two DirectoryPoints.
func (d DirectoryPoint) Distance(c kdtree.Comparable) float64 {
	other, ok := c.(DirectoryPoint)
	if !ok {
		return math.Inf(1)
	}

	if len(d.Metadata) != len(other.Metadata) {
		return math.Inf(1)
	}

	dist := 0.0
	for i := range d.Metadata {
		delta := d.Metadata[i] - other.Metadata[i]
		dist += delta * delta
	}
	return math.Sqrt(dist)
}
