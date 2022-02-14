package vars

import (
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCommandSet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set NAME VALUE",
		Short: "Create or update global variable.",
		Long:  "Create or update global variable.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			createVar(args[0], args[1])
		},
	}

	return cmd
}

func createVar(key, value string) {
	cfg := config.MustLoad()

	cfg.Vars[key] = value

	err := cfg.Save()
	if err != nil {
		logrus.Fatalf("%e", err)
	}
}
