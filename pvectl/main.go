// Pvectl module: LabUL Proxmox template work in Dagger pipelines
//
// Runs pvectl (https://github.com/stuttgart-things/pvectl) from its
// container image in programmatic mode (-o json) and returns the JSON it
// prints. Typical use: after a Packer build, check the new template and
// promote it (move to ul-pve11, FS9100 copy, repins).
//
// Rules that carry over from pvectl:
//   - VM descriptions (notes) are never returned; in LabUL they hold
//     plaintext passwords.
//   - TLS is always verified: pass the cluster CA (caCert) or the pinned
//     SHA-256 of the API certificate (fingerprint). There is no insecure
//     switch.
//   - template-promote only shows the plan (--dry-run) unless apply is set.
//
// Credentials are Dagger secrets: an API token (tokenId + tokenSecret,
// preferred) or user + password.

package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"dagger/pvectl/internal/dagger"
)

// defaultImage is the pvectl image used when image is empty. Keep it in
// step with the +default of New, which has to be a literal.
const defaultImage = "ghcr.io/stuttgart-things/pvectl:v0.6.1"

// knownHostsPath is where a given fingerprint is pinned in the container.
const knownHostsPath = "/tmp/pvectl/known_hosts"

// Pvectl runs pvectl against one Proxmox VE cluster
type Pvectl struct {
	// connection, set by New; private: secrets are not exposed as fields
	// +private
	Image string
	// +private
	URL string
	// +private
	TokenID *dagger.Secret
	// +private
	TokenSecret *dagger.Secret
	// +private
	User *dagger.Secret
	// +private
	Password *dagger.Secret
	// +private
	CaCert *dagger.Secret
	// +private
	Fingerprint string
}

func New(
	// Proxmox API URL, e.g. https://ul-pve01.labul.sva.de:8006/api2/json
	url string,
	// API token ID: user@realm!tokenname
	// +optional
	tokenId *dagger.Secret,
	// API token secret
	// +optional
	tokenSecret *dagger.Secret,
	// User (user@realm), instead of a token
	// +optional
	user *dagger.Secret,
	// Password, instead of a token
	// +optional
	password *dagger.Secret,
	// Cluster CA (/etc/pve/pve-root-ca.pem), PEM
	// +optional
	caCert *dagger.Secret,
	// SHA-256 of the API certificate (AB:CD:...), pinned for this run
	// +optional
	fingerprint string,
	// pvectl image
	// +optional
	// +default="ghcr.io/stuttgart-things/pvectl:v0.6.1"
	image string,
) *Pvectl {
	return &Pvectl{
		Image: image, URL: url,
		TokenID: tokenId, TokenSecret: tokenSecret,
		User: user, Password: password,
		CaCert: caCert, Fingerprint: fingerprint,
	}
}

// Vms lists VMs or templates by name substring or VMID (JSON array)
func (m *Pvectl) Vms(
	ctx context.Context,
	// name substring or VMID
	// +optional
	filter string,
	// list templates instead of VMs
	// +optional
	templates bool,
	// list VMs and templates
	// +optional
	all bool,
) (string, error) {
	args := []string{"vms"}
	if filter != "" {
		args = append(args, filter)
	}
	if templates {
		args = append(args, "--templates")
	}
	if all {
		args = append(args, "--all")
	}
	return m.run(ctx, args, true)
}

// Config shows a VM or template config (without description) and its
// checks; fails with the JSON when a check fails
func (m *Pvectl) Config(
	ctx context.Context,
	// VMID
	vmid int,
) (string, error) {
	return m.run(ctx, []string{"config", strconv.Itoa(vmid)}, true)
}

// Tasks lists recent Proxmox tasks (JSON array)
func (m *Pvectl) Tasks(
	ctx context.Context,
	// only tasks of this node
	// +optional
	node string,
	// only tasks of this VMID
	// +optional
	vmid int,
	// only failed tasks
	// +optional
	errors bool,
	// only tasks started within this duration, e.g. 96h
	// +optional
	since string,
) (string, error) {
	args := []string{"tasks"}
	if node != "" {
		args = append(args, "--node", node)
	}
	if vmid != 0 {
		args = append(args, "--vmid", strconv.Itoa(vmid))
	}
	if errors {
		args = append(args, "--errors")
	}
	if since != "" {
		args = append(args, "--since", since)
	}
	return m.run(ctx, args, true)
}

// Tls shows the API certificate and how this run trusts it (no login)
func (m *Pvectl) Tls(ctx context.Context) (string, error) {
	return m.run(ctx, []string{"tls"}, false)
}

// TemplatePromote promotes a fresh Packer template: checks, migrate to
// the target node, FS9100 copy, repins. Without apply it only returns the
// plan (--dry-run) and changes nothing.
func (m *Pvectl) TemplatePromote(
	ctx context.Context,
	// template VMID
	vmid int,
	// migrate the template to this node
	// +optional
	// +default="ul-pve11"
	to string,
	// make the full FS9100 copy for Telmate Terraform (--fs-copy=false: none)
	// +optional
	// +default=true
	fsCopy bool,
	// promote even with warnings (a newer template exists, or the source is still pinned)
	// +optional
	allowOld bool,
	// stuttgart-things checkout to read the current repin values from (read only)
	// +optional
	repo *dagger.Directory,
	// carry out the plan (--yes). Default: --dry-run, nothing is written
	// +optional
	apply bool,
) (string, error) {
	args := []string{"template", "promote", strconv.Itoa(vmid), "--to", to}
	if fsCopy {
		args = append(args, "--fs9100-copy")
	} else {
		args = append(args, "--no-fs9100-copy")
	}
	if allowOld {
		args = append(args, "--allow-old")
	}
	if apply {
		args = append(args, "--yes")
	} else {
		args = append(args, "--dry-run")
	}
	ctr, err := m.container(true)
	if err != nil {
		return "", err
	}
	if repo != nil {
		ctr = ctr.WithMountedDirectory("/repo", repo)
		args = append(args, "--repo", "/repo")
	}
	return exec(ctx, ctr, args)
}

// container is the pvectl image with the environment for one run.
// login: credentials are required (tls needs only the URL).
func (m *Pvectl) container(login bool) (*dagger.Container, error) {
	if m.URL == "" {
		return nil, fmt.Errorf("url is required")
	}
	image := m.Image
	if image == "" {
		image = defaultImage
	}
	ctr := dag.Container().From(image).
		WithEnvVariable("PVE_API_URL", m.URL).
		WithEnvVariable("PVE_KNOWN_HOSTS", knownHostsPath)
	if login {
		switch {
		case m.TokenID != nil && m.TokenSecret != nil:
			ctr = ctr.WithSecretVariable("PVE_API_TOKEN_ID", m.TokenID).
				WithSecretVariable("PVE_API_TOKEN_SECRET", m.TokenSecret)
		case m.User != nil && m.Password != nil:
			ctr = ctr.WithSecretVariable("PVE_API_USER", m.User).
				WithSecretVariable("PVE_API_PASSWORD", m.Password)
		default:
			return nil, fmt.Errorf("credentials required: token-id + token-secret, or user + password")
		}
	}
	if m.CaCert != nil {
		ctr = ctr.WithSecretVariable("PVE_CA_CERT", m.CaCert)
	}
	if m.Fingerprint != "" {
		hp, err := hostPort(m.URL)
		if err != nil {
			return nil, err
		}
		ctr = ctr.WithNewFile(knownHostsPath, hp+" "+strings.TrimSpace(m.Fingerprint)+"\n")
	}
	return ctr, nil
}

func (m *Pvectl) run(ctx context.Context, args []string, login bool) (string, error) {
	ctr, err := m.container(login)
	if err != nil {
		return "", err
	}
	return exec(ctx, ctr, args)
}

// exec runs pvectl -o json and returns stdout. A non-zero exit fails with
// the exit code and the JSON pvectl printed (an error object, or the
// result of a failed check/task).
func exec(ctx context.Context, ctr *dagger.Container, args []string) (string, error) {
	args = append(args, "-o", "json", "--non-interactive")
	ctr = ctr.WithExec(args, dagger.ContainerWithExecOpts{
		UseEntrypoint: true,
		Expect:        dagger.ReturnTypeAny,
	})
	code, err := ctr.ExitCode(ctx)
	if err != nil {
		return "", err
	}
	out, err := ctr.Stdout(ctx)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("pvectl %s: exit %d\n%s", strings.Join(args[:min(2, len(args))], " "), code, out)
	}
	return out, nil
}

// hostPort is the known-hosts key of an API URL (default port 8006),
// as pvectl writes it.
func hostPort(apiURL string) (string, error) {
	u, err := url.Parse(apiURL)
	if err != nil || u.Hostname() == "" {
		return "", fmt.Errorf("url %q: no host", apiURL)
	}
	port := u.Port()
	if port == "" {
		port = "8006"
	}
	return net.JoinHostPort(u.Hostname(), port), nil
}
