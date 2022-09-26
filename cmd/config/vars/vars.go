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
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.MustLoad()
			out, err := ListVars(cfg, output, args...)
			if err != nil {
				return err
			}
			fmt.Println(out)
			return nil
		},
	}

	cmd.Flags().VarP(&output, "output", "o", "Output format. One of: table|json|yaml")

	cmd.AddCommand(
		NewCommandSet(),
		NewCommandDelete())
	return cmd
}

func ListVars(cfg config.Config, output pkgcmd.OutputFormat, args ...string) (string, error) {
	if len(args) == 1 {
		return fmt.Sprintf("%v", cfg.Vars[args[0]]), nil
	}

	vars := map[string]any{}
	if len(args) > 1 {
		for _, k := range args {
			vars[k] = cfg.Vars[k]
		}
	} else {
		vars = cfg.Vars
	}

	return pkgcmd.Output(vars, output, pkgcmd.OutputOptions{
		Key:   "NAME",
		Value: "VALUE",
	})
}
