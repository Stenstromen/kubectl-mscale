# kubectl-mscale

![kubectl-mscale logo](./kubectl-mscale.webp)

- [kubectl-mscale](#kubectl-mscale)
  - [Description](#description)
  - [Installation via Homebrew (MacOS/Linux - x86\_64/arm64)](#installation-via-homebrew-macoslinux---x86_64arm64)
  - [Download and Run Binary](#download-and-run-binary)
  - [Build and Run Binary](#build-and-run-binary)
  - [Usage](#usage)
    - [Scale all resources of a specific type across multiple namespaces](#scale-all-resources-of-a-specific-type-across-multiple-namespaces)
    - [Scale one resource with a specific name across multiple namespaces](#scale-one-resource-with-a-specific-name-across-multiple-namespaces)
    - [Scale all resources of a specific type across all namespaces](#scale-all-resources-of-a-specific-type-across-all-namespaces)
    - [Scale from a file](#scale-from-a-file)
    - [Scale with verification of current replicas](#scale-with-verification-of-current-replicas)
  - [Supported Resource Types](#supported-resource-types)
  - [Configuration](#configuration)
  - [Requirements](#requirements)
  - [License](#license)

## Description

A kubectl plugin for scaling resources across multiple namespaces simultaneously. This tool is particularly useful when you need to scale the same resource across different environments or namespaces.

## Installation via Homebrew (MacOS/Linux - x86_64/arm64)

```bash
brew install stenstromen/tap/kubectl-mscale
```

## Download and Run Binary

- For **MacOS** and **Linux**: Checkout and download the latest binary from [Releases page](https://github.com/Stenstromen/kubectl-mscale/releases/latest/)
- For **Windows**: Build the binary yourself.

## Build and Run Binary

```bash
go build
./kubectl-mscale
```

## Usage

The primary use case is scaling resources across multiple namespaces in a single command. Here are some examples:

### Scale all resources of a specific type across multiple namespaces

```bash
# Scale all deployments to 3 replicas across multiple namespaces
kubectl-mscale deployment --replicas=3 -n default,staging,production

# Scale all statefulsets to 2 replicas across multiple namespaces
kubectl-mscale statefulset --replicas=2 -n default,staging,production

# Scale all replicaset to 5 replicas across multiple namespaces
kubectl-mscale replicaset --replicas=5 -n default,staging,production
```

### Scale one resource with a specific name across multiple namespaces

```bash
# Scale a deployment named 'nginx' to 0 replicas across multiple namespaces
kubectl-mscale deployment nginx --replicas=0 -n default,staging,production

# Scale a statefulset named 'mysql' to 1 replica across multiple namespaces
kubectl-mscale statefulset mysql --replicas=1 -n default,staging,production
```

### Scale all resources of a specific type across all namespaces

```bash
# Scale ALL deployments to 0 replicas across multiple namespaces
kubectl-mscale deployment --replicas=0 -n default,staging,production --all

# Scale ALL statefulsets to 1 replica in the default namespace
kubectl-mscale statefulset --replicas=1 --all
```

### Scale from a file

```bash
# Scale resources defined in a YAML file
kubectl-mscale statefulset --filename=statefulset.yaml --replicas=3
```

### Scale with verification of current replicas

```bash
# Only scale if current replicas match the expected value
kubectl-mscale deployment nginx --replicas=5 --current-replicas=3 -n production
```

## Supported Resource Types

The following resource types can be scaled with kubectl-mscale:

- Deployments (`deployment`, `deploy`, `deployments`)
- StatefulSets (`statefulset`, `sts`, `statefulsets`)
- ReplicaSets (`replicaset`, `rs`, `replicasets`)
- ReplicationControllers (`replicationcontroller`, `rc`, `replicationcontrollers`)
- Jobs (`job`, `jobs`)
- CronJobs (`cronjob`, `cj`, `cronjobs`)
- HorizontalPodAutoscalers (`horizontalpodautoscaler`, `hpa`, `horizontalpodautoscalers`)

## Configuration

The plugin will use the Kubernetes configuration from:

1. The KUBECONFIG environment variable if set
2. The default location at ~/.kube/config if KUBECONFIG is not set

## Requirements

- kubectl
- Kubernetes cluster access

## License

MIT
