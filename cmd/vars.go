package cmd

import (
    "os"

    "github.com/jedib0t/go-pretty/v6/table"
    "github.com/rancherlabs/corral/pkg/config"
    "github.com/rancherlabs/corral/pkg/corral"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

const varsDescription = `
Show the given corral's variables. If not variable is specified all variables are returned.  If one variables
is specified only that variables value is returned.  If multiple variables are specified they will be returned as a table.

Examples:

corral vars k3s
corral vars k3s kube_api_host node_token
corral vars k3s kubeconfig | base64 --decode > ~/.kube/config
`

func NewCommandVars() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "vars NAME [VAR | [VAR...]]",
        Short: "Show the given corral's variables.",
        Long:  varsDescription,
        Run:   vars,
        Args:  cobra.MinimumNArgs(1),
        PreRun: func(_ *cobra.Command, _ []string) {
            cfg = config.Load()
        },
    }

    return cmd
}

func vars(_ *cobra.Command, args []string) {
    corralName := args[0]

    c, err := corral.Load(cfg.CorralPath(corralName))
    if err != nil {
        logrus.Fatal("failed to load corral: ", err)
    }

    // if only one output is requested return the raw value
    if len(args) == 2 {
        _, _ = os.Stdout.Write([]byte(c.Vars[args[1]]))
        return
    }

    tbl := table.NewWriter()
    tbl.SetOutputMirror(os.Stdout)
    tbl.AppendHeader(table.Row{"NAME", "VALUE"})
    tbl.AppendSeparator()

    var found bool
    for k, v := range c.Vars {
        found = false
        if len(args) > 1 {
            for _, kk := range args[1:] {
                if kk == k {
                    found = true
                    break
                }
            }
        } else {
            found = true
        }

        if !found {
            continue
        }

        if len(v) > 16 {
            v = v[0:16] + "..."
        }

        tbl.AppendRow(table.Row{k, v})
    }

    tbl.Render()
}
