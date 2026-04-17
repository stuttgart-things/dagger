package main

import (
	"context"
	"dagger/argocd/internal/dagger"
	"fmt"
)

// AddClusterK8s registers a Kubernetes cluster in ArgoCD without calling the ArgoCD
// HTTP/gRPC API. It creates (or reuses) a ServiceAccount with cluster-admin permissions in
// the target cluster, mints a token via `kubectl create token`, extracts the cluster's
// server URL and CA from the kubeconfig, and applies the resulting ArgoCD cluster Secret
// (labelled argocd.argoproj.io/secret-type=cluster) in the cluster where ArgoCD runs.
func (m *Argocd) AddClusterK8s(
	ctx context.Context,
	// Kubeconfig of the target cluster to register (where the SA is created)
	kubeConfig *dagger.Secret,
	// Display name for the cluster in ArgoCD (also the Secret name)
	clusterName string,
	// Kubeconfig of the cluster where ArgoCD runs. If omitted, kubeConfig is used.
	// +optional
	argocdKubeConfig *dagger.Secret,
	// Namespace where ArgoCD is installed
	// +optional
	// +default="argocd"
	argocdNamespace string,
	// ServiceAccount name created/reused in the target cluster
	// +optional
	// +default="argocd-manager"
	serviceAccountName string,
	// Namespace for the ServiceAccount in the target cluster
	// +optional
	// +default="kube-system"
	serviceAccountNamespace string,
	// Kubeconfig context of the target cluster. Empty = current-context.
	// +optional
	sourceContext string,
	// Kubeconfig context of the ArgoCD cluster. Empty = current-context of argocdKubeConfig.
	// +optional
	argocdContext string,
	// Override the server URL written into the cluster Secret. Empty = server from kubeconfig.
	// +optional
	serverURL string,
	// Duration passed to `kubectl create token`. Subject to the cluster's max.
	// +optional
	// +default="8760h"
	tokenDuration string,
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	baseImage string,
) (string, error) {

	if clusterName == "" {
		return "", fmt.Errorf("clusterName must not be empty")
	}

	const srcPath = "/tmp/src-kubeconfig"
	const argoPath = "/tmp/argocd-kubeconfig"

	ctr := dag.Container().
		From(baseImage).
		WithExec([]string{"apk", "add", "--no-cache", "kubectl", "jq"}).
		WithMountedSecret(srcPath, kubeConfig, dagger.ContainerWithMountedSecretOpts{
			Mode: 0444,
		})

	argoSecret := argocdKubeConfig
	if argoSecret == nil {
		argoSecret = kubeConfig
	}
	ctr = ctr.WithMountedSecret(argoPath, argoSecret, dagger.ContainerWithMountedSecretOpts{
		Mode: 0444,
	}).
		WithEnvVariable("CLUSTER_NAME", clusterName).
		WithEnvVariable("ARGOCD_NAMESPACE", argocdNamespace).
		WithEnvVariable("SA_NAME", serviceAccountName).
		WithEnvVariable("SA_NAMESPACE", serviceAccountNamespace).
		WithEnvVariable("SOURCE_CONTEXT", sourceContext).
		WithEnvVariable("ARGOCD_CONTEXT", argocdContext).
		WithEnvVariable("SERVER_URL_OVERRIDE", serverURL).
		WithEnvVariable("TOKEN_DURATION", tokenDuration)

	script := `set -eu
cp ` + srcPath + ` /tmp/src.kc
cp ` + argoPath + ` /tmp/argo.kc
chmod 600 /tmp/src.kc /tmp/argo.kc

export KUBECONFIG=/tmp/src.kc
SRC_JSON=$(kubectl config view --flatten --raw -o json)

SRC_CTX="${SOURCE_CONTEXT:-}"
if [ -z "$SRC_CTX" ]; then
  SRC_CTX=$(printf '%s' "$SRC_JSON" | jq -r '."current-context" // empty')
fi
if [ -z "$SRC_CTX" ]; then
  echo "could not determine source context; set --source-context explicitly" >&2
  exit 1
fi

CLUSTER_REF=$(printf '%s' "$SRC_JSON" | jq -r --arg c "$SRC_CTX" '.contexts[] | select(.name==$c) | .context.cluster')
if [ -z "$CLUSTER_REF" ] || [ "$CLUSTER_REF" = "null" ]; then
  echo "context '$SRC_CTX' has no cluster reference" >&2
  exit 1
fi

SERVER=$(printf '%s' "$SRC_JSON" | jq -r --arg c "$CLUSTER_REF" '.clusters[] | select(.name==$c) | .cluster.server')
if [ -n "${SERVER_URL_OVERRIDE:-}" ]; then
  SERVER="$SERVER_URL_OVERRIDE"
fi
if [ -z "$SERVER" ] || [ "$SERVER" = "null" ]; then
  echo "could not determine server URL for cluster '$CLUSTER_REF'" >&2
  exit 1
fi

CA_DATA=$(printf '%s' "$SRC_JSON" | jq -r --arg c "$CLUSTER_REF" '.clusters[] | select(.name==$c) | .cluster["certificate-authority-data"] // empty')
INSECURE=$(printf '%s' "$SRC_JSON" | jq -r --arg c "$CLUSTER_REF" '.clusters[] | select(.name==$c) | .cluster["insecure-skip-tls-verify"] // false')

kubectl --context "$SRC_CTX" apply -f - <<YAML
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${SA_NAME}
  namespace: ${SA_NAMESPACE}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ${SA_NAME}-role
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
- nonResourceURLs: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${SA_NAME}-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ${SA_NAME}-role
subjects:
- kind: ServiceAccount
  name: ${SA_NAME}
  namespace: ${SA_NAMESPACE}
YAML

TOKEN=$(kubectl --context "$SRC_CTX" -n "$SA_NAMESPACE" create token "$SA_NAME" --duration="$TOKEN_DURATION")

if [ -n "$CA_DATA" ]; then
  CONFIG_JSON=$(jq -cn --arg t "$TOKEN" --arg ca "$CA_DATA" '{bearerToken:$t, tlsClientConfig:{caData:$ca}}')
elif [ "$INSECURE" = "true" ]; then
  CONFIG_JSON=$(jq -cn --arg t "$TOKEN" '{bearerToken:$t, tlsClientConfig:{insecure:true}}')
else
  CONFIG_JSON=$(jq -cn --arg t "$TOKEN" '{bearerToken:$t, tlsClientConfig:{}}')
fi

NAME_B64=$(printf '%s' "$CLUSTER_NAME"   | base64 | tr -d '\n')
SERVER_B64=$(printf '%s' "$SERVER"       | base64 | tr -d '\n')
CONFIG_B64=$(printf '%s' "$CONFIG_JSON"  | base64 | tr -d '\n')

cat >/tmp/cluster-secret.yaml <<YAML
apiVersion: v1
kind: Secret
metadata:
  name: ${CLUSTER_NAME}
  namespace: ${ARGOCD_NAMESPACE}
  labels:
    argocd.argoproj.io/secret-type: cluster
type: Opaque
data:
  name: ${NAME_B64}
  server: ${SERVER_B64}
  config: ${CONFIG_B64}
YAML

export KUBECONFIG=/tmp/argo.kc
ARGO_CTX_ARG=""
if [ -n "${ARGOCD_CONTEXT:-}" ]; then
  ARGO_CTX_ARG="--context=${ARGOCD_CONTEXT}"
fi

kubectl $ARGO_CTX_ARG apply -f /tmp/cluster-secret.yaml
echo "registered cluster '$CLUSTER_NAME' -> $SERVER in namespace $ARGOCD_NAMESPACE"
`

	return ctr.WithExec([]string{"sh", "-c", script}).Stdout(ctx)
}
