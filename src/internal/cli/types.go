package cli

import (
	"file4you/internal/db"
	"file4you/internal/deskfs"
	"file4you/internal/ui"

	"github.com/firebase/genkit/go/genkit"
	"github.com/spf13/cobra"
)

type CmdParams struct {
	Interactor ui.Interactor // New field
	DeskFS     *deskfs.DesktopFileSystem
	Palette    []*cobra.Command
	CentralDB  *db.CentralDBProvider
	Genkit     *genkit.Genkit
}

type File4YouCMD struct {
	Root *cobra.Command
}

func NewFile4YouCMD(cmdRoot *cobra.Command) *File4YouCMD {
	return &File4YouCMD{
		Root: cmdRoot,
	}
}
