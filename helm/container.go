package main

import (
	"dagger/helm/internal/dagger"
	"fmt"
)

func (m *Helm) container() *dagger.Container {
	// Return cached container if available
	if m.cachedContainer != nil {
		return m.cachedContainer
	}

	arch := "linux_amd64"

	// --- BASE IMAGE ---
	if m.BaseImage == "" {
		m.BaseImage = "cgr.dev/chainguard/wolfi-base:latest"
	}

	// Install all required packages in a single apk call for efficiency
	ctr := dag.Container().
		From(m.BaseImage).
		WithExec([]string{"apk", "add", "--no-cache", "wget", "curl", "git", "kubectl"})

	// ======================================================
	// INSTALL HELM (manual: Wolfi does NOT ship Helm 3.x)
	// ======================================================
	helmVersion := "v3.19.2"

	// Removed duplicate apk add call - packages already installed above
	ctr = ctr.
		WithExec([]string{"sh", "-c", "wget https://get.helm.sh/helm-" + helmVersion + "-linux-amd64.tar.gz -O /tmp/helm.tar.gz"}).
		WithExec([]string{"tar", "-xzvf", "/tmp/helm.tar.gz", "-C", "/tmp/"}).
		WithExec([]string{"mv", "/tmp/linux-amd64/helm", "/usr/bin/helm"}).
		WithExec([]string{"chmod", "+x", "/usr/bin/helm"}).
		WithExec([]string{"rm", "-rf", "/tmp/helm.tar.gz", "/tmp/linux-amd64"}) // Cleanup temp files

	// ======================================================
	// INSTALL HELMFILE
	// ======================================================
	helmfileVersion := "1.1.9"
	helmfileBin := "helmfile"
	helmfileTar := fmt.Sprintf("%s_%s_%s.tar.gz", helmfileBin, helmfileVersion, arch)
	helmfileURL := fmt.Sprintf("https://github.com/helmfile/helmfile/releases/download/v%s/%s", helmfileVersion, helmfileTar)
	helmfileBinPath := "/usr/bin/" + helmfileBin

	ctr = ctr.
		WithExec([]string{"wget", "-O", "/tmp/" + helmfileTar, helmfileURL}).
		WithExec([]string{"tar", "-xzf", "/tmp/" + helmfileTar, "-C", "/tmp/"}).
		WithExec([]string{"mv", "/tmp/" + helmfileBin, helmfileBinPath}).
		WithExec([]string{"chmod", "+x", helmfileBinPath}).
		WithExec([]string{"rm", "-f", "/tmp/" + helmfileTar}). // Cleanup temp file
		WithExec([]string{"helmfile", "init", "--force"})

	// ======================================================
	// INSTALL POLARIS
	// ======================================================
	polarisVersion := "10.1.2"
	polarisTar := fmt.Sprintf("polaris_%s.tar.gz", arch)
	polarisURL := fmt.Sprintf("https://github.com/FairwindsOps/polaris/releases/download/%s/polaris_linux_amd64.tar.gz", polarisVersion)
	polarisBinPath := "/usr/bin/polaris"

	ctr = ctr.
		WithExec([]string{"wget", "-O", "/tmp/" + polarisTar, polarisURL}).
		WithExec([]string{"tar", "-xzf", "/tmp/" + polarisTar, "-C", "/tmp/"}).
		WithExec([]string{"mv", "/tmp/polaris", polarisBinPath}).
		WithExec([]string{"chmod", "+x", polarisBinPath}).
		WithExec([]string{"rm", "-f", "/tmp/" + polarisTar}) // Cleanup temp file

	// ======================================================
	// INSTALL VALS
	// ======================================================
	valsVersion := "0.42.4"
	valsBin := "vals"
	valsTar := fmt.Sprintf("%s_%s_%s.tar.gz", valsBin, valsVersion, arch)
	valsURL := fmt.Sprintf("https://github.com/helmfile/vals/releases/download/v%s/%s", valsVersion, valsTar)
	valsBinPath := "/usr/bin/" + valsBin

	ctr = ctr.
		WithExec([]string{"wget", "-O", "/tmp/" + valsTar, valsURL}).
		WithExec([]string{"tar", "-xzf", "/tmp/" + valsTar, "-C", "/tmp/"}).
		WithExec([]string{"mv", "/tmp/" + valsBin, valsBinPath}).
		WithExec([]string{"chmod", "+x", valsBinPath}).
		WithExec([]string{"rm", "-f", "/tmp/" + valsTar}) // Cleanup temp file

	// Cache the container for reuse
	m.cachedContainer = ctr
	return ctr
}
