package vars

import (
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/spf13/cobra"
)

func NewVarsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vars [VAR | [VAR...]]",
		Short: "List and modify global configuration.",
		Long:  "List and modify global configuration.",
		Run:   listVars,
	}

	cmd.AddCommand(
		NewCommandSet(),
		NewCommandDelete())
	return cmd
}

func listVars(_ *cobra.Command, args []string) {
	cfg := config.MustLoad()

	if len(args) == 1 {
		println(cfg.Vars[args[0]])
		return
	}

	if len(args) > 1 {
		println("NAME\tVALUE")
		for _, k := range args {
			println(k + "\t" + cfg.Vars[k])
		}
		return
	}

	println("NAME\tVALUE")
	for k, v := range cfg.Vars {
		println(k + "\t" + v)
	}
}
