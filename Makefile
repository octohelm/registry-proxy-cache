PKG = $(shell cat go.mod | grep "^module " | sed -e "s/module //g")
VERSION = $(shell cat .version)
COMMIT_SHA ?= $(shell git rev-parse --short HEAD)

GOBUILD = CGO_ENABLED=0 STATIC=0 go build -ldflags "-extldflags -static -s -w -X $(PKG)/pkg/version.Version=$(VERSION)+sha.$(COMMIT_SHA)"
GOTEST = go test -v -race
GOBIN ?= ./bin
GOBUILD_TAGS= -tags include_oss

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

APP ?= registry-proxy-cache
WORKSPACE ?= ./cmd/$(APP)

DOCKERX_LABELS ?=
HUB ?= docker.io/octohelm
TAG ?= $(COMMIT_SHA)

REGISTRY_CONFIGURATION_PATH = $(WORKSPACE)/config-dev.yaml
NAMESPACE = registry-proxy-cache

REGISTRY_PROXY_CACHE = REGISTRY_CONFIGURATION_PATH=$(REGISTRY_CONFIGURATION_PATH) go run $(WORKSPACE)

up:
	$(REGISTRY_PROXY_CACHE)

gc:
	$(REGISTRY_PROXY_CACHE) gc

test:
	$(GOTEST) ./...

cover:
	$(GOTEST) -coverprofile=coverage.txt -covermode=atomic ./...

build:
	$(GOBUILD) $(GOBUILD_TAGS) -o $(GOBIN)/$(APP)-$(GOOS)-$(GOARCH) $(WORKSPACE)

fmt:
	goimports -w -l .

PLATFORMS = amd64 arm64
BUILDER ?= docker

buildx:
	for arch in $(PLATFORMS); do \
  		$(MAKE) build GOOS=linux GOARCH=$${arch}; \
  	done

dockerx:
	docker buildx build \
		--push \
		--build-arg=BUILDER=$(BUILDER) \
		--build-arg=APP=$(APP) \
		$(foreach label,$(DOCKERX_LABELS),--label=$(label)) \
		$(foreach arch,$(PLATFORMS),--platform=linux/$(arch)) \
		$(foreach hub,$(HUB),--tag=$(hub)/$(APP):$(TAG)) \
		-f $(WORKSPACE)/Dockerfile .

dockerx.dev: buildx
	$(MAKE) dockerx BUILDER=local
