package cmd_package

import (
	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/spf13/cobra"
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
		Use:   "template PACKAGE[S] NAME",
		Short: "Create a package from a template",
		Long:  templateDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			return _package.Template(args[len(args)-1], description, args[:len(args)-1]...)
		},
		Args: cobra.MinimumNArgs(2),
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "description of the rendered package")

	return cmd
}
