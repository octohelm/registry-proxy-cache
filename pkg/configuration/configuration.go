package configuration

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/distribution/distribution/v3/configuration"

	_ "github.com/distribution/distribution/v3/registry/auth/htpasswd"
	_ "github.com/distribution/distribution/v3/registry/auth/silly"
	_ "github.com/distribution/distribution/v3/registry/auth/token"

	_ "github.com/distribution/distribution/v3/registry/proxy"

	_ "github.com/distribution/distribution/v3/registry/storage/driver/filesystem"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/middleware/alicdn"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/oss"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/s3-aws"
)

type Configuration struct {
	Log struct {
		// AccessLog configures access logging.
		AccessLog struct {
			// Disabled disables access logging.
			Disabled bool `yaml:"disabled,omitempty"`
		} `yaml:"accesslog,omitempty"`

		// Level is the granularity at which registry operations are logged.
		Level configuration.Loglevel `yaml:"level,omitempty"`

		// Formatter overrides the default formatter with another. Options
		// include "text", "json" and "logstash".
		Formatter string `yaml:"formatter,omitempty"`

		// Fields allows users to specify static string fields to include in
		// the logger context.
		Fields map[string]interface{} `yaml:"fields,omitempty"`

		// Hooks allows users to configure the log hooks, to enabling the
		// sequent handling behavior, when defined levels of log message emit.
		Hooks []configuration.LogHook `yaml:"hooks,omitempty"`
	}

	// Storage is the configuration for the registry's storage driver
	Storage configuration.Storage `yaml:"storage"`

	// Auth allows configuration of various authorization methods that may be
	// used to gate requests.
	Auth configuration.Auth `yaml:"auth,omitempty"`

	// Middleware lists all middlewares to be used by the registry.
	Middleware map[string][]configuration.Middleware `yaml:"middleware,omitempty"`

	// Reporting is the configuration for error reporting
	Reporting configuration.Reporting `yaml:"reporting,omitempty"`

	// HTTP contains configuration parameters for the registry's http
	// interface.
	HTTP struct {
		// Addr specifies the bind address for the registry instance.
		Addr string `yaml:"addr,omitempty"`

		// Net specifies the net portion of the bind address. A default empty value means tcp.
		Net string `yaml:"net,omitempty"`

		// Host specifies an externally-reachable address for the registry, as a fully
		// qualified URL.
		Host string `yaml:"host,omitempty"`

		Prefix string `yaml:"prefix,omitempty"`

		// Secret specifies the secret key which HMAC tokens are created with.
		Secret string `yaml:"secret,omitempty"`

		// RelativeURLs specifies that relative URLs should be returned in
		// Location headers
		RelativeURLs bool `yaml:"relativeurls,omitempty"`

		// Amount of time to wait for connection to drain before shutting down when registry
		// receives a stop signal
		DrainTimeout time.Duration `yaml:"draintimeout,omitempty"`

		// TLS instructs the http server to listen with a TLS configuration.
		// This only support simple tls configuration with a cert and key.
		// Mostly, this is useful for testing situations or simple deployments
		// that require tls. If more complex configurations are required, use
		// a proxy or make a proposal to add support here.
		TLS struct {
			// Certificate specifies the path to an x509 certificate file to
			// be used for TLS.
			Certificate string `yaml:"certificate,omitempty"`

			// Key specifies the path to the x509 key file, which should
			// contain the private portion for the file specified in
			// Certificate.
			Key string `yaml:"key,omitempty"`

			// Specifies the CA certs for client authentication
			// A file may contain multiple CA certificates encoded as PEM
			ClientCAs []string `yaml:"clientcas,omitempty"`

			// LetsEncrypt is used to configuration setting up TLS through
			// Let's Encrypt instead of manually specifying certificate and
			// key. If a TLS certificate is specified, the Let's Encrypt
			// section will not be used.
			LetsEncrypt struct {
				// CacheFile specifies cache file to use for lets encrypt
				// certificates and keys.
				CacheFile string `yaml:"cachefile,omitempty"`

				// Email is the email to use during Let's Encrypt registration
				Email string `yaml:"email,omitempty"`

				// Hosts specifies the hosts which are allowed to obtain Let's
				// Encrypt certificates.
				Hosts []string `yaml:"hosts,omitempty"`
			} `yaml:"letsencrypt,omitempty"`
		} `yaml:"tls,omitempty"`

		// Headers is a set of headers to include in HTTP responses. A common
		// use case for this would be security headers such as
		// Strict-Transport-Security. The map keys are the header names, and
		// the values are the associated header payloads.
		Headers http.Header `yaml:"headers,omitempty"`

		// Debug configures the http debug interface, if specified. This can
		// include services such as pprof, expvar and other data that should
		// not be exposed externally. Left disabled by default.
		Debug struct {
			// Addr specifies the bind address for the debug server.
			Addr string `yaml:"addr,omitempty"`
			// Prometheus configures the Prometheus telemetry endpoint.
			Prometheus struct {
				Enabled bool   `yaml:"enabled,omitempty"`
				Path    string `yaml:"path,omitempty"`
			} `yaml:"prometheus,omitempty"`
		} `yaml:"debug,omitempty"`

		// HTTP2 configuration options
		HTTP2 struct {
			// Specifies whether the registry should disallow clients attempting
			// to connect via http2. If set to true, only http/1.1 is supported.
			Disabled bool `yaml:"disabled,omitempty"`
		} `yaml:"http2,omitempty"`
	} `yaml:"http,omitempty"`

	// Notifications specifies configuration about various endpoint to which
	// registry events are dispatched.
	Notifications configuration.Notifications `yaml:"notifications,omitempty"`

	// Redis configures the redis pool available to the registry webapp.
	Redis struct {
		// Addr specifies the the redis instance available to the application.
		Addr string `yaml:"addr,omitempty"`

		// Password string to use when making a connection.
		Password string `yaml:"password,omitempty"`

		// DB specifies the database to connect to on the redis instance.
		DB int `yaml:"db,omitempty"`

		DialTimeout  time.Duration `yaml:"dialtimeout,omitempty"`  // timeout for connect
		ReadTimeout  time.Duration `yaml:"readtimeout,omitempty"`  // timeout for reads of data
		WriteTimeout time.Duration `yaml:"writetimeout,omitempty"` // timeout for writes of data

		// Pool configures the behavior of the redis connection pool.
		Pool struct {
			// MaxIdle sets the maximum number of idle connections.
			MaxIdle int `yaml:"maxidle,omitempty"`

			// MaxActive sets the maximum number of connections that should be
			// opened before blocking a connection request.
			MaxActive int `yaml:"maxactive,omitempty"`

			// IdleTimeout sets the amount time to wait before closing
			// inactive connections.
			IdleTimeout time.Duration `yaml:"idletimeout,omitempty"`
		} `yaml:"pool,omitempty"`
	} `yaml:"redis,omitempty"`

	Health configuration.Health `yaml:"health,omitempty"`

	// Compatibility is used for configurations of working with older or deprecated features.
	Compatibility struct {
		// Schema1 configures how schema1 manifests will be handled
		Schema1 struct {
			// TrustKey is the signing key to use for adding the signature to
			// schema1 manifests.
			TrustKey string `yaml:"signingkeyfile,omitempty"`
			// Enabled determines if schema1 manifests should be pullable
			Enabled bool `yaml:"enabled,omitempty"`
		} `yaml:"schema1,omitempty"`
	} `yaml:"compatibility,omitempty"`

	// Validation configures validation options for the registry.
	Validation struct {
		// Enabled enables the other options in this section. This field is
		// deprecated in favor of Disabled.
		Enabled bool `yaml:"enabled,omitempty"`
		// Disabled disables the other options in this section.
		Disabled bool `yaml:"disabled,omitempty"`
		// Manifests configures manifest validation.
		Manifests struct {
			// URLs configures validation for URLs in pushed manifests.
			URLs struct {
				// Allow specifies regular expressions (https://godoc.org/regexp/syntax)
				// that URLs in pushed manifests must match.
				Allow []string `yaml:"allow,omitempty"`
				// Deny specifies regular expressions (https://godoc.org/regexp/syntax)
				// that URLs in pushed manifests must not match.
				Deny []string `yaml:"deny,omitempty"`
			} `yaml:"urls,omitempty"`
		} `yaml:"manifests,omitempty"`
	} `yaml:"validation,omitempty"`

	// Policy configures registry policy options.
	Policy struct {
		// Repository configures policies for repositories
		Repository struct {
			// Classes is a list of repository classes which the
			// registry allows content for. This class is matched
			// against the configuration media type inside uploaded
			// manifests. When non-empty, the registry will enforce
			// the class in authorized resources.
			Classes []string `yaml:"classes"`
		} `yaml:"repository,omitempty"`
	} `yaml:"policy,omitempty"`

	Proxies map[string]Proxy `yaml:"proxies"`
}

// Parse parses an input configuration yaml document into a Configuration struct
// This should generally be capable of handling old configuration format versions
//
// Environment variables may be used to override configuration parameters other than version,
// following the scheme below:
// Configuration.Abc may be replaced by the value of REGISTRY_ABC,
// Configuration.Abc.Xyz may be replaced by the value of REGISTRY_ABC_XYZ, and so forth
func Parse(rd io.Reader) (*Configuration, error) {
	in, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, err
	}

	p := configuration.NewParser("registry", []configuration.VersionedParseInfo{
		{
			Version: configuration.MajorMinorVersion(0, 1),
			ParseAs: reflect.TypeOf(Configuration{}),
			ConversionFunc: func(c interface{}) (interface{}, error) {
				if conf, ok := c.(*Configuration); ok {
					if conf.Storage.Type() == "" {
						return nil, errors.New("no storage configuration provided")
					}
					return conf, nil
				}
				return nil, fmt.Errorf("expected *Configuration, received %#v", c)
			},
		},
	})

	config := new(Configuration)
	err = p.Parse(in, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
