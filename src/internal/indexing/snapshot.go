package indexing

import (
	"encoding/binary"
	"os"
	"time"
)

// PersistSnapshot writes a minimal binary columnar snapshot to path.
// Format (versioned, little-endian):
// [magic 'F4SN'] [u32 version] [u64 n] [u64 buildUnix]
// For each column present, a presence bitset could be used; here we write sizes, modtimes, isDirs, extIDs, depths and omit paths for compactness.
func PersistSnapshot(path string, snap *ColumnarSnapshot) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	// header
	f.Write([]byte{'F', '4', 'S', 'N'})
	var u32 = func(v uint32) { _ = binary.Write(f, binary.LittleEndian, v) }
	var u64 = func(v uint64) { _ = binary.Write(f, binary.LittleEndian, v) }
	var s64 = func(v int64) { _ = binary.Write(f, binary.LittleEndian, v) }
	var u16 = func(v uint16) { _ = binary.Write(f, binary.LittleEndian, v) }
	var b = func(v bool) {
		if v {
			f.Write([]byte{1})
		} else {
			f.Write([]byte{0})
		}
	}

	u32(1) // version
	n := uint64(len(snap.Sizes))
	u64(n)
	u64(uint64(snap.Meta.BuildUnixSec))
	// extension dictionary
	extCount := uint32(len(snap.ExtDict))
	u32(extCount)
	for _, s := range snap.ExtDict {
		bts := []byte(s)
		u32(uint32(len(bts)))
		_, _ = f.Write(bts)
	}
	// sizes
	for _, v := range snap.Sizes {
		s64(v)
	}
	// modtimes
	for _, v := range snap.ModTimes {
		s64(v)
	}
	// isDirs
	for _, v := range snap.IsDirs {
		b(v)
	}
	// extIDs
	for _, v := range snap.ExtIDs {
		_ = binary.Write(f, binary.LittleEndian, v)
	}
	// depths
	for _, v := range snap.Depths {
		u16(v)
	}
	return nil
}

// LoadSnapshot reads a snapshot persisted with PersistSnapshot.
func LoadSnapshot(path string) (*ColumnarSnapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, 4)
	if _, err := f.Read(buf); err != nil {
		return nil, err
	}
	if string(buf) != "F4SN" {
		return nil, os.ErrInvalid
	}
	var ver uint32
	if err := binary.Read(f, binary.LittleEndian, &ver); err != nil {
		return nil, err
	}
	var n uint64
	if err := binary.Read(f, binary.LittleEndian, &n); err != nil {
		return nil, err
	}
	var build uint64
	if err := binary.Read(f, binary.LittleEndian, &build); err != nil {
		return nil, err
	}
	snap := &ColumnarSnapshot{Meta: SnapshotMeta{BuildUnixSec: int64(build)}}
	// read extension dictionary
	var extCount uint32
	if err := binary.Read(f, binary.LittleEndian, &extCount); err != nil {
		return nil, err
	}
	if extCount > 0 {
		snap.ExtDict = make([]string, extCount)
		for i := uint32(0); i < extCount; i++ {
			var slen uint32
			if err := binary.Read(f, binary.LittleEndian, &slen); err != nil {
				return nil, err
			}
			if slen > 0 {
				buf2 := make([]byte, slen)
				if _, err := f.Read(buf2); err != nil {
					return nil, err
				}
				snap.ExtDict[i] = string(buf2)
			}
		}
	}
	snap.Sizes = make([]int64, n)
	for i := range snap.Sizes {
		if err := binary.Read(f, binary.LittleEndian, &snap.Sizes[i]); err != nil {
			return nil, err
		}
	}
	snap.ModTimes = make([]int64, n)
	for i := range snap.ModTimes {
		if err := binary.Read(f, binary.LittleEndian, &snap.ModTimes[i]); err != nil {
			return nil, err
		}
	}
	snap.IsDirs = make([]bool, n)
	for i := range snap.IsDirs {
		var x [1]byte
		if _, err := f.Read(x[:]); err != nil {
			return nil, err
		}
		snap.IsDirs[i] = x[0] == 1
	}
	snap.ExtIDs = make([]uint32, n)
	for i := range snap.ExtIDs {
		if err := binary.Read(f, binary.LittleEndian, &snap.ExtIDs[i]); err != nil {
			return nil, err
		}
	}
	snap.Depths = make([]uint16, n)
	for i := range snap.Depths {
		if err := binary.Read(f, binary.LittleEndian, &snap.Depths[i]); err != nil {
			return nil, err
		}
	}
	return snap, nil
}

// BuildColumnar constructs a snapshot from file records and dictionary encoders.
func BuildColumnar(records []FileRecord, extDict map[string]uint32) *ColumnarSnapshot {
	n := len(records)
	snap := &ColumnarSnapshot{
		Meta:     SnapshotMeta{NumFiles: n, BuildUnixSec: time.Now().Unix()},
		Sizes:    make([]int64, n),
		ModTimes: make([]int64, n),
		IsDirs:   make([]bool, n),
		ExtIDs:   make([]uint32, n),
		Depths:   make([]uint16, n),
	}
	for i, r := range records {
		snap.Sizes[i] = r.Size
		snap.ModTimes[i] = r.ModTime.Unix()
		snap.IsDirs[i] = r.IsDir
		snap.ExtIDs[i] = r.ExtID
		snap.Depths[i] = r.Depth
	}
	// construct reverse extension dict for persistence
	if len(extDict) > 0 {
		var maxID uint32
		for _, id := range extDict {
			if id > maxID {
				maxID = id
			}
		}
		rev := make([]string, maxID+1)
		for s, id := range extDict {
			rev[id] = s
		}
		snap.ExtDict = rev
	}
	return snap
}
