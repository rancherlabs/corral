package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/corral"
	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const deleteDescription = `
Delete the given corral(s) and the associated infrastructure. If multiple corrals are given they will be deleted in
the order they appear one at a time.
`

func NewCommandDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [NAME [NAME ...]",
		Short: "Delete the given corral(s) and the associated infrastructure.",
		Long:  deleteDescription,
		Args:  cobra.MinimumNArgs(1),
		Run:   deleteCorral,
	}

	cmd.Flags().Bool("dry-run", false, "Display what resources this corral will create.")
	_ = cfgViper.BindPFlag("dry-run", cmd.Flags().Lookup("dry-run"))

	return cmd
}

func deleteCorral(_ *cobra.Command, args []string) {
	for _, name := range args {
		c, err := corral.Load(config.CorralPath(name))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				logrus.Warnf("skipping corral [%s], does not exist", name)
				continue
			} else {
				logrus.Fatal(err)
			}
		}

		c.SetStatus(corral.StatusDeleting)

		logrus.Infof("running terraform destroy for [%s]", name)
		pkg, err := _package.LoadPackage(c.Source)
		if err != nil {
			logrus.Error("could not load package associated with corral: ", err)
		}

		tf, err := config.NewTerraform(c.TerraformPath(), pkg.TerraformVersion())
		if err != nil {
			logrus.Fatal("failed to load terraform: ", err)
		}
		if debug {
			tf.SetStdout(os.Stdout)
		}

		err = tf.Destroy(context.Background())
		if err != nil {
			logrus.Errorf("failed to cleanup corral resources [%s]: %s", c.Name, err)
			continue
		}

		err = c.Delete()
		if err != nil {
			logrus.Errorf("failed to cleanup corral [%s]: %s", c.Name, err)
		}
	}
}
