package vars

import (
	"fmt"

	pkgcmd "github.com/rancherlabs/corral/pkg/cmd"
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/spf13/cobra"
)

var output = pkgcmd.OutputFormatTable

func NewVarsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vars [VAR | [VAR...]]",
		Short: "List and modify global configuration.",
		Long:  "List and modify global configuration.",
		RunE:  listVars,
	}

	cmd.Flags().VarP(&output, "output", "o", "Output format. One of: table|json|yaml")

	cmd.AddCommand(
		NewCommandSet(),
		NewCommandDelete())
	return cmd
}

func listVars(cmd *cobra.Command, args []string) error {
	cfg := config.MustLoad()

	if len(args) == 1 {
		println(cfg.Vars[args[0]])
		return nil
	}

	vars := map[string]string{}
	if len(args) > 1 {
		for _, k := range args {
			vars[k] = cfg.Vars[k]
		}
	} else {
		vars = cfg.Vars
	}

	out, err := pkgcmd.Output(vars, output, pkgcmd.OutputOptions{
		Key:   "NAME",
		Value: "VALUE",
	})
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
