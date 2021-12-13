package cmd

import (
    "github.com/rancherlabs/corral/pkg/config"
    _package "github.com/rancherlabs/corral/pkg/package"
    "github.com/rancherlabs/corral/pkg/version"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

const publishDescription = `
Upload the package found at the given target to the given registry.

Examples:
corral publish /home/rancher/my_pkg ghcr.io/rancher/my_pkg:latest
`

func NewCommandPublish() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "publish TARGET REFERENCE",
        Short: "Upload a package to an OCI registry.",
        Long:  publishDescription,
        Run:   publish,
        Args:  cobra.ExactArgs(2),
        PreRun: func(_ *cobra.Command, _ []string) {
            cfg = config.Load()
        },
    }

    return cmd
}

const corralUserAgent = "Corral/" + version.Version

func publish(_ *cobra.Command, args []string) {
    err := _package.UploadPackage(args[0], args[1], cfg.RegistryCredentialsFile())
    if err != nil {
        logrus.Fatal("failed to push package: %s", err)
    }

    logrus.Info("success")
}
