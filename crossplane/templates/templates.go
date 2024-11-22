/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package templates

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

var Claim = `apiVersion: {{ .apiGroup }}/{{ .claimApiVersion }}
kind: {{ .claimKind }}
metadata:
  name: {{ .claimName }}
  namespace: {{ .namespace }}
spec:
`

var Composition = `apiVersion: {{ .compositionApiVersion }}
kind: Composition
metadata:
  labels:
    crossplane.io/xrd: {{ .kindLowerX }}.{{ .apiGroup }}
  name: {{ .kindLower }}
spec:
  compositeTypeRef:
    apiVersion: {{ .apiGroup }}/{{ .claimApiVersion }}
    kind: {{ .kind }}
  mode: Pipeline
  pipeline:
`

var Definition = `apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: {{ .plural }}.{{ .apiGroup }}
spec:
  connectionSecretKeys:
    - kubeconfig
  group: {{ .apiGroup }}
  names:
    kind: {{ .kind }}
    plural: {{ .plural }}
  claimNames:
    kind: {{ .claimKind }}
    plural: {{ .claimPlural }}
`

var Configuration = `apiVersion: meta.pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: {{ .kind }}
  annotations:
    meta.crossplane.io/maintainer: {{ .maintainer }}
    meta.crossplane.io/source: {{ .source }}
    meta.crossplane.io/license: {{ .license }}
    meta.crossplane.io/description: |
      deploys {{ .kind }} w/ crossplane
    meta.crossplane.io/readme: |
      deploys {{ .kind }} w/ crossplane
spec:
  crossplane:
    version: ">={{ .crossplaneVersion }}"
  dependsOn:
    - provider: xpkg.upbound.io/crossplane-contrib/provider-helm
      version: ">=v0.19.0"
    - provider: xpkg.upbound.io/crossplane-contrib/provider-kubernetes
      version: ">=v0.14.1"
`

var Readme = `# CROSSPLANE CLAIM
`
