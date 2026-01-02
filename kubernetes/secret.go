package main

import (
	"context"
	"dagger/kubernetes/internal/dagger"
	"fmt"
)

// CreateKubeconfigSecret creates or updates a Kubernetes secret from a kubeconfig file
// This is an idempotent operation using kubectl apply
//
// Example usage:
//
//	dagger call -m kcl create-kubeconfig-secret \
//	  --namespace crossplane-system \
//	  --secret-name dev \
//	  --kubeconfig-file ./kubeconfig.yaml \
//	  --kube-config file:///home/user/.kube/config
//
// Parameters:
// - namespace: Kubernetes namespace where the secret will be created
// - secretName: Name of the secret to create
// - kubeconfigFile: kubeconfig file to use as secret data
// - kubeConfig: kubeconfig for authentication (file:// or contents)
//
// Returns the secret creation status
func (m *Kubernetes) CreateKubeconfigSecret(
	ctx context.Context,
	// Kubernetes namespace where secret will be created
	// +optional
	// +default="crossplane-system"
	namespace string,
	// Name of the secret to create
	// +optional
	// +default="kubeconfig"
	secretName string,
	// Kubeconfig secret to create secret from
	ToBeCreatedKubeconfigSecret *dagger.Secret,
	// Kubeconfig secret for kubectl authentication
	// +optional
	kubeConfigCluster *dagger.Secret,
) (string, error) {

	ctr := m.container()

	// Mount the kubeconfig secret to use as secret data
	ctr = ctr.WithMountedSecret("/kubeconfig.yaml", ToBeCreatedKubeconfigSecret)

	// Mount kubectl authentication kubeconfig secret if provided
	if kubeConfigCluster != nil {
		ctr = ctr.WithMountedSecret("/root/.kube/config", kubeConfigCluster)
	}

	// Create the secret using kubectl (idempotent with apply)
	cmd := fmt.Sprintf(`
kubectl -n "%s" create secret generic "%s" \
  --from-file=config=/kubeconfig.yaml \
  --dry-run=client -o yaml \
  | kubectl apply -f -
`, namespace, secretName)

	result, err := ctr.WithExec([]string{"sh", "-c", cmd}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return result, nil
}
