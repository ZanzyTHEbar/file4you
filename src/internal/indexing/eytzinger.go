package indexing

// BuildEytzinger builds an eytzinger layout from sorted keys and parallel pathIDs.
// Reference: "Eytzinger Layout" (in-order to breadth-first array mapping for cache efficiency).
func BuildEytzinger(sorted []int64, ids []PathID) (layout []int64, idx []PathID) {
	n := len(sorted)
	layout = make([]int64, n)
	idx = make([]PathID, n)
	pos := 0
	var dfs func(i int)
	dfs = func(i int) {
		if i > n {
			return
		}
		dfs(i << 1)
		layout[i-1] = sorted[pos]
		idx[i-1] = ids[pos]
		pos++
		dfs((i << 1) | 1)
	}
	dfs(1)
	return
}

// EytzingerSearch performs a branchless search on an eytzinger array.
func EytzingerSearch(a []int64, x int64) int {
	i := 1
	n := len(a)
	for i <= n {
		if x >= a[i-1] {
			i = (i << 1) | 1
		} else {
			i = i << 1
		}
	}
	i >>= 1
	if i == 0 {
		return -1
	}
	if a[i-1] == x {
		return i - 1
	}
	// predecessor
	// return index of last <= x
	// walk up until parent is smaller
	j := i
	for j > 1 && (j&1) == 1 {
		j >>= 1
	}
	if j > 1 {
		j >>= 1
		return j - 1
	}
	return i - 1
}
