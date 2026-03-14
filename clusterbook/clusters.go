package main

import (
	"context"
	"fmt"
)

// ListClusters returns all clusters with their IP counts
func (c *Clusterbook) ListClusters(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
) (string, error) {
	return doGet(ctx, server, "/api/v1/clusters")
}

// GetCluster returns all IPs assigned to a specific cluster
func (c *Clusterbook) GetCluster(
	ctx context.Context,
	// clusterbook server address (e.g. "localhost:8080")
	server string,
	// cluster name
	clusterName string,
) (string, error) {
	return doGet(ctx, server, fmt.Sprintf("/api/v1/clusters/%s", clusterName))
}
