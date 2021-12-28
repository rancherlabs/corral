package cmd_package

import (
	"github.com/rancherlabs/corral/pkg/config"
	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCommandInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info PACKAGE",
		Short: "Display details about the given package.",
		Args:  cobra.ExactArgs(1),
		PreRun: func(_ *cobra.Command, _ []string) {
			cfg = config.Load()
		},
		Run: info,
	}

	return cmd
}

func info(_ *cobra.Command, args []string) {
	pkg, err := _package.LoadPackage(args[0], cfg.PackageCachePath(), cfg.RegistryCredentialsFile())
	if err != nil {
		logrus.Fatal(err)
	}

	println(pkg.Name)
	println()
	println(pkg.Description)
	println()
	println("VARIABLE\tDESCRIPTION")
	for k, v := range pkg.VariableSchemas {
		println(k, "\t", v.Description)
	}
	println()
}
