package corral

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/rancherlabs/corral/pkg/version"

	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/rancherlabs/corral/pkg/vars"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const nodePoolVarName = "corral_node_pools"

type Corral struct {
	RootPath string `yaml:"rootPath"`
	Source   string `yaml:"source"`

	Name       string `yaml:"name"`
	Status     Status `yaml:"status" json:"status,omitempty"`
	PublicKey  string `yaml:"public_key"`
	PrivateKey string `yaml:"private_key"`

	NodePools map[string][]Node `yaml:"node_pools" json:"node_pools,omitempty"`
	Vars      vars.VarSet       `yaml:"vars" json:"vars,omitempty"`
}

func Load(path string) (*Corral, error) {
	var c Corral
	b, err := ioutil.ReadFile(filepath.Join(path, "corral.yaml"))
	if err != nil {
		return nil, err
	}
	return &c, yaml.Unmarshal(b, &c)
}

func (c *Corral) TerraformPath(name string) string {
	return filepath.Join(c.RootPath, "terraform", name)
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

	err = os.MkdirAll(c.TerraformPath(""), 0700)
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

func (c *Corral) ApplyModule(src, name string) error {
	if err := os.MkdirAll(c.TerraformPath(name), 0700); err != nil {
		return err
	}

	tf, err := config.NewTerraform(c.TerraformPath(name), version.TerraformVersion)
	if err != nil {
		return errors.Wrap(err, "failed to initialized terraform")
	}

	err = tf.Init(context.Background(),
		tfexec.Upgrade(false),
		tfexec.FromModule(src))
	if err != nil {
		return errors.Wrap(err, "failed to initialized terraform module")
	}

	f, err := os.Create(filepath.Join(c.TerraformPath(name), "terraform.tfvars.json"))
	if err != nil {
		return errors.Wrap(err, "failed to create tfvars file")
	}
	tfVars := map[string]interface{}{}
	for k, v := range c.Vars {
		if k == "corral_node_pools" {
			tfVars[k] = c.NodePools
		}

		tfVars[k] = v
	}

	_ = json.NewEncoder(f).Encode(tfVars)
	_ = f.Close()

	err = tf.Apply(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to apply terraform module")
	}

	tfOutput, err := tf.Output(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed read terraform output")
	}

	for k, v := range tfOutput {
		if k == nodePoolVarName {
			np := map[string][]Node{}
			if err = json.Unmarshal(v.Value, &np); err != nil {
				return errors.Wrap(err, "failed to parse node pools from output")
			}

			for s, nodes := range np {
				c.NodePools[s] = append(c.NodePools[s], nodes...)
			}

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(c.NodePools)
			c.Vars[nodePoolVarName] = vars.Escape(&buf)
		}

		c.Vars[k] = vars.FromTerraformOutputMeta(v)
	}

	return nil
}

func (c *Corral) DestroyModule(name string) error {
	if _, err := os.Stat(c.TerraformPath(name)); os.IsNotExist(err) {
		return nil
	}

	tf, err := config.NewTerraform(c.TerraformPath(name), version.TerraformVersion)
	if err != nil {
		return errors.Wrap(err, "failed to initialized terraform")
	}

	err = tf.Destroy(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to destroy terraform module")
	}

	return nil
}
