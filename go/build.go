package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
)

func (m *Go) build(
	ctx context.Context,
	src *dagger.Directory,
	opts GoBuildOpts, // Use the struct for parameters
) *dagger.Directory {
	// MOUNT CLONED REPOSITORY INTO `GOLANG` IMAGE
	golang := m.
		GetGoLangContainer(opts.GoVersion).
		WithDirectory("/src", src).
		WithWorkdir("/src")

	fmt.Println("DIR", src)

	// DEFINE THE APPLICATION BUILD COMMAND
	path := "build/"
	buildCmd := []string{
		"env",
		"GOOS=" + opts.Os,
		"GOARCH=" + opts.Arch,
		"go",
		"build",
		"-o",
		path + "/" + opts.BinName,
	}

	// Add ldflags if provided
	if opts.Ldflags != "" {
		buildCmd = append(buildCmd, "-ldflags", opts.Ldflags)
	}

	// Add the main Go file to the build command
	buildCmd = append(buildCmd, opts.GoMainFile)

	// Execute the build command
	golang = golang.WithExec(buildCmd)

	// GET REFERENCE TO BUILD OUTPUT DIRECTORY IN CONTAINER
	outputDir := golang.Directory(path)

	return outputDir
}
func (m *Go) Binary(
	ctx context.Context,
	src *dagger.Directory,
	// +optional
	// +default="1.24.2"
	goVersion string,
	// +optional
	// +default="linux"
	os string,
	// +optional
	// +default="amd64"
	arch string,
	// +optional
	// +default="main.go"
	goMainFile string,
	// +optional
	// +default="main"
	binName string,
	// +optional
	ldflags string, // Add ldflags as an optional parameter
) *dagger.Directory {
	// Call the core build function with the struct
	return m.build(ctx, src, GoBuildOpts{
		GoVersion:  goVersion,
		Os:         os,
		Arch:       arch,
		GoMainFile: goMainFile,
		BinName:    binName,
		Ldflags:    ldflags, // Pass ldflags to the build function
	})
}

func (m *Go) Build(
	ctx context.Context,
	src *dagger.Directory,
	opts GoBuildOpts, // Use the struct for parameters
) *dagger.Directory {
	// Call the core build function
	return m.build(ctx, src, opts)
}
