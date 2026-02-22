# sixnetd Roadmap

## Context

sixnetd is the privileged daemon for the sixnet VPN platform. It wraps ZeroTier
operations that require root, and exposes them to unprivileged clients (Swift GUI app,
bash wrapper, future platform clients) over a Unix socket at `/var/run/sixnetd.sock`.

The reference implementation for everything this daemon does is `vpn/zt` in the
sixnet core repo (`~/projects/six.net`). All logic decisions should be validated
against that script.

The immediate goal is to support the Q1 deployment: a small group of real users
connecting to hosted apps (Authentik, OpenProject, Jellyfin) without a terminal.

---

## Phase 1 — Foundation: State Reading

**Goal:** daemon starts, reads ZeroTier state, answers `{"cmd":"status"}` over the socket.

**Milestone:** `echo '{"cmd":"status"}' | nc -U /var/run/sixnetd.sock` returns real
ZeroTier state — node ID, daemon status, network auth state, assigned IP.

### What to build

`internal/zerotier/state.go`
- Structs: `NodeState`, `NetworkState`, `DaemonState`
- `DaemonState.Status`: `not_installed | not_running | running`
- `NetworkState.Status`: mirrors ZeroTier — `OK | ACCESS_DENIED | NOT_FOUND | ...`
- `NetworkState.Mode`: current mode derived from allowGlobal/allowDefault flags: `vpn | lan | exit | disconnected`

`internal/zerotier/client.go` — ZeroTier HTTP API client
- Reads authtoken from known paths (macOS + Linux)
- `GET /status` → node ID, version, online
- `GET /network/<nwid>` → full network state, DNS config, assigned IPs, flags
- Auth token paths:
  - macOS: `/Library/Application Support/ZeroTier/One/authtoken.secret`
  - Linux: `/var/lib/zerotier-one/authtoken.secret`

`internal/zerotier/cli.go` — zerotier-cli wrapper
- `IsInstalled()` — checks known binary paths
- `Info()` — runs `zerotier-cli info`, parses node address + version
- Paths to check: `/usr/local/bin/zerotier-cli`, `/usr/sbin/zerotier-cli`

`internal/socket/protocol.go` — request/response types
```json
Request:  {"cmd":"status","networkId":"31655f6ec3a15f6d"}
Response: {
  "daemon": "running",
  "nodeId": "a1b2c3d4e5",
  "network": {
    "id": "31655f6ec3a15f6d",
    "name": "Q1 Office VPN",
    "status": "OK",
    "mode": "vpn",
    "assignedIP": "10.147.20.42",
    "authorized": true
  }
}
```

`internal/socket/server.go` — Unix socket listener
- Listen on `/var/run/sixnetd.sock`
- One goroutine per client connection
- Newline-delimited JSON (one request → one response for now)
- Handles `status` command

`cmd/sixnetd/main.go` — wire together
- Start socket server
- Background state poller (5-second interval, updates cached state)
- Signal handling (SIGTERM/SIGINT → clean shutdown, remove socket file)

### Reference
`vpn/zt`: `get_network_status()`, `get_zt_dns()`, `get_setting()`, `is_joined()`

---

## Phase 2 — Connect / Disconnect

**Goal:** daemon manages the full VPN lifecycle through all three modes.

**Milestone:** connect and disconnect work via socket commands, DNS resolves after
connect, `/etc/resolver/<domain>` is created and removed correctly.

### What to build

`internal/zerotier/ops.go` — privileged ZeroTier operations
- `Join(networkId)` — `zerotier-cli join <nwid>`
- `Leave(networkId)` — `zerotier-cli leave <nwid>`
- `SetFlags(networkId, mode)` — set allowDNS, allowManaged, allowGlobal, allowDefault
  based on mode:

  | Mode | allowManaged | allowDNS | allowGlobal | allowDefault |
  |------|-------------|----------|-------------|--------------|
  | vpn  | true | true | false | false |
  | lan  | true | true | true  | false |
  | exit | true | true | true  | true  |

- `ClearFlags(networkId)` — set all flags to false (disconnect state)

`internal/dns/resolver.go` — macOS DNS resolver management
- `Setup(domain, serverIP)` — write `/etc/resolver/<domain>` pointing to VPN DNS server
  - DNS config comes from ZeroTier HTTP API: `network.dns.domain` + `network.dns.servers[0]`
- `Remove(domain)` — delete `/etc/resolver/<domain>`
- Linux equivalent: TBD (may need resolvconf or systemd-resolved integration)

Socket protocol additions:
```json
{"cmd":"connect","networkId":"31655f6ec3a15f6d","mode":"vpn|lan|exit"}
{"cmd":"disconnect","networkId":"31655f6ec3a15f6d"}
{"cmd":"join","networkId":"31655f6ec3a15f6d"}
{"cmd":"leave","networkId":"31655f6ec3a15f6d"}
```

Connect sequence (mirrors `zt up` from the bash reference):
1. Verify joined (`zerotier-cli listnetworks | grep <nwid>`)
2. Set flags for requested mode
3. Query ZeroTier HTTP API for DNS config (domain + server IP)
4. Write `/etc/resolver/<domain>`
5. Return updated state

Disconnect sequence (mirrors `zt down`):
1. Clear all flags
2. Remove `/etc/resolver/<domain>`
3. Return updated state

### Reference
`vpn/zt`: `cmd_up()`, `cmd_down()`, `setup_dns_resolver()`, `remove_dns_resolver()`

---

## Phase 3 — State Push

**Goal:** clients receive state updates automatically without polling the daemon.

**Milestone:** Swift app UI updates in real time when connection state changes
(authorized, IP assigned, DNS working) without the app doing its own polling.

### What to build

- Client registry in socket server — track all open connections
- Poller detects state changes (compare new state to previous)
- Push state to all connected clients on change
- Protocol addition: daemon pushes unsolicited `{"event":"state","data":{...}}` messages
- Swift app switches from polling to event-driven updates

Note: phase 1 and 2 can use a simpler poll-from-client model. Phase 3 is a UX
improvement — the Swift app polls every 3-5 seconds in the interim.

---

## Phase 4 — macOS LaunchDaemon Installer

**Goal:** Swift app can install sixnetd as a persistent system service on first launch.

**Milestone:** app installs daemon with one admin dialog, daemon survives reboot,
app reconnects to socket automatically on subsequent launches.

### What to build

`internal/install/` — install/uninstall logic
- Copy daemon binary to `/Library/Application Support/Sixnet/sixnetd`
- Write LaunchDaemon plist to `/Library/LaunchDaemons/de.mcp.sixnet.daemon.plist`
- `launchctl bootstrap system` to start it
- Version check: compare installed binary version against bundled version
- Upgrade path: stop old daemon, replace binary, restart

`cmd/sixnetd/main.go` additions:
- `sixnetd --version` — prints version string (for installer version check)
- `sixnetd --install` — self-install (called via NSAppleScript from Swift app)
- `sixnetd --uninstall` — clean removal

LaunchDaemon plist:
```xml
<key>Label</key><string>de.mcp.sixnet.daemon</string>
<key>ProgramArguments</key><array>
  <string>/Library/Application Support/Sixnet/sixnetd</string>
</array>
<key>RunAtLoad</key><true/>
<key>KeepAlive</key><true/>
```

Swift app first-launch flow:
1. Check if `/var/run/sixnetd.sock` exists and responds
2. If not: extract bundled sixnetd, run `sixnetd --install` via NSAppleScript
3. Wait for socket, connect

---

## Phase 5 — End-to-End Integration with sixnet-client

**Goal:** Swift app fully replaced with daemon-backed UI. Q1-ready.

**Milestone:** real users can install the app, connect to Q1, access hosted services,
disconnect — entirely from the menu bar, no terminal.

### What to build (in sixnet-client)

- `DaemonClient.swift` — connects to `/var/run/sixnetd.sock`, sends JSON commands
- Replace all `// TODO` in `MenuBarView.swift` with daemon calls
- Status display wired to daemon state (via polling or push from Phase 3)
- Onboarding flow:
  1. Check ZeroTier installed (from `status` response)
  2. If not: open zerotier.com/download
  3. Join network (`join` command)
  4. Show node ID → user copies it to enrollment portal
  5. Open enrollment URL in browser
  6. Wait for authorization (poll `status`)
  7. Connect
- Error states: ZeroTier not installed, daemon not running, not authorized

---

## Post-Q1

- Rewrite `vpn/zt` bash wrapper to use sixnetd socket (all logic already in daemon)
- Linux packaging: systemd unit, Makefile target for `.deb` / Homebrew formula
- Config management: how network ID + enrollment URL reach the client (currently hardcoded)
- Windows: stretch goal, different service mechanism

---

## Critical path to Q1

```
Phase 1 (state reading)
    → Phase 2 (connect/disconnect)
        → Phase 4 (installer)
            → Phase 5 (Swift integration)
```

Phase 3 (push) is a UX improvement — not on the critical path. The Swift app
polls the daemon at 3-5s intervals until Phase 3 is done.
