package main

import (
	"context"
	"fmt"
)

// AssignIP assigns an IP address to a cluster
func (c *Clusterbook) AssignIP(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
	// network key (e.g. "10.31.103")
	networkKey string,
	// full IP address (e.g. "10.31.103.5")
	ip string,
	// cluster name
	cluster string,
	// status (PENDING or ASSIGNED)
	status string,
	// create a DNS record for this assignment
	// +optional
	createDns bool,
) (string, error) {
	body := map[string]interface{}{
		"ip":         ip,
		"cluster":    cluster,
		"status":     status,
		"create_dns": createDns,
	}
	return doPost(ctx, server, fmt.Sprintf("/api/v1/networks/%s/assign", networkKey), body)
}

// ReleaseIP releases an IP address back to the available pool
func (c *Clusterbook) ReleaseIP(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
	// network key (e.g. "10.31.103")
	networkKey string,
	// full IP address (e.g. "10.31.103.5")
	ip string,
) (string, error) {
	body := map[string]interface{}{
		"ip": ip,
	}
	return doPost(ctx, server, fmt.Sprintf("/api/v1/networks/%s/release", networkKey), body)
}

// AddIPs adds IPs to an existing network
func (c *Clusterbook) AddIPs(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
	// network key (e.g. "10.31.103")
	networkKey string,
	// list of last-octet IPs to add (e.g. ["11","12","13"])
	ips []string,
) (string, error) {
	body := map[string]interface{}{
		"ips": ips,
	}
	return doPost(ctx, server, fmt.Sprintf("/api/v1/networks/%s/ips/add", networkKey), body)
}

// DeleteIP removes an IP from a network
func (c *Clusterbook) DeleteIP(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
	// network key (e.g. "10.31.103")
	networkKey string,
	// last-octet IP to delete (e.g. "5")
	ip string,
) (string, error) {
	return doDelete(ctx, server, fmt.Sprintf("/api/v1/networks/%s/ips/%s", networkKey, ip))
}
