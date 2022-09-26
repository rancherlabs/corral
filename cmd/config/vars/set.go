package vars

import (
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/vars"
	"github.com/spf13/cobra"
)

func NewCommandSet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set NAME VALUE",
		Short: "Create or update global variable.",
		Long:  "Create or update global variable.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.MustLoad()
			err := CreateVar(&cfg, args[0], args[1])
			if err != nil {
				return err
			}
			return cfg.Save()
		},
	}

	return cmd
}

func CreateVar(cfg *config.Config, key, value string) error {
	v, err := vars.FromJson(value)
	if err != nil {
		return err
	}

	cfg.Vars[key] = v
	return nil
}
