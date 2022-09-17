package magetools

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/sh"
)

type Go struct {
	Arch       string
	OS         string
	Version    string
	Commit     string
	CGoEnabled string
	Verbose    string
}

func NewGo(arch, goos, version, commit string, cgoEnabled, verbose bool) *Go {
	return &Go{
		Arch:       arch,
		OS:         goos,
		Version:    version,
		Commit:     commit,
		CGoEnabled: boolToIntString(cgoEnabled),
		Verbose:    boolToIntString(verbose),
	}
}

func boolToIntString(b bool) string {
	return map[bool]string{true: "1", false: "0"}[b]
}

func (g *Go) Build(target, output string) error {
	envs := map[string]string{"GOOS": g.OS, "GOARCH": g.Arch, "CGO_ENABLED": g.CGoEnabled, "MAGEFILE_VERBOSE": g.Verbose}
	return sh.RunWithV(envs, "go", "build", "-a", "-o", output, "--ldflags="+g.versionFlag(), target)
}

func (g *Go) Test(coverpkg string, targets ...string) error {
	envs := map[string]string{"GOOS": g.OS, "ARCH": g.Arch, "CGO_ENABLED": g.CGoEnabled, "MAGEFILE_VERBOSE": g.Verbose}
	if coverpkg != "" {
		coverpkg = "-coverpkg=" + coverpkg
	}
	return sh.RunWithV(envs, "go", append([]string{"test", "-v", "-cover", coverpkg, "--ldflags=" + g.versionFlag()}, targets...)...)
}

type Mod struct {
	*Go
}

func (g *Go) Mod() *Mod {
	return &Mod{g}
}

func (m *Mod) Download() error {
	envs := map[string]string{"GOOS": m.OS, "ARCH": m.Arch}
	return sh.RunWithV(envs, "go", "mod", "download")
}

func (m *Mod) Tidy() error {
	envs := map[string]string{"GOOS": m.OS, "ARCH": m.Arch}
	return sh.RunWithV(envs, "go", "mod", "tidy")
}

func (m *Mod) Verify() error {
	envs := map[string]string{"GOOS": m.OS, "ARCH": m.Arch}
	return sh.RunWithV(envs, "go", "mod", "verify")
}

func (g *Go) Fmt(target string) error {
	envs := map[string]string{"GOOS": g.OS, "ARCH": g.Arch}
	return sh.RunWithV(envs, "go", "fmt", target)
}

func (g *Go) Lint(targets ...string) error {
	envs := map[string]string{"GOOS": g.OS, "ARCH": g.Arch, "CGO_ENABLED": g.CGoEnabled, "MAGEFILE_VERBOSE": g.Verbose}
	return sh.RunWithV(envs, "golangci-lint", append([]string{"run"}, targets...)...)
}

func (g *Go) versionFlag() string {
	return fmt.Sprintf(`-X 'github.com/rancherlabs/corral/pkg/versionFlag.Version=%s'`, g.Version)
}

func GetCommit() (string, error) {
	result, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result), nil
}

func IsGitClean() error {
	result, err := sh.Output("git", "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return err
	}
	if result != "" {
		return errors.New("encountered dirty repo")
	}
	return nil
}

func GetVersion() (string, error) {
	commit, err := GetCommit()
	if err != nil {
		return "", err
	}
	ref := os.Getenv("GITHUB_REF_NAME")
	if ref != "" {
		ref = strings.TrimPrefix(ref, "v") + "+" + commit // append build metadata
	} else {
		ref = commit
	}
	return ref, nil
}
