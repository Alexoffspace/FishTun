# FishTun

Native GUI app for creating and managing multiple SSH tunnels to remote hosts.

## Build

```
wails build -tags webkit2_41
```

Requires Go 1.25+, CGO, Wails CLI, and Linux dev headers:

```bash
sudo apt install gcc libx11-dev libxrandr-dev libxinerama-dev libxi-dev \
  libgl1-mesa-dev libxxf86vm-dev libgtk-3-dev libwebkit2gtk-4.1-dev
```

## Dependencies

- `github.com/wailsapp/wails/v2` — native WebView GUI (Vue 3 frontend)
- `golang.org/x/crypto/ssh` — direct SSH tunnel (no exec)
- `golang.org/x/term` — terminal password prompt fallback
- `github.com/getlantern/systray` — system tray

## Config

Stored at `~/.config/FishTun/config` in `KEY="value"` format.
Persisted between runs; GUI pre-fills with saved values.

## Source files

| File | Lines | Purpose |
|---|---|---|
| `main.go` | ~50 | Wails app entry, window options |
| `app.go` | ~320 | Bridge: Connect, Disconnect, OpenBrowser, config, event emit |
| `ssh.go` | ~235 | Tunnel struct — ssh.Dial, local TCP listener, connection relay (io.Copy), remote command, keepalive, graceful shutdown |
| `config.go` | ~150 | Config struct (LocalPort, RemotePort, SSHHost, SSHUser, SSHPort, LocalHost, RemoteCmd, DisconnectCmd), LoadConfig(), SaveConfig() |
| `tunnel.go` | ~40 | openURL(), waitForLocalPort() helpers |
| `systray.go` | ~40 | System tray menu and event pipes |
| `frontend/src/App.vue` | ~380 | Root Vue component: tunnels list, form toggle, log, error, events |
| `frontend/src/components/Form.vue` | ~240 | SSH connection form (ports, host, user, password, commands) |
| `frontend/src/components/TunnelsView.vue` | ~215 | Connected tunnels list, Open Browser / Disconnect buttons |
| `frontend/src/components/LogView.vue` | ~85 | Scrollable console log panel |

## Architecture

1. **GUI** (Wails + Vue 3 + TypeScript) — user enters SSH params, ports, optional password, remote command; shows connection status, console log, tunnel list
2. **SSH tunnel** — `ssh.Dial` with key-based auth (id_rsa) + password fallback (GUI or terminal prompt); local TCP listener forwards to remote via `sshClient.Dial`
3. **Remote command** — optional command executed via SSH session on connect; optional disconnect command on teardown
4. **Keepalive** — `keepalive@openssh.com` every 30s; auto-disconnects on failure
5. **OS browser** — tunnel URL opened via `xdg-open` / `open` / `rundll32`
6. **Config persistence** — `~/.config/FishTun/config` in `KEY="value"` format, loaded on start, saved on submit
7. **System tray** — background operation with Show Window, Disconnect All, Make More Tunnels, Quit
8. **Events** — Go bridge emits `log`/`status`/`error`/`tunnels` events via Wails EventsEmit; Vue listens via EventsOn

## Release

1. Bump version in `wails.json` → `info.productVersion`
2. Commit, tag, push:
   ```
   git add -A && git commit -m "chore: release vX.Y.Z"
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
3. GitHub Actions builds 6 variants (linux amd64/arm64 .tar.gz, macos amd64/arm64 .dmg, windows amd64/arm64 .exe) and creates a Release with auto-generated release notes.


# SYSTEM ROLE & OPERATIONAL RULES

### 1. Mandatory Root-Cause Analysis (Before Code)
- Before writing any code snippet, you MUST output a mandatory multi-step technical analysis.
- Use the following rigid structural format inside the markdown code block before the code itself:
  /*
  [ANALYSIS]
  - Root Issue: [Identify the true underlying cause, not just the symptom]
  - Side-Effects: [List potential breaking changes in related systems/configs]
  - Edge Cases: [Identify at least 2 hidden edge cases or fail-states]
  - Strategy: [Step-by-step logic of the proposed structural fix]
  */
- Never skip this block. Do not write code until the analysis is fully detailed.

### 2. AI Communication (Extreme Economy)
- **ZERO CHAT EXPLANATIONS.**
- No introductory or concluding phrases (e.g., "Sure, I can help", "Here is the code").
- After successful execution: Output exactly "Done." or "Fixed." after the final code block.
- If clarification is strictly required: Output exactly "Q: [option A] or [option B]?"
- Error reporting: Output exactly "Error: [3-word reason]"

### 3. Output Formatting & Context Hygiene
- **No Full File Echoing.** Provide only the modified code snippets, functions, or diffs.
- **Markdown Only.** Wrap all code outputs strictly in appropriate markdown code blocks.
- **No Text Summaries.** Do not explain what changed or how it works in the chat.

### 4. Dependency & Agent Maintenance
- **Config Sync.** Automatically update related configuration files (`package.json`, `config.json`) in the same compact snippet format.
- **Agent Prompt Upkeep.** If a systemic rule changes, append exactly one short line to the system prompt file.

### 5. Stuck Loop & Expert Handoff
- If a solution fails after 2 distinct attempts, output: "Prompt for expert AI:" followed by a compact, self-contained prompt containing all context, inputs, and desired output, wrapped in a single markdown code block. Do not retry further.
