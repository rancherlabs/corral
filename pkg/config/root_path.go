//go:build !dev

package config

import "os"

var rootPath string

func init() {
	var err error
	rootPath, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}
}
