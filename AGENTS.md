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

Saved hosts stored at `~/.config/FishTun/hosts` in INI format (`[name]` sections with `KEY="value"`).
Passwords are never persisted.

## Source files

| File | Lines | Purpose |
|---|---|---|
| `main.go` | ~50 | Wails app entry, window options |
| `app.go` | ~320 | Bridge: Connect, Disconnect, OpenBrowser, GetHosts/SaveHost/DeleteHost, config, event emit |
| `ssh.go` | ~235 | Tunnel struct — ssh.Dial, local TCP listener, connection relay (io.Copy), remote command, keepalive, graceful shutdown |
| `config.go` | ~150 | Config struct, LoadConfig()/SaveConfig(); SavedHost, LoadHosts()/SaveHosts() |
| `tunnel.go` | ~40 | openURL(), waitForLocalPort() helpers |
| `systray.go` | ~40 | System tray menu and event pipes |
| `frontend/src/App.vue` | ~380 | Root Vue component: tunnels/hosts views, form toggle, log, error, events |
| `frontend/src/components/Form.vue` | ~240 | SSH connection form (host name, ports, host, user, password, commands) |
| `frontend/src/components/HostsView.vue` | ~150 | Saved hosts list, Connect / Edit / Delete actions |
| `frontend/src/components/TunnelsView.vue` | ~215 | Connected tunnels list, Open Browser / Disconnect buttons |
| `frontend/src/components/LogView.vue` | ~120 | Console log panel with full-window expand toggle (fixed-inset overlay, Escape to collapse; state lives in `App.vue` as `logExpanded`) |

## Architecture

1. **GUI** (Wails + Vue 3 + TypeScript) — user enters SSH params, ports, optional password, remote command; shows connection status, console log, tunnel list
2. **SSH tunnel** — `ssh.Dial` with key-based auth (id_rsa) + password fallback (GUI or terminal prompt); local TCP listener forwards to remote via `sshClient.Dial`
3. **Remote command** — optional command executed via SSH session on connect; optional disconnect command on teardown
4. **Keepalive** — `keepalive@openssh.com` every 30s; auto-disconnects on failure
5. **OS browser** — tunnel URL opened via `xdg-open` / `open` / `rundll32`
6. **Config persistence** — `~/.config/FishTun/config` in `KEY="value"` format, loaded on start, saved on submit
7. **System tray** — background operation with Show Window, Disconnect All, Make More Tunnels, Quit
8. **Events** — Go bridge emits `log`/`status`/`error`/`tunnels` events via Wails EventsEmit; Vue listens via EventsOn
9. **Single instance** — `instance.go` uses an O_EXCL lock file (`~/.config/FishTun/instance.lock`) holding a loopback IPC port; a second launch dials it, sends `show`, and exits, so only one process runs. Stale locks from a crash are recovered by removing the file and retrying.

Tray uses `github.com/energye/systray` via `RunWithExternalLoop` (started in `main` before `wails.Run`); click handlers dispatch to Wails only through non-blocking channel sends. `HideWindowOnClose` must stay off: with it enabled Wails hides the window in `OnClose` and never calls `OnBeforeClose`, so the X button bypasses the tray-hide logic. `OnBeforeClose` is the single hide path and never blocks: it shows the window, emits the `confirm-close` event and returns `true` (close cancelled). The Vue app renders an in-app popup with **System tray** (`HideToTray`: `runtime.Hide`) and **Close app** (`ExitApp`: sets `quitting` then `runtime.Quit`). Never use `runtime.MessageDialog` here — Win32 `QuestionDialog` is hard-limited to Yes/No labels.

Icons: `icon.png` (repository root) is the single source of truth. Windows tray/window need an ICO container (Win32 rejects raw PNG), so run `go run genico.go` after changing `icon.png`: it regenerates root `icon.ico` (16/32/48 as DIB, 256 as PNG) and mirrors it to `build/windows/icon.ico` for the Wails exe resource. `main.go` embeds both and picks via `trayIcon()` — ICO on Windows, PNG elsewhere (Linux GTK/`Linux.Options.Icon`, macOS).

## Release

1. Bump version in `wails.json` → `info.productVersion`
2. Commit, tag, push:
   ```
   git add -A && git commit -m "chore: release vX.Y.Z"
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
3. GitHub Actions builds 5 variants (linux amd64/arm64 .tar.gz, macos universal .dmg, windows amd64/arm64 .exe) and creates a Release with auto-generated release notes. macOS is built once on `macos-15` with `-platform darwin/universal` (GitHub retired Intel `macos-13`/`macos-14` runners); Wails `-platform` sets GOOS/GOARCH/CGO_ENABLED itself, so the workflow must not set them (only `CC` for linux/arm64 cross-compile), and `-skipbindings` is forbidden because `frontend/wailsjs` is gitignored and generated during `wails build` before the frontend build. linux/arm64 cannot be cross-compiled from the x86_64 runner (cgo+GTK needs aarch64 dev packages, not present there), so it builds inside an emulated `ubuntu:24.04 --platform linux/arm64` container via `docker run` with `wails build -s` (frontend and bindings are prepared on the host first).


# SYSTEM ROLE & OPERATIONAL RULES

### 1. Pre-Execution Root-Cause Analysis
Before making any edits or writing code, you MUST output a concise multi-step technical analysis using this exact structure:
[ANALYSIS]
- Root Issue: [True underlying cause, not just the symptom]
- Side-Effects: [Potential breaking changes in related systems/configs]
- Edge Cases: [At least 2 hidden edge cases or fail-states]
- Strategy: [Step-by-step logic of the proposed structural fix]

Never skip this block. Complete the analysis before calling file-editing tools or outputting code. Keep each point strictly factual and telegraphic to save tokens.

### 2. Token Economy & Language Rules
- **Strict Token Economy:** Maximize token efficiency. Output ONLY what is directly relevant to the task. Zero conversational filler, pleasantries, meta-commentary, or summaries of what was changed. Every word must carry technical payload.
- **Language Match:** Always communicate, fill out the `[ANALYSIS]` block, and write status updates in the exact language used by the user in their prompt.
- **Code Comments:** All inline comments and docstrings inside the code MUST be written strictly in English, regardless of the chat language.
- **Completion Trigger:** After successful execution, output only a single word equivalent to "Done." or "Fixed." in the user's language (e.g., "Готово." / "Исправлено.").

### 3. Code Editing & Display Hygiene
- **File Editing:** Use your native editing/diff tools freely in whatever way is most efficient and reliable to apply changes to the codebase.
- **Chat Code Output:** Do not echo modified code into the chat after editing files. ONLY display code in the chat if the user explicitly asks to see it — and when asked, output only the relevant snippets or target functions wrapped in markdown code blocks, never the full file.

### 4. Agent Memory Upkeep (AGENTS.md)
- If a systemic architectural rule, project convention, or operational constraint is established or modified during the session, append exactly one concise line to `AGENTS.md`. Do not touch other configuration files unless explicitly required by the task.

### 5. Stuck Loop & Expert Handoff
- If an issue remains unresolved after 2 distinct attempts, stop retrying immediately.
- Output the equivalent of "Prompt for expert AI:" in the user's language, followed by a single markdown code block containing a compact, self-contained prompt (context, current code, failed attempts, and desired outcome).