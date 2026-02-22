package zerotier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const apiBase = "http://localhost:9993"

// known authtoken locations across macOS and Linux
var tokenPaths = []string{
	"/Library/Application Support/ZeroTier/One/authtoken.secret",
	"/var/lib/zerotier-one/authtoken.secret",
}

// Client wraps the local ZeroTier HTTP API on port 9993.
type Client struct {
	token string
	http  *http.Client
}

// NewClient reads the authtoken and returns a ready API client.
// Returns an error if the authtoken cannot be read — likely means
// ZeroTier is not running or daemon lacks root.
func NewClient() (*Client, error) {
	token, err := readToken()
	if err != nil {
		return nil, err
	}
	return &Client{
		token: token,
		http:  &http.Client{Timeout: 3 * time.Second},
	}, nil
}

func readToken() (string, error) {
	for _, p := range tokenPaths {
		data, err := os.ReadFile(p)
		if err == nil {
			return strings.TrimSpace(string(data)), nil
		}
	}
	return "", fmt.Errorf("authtoken not found (tried %v)", tokenPaths)
}

func (c *Client) post(path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", apiBase+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("X-ZT1-AUTH", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("ZeroTier API %s returned %d", path, resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) del(path string) error {
	req, err := http.NewRequest("DELETE", apiBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-ZT1-AUTH", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("ZeroTier API %s returned %d", path, resp.StatusCode)
	}
	return nil
}

func (c *Client) get(path string, out any) error {
	req, err := http.NewRequest("GET", apiBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-ZT1-AUTH", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("ZeroTier API %s returned %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// ztStatus mirrors the relevant fields from GET /status
type ztStatus struct {
	Address string `json:"address"`
	Version string `json:"version"`
	Online  bool   `json:"online"`
}

// NodeStatus returns the local node ID and ZeroTier version.
func (c *Client) NodeStatus() (nodeID, version string, err error) {
	var s ztStatus
	if err = c.get("/status", &s); err != nil {
		return
	}
	return s.Address, s.Version, nil
}

// ztNetwork mirrors the relevant fields from GET /network/<nwid>
type ztNetwork struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Status            string   `json:"status"`
	AssignedAddresses []string `json:"assignedAddresses"`
	AllowDNS          bool     `json:"allowDNS"`
	AllowManaged      bool     `json:"allowManaged"`
	AllowGlobal       bool     `json:"allowGlobal"`
	AllowDefault      bool     `json:"allowDefault"`
	DNS               struct {
		Domain  string   `json:"domain"`
		Servers []string `json:"servers"`
	} `json:"dns"`
}

// JoinNetwork tells ZeroTier to join a network.
func (c *Client) JoinNetwork(networkID string) error {
	return c.post("/network/"+networkID, map[string]any{}, nil)
}

// LeaveNetwork tells ZeroTier to leave a network.
func (c *Client) LeaveNetwork(networkID string) error {
	return c.del("/network/" + networkID)
}

// SetNetworkFlags updates the allow* flags on a joined network.
func (c *Client) SetNetworkFlags(networkID string, flags map[string]bool) error {
	return c.post("/network/"+networkID, flags, nil)
}

// NetworkState returns the full state of a network the node is joined to.
func (c *Client) NetworkState(networkID string) (*NetworkState, error) {
	var n ztNetwork
	if err := c.get("/network/"+networkID, &n); err != nil {
		return nil, err
	}

	ns := &NetworkState{
		ID:           n.ID,
		Name:         n.Name,
		Status:       n.Status,
		Authorized:   n.Status == "OK",
		AllowDNS:     n.AllowDNS,
		AllowManaged: n.AllowManaged,
		AllowGlobal:  n.AllowGlobal,
		AllowDefault: n.AllowDefault,
		Mode:         DeriveMode(n.AllowManaged, n.AllowDNS, n.AllowGlobal, n.AllowDefault),
		DNSDomain:    n.DNS.Domain,
	}

	if len(n.DNS.Servers) > 0 {
		ns.DNSServer = n.DNS.Servers[0]
	}

	// Addresses come as CIDR (e.g. "10.147.20.42/24") — strip the prefix
	if len(n.AssignedAddresses) > 0 {
		ip := n.AssignedAddresses[0]
		if i := strings.Index(ip, "/"); i >= 0 {
			ip = ip[:i]
		}
		ns.AssignedIP = ip
	}

	return ns, nil
}
