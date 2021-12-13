package config

import (
    "errors"
    "os"
    "path/filepath"

    "github.com/rancherlabs/corral/pkg/version"
    "github.com/sirupsen/logrus"
    "gopkg.in/yaml.v3"
)

type Config struct {
    UserID            string `yaml:"user_id"`
    UserPublicKeyPath string `yaml:"user_public_key_path"`

    AppRootPath string `yaml:"app_root_path"`
    Version     string `yaml:"version"`

    Vars map[string]string `yaml:"vars"`
}

func Load() Config {
    var c Config
    var err error

    c.AppRootPath, err = defaultAppRoot()
    if err != nil {
        logrus.Fatal(err)
    }

    f, err := os.OpenFile(c.AppPath("config.yaml"), os.O_RDONLY, 0o644)
    defer func(f *os.File) { _ = f.Close() }(f)
    if errors.Is(err, os.ErrNotExist) {
        logrus.Fatal("You must call `corral config` before using this command.")
    }
    if err != nil {
        logrus.Fatal("An unknown error occurred loading the configuration", err)
    }

    err = yaml.NewDecoder(f).Decode(&c)
    if err != nil {
        logrus.Fatal("Configuration file is invalid", err)
    }

    if version.Version != c.Version {
        logrus.Warn("Your corral config does not match your binary version. Run `corral config` to upgrade.")
    }

    return c
}

func (c Config) AppPath(path ...string) string {
    return filepath.Join(append([]string{c.AppRootPath}, path...)...)
}

func (c Config) CorralPath(name string) string {
    return c.AppPath("corrals", name)
}

func (c Config) PackageCachePath() string {
    return c.AppPath("packages")
}

func (c Config) RegistryCredentialsFile() string {
    return c.AppPath("registry-creds.json")
}

func (c *Config) Save() (err error) {
    c.AppRootPath, err = defaultAppRoot()
    if err != nil {
        return err
    }

    err = os.MkdirAll(c.AppRootPath, 0700)
    if err != nil {
        return err
    }

    f, err := os.Create(c.AppPath("config.yaml"))
    defer func(f *os.File) { _ = f.Close() }(f)
    if err != nil {
        return err
    }

    return yaml.NewEncoder(f).Encode(c)
}

func defaultAppRoot() (string, error) {
    userHome, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }

    return filepath.Join(userHome, ".corral"), nil
}
