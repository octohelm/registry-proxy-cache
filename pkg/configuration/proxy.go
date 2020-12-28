package configuration

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// Proxy configures the registry as a pull through cache
type Proxy struct {
	// RemoteURL is the URL of the remote registry
	RemoteURL string `yaml:"remoteurl"`

	// Username of the hub user
	Username string `yaml:"username"`

	// Password of the hub user
	Password string `yaml:"password"`
}

func ParseProxies(proxyList string) (map[string]Proxy, error) {
	list := strings.Split(strings.TrimSpace(proxyList), " ")

	proxies := map[string]Proxy{}

	for _, p := range list {
		i := strings.Index(p, "+")
		if i == -1 {
			continue
		}

		hub := p[0:i]
		endpoint := p[i+1:]

		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, errors.Wrapf(err, "hub `%s` parse failed", hub)
		}

		p := Proxy{}

		p.RemoteURL = fmt.Sprintf("%s://%s", u.Scheme, u.Host)

		if u.User != nil {
			p.Username = u.User.Username()
			p.Password, _ = u.User.Password()
		}

		proxies[hub] = p
	}

	return proxies, nil
}
