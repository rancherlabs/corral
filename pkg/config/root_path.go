package config

import (
	"os"

	"github.com/samber/lo"
)

var rootPath string

func InitializeRootPath(path string) {
	rootPath = path
	lo.Must(os.Stat(rootPath))
}
