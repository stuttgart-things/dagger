package main

import (
	"context"
	"strings"

	"dagger/gitlab/internal/dagger"
)

// ClonePrivateRepo clones a private repository and returns a Dagger Directory

// Clone clones a git repo using a container and returns the Directory
func (g *Gitlab) CloneWithToken(
	ctx context.Context,
	repoURL string,
	token dagger.Secret,
	branch string,
) (*dagger.Directory, error) {
	// The token stays a Secret: it is only expanded inside the shell, so it is
	// neither an op argument (printed by --progress plain) nor left in the
	// cloned repo's .git/config (#318).
	repoPath := strings.TrimPrefix(repoURL, "https://")

	container := dag.Container().From("alpine/git").
		WithSecretVariable("GITLAB_TOKEN", &token).
		WithEnvVariable("REPO_PATH", repoPath).
		WithEnvVariable("BRANCH", branch).
		WithExec([]string{"sh", "-c",
			`git clone --branch "$BRANCH" "https://oauth2:${GITLAB_TOKEN}@${REPO_PATH}" /repo && ` +
				`git -C /repo remote set-url origin "https://${REPO_PATH}"`,
		})

	// Return the directory where the repo is cloned
	return container.Directory("/repo"), nil
}
