package _package

import (
    "archive/tar"
    "bytes"
    "compress/gzip"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "io/fs"
    "net/http"
    "os"
    "path/filepath"
    "strings"

    v1 "github.com/opencontainers/image-spec/specs-go/v1"
    "github.com/rancherlabs/corral/pkg/version"
    "github.com/sirupsen/logrus"
    "gopkg.in/yaml.v2"
    "oras.land/oras-go/pkg/auth"
    dockerauth "oras.land/oras-go/pkg/auth/docker"
    "oras.land/oras-go/pkg/content"
    "oras.land/oras-go/pkg/oras"
)

const corralUserAgent = "Corral/" + version.Version

type Package struct {
    RootPath string

    Manifest
}

func (b Package) ManifestPath() string {
    return filepath.Join(b.RootPath, "manifest.yaml")
}

func (b Package) TerraformModulePath() string {
    return filepath.Join(b.RootPath, "terraform", "module")
}

func (b Package) TerraformPluginPath() string {
    return filepath.Join(b.RootPath, "terraform", "plugins")
}

func (b *Package) ScriptPath() string {
    return filepath.Join(b.RootPath, "scripts")
}

func LoadPackage(ref, cachePath, registryCredentialsFile string) (Package, error) {
    path, _ := filepath.Abs(ref)
    if _, err := os.Stat(path); err == nil {
        return loadLocalPackage(path)

    }

    return loadRemotePackage(ref, cachePath, registryCredentialsFile)
}

func loadLocalPackage(src string) (Package, error) {
    var pkg Package

    if _, err := os.Stat(src); err != nil {
        return pkg, err
    }

    pkg.RootPath = src

    mf, err := os.Open(pkg.ManifestPath())
    if err != nil {
        return pkg, fmt.Errorf("%s could not find package manifest", src)
    }

    err = yaml.NewDecoder(mf).Decode(&pkg.Manifest)
    if err != nil {
        return pkg, fmt.Errorf("failed to parse package manifest: %s", err.Error())
    }

    return pkg, nil
}

func loadRemotePackage(ref, cachePath, registryCredentialsFile string) (pkg Package, err error) {
    registryStore, err := newRegistryStore(registryCredentialsFile)
    if err != nil {
        return
    }

    name, desc, err := registryStore.Resolve(context.Background(), ref)
    if err != nil {
        return
    }

    dest := filepath.Join(cachePath, getRefPath(ref, string(desc.Digest)))
    _, err = os.Stat(dest)
    if err == nil {
        return loadLocalPackage(dest)
    }
    if !errors.Is(err, os.ErrNotExist) {
        return
    }

    fetcher, err := registryStore.Fetcher(context.Background(), name)
    if err != nil {
        return
    }

    r, err := fetcher.Fetch(context.Background(), desc)
    if err != nil {
        return
    }

    var manifest v1.Manifest
    err = json.NewDecoder(r).Decode(&manifest)
    if err != nil {
        return
    }

    rr, err := fetcher.Fetch(context.Background(), manifest.Layers[0])
    if err != nil {
        return
    }

    err = os.MkdirAll(dest, 0o700)
    if err != nil {
        return
    }

    err = extractPackage(dest, rr)
    if err != nil {
        return
    }

    return loadLocalPackage(dest)
}

func UploadPackage(root, ref, credentialsFile string) error {
    manifest, err := os.ReadFile(filepath.Join(root, "manifest.yaml"))
    if err != nil {
        return err
    }

    logrus.Info("compressing files")
    buf, err := compressPackage(root)
    if err != nil {
        return err
    }

    logrus.Info("building manifest")
    memoryStore := content.NewMemory()
    configDescriptor, _ := memoryStore.Add("", v1.MediaTypeImageLayer, manifest)
    contentDescriptor, _ := memoryStore.Add("", v1.MediaTypeImageLayerGzip, buf)
    pushContents := []v1.Descriptor{contentDescriptor}

    manifestData, manifestDescriptor, err := content.GenerateManifest(&configDescriptor, nil, pushContents...)
    _ = memoryStore.StoreManifest(ref, manifestDescriptor, manifestData)

    registryStore, err := newRegistryStore(credentialsFile)
    if err != nil {
        return err
    }

    logrus.Info("pushing to registry")
    _, err = oras.Copy(context.Background(), memoryStore, ref, registryStore, "")
    return err
}

func newRegistryStore(credentialFiles ...string) (reg content.Registry, err error) {
    authorizer, err := dockerauth.NewClient(credentialFiles...)
    if err != nil {
        return
    }

    headers := http.Header{}
    headers.Set("User-Agent", corralUserAgent)
    opts := []auth.ResolverOption{auth.WithResolverHeaders(headers)}
    resolver, err := authorizer.ResolverWithOpts(opts...)
    if err != nil {
        return
    }

    reg = content.Registry{Resolver: resolver}

    return
}

func extractPackage(dest string, r io.Reader) error {
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

func compressPackage(root string) ([]byte, error) {
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
                Name:     path,
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
            Name: path,
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

func getRefPath(ref, digest string) string {
    ref = strings.Split(ref, ":")[0]
    digest = strings.Split(digest, ":")[1]

    return filepath.Join(append(strings.Split(ref, "/"), digest)...)
}
