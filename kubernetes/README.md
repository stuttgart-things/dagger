# Kubernetes Dagger Module

A Dagger module for executing kubectl commands in a containerized environment with kubectl and helm pre-installed.

## Features

- Execute kubectl commands with custom operations and resource kinds
- Support for namespace filtering (specific namespace, all namespaces, or cluster-wide)
- Pipe kubectl output through additional shell commands
- Use custom kubeconfig files
- Built on Chainguard Wolfi base image with kubectl and helm

## Examples

### Basic Operations

#### Get all pods across all namespaces
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind pods \
  --namespace ALL \
  --kube-config file://~/.kube/config \
  --progress plain
```

#### Get pods in a specific namespace
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind pods \
  --namespace default \
  --kube-config file://~/.kube/config \
  --progress plain
```

#### Get cluster-wide resources (nodes, persistent volumes, etc.)
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind nodes \
  --kube-config file://~/.kube/config \
  --progress plain
```

#### Apply kustomize resources

```bash
dagger call -m kubernetes kubectl \
  --operation apply \
  --kustomize-source https://github.com/stuttgart-things/helm/infra/crds/cilium \
  --namespace kube-system \
  --kube-config file://~/.kube/config
```

### Filtering with Additional Commands

#### Find pods with "ingress" in the name
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind pods \
  --namespace ALL \
  --additional-command="grep ingress" \
  --kube-config file://~/.kube/demo-infra \
  --progress plain
```

#### Count running pods across all namespaces
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind pods \
  --namespace ALL \
  --additional-command="grep -c Running" \
  --kube-config file://~/.kube/config \
  --progress plain
```

#### Get pods and format output with awk
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind pods \
  --namespace ALL \
  --additional-command="awk '{print \$1, \$3}'" \
  --kube-config file://~/.kube/config \
  --progress plain
```

### Resource Management

#### Describe a specific deployment
```bash
dagger call -m kubernetes command \
  --operation describe \
  --resource-kind deployment/nginx \
  --namespace production \
  --kube-config file://~/.kube/config \
  --progress plain
```

#### Get services in JSON format
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind services \
  --namespace default \
  --additional-command="kubectl get svc -o json" \
  --kube-config file://~/.kube/config \
  --progress plain
```

#### List all deployments with specific labels
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind deployments \
  --namespace ALL \
  --additional-command="grep app=frontend" \
  --kube-config file://~/.kube/config \
  --progress plain
```

### Troubleshooting

#### Get events in a namespace
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind events \
  --namespace kube-system \
  --additional-command="tail -20" \
  --kube-config file://~/.kube/config \
  --progress plain
```

#### Check pod logs (using describe to see pod details)
```bash
dagger call -m kubernetes command \
  --operation describe \
  --resource-kind pod/my-pod-name \
  --namespace default \
  --kube-config file://~/.kube/config \
  --progress plain
```

#### Find pods that are not running
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind pods \
  --namespace ALL \
  --additional-command="grep -v Running | grep -v Completed" \
  --kube-config file://~/.kube/config \
  --progress plain
```

### Advanced Usage

#### Get resource usage with top
```bash
dagger call -m kubernetes command \
  --operation top \
  --resource-kind pods \
  --namespace ALL \
  --kube-config file://~/.kube/config \
  --progress plain
```

#### List custom resources
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind crd \
  --kube-config file://~/.kube/config \
  --progress plain
```

#### Search for specific configuration in configmaps
```bash
dagger call -m kubernetes command \
  --operation get \
  --resource-kind configmaps \
  --namespace default \
  --additional-command="grep -A 5 'key-name'" \
  --kube-config file://~/.kube/config \
  --progress plain
```

## Parameters

- **operation**: kubectl operation (default: "get")
  - Examples: get, describe, top, explain, api-resources
- **resource-kind**: Kubernetes resource type (default: "pods")
  - Examples: pods, deployments, services, nodes, configmaps, secrets, ingresses
- **namespace**: Target namespace
  - Empty: cluster-wide resources
  - Specific namespace: `--namespace default`
  - All namespaces: `--namespace ALL` (or `all`, `*`)
- **kube-config**: Path to kubeconfig file
  - Local file: `file://~/.kube/config`
  - Environment variable: Uses Dagger secret mounting
- **additional-command**: Optional shell command to pipe kubectl output through
  - Examples: grep, awk, sed, tail, head, wc

## Notes

- The module uses `cgr.dev/chainguard/wolfi-base:latest` as the base image
- kubectl and helm are pre-installed in the container
- Kubeconfig is securely mounted as a Dagger secret at `/root/.kube/config`
