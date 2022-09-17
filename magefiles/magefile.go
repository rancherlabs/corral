//go:build mage

package main

import (
	"context"
	"log"
	"os"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/rancherlabs/corral/magetools"
)

var Default = Build
var g *magetools.Go
var version string
var commit string

func Setup() error {
	var err error
	version, err = magetools.GetVersion()
	if err != nil {
		return err
	}
	commit, err = magetools.GetCommit()
	if err != nil {
		return err
	}
	g = magetools.NewGo(goarch(), goos(), version, commit, false, true)
	return nil
}

func Dependencies() error {
	mg.Deps(Setup)
	return g.Mod().Download()
}

func Build(ctx context.Context) error {
	mg.Deps(Dependencies)
	return g.Build()
}

func Validate() error {
	mg.Deps(Setup)
	log.Println("[Validate] Running: golangci-lint")
	if err := g.Lint(); err != nil {
		return err
	}

	log.Println("[Validate] Running: go fmt")
	if err := g.Fmt("./..."); err != nil {
		return err
	}

	log.Println("[Validate] Running: go mod tidy")
	if err := g.Mod().Tidy(); err != nil {
		return err
	}

	log.Println("[Validate] Running: go mod verify")
	if err := g.Mod().Verify(); err != nil {
		return err
	}

	log.Println("[Validate] Checking for dirty repo")
	if err := magetools.IsGitClean(); err != nil {
		return err
	}

	log.Println("[Validate] corral has been successfully validated")
	return nil
}

func Test() error {
	mg.Deps(Setup)
	log.Println("[Test] Running unit tests")
	if err := g.Test("", "./cmd/...", "./pkg/..."); err != nil {
		return err
	}
	log.Println("[Test] Running integration tests")
	if err := g.Test("./...", "./tests/..."); err != nil {
		return err
	}
	log.Println("[Test] corral has been successfully tested")
	return nil
}

func goos() string {
	if goos := os.Getenv("GOOS"); goos != "" {
		return goos
	}
	return runtime.GOOS
}

func goarch() string {
	if goarch := os.Getenv("GOARCH"); goarch != "" {
		return goarch
	}
	return runtime.GOARCH
}
