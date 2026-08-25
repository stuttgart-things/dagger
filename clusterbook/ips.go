package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	return c.doPost(ctx, server, fmt.Sprintf("/api/v1/networks/%s/assign", networkKey), body)
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
	return c.doPost(ctx, server, fmt.Sprintf("/api/v1/networks/%s/release", networkKey), body)
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
	return c.doPost(ctx, server, fmt.Sprintf("/api/v1/networks/%s/ips/add", networkKey), body)
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
	return c.doDelete(ctx, server, fmt.Sprintf("/api/v1/networks/%s/ips/%s", networkKey, ip))
}

// ReserveIP allocates the next free IP in a network to a cluster, optionally
// creating the wildcard DNS record for it.
//
// IDEMPOTENT ON PURPOSE. A plain POST /reserve hands out a NEW address every
// time it is called, so a pipeline retry silently leaks an IP and a second DNS
// record -- and neither shows up as a failure anywhere. This first asks whether
// the cluster already holds an address in this network and returns that one if
// so, which makes re-running a bake safe.
//
// The reply always has the same shape, reused telling the two paths apart:
//
//	{"ip":"10.31.103.6","digit":"6","status":"ASSIGNED:DNS","cluster":"c","reused":false}
func (c *Clusterbook) ReserveIP(
	ctx context.Context,
	// clusterbook server address ("host:port" for http, or a full https:// URL)
	server string,
	// network key (e.g. "10.31.103")
	networkKey string,
	// cluster name -- also the label of the wildcard record, *.{cluster}.{zone}
	cluster string,
	// status to record
	// +optional
	// +default="ASSIGNED"
	status string,
	// create the wildcard DNS record for this assignment
	// +optional
	createDns bool,
	// TTL in seconds after which the assignment is reclaimed (0 = permanent)
	// +optional
	leaseDurationSeconds int,
) (string, error) {
	if cluster == "" {
		return "", fmt.Errorf("cluster is required")
	}

	if existing, err := c.lookupClusterIP(ctx, server, networkKey, cluster); err != nil {
		return "", err
	} else if existing != "" {
		return existing, nil
	}

	body := map[string]interface{}{
		"cluster":    cluster,
		"status":     status,
		"create_dns": createDns,
	}
	if leaseDurationSeconds > 0 {
		body["lease_duration_seconds"] = leaseDurationSeconds
	}

	raw, err := c.doPost(ctx, server, fmt.Sprintf("/api/v1/networks/%s/reserve", networkKey), body)
	if err != nil {
		return "", err
	}

	var reserved struct {
		IP      string `json:"ip"`
		Digit   string `json:"digit"`
		Status  string `json:"status"`
		Cluster string `json:"cluster"`
	}
	if err := json.Unmarshal([]byte(raw), &reserved); err != nil {
		return "", fmt.Errorf("reserve returned unparsable body %q: %w", raw, err)
	}

	// Rebuilt rather than passed through: the reserve endpoint also returns an
	// "ips" array that the reuse path has no equivalent for, and a reply whose
	// keys depend on which branch ran is exactly the kind of difference a
	// caller only discovers on the rare path.
	return marshalReservation(reserved.IP, reserved.Digit, reserved.Status, cluster, false)
}

// lookupClusterIP returns a reserve-shaped reply if the cluster already holds
// an IP in networkKey, or "" if it holds none.
func (c *Clusterbook) lookupClusterIP(ctx context.Context, server, networkKey, cluster string) (string, error) {
	raw, found, err := c.doGetAllowMissing(ctx, server, "/api/v1/clusters/"+cluster)
	if err != nil || !found {
		return "", err
	}

	var info struct {
		IPs []struct {
			Network string `json:"network"`
			IP      string `json:"ip"`
			Status  string `json:"status"`
		} `json:"ips"`
	}
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		return "", fmt.Errorf("cluster info returned unparsable body %q: %w", raw, err)
	}

	for _, ip := range info.IPs {
		if ip.Network != networkKey {
			continue
		}
		digit := ip.IP
		if i := strings.LastIndex(digit, "."); i >= 0 {
			digit = digit[i+1:]
		}
		return marshalReservation(ip.IP, digit, ip.Status, cluster, true)
	}

	return "", nil
}

// marshalReservation is the single reply shape ReserveIP returns, whichever
// path produced it.
func marshalReservation(ip, digit, status, cluster string, reused bool) (string, error) {
	out, err := json.Marshal(map[string]interface{}{
		"ip":      ip,
		"digit":   digit,
		"status":  status,
		"cluster": cluster,
		"reused":  reused,
	})
	if err != nil {
		return "", err
	}
	return string(out), nil
}
