package main

import (
	"context"
	"crypto/tls"
	"dagger/packer/internal/dagger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type proxmoxClient struct {
	baseURL    string
	authHeader string
	http       *http.Client
}

func newProxmoxClient(ctx context.Context, pmURL, tokenID, tokenSecret *dagger.Secret) (*proxmoxClient, error) {
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

	return &proxmoxClient{
		baseURL:    strings.TrimRight(base, "/") + "/api2/json",
		authHeader: fmt.Sprintf("PVEAPIToken=%s=%s", id, sec),
		http: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}, nil
}

func (c *proxmoxClient) do(ctx context.Context, method, path string, form url.Values) (string, error) {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", c.authHeader)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("proxmox %s %s: %s: %s", method, path, resp.Status, string(raw))
	}
	return prettyJSON(raw), nil
}

func prettyJSON(raw []byte) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return string(raw)
	}
	return string(out)
}

// Proxmoxoperation performs move/rename/delete on a Proxmox VM or template.
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
) (string, error) {
	if node == "" || vmid == "" {
		return "", fmt.Errorf("node and vmid must be specified")
	}
	c, err := newProxmoxClient(ctx, proxmoxUrl, tokenId, tokenSecret)
	if err != nil {
		return "", err
	}

	switch operation {
	case "move":
		if target == "" {
			return "", fmt.Errorf("target (destination node) must be specified for move")
		}
		return c.do(ctx, http.MethodPost,
			fmt.Sprintf("/nodes/%s/qemu/%s/migrate", node, vmid),
			url.Values{"target": {target}})

	case "rename":
		if target == "" {
			return "", fmt.Errorf("target (new VM name) must be specified for rename")
		}
		// POST (async) on the qemu config endpoint: PVE returns
		// "501 Method 'PUT ...' not implemented" for PUT here.
		return c.do(ctx, http.MethodPost,
			fmt.Sprintf("/nodes/%s/qemu/%s/config", node, vmid),
			url.Values{"name": {target}})

	case "delete":
		return c.do(ctx, http.MethodDelete,
			fmt.Sprintf("/nodes/%s/qemu/%s?purge=1&destroy-unreferenced-disks=1", node, vmid),
			nil)

	default:
		return "", fmt.Errorf("unsupported operation: %s", operation)
	}
}

// CheckProxmoxStorage lists storage pools (cluster-wide or per-node) with usage.
func (m *Packer) CheckProxmoxStorage(
	ctx context.Context,
	proxmoxUrl *dagger.Secret,
	tokenId *dagger.Secret,
	tokenSecret *dagger.Secret,
	// Node to query (e.g. "pve1"); if empty, queries cluster-wide /storage
	// +optional
	node string,
) (string, error) {
	c, err := newProxmoxClient(ctx, proxmoxUrl, tokenId, tokenSecret)
	if err != nil {
		return "", err
	}
	path := "/storage"
	if node != "" {
		path = fmt.Sprintf("/nodes/%s/storage", node)
	}
	return c.do(ctx, http.MethodGet, path, nil)
}

// CheckProxmoxNetworks lists network interfaces/bridges on a node.
func (m *Packer) CheckProxmoxNetworks(
	ctx context.Context,
	proxmoxUrl *dagger.Secret,
	tokenId *dagger.Secret,
	tokenSecret *dagger.Secret,
	// Node to query (e.g. "pve1")
	node string,
) (string, error) {
	if node == "" {
		return "", fmt.Errorf("node must be specified")
	}
	c, err := newProxmoxClient(ctx, proxmoxUrl, tokenId, tokenSecret)
	if err != nil {
		return "", err
	}
	return c.do(ctx, http.MethodGet, fmt.Sprintf("/nodes/%s/network", node), nil)
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
) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name must be specified")
	}
	c, err := newProxmoxClient(ctx, proxmoxUrl, tokenId, tokenSecret)
	if err != nil {
		return "", err
	}

	body, err := c.do(ctx, http.MethodGet, "/cluster/resources?type=vm", nil)
	if err != nil {
		return "", err
	}

	var resp struct {
		Data []struct {
			VMID int    `json:"vmid"`
			Name string `json:"name"`
			Node string `json:"node"`
			Pool string `json:"pool"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return "", fmt.Errorf("parse cluster resources: %w", err)
	}

	var matches []int
	for _, r := range resp.Data {
		if r.Name != name {
			continue
		}
		if pool != "" && r.Pool != pool {
			continue
		}
		if node != "" && r.Node != node {
			continue
		}
		matches = append(matches, r.VMID)
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no VM/template found with name %q (pool=%q, node=%q)", name, pool, node)
	case 1:
		return fmt.Sprintf("%d", matches[0]), nil
	default:
		return "", fmt.Errorf("multiple VMs/templates found with name %q: vmids=%v", name, matches)
	}
}

// ListProxmoxResources returns datacenter-wide resources (nodes, VMs, storage, sdn).
func (m *Packer) ListProxmoxResources(
	ctx context.Context,
	proxmoxUrl *dagger.Secret,
	tokenId *dagger.Secret,
	tokenSecret *dagger.Secret,
	// Resource type filter: vm | storage | node | sdn (empty = all)
	// +optional
	resourceType string,
) (string, error) {
	c, err := newProxmoxClient(ctx, proxmoxUrl, tokenId, tokenSecret)
	if err != nil {
		return "", err
	}
	path := "/cluster/resources"
	if resourceType != "" {
		path += "?type=" + url.QueryEscape(resourceType)
	}
	return c.do(ctx, http.MethodGet, path, nil)
}
