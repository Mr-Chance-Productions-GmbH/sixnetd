package socket

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"

	"github.com/Mr-Chance-Productions-GmbH/sixnetd/internal/dns"
	"github.com/Mr-Chance-Productions-GmbH/sixnetd/internal/zerotier"
)

const SocketPath = "/var/run/sixnetd.sock"

// Server listens on the Unix socket and dispatches JSON requests.
type Server struct {
	listener net.Listener
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start() error {
	os.Remove(SocketPath)

	l, err := net.Listen("unix", SocketPath)
	if err != nil {
		return err
	}
	// 0666 so unprivileged clients (the GUI app) can connect
	os.Chmod(SocketPath, 0666)

	s.listener = l
	go s.accept()
	return nil
}

func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
	os.Remove(SocketPath)
}

func (s *Server) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return // listener closed
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	enc := json.NewEncoder(conn)
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			enc.Encode(errResponse("invalid JSON"))
			continue
		}
		resp := s.dispatch(req)
		if err := enc.Encode(resp); err != nil {
			log.Printf("write error: %v", err)
			return
		}
	}
}

func (s *Server) dispatch(req Request) Response {
	switch req.Cmd {
	case "status":
		return s.handleStatus(req)
	case "join":
		return s.handleJoin(req)
	case "leave":
		return s.handleLeave(req)
	case "connect":
		return s.handleConnect(req)
	case "disconnect":
		return s.handleDisconnect(req)
	default:
		return errResponse("unknown command: " + req.Cmd)
	}
}

func (s *Server) handleJoin(req Request) Response {
	if req.NetworkID == "" {
		return errResponse("networkId required")
	}
	client, err := zerotier.NewClient()
	if err != nil {
		return errResponse("zerotier not running: " + err.Error())
	}
	if err := client.JoinNetwork(req.NetworkID); err != nil {
		return errResponse(err.Error())
	}
	return okResponse()
}

func (s *Server) handleLeave(req Request) Response {
	if req.NetworkID == "" {
		return errResponse("networkId required")
	}
	client, err := zerotier.NewClient()
	if err != nil {
		return errResponse("zerotier not running: " + err.Error())
	}
	// Fetch DNS domain before leaving so we can clean up the resolver file.
	ns, _ := client.NetworkState(req.NetworkID)
	if err := client.LeaveNetwork(req.NetworkID); err != nil {
		return errResponse(err.Error())
	}
	if ns != nil && ns.DNSDomain != "" {
		dns.Remove(ns.DNSDomain)
	}
	return okResponse()
}

func (s *Server) handleConnect(req Request) Response {
	if req.NetworkID == "" {
		return errResponse("networkId required")
	}
	mode := zerotier.NetworkMode(req.Mode)
	if mode == "" {
		mode = zerotier.ModeVPN
	}
	client, err := zerotier.NewClient()
	if err != nil {
		return errResponse("zerotier not running: " + err.Error())
	}
	if err := client.SetNetworkFlags(req.NetworkID, zerotier.ModeFlags(mode)); err != nil {
		return errResponse(err.Error())
	}
	// Write DNS resolver if the controller has pushed DNS config.
	ns, err := client.NetworkState(req.NetworkID)
	if err != nil {
		return errResponse(err.Error())
	}
	if ns.DNSDomain != "" && ns.DNSServer != "" {
		if err := dns.Write(ns.DNSDomain, ns.DNSServer); err != nil {
			return errResponse("DNS resolver: " + err.Error())
		}
	}
	return okResponse()
}

func (s *Server) handleDisconnect(req Request) Response {
	if req.NetworkID == "" {
		return errResponse("networkId required")
	}
	client, err := zerotier.NewClient()
	if err != nil {
		return errResponse("zerotier not running: " + err.Error())
	}
	// Fetch DNS domain before clearing flags.
	ns, _ := client.NetworkState(req.NetworkID)
	if err := client.SetNetworkFlags(req.NetworkID, zerotier.ModeFlags(zerotier.ModeDisconnected)); err != nil {
		return errResponse(err.Error())
	}
	if ns != nil && ns.DNSDomain != "" {
		dns.Remove(ns.DNSDomain)
	}
	return okResponse()
}

func (s *Server) handleStatus(req Request) Response {
	resp := Response{}

	if !zerotier.IsInstalled() {
		resp.Daemon = string(zerotier.StatusNotInstalled)
		return resp
	}

	client, err := zerotier.NewClient()
	if err != nil {
		resp.Daemon = string(zerotier.StatusNotRunning)
		return resp
	}

	nodeID, version, err := client.NodeStatus()
	if err != nil {
		resp.Daemon = string(zerotier.StatusNotRunning)
		return resp
	}

	resp.Daemon = string(zerotier.StatusRunning)
	resp.NodeID = nodeID
	resp.Version = version

	if req.NetworkID != "" {
		ns, err := client.NetworkState(req.NetworkID)
		if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Network = ns
		}
	}

	return resp
}
