# How to

```bash
# kube config rule
~/.kube/config--${ENV}.yaml

# values
cp values.yaml values--${ENV}.yaml

# check
make apply.registry-proxy-cache ENV=hw-sg

# deploy
make apply.registry-proxy-cache ENV=hw-sg DEBUG=0 
```