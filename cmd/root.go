package cmd

import (
	"github.com/rancherlabs/corral/cmd/config"
	cmd_package "github.com/rancherlabs/corral/cmd/package"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var debug bool

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "corral",
		Short: "Corral is a CLI tool for creating and packaging reproducible development environments.",
		Long:  "Corral is a CLI tool for creating and packaging reproducible development environments.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Usage(); err != nil {
				logrus.Fatalln(err)
			}
		},
	}

	rootCmd.AddCommand(
		config.NewCommandConfig(),
		NewCommandDelete(),
		NewCommandList(),
		NewCommandVars(),
		NewCommandCreate(),
		cmd_package.NewCommandPackage())

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable verbose logging.")

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
}
