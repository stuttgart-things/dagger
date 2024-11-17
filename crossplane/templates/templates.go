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
		Destination: "example/claim.yaml",
	},
	{
		Template:    Composition,
		Destination: "apis/composition.yaml",
	},
	{
		Template:    Function,
		Destination: "apis/function.yaml",
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

var Claim = `
apiVersion: resources.stuttgart-things.com/v1alpha1
kind: {{ .kind }}
metadata:
  name: {{ .claimName }}
  namespace: {{ .namespace }}
spec:
`

var Composition = `
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  labels:
    crossplane.io/xrd: xminio.resources.stuttgart-things.com
  name: {{ .kind }}
spec:
  compositeTypeRef:
    apiVersion: resources.stuttgart-things.com/v1alpha1
    kind: XMinio
  mode: Pipeline
  pipeline:
`

var Function = `
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-patch-and-transform
spec:
  package: xpkg.upbound.io/crossplane-contrib/function-patch-and-transform:v0.1.4
`

var Definition = `
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: xminios.resources.stuttgart-things.com
spec:
`

var Configuration = `
apiVersion: meta.pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: minio
  annotations:
    meta.crossplane.io/maintainer: patrick.hermann@sva.de
    meta.crossplane.io/source: github.com/stuttgart-things/stuttgart-things
    meta.crossplane.io/license: Apache-2.0
    meta.crossplane.io/description: |
      deploys minio with crossplane based on the official minio helm chart
    meta.crossplane.io/readme: |
      deploys minio with crossplane based on the official minio helm chart
spec:
  crossplane:
    version: ">=v1.14.1-0"
  dependsOn:
    - provider: xpkg.upbound.io/crossplane-contrib/provider-helm
      version: "v0.19.0"
    - provider: xpkg.upbound.io/crossplane-contrib/provider-kubernetes
      version: "v0.14.1"
`

var Readme = `
# CROSSPLANE CLAIM
`
