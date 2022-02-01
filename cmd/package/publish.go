package cmd_package

import (
	"time"

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
	}

	return cmd
}

func publish(_ *cobra.Command, args []string) {
	cfg := config.Load()

	pkg, err := _package.LoadPackage(args[0])
	if err != nil {
		logrus.Fatal("failed to load package: ", err)
	}

	setAnnotationIfEmpty(pkg.Annotations, _package.TerraformVersionAnnotation, version.TerraformVersion)
	setAnnotationIfEmpty(pkg.Annotations, _package.PublisherAnnotation, cfg.UserID)
	setAnnotationIfEmpty(pkg.Annotations, _package.CorralVersionAnnotation, version.Version)
	setAnnotationIfEmpty(pkg.Annotations, _package.PublishTimestampAnnotation, time.Now().UTC().Format(time.RFC3339))

	err = _package.UploadPackage(pkg, args[1])
	if err != nil {
		logrus.Fatal("failed to push package: ", err)
	}

	logrus.Info("success")
}

func setAnnotationIfEmpty(annotations map[string]string, key, value string) {
	if annotations != nil && annotations[key] == "" {
		annotations[key] = value
	}
}
