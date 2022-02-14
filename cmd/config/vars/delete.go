package vars

import (
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCommandDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [NAME | [NAME...]]",
		Short: "Remove existing global variable.",
		Long:  "Remove existing global variable.",
		Args:  cobra.MinimumNArgs(1),
		Run:   deleteVar,
	}

	return cmd
}

func deleteVar(_ *cobra.Command, args []string) {
	cfg := config.MustLoad()

	for _, arg := range args {
		delete(cfg.Vars, arg)
	}

	err := cfg.Save()
	if err != nil {
		logrus.Fatalf("%e", err)
	}
}
