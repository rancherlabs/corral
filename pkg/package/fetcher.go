package _package

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/remotes"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rancherlabs/corral/pkg/config"
)

type CachedFetcher struct {
	cachePath string
	source    remotes.Fetcher
}

func NewCachedFetcher(source remotes.Fetcher) remotes.Fetcher {
	return &CachedFetcher{
		cachePath: config.CorralRoot("cache", "layers"),
		source:    source,
	}
}

func (c *CachedFetcher) Fetch(ctx context.Context, desc v1.Descriptor) (io.ReadCloser, error) {
	isCached, err := c.isCached(desc)
	if err != nil {
		return nil, err
	}

	if !isCached {
		if err := c.cacheFromSource(ctx, desc); err != nil {
			return nil, err
		}
	}

	f, err := os.Open(c.descriptorPath(desc))
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (c *CachedFetcher) isCached(desc v1.Descriptor) (bool, error) {
	info, err := os.Stat(c.descriptorPath(desc))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return info.Size() == desc.Size, nil
}

func (c *CachedFetcher) cacheFromSource(ctx context.Context, desc v1.Descriptor) error {
	// download the image from the source
	r, err := c.source.Fetch(ctx, desc)
	if err != nil {
		return err
	}

	defer func() { _ = r.Close() }()

	// create the cache file
	f, err := os.Create(c.descriptorPath(desc))
	if err != nil {
		return err
	}

	// copy the downloaded layer to the cache
	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	return nil
}

func (c *CachedFetcher) descriptorPath(desc v1.Descriptor) string {
	return filepath.Join(c.cachePath, desc.Digest.String())
}
