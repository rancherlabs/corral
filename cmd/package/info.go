package cmd_package

import (
	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCommandInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info PACKAGE",
		Short: "Display details about the given package.",
		Args:  cobra.ExactArgs(1),
		Run:   info,
	}

	return cmd
}

func info(_ *cobra.Command, args []string) {
	pkg, err := _package.LoadPackage(args[0])
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
