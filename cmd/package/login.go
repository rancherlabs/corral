package cmd_package

import (
	"fmt"

	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCommandLogin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login REGISTRY",
		Short: "Login to an OCI registry.",
		Args:  cobra.ExactArgs(1),
		Run:   login,
	}

	cmd.Flags().String("username", "", "The username for the registry.")
	cmd.Flags().String("password", "", "The password for the user.")

	return cmd
}

func login(cmd *cobra.Command, args []string) {
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")

	if username == "" {
		username = prompt(fmt.Sprintf("username: "))
	}

	if password == "" {
		password = prompt(fmt.Sprintf("password: "))
	}

	err := _package.AddRegistryCredentials(args[0], username, password)
	if err != nil {
		logrus.Fatalf("error authenticating with registry: %s", err)
	}

	logrus.Info("success!")
}

func prompt(message string) string {
	var buf string

	print(message)
	_, _ = fmt.Scanln(&buf)

	return buf
}
