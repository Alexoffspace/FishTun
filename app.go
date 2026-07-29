package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/getlantern/systray"
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
	ctx              context.Context
	config           Config
	tunnels          map[string]*Tunnel
	tunnelInfo       map[string]TunnelInfo
	mu               sync.Mutex
	iconData         []byte
	makeMoreCh       chan struct{}
	showWindowCh     chan struct{}
	disconnectAllCh  chan struct{}
	quitCh           chan struct{}
	trayStarted      bool
	trayMu           sync.Mutex
	tooltipCh        chan string
}

func NewApp() *App {
	return &App{
		config:    LoadConfig(),
		tunnels:   make(map[string]*Tunnel),
		tunnelInfo: make(map[string]TunnelInfo),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.makeMoreCh = make(chan struct{}, 10)
	a.showWindowCh = make(chan struct{}, 10)
	a.disconnectAllCh = make(chan struct{}, 10)
	a.quitCh = make(chan struct{}, 10)
	a.tooltipCh = make(chan string, 1)

	go a.handleTrayEvents()
}

func (a *App) handleTrayEvents() {
	for {
		select {
		case <-a.makeMoreCh:
			runtime.Show(a.ctx)
		case <-a.showWindowCh:
			runtime.Show(a.ctx)
		case <-a.disconnectAllCh:
			a.DisconnectAll()
		case <-a.quitCh:
			a.DisconnectAll()
			go a.quit()
			return
		}
	}
}

func (a *App) startTray() {
	a.trayMu.Lock()
	if a.trayStarted {
		a.trayMu.Unlock()
		return
	}
	a.trayStarted = true
	a.trayMu.Unlock()

	go runTray(a.iconData, a.makeMoreCh, a.showWindowCh, a.disconnectAllCh, a.quitCh, a.tooltipCh)
}

func (a *App) quit() {
	runtime.Quit(a.ctx)
	systray.Quit()
}

func (a *App) GetConfig() Config {
	return a.config
}

func (a *App) SaveConfig(cfg Config) {
	a.config = cfg
	SaveConfig(cfg)
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
	select {
	case a.tooltipCh <- label:
	default:
	}
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

	a.startTray()

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
	runtime.Hide(a.ctx)
}
