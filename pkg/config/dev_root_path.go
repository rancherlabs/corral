//go:build dev

package config

import (
	"os"
	"path/filepath"
)

var rootPath string

func init() {
	var err error
	rootPath, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	rootPath = filepath.Join(rootPath, ".dev/corral")
}
