# sixnetd

Privileged background daemon for [sixnet](https://github.com/Mr-Chance-Productions-GmbH/sixnet) —
a self-hosted VPN platform built on ZeroTier.

sixnetd runs as root and owns all ZeroTier operations. Clients (GUI app, CLI tools)
talk to it via a Unix socket using a simple JSON protocol — no privilege escalation
needed after the one-time daemon installation.

## Architecture

```
zerotier-one daemon
    ↓
zerotier-cli + ZeroTier HTTP API (:9993)
    ↓
sixnetd  ←  /var/run/sixnetd.sock  ←  SixnetClient, zt, ...
```

## Build

```bash
make build    # produces build/sixnetd
make run      # build + run as root (sudo)
make clean
```

Requires Go 1.21+.

Build artifacts must not live on a cloud-synced filesystem.
Use `mkdir-nosync build` to create a symlink to a local path.

## Socket protocol

Unix socket at `/var/run/sixnetd.sock`. Newline-delimited JSON.

### Commands

```
{"cmd":"status","networkId":"<nwid>"}
{"cmd":"connect","networkId":"<nwid>","mode":"vpn|lan|exit"}
{"cmd":"disconnect","networkId":"<nwid>"}
{"cmd":"join","networkId":"<nwid>"}
{"cmd":"leave","networkId":"<nwid>"}
```

### Responses

```json
{
  "daemon": "running",
  "nodeId": "a1b2c3d4e5",
  "network": {
    "id": "31655f6ec3a15f6d",
    "name": "Q1 Office VPN",
    "status": "OK",
    "authorized": true,
    "mode": "vpn",
    "assignedIP": "10.147.20.42"
  }
}
```

```json
{"ok": true}
{"ok": false, "error": "not joined"}
```

### Connection modes

| Mode  | allowManaged | allowDNS | allowGlobal | allowDefault | Access |
|-------|-------------|----------|-------------|--------------|--------|
| `vpn` | true | true | false | false | ZeroTier devices only |
| `lan` | true | true | true  | false | + office LAN via gateway |
| `exit`| true | true | true  | true  | + internet via gateway |

### Key file paths

| Path | Purpose |
|------|---------|
| `/Library/Application Support/ZeroTier/One/authtoken.secret` | ZeroTier API auth (macOS) |
| `/var/lib/zerotier-one/authtoken.secret` | ZeroTier API auth (Linux) |
| `/usr/local/bin/zerotier-cli` | ZeroTier CLI (macOS standard install) |
| `/var/run/sixnetd.sock` | Unix socket for client connections |
| `/etc/resolver/<domain>` | macOS DNS resolver config (written on connect) |

## Installation

Via Homebrew (tap: `Mr-Chance-Productions-GmbH/sixnet`):

```bash
# Daemon only
brew install Mr-Chance-Productions-GmbH/sixnet/sixnetd
brew services start sixnetd

# GUI app + daemon (recommended — installs sixnetd as a dependency)
brew install --cask Mr-Chance-Productions-GmbH/sixnet/sixnet-client
```

The GUI app handles `brew services start sixnetd` on first launch via a one-time
admin dialog. No manual steps needed after the cask install.

**Uninstall:**
```bash
brew services stop sixnetd
brew uninstall --cask sixnet-client
brew uninstall sixnetd
```

**Linux:** same formula, `brew services` uses systemd instead of LaunchDaemon.

## Related

- [sixnet-client](https://github.com/Mr-Chance-Productions-GmbH/sixnet-client) — macOS GUI app (bundles sixnetd)
- [sixnet](https://github.com/Mr-Chance-Productions-GmbH/sixnet) — server-side stack
  - `vpn/zt` — bash reference implementation (the spec for what sixnetd wraps)
