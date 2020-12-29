package registryproxycache

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/octohelm/registry-proxy-cache/pkg/configuration"
	. "github.com/onsi/gomega"
)

var c, _ = configuration.Parse(bytes.NewBuffer([]byte(`
version: 0.1

log:
  level: info

http:
  addr: 0.0.0.0:5001

proxies:
  docker.io:
    remoteurl: https://registry-1.docker.io

storage:
  filesystem:
    maxthreads: 100
    rootdirectory: /tmp/registry-cache
`)))

var c2, _ = configuration.Parse(bytes.NewBuffer([]byte(`
version: 0.1

log:
  level: info

http:
  addr: 0.0.0.0:5002

proxies:
  docker.io:
    remoteurl: http://` + c.HTTP.Addr + `

storage:
  filesystem:
    maxthreads: 100
    rootdirectory: /tmp/registry-cache-2
`)))

func TestRegistryProxyCache(t *testing.T) {
	t.Run("proxy one", func(t *testing.T) {
		r, err := NewRegistryProxyCache(context.Background(), c)
		NewWithT(t).Expect(err).Should(BeNil())

		go func() {
			_ = r.ListenAndServe()
		}()

		t.Run("as mirror", func(t *testing.T) {
			data, err := manifest("library/busybox", "latest", "/mirrors/docker.io/v2/", c.HTTP.Addr)
			NewWithT(t).Expect(err).Should(BeNil())
			NewWithT(t).Expect(data["manifests"]).ShouldNot(BeNil())

			data2, err := manifest("library/busybox", "latest", "/mirrors/docker.io/", c.HTTP.Addr)
			NewWithT(t).Expect(err).Should(BeNil())
			NewWithT(t).Expect(data2["manifests"]).ShouldNot(BeNil())
		})

		t.Run("hub prefix", func(t *testing.T) {
			data, err := manifest("docker.io/library/busybox", "latest", "/v2/", c.HTTP.Addr)
			NewWithT(t).Expect(err).Should(BeNil())
			NewWithT(t).Expect(data["manifests"]).ShouldNot(BeNil())
		})

		_ = r.Server.Shutdown(context.Background())
	})

	t.Run("proxy chain", func(t *testing.T) {
		r, err := NewRegistryProxyCache(context.Background(), c)
		NewWithT(t).Expect(err).Should(BeNil())

		go func() {
			_ = r.ListenAndServe()
		}()

		r2, err := NewRegistryProxyCache(context.Background(), c2)
		NewWithT(t).Expect(err).Should(BeNil())

		go func() {
			_ = r2.ListenAndServe()
		}()

		t.Run("hub prefix mirror", func(t *testing.T) {
			data, err := manifest("library/busybox", "latest", "/hub-prefix-mirrors/docker.io/v2/", c2.HTTP.Addr)
			NewWithT(t).Expect(err).Should(BeNil())
			NewWithT(t).Expect(data["manifests"]).ShouldNot(BeNil())
		})

		_ = r.Server.Shutdown(context.Background())
		_ = r2.Server.Shutdown(context.Background())
	})
}

func manifest(name string, ref string, prefix string, host string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", "http://"+host+prefix+name+"/manifests/"+ref, nil)
	if err != nil {
		return nil, err
	}

	req.Header["Accept"] = []string{
		"application/vnd.oci.image.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.v1+prettyjws",
		"application/json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.oci.image.index.v1+json",
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := map[string]interface{}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
