package config

import (
	"os"
	"path/filepath"
)

func CorralRoot(parts ...string) string {
	root, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(append([]string{root, ".corral"}, parts...)...)
}

func CorralPath(name string) string {
	return CorralRoot("corrals", name)
}
