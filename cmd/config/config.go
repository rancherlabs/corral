package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/rancherlabs/corral/cmd/config/vars"
	"github.com/rancherlabs/corral/pkg/config"
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

	cmd.Flags().String("user_id", "", "The user id is used by packages to help identify resources.")
	cmd.Flags().String("public_key", "", "Path to a public key you want packages to install on nodes.")

	cmd.AddCommand(vars.NewVarsCommand())

	return cmd
}

func configCorral(cmd *cobra.Command, _ []string) {
	cfg, _ := config.Load()
	userId, _ := cmd.Flags().GetString("user_id")
	if userId != "" {
		cfg.UserID = userId
	}
	userPublicKeyPath, _ := cmd.Flags().GetString("public_key")
	if userPublicKeyPath != "" {
		cfg.UserPublicKeyPath = userPublicKeyPath
	}

	if userId == "" {
		if cfg.UserID == "" {
			u, _ := user.Current()
			if u != nil {
				cfg.UserID = u.Username
			}
		}

		if input := prompt(fmt.Sprintf("How should packages identify you(%s): ", cfg.UserID)); len(input) > 0 {
			cfg.UserID = input
		}
	}

	if userPublicKeyPath == "" {
		if cfg.UserPublicKeyPath == "" {
			userRoot, _ := os.UserHomeDir()
			if userRoot != "" {
				cfg.UserPublicKeyPath = filepath.Join(userRoot, ".ssh", "id_rsa.pub")
			}
		}

		if input := prompt(fmt.Sprintf("What ssh public key should packages use (%s): ", cfg.UserPublicKeyPath)); len(input) > 0 {
			cfg.UserPublicKeyPath = input
		}
	}

	logrus.Info("installing corral, this can take a minute")

	if err := config.Install(); err != nil {
		logrus.Fatal(err)
	}

	if err := cfg.Save(); err != nil {
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
