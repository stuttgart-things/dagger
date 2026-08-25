# Clusterbook Dagger Module

Thin Dagger wrapper around the [clusterbook](https://github.com/stuttgart-things/clusterbook) HTTP API for managing cluster IP allocations. Use it from CI to query networks, allocate IPs to a cluster, and release them on teardown.

All functions take `--server` and return the raw JSON response from the server.

`--server` accepts either a bare `host:port` (assumed `http://`, the historical
behaviour) or a full URL. Pass `--insecure` **before** the function name when the
server is behind a certificate signed by an internal CA:

```bash
dagger call -m clusterbook --insecure list-networks \
  --server https://clusterbook.infra.sthings-vsphere.labul.sva.de
```

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
| `assign-ip` | Assign a **specific** IP to a cluster (`PENDING` or `ASSIGNED`), optionally creating a DNS record. |
| `reserve-ip` | Allocate the **next free** IP to a cluster, optionally creating its wildcard DNS record. Idempotent. |
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

## reserve-ip

Picks the next free address in a pool and, with `--create-dns`, has clusterbook
write the wildcard `*.{cluster}.{zone}` for it — IPAM and DNS in one call.

```bash
dagger call -m clusterbook --insecure reserve-ip \
  --server https://clusterbook.infra.sthings-vsphere.labul.sva.de \
  --network-key 10.31.104 \
  --cluster my-cluster \
  --create-dns
# → {"cluster":"my-cluster","digit":"15","ip":"10.31.104.15","reused":false,"status":"ASSIGNED:DNS"}
```

**It is idempotent, and that is the point.** The bare `POST /reserve` endpoint
hands out a *new* address on every call, so a pipeline retry silently leaks an
IP and a second DNS record — with nothing anywhere reporting a failure.
`reserve-ip` first asks whether the cluster already holds an address in that
network and returns it if so, marked `"reused": true`. Both paths return the
same five keys, so a caller never has to know which one ran.

Use `assign-ip` instead when the address is already decided; use `reserve-ip`
when clusterbook should decide.
