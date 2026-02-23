# Todo

## Phase 1 — State reading ✓

- [x] `internal/zerotier/state.go` — DaemonState, NodeState, NetworkState structs
- [x] `internal/zerotier/client.go` — authtoken read, HTTP API client (/status, /network/<nwid>)
- [x] `internal/zerotier/cli.go` — IsInstalled(), binary path detection
- [x] `internal/socket/protocol.go` — request/response types
- [x] `internal/socket/server.go` — Unix socket listener, handle `status` command
- [x] `cmd/sixnetd/main.go` — wire up: socket server + signal handling
- [x] smoke test: `echo '{"cmd":"status"}' | nc -U /var/run/sixnetd.sock`

## Phase 2 — Connect / Disconnect ✓

- [x] `internal/zerotier/client.go` — `JoinNetwork()`, `LeaveNetwork()`, `SetNetworkFlags()`
- [x] `internal/zerotier/ops.go` — `ModeFlags(mode)` — returns allow flag map for a mode
- [x] `internal/dns/resolver.go` — `Write(domain, server)` and `Remove(domain)` for /etc/resolver/
- [x] `internal/socket/server.go` — handle `join`, `leave`, `connect`, `disconnect` commands
- [x] smoke test: connect vpn → `/etc/resolver/q1.zt` written, flags set
- [x] smoke test: disconnect → resolver removed, all allow* flags false

## Phase 3 — State push

- [ ] client registry in socket server
- [ ] push state to all clients on change
- [ ] Swift app switches from polling to event-driven

## Phase 4 — Homebrew packaging ✓

New repo: `Mr-Chance-Productions-GmbH/homebrew-sixnet` (tap)

- [x] `sixnetd --version` flag — prints `0.1.0`, tested
- [x] Tag v0.1.0 on sixnetd GitHub — source tarball URL needed by formula
- [x] Formula: `Formula/sixnetd.rb` — builds from source, binary only, no service block
- [x] Cask: `Casks/sixnet-client.rb` — stub, declares sixnetd as dependency
- [x] Test: `brew install Mr-Chance-Productions-GmbH/sixnet/sixnetd` → builds, installs
- [x] Test: `brew test sixnetd` → --version passes
- [x] Test: `sudo sixnetd` → daemon starts, socket responds
- [x] Test: `brew uninstall sixnetd` → clean, no root needed

## Phase 5 — Swift integration (in sixnet-client) ✓

- [x] `DaemonClient.swift` — multi-network model, POSIX Unix socket, JSON protocol, polling timer
- [x] `AddNetworkView.swift` — modal sheet: URL → fetch client.json → save network config
- [x] `MenuBarView.swift` — per-network rows: status, join/connect/disconnect, enrollment prompt
- [x] `SixnetClientApp.swift` — AppDelegate wires DaemonClient, first-launch daemon start via NSAppleScript
- [x] Exit mode conflict handling — disconnect previous exit-mode network before connecting new one
- [x] Network persistence — [SavedNetwork] stored in UserDefaults as JSON
- [x] smoke test: build + launch app against live sixnetd
- [x] smoke test: Add Network → fetch client.json → network appears in list
- [x] smoke test: Connect → green dot, mode badge, IP, DNS resolves, ping works
- [x] smoke test: Disconnect → orange dot, DNS removed, ping stops
- [x] smoke test: reconnect cycle verified

## Phase 6 — Onboarding flow (design session needed)

Design: TBD with explicit UI session

- [ ] Step 1: Add Network (URL → fetch config → join → show node ID)
- [ ] Step 2: Enrollment prompt (open portal URL in browser, wait for auth)
- [ ] Step 3: Connect (first connection after authorization)

## Phase 7 — config.json endpoint on enrollment app

- [ ] `GET /client.json` on enrollment app → serves networkId, name, enrollUrl
- [ ] (currently: mock with static file during development)
