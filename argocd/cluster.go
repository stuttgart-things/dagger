package main

import (
	"context"
	"dagger/argocd/internal/dagger"
	"fmt"
)

// AddClusterK8s registers a Kubernetes cluster in ArgoCD without calling the ArgoCD
// HTTP/gRPC API. It creates (or reuses) a ServiceAccount with cluster-admin permissions
// in the target cluster, mints a token via `kubectl create token`, extracts the cluster's
// server URL and CA from the kubeconfig, and assembles the ArgoCD cluster Secret
// (labelled argocd.argoproj.io/secret-type=cluster).
//
// The rendered Secret is always returned in the output Directory as `<clusterName>.yaml`.
// When applyToCluster is true the Secret is also applied to the ArgoCD-hosting cluster;
// when false (the default, matching create-app-project / create-application*) you get
// the file back without touching the ArgoCD cluster — handy for git-committing or
// inspecting before apply. Either way, the target cluster IS mutated (SA + RBAC
// created, token minted) because the Secret can't be built without a live token.
func (m *Argocd) AddClusterK8s(
	ctx context.Context,
	// Kubeconfig of the target cluster to register (where the SA is created)
	kubeConfig *dagger.Secret,
	// Display name for the cluster in ArgoCD (also the Secret name and output filename)
	clusterName string,
	// Kubeconfig of the cluster where ArgoCD runs. Required when applyToCluster is true
	// (and you want to apply somewhere other than the target cluster). Ignored when
	// applyToCluster is false.
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
	// Apply the generated cluster Secret to the ArgoCD cluster. When false (default),
	// the Secret is only rendered and returned — inspect/commit it, apply later with
	// your own tooling (or pipe it through SOPS first).
	// +optional
	// +default=false
	applyToCluster bool,
	// +optional
	// +default="cgr.dev/chainguard/wolfi-base:latest"
	baseImage string,
) (*dagger.Directory, error) {

	if clusterName == "" {
		return nil, fmt.Errorf("clusterName must not be empty")
	}

	const srcPath = "/tmp/src-kubeconfig"
	const argoPath = "/tmp/argocd-kubeconfig"
	const secretPath = "/tmp/cluster-secret.yaml"

	ctr := dag.Container().
		From(baseImage).
		WithExec([]string{"apk", "add", "--no-cache", "kubectl", "jq"}).
		WithMountedSecret(srcPath, kubeConfig, dagger.ContainerWithMountedSecretOpts{
			Mode: 0444,
		})

	apply := "0"
	if applyToCluster {
		apply = "1"
		argoSecret := argocdKubeConfig // pragma: allowlist secret
		if argoSecret == nil {         // pragma: allowlist secret
			argoSecret = kubeConfig // pragma: allowlist secret
		}
		ctr = ctr.WithMountedSecret(argoPath, argoSecret, dagger.ContainerWithMountedSecretOpts{
			Mode: 0444,
		})
	}

	ctr = ctr.
		WithEnvVariable("CLUSTER_NAME", clusterName).
		WithEnvVariable("ARGOCD_NAMESPACE", argocdNamespace).
		WithEnvVariable("SA_NAME", serviceAccountName).
		WithEnvVariable("SA_NAMESPACE", serviceAccountNamespace).
		WithEnvVariable("SOURCE_CONTEXT", sourceContext).
		WithEnvVariable("ARGOCD_CONTEXT", argocdContext).
		WithEnvVariable("SERVER_URL_OVERRIDE", serverURL).
		WithEnvVariable("TOKEN_DURATION", tokenDuration).
		WithEnvVariable("APPLY", apply)

	script := `set -eu
cp ` + srcPath + ` /tmp/src.kc
chmod 600 /tmp/src.kc

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

cat >` + secretPath + ` <<YAML
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

if [ "${APPLY:-0}" = "1" ]; then
  cp ` + argoPath + ` /tmp/argo.kc
  chmod 600 /tmp/argo.kc
  export KUBECONFIG=/tmp/argo.kc
  ARGO_CTX_ARG=""
  if [ -n "${ARGOCD_CONTEXT:-}" ]; then
    ARGO_CTX_ARG="--context=${ARGOCD_CONTEXT}"
  fi
  kubectl $ARGO_CTX_ARG apply -f ` + secretPath + `
  echo "registered cluster '$CLUSTER_NAME' -> $SERVER in namespace $ARGOCD_NAMESPACE"
else
  echo "rendered cluster Secret for '$CLUSTER_NAME' -> $SERVER (apply skipped)"
fi
`

	execCtr := ctr.WithExec([]string{"sh", "-c", script})

	secretContents, err := execCtr.File(secretPath).Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated cluster Secret: %w", err)
	}

	outputDir := dag.Directory().WithNewFile(clusterName+".yaml", secretContents)
	return outputDir, nil
}
