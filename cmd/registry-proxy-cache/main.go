package main

import (
	"fmt"
	"os"

	dcontext "github.com/docker/distribution/context"
	"github.com/docker/distribution/version"
	"github.com/octohelm/registry-proxy-cache/pkg/configuration"
	"github.com/octohelm/registry-proxy-cache/pkg/registryproxycache"
)

func main() {
	// setup context
	ctx := dcontext.WithVersion(dcontext.Background(), version.Version)

	log := dcontext.GetLogger(ctx)

	config, err := resolveConfiguration()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
		os.Exit(1)
	}

	r, err := registryproxycache.NewRegistryProxyCache(ctx, config)
	if err != nil {
		log.Fatalln(err)
		return
	}

	if err = r.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}

func resolveConfiguration() (*configuration.Configuration, error) {
	var configurationPath string

	if p := os.Getenv("REGISTRY_CONFIGURATION_PATH"); p != "" {
		configurationPath = p
	}

	if configurationPath == "" {
		return nil, fmt.Errorf("configuration path unspecified")
	}

	fp, err := os.Open(configurationPath)
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	config, err := configuration.Parse(fp)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", configurationPath, err)
	}

	if v := os.Getenv("REGISTRY_PROXIES"); v != "" {
		registryProxies, err := configuration.ParseProxies(v)
		if err != nil {
			return nil, err
		}
		config.Proxies = registryProxies
	}

	return config, nil
}
