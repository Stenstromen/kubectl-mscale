# kubectl-mscale

A kubectl plugin for scaling resources across multiple namespaces simultaneously. This tool is particularly useful when you need to scale the same resource across different environments or namespaces.

## Installation

```bash
# TBA
```

## Usage

The primary use case is scaling resources across multiple namespaces in a single command. Here are some examples:

### Scale a single resource across multiple namespaces

```bash
# Scale a deployment named 'nginx' to 3 replicas across multiple namespaces
kubectl-mscale deployment/nginx --replicas=3 -n default,staging,production

# Scale a statefulset named 'mysql' to 2 replicas across multiple namespaces
kubectl-mscale statefulset/mysql --replicas=2 -n default,staging,production

# Scale a replicaset named 'web' to 5 replicas across multiple namespaces
kubectl-mscale replicaset/web --replicas=5 -n default,staging,production
```

### Scale multiple resources across multiple namespaces

```bash
# Scale multiple deployments to 0 replicas across multiple namespaces
kubectl-mscale deployment/nginx deployment/redis --replicas=0 -n default,staging,production

# Scale multiple statefulsets to 1 replica across multiple namespaces
kubectl-mscale statefulset/mysql statefulset/redis --replicas=1 -n default,staging,production
```

## Supported Resource Types

- Deployments
- StatefulSets
- ReplicaSets

## Requirements

- kubectl
- krew
- Kubernetes cluster access

## License

MIT
