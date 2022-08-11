package _package

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type TemplateManifest struct {
	Name            string            `yaml:"name"`
	Annotations     map[string]string `yaml:"annotations,omitempty"`
	Description     string            `yaml:"description"`
	Commands        []Command         `yaml:"commands"`
	Overlay         map[string]string `yaml:"overlay,omitempty"`
	VariableSchemas map[string]any    `yaml:"variables,omitempty"`
}

func Template(name, description string, packages ...string) error {
	pkgs := make([]Package, len(packages))

	for i, p := range packages {
		pkg, err := LoadPackage(p) // ensures pkg is in cache
		if err != nil {
			return fmt.Errorf("failed to load [%s] package: %w", p, err)
		}
		pkgs[i] = pkg
	}

	manifest, err := MergePackages(name, description, pkgs)
	if err != nil {
		return err
	}

	buf, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(name, "manifest.yaml"), buf, 0664)
	if err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	err = ValidateManifest(buf)
	if err != nil {
		return fmt.Errorf("rendered package is not a valid package: %w", err)
	}
	return nil
}

func MergePackages(name, description string, pkgs []Package) (TemplateManifest, error) {
	if description == "" {
		for i := range pkgs {
			if i > 0 {
				description += "\n"
			}

			if pkgs[i].Description != "" {
				description += pkgs[i].Description
			}
		}
	}

	t := TemplateManifest{
		Name:            filepath.Base(name),
		Description:     description,
		Overlay:         map[string]string{},
		VariableSchemas: map[string]any{},
	}

	for _, pkg := range pkgs {
		buf, err := ioutil.ReadFile(filepath.Join(pkg.RootPath, "manifest.yaml"))
		if err != nil {
			return t, err
		}

		yml := struct {
			VariableSchemas map[string]any `yaml:"variables,omitempty"`
		}{
			VariableSchemas: map[string]any{},
		}

		err = yaml.Unmarshal(buf, &yml)
		if err != nil {
			return t, err
		}

		for _, c := range pkg.Commands {
			if c.Module != "" {
				c.Module = filepath.Join(pkg.Name, c.Module)
			}
			t.Commands = append(t.Commands, c)
		}
		for k, v := range pkg.Overlay {
			t.Overlay[k] = v
		}
		for k, v := range yml.VariableSchemas {
			if _, ok := yml.VariableSchemas[k]; ok {
				t.VariableSchemas[k] = mergeVariable(yml.VariableSchemas[k], v)
			} else {
				t.VariableSchemas[k] = v
			}
		}

		logrus.Infof("Copying modules from %s", pkg.Name)

		err = copyTerraform(name, pkg)
		if err != nil {
			return t, err
		}

		logrus.Infof("Copying overlay from %s", pkg.Name)

		err = copyOverlay(name, pkg)
		if err != nil {
			return t, err
		}
	}
	return t, nil
}

func copyTerraform(root string, pkg Package) error {
	return copyFiles(filepath.Join(root, "terraform", pkg.Name), pkg.TerraformModulePath(""), pkg)
}

func copyOverlay(root string, pkg Package) error {
	return copyFiles(filepath.Join(root, "overlay"), pkg.OverlayPath(), pkg)
}

func copyFiles(root, dir string, pkg Package) error {
	_, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		orig := path[len(dir):]
		destPath := root + orig

		if d.IsDir() {
			logrus.Debugf("creating %s", destPath)
			if err = os.MkdirAll(destPath, 0700); err != nil {
				return err
			}
		} else {
			// skip manifests
			if strings.HasSuffix(path, "manifest.yaml") {
				return nil
			}
			logrus.Debugf("%s: %s -> %s", pkg.Name, orig, destPath)

			if _, err = os.Stat(destPath); err == nil {
				_ = os.Remove(destPath)
			}

			f, err := os.Create(destPath)
			if err != nil {
				return err
			}

			in, err := os.Open(path)
			if err != nil {
				return err
			}

			_, err = io.Copy(f, in)
			_ = f.Close()
			_ = in.Close()

			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func mergeVariable(vars ...any) any {
	out := map[string]any{}

	for _, v := range vars {
		vm := v.(map[string]any)

		for k, v := range vm {
			out[k] = v
		}
	}

	return out

}
