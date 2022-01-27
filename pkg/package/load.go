package _package

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rancherlabs/corral/pkg/config"
	"github.com/sirupsen/logrus"
)

func LoadPackage(ref string) (Package, error) {
	path, _ := filepath.Abs(ref)
	if _, err := os.Stat(path); err == nil {
		return loadLocalPackage(path)

	}

	if !strings.Contains(ref, ":") {
		logrus.Info("Defaulting to latest tag.")
		ref += ":latest"
	}

	return loadRemotePackage(ref)
}

func loadLocalPackage(src string) (pkg Package, err error) {
	if _, err := os.Stat(src); err != nil {
		return pkg, err
	}

	pkg.RootPath = src

	pkg.Manifest, err = LoadManifest(os.DirFS(pkg.RootPath), "manifest.yaml")
	if err != nil {
		return pkg, err
	}

	return pkg, nil
}

func loadRemotePackage(ref string) (pkg Package, err error) {
	registryStore, err := newRegistryStore()
	if err != nil {
		return
	}

	// get the latest digest for the ref
	_, desc, err := registryStore.Resolve(context.Background(), ref)
	if err != nil {
		return
	}

	// check if this digest has already been cached
	dest := filepath.Join(config.CorralRoot("cache", "packages"), getRefPath(ref, string(desc.Digest)))
	_, err = os.Stat(dest)
	if err == nil {
		return loadLocalPackage(dest)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return
	}

	// use a cached fetcher to cache the layers
	storeFetcher, err := registryStore.Fetcher(context.Background(), ref)
	if err != nil {
		return
	}

	fetcher := NewCachedFetcher(storeFetcher)
	if err != nil {
		return
	}

	// fetch the image manifest
	r, err := fetcher.Fetch(context.Background(), desc)
	if err != nil {
		return
	}

	var manifest v1.Manifest
	err = json.NewDecoder(r).Decode(&manifest)
	if err != nil {
		return
	}

	// create the destination directory
	err = os.MkdirAll(dest, 0o700)
	if err != nil {
		return
	}

	// extract the layers to the destination
	for _, layer := range manifest.Layers {
		var layerReader io.ReadCloser
		layerReader, err = fetcher.Fetch(context.Background(), layer)
		if err != nil {
			return
		}

		if err = extractLayer(dest, layerReader); err != nil {
			return
		}
	}

	return loadLocalPackage(dest)
}

func extractLayer(dest string, r io.Reader) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}

		if header.Typeflag == tar.TypeDir {
			err = os.MkdirAll(filepath.Join(dest, header.Name), 0o700)
			if err != nil {
				return err
			}
			continue
		}

		destPath, _ := filepath.Split(header.Name)
		err = os.MkdirAll(filepath.Join(dest, destPath), 0o700)
		if err != nil {
			return err
		}

		f, err := os.Create(filepath.Join(dest, header.Name))
		if err != nil {
			return err
		}

		_, err = io.Copy(f, tr)
		if err != nil {
			_ = f.Close()
			return err
		}

		_ = f.Close()
	}

	return nil
}

func getRefPath(ref, digest string) string {
	ref = strings.Split(ref, ":")[0]
	digest = strings.Split(digest, ":")[1]

	return filepath.Join(append(strings.Split(ref, "/"), digest)...)
}
