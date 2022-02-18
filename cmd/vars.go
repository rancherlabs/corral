package cmd

import (
	"os"

	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/corral"
	_package "github.com/rancherlabs/corral/pkg/package"
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
		Run:   listVars,
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.Flags().Bool("sensitive", false, "Sensitive values will be displayed if this flag is true.")

	return cmd
}

func listVars(cmd *cobra.Command, args []string) {
	corralName := args[0]

	c, err := corral.Load(config.CorralPath(corralName))
	if err != nil {
		logrus.Fatal("failed to load corral: ", err)
	}

	// if only one output is requested return the raw value
	if len(args) == 2 {
		_, _ = os.Stdout.Write([]byte(c.Vars[args[1]]))
		return
	}

	pkg, err := _package.LoadPackage(c.Source)
	if err != nil {
		logrus.Fatal("failed to load corral package: ", err)
	}

	vs := c.Vars

	if !debug {
		vs = pkg.FilterVars(c.Vars)
	}

	if sensitive, _ := cmd.Flags().GetBool("sensitive"); !sensitive {
		vs = pkg.FilterSensitiveVars(vs)
	}

	println("NAME\tVALUE")
	for k, v := range vs {
		println(k + "\t" + v)
	}
	println()
}
