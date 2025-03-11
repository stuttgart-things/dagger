package main

import (
	"context"
	"dagger/docker/internal/dagger"

	"github.com/creasty/defaults"
	"github.com/disaster37/dagger-library-go/lib/helper"
	"github.com/gookit/validate"
)

type LintOption struct {
	Source     *dagger.Directory `validate:"required"`
	Dockerfile string            `default:"Dockerfile"`
	Threashold string            `default:"error"`
}

// Lint permit to lint dockerfile image
func (m *Docker) Lint(
	ctx context.Context,

	// the source directory
	source *dagger.Directory,

	// The dockerfile path
	// +optional
	dockerfile string,

	// The failure threshold
	// +optional
	threshold string,
) (string, error) {
	option := &LintOption{
		Source:     source,
		Dockerfile: dockerfile,
	}

	if err := defaults.Set(option); err != nil {
		panic(err)
	}

	if err := validate.Struct(option).ValidateErr(); err != nil {
		panic(err)
	}

	return m.BaseHadolintContainer.
		WithDirectory("/project", option.Source).
		WithWorkdir("/project").
		WithExec(helper.ForgeCommandf("/bin/hadolint --failure-threshold %s %s", option.Threashold, option.Dockerfile)).
		Stdout(ctx)
}
