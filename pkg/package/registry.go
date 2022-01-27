package _package

import (
	"net/http"

	"github.com/rancherlabs/corral/pkg/config"
	"oras.land/oras-go/pkg/auth"
	"oras.land/oras-go/pkg/auth/docker"
	dockerauth "oras.land/oras-go/pkg/auth/docker"
	"oras.land/oras-go/pkg/content"
)

var registryCredentials = config.CorralRoot("registry-creds.json")

func newRegistryStore() (reg content.Registry, err error) {
	authorizer, err := docker.NewClient(registryCredentials)
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

func AddRegistryCredentials(registry, username, password string) error {
	da, err := dockerauth.NewClient(registryCredentials)
	if err != nil {
		return err
	}

	return da.LoginWithOpts(auth.WithLoginHostname(registry),
		auth.WithLoginUsername(username),
		auth.WithLoginSecret(password),
		auth.WithLoginUserAgent(corralUserAgent))
}
