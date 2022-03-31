package cmd_package

import (
	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func NewCommandValidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate PACKAGE",
		Short: "Validate the given package's manifest and structure.",
		Args:  cobra.ExactArgs(1),
		Run:   validate,
	}

	return cmd
}

func validate(_ *cobra.Command, args []string) {
	pkg, err := _package.LoadPackage(args[0])
	if err != nil {
		logrus.Fatal(err)
	}

	if i, err := os.Stat(pkg.OverlayPath()); err != nil || !i.IsDir() {
		logrus.Fatal("overlay folder not found.")
	}

	for _, cmd := range pkg.Commands {
		if cmd.Module != "" {
			if i, err := os.Stat(pkg.TerraformModulePath(cmd.Module)); err != nil || !i.IsDir() {
				logrus.Fatalf("Terraform module [%s] not found.", cmd.Module)
			}
		}
	}

	err = pkg.ValidateDefaults()
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("package is valid")
}
