package trees

import "gonum.org/v1/gonum/spatial/kdtree"

// DirectoryPointCollection is a collection of DirectoryPoint structs that implements kdtree.Interface.
type DirectoryPointCollection []DirectoryPoint

// Index returns the DirectoryPoint at index i.
func (d DirectoryPointCollection) Index(i int) kdtree.Comparable {
	return d[i]
}

// Len returns the length of the collection.
func (d DirectoryPointCollection) Len() int {
	return len(d)
}

// Slice returns a subset of the collection between start and end indices.
func (d DirectoryPointCollection) Slice(start, end int) kdtree.Interface {
	return d[start:end]
}

// Pivot finds the median for a specified dimension.
func (d DirectoryPointCollection) Pivot(dim kdtree.Dim) int {
	plane := DirectoryPointPlane{Dim: dim, Points: d}
	return kdtree.Partition(plane, kdtree.MedianOfMedians(plane))
}

