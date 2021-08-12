package registryproxycache

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	_ "unsafe"

	dconfiguration "github.com/distribution/distribution/v3/configuration"
	dcontext "github.com/distribution/distribution/v3/context"
	"github.com/distribution/distribution/v3/health"
	_ "github.com/distribution/distribution/v3/registry"
	"github.com/distribution/distribution/v3/registry/handlers"
	"github.com/distribution/distribution/v3/uuid"
	"github.com/octohelm/registry-proxy-cache/pkg/configuration"
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
					originPath := req.URL.Path

					req.URL.Path = "/v2/" + req.URL.Path[len(p):]
					req.RequestURI = req.URL.RequestURI()

					req.Header.Set("User-Agent", fmt.Sprintf("Proxy-Cache/%s %s", hub, req.Header.Get("User-Agent")))
					registries[hub].ServeHTTP(rw, req)

					req.URL.Path = originPath
					req.RequestURI = req.URL.RequestURI()

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
					m := "/hub-prefix-mirrors/" + hub + "/"

					if strings.HasPrefix(req.URL.Path, m) {
						v2 := "/hub-prefix-mirrors/" + hub + "/v2/"

						if !strings.HasPrefix(req.URL.Path, v2) {
							http.Redirect(rw, req, v2+req.URL.Path[len(m):], http.StatusMovedPermanently)
							return
						}

						originPath := req.URL.Path

						// skip _catalog
						if strings.HasPrefix(req.URL.Path, v2+"_") {
							req.URL.Path = "/v2/" + req.URL.Path[len(v2):]
						} else {
							req.URL.Path = "/v2/" + hub + "/" + req.URL.Path[len(v2):]
						}
						req.RequestURI = req.URL.RequestURI()

						req.Header.Set("User-Agent", fmt.Sprintf("Proxy-Cache/%s %s", hub, req.Header.Get("User-Agent")))
						registries[hub].ServeHTTP(rw, req)

						req.URL.Path = originPath
						req.RequestURI = req.URL.RequestURI()

						return
					}
				}

				rw.WriteHeader(http.StatusNotFound)
				_, _ = rw.Write(nil)

				return
			}

			if strings.HasPrefix(req.URL.Path, "/mirrors/") {
				for hub := range registries {
					m := "/mirrors/" + hub + "/"

					if strings.HasPrefix(req.URL.Path, m) {

						v2 := "/mirrors/" + hub + "/v2/"

						originPath := req.URL.Path
						if !strings.HasPrefix(req.URL.Path, v2) {
							http.Redirect(rw, req, v2+req.URL.Path[len(m):], http.StatusMovedPermanently)
							return
						}

						req.URL.Path = "/v2/" + req.URL.Path[len(v2):]
						req.RequestURI = req.URL.RequestURI()

						req.Header.Set("User-Agent", fmt.Sprintf("Proxy-Cache/%s %s", hub, req.Header.Get("User-Agent")))
						registries[hub].ServeHTTP(rw, req)

						req.URL.Path = originPath
						req.RequestURI = req.URL.RequestURI()

						return
					}
				}

				rw.WriteHeader(http.StatusNotFound)
				_, _ = rw.Write(nil)
				return
			}

			next.ServeHTTP(rw, req)
		})
	}
}

func NewRegistryHandlers(ctx context.Context, config *configuration.Configuration) (RegistryHandlers, error) {
	dconfigs := configuration.ToDistributionConfigurations(config)

	registryHandlers := RegistryHandlers{}

	for hub := range dconfigs {
		r, err := NewRegistryHandler(ctx, dconfigs[hub])
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

//go:linkname  configureBugsnag github.com/distribution/distribution/v3/registry.configureBugsnag
func configureBugsnag(config *dconfiguration.Configuration)

//go:linkname  configureLogging github.com/distribution/distribution/v3/registry.configureLogging
func configureLogging(ctx context.Context, config *dconfiguration.Configuration) (context.Context, error)

//go:linkname  panicHandler github.com/distribution/distribution/v3/registry.panicHandler
func panicHandler(handler http.Handler) http.Handler

//go:linkname  configureReporting github.com/distribution/distribution/v3/registry.configureReporting
func configureReporting(app *handlers.App) http.Handler

//go:linkname  alive github.com/distribution/distribution/v3/registry.alive
func alive(path string, handler http.Handler) http.Handler
