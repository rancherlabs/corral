package cmd_package

import (
	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/spf13/cobra"
)

func NewCommandValidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate PACKAGE",
		Short: "Validate the given package's manifest and structure.",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return _package.Validate(args[0])
		},
	}

	return cmd
}
