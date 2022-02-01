package config

import (
	"os"
)

func Install() error {
	initialPaths := []string{
		CorralRoot("cache", "layers"),
		CorralRoot("cache", "packages"),
		CorralRoot("cache", "terraform", "bin"),
	}

	for _, p := range initialPaths {
		if err := os.MkdirAll(p, 0o700); err != nil {
			return err
		}
	}

	return nil
}
