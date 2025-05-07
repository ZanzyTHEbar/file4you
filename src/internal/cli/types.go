package cli

import (
	"file4you/internal/db"
	"file4you/internal/deskfs"
	"file4you/internal/ui" // New import

	"github.com/firebase/genkit/go/genkit" // Added for Genkit instance
	"github.com/spf13/cobra"
)

type CmdParams struct {
	Interactor ui.Interactor // New field
	DeskFS     *deskfs.DesktopFS
	Palette    []*cobra.Command
	CentralDB  *db.CentralDBProvider
	Genkit     *genkit.Genkit // Added for Genkit instance
}

type File4YouCMD struct {
	Root *cobra.Command
}

func NewFile4YouCMD(cmdRoot *cobra.Command) *File4YouCMD {
	return &File4YouCMD{
		Root: cmdRoot,
	}
}
