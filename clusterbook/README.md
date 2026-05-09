# Clusterbook Dagger Module

Thin Dagger wrapper around the [clusterbook](https://github.com/stuttgart-things/clusterbook) HTTP API for managing cluster IP allocations. Use it from CI to query networks, allocate IPs to a cluster, and release them on teardown.

All functions take `--server <host:port>` (HTTP) and return the raw JSON response from the server.

## Functions

### Networks
| Function | Purpose |
|----------|---------|
| `list-networks` | List all network pools with stats. |
| `get-network-ips` | List IPs in a network with status and assigned cluster. |
| `create-network` | Create a network from a flat list of last-octet IPs. |
| `create-network-from-cidr` | Create a network from CIDR, optionally reserving IPs (e.g. gateway). |
| `delete-network` | Delete a network pool. |

### Clusters
| Function | Purpose |
|----------|---------|
| `list-clusters` | List all clusters with their IP counts. |
| `get-cluster` | List IPs assigned to a specific cluster. |

### IPs
| Function | Purpose |
|----------|---------|
| `assign-ip` | Assign an IP to a cluster (`PENDING` or `ASSIGNED`), optionally creating a DNS record. |
| `release-ip` | Release an IP back to the pool. |
| `add-ips` | Add IPs to an existing network. |
| `delete-ip` | Remove an IP from a network. |

## Quick Start

```bash
export CB=clusterbook.example.com:8080

# Create a /24 network, reserve .1 for gateway
dagger call -m clusterbook create-network-from-cidr \
  --server $CB \
  --cidr 10.31.103.0/24 \
  --reserved 1
```

```bash
# Assign an IP to a cluster (with DNS record)
dagger call -m clusterbook assign-ip \
  --server $CB \
  --network-key 10.31.103 \
  --ip 10.31.103.5 \
  --cluster sthings-app-4 \
  --status ASSIGNED \
  --create-dns
```

```bash
# Release on teardown
dagger call -m clusterbook release-ip \
  --server $CB \
  --network-key 10.31.103 \
  --ip 10.31.103.5
```

```bash
# Inspect state
dagger call -m clusterbook list-networks --server $CB
dagger call -m clusterbook get-network-ips --server $CB --network-key 10.31.103
dagger call -m clusterbook get-cluster --server $CB --cluster-name sthings-app-4
```
