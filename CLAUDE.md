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
- All subsequent operations go through the socket — no privilege escalation needed
- The macOS GUI app, bash wrapper, and any future client are all unprivileged

## Reference implementation

`vpn/zt` in `~/projects/six.net` is the bash script this daemon replaces. It is the
authoritative reference for every operation sixnetd must implement. When in doubt
about behavior, read that script first.

Key functions to reference:
- `get_network_status()` — parses `zerotier-cli listnetworks` output
- `get_zt_dns()` — queries ZeroTier HTTP API for DNS domain + server IP
- `get_setting()` — reads individual network flags via `zerotier-cli get`
- `is_joined()` — checks network membership
- `cmd_up()` / `cmd_down()` — connect/disconnect with mode management
- `setup_dns_resolver()` / `remove_dns_resolver()` — /etc/resolver/ management

Once sixnetd is stable, `vpn/zt` will be rewritten to talk to the socket instead
of calling zerotier-cli directly.

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
- `GET /network/<nwid>` — full network state, DNS config (domain + servers), assigned IPs

## Packaging

Distribution is via Homebrew — tap at `Mr-Chance-Productions-GmbH/homebrew-sixnet`.

**`brew install Mr-Chance-Productions-GmbH/sixnet/sixnetd`**
- Builds binary from source into the Cellar, symlinks to `/opt/homebrew/bin/sixnetd`
- Formula includes a `service` block → `brew services start sixnetd` loads the
  LaunchDaemon on macOS, a systemd unit on Linux
- `brew uninstall sixnetd` removes binary and plist cleanly

**`brew install --cask Mr-Chance-Productions-GmbH/sixnet/sixnet-client`**
- Installs sixnetd formula as a dependency (binary arrives automatically)
- Installs SixnetClient.app to /Applications
- Homebrew Cask runs `xattr -dr com.apple.quarantine` — no Gatekeeper dialog
- `brew uninstall --cask sixnet-client` removes the app cleanly

**First-launch flow (in Swift app):**
1. Check if `/var/run/sixnetd.sock` is alive
2. If not: show one-time setup screen
3. Run `brew services start sixnetd` via NSAppleScript — one admin dialog, ever
4. Daemon is now running and auto-starts at every boot

**No `--install` / `--uninstall` flags in sixnetd** — Homebrew owns the lifecycle.
`--version` is implemented for debugging and upgrade checks.

## Related repos

- `sixnet-client` — macOS Swift GUI app (bundles this daemon, talks to socket)
- `six.net` — server-side sixnet stack (ZeroTier controller, Authentik, Caddy, DNS)

## Current state

See `plan.md` for current implementation approach and `todo.md` for task tracking.
