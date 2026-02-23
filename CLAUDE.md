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

Running as root is a known constraint, not a preferred design. The better long-term
model is Apple's NetworkExtension framework (no root, no /etc/resolver/ writes) but
that requires a paid Apple Developer account and a full rearchitecting effort.
See `~/projects/six.net/docs/architecture/client-privilege-model.md` for the full
decision record, constraints, and future path.

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
- Binary is user-owned — `brew uninstall sixnetd` removes it cleanly, no root needed

**`brew install --cask Mr-Chance-Productions-GmbH/sixnet/sixnet-client`**
- Installs sixnetd formula as a dependency (binary arrives automatically)
- Installs SixnetClient.app to /Applications
- Homebrew Cask runs `xattr -dr com.apple.quarantine` — no Gatekeeper dialog
- `brew uninstall --cask sixnet-client` removes the app cleanly

**No LaunchDaemon. No system integration. No plist anywhere.**
The formula installs a binary only. The Swift app starts it on demand.

**Daemon lifecycle:**
- App launches → checks if socket `/var/run/sixnetd.sock` is alive
- If not: runs `sudo /opt/homebrew/bin/sixnetd` via NSAppleScript — one admin dialog
- Daemon keeps running when the app quits — VPN stays connected
- Next app launch → socket still alive → no dialog
- After reboot → first app launch → one admin dialog again

**Uninstall (from app menu or CLI):**
```bash
brew uninstall --cask sixnet-client
brew uninstall sixnetd
```
No root required. The running daemon process exits on next reboot or can be
killed manually. No traces left in `/Library/` or anywhere system-level.

`--version` is implemented for debugging and upgrade checks.

## Related repos

- `sixnet-client` — macOS Swift GUI app (bundles this daemon, talks to socket)
- `six.net` — server-side sixnet stack (ZeroTier controller, Authentik, Caddy, DNS)

## Current state

See `plan.md` for current implementation approach and `todo.md` for task tracking.
