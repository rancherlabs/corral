package cmd

import (
    "fmt"

    "github.com/rancherlabs/corral/pkg/config"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "oras.land/oras-go/pkg/auth"
    dockerauth "oras.land/oras-go/pkg/auth/docker"
)

func NewCommandLogin() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "login REGISTRY",
        Short: "Login to an OCI registry.",
        Args:  cobra.ExactArgs(1),
        PreRun: func(_ *cobra.Command, _ []string) {
            cfg = config.Load()
        },
        Run: login,
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

    da, err := dockerauth.NewClient(cfg.RegistryCredentialsFile())
    if err != nil {
        logrus.Fatalf("error getting auth client: %s", err)
    }

    err = da.LoginWithOpts(auth.WithLoginHostname(args[0]),
        auth.WithLoginUsername(username),
        auth.WithLoginSecret(password),
        auth.WithLoginUserAgent(corralUserAgent))
    if err != nil {
        logrus.Fatalf("error authenticating with registry: %s", err)
    }

    logrus.Info("success!")
}
