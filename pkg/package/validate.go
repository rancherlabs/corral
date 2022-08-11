package _package

import (
	"errors"
	"os"

	"github.com/sirupsen/logrus"
)

var ErrOverlayNotFound = errors.New("overlay folder not found")

func Validate(name string) error {
	pkg, err := LoadPackage(name)
	if err != nil {
		return err
	}

	if i, err := os.Stat(pkg.OverlayPath()); err != nil || !i.IsDir() {
		return ErrOverlayNotFound
	}

	for _, cmd := range pkg.Commands {
		if cmd.Module != "" {
			if i, err := os.Stat(pkg.TerraformModulePath(cmd.Module)); err != nil || !i.IsDir() {
				return err
			}
		}
	}

	err = pkg.ValidateDefaults()
	if err != nil {
		return err
	}

	logrus.Info("package is valid")
	return nil
}
