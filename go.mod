module github.com/octohelm/registry-proxy-cache

go 1.15

require (
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/go-metrics v0.0.1
	github.com/gorilla/handlers v0.0.0-20150720190736-60c7bfde3e33
	github.com/onsi/gomega v1.12.0
	github.com/pkg/errors v0.9.1
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v0.0.0-20201218233920-35f1369d3770
