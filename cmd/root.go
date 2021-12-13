package cmd

import (
    "github.com/rancherlabs/corral/pkg/config"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

var cfg config.Config
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
        NewCommandConfig(),
        NewCommandDelete(),
        NewCommandList(),
        NewCommandPublish(),
        NewCommandVars(),
        NewCommandLogin(),
        NewCommandCreate())

    rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable verbose logging.")

    if err := rootCmd.Execute(); err != nil {
        logrus.Fatalln(err)
    }
}
