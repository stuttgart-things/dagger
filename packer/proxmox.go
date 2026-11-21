package main

import (
	"context"
	"dagger/packer/internal/dagger"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/stuttgart-things/pvectl/pkg/pve"
)

// The Proxmox functions use the pvectl library
// (github.com/stuttgart-things/pvectl/pkg/pve): API token login, verified
// TLS (cluster CA or a pinned certificate fingerprint), and VM descriptions
// never leave the library. Every function returns {"data": ...} JSON.

// newProxmoxClient builds a pvectl client. TLS needs proxmoxCaCert (the
// cluster CA, /etc/pve/pve-root-ca.pem) or proxmoxFingerprint (the SHA-256
// of the API certificate as the web UI shows it).
func newProxmoxClient(ctx context.Context, pmURL, tokenID, tokenSecret, caCert *dagger.Secret, fingerprint string) (*pve.Client, error) {
	base, err := pmURL.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("read proxmoxUrl: %w", err)
	}
	id, err := tokenID.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("read tokenId: %w", err)
	}
	sec, err := tokenSecret.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("read tokenSecret: %w", err)
	}
	o := pve.Options{URL: base, TokenID: id, TokenSecret: sec}
	switch {
	case caCert != nil:
		if o.CACert, err = caCert.Plaintext(ctx); err != nil {
			return nil, fmt.Errorf("read proxmoxCaCert: %w", err)
		}
	case fingerprint != "":
		if o.KnownHosts, err = pve.PinnedKnownHosts(base, fingerprint); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("TLS: pass --proxmox-ca-cert (cluster CA, /etc/pve/pve-root-ca.pem) " +
			"or --proxmox-fingerprint (SHA-256 of the API certificate: node > System > Certificates)")
	}
	return pve.New(o)
}

// result wraps v as {"data": v}, indented.
func result(v any) (string, error) {
	out, err := json.MarshalIndent(map[string]any{"data": v}, "", "  ")
	return string(out), err
}

func parseVMID(vmid string) (int, error) {
	n, err := strconv.Atoi(vmid)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("vmid %q is not a VMID", vmid)
	}
	return n, nil
}

// Proxmoxoperation performs move/rename/delete on a Proxmox VM or template
// and returns the task UPID.
func (m *Packer) Proxmoxoperation(
	ctx context.Context,
	// Operation: move | rename | delete
	operation string,
	// Source node hosting the VM (e.g. "pve1")
	node string,
	// VMID of the VM/template
	vmid string,
	// For move: target node. For rename: new name. For delete: ignored.
	// +optional
	target string,
	proxmoxUrl *dagger.Secret,
	tokenId *dagger.Secret,
	tokenSecret *dagger.Secret,
	// Cluster CA (PEM) to verify the API certificate
	// +optional
	proxmoxCaCert *dagger.Secret,
	// SHA-256 of the API certificate (AB:CD:...), instead of a CA
	// +optional
	proxmoxFingerprint string,
) (string, error) {
	if node == "" || vmid == "" {
		return "", fmt.Errorf("node and vmid must be specified")
	}
	id, err := parseVMID(vmid)
	if err != nil {
		return "", err
	}
	if (operation == "move" || operation == "rename") && target == "" {
		return "", fmt.Errorf("target must be specified for %s (move: destination node, rename: new name)", operation)
	}
	if operation != "move" && operation != "rename" && operation != "delete" {
		return "", fmt.Errorf("unsupported operation: %s", operation)
	}
	c, err := newProxmoxClient(ctx, proxmoxUrl, tokenId, tokenSecret, proxmoxCaCert, proxmoxFingerprint)
	if err != nil {
		return "", err
	}
	var upid string
	switch operation {
	case "move":
		upid, err = c.Migrate(ctx, pve.MigrateRequest{Node: node, VMID: id, Target: target})
	case "rename":
		upid, err = c.Rename(ctx, node, id, target)
	case "delete":
		upid, err = c.Delete(ctx, node, id)
	}
	if err != nil {
		return "", err
	}
	return result(upid)
}

// CheckProxmoxStorage lists storages with usage: cluster-wide (every node),
// or the VM-disk storages of one node.
func (m *Packer) CheckProxmoxStorage(
	ctx context.Context,
	proxmoxUrl *dagger.Secret,
	tokenId *dagger.Secret,
	tokenSecret *dagger.Secret,
	// Node to query (e.g. "pve1"); if empty, every storage on every node
	// +optional
	node string,
	// Cluster CA (PEM) to verify the API certificate
	// +optional
	proxmoxCaCert *dagger.Secret,
	// SHA-256 of the API certificate (AB:CD:...), instead of a CA
	// +optional
	proxmoxFingerprint string,
) (string, error) {
	c, err := newProxmoxClient(ctx, proxmoxUrl, tokenId, tokenSecret, proxmoxCaCert, proxmoxFingerprint)
	if err != nil {
		return "", err
	}
	if node != "" {
		ss, err := c.Storages(ctx, node)
		if err != nil {
			return "", err
		}
		return result(ss)
	}
	ss, err := c.ClusterStorages(ctx)
	if err != nil {
		return "", err
	}
	return result(ss)
}

// CheckProxmoxNetworks lists network interfaces/bridges on a node.
func (m *Packer) CheckProxmoxNetworks(
	ctx context.Context,
	proxmoxUrl *dagger.Secret,
	tokenId *dagger.Secret,
	tokenSecret *dagger.Secret,
	// Node to query (e.g. "pve1")
	node string,
	// Cluster CA (PEM) to verify the API certificate
	// +optional
	proxmoxCaCert *dagger.Secret,
	// SHA-256 of the API certificate (AB:CD:...), instead of a CA
	// +optional
	proxmoxFingerprint string,
) (string, error) {
	if node == "" {
		return "", fmt.Errorf("node must be specified")
	}
	c, err := newProxmoxClient(ctx, proxmoxUrl, tokenId, tokenSecret, proxmoxCaCert, proxmoxFingerprint)
	if err != nil {
		return "", err
	}
	ns, err := c.Networks(ctx, node)
	if err != nil {
		return "", err
	}
	return result(ns)
}

// GetProxmoxTemplateIdByName looks up a VM/template VMID by its name.
// Errors if zero or multiple matches are found. Optional pool/node narrow the search.
func (m *Packer) GetProxmoxTemplateIdByName(
	ctx context.Context,
	// Name of the VM/template to look up
	name string,
	proxmoxUrl *dagger.Secret,
	tokenId *dagger.Secret,
	tokenSecret *dagger.Secret,
	// Restrict search to this pool
	// +optional
	pool string,
	// Restrict search to this node
	// +optional
	node string,
	// Cluster CA (PEM) to verify the API certificate
	// +optional
	proxmoxCaCert *dagger.Secret,
	// SHA-256 of the API certificate (AB:CD:...), instead of a CA
	// +optional
	proxmoxFingerprint string,
) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name must be specified")
	}
	c, err := newProxmoxClient(ctx, proxmoxUrl, tokenId, tokenSecret, proxmoxCaCert, proxmoxFingerprint)
	if err != nil {
		return "", err
	}
	r, err := c.FindByName(ctx, name, pool, node)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(r.VMID), nil
}

// ListProxmoxResources returns datacenter-wide resources: VMs/templates,
// storages, nodes.
func (m *Packer) ListProxmoxResources(
	ctx context.Context,
	proxmoxUrl *dagger.Secret,
	tokenId *dagger.Secret,
	tokenSecret *dagger.Secret,
	// Resource type filter: vm | storage | node (empty = all three)
	// +optional
	resourceType string,
	// Cluster CA (PEM) to verify the API certificate
	// +optional
	proxmoxCaCert *dagger.Secret,
	// SHA-256 of the API certificate (AB:CD:...), instead of a CA
	// +optional
	proxmoxFingerprint string,
) (string, error) {
	if resourceType != "" && resourceType != "vm" && resourceType != "storage" && resourceType != "node" {
		return "", fmt.Errorf("unsupported resource type %q: vm | storage | node", resourceType)
	}
	c, err := newProxmoxClient(ctx, proxmoxUrl, tokenId, tokenSecret, proxmoxCaCert, proxmoxFingerprint)
	if err != nil {
		return "", err
	}
	all := map[string]any{}
	if resourceType == "" || resourceType == "vm" {
		vms, err := c.Resources(ctx)
		if err != nil {
			return "", err
		}
		all["vm"] = vms
	}
	if resourceType == "" || resourceType == "storage" {
		ss, err := c.ClusterStorages(ctx)
		if err != nil {
			return "", err
		}
		all["storage"] = ss
	}
	if resourceType == "" || resourceType == "node" {
		ns, err := c.Nodes(ctx)
		if err != nil {
			return "", err
		}
		all["node"] = ns
	}
	if resourceType != "" {
		return result(all[resourceType])
	}
	return result(all)
}
