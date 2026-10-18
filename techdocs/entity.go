package main

import (
	"context"
	"dagger/techdocs/internal/dagger"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

const techdocsRefAnnotation = "backstage.io/techdocs-ref"

type catalogEntity struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name        string            `yaml:"name"`
		Namespace   string            `yaml:"namespace"`
		Annotations map[string]string `yaml:"annotations"`
	} `yaml:"metadata"`
}

// Entity returns <namespace>/<kind>/<name> of the entity in catalogFile that
// carries the backstage.io/techdocs-ref annotation -- the path Backstage
// looks up the docs under. The namespace defaults to "default".
func (m *Techdocs) Entity(
	ctx context.Context,
	// Repository directory containing catalogFile
	src *dagger.Directory,
	// Catalog file inside src to read the entity from
	// +optional
	// +default="catalog-info.yaml"
	catalogFile string,
) (string, error) {
	content, err := src.File(catalogFile).Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", catalogFile, err)
	}
	return entityRef(content)
}

// entityRef picks the entity with a techdocs-ref annotation out of a
// (possibly multi-document) catalog file.
func entityRef(content string) (string, error) {
	dec := yaml.NewDecoder(strings.NewReader(content))
	var withDocs []catalogEntity
	for {
		var e catalogEntity
		err := dec.Decode(&e)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("parsing catalog file: %w", err)
		}
		if _, ok := e.Metadata.Annotations[techdocsRefAnnotation]; ok {
			withDocs = append(withDocs, e)
		}
	}

	switch len(withDocs) {
	case 0:
		return "", fmt.Errorf("no entity with a %s annotation; pass --entity", techdocsRefAnnotation)
	case 1:
	default:
		return "", fmt.Errorf("%d entities carry %s; pass --entity to pick one", len(withDocs), techdocsRefAnnotation)
	}

	e := withDocs[0]
	if e.Kind == "" || e.Metadata.Name == "" {
		return "", errors.New("the entity with techdocs-ref has no kind or metadata.name")
	}
	ns := e.Metadata.Namespace
	if ns == "" {
		ns = "default"
	}
	// Backstage looks docs up lower-cased (unless
	// legacyUseCaseSensitiveTripletPaths); techdocs-cli publish lower-cases
	// too, so this only keeps the log line honest.
	return strings.ToLower(fmt.Sprintf("%s/%s/%s", ns, e.Kind, e.Metadata.Name)), nil
}
