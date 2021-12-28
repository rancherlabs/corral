package cmd_package

import (
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfg config.Config

const corralUserAgent = "Corral/" + version.Version

func NewCommandPackage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Commands related to managing packages.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Usage(); err != nil {
				logrus.Fatalln(err)
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cfg = config.Load()
		},
	}

	cmd.AddCommand(
		NewCommandPublish(),
		NewCommandLogin(),
		NewCommandInfo(),
		NewCommandValidate())
	return cmd
}
