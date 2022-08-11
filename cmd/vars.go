package cmd

import (
	"fmt"
	"os"

	pkgcmd "github.com/rancherlabs/corral/pkg/cmd"
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/corral"
	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/spf13/cobra"
)

const varsDescription = `
Show the given corral's variables. If not variable is specified all variables are returned.  If one variables
is specified only that variables value is returned.  If multiple variables are specified they will be returned as a table.

Examples:

corral vars k3s
corral vars k3s kube_api_host node_token
corral vars k3s kubeconfig | base64 --decode > ~/.kube/config
corral vars k3s -a
`

func NewCommandVars() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vars NAME [VAR | [VAR...]]",
		Short: "Show the given corral's variables",
		Long:  varsDescription,
		RunE:  listVars,
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.Flags().Bool("sensitive", false, "Sensitive values will be displayed if this flag is true.")
	cmd.Flags().VarP(&output, "output", "o", "Output format. One of: table|json|yaml")
	cmd.Flags().BoolP("all", "a", false, "All values will be displayed if this flag is true.")

	return cmd
}

func listVars(cmd *cobra.Command, args []string) error {
	corralName := args[0]

	c, err := corral.Load(config.CorralPath(corralName))
	if err != nil {
		return err
	}

	// if only one output is requested return the raw value
	if len(args) == 2 {
		_, _ = os.Stdout.WriteString(fmt.Sprintf("%v\n", c.Vars[args[1]]))
		return nil
	}

	pkg, err := _package.LoadPackage(c.Source)
	if err != nil {
		return err
	}

	vs := c.Vars

	if all, _ := cmd.Flags().GetBool("all"); !all {
		vs = pkg.FilterVars(vs)
	}

	if sensitive, _ := cmd.Flags().GetBool("sensitive"); !sensitive {
		vs = pkg.FilterSensitiveVars(vs)
	}

	out, err := pkgcmd.Output(vs, output, pkgcmd.OutputOptions{
		Key:   "NAME",
		Value: "VALUE",
	})
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
