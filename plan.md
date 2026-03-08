# Plan

## Current focus: Distribution (cask completion + first real-user deployment)

Phases 1–7 are complete. sixnetd is stable and fully integrated with the Swift client.
Next: build and release the sixnet-client .app so the Homebrew cask can be completed
and real users can install via `brew install --cask sixnet-client`.

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
Version injected via ldflags at build time (`-X main.version=#{version}`); source defaults to "dev".
App starts daemon on demand via NSAppleScript. No LaunchDaemon, no system integration,
clean uninstall without root.

**Phase 5 — Swift integration** ✓ done
Multi-network model. DaemonClient (POSIX Unix socket, 5s polling, per-network state).
AddNetworkView modal sheet: user pastes enrollment URL → client fetches `client.json`
→ gets networkId/name/enrollUrl automatically. MenuBarView: per-network rows with
status dot, mode, IP, join/connect/disconnect. Exit-mode conflict handling.
Smoke-tested: add → connect (DNS resolves, ping) → disconnect → reconnect.

**Phase 6 — Mode 2 PKCE enrollment** ✓ done (in sixnet-client)
Device-initiated OAuth PKCE flow. Client fetches OIDC discovery, opens system browser,
catches callback on localhost:12345, exchanges code for id_token, POSTs to /claim.
On success: auto-joins the ZeroTier network, shows "Enrolled — ready to connect".
User then presses Connect in the menu bar to choose mode (vpn/lan/exit).
Onboarding flow emerges naturally from the UI state machine — no separate wizard needed.

**Phase 7 — /client.json endpoint** ✓ done (in six.net enroll app)
`GET /client.json` on the enrollment server returns networkId, name, enrollUrl, issuer, clientId.
Fetched by AddNetworkView when user pastes the enrollment URL.
