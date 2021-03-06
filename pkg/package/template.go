package _package

import (
	"errors"
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
	Name            string                 `yaml:"name"`
	Annotations     map[string]string      `yaml:"annotations,omitempty"`
	Description     string                 `yaml:"description"`
	Commands        []Command              `yaml:"commands"`
	Overlay         map[string]string      `yaml:"overlay,omitempty"`
	VariableSchemas map[string]interface{} `yaml:"variables,omitempty"`
}

func MergePackages(name string, pkgs []Package) (TemplateManifest, error) {
	t := TemplateManifest{
		Name:            name,
		Overlay:         map[string]string{},
		VariableSchemas: map[string]interface{}{},
	}
	var srcs = map[string]string{}
	for _, pkg := range pkgs {
		buf, err := ioutil.ReadFile(filepath.Join(pkg.RootPath, "manifest.yaml"))
		if err != nil {
			return t, err
		}

		yml := struct {
			VariableSchemas map[string]interface{} `yaml:"variables,omitempty"`
		}{
			VariableSchemas: map[string]interface{}{},
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
			srcs[k] = pkg.Name
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
			f.Close()
			in.Close()

			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func mergeVariable(vars ...interface{}) interface{} {
	out := map[string]interface{}{}

	for _, v := range vars {
		vm := v.(map[string]interface{})

		for k, v := range vm {
			out[k] = v
		}
	}

	return out

}
