package indexing

import (
	roaring "github.com/RoaringBitmap/roaring"
)

// AttributeBitmaps holds roaring bitmaps keyed by attribute value id.
// Example: ExtID -> bitmap of PathIDs with that extension.
type AttributeBitmaps struct {
	Ext map[uint32]*roaring.Bitmap
}

func NewAttributeBitmaps() *AttributeBitmaps {
	return &AttributeBitmaps{Ext: make(map[uint32]*roaring.Bitmap)}
}

func (ab *AttributeBitmaps) AddExt(extID uint32, pid PathID) {
	bm, ok := ab.Ext[extID]
	if !ok {
		bm = roaring.New()
		ab.Ext[extID] = bm
	}
	bm.Add(uint32(pid))
}

// AndExt returns the intersection of multiple extension bitmaps.
func (ab *AttributeBitmaps) AndExt(extIDs ...uint32) *roaring.Bitmap {
	if len(extIDs) == 0 {
		return roaring.New()
	}
	// copy first
	res := ab.clone(ab.Ext[extIDs[0]])
	for _, id := range extIDs[1:] {
		res.And(ab.Ext[id])
	}
	return res
}

func (ab *AttributeBitmaps) clone(b *roaring.Bitmap) *roaring.Bitmap {
	if b == nil {
		return roaring.New()
	}
	c := roaring.New()
	c.Or(b) // copy
	return c
}
