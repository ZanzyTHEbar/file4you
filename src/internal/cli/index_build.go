package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"file4you/internal/indexing"

	"github.com/spf13/cobra"
)

func newIndexBuildCmd(params *CmdParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "index-build [DIR] [OUT]",
		Short:  "Experimental: build a columnar index snapshot for a directory",
		Hidden: true,
		Args:   cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			out := "snapshot.f4s"
			if len(args) > 1 {
				out = args[1]
			}
			dir = filepath.Clean(dir)
			// Build mapper
			mapper := indexing.NewSimplePathIDMapper()
			paths, err := mapper.BuildFromDir(dir)
			if err != nil {
				return err
			}
			// Collect records
			recs := make([]indexing.FileRecord, 0, len(paths))
			extDict := map[string]uint32{"": 0}
			nextExt := uint32(1)
			for _, p := range paths {
				info, err := os.Lstat(p)
				if err != nil {
					continue
				}
				ext := strings.ToLower(filepath.Ext(p))
				if _, ok := extDict[ext]; !ok {
					extDict[ext] = nextExt
					nextExt++
				}
				depth := uint16(strings.Count(strings.TrimPrefix(p, dir), string(os.PathSeparator)))
				recs = append(recs, indexing.FileRecord{
					Path:    p,
					Size:    sizeOf(info),
					ModTime: info.ModTime(),
					IsDir:   info.IsDir(),
					ExtID:   extDict[ext],
					Depth:   depth,
				})
			}
			// Sort by path for determinism (already sorted, but ensure)
			sort.SliceStable(recs, func(i, j int) bool { return recs[i].Path < recs[j].Path })
			snap := indexing.BuildColumnar(recs, extDict)
			if err := indexing.PersistSnapshot(out, snap); err != nil {
				return err
			}
			params.Interactor.Success(fmt.Sprintf("Built snapshot with %d entries at %s", len(recs), out))
			return nil
		},
	}
	return cmd
}

func sizeOf(fi os.FileInfo) int64 {
	if fi.IsDir() {
		return 0
	}
	return fi.Size()
}
