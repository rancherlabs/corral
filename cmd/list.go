package cmd

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/corral"
	"github.com/spf13/cobra"
)

func NewCommandList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all corrals on this system.",
		Long:  "List all corrals on this system.",
		Run:   list,
	}

	return cmd
}

func list(_ *cobra.Command, _ []string) {
	corralNames, _ := os.ReadDir(config.CorralRoot("corrals"))

	tbl := table.NewWriter()
	tbl.SetOutputMirror(os.Stdout)
	tbl.AppendHeader(table.Row{"NAME", "STATUS"})
	tbl.AppendSeparator()
	for _, entry := range corralNames {
		c, err := corral.Load(config.CorralRoot("corrals", entry.Name()))
		if err != nil {
			continue
		}

		tbl.AppendRow(table.Row{c.Name, c.Status.String()})
	}
	tbl.Render()
}
