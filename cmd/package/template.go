package cmd_package

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"

	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const templateDescription = `
Create a package from existing package(s).

Examples:
corral package template a b c OUT 
corral package template --description "my description" a b c OUT
`

func NewCommandTemplate() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "template",
		Short: "Create a package from a template",
		Long:  templateDescription,
		Run: func(cmd *cobra.Command, args []string) {
			template(args[len(args)-1], description, args[:len(args)-1]...)
		},
		Args: cobra.MinimumNArgs(2),
	}

	cmd.Flags().StringVar(&description, "description", "", "description of the rendered package")

	return cmd
}

// embed a template
func template(name, description string, packages ...string) {
	pkgs := make([]_package.Package, len(packages))

	for i, p := range packages {
		pkg, err := _package.LoadPackage(p) // ensures pkg is in cache
		if err != nil {
			logrus.Fatalf("failed to load [%s] package", p)
		}
		pkgs[i] = pkg
	}

	manifest, err := _package.MergePackages(name, pkgs)
	if err != nil {
		logrus.Fatal(err)
	}

	if description == "" {
		for i := range pkgs {
			if i > 0 {
				description += "\n"
			}

			if pkgs[i].Description != "" {
				description += pkgs[i].Description
			}
		}
	}
	manifest.Name = filepath.Base(name)
	manifest.Description = description

	buf, _ := yaml.Marshal(manifest)

	err = os.WriteFile(filepath.Join(name, "manifest.yaml"), buf, 0664)
	if err != nil {
		logrus.Fatal("failed to write manifest: ", err)
	}

	err = _package.ValidateManifest(buf)
	if err != nil {
		logrus.Fatal("rendered package is not a valid package: ", err)
	}
}
