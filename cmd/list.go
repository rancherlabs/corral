package cmd

import (
	"fmt"
	"os"

	pkgcmd "github.com/rancherlabs/corral/pkg/cmd"
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/corral"
	"github.com/spf13/cobra"
)

func NewCommandList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all corrals on this system.",
		Long:  "List all corrals on this system.",
		RunE:  list,
	}

	cmd.Flags().VarP(&output, "output", "o", "Output format. One of: table|json|yaml")

	return cmd
}

func list(_ *cobra.Command, _ []string) error {
	corralNames, _ := os.ReadDir(config.CorralRoot("corrals"))

	corrals := map[string]string{}
	for _, entry := range corralNames {
		c, err := corral.Load(config.CorralRoot("corrals", entry.Name()))
		if err != nil {
			continue
		}

		corrals[c.Name] = c.Status.String()
	}

	out, err := pkgcmd.Output(corrals, output, pkgcmd.OutputOptions{
		Key:   "NAME",
		Value: "STATUS",
	})
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
