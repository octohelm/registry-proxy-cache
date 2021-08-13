package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/octohelm/registry-proxy-cache/pkg/kubernetes"
	"github.com/octohelm/registry-proxy-cache/pkg/storage"

	dcontext "github.com/distribution/distribution/v3/context"
	"github.com/distribution/distribution/v3/version"
	"github.com/octohelm/registry-proxy-cache/pkg/configuration"
	"github.com/octohelm/registry-proxy-cache/pkg/registryproxycache"
)

func init() {
	flag.Parse()
}

func main() {
	// setup context
	ctx := dcontext.WithVersion(dcontext.Background(), version.Version)

	c, err := resolveConfiguration()
	if err != nil {
		dcontext.GetLogger(ctx).Fatalf("configuration error: %v", err)
		os.Exit(1)
	}

	if args := flag.CommandLine.Args(); len(args) > 0 && args[0] == "gc" {
		gc(ctx, c)
		return
	}

	server(ctx, c)
}

func gc(ctx context.Context, c *configuration.Configuration) {
	log := dcontext.GetLogger(ctx)

	images, err := kubernetes.GetClusterContainerImages(ctx)
	if err != nil {
		log.Fatalf("get cluster container images failed: %v", err)
		os.Exit(1)
	}

	if err := storage.ClusterGarbageCollect(ctx, configuration.ToDistributionConfigurations(c), images, false); err != nil {
		log.Fatalf("garbage collect failed: %v", err)
		os.Exit(1)
	}
}

func server(ctx context.Context, c *configuration.Configuration) {
	log := dcontext.GetLogger(ctx)

	r, err := registryproxycache.NewRegistryProxyCache(ctx, c)
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
