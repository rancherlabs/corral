package cmd

import (
    "fmt"
    "os"
    "os/user"
    "path/filepath"

    "github.com/rancherlabs/corral/pkg/corral"
    "github.com/rancherlabs/corral/pkg/install"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

func NewCommandConfig() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "config",
        Short: "Set global configuration and download dependencies.",
        Long:  "Set global configuration and download dependencies.",
        Run:   configCorral,
    }

    cmd.Flags().String("user_id", cfg.UserID, "The user id is used by packages to help identify resources.")
    cmd.Flags().String("public_key", cfg.UserPublicKeyPath, "Path to a public key you want packages to install on nodes.")
    cmd.Flags().StringArrayP("variable", "v", []string{}, "Global variable applied to all corrals.")

    return cmd
}

func configCorral(cmd *cobra.Command, _ []string) {
    userId, _ := cmd.Flags().GetString("user_id")
    userPublicKeyPath, _ := cmd.Flags().GetString("public_key")

    if userId == "" {
        u, _ := user.Current()
        if u != nil {
            userId = u.Username
        }

        if input := prompt(fmt.Sprintf("How should packages identify you(%s): ", userId)); len(input) > 0 {
            userId = input
        }
    }

    if userPublicKeyPath == "" {
        userRoot, _ := os.UserHomeDir()
        if userRoot != "" {
            userPublicKeyPath = filepath.Join(userRoot, ".ssh", "id_rsa.pub")
        }

        if input := prompt(fmt.Sprintf("What ssh public key should packages use (%s): ", userPublicKeyPath)); len(input) > 0 {
            userPublicKeyPath = input
        }
    }

    globalVars := map[string]string{}
    globalVarsSlice, _ := cmd.Flags().GetStringArray("variable")
    for _, raw := range globalVarsSlice {
        k, v := corral.ToVar(raw)
        if k == "" {
            logrus.Fatal("variables should be in the format <key>=<value>")
        }
        globalVars[k] = v
    }

    logrus.Info("installing corral, this can take a minute")

    err := install.Install(userId, userPublicKeyPath, globalVars)
    if err != nil {
        logrus.Fatal("error saving configuration: ", err)
    }

    logrus.Info("corral installed successfully!")
}

func prompt(message string) string {
    var buf string

    print(message)
    _, _ = fmt.Scanln(&buf)

    return buf
}
