// A generated module for Docker functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/docker/internal/dagger"
	"fmt"
	"sync"
)

type Docker struct {
	// +private
	BaseHadolintContainer *dagger.Container
	BaseTrivyContainer    *dagger.Container

	// +private
	BuildContainer *dagger.Container
}

func New(
	// base hadolint container
	// It need contain hadolint
	// +optional
	baseHadolintContainer *dagger.Container,

	// base hadolint container
	// It need contain hadolint
	// +optional
	baseTrivyContainer *dagger.Container,

	// The external build of container
	// Usefull when need build args
	// +optional
	buildContainer *dagger.Container,
) *Docker {
	image := &Docker{
		BuildContainer: buildContainer,
	}

	if baseHadolintContainer != nil {
		image.BaseHadolintContainer = baseHadolintContainer
	} else {
		image.BaseHadolintContainer = image.GetBaseHadolintContainer()
	}

	if baseTrivyContainer != nil {
		image.BaseTrivyContainer = baseTrivyContainer
	} else {
		image.BaseTrivyContainer = image.GetTrivyContainer()
	}

	return image
}

// GetBaseHadolintContainer return the default image for hadolint
func (m *Docker) GetBaseHadolintContainer() *dagger.Container {
	return dag.Container().
		From("ghcr.io/hadolint/hadolint:2.12.0")
}

// GetBaseHadolintContainer return the default image for hadolint
func (m *Docker) GetTrivyContainer() *dagger.Container {
	return dag.Container().
		From("aquasec/trivy:0.60.0")
}

func (m *Docker) BuildAndPush(
	ctx context.Context,
	// The source directory
	source *dagger.Directory,
	// The repository name
	repositoryName string,
	// The version
	version string,
	// The registry username
	// +optional
	withRegistryUsername *dagger.Secret,
	// The registry password
	// +optional
	withRegistryPassword *dagger.Secret,
	// The registry URL
	registryUrl string,
	// The Dockerfile path
	// +optional
	// +default="Dockerfile"
	dockerfile string,
	// Set extra directories
	// +optional
	withDirectories []*dagger.Directory,
) (string, error) {
	// Create a WaitGroup to wait for linting, building, and Trivy scan to complete
	var wg sync.WaitGroup
	wg.Add(3) // Wait for 3 goroutines (linting, building, and Trivy scan)

	// Channel to collect errors from goroutines
	errChan := make(chan error, 3) // Buffer for 3 errors

	// Channel to collect linting results
	lintResultChan := make(chan string, 1)

	// Channel to collect Trivy scan results
	trivyResultChan := make(chan string, 1)

	// Step 1: Run linting concurrently
	go func() {
		defer wg.Done()
		lintOutput, err := m.Lint(ctx, source, dockerfile, "error") // Use default threshold
		if err != nil {
			errChan <- fmt.Errorf("linting failed: %w", err)
			lintResultChan <- lintOutput // Send linting output even if there's an error
			return
		}
		lintResultChan <- lintOutput // Send linting output
		errChan <- nil
	}()

	// Step 2: Run building concurrently
	var buildErr error
	go func() {
		defer wg.Done()
		// Debug: Log the source directory
		fmt.Println("Source Directory:", source)

		// Call the Build function
		builtImage := m.Build(source, dockerfile, withDirectories)
		if builtImage == nil {
			buildErr = fmt.Errorf("build failed: builtImage is nil")
			errChan <- buildErr
			return
		}

		// Debug: Log the built image
		fmt.Println("Built Image:", builtImage)

		// Push the image to the registry
		_, err := builtImage.Push(ctx, repositoryName, version, withRegistryUsername, withRegistryPassword, registryUrl)
		if err != nil {
			buildErr = fmt.Errorf("push failed: %w", err)
			errChan <- buildErr
			return
		}
		errChan <- nil
	}()

	// Step 3: Run Trivy scan on the built image
	go func() {
		defer wg.Done()
		// Wait for the build to complete and check for errors
		if buildErr != nil {
			errChan <- fmt.Errorf("Trivy scan skipped due to build failure: %w", buildErr)
			trivyResultChan <- "Trivy scan skipped due to build failure"
			return
		}

		// Construct the fully qualified image reference
		imageRef := fmt.Sprintf("%s/%s:%s", registryUrl, repositoryName, version)

		// Run Trivy scan on the image reference with registry authentication
		trivyOutput, err := m.TrivyScan(ctx, imageRef, withRegistryUsername, withRegistryPassword)
		if err != nil {
			errChan <- fmt.Errorf("Trivy scan failed: %w", err)
			trivyResultChan <- trivyOutput // Send Trivy output even if there's an error
			return
		}
		trivyResultChan <- trivyOutput // Send Trivy output
		errChan <- nil
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Collect linting results
	lintOutput := <-lintResultChan

	// Collect Trivy scan results
	trivyOutput := <-trivyResultChan

	// Collect errors from the channel
	close(errChan)
	var errs []error
	for err := range errChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Output linting results
	fmt.Println("Linting Results:")
	fmt.Println(lintOutput)

	// Output Trivy scan results
	fmt.Println("Trivy Scan Results:")
	fmt.Println(trivyOutput)

	// If there are any errors, return them along with the linting and Trivy results
	if len(errs) > 0 {
		return fmt.Sprintf("Linting Results:\n%s\nTrivy Scan Results:\n%s", lintOutput, trivyOutput), fmt.Errorf("errors occurred: %v", errs)
	}

	// Return success along with the linting and Trivy results
	return fmt.Sprintf("Successfully built and pushed %s/%s:%s\nLinting Results:\n%s\nTrivy Scan Results:\n%s", registryUrl, repositoryName, version, lintOutput, trivyOutput), nil
}
