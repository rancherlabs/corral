package cmd_package

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCommandPackage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Commands related to managing packages.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Usage(); err != nil {
				logrus.Fatalln(err)
			}
		},
	}

	cmd.AddCommand(
		NewCommandPublish(),
		NewCommandLogin(),
		NewCommandInfo(),
		NewCommandValidate(),
		NewCommandDownload())
	return cmd
}
