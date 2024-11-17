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
