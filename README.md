# Registry Proxy Cache

## Usage

### Quick Start

```bash
export REGISTRY_PROXIES=(
    docker.io+https://${DOCKER_USERNAME}:${DOCKER_PASSWORD}@registry-1.docker.io
    quay.io+https://quay.io
    ghcr.io+https://ghcr.io
    gcr.io+https://gcr.io
    k8s.gcr.io+https://k8s.gcr.io
)

docker run \
    -p 5000:5000 \
    -v /data/registry-cache:/data/registry-cache \
    -e REGISTRY_PROXIES="${REGISTRY_PROXIES}" \
    docker.io/octohelm/registry-proxy-cache:master    
```

### Pull Directly

```bash

docker pull ${registry_proxy_cache_ip}:5000/docker.io/library/nginx
```

### Use as Registry Mirrors

[configure registry endpoint of containerd](https://github.com/containerd/cri/blob/master/docs/registry.md#configure-registry-endpoint)

```toml
version = 2

[plugins."io.containerd.grpc.v1.cri".registry]
[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
endpoint = ["http://${registry_proxy_cache_ip}:5000/mirrors/docker.io/"]
# if have pull like `docker pull myhub/docker.io/library/busybox`
# should use `/hub-prefix-mirrors/{hub}`
# endpoint = ["http://${registry_proxy_cache_ip}:5000/hub-prefix-mirrors/docker.io/"]
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."quay.io"]
endpoint = ["http://${registry_proxy_cache_ip}:5000/mirrors/quay.io/"]
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."gcr.io"]
endpoint = ["http://${registry_proxy_cache_ip}:5000/mirrors/gcr.io/"]
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."ghcr.io"]
endpoint = ["http://${registry_proxy_cache_ip}:5000/mirrors/ghcr.io/"]
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."k8s.gcr.io"]
endpoint = ["http://${registry_proxy_cache_ip}:5000/mirrors/k8s.gcr.io/"]
```