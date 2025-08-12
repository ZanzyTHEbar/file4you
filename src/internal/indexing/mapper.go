package indexing

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SimplePathIDMapper is a compact map-based mapper from canonicalized path -> PathID.
// It is designed to be replaced by an MPH-backed implementation without changing callers.
type SimplePathIDMapper struct {
	pathToID map[string]PathID
}

func NewSimplePathIDMapper() *SimplePathIDMapper {
	return &SimplePathIDMapper{pathToID: make(map[string]PathID)}
}

func (m *SimplePathIDMapper) Lookup(path string) (PathID, bool) {
	id, ok := m.pathToID[canonicalize(path)]
	return id, ok
}

func (m *SimplePathIDMapper) Size() int { return len(m.pathToID) }

// BuildFromDir assigns stable PathIDs in lexical path order for determinism.
// Returns the ordered list of canonicalized paths corresponding to PathIDs [0..N-1].
func (m *SimplePathIDMapper) BuildFromDir(root string) ([]string, error) {
	canonicalRoot := canonicalize(root)
	var paths []string
	// Collect all regular files and directories
	err := filepath.WalkDir(canonicalRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		cp := canonicalize(p)
		paths = append(paths, cp)
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Stable ordering for deterministic IDs
	sort.Strings(paths)
	m.pathToID = make(map[string]PathID, len(paths))
	for i, p := range paths {
		m.pathToID[p] = PathID(i)
	}
	return paths, nil
}

func canonicalize(p string) string {
	// Convert backslashes to forward slashes and clean components.
	p = strings.ReplaceAll(p, "\\", "/")
	// filepath.Clean preserves platform separators; convert after cleaning.
	p = filepath.ToSlash(filepath.Clean(p))
	// Drop trailing slash except for root
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}
