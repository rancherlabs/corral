package cmd

import (
    "context"
    "errors"
    "os"

    "github.com/hashicorp/terraform-exec/tfexec"
    "github.com/rancherlabs/corral/pkg/config"
    "github.com/rancherlabs/corral/pkg/corral"
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
        PreRun: func(_ *cobra.Command, _ []string) {
            cfg = config.Load()
        },
    }

    cmd.Flags().Bool("dry-run", false, "Display what resources this corral will create.")
    _ = cfgViper.BindPFlag("dry-run", cmd.Flags().Lookup("dry-run"))

    return cmd
}

func deleteCorral(_ *cobra.Command, args []string) {
    for _, name := range args {
        c, err := corral.Load(cfg.CorralPath(name))
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
        tf, _ := tfexec.NewTerraform(c.TerraformPath(), cfg.AppPath("bin", "terraform"))
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
