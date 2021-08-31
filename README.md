# kube-svcauth-proxy

Multi-channel reverse proxy, to secure pod service from unauthorized access.

![Diagram](diag.png)

_kube-svcauth-proxy_ can server multiple routes, listen->upstream pairs, each configured separately.
Authorization procedure is simple get request to the configured service name in configured
namespace. With this we don't need any cluster-wide permissions.

## Config

```yaml
proxy:
  - listen: ":9090"
    upstream: "http://127.0.0.1:8080"
    namespace: "app-namespace"
    service: "app-service"
  - listen: ":9091"
    upstream: "http://127.0.0.1:8081"
    namespace: "app-namespace"
    service: "app-service"
```

listen - listen on this address upstream - forward to this address namespace - check access in this
namespace service - check access to this service

## Motivation

We wanted to use [`kube-rbac-proxy`](https://github.com/brancz/kube-rbac-proxy), but that have
several downsides:

- kube-rbac-proxy does not support multiple routes, so that would mean that we would need one
  kube-rbac-proxy instance per port, which is huge overhead. In Eclipse Che, that would mean extra
  container per secured workspace endpoint. That's 5+ extra containers per workspace.
- kube-rbac-proxy needs cluster permissions to `tokenreviews` and `subjectaccessreviews`. That would
  mean that every workspace `ServiceAccount` would need to have these permissions. In DevWorkspace
  world that would mean to have extra `ClusterRoleBinding` per workspace.
