package main

import (
	"context"
	"dagger/docker/internal/dagger"

	"github.com/creasty/defaults"
	"github.com/gookit/validate"
)

type LintOption struct {
	src        *dagger.Directory `validate:"required"`
	Dockerfile string            `default:"Dockerfile"`
	Threashold string            `default:"error"`
}

// Lint permit to lint dockerfile image
func (m *Docker) Lint(
	ctx context.Context,
	// the src directory
	src *dagger.Directory,
	// The dockerfile path
	// +optional
	dockerfile string,
	// The failure threshold
	// +optional
	threshold string,
) (string, error) {
	option := &LintOption{
		src:        src,
		Dockerfile: dockerfile,
	}

	if err := defaults.Set(option); err != nil {
		panic(err)
	}

	if err := validate.Struct(option).ValidateErr(); err != nil {
		panic(err)
	}

	return m.BaseHadolintContainer.
		WithDirectory("/project", option.src).
		WithWorkdir("/project").
		WithExec([]string{"/bin/hadolint", "--failure-threshold", option.Threashold, option.Dockerfile}).
		Stdout(ctx)
}
