package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
	"strings"
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
		buildCmd = append(buildCmd, "-ldflags", FormatLdflags(opts.Ldflags, opts.Package))
	}

	// Add the main Go file to the build command
	buildCmd = append(buildCmd, opts.GoMainFile)

	// Execute the build command
	golang = golang.WithExec(buildCmd)

	// GET REFERENCE TO BUILD OUTPUT DIRECTORY IN CONTAINER
	outputDir := golang.Directory(path)

	return outputDir
}

func (m *Go) BuildBinary(
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
	ldflags string,
	// +optional
	// +default=""
	packageName string,
) *dagger.Directory {
	// Call the core build function with the struct
	return m.build(ctx, src, GoBuildOpts{
		GoVersion:  goVersion,
		Os:         os,
		Arch:       arch,
		GoMainFile: goMainFile,
		BinName:    binName,
		Ldflags:    ldflags,
		Package:    packageName,
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

func FormatLdflags(ldflags string, pkg string) string {
	var result []string

	// Ensure pkg ends with "/"
	if pkg != "" && !strings.HasSuffix(pkg, "/") {
		pkg += "/"
	}

	parts := strings.Split(ldflags, ";")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			kv := strings.SplitN(trimmed, "=", 2)
			if len(kv) == 2 {
				flag := fmt.Sprintf("-X %s%s=%s", pkg, kv[0], kv[1])
				result = append(result, flag)
			}
		}
	}
	return strings.Join(result, " ")
}
