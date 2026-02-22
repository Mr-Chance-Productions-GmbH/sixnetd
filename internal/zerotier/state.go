package zerotier

// DaemonStatus represents the ZeroTier installation and runtime state.
type DaemonStatus string

const (
	StatusNotInstalled DaemonStatus = "not_installed"
	StatusNotRunning   DaemonStatus = "not_running"
	StatusRunning      DaemonStatus = "running"
)

// NetworkMode is the active connection mode, derived from ZeroTier flag state.
type NetworkMode string

const (
	ModeDisconnected NetworkMode = "disconnected"
	ModeVPN          NetworkMode = "vpn"  // allowManaged + allowDNS
	ModeLAN          NetworkMode = "lan"  // + allowGlobal
	ModeExit         NetworkMode = "exit" // + allowDefault
)

// NetworkState is the full state of a single ZeroTier network membership.
type NetworkState struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Status       string      `json:"status"` // OK, ACCESS_DENIED, NOT_FOUND, ...
	Authorized   bool        `json:"authorized"`
	Mode         NetworkMode `json:"mode"`
	AssignedIP   string      `json:"assignedIP,omitempty"`
	AllowDNS     bool        `json:"allowDNS"`
	AllowManaged bool        `json:"allowManaged"`
	AllowGlobal  bool        `json:"allowGlobal"`
	AllowDefault bool        `json:"allowDefault"`
	DNSDomain    string      `json:"dnsDomain,omitempty"`
	DNSServer    string      `json:"dnsServer,omitempty"`
}

// State is the full snapshot returned to clients on a status request.
type State struct {
	Daemon  DaemonStatus  `json:"daemon"`
	NodeID  string        `json:"nodeId,omitempty"`
	Version string        `json:"version,omitempty"`
	Network *NetworkState `json:"network,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// DeriveMode returns the connection mode from ZeroTier network flags.
// Mirrors the logic in vpn/zt: up sets allowManaged+allowDNS,
// up lan adds allowGlobal, up exit adds allowDefault.
func DeriveMode(allowManaged, allowDNS, allowGlobal, allowDefault bool) NetworkMode {
	if !allowManaged || !allowDNS {
		return ModeDisconnected
	}
	if allowDefault {
		return ModeExit
	}
	if allowGlobal {
		return ModeLAN
	}
	return ModeVPN
}
