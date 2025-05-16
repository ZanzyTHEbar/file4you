package terminal

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// CmdDesc is a map of command names to their descriptions
var CmdDesc = map[string][2]string{}

func printCmds(w io.Writer, prefix string, colors []lipgloss.Color, cmds ...string) {
	if os.Getenv("FILE4YOU_DISABLE_SUGGESTIONS") != "" {
		return
	}

	for i, cmd := range cmds {
		config, ok := CmdDesc[cmd]
		if !ok {
			continue
		}

		alias := config[0]
		desc := config[1]

		if alias != "" {
			containsFull := strings.Contains(cmd, alias)
			if containsFull {
				cmd = strings.Replace(cmd, alias, fmt.Sprintf("(%s)", alias), 1)
			} else {
				cmd = fmt.Sprintf("%s (%s)", cmd, alias)
			}
		}

		styled := lipgloss.NewStyle().Foreground(colors[i%len(colors)])
		fmt.Fprintf(w, "%s%s 👉 %s\n", prefix, styled.Bold(true).Render(cmd), styled.Italic(true).Render(desc))
	}
}

func PrintCustomCmd(prefix, cmd, alias, desc string) {
	cmd = strings.Replace(cmd, alias, fmt.Sprintf("(%s)", alias), 1)
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF"))
	slog.Info(fmt.Sprintf("%s%s 👉 %s\n", prefix, styled.Bold(true).Render(cmd), styled.Italic(true).Render(desc)))
}
