ARG BUILDER=docker
FROM --platform=${BUILDPLATFORM} golang:1.15-buster AS builder-docker

ARG GOPROXY
WORKDIR /go/src
COPY ./ ./

ENV GOBIN=/go/bin
ARG TARGETARCH
RUN GOARCH=${TARGETARCH} make build

FROM busybox as builder-local

COPY ./bin /go/bin

FROM builder-${BUILDER} AS builder

FROM alpine:3.12

RUN set -eux; \
    \
    apk add --no-cache \
        ca-certificates apache2-utils

EXPOSE 5000
# debug
EXPOSE 5005

ARG TARGETARCH
COPY --from=builder /go/bin/registry-proxy-cache-linux-${TARGETARCH} /go/bin/registry-proxy-cache
COPY cmd/registry-proxy-cache/config.yaml /etc/docker/registry/config.yml

ENV REGISTRY_CONFIGURATION_PATH /etc/docker/registry/config.yml

ENTRYPOINT ["/go/bin/registry-proxy-cache"]
