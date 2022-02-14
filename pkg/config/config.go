package config

import (
	"context"
	"errors"
	"os"

	tfversion "github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/rancherlabs/corral/pkg/version"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	UserID            string `yaml:"user_id"`
	UserPublicKeyPath string `yaml:"user_public_key_path"`

	Version string `yaml:"version"`

	Vars map[string]string `yaml:"vars"`
}

func MustLoad() Config {
	c, err := Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logrus.Fatal("You must call `corral config` before using this command.")
		}
		logrus.Fatal("Configuration file is invalid", err)
	}
	return c
}

func Load() (Config, error) {
	var c Config
	var err error

	body, err := os.ReadFile(CorralRoot("config.yaml"))
	if err != nil {
		return Config{}, err
	}

	if err := yaml.Unmarshal(body, &c); err != nil {
		return Config{}, err
	}

	if version.Version != c.Version {
		logrus.Infof("Upgrading corral to %s.", version.Version)
		if err := Install(); err != nil {
			panic(err)
		}

		c.Version = version.Version
		if err := c.Save(); err != nil {
			panic(err)
		}

	}

	return c, nil
}

func (c *Config) Save() (err error) {
	f, err := os.Create(CorralRoot("config.yaml"))
	defer func(f *os.File) { _ = f.Close() }(f)
	if err != nil {
		return err
	}

	return yaml.NewEncoder(f).Encode(c)
}

func NewTerraform(modulePath, version string) (*tfexec.Terraform, error) {
	v, err := tfversion.NewVersion(version)
	if err != nil {
		return nil, err
	}

	versionPath := CorralRoot("cache", "terraform", "bin", version)

	if err = os.MkdirAll(versionPath, 0o700); err != nil {
		return nil, err
	}

	i := install.NewInstaller()
	tfPath, err := i.Ensure(context.Background(), []src.Source{
		&fs.ExactVersion{
			Product:    product.Terraform,
			ExtraPaths: []string{versionPath},
			Version:    v,
		},
		&releases.ExactVersion{
			Product:    product.Terraform,
			InstallDir: versionPath,
			Version:    v,
		},
	})
	if err != nil {
		return nil, err
	}

	return tfexec.NewTerraform(modulePath, tfPath)
}
