package cmd_package

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	_package "github.com/rancherlabs/corral/pkg/package"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const downloadDescription = `
Download a package from an OCI registry to the local filesystem.

Examples:
corral package download ghcr.io/rancher/my_pkg:latest
corral package download ghcr.io/rancher/my_pkg:latest dest
`

func NewCommandDownload() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download PACKAGE [DEST]",
		Short: "Download a package from an OCI registry.",
		Long:  downloadDescription,
		Run:   download,
		Args:  cobra.RangeArgs(1, 2),
	}

	return cmd
}

func download(_ *cobra.Command, args []string) {
	pkg, err := _package.LoadPackage(args[0])
	if err != nil {
		logrus.Fatalf("failed to load package: %s", err)
	}

	dest := pkg.Name
	if len(args) > 1 {
		dest = args[1]
	}

	err = filepath.WalkDir(pkg.RootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := dest + path[len(pkg.RootPath):]

		if d.IsDir() {
			if err := os.Mkdir(destPath, 0700); err != nil {
				return err
			}
		} else {
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

	if err != nil {
		logrus.Fatalf("failed to copy package files to destination: %s", err)
	}
}
