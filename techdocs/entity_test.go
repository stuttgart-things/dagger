package main

import "testing"

func TestEntityRef(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name: "single component, default namespace",
			content: `apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: flux
  annotations:
    backstage.io/techdocs-ref: dir:.
`,
			want: "default/component/flux",
		},
		{
			name: "multi-document, only one with docs",
			content: `apiVersion: backstage.io/v1alpha1
kind: System
metadata:
  name: homerun2
---
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: homerun2-k8s-pitcher
  namespace: Platform
  annotations:
    backstage.io/techdocs-ref: url:https://github.com/stuttgart-things/homerun2-k8s-pitcher/tree/main
`,
			want: "platform/component/homerun2-k8s-pitcher",
		},
		{
			name: "no techdocs-ref",
			content: `kind: Component
metadata:
  name: x
`,
			wantErr: true,
		},
		{
			name: "two entities with techdocs-ref",
			content: `kind: Component
metadata:
  name: a
  annotations: {backstage.io/techdocs-ref: dir:.}
---
kind: Component
metadata:
  name: b
  annotations: {backstage.io/techdocs-ref: dir:.}
`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := entityRef(tt.content)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
