# Claude Briefing - sixnetd

## What is this repo?

Privileged background daemon for sixnet — the self-hosted VPN platform built on ZeroTier.

sixnetd is the business logic and privilege layer between ZeroTier and any sixnet client
(GUI app, CLI wrapper, future platform clients). It runs as root, owns all ZeroTier
operations, and exposes a Unix socket with a JSON protocol that unprivileged clients use.

## Why a daemon?

ZeroTier operations require root:
- `zerotier-cli join/leave/set` — all need sudo
- Reading `/Library/Application Support/ZeroTier/One/authtoken.secret` — root-only (0600)
- Writing `/etc/resolver/<domain>` — root-only

Without a daemon, every connect/disconnect triggers an OS auth dialog. With sixnetd:
- Daemon installs once (one admin dialog at app first-launch)
- All subsequent operations go through the socket — no more privilege escalation
- The macOS GUI app, bash wrapper, and any future client are all unprivileged

## Architecture

```
ZeroTier daemon (zerotier-one)
    ↓
zerotier-cli + ZeroTier HTTP API (:9993)
    ↓
sixnetd                              ← this repo
    /var/run/sixnetd.sock (JSON)
    ↓
clients: SixnetClient (macOS GUI), zt (bash), ...
```

## Socket protocol

Unix socket at `/var/run/sixnetd.sock`.
Newline-delimited JSON — one request per line, one response per line.

**Commands (to be defined):**
- `{"cmd":"status"}` — ZeroTier install status, daemon status, node ID, network state
- `{"cmd":"connect","mode":"vpn|lan|exit"}` — connect with given mode
- `{"cmd":"disconnect"}` — disconnect
- `{"cmd":"join","networkId":"<nwid>"}` — join a network
- `{"cmd":"leave","networkId":"<nwid>"}` — leave a network

## ZeroTier operations wrapped

```bash
zerotier-cli join <nwid>
zerotier-cli leave <nwid>
zerotier-cli set <nwid> allowDNS=true|false
zerotier-cli set <nwid> allowManaged=true|false
zerotier-cli set <nwid> allowGlobal=true|false      # lan mode
zerotier-cli set <nwid> allowDefault=true|false     # exit mode
zerotier-cli listnetworks
zerotier-cli info
```

Local ZeroTier HTTP API (port 9993, authtoken at known path):
- `GET /status` — node ID, version, online status
- `GET /network/<nwid>` — full network state, DNS config, assigned IPs

DNS resolver (macOS):
- Create/remove `/etc/resolver/<domain>` pointing to VPN DNS server

## Key file paths

| Path | Purpose |
|------|---------|
| `/Library/Application Support/ZeroTier/One/authtoken.secret` | ZeroTier API auth (macOS) |
| `/var/lib/zerotier-one/authtoken.secret` | ZeroTier API auth (Linux) |
| `/usr/local/bin/zerotier-cli` | ZeroTier CLI (macOS standard install) |
| `/var/run/sixnetd.sock` | Unix socket for client connections |
| `/etc/resolver/<domain>` | macOS DNS resolver config |

## Packaging

**macOS:** bundled inside `SixnetClient.app/Contents/MacOS/sixnetd`.
On first app launch, installed to `/Library/Application Support/Sixnet/sixnetd`
and registered as a LaunchDaemon at `/Library/LaunchDaemons/de.mcp.sixnet.daemon.plist`.

**Linux:** same binary, different packaging (systemd unit, package manager).
The daemon code has no macOS-specific dependencies.

## Reference implementation

The bash `vpn/zt` wrapper in `~/projects/six.net` is the reference for all operations
this daemon replicates. Once sixnetd is stable, the `zt` wrapper will be rewritten
to talk to the socket instead of calling zerotier-cli directly.

## Related repos

- `sixnet-client` — macOS Swift GUI app (bundles this daemon)
- `six.net` — server-side sixnet stack (ZeroTier controller, Authentik, Caddy, DNS)
  - `vpn/zt` — bash reference implementation
