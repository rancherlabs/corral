package _package

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

func UploadPackage(pkg Package, ref string) error {
	logrus.Info("building manifest")
	memoryStore := content.NewMemory()
	configDescriptor, err := getManifestDescriptor(memoryStore, pkg)

	logrus.Info("compressing package contents")
	var contents []v1.Descriptor

	if desc, err := addManifestLayer(memoryStore, pkg); err == nil {
		contents = append(contents, desc)
	} else {
		return err
	}

	if ds, err := addTerraformModuleLayers(memoryStore, pkg); err == nil {
		contents = append(contents, ds...)
	} else {
		return err
	}

	if desc, err := addOverlayLayer(memoryStore, pkg); err == nil {
		contents = append(contents, desc)
	} else {
		return err
	}

	manifestData, manifestDescriptor, err := content.GenerateManifest(&configDescriptor, pkg.Annotations, contents...)
	_ = memoryStore.StoreManifest(ref, manifestDescriptor, manifestData)

	registryStore, err := newRegistryStore()
	if err != nil {
		return err
	}

	logrus.Info("pushing to registry")
	_, err = oras.Copy(context.Background(), memoryStore, ref, registryStore, "")
	return err
}

func getManifestDescriptor(memoryStore *content.Memory, pkg Package) (v1.Descriptor, error) {
	buf, err := os.ReadFile(pkg.ManifestPath())
	if err != nil {
		return v1.Descriptor{}, err
	}

	return memoryStore.Add("", v1.MediaTypeImageLayer, buf)
}

func addManifestLayer(memoryStore *content.Memory, pkg Package) (v1.Descriptor, error) {
	manifest, err := os.ReadFile(pkg.ManifestPath())
	if err != nil {
		return v1.Descriptor{}, err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	hdr := &tar.Header{
		Name: "manifest.yaml",
		Size: int64(len(manifest)),
	}

	if err = tw.WriteHeader(hdr); err != nil {
		return v1.Descriptor{}, err
	}

	if _, err = tw.Write(manifest); err != nil {
		return v1.Descriptor{}, err
	}

	err = tw.Close()
	if err != nil {
		return v1.Descriptor{}, err
	}

	err = gz.Close()
	if err != nil {
		return v1.Descriptor{}, err
	}

	return memoryStore.Add("", v1.MediaTypeImageLayerGzip, buf.Bytes())
}

func addTerraformModuleLayers(memoryStore *content.Memory, pkg Package) ([]v1.Descriptor, error) {
	var ds []v1.Descriptor

	var desc v1.Descriptor
	for _, cmd := range pkg.Commands {
		if cmd.Module != "" {
			buf, err := compressPath(filepath.Join("terraform", cmd.Module), pkg.TerraformModulePath(cmd.Module))
			if err != nil {
				return nil, err
			}

			desc, err = memoryStore.Add("", v1.MediaTypeImageLayerGzip, buf)
			if err != nil {
				return nil, err
			}

			ds = append(ds, desc)
		}
	}

	return ds, nil
}

func addOverlayLayer(memoryStore *content.Memory, pkg Package) (v1.Descriptor, error) {
	buf, err := compressPath("overlay", pkg.OverlayPath())
	if err != nil {
		return v1.Descriptor{}, err
	}

	return memoryStore.Add("", v1.MediaTypeImageLayerGzip, buf)
}

func compressPath(prefix, root string) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	sfs := os.DirFS(root)
	err := fs.WalkDir(sfs, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}

		if d.IsDir() {
			hdr := &tar.Header{
				Name:     filepath.Join(prefix, path),
				Typeflag: tar.TypeDir,
			}
			if err = tw.WriteHeader(hdr); err != nil {
				return err
			}
			return nil
		}

		stat, err := os.Stat(filepath.Join(root, path))
		if err != nil {
			return err
		}

		f, err := os.Open(filepath.Join(root, path))
		if err != nil {
			return err
		}
		defer func(f *os.File) { _ = f.Close() }(f)

		hdr := &tar.Header{
			Name: filepath.Join(prefix, path),
			Size: stat.Size(),
		}

		if err = tw.WriteHeader(hdr); err != nil {
			return err
		}

		_, err = io.Copy(tw, f)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = tw.Close()
	if err != nil {
		return nil, err
	}

	err = gz.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
