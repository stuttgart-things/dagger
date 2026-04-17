package main

import (
	"bytes"
	"context"
	"dagger/argocd/internal/dagger"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
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

	access, err := ensureTargetAccess(ctx, targetAccessInput{
		kubeConfig:              kubeConfig,
		serviceAccountName:      serviceAccountName,
		serviceAccountNamespace: serviceAccountNamespace,
		sourceContext:           sourceContext,
		tokenDuration:           tokenDuration,
		serverURLOverride:       serverURL,
		baseImage:               baseImage,
	})
	if err != nil {
		return nil, fmt.Errorf("prepare target cluster: %w", err)
	}

	secretYAML, err := renderClusterSecret(clusterName, argocdNamespace, access)
	if err != nil {
		return nil, fmt.Errorf("render cluster Secret: %w", err)
	}

	outputPath := clusterName + ".yaml"
	outputDir := dag.Directory().WithNewFile(outputPath, secretYAML)

	if applyToCluster {
		argoKC := argocdKubeConfig
		if argoKC == nil {
			argoKC = kubeConfig
		}
		if err := applyClusterSecret(ctx, applyInput{
			kubeConfig:      argoKC,
			kubeContext:     argocdContext,
			secretFile:      outputDir.File(outputPath),
			argocdNamespace: argocdNamespace,
			baseImage:       baseImage,
		}); err != nil {
			return nil, fmt.Errorf("apply cluster Secret: %w", err)
		}
	}

	return outputDir, nil
}

type applyInput struct {
	kubeConfig      *dagger.Secret
	kubeContext     string
	secretFile      *dagger.File
	argocdNamespace string
	baseImage       string
}

// applyClusterSecret runs `kubectl apply` against the ArgoCD-hosting cluster. When
// no explicit context is requested, it delegates to the shared kubernetes module so
// the apply path matches project.go / application.go. When a context is set, the
// shared Kubectl has no --context option, so we fall back to a thin container.
func applyClusterSecret(ctx context.Context, in applyInput) error {
	if in.kubeContext == "" {
		_, err := dag.Kubernetes().Kubectl(ctx, dagger.KubernetesKubectlOpts{
			Operation:  "apply",
			SourceFile: in.secretFile,
			KubeConfig: in.kubeConfig,
			Namespace:  in.argocdNamespace,
		})
		return err
	}

	const kubeconfigPath = "/tmp/argo.kc"
	const secretPath = "/tmp/cluster-secret.yaml"

	_, err := dag.Container().
		From(in.baseImage).
		WithExec([]string{"apk", "add", "--no-cache", "kubectl"}).
		WithMountedSecret(kubeconfigPath, in.kubeConfig, dagger.ContainerWithMountedSecretOpts{Mode: 0444}).
		WithMountedFile(secretPath, in.secretFile).
		WithEnvVariable("KUBECONFIG", kubeconfigPath).
		WithExec([]string{
			"kubectl",
			"--context=" + in.kubeContext,
			"-n", in.argocdNamespace,
			"apply", "-f", secretPath,
		}).
		Stdout(ctx)
	return err
}

// targetClusterAccess is everything renderClusterSecret needs from the target cluster.
type targetClusterAccess struct {
	ServerURL string
	CAData    string // base64-encoded, empty when absent
	Insecure  bool
	Token     string
}

type targetAccessInput struct {
	kubeConfig              *dagger.Secret
	serviceAccountName      string
	serviceAccountNamespace string
	sourceContext           string
	tokenDuration           string
	serverURLOverride       string
	baseImage               string
}

// ensureTargetAccess provisions the SA + cluster-admin RBAC in the target cluster,
// mints a token, and extracts the server URL / CA from the kubeconfig. The shell
// script emits a single JSON line on stdout so parsing stays in Go.
func ensureTargetAccess(ctx context.Context, in targetAccessInput) (*targetClusterAccess, error) {
	const srcPath = "/tmp/src-kubeconfig"
	const rbacPath = "/tmp/rbac.yaml"

	rbac, err := renderRBAC(in.serviceAccountName, in.serviceAccountNamespace)
	if err != nil {
		return nil, fmt.Errorf("render RBAC: %w", err)
	}

	ctr := dag.Container().
		From(in.baseImage).
		WithExec([]string{"apk", "add", "--no-cache", "kubectl", "jq"}).
		WithMountedSecret(srcPath, in.kubeConfig, dagger.ContainerWithMountedSecretOpts{Mode: 0444}).
		WithNewFile(rbacPath, rbac).
		WithEnvVariable("SA_NAME", in.serviceAccountName).
		WithEnvVariable("SA_NAMESPACE", in.serviceAccountNamespace).
		WithEnvVariable("SOURCE_CONTEXT", in.sourceContext).
		WithEnvVariable("TOKEN_DURATION", in.tokenDuration).
		WithEnvVariable("SERVER_URL_OVERRIDE", in.serverURLOverride)

	script := `set -eu
cp ` + srcPath + ` /tmp/src.kc
chmod 600 /tmp/src.kc
export KUBECONFIG=/tmp/src.kc

SRC_JSON=$(kubectl config view --flatten --raw -o json)

SRC_CTX="${SOURCE_CONTEXT:-}"
[ -n "$SRC_CTX" ] || SRC_CTX=$(printf '%s' "$SRC_JSON" | jq -r '."current-context" // empty')
[ -n "$SRC_CTX" ] || { echo "could not determine source context; set --source-context explicitly" >&2; exit 1; }

CLUSTER_REF=$(printf '%s' "$SRC_JSON" | jq -r --arg c "$SRC_CTX" '.contexts[] | select(.name==$c) | .context.cluster')
[ -n "$CLUSTER_REF" ] && [ "$CLUSTER_REF" != "null" ] || { echo "context '$SRC_CTX' has no cluster reference" >&2; exit 1; }

SERVER=$(printf '%s' "$SRC_JSON" | jq -r --arg c "$CLUSTER_REF" '.clusters[] | select(.name==$c) | .cluster.server')
[ -n "${SERVER_URL_OVERRIDE:-}" ] && SERVER="$SERVER_URL_OVERRIDE"
[ -n "$SERVER" ] && [ "$SERVER" != "null" ] || { echo "could not determine server URL for cluster '$CLUSTER_REF'" >&2; exit 1; }

CA_DATA=$(printf '%s' "$SRC_JSON" | jq -r --arg c "$CLUSTER_REF" '.clusters[] | select(.name==$c) | .cluster["certificate-authority-data"] // empty')
INSECURE=$(printf '%s' "$SRC_JSON" | jq -r --arg c "$CLUSTER_REF" '.clusters[] | select(.name==$c) | .cluster["insecure-skip-tls-verify"] // false')

kubectl --context "$SRC_CTX" apply -f ` + rbacPath + ` >&2

TOKEN=$(kubectl --context "$SRC_CTX" -n "$SA_NAMESPACE" create token "$SA_NAME" --duration="$TOKEN_DURATION")

jq -cn \
  --arg server "$SERVER" \
  --arg ca "$CA_DATA" \
  --arg insecure "$INSECURE" \
  --arg token "$TOKEN" \
  '{serverURL:$server, caData:$ca, insecure:($insecure=="true"), token:$token}'
`

	out, err := ctr.WithExec([]string{"sh", "-c", script}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	var access targetClusterAccess
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &access); err != nil {
		return nil, fmt.Errorf("parse access JSON: %w", err)
	}
	if access.Token == "" {
		return nil, fmt.Errorf("empty token returned from target cluster")
	}
	return &access, nil
}

const rbacTmpl = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{ .Name }}-role
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
  name: {{ .Name }}-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ .Name }}-role
subjects:
- kind: ServiceAccount
  name: {{ .Name }}
  namespace: {{ .Namespace }}
`

func renderRBAC(saName, saNamespace string) (string, error) {
	return execTemplate("rbac", rbacTmpl, map[string]string{
		"Name":      saName,
		"Namespace": saNamespace,
	})
}

const clusterSecretTmpl = `apiVersion: v1
kind: Secret
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    argocd.argoproj.io/secret-type: cluster
type: Opaque
data:
  name: {{ .NameB64 }}
  server: {{ .ServerB64 }}
  config: {{ .ConfigB64 }}
`

// renderClusterSecret returns the ArgoCD cluster Secret YAML as a string. Pure Go,
// no container, no shell. The config blob is the ArgoCD cluster-secret schema
// documented at https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#clusters.
func renderClusterSecret(clusterName, argocdNamespace string, access *targetClusterAccess) (string, error) {
	type tlsClientConfig struct {
		CAData   string `json:"caData,omitempty"`
		Insecure bool   `json:"insecure,omitempty"`
	}
	type clusterConfig struct {
		BearerToken     string          `json:"bearerToken"`
		TLSClientConfig tlsClientConfig `json:"tlsClientConfig"`
	}

	cfg := clusterConfig{BearerToken: access.Token}
	switch {
	case access.CAData != "":
		cfg.TLSClientConfig.CAData = access.CAData
	case access.Insecure:
		cfg.TLSClientConfig.Insecure = true
	}

	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal cluster config: %w", err)
	}

	return execTemplate("cluster-secret", clusterSecretTmpl, map[string]string{
		"Name":      clusterName,
		"Namespace": argocdNamespace,
		"NameB64":   base64.StdEncoding.EncodeToString([]byte(clusterName)),
		"ServerB64": base64.StdEncoding.EncodeToString([]byte(access.ServerURL)),
		"ConfigB64": base64.StdEncoding.EncodeToString(cfgJSON),
	})
}

func execTemplate(name, body string, data any) (string, error) {
	t, err := template.New(name).Option("missingkey=error").Parse(body)
	if err != nil {
		return "", fmt.Errorf("parse %s template: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute %s template: %w", name, err)
	}
	return buf.String(), nil
}
