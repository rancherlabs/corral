package cmd_package

import (
	"os"
	"path/filepath"

	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const templateDescription = `
Create a package from existing package(s).

Examples:
corral package template a b c OUT 
corral package template -f config.yaml
`

func NewCommandTemplate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Create a package from a template",
		Long:  templateDescription,
		Run: func(cmd *cobra.Command, args []string) {
			err := template(cmd, args)
			if err != nil {
				logrus.Fatalf("Error rendering package template: %v", err)
			}
		},
		Args: cobra.MinimumNArgs(1),
	}

	cmd.Flags().StringP("file", "f", "", "yaml file to define template values")

	return cmd
}

// embed a template
func template(cmd *cobra.Command, args []string) error {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	body, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var t _package.TemplateSpec

	if err = yaml.Unmarshal(body, &t); err != nil {
		return err
	}

	srcs := append(t.Packages, args[:len(args)-1]...)
	pkgs := make([]_package.Package, len(srcs))

	for i, p := range srcs {
		pkg, err := _package.LoadPackage(p) // ensures pkg is in cache
		if err != nil {
			return err
		}
		pkgs[i] = pkg
	}

	name := args[len(args)-1]

	manifest, err := _package.MergePackages(name, pkgs)
	if err != nil {
		return err
	}

	manifest.Description = t.Description

	buf, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	err = _package.ValidateManifest(buf)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(name, "manifest.yaml"), buf, 0664)
	if err != nil {
		return err
	}

	return nil
}
