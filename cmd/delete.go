package cmd

import (
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

	cmd.Flags().Bool("skip-cleanup", false, "Do not run terraform destroy just delete the package.  This can result in un-tracked infrastructure resources!")

	return cmd
}

func deleteCorral(cmd *cobra.Command, args []string) {
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

		if skip, _ := cmd.Flags().GetBool("skip-cleanup"); !skip {
			logrus.Infof("cleaning up corral: %s", name)
			pkg, err := _package.LoadPackage(c.Source)
			if err != nil {
				logrus.Error("could not load package associated with corral: ", err)
			}

			for i := len(pkg.Commands) - 1; i >= 0; i-- {
				if pkg.Commands[i].Module != "" {
					if pkg.Commands[i].SkipCleanup {
						continue
					}

					logrus.Debugf("destroying module: %s", pkg.Commands[i].Module)
					if err = c.DestroyModule(pkg.Commands[i].Module); err != nil {
						logrus.Errorf("failed to cleanup module [%s]: %v", pkg.Commands[i].Module, err)
						continue
					}
				}
			}
		}

		err = c.Delete()
	}
}
