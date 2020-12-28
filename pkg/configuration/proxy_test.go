package configuration

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestParseRegistryProxies(t *testing.T) {
	proxies := "docker.io+https://username:password@registry-1.docker.io quay.io+https://quay.io ghcr.io+https://:@ghcr.io"

	ps, err := ParseProxies(proxies)
	NewWithT(t).Expect(err).Should(BeNil())

	NewWithT(t).Expect(ps["docker.io"]).Should(Equal(Proxy{
		RemoteURL: "https://registry-1.docker.io",
		Username:  "username",
		Password:  "password",
	}))

	NewWithT(t).Expect(ps["quay.io"]).Should(Equal(Proxy{
		RemoteURL: "https://quay.io",
	}))

	NewWithT(t).Expect(ps["ghcr.io"]).Should(Equal(Proxy{
		RemoteURL: "https://ghcr.io",
	}))
}
