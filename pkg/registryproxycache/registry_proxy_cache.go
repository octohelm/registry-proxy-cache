package registryproxycache

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	dcontext "github.com/docker/distribution/context"
	"github.com/docker/go-metrics"

	"github.com/docker/distribution/registry/listener"
	"github.com/octohelm/registry-proxy-cache/pkg/configuration"

	gorhandlers "github.com/gorilla/handlers"

	_ "net/http/pprof"

	_ "github.com/docker/distribution/registry/auth/htpasswd"
	_ "github.com/docker/distribution/registry/auth/silly"
	_ "github.com/docker/distribution/registry/auth/token"
	_ "github.com/docker/distribution/registry/proxy"
	_ "github.com/docker/distribution/registry/storage/driver/azure"
	_ "github.com/docker/distribution/registry/storage/driver/filesystem"
	_ "github.com/docker/distribution/registry/storage/driver/gcs"
	_ "github.com/docker/distribution/registry/storage/driver/inmemory"
	_ "github.com/docker/distribution/registry/storage/driver/middleware/alicdn"
	_ "github.com/docker/distribution/registry/storage/driver/middleware/cloudfront"
	_ "github.com/docker/distribution/registry/storage/driver/middleware/redirect"
	_ "github.com/docker/distribution/registry/storage/driver/oss"
	_ "github.com/docker/distribution/registry/storage/driver/s3-aws"
	_ "github.com/docker/distribution/registry/storage/driver/swift"
)

var quit = make(chan os.Signal, 1)

type RegistryProxyCache struct {
	config *configuration.Configuration
	Server *http.Server
	ctx    context.Context
}

func NewRegistryProxyCache(ctx context.Context, config *configuration.Configuration) (*RegistryProxyCache, error) {
	registries, err := NewRegistryHandlers(ctx, config)
	if err != nil {
		return nil, err
	}

	server := &http.Server{}

	server.Handler = registries.RegistryMirrorMiddleware()(registries.RegistryHandler())

	if !config.Log.AccessLog.Disabled {
		server.Handler = gorhandlers.CombinedLoggingHandler(os.Stdout, server.Handler)
	}

	return &RegistryProxyCache{
		ctx:    ctx,
		config: config,
		Server: server,
	}, nil
}

// ListenAndServe runs the registry's HTTP Server.
func (r *RegistryProxyCache) ListenAndServe() error {
	config := r.config
	log := dcontext.GetLogger(r.ctx)

	if config.HTTP.Debug.Addr != "" {
		go func(addr string) {
			log.Infof("debug server listening %v (%s,%s)", addr)

			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Fatalf("error listening on debug interface: %v", err)
			}
		}(config.HTTP.Debug.Addr)
	}

	if config.HTTP.Debug.Prometheus.Enabled {
		path := config.HTTP.Debug.Prometheus.Path
		if path == "" {
			path = "/metrics"
		}
		http.Handle(path, metrics.Handler())
	}

	ln, err := listener.NewListener(config.HTTP.Net, config.HTTP.Addr)
	if err != nil {
		return err
	}

	if config.HTTP.DrainTimeout == 0 {
		return r.Server.Serve(ln)
	}

	// setup channel to get notified on SIGTERM signal
	signal.Notify(quit, syscall.SIGTERM)
	serveErr := make(chan error)

	// Start serving in goroutine and listen for stop signal in main thread
	go func() {
		log.Infof("server listening %v (%s,%s)", config.HTTP.Addr, runtime.GOOS, runtime.GOARCH)
		serveErr <- r.Server.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		return err
	case <-quit:
		// shutdown the Server with a grace period of configured timeout
		c, cancel := context.WithTimeout(context.Background(), config.HTTP.DrainTimeout)
		defer cancel()
		return r.Server.Shutdown(c)
	}
}
