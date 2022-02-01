package cmd_package

import (
	"os"
	"path/filepath"

	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	if i, err := os.Stat(filepath.Join(args[0], "manifest.yaml")); err != nil || i.IsDir() {
		logrus.Fatal("Package manifest not found.")
	}

	if i, err := os.Stat(filepath.Join(args[0], "scripts")); err != nil || !i.IsDir() {
		logrus.Fatal("Scripts folder not found.")
	}

	if i, err := os.Stat(filepath.Join(args[0], "terraform", "module")); err != nil || !i.IsDir() {
		logrus.Fatal("Terraform module not found.")
	}

	_, err := _package.LoadPackage(args[0])
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("package is valid")
}
