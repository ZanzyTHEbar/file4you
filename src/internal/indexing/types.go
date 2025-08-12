package indexing

import (
	"time"
)

// PathID is a stable identifier for a path within a snapshot.
// It is intentionally small and contiguous to support compact columnar layouts
// and roaring bitmap usage.
type PathID = uint64

// FileRecord holds the minimal attributes needed for core indexing operations.
// Additional attributes can be joined out-of-band via the central/workspace DBs.
type FileRecord struct {
	Path    string
	Size    int64
	ModTime time.Time
	IsDir   bool
	ExtID   uint32 // small integer for extension dictionary
	Depth   uint16 // directory depth
}

// SnapshotMeta captures summary information for a built snapshot.
type SnapshotMeta struct {
	NumFiles     int
	NumDirs      int
	BuildUnixSec int64
}

// PathIDMapper maps canonicalized paths to stable PathIDs.
// Implementations may use Minimal Perfect Hash (MPH) or a fallback map.
type PathIDMapper interface {
	Lookup(path string) (PathID, bool)
	Size() int
}

// ColumnarSnapshot is the in-memory representation of the on-disk snapshot.
// It is optimized for scan-heavy workloads and vectorized computations.
type ColumnarSnapshot struct {
	Meta SnapshotMeta

	// Dictionaries
	ExtDict []string // extId -> ".pdf" etc

	// Core columns (same length, indexed by PathID)
	Paths    []string // optional (may be omitted in persisted snapshot)
	Sizes    []int64  // bytes
	ModTimes []int64  // unix seconds
	IsDirs   []bool   // directories vs files
	ExtIDs   []uint32 // extension id
	Depths   []uint16 // directory depth

	// Indexes/accelerators (optional)
	// Eytzinger-order arrays for cache-efficient numeric search
	EytzingerSizes []int64
	EytzingerIdx   []PathID // back-reference from Eytzinger position -> PathID

	// Roaring bitmaps keyed by attribute values (e.g., extId)
	// Managed by a separate component to keep ColumnarSnapshot lightweight.
}
