package cmd

import (
	"github.com/rancherlabs/corral/cmd/config"
	cmdpackage "github.com/rancherlabs/corral/cmd/package"
	pkgcmd "github.com/rancherlabs/corral/pkg/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var output = pkgcmd.OutputFormatTable

func Execute() {
	var debug bool
	var trace bool

	rootCmd := &cobra.Command{
		Use:   "corral",
		Short: "Corral is a CLI tool for creating and packaging reproducible development environments.",
		Long:  "Corral is a CLI tool for creating and packaging reproducible development environments.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if trace {
				logrus.SetLevel(logrus.TraceLevel)
			} else if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Usage(); err != nil {
				logrus.Fatalln(err)
			}
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	//rootCmd.SetErr(logrus.StandardLogger().Out)

	rootCmd.AddCommand(
		config.NewCommandConfig(),
		NewCommandDelete(),
		NewCommandList(),
		NewCommandVars(),
		NewCommandCreate(),
		cmdpackage.NewCommandPackage())

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&trace, "trace", false, "Enable verboser logging")

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
}
