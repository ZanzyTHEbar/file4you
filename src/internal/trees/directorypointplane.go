package trees

import "gonum.org/v1/gonum/spatial/kdtree"

// DirectoryPointPlane wraps DirectoryPointCollection for sorting by a specific dimension.
type DirectoryPointPlane struct {
	Dim    kdtree.Dim               // Dimension to sort by
	Points DirectoryPointCollection // Collection of DirectoryPoints
}

// Len returns the length of the DirectoryPoint collection.
func (p DirectoryPointPlane) Len() int {
	return len(p.Points)
}

// Swap exchanges two points.
func (p DirectoryPointPlane) Swap(i, j int) {
	p.Points[i], p.Points[j] = p.Points[j], p.Points[i]
}

// Less returns true if point i is less than point j in the specified dimension.
func (p DirectoryPointPlane) Less(i, j int) bool {
	return p.Points[i].Metadata[p.Dim] < p.Points[j].Metadata[p.Dim]
}

// Slice returns a subset of DirectoryPointPlane from start to end.
// This allows DirectoryPointPlane to fulfill the kdtree.SortSlicer interface.
func (p DirectoryPointPlane) Slice(start, end int) kdtree.SortSlicer {
	return DirectoryPointPlane{
		Dim:    p.Dim,
		Points: p.Points[start:end],
	}
}
