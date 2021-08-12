package registryproxycache

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	dcontext "github.com/distribution/distribution/v3/context"
	"github.com/distribution/distribution/v3/registry/listener"

	"github.com/docker/go-metrics"
	gorhandlers "github.com/gorilla/handlers"
	"github.com/octohelm/registry-proxy-cache/pkg/configuration"
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
			log.Infof("debug server listening %v (%s,%s)", addr, runtime.GOOS, runtime.GOARCH)

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
