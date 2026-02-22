package socket

import "github.com/Mr-Chance-Productions-GmbH/sixnetd/internal/zerotier"

// Request is sent by clients over the Unix socket.
type Request struct {
	Cmd       string `json:"cmd"`
	NetworkID string `json:"networkId,omitempty"`
	Mode      string `json:"mode,omitempty"` // vpn | lan | exit
}

// Response is returned by the daemon for every request.
// Status responses populate Daemon/NodeID/Network.
// Action responses (Phase 2+) populate OK/Error.
type Response struct {
	// action response fields
	OK    *bool  `json:"ok,omitempty"`
	Error string `json:"error,omitempty"`

	// status response fields
	Daemon  string                 `json:"daemon,omitempty"`
	NodeID  string                 `json:"nodeId,omitempty"`
	Version string                 `json:"version,omitempty"`
	Network *zerotier.NetworkState `json:"network,omitempty"`
}

func errResponse(msg string) Response {
	return Response{Error: msg}
}

func okResponse() Response {
	t := true
	return Response{OK: &t}
}
