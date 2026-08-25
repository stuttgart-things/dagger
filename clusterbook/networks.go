package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ListNetworks returns all network pools with their stats
func (c *Clusterbook) ListNetworks(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
) (string, error) {
	return c.doGet(ctx, server, "/api/v1/networks")
}

// GetNetworkIPs returns all IPs in a network with their status and cluster info
func (c *Clusterbook) GetNetworkIPs(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
	// network key (e.g. "10.31.103")
	networkKey string,
) (string, error) {
	return c.doGet(ctx, server, fmt.Sprintf("/api/v1/networks/%s/ips", networkKey))
}

// CreateNetwork creates a new network with a flat list of IPs
func (c *Clusterbook) CreateNetwork(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
	// network prefix (e.g. "10.31.103")
	network string,
	// list of last-octet IPs (e.g. ["3","4","5"])
	ips []string,
) (string, error) {
	body := map[string]interface{}{
		"network": network,
		"ips":     ips,
	}
	return c.doPost(ctx, server, "/api/v1/networks", body)
}

// CreateNetworkFromCidr creates a new network from CIDR notation
func (c *Clusterbook) CreateNetworkFromCidr(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
	// CIDR notation (e.g. "10.31.103.0/24")
	cidr string,
	// reserved last-octet IPs to exclude (e.g. ["1"] for gateway)
	// +optional
	reserved []string,
) (string, error) {
	body := map[string]interface{}{
		"cidr": cidr,
	}
	if len(reserved) > 0 {
		body["reserved"] = reserved
	}
	return c.doPost(ctx, server, "/api/v1/networks/cidr", body)
}

// DeleteNetwork deletes a network pool
func (c *Clusterbook) DeleteNetwork(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
	// network key (e.g. "10.31.103")
	networkKey string,
) (string, error) {
	return c.doDelete(ctx, server, fmt.Sprintf("/api/v1/networks/%s", networkKey))
}

func (c *Clusterbook) doGet(ctx context.Context, server, path string) (string, error) {
	url := baseURL(server) + path

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, body)
	}

	return string(body), nil
}

func (c *Clusterbook) doPost(ctx context.Context, server, path string, payload interface{}) (string, error) {
	url := baseURL(server) + path

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, body)
	}

	return string(body), nil
}

func (c *Clusterbook) doDelete(ctx context.Context, server, path string) (string, error) {
	url := baseURL(server) + path

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, body)
	}

	return string(body), nil
}

// baseURL turns a server argument into a URL prefix.
//
// A bare "host:port" keeps the historical http:// default, so every existing
// caller behaves exactly as before. Anything already carrying a scheme is used
// verbatim -- the deployed clusterbook sits behind a Gateway on https, and
// hardcoding http:// there produced a connection error that read like the
// service was down.
func baseURL(server string) string {
	server = strings.TrimSuffix(server, "/")
	if strings.HasPrefix(server, "http://") || strings.HasPrefix(server, "https://") {
		return server
	}
	return "http://" + server
}

// doGetAllowMissing is doGet that reports a 404 instead of turning it into an
// error, for endpoints where "not there yet" is a normal answer.
func (c *Clusterbook) doGetAllowMissing(ctx context.Context, server, path string) (body string, found bool, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL(server)+path, nil)
	if err != nil {
		return "", false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, raw)
	}

	return string(raw), true, nil
}
