# FishTun

<p align="left">
  <img src="icon.png" alt="FishTun logo" width="96" style="vertical-align: middle; margin-right: 10px;">
  <span style="font-weight: bold; vertical-align: middle;font-style: italic;">Native GUI app — multi-tunnel SSH port forwarding</span>
</p>

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Wails](https://img.shields.io/badge/Wails-v2-DF0000?logo=wails)](https://wails.io)

## Overview

FishTun is a desktop application built with Go, Wails, and Vue 3 that lets you create and manage multiple SSH tunnels through a native GUI. It forwards a remote port to your local machine and can optionally execute a command on the remote host (start/stop hooks).

## Features

- 🔐 SSH tunnel via `golang.org/x/crypto/ssh`
- 📡 Multiple tunnels simultaneously
- 🔄 Keepalive (30s interval), auto-disconnect on failure
- 🖥️ Native GUI (Wails WebView)
- 🔑 Auth: key-based (id_rsa) + password fallback (GUI or terminal prompt)
- ⚙️ Config persistence in `~/.config/FishTun/config`
- 🪟 System tray: minimize, disconnect all, quick actions
- 🌗 Light/dark theme (follows system preference)
- ▶️ Remote command execution on connect (e.g. start your app)
- ⏹️ Disconnect hook on tunnel teardown

## Requirements

- Go 1.25+
- CGO enabled
- Linux dev headers: GTK3, WebKit2GTK 4.1, X11
- Wails CLI v2

```bash
sudo apt install gcc libx11-dev libxrandr-dev libxinerama-dev libxi-dev \
  libgl1-mesa-dev libxxf86vm-dev libgtk-3-dev libwebkit2gtk-4.1-dev

go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Quick Start

```bash
git clone https://github.com/alexoffspace/FishTun.git
cd FishTun
wails build -tags webkit2_41
./build/bin/FishTun
```

## Configuration

Saved automatically to `~/.config/FishTun/config`:

```
LOCAL_PORT="2222"
REMOTE_PORT="8080"
SSH_HOST="my.server.com"
SSH_USER="root"
SSH_PORT="22"
LOCAL_HOST="127.0.0.1"
REMOTE_CMD="nohup myapp --port 8080 &"
DISCONNECT_CMD="pkill -f myapp"
```

Fields are editable from the GUI. The config persists between runs.

## Project Structure

| File | Purpose |
|---|---|
| `main.go` | Wails app entry, window options |
| `app.go` | Go ↔ frontend bridge: Connect, Disconnect, config, event emit |
| `ssh.go` | Tunnel core: SSH dial, TCP listener, connection relay, remote command, keepalive |
| `config.go` | Config load/save (`~/.config/FishTun/config`) |
| `tunnel.go` | Helpers: `openURL`, `waitForLocalPort` |
| `systray.go` | System tray setup |
| `frontend/` | Vue 3 + TypeScript UI |

## Architecture

1. **GUI** (Wails + Vue 3) — user enters SSH params, sees connection status, opens browser, manages tunnels
2. **SSH tunnel** — `ssh.Dial` with key-based + password auth; local TCP listener forwards to remote via `sshClient.Dial`
3. **Remote command** — optional command executed via SSH session on connect; optional disconnect hook on teardown
4. **Keepalive** — `keepalive@openssh.com` every 30s; auto-disconnects on failure
5. **Browser** — tunnel URL opened via `xdg-open` / `open` / `rundll32`
6. **System tray** — background operation, quick disconnect, quit
7. **Events** — Go bridge emits `log`/`status`/`error`/`tunnels` events; Vue listens via `EventsOn`

## License

MIT © 2026 Alexey Belov
