package cli

import (
	"file4you/internal/db"
	"file4you/internal/deskfs"
	"file4you/internal/terminal"

	"github.com/spf13/cobra"
)

type CmdParams struct {
	Term      *terminal.Terminal
	DeskFS    *deskfs.DesktopFS
	Palette   []*cobra.Command
	CentralDB *db.CentralDBProvider
}

type File4YouCMD struct {
	Root *cobra.Command
}

func NewFile4YouCMD(cmdRoot *cobra.Command) *File4YouCMD {
	return &File4YouCMD{
		Root: cmdRoot,
	}
}
