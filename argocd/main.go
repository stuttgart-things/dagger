// Argocd provides Dagger functions for registering clusters in ArgoCD and
// rendering ArgoCD resources from OCI-hosted KCL modules.
//
// Functions live in dedicated files:
//   - cli.go      AddClusterCli    (logs in with the argocd CLI and runs `cluster add`)
//   - cluster.go  AddClusterK8s    (applies a cluster Secret directly; no ArgoCD API call)
//   - project.go  CreateAppProject (renders + optionally applies an AppProject via the shared KCL module)

package main

type Argocd struct{}
