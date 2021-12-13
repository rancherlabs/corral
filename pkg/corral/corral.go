package corral

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Corral struct {
	RootPath string `yaml:"rootPath"`

	Name       string `yaml:"name"`
	Status     Status `yaml:"status" json:"status,omitempty"`
	PublicKey  string `yaml:"public_key"`
	PrivateKey string `yaml:"private_key"`

	NodePools map[string][]Node `yaml:"node_pools" json:"node_pools,omitempty"`
	Vars      map[string]string `yaml:"vars" json:"vars,omitempty"`
}

func Load(path string) (*Corral, error) {
	f, err := os.Open(filepath.Join(path, "corral.yaml"))
	defer func(f *os.File) { _ = f.Close() }(f)
	if err != nil {
		return nil, err
	}

	var c Corral
	return &c, yaml.NewDecoder(f).Decode(&c)
}

func (c *Corral) TerraformPath() string {
	return filepath.Join(c.RootPath, "terraform")
}

func (c *Corral) Exists() bool {
	_, err := os.Stat(c.RootPath)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func (c *Corral) Save() error {
	err := os.MkdirAll(c.RootPath, 0700)
	if err != nil {
		return err
	}

	err = os.MkdirAll(c.TerraformPath(), 0700)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(c.RootPath, "corral.yaml"))
	defer func(f *os.File) { _ = f.Close() }(f)
	if err != nil {
		return err
	}

	return yaml.NewEncoder(f).Encode(c)
}

func (c *Corral) Delete() error {
	err := os.RemoveAll(c.RootPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *Corral) SetStatus(status Status) {
	c.Status = status
	err := c.Save()
	if err != nil {
		logrus.Warn("Failed to save corral status")
	}
}
