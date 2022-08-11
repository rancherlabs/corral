package main

import (
	"os"

	"github.com/rancherlabs/corral/cmd"
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/samber/lo"
)

func main() {
	root := os.Getenv("CORRAL_ROOT")
	if root == "" {
		root = lo.Must(os.UserHomeDir())
	}
	config.InitializeRootPath(root)
	cmd.Execute()
}
