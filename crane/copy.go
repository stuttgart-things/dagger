package main

import (
	"context"
	"dagger/crane/internal/dagger"
	"fmt"
	"strings"
)

// Copy copies an image between registries with authentication
// +call
func (m *Crane) Copy(
	ctx context.Context,
	source string,
	target string,
	// +optional
	sourceRegistry string,
	// +optional
	sourceUsername string,
	// +optional
	sourcePassword *dagger.Secret,
	// +optional
	targetRegistry string,
	// +optional
	targetUsername string,
	// +optional
	targetPassword *dagger.Secret,
	// +optional
	// +flag
	// +default=false
	insecure bool,
	// +optional
	// +flag
	// +default="linux/amd64"
	platform string,
	// +optional
	dockerConfigSecret *dagger.Secret, // NEW: Docker config.json secret
) (string, error) {
	if platform == "" {
		platform = "linux/amd64"
	}

	if sourceRegistry == "" {
		sourceRegistry = extractRegistry(source)
	}
	if targetRegistry == "" {
		targetRegistry = extractRegistry(target)
	}

	var sourceAuth, targetAuth *RegistryAuth

	if sourceRegistry != "" && sourceUsername != "" && sourcePassword != nil {
		sourceAuth = &RegistryAuth{
			URL:      sourceRegistry,
			Username: sourceUsername,
			Password: sourcePassword,
		}
	}

	if targetRegistry != "" && targetUsername != "" && targetPassword != nil {
		targetAuth = &RegistryAuth{
			URL:      targetRegistry,
			Username: targetUsername,
			Password: targetPassword,
		}
	}

	return m.copyImage(
		ctx,
		source,
		target,
		sourceAuth,
		targetAuth,
		insecure,
		platform,
		dockerConfigSecret, // Pass to copyImage
	)
}

func extractRegistry(imageRef string) string {
	parts := strings.Split(imageRef, "/")
	if len(parts) > 1 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":")) {
		return parts[0]
	}
	return ""
}

func (m *Crane) copyImage(
	ctx context.Context,
	source string,
	target string,
	sourceAuth *RegistryAuth,
	targetAuth *RegistryAuth,
	insecure bool,
	platform string,
	dockerConfigSecret *dagger.Secret, // NEW: Docker config secret
) (string, error) {
	ctr := m.container(insecure)

	// MOUNT DOCKER CONFIG SECRET IF PROVIDED
	if dockerConfigSecret != nil {
		// Set DOCKER_CONFIG environment variable to custom path
		configDir := "/.docker"
		configPath := configDir + "/config.json"

		ctr = ctr.
			WithEnvVariable("DOCKER_CONFIG", configDir).
			WithMountedSecret(configPath, dockerConfigSecret)
	} else {
		// REGULAR AUTHENTICATION IF NO CONFIG SECRET
		if sourceAuth != nil {
			ctr = authenticate(ctr, sourceAuth, insecure)
		}
		if targetAuth != nil {
			ctr = authenticate(ctr, targetAuth, insecure)
		}
	}

	cmd := []string{"crane", "copy", "--platform", platform}
	if insecure {
		cmd = append(cmd, "--insecure")
	}
	cmd = append(cmd, source, target)

	fmt.Println("Executing command:", strings.Join(cmd, " "))

	result := ctr.WithExec(cmd)
	out, err := result.Stdout(ctx)
	if err != nil {
		stderr, _ := result.Stderr(ctx)
		return "", fmt.Errorf("copy failed: %w\nStdout: %s\nStderr: %s", err, out, stderr)
	}

	return out, nil
}
