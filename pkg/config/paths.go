package config

import (
	"path/filepath"
)

func CorralRoot(parts ...string) string {
	return filepath.Join(append([]string{rootPath, ".corral"}, parts...)...)
}

func CorralPath(name string) string {
	return CorralRoot("corrals", name)
}
