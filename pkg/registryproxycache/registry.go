package registryproxycache

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	dconfiguration "github.com/docker/distribution/configuration"
	dcontext "github.com/docker/distribution/context"
	"github.com/docker/distribution/health"
	_ "github.com/docker/distribution/registry"
	"github.com/docker/distribution/registry/handlers"
	"github.com/docker/distribution/uuid"
	"github.com/octohelm/registry-proxy-cache/pkg/configuration"
	"gopkg.in/yaml.v2"

	_ "unsafe"
)

type RegistryHandlers map[string]http.Handler

// {registry_host}/v2/{hub}/
// {registry_host}/v2/
func (registries RegistryHandlers) RegistryHandler() http.Handler {
	defaultHandler := registries["docker.io"]

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/v2/") {
			for hub := range registries {
				p := "/v2/" + hub + "/"

				if strings.HasPrefix(req.URL.Path, p) {
					req.URL.Path = "/v2/" + req.URL.Path[len(p):]

					req.RequestURI = req.URL.RequestURI()

					registries[hub].ServeHTTP(rw, req)
					return
				}
			}
		}

		defaultHandler.ServeHTTP(rw, req)
	})
}

// {registry_host}/hub-prefix-mirrors/{hub}/v2/{hub}/
// {registry_host}/mirrors/{hub}/v2/
func (registries RegistryHandlers) RegistryMirrorMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Path, "/hub-prefix-mirrors/") {

				for hub := range registries {
					p := "/hub-prefix-mirrors/" + hub + "/v2/"

					if strings.HasPrefix(req.URL.Path, p) {
						// skip _catalog
						if strings.HasPrefix(req.URL.Path, p+"_") {
							req.URL.Path = "/v2/" + req.URL.Path[len(p):]
							req.RequestURI = req.URL.RequestURI()

							registries[hub].ServeHTTP(rw, req)
							return
						}

						req.URL.Path = "/v2/" + hub + "/" + req.URL.Path[len(p):]
						req.RequestURI = req.URL.RequestURI()

						registries[hub].ServeHTTP(rw, req)
						return
					}
				}

				return
			}

			if strings.HasPrefix(req.URL.Path, "/mirrors/") {
				for hub := range registries {
					p := "/mirrors/" + hub + "/v2/"

					if strings.HasPrefix(req.URL.Path, p) {
						req.URL.Path = "/v2/" + req.URL.Path[len(p):]

						req.RequestURI = req.URL.RequestURI()

						registries[hub].ServeHTTP(rw, req)
						return
					}
				}

				return
			}

			next.ServeHTTP(rw, req)
		})
	}
}

var rerootdirectory = regexp.MustCompile("rootdirectory: ([^\n]+)")

func NewRegistryHandlers(ctx context.Context, config *configuration.Configuration) (RegistryHandlers, error) {
	if len(config.Proxies) == 0 {
		config.Proxies = map[string]configuration.Proxy{}
	}

	if _, ok := config.Proxies["docker.io"]; !ok {
		config.Proxies["docker.io"] = configuration.Proxy{
			RemoteURL: "https://registry-1.docker.io",
		}
	}

	proxies := make(map[string]configuration.Proxy)

	for hub := range config.Proxies {
		proxies[hub] = config.Proxies[hub]
	}

	config.Proxies = nil
	configYAML, _ := yaml.Marshal(config)

	registryHandlers := RegistryHandlers{}

	for hub := range proxies {
		proxy := proxies[hub]

		configYAML2 := rerootdirectory.ReplaceAllFunc(configYAML, func(bytes []byte) []byte {
			return append(rerootdirectory.FindSubmatch(bytes)[0], []byte("/"+hub)...)
		})

		c := &dconfiguration.Configuration{}
		_ = yaml.Unmarshal(configYAML2, c)

		c.Proxy.RemoteURL = proxy.RemoteURL
		c.Proxy.Username = proxy.Username
		c.Proxy.Password = proxy.Password

		if c.Log.Fields == nil {
			c.Log.Fields = map[string]interface{}{}
		}

		c.Log.Fields["hub"] = hub

		// registry must disable debug
		c.HTTP.Debug.Addr = ""
		c.HTTP.Debug.Prometheus.Enabled = false

		r, err := NewRegistryHandler(ctx, c)
		if err != nil {
			return nil, err
		}
		registryHandlers[hub] = r
	}

	return registryHandlers, nil
}

func NewRegistryHandler(ctx context.Context, config *dconfiguration.Configuration) (http.Handler, error) {
	var err error
	ctx, err = configureLogging(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error configuring logger: %v", err)
	}

	configureBugsnag(config)

	// inject a logger into the uuid library. warns us if there is a problem
	// with uuid generation under low entropy.
	uuid.Loggerf = dcontext.GetLogger(ctx).Warnf

	app := handlers.NewApp(ctx, config)
	// TODO(aaronl): The global scope of the health checks means NewRegistryHandler
	// can only be called once per process.
	app.RegisterHealthChecks()
	handler := configureReporting(app)
	handler = alive("/", handler)
	handler = health.Handler(handler)
	handler = panicHandler(handler)

	return handler, nil

}

//go:linkname  configureBugsnag github.com/docker/distribution/registry.configureBugsnag
func configureBugsnag(config *dconfiguration.Configuration)

//go:linkname  configureLogging github.com/docker/distribution/registry.configureLogging
func configureLogging(ctx context.Context, config *dconfiguration.Configuration) (context.Context, error)

//go:linkname  panicHandler github.com/docker/distribution/registry.panicHandler
func panicHandler(handler http.Handler) http.Handler

//go:linkname  configureReporting github.com/docker/distribution/registry.configureReporting
func configureReporting(app *handlers.App) http.Handler

//go:linkname  alive github.com/docker/distribution/registry.alive
func alive(path string, handler http.Handler) http.Handler
