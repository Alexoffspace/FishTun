package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type TunnelInfo struct {
	ID            string `json:"id"`
	LocalHost     string `json:"localHost"`
	LocalPort     int    `json:"localPort"`
	RemotePort    int    `json:"remotePort"`
	SSHUser       string `json:"sshUser"`
	SSHHost       string `json:"sshHost"`
	SSHPort       int    `json:"sshPort"`
	RemoteCmd     string `json:"remoteCmd"`
	DisconnectCmd string `json:"disconnectCmd"`
}

type App struct {
	ctx             context.Context
	config          Config
	hosts           []SavedHost
	tunnels         map[string]*Tunnel
	tunnelInfo      map[string]TunnelInfo
	mu              sync.Mutex
	iconData        []byte
	makeMoreCh      chan struct{}
	showWindowCh    chan struct{}
	disconnectAllCh chan struct{}
	quitCh          chan struct{}
	trayStarted     bool
	trayMu          sync.Mutex
	quitting        bool
}

func NewApp() *App {
	return &App{
		config:          LoadConfig(),
		hosts:           LoadHosts(),
		tunnels:         make(map[string]*Tunnel),
		tunnelInfo:      make(map[string]TunnelInfo),
		makeMoreCh:      make(chan struct{}, 10),
		showWindowCh:    make(chan struct{}, 10),
		disconnectAllCh: make(chan struct{}, 10),
		quitCh:          make(chan struct{}, 10),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.handleTrayEvents()
}

func (a *App) handleTrayEvents() {
	for {
		select {
		case <-a.makeMoreCh:
			a.showWindow()
		case <-a.showWindowCh:
			a.showWindow()
		case <-instanceShowCh:
			a.showWindow()
		case <-a.disconnectAllCh:
			a.DisconnectAll()
		case <-a.quitCh:
			a.DisconnectAll()
			a.quit()
			return
		}
	}
}

// showWindow reliably brings the window to the front, even if it is hidden or
// minimised. The transient always-on-top toggle raises it above other windows.
func (a *App) showWindow() {
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
}

// onBeforeClose handles the window close button. When no tunnel is active the
// app quits immediately (nothing to keep alive in the tray); otherwise the close
// is cancelled and the frontend is asked (in-app popup) whether to exit or keep
// running in the tray. The decision arrives asynchronously via ExitApp/HideToTray.
func (a *App) onBeforeClose(ctx context.Context) bool {
	a.trayMu.Lock()
	quitting := a.quitting
	a.trayMu.Unlock()
	if quitting {
		return false
	}
	// With no active tunnels there is nothing to keep running in the tray, so
	// the close button quits directly instead of showing the tray prompt.
	a.mu.Lock()
	hasTunnels := len(a.tunnels) > 0
	a.mu.Unlock()
	if !hasTunnels {
		a.trayMu.Lock()
		a.quitting = true
		a.trayMu.Unlock()
		return false
	}
	runtime.WindowShow(a.ctx)
	runtime.EventsEmit(a.ctx, "confirm-close")
	return true
}

// ExitApp quits the application after the user confirmed the close popup.
func (a *App) ExitApp() {
	a.trayMu.Lock()
	a.quitting = true
	a.trayMu.Unlock()
	runtime.Quit(a.ctx)
}

// HideToTray keeps the application running in the tray after the user chose so
// in the close popup.
func (a *App) HideToTray() {
	runtime.Hide(a.ctx)
}

// startTray launches the system tray with an external event loop so it can run
// alongside the Wails/WebView2 loop. start() must be called from a goroutine,
// never the UI thread. end() may be called at most once.
func (a *App) startTray() {
	a.trayMu.Lock()
	if a.trayStarted {
		a.trayMu.Unlock()
		return
	}
	a.trayStarted = true
	a.trayMu.Unlock()

	iconData := a.iconData
	start, _ := systray.RunWithExternalLoop(func() {
		onTrayReady(iconData, a.makeMoreCh, a.showWindowCh, a.disconnectAllCh, a.quitCh)
	}, func() {})

	go start()
}

func (a *App) quit() {
	a.trayMu.Lock()
	a.quitting = true
	a.trayMu.Unlock()

	// Stop the tray pump first so it can remove the icon (NIM_DELETE), then quit
	// Wails. Reversing the order can hang and orphan the tray icon.
	systray.Quit()
	time.Sleep(150 * time.Millisecond)
	runtime.Quit(a.ctx)
}

func (a *App) GetConfig() Config {
	return a.config
}

func (a *App) SaveConfig(cfg Config) {
	a.config = cfg
	SaveConfig(cfg)
}

func (a *App) GetHosts() []SavedHost {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]SavedHost, len(a.hosts))
	copy(out, a.hosts)
	return out
}

func (a *App) SaveHost(name string, cfg Config) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name required")
	}
	a.mu.Lock()
	found := false
	for i := range a.hosts {
		if a.hosts[i].Name == name {
			a.hosts[i] = SavedHost{Name: name, Config: cfg}
			found = true
			break
		}
	}
	if !found {
		a.hosts = append(a.hosts, SavedHost{Name: name, Config: cfg})
	}
	hosts := make([]SavedHost, len(a.hosts))
	copy(hosts, a.hosts)
	a.mu.Unlock()

	SaveHosts(hosts)
	a.emitLog(fmt.Sprintf("Host \"%s\" saved", name))
	a.emitHostsUpdate()
	return nil
}

func (a *App) DeleteHost(name string) error {
	a.mu.Lock()
	out := a.hosts[:0]
	for _, h := range a.hosts {
		if h.Name != name {
			out = append(out, h)
		}
	}
	a.hosts = out
	hosts := make([]SavedHost, len(a.hosts))
	copy(hosts, a.hosts)
	a.mu.Unlock()

	SaveHosts(hosts)
	a.emitLog(fmt.Sprintf("Host \"%s\" deleted", name))
	a.emitHostsUpdate()
	return nil
}

func (a *App) emitHostsUpdate() {
	a.mu.Lock()
	hosts := make([]SavedHost, len(a.hosts))
	copy(hosts, a.hosts)
	a.mu.Unlock()
	runtime.EventsEmit(a.ctx, "hosts", hosts)
}

func (a *App) tunnelLabel() string {
	a.mu.Lock()
	count := len(a.tunnels)
	a.mu.Unlock()
	if count == 0 {
		return "FishTun"
	}
	return fmt.Sprintf("FishTun — %d tunnel(s)", count)
}

func (a *App) setTitleAndTooltip() {
	label := a.tunnelLabel()
	runtime.WindowSetTitle(a.ctx, label)
	systray.SetTooltip(label)
}

func (a *App) emitLog(msg string) {
	runtime.EventsEmit(a.ctx, "log", msg)
}

func (a *App) emitStatus(status string) {
	runtime.EventsEmit(a.ctx, "status", status)
}

func (a *App) emitError(msg string) {
	runtime.EventsEmit(a.ctx, "error", msg)
}

func (a *App) emitTunnelsUpdate() {
	a.mu.Lock()
	infos := make([]TunnelInfo, 0, len(a.tunnelInfo))
	for _, info := range a.tunnelInfo {
		infos = append(infos, info)
	}
	a.mu.Unlock()
	runtime.EventsEmit(a.ctx, "tunnels", infos)
}

func (a *App) Connect(sshUser, sshHost string, sshPort, localPort, remotePort int, password, localHost, remoteCmd, disconnectCmd string) error {
	cfg := Config{
		SSHUser:       sshUser,
		SSHHost:       sshHost,
		SSHPort:       sshPort,
		LocalPort:     localPort,
		RemotePort:    remotePort,
		LocalHost:     localHost,
		RemoteCmd:     remoteCmd,
		DisconnectCmd: disconnectCmd,
	}

	tunnelID := tunnelID(localHost, localPort)

	a.emitLog(fmt.Sprintf("Creating tunnel %s ...", tunnelID))
	a.emitStatus("connecting")

	a.config.LocalHost = localHost

	tunnel := NewTunnel(cfg)
	tunnel.LogFunc = func(format string, args ...any) {
		a.emitLog(fmt.Sprintf("[%s] %s", tunnelID, fmt.Sprintf(format, args...)))
	}

	if err := tunnel.Start(password); err != nil {
		a.emitLog(fmt.Sprintf("[%s] Connection failed: %v", tunnelID, err))
		a.emitError(fmt.Sprintf("Connection failed: %v", err))
		a.emitStatus("disconnected")
		return err
	}

	a.emitLog(fmt.Sprintf("[%s] SSH connection established, waiting for remote port ...", tunnelID))
	a.emitStatus("waiting")

	if err := waitForLocalPort(cfg.LocalHost, cfg.LocalPort, 30); err != nil {
		a.emitLog(fmt.Sprintf("[%s] Port wait failed: %v", tunnelID, err))
		tunnel.Stop()
		a.emitError(fmt.Sprintf("Port wait failed: %v", err))
		a.emitStatus("disconnected")
		return err
	}

	a.emitLog(fmt.Sprintf("[%s] Tunnel established: %s:%d", tunnelID, cfg.LocalHost, cfg.LocalPort))

	a.mu.Lock()
	a.tunnels[tunnelID] = tunnel
	a.tunnelInfo[tunnelID] = TunnelInfo{
		ID:            tunnelID,
		LocalHost:     cfg.LocalHost,
		LocalPort:     cfg.LocalPort,
		RemotePort:    cfg.RemotePort,
		SSHUser:       sshUser,
		SSHHost:       sshHost,
		SSHPort:       sshPort,
		RemoteCmd:     cfg.RemoteCmd,
		DisconnectCmd: cfg.DisconnectCmd,
	}
	a.mu.Unlock()

	a.emitTunnelsUpdate()
	a.emitStatus("connected")
	a.setTitleAndTooltip()

	go func() {
		tunnel.Wait()
		a.mu.Lock()
		delete(a.tunnels, tunnelID)
		delete(a.tunnelInfo, tunnelID)
		count := len(a.tunnels)
		a.mu.Unlock()
		a.emitLog(fmt.Sprintf("[%s] Tunnel connection lost", tunnelID))
		a.emitTunnelsUpdate()
		if count == 0 {
			a.emitStatus("disconnected")
		}
		a.setTitleAndTooltip()
	}()

	return nil
}

func tunnelID(localHost string, localPort int) string {
	return fmt.Sprintf("%s:%d", localHost, localPort)
}

func (a *App) Disconnect(tunnelID string) error {
	a.mu.Lock()
	tunnel, ok := a.tunnels[tunnelID]
	if !ok {
		a.mu.Unlock()
		return nil
	}
	delete(a.tunnels, tunnelID)
	delete(a.tunnelInfo, tunnelID)
	count := len(a.tunnels)
	a.mu.Unlock()

	a.emitLog(fmt.Sprintf("[%s] Disconnecting by user", tunnelID))
	tunnel.Stop()

	a.emitTunnelsUpdate()
	if count == 0 {
		a.emitStatus("disconnected")
	}
	a.setTitleAndTooltip()
	return nil
}

func (a *App) DisconnectAll() {
	a.mu.Lock()
	tunnels := make(map[string]*Tunnel)
	for id, t := range a.tunnels {
		tunnels[id] = t
	}
	a.tunnels = make(map[string]*Tunnel)
	a.tunnelInfo = make(map[string]TunnelInfo)
	a.mu.Unlock()

	for id, tunnel := range tunnels {
		a.emitLog(fmt.Sprintf("[%s] Disconnecting by user", id))
		tunnel.Stop()
	}

	a.emitTunnelsUpdate()
	a.emitStatus("disconnected")
	a.setTitleAndTooltip()
}

func (a *App) GetTunnels() []TunnelInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	infos := make([]TunnelInfo, 0, len(a.tunnelInfo))
	for _, info := range a.tunnelInfo {
		infos = append(infos, info)
	}
	return infos
}

func (a *App) OpenBrowser(tunnelID string) error {
	a.mu.Lock()
	info, ok := a.tunnelInfo[tunnelID]
	a.mu.Unlock()
	if !ok {
		return fmt.Errorf("tunnel %s not found", tunnelID)
	}
	targetURL := fmt.Sprintf("http://%s:%d", info.LocalHost, info.LocalPort)
	a.emitLog(fmt.Sprintf("[%s] Opening browser: %s", tunnelID, targetURL))
	return openURL(targetURL)
}

func (a *App) IsConnected() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.tunnels) > 0
}

func (a *App) GetConnectedInfo() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.tunnels) == 0 {
		return ""
	}
	var result string
	for _, info := range a.tunnelInfo {
		result += fmt.Sprintf("Host: %s@%s:%d\nTunnel: %s:%d \u2194 remote 127.0.0.1:%d\n\n",
			info.SSHUser, info.SSHHost, info.SSHPort,
			info.LocalHost, info.LocalPort, info.RemotePort)
	}
	return result
}

func (a *App) Config() Config {
	return a.config
}

func (a *App) MinimizeToTray() {
	a.HideToTray()
}
