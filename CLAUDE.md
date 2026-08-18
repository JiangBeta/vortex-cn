# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Vortex is a keyboard-first BubbleTea TUI (in Go) for managing a fleet of Linux VPS servers over SSH. It ships as **two binaries**:

- `cmd/vps-manager` — the TUI, run locally. The `Router` in its `main.go` is the single BubbleTea root model.
- `cmd/vortex-agent` — a tiny stat-gathering CLI that prints a single JSON `payload` to stdout. It is cross-compiled for the target OS, pushed over SSH, and executed remotely.

## Commands

The Go module is named `main` (not a domain), so import paths look like `main/internal/...`. `go.mod` pins `go 1.26.5`; the Nix flake overrides nixpkgs' Go to match.

```bash
# Build the TUI (README's convention; the .exe suffix is a naming quirk, not Windows-only)
go build -o bin/vps-manager.exe ./cmd/vps-manager

# Run the TUI
go run ./cmd/vps-manager            # or ./bin/vps-manager.exe

# Build the remote agent for Linux (this is also done automatically at connect time)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o vortex-agent ./cmd/vortex-agent

# Tests: there are currently no *_test.go files; this only type-checks packages.
go test ./...
go vet ./...

# Lint (golangci-lint is provided by the Nix devShell; no config file is committed)
golangci-lint run

# Nix
nix develop                          # shell with go, gopls, golangci-lint, delve
nix build .#default                  # builds via buildGoModule; subPackages = [ "cmd/vps-manager" ]
nix flake check --all-systems
```

Config lives at `~/.config/vortex/config.yaml` (XDG-aware), not in the repo.

## Architecture

### Telemetry flow

1. The Servers page calls `ssh.Connect(...)` → emits `ssh.ConnectedMsg`.
2. `Router` handles `ConnectedMsg`: it calls `ssh.Client.DeployAndRunAgent()`, which runs `go build` locally to compile `cmd/vortex-agent` for the remote OS, streams the binary over SSH (raw for Linux, base64+PowerShell for Windows), then executes `./vortex-agent payload` remotely.
3. `internal/engine/system`'s `Engine.FetchPayload()` polls `./vortex-agent payload` on a tick (interval from `config.Monitoring.RefreshInterval`), unmarshals it into `agent.Payload`, and the `Router` broadcasts that payload to every page.

The `agent.Payload` JSON (CPU/RAM/disk, network, docker, systemd services, optional logs) is the shared state all pages render from.

### Package layering

There are two distinct "engine" concepts that are easy to confuse:

- **Agent-side gatherers** — `internal/{stats,network,docker,services}` — run *on the remote host inside the compiled agent*. They inspect local state directly (gopsutil, `exec.Command("ip"/"ss"/"systemctl")`, and `internal/docker` uses the real Docker client SDK `github.com/docker/docker` against the local daemon). These packages are imported by `cmd/vortex-agent`, never by the TUI's page code.
- **Page-side engines** — `internal/engine/*` (docker, systemd, cron, certs, files, logs, …) — run *inside the TUI* and shell out to the remote host by calling `ssh.Client.Run(...)`. One engine per module.

The remaining layers:

- `internal/pages/*` — one BubbleTea model per sidebar module. Each implements the `pages.Page` interface (`tea.Model` + `Title()` + `Icon()`).
- `internal/components` — reusable UI widgets: `Card`, `Toast`, `Palette` (command palette), `Sparkline`, `Globe`, `Startup`, `Progress`.
- `internal/config` — `config.yaml` load/save plus a `Registry` of typed settings that back the Settings page; `config.CurrentConfig` is a global the whole app reads from.
- `internal/theme` — a fixed slice of `Theme` structs (lipgloss colors) with `Current` selected by name.
- `internal/ssh` — the SSH wrapper (`Connect`, `Run`, `RunCommand`, `DeployAndRunAgent`, `Close`).

### How pages talk to the Router

Pages never touch the SSH client directly for one-off commands. Instead they return `tea.Msg` values that `Router` (in `cmd/vps-manager/main.go`) handles centrally, defined in `internal/pages/page.go`:

- `RunRemoteCmdMsg{Command}` — fire-and-forget remote command.
- `RunRemoteQueryMsg{Command, ResponseHandler}` — run a remote command and route the stdout through a handler that returns the next message.
- `LogActivityMsg{Message}` — append to the Mission Control activity feed.
- `EngineReadyMsg` — broadcast on connect so each page can construct its own `internal/engine/*` engine.

`Router` also implements the `IsInputActive() bool` optional interface check: a page returns `true` while a text input/editor is focused, which suppresses the global keybinds (`g`, `f`, `x`, …) so keystrokes reach the input instead.

### Adding a module

Sidebar order and page indexing are **hardcoded**, not driven by the registry. To add a page you must touch several places in `cmd/vps-manager/main.go`:

1. The `pages: []pages.Page{...}` slice in `initialModel()` (its index is the page's identity).
2. The global-keybind `switch` block (each keybind maps a key to a literal index).
3. Optionally a `Command Palette` registration.

`internal/pages` also exports a `Register`/`GetAll` registry, but the `Router` does not use it — it builds its own slice, so registration alone won't add a page to the UI.

## Gotchas

- **Host key verification is disabled**: `ssh.Client.Connect` uses `ssh.InsecureIgnoreHostKey()`. The `SSHConfig.KnownHosts` config field exists but is not wired into connection logic yet.
- **The TUI requires the Go toolchain at runtime**: connecting to a server invokes `go build` to cross-compile the agent, so a deployed TUI still needs Go installed locally (unless the agent binary is pre-staged).
- **Compiled binaries are committed at the repo root** (`vortex.exe`, `tmp.exe`, `agent-test`, etc.) and are not covered by `.gitignore`.
- `internal/engine/docker` shells out to the `docker` CLI over SSH; `internal/docker` uses the Docker SDK — same feature area, two unrelated code paths.
