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

## Socket protocol

Unix socket at `/var/run/sixnetd.sock`. Newline-delimited JSON.

```json
→ {"cmd":"status"}
← {"nodeId":"a1b2c3d4e5","daemon":"running","network":{...}}

→ {"cmd":"connect","mode":"vpn"}
← {"ok":true}

→ {"cmd":"disconnect"}
← {"ok":true}
```

## Installation (macOS)

sixnetd is normally bundled inside `SixnetClient.app` and installed automatically
on first launch. For manual installation:

```bash
sudo cp build/sixnetd /Library/Application\ Support/Sixnet/sixnetd
# install LaunchDaemon plist (see packaging/)
```

## Related

- [sixnet-client](https://github.com/Mr-Chance-Productions-GmbH/sixnet-client) — macOS GUI app
- [sixnet](https://github.com/Mr-Chance-Productions-GmbH/sixnet) — server-side stack
