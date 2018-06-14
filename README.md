# Drone Kube

Drone plugin to update kubernetes deployments

See the [DOC](DOCS.md) file for usage. 



## Usage

```yaml
pipeline:
  deploy:
    # plugins image
    image: win7/drone-kube:v1.0 
    # this will delete all pod for Deployment/myapp, then start new
    kind: Deployment
    workload: myapp
    # workload namespace
    namespace: kube-system
    // k8s server
    server: https://192.168.88.200:6443
    ca: base64ofca
    token: base64oftoken

```
