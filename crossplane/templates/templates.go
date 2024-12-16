/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package templates

// FunctionPackage represents the details of a Crossplane function package
type FunctionPackage struct {
	Name       string
	PackageURL string
	Version    string
	ApiVersion string
}

type TemplateDestination struct {
	Template    string
	Destination string
}

var PackageFiles = []TemplateDestination{
	{
		Template:    Claim,
		Destination: "examples/claim.yaml",
	},
	{
		Template:    Functions,
		Destination: "examples/functions.yaml",
	},
	{
		Template:    Composition,
		Destination: "apis/composition.yaml",
	},
	{
		Template:    Readme,
		Destination: "README.md",
	},
	{
		Template:    Definition,
		Destination: "apis/definition.yaml",
	},
	{
		Template:    Configuration,
		Destination: "crossplane.yaml",
	},
}

var Claim = `---
apiVersion: {{ .apiGroup }}/{{ .claimApiVersion }}
kind: {{ .claimKind }}
metadata:
  name: {{ .claimName }}
  namespace: {{ .namespace }}
spec:
`

var Functions = `---{{- range .functions }}
apiVersion: {{ .ApiVersion }}
kind: Function
metadata:
  name: {{ .Name }}
spec:
  package: {{ .PackageURL }}:{{ .Version }}
---
{{- end }}
`

var Composition = `---
apiVersion: {{ .compositionApiVersion }}
kind: Composition
metadata:
  labels:
    crossplane.io/xrd: {{ .plural }}.{{ .apiGroup }}
  name: {{ .name }}
spec:
  compositeTypeRef:
    apiVersion: {{ .apiGroup }}/{{ .claimApiVersion }}
    kind: {{ .kind }}
  #pipeline:
  #  - step: <REPLACE_ME>
  #    functionRef:
  #      name: function-go-templating
  #    input:
  #      apiVersion: gotemplating.fn.crossplane.io/v1beta1
  #      kind: GoTemplate
  #      source: Inline
  #      inline:
  #        template: |
  #          apiVersion: <REPLACE_ME>
  #          kind: <REPLACE_ME>
  #          metadata:
  #            annotations:
  #              gotemplating.fn.crossplane.io/composition-resource-name: $CLAIMNAME
  #              gotemplating.fn.crossplane.io/ready: "True"
  #  - step: <REPLACE_ME>
  #    functionRef:
  #      name: function-patch-and-transform
  #    input:
  #      apiVersion: pt.fn.crossplane.io/v1beta1
  #      environment: null
  #      kind: Resources
  #      patchSets: []
  #      resources:
  #        - name: <REPLACE_ME>
  #          base:
  #            apiVersion: <REPLACE_ME>
  #            kind: <REPLACE_ME>
  #          patches: {}
`

var Definition = `---
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: {{ .plural }}.{{ .apiGroup }}
spec:
  group: {{ .apiGroup }}
  names:
    kind: {{ .kind }}
    plural: {{ .plural }}
  claimNames:
    kind: {{ .claimKind }}
    plural: {{ .claimPlural }}
  versions:
    - name: v1alpha1
      served: true
      referenceable: true
      schema:
        openAPIV3Schema:
          description: A {{ .claimKind }} is a composite resource that represents
          type: object
          properties:
            spec:
              type: object
              properties:
                <REPLACE_ME>:
                  type: string
                  default: <REPLACE_ME>
                  description: <REPLACE_ME>
            status:
              description: A Status represents the observed state
              properties:
                <REPLACE_ME>:
                  description: Freeform field containing status information
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
              type: object
`

var Configuration = `---
apiVersion: meta.pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: {{ .kind }}
  annotations:
    meta.crossplane.io/maintainer: {{ .maintainer }}
    meta.crossplane.io/source: {{ .source }}
    meta.crossplane.io/license: {{ .license }}
    meta.crossplane.io/description: |
      deploys {{ .claimKind }} w/ crossplane
    meta.crossplane.io/readme: |
      deploys {{ .claimKind }} w/ crossplane
spec:
  crossplane:
    version: ">={{ .crossplaneVersion }}"
  dependsOn:
    - provider: xpkg.upbound.io/crossplane-contrib/provider-helm
      version: ">=v0.19.0"
    - provider: xpkg.upbound.io/crossplane-contrib/provider-kubernetes
      version: ">=v0.14.1"
`

var Readme = `# {{ .claimKind }}

// ## PROVIDER-CONFIG

// ### CREATE KUBECONFIG AS A SECRET FROM LOCAL FILE

// ```bash
// CROSSPLANE_NAMESPACE=crossplane-system
// CLUSTER_NAME=local
// FOLDER_KUBECONFIG=~/.kube/
// FILENAME_KUBECONFIG=rke2.yaml
// ```

// ```bash
// kubectl -n ${CROSSPLANE_NAMESPACE} create secret generic ${CLUSTER_NAME} --from-file=${FOLDER_KUBECONFIG}/${FILENAME_KUBECONFIG}
// ```

// ### CREATE KUBERNETES PROVIDER CONFIG

// ```bash
// kubectl apply -f - <<EOF
// apiVersion: kubernetes.crossplane.io/v1alpha1
// kind: ProviderConfig
// metadata:
//   name: ${CLUSTER_NAME}
// spec:
//   credentials:
//     source: Secret
//     secretRef:
//       namespace: ${CROSSPLANE_NAMESPACE}
//       name: ${CLUSTER_NAME}
//       key: ${FILENAME_KUBECONFIG}
// EOF
// ```

// ### CREATE KUBERNETES PROVIDER CONFIG

// ```bash
// kubectl apply -f - <<EOF
// apiVersion: kubernetes.crossplane.io/v1alpha1
// kind: ProviderConfig
// metadata:
//   name: ${CLUSTER_NAME}
// spec:
//   credentials:
//     source: Secret
//     secretRef:
//       namespace: ${CROSSPLANE_NAMESPACE}
//       name: ${CLUSTER_NAME}
//       key: ${FILENAME_KUBECONFIG}
// EOF
// ```

// ### CREATE HELM PROVIDER CONFIG

// ```bash
// kubectl apply -f - <<EOF
// apiVersion: helm.crossplane.io/v1beta1
// kind: ProviderConfig
// metadata:
//   name: ${CLUSTER_NAME}
// spec:
//   credentials:
//     source: Secret
//     secretRef:
//       namespace: ${CROSSPLANE_NAMESPACE}
//       name: ${CLUSTER_NAME}
//       key: ${FILENAME_KUBECONFIG}
// EOF
// ```

`
