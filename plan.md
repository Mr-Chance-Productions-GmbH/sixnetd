# Plan

## Current focus: Phase 6 — Onboarding flow (design session needed)

Phases 1, 2, 4, and 5 are complete and smoke-tested.

## What works

**Socket protocol** — newline-delimited JSON over `/var/run/sixnetd.sock` (0666).
Request/response. Phase 3 (push) deferred until after Q1.

**Commands implemented:**
- `status` — ZeroTier daemon state, node ID, version, full network state
- `join` — POST /network/<nwid> (ZeroTier HTTP API)
- `leave` — DELETE /network/<nwid>, removes DNS resolver file
- `connect` — sets allow flags for requested mode, writes /etc/resolver/<domain>
- `disconnect` — clears all allow flags, removes /etc/resolver/<domain>

**Mode → allow flags** (from `vpn/zt` reference):
- `vpn`  → allowManaged=true,  allowDNS=true,  allowGlobal=false, allowDefault=false
- `lan`  → allowManaged=true,  allowDNS=true,  allowGlobal=true,  allowDefault=false
- `exit` → allowManaged=true,  allowDNS=true,  allowGlobal=true,  allowDefault=true
- (disconnected) → all false

**Internal package structure:**
```
internal/
  zerotier/
    state.go    — structs, DeriveMode
    client.go   — ZeroTier HTTP API (get/post/del, JoinNetwork, LeaveNetwork, SetNetworkFlags, NodeStatus, NetworkState)
    cli.go      — IsInstalled(), CLIPath()
    ops.go      — ModeFlags(mode)
  dns/
    resolver.go — Write(domain, server), Remove(domain) for /etc/resolver/
  socket/
    server.go   — Unix socket listener, dispatch, all command handlers
    protocol.go — Request, Response types
```

## Phases

**Phase 1 — State reading** ✓ done
Daemon answers `{"cmd":"status"}`. Smoke-tested: socket → ZeroTier API → response.

**Phase 2 — Connect / Disconnect** ✓ done
Join, leave, set flags, DNS resolver management. Smoke-tested: connect/disconnect cycle verified.

**Phase 3 — State push** (post-Q1)
Daemon pushes state changes to connected clients. Swift app becomes event-driven.

**Phase 4 — Homebrew packaging** ✓ done
Formula (binary only, no service block) + cask stub in tap `Mr-Chance-Productions-GmbH/sixnet`.
`brew install --cask sixnet-client` installs everything. App starts daemon on demand via
NSAppleScript. No LaunchDaemon, no system integration, clean uninstall without root.

**Phase 5 — Swift integration** ✓ done
Multi-network model. DaemonClient (POSIX Unix socket, 5s polling, per-network state).
AddNetworkView modal sheet: user pastes enrollment URL → client fetches `client.json`
→ gets networkId/name/enrollUrl automatically. MenuBarView: per-network rows with
status dot, mode, IP, join/connect/disconnect. Exit-mode conflict handling.
Smoke-tested: add → connect (DNS resolves, ping) → disconnect → reconnect.

**Phase 6 — Onboarding flow** ← current (design session needed)
Guided first-use: add network → join → show node ID → open enrollment portal →
wait for authorization → connect. TBD with explicit UI design session.

**Phase 7 — config.json endpoint on enrollment app**
`GET /client.json` on the enrollment server returns networkId, name, enrollUrl.
Currently mocked with a static file during development.
