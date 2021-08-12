package configuration

import (
	"bytes"
	"path"

	"github.com/distribution/distribution/v3/configuration"
	"gopkg.in/yaml.v3"
)

func ToDistributionConfigurations(c *Configuration) map[string]*configuration.Configuration {
	if len(c.Proxies) == 0 {
		c.Proxies = map[string]Proxy{}
	}

	if _, ok := c.Proxies["docker.io"]; !ok {
		c.Proxies["docker.io"] = Proxy{
			RemoteURL: "https://registry-1.docker.io",
		}
	}

	m := map[string]*configuration.Configuration{}

	for hub := range c.Proxies {
		proxy := c.Proxies[hub]

		dc := &configuration.Configuration{}

		//dc.Storage

		//var rerootdirectory = regexp.MustCompile("rootdirectory: ([^\n]+)")

		//configYAML2 := rerootdirectory.ReplaceAllFunc(configYAML, func(bytes []byte) []byte {
		//	return append(rerootdirectory.FindSubmatch(bytes)[0], []byte("/"+hub)...)
		//})

		buf := bytes.NewBuffer(nil)

		_ = yaml.NewEncoder(buf).Encode(c)
		_ = yaml.NewDecoder(buf).Decode(dc)

		if dc.Storage != nil {
			parameters := dc.Storage.Parameters()
			if v, ok := parameters["rootdirectory"]; ok {
				if rootDir, ok := v.(string); ok {
					parameters["rootdirectory"] = path.Join(rootDir, hub)
				}
			}
		}

		dc.Proxy.RemoteURL = proxy.RemoteURL
		dc.Proxy.Username = proxy.Username
		dc.Proxy.Password = proxy.Password

		if dc.Log.Fields == nil {
			dc.Log.Fields = map[string]interface{}{}
		}

		dc.Log.Fields["hub"] = hub

		// registry must disable debug
		dc.HTTP.Debug.Addr = ""
		dc.HTTP.Debug.Prometheus.Enabled = false

		m[hub] = dc
	}

	return m
}
