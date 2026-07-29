package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type Tunnel struct {
	config        Config
	sshClient     *ssh.Client
	listener      net.Listener
	done          chan struct{}
	remoteSession *ssh.Session
	LogFunc       func(format string, args ...any)
}

func NewTunnel(cfg Config) *Tunnel {
	return &Tunnel{
		config: cfg,
		done:   make(chan struct{}),
	}
}

func (t *Tunnel) Start(password string) error {
	addr := net.JoinHostPort(t.config.SSHHost, strconv.Itoa(t.config.SSHPort))
	cfg := &ssh.ClientConfig{
		User:            t.config.SSHUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	home, _ := os.UserHomeDir()
	keyPath := filepath.Join(home, ".ssh", "id_rsa")
	key, err := os.ReadFile(keyPath)
	if err == nil {
		signer, signerErr := ssh.ParsePrivateKey(key)
		if signerErr == nil {
			cfg.Auth = append(cfg.Auth, ssh.PublicKeys(signer))
		}
	}

	if password != "" {
		cfg.Auth = append(cfg.Auth, ssh.Password(password))
	}

	if password == "" && len(cfg.Auth) == 0 {
		fmt.Print("SSH password: ")
		passBytes, _ := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		cfg.Auth = append(cfg.Auth, ssh.Password(string(passBytes)))
	}

	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return fmt.Errorf("SSH dial failed: %w", err)
	}
	t.sshClient = client

	localAddr := fmt.Sprintf("%s:%d", t.config.LocalHost, t.config.LocalPort)
	l, err := net.Listen("tcp", localAddr)
	if err != nil {
		t.sshClient.Close()
		return fmt.Errorf("local listen failed: %w", err)
	}
	t.listener = l

	go t.acceptLoop()
	go t.startRemoteCommand()
	go t.keepaliveLoop()

	t.LogFunc("Tunnel open: %s:%d <-> %s@%s:%d (remote 127.0.0.1:%d)",
		t.config.LocalHost, t.config.LocalPort, t.config.SSHUser, t.config.SSHHost,
		t.config.SSHPort, t.config.RemotePort)

	return nil
}

func (t *Tunnel) acceptLoop() {
	for {
		localConn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.done:
				return
			default:
				t.LogFunc("Accept error: %v", err)
				return
			}
		}
		go t.handleConnection(localConn)
	}
}

func (t *Tunnel) handleConnection(localConn net.Conn) {
	defer localConn.Close()

	remoteAddr := fmt.Sprintf("127.0.0.1:%d", t.config.RemotePort)
	remoteConn, err := t.sshClient.Dial("tcp", remoteAddr)
	if err != nil {
		t.LogFunc("Remote dial error: %v", err)
		return
	}
	defer remoteConn.Close()

	done := make(chan struct{}, 2)
	go func() {
		io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()
	<-done
}

func (t *Tunnel) startRemoteCommand() {
	cmd := t.config.RemoteCmd
	if cmd == "" {
		t.LogFunc("No remote command configured, skipping")
		return
	}

	session, err := t.sshClient.NewSession()
	if err != nil {
		t.LogFunc("Remote command session error: %v", err)
		return
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 80, 40, modes); err != nil {
		t.LogFunc("RequestPty error: %v", err)
		session.Close()
		return
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Start(cmd); err != nil {
		t.LogFunc("Remote command start failed: %v", err)
		session.Close()
		return
	}

	t.remoteSession = session
	t.LogFunc("Remote command started: %s", cmd)

	go func() {
		err := session.Wait()
		t.LogFunc("Remote command exited: %v", err)
	}()
}

func (t *Tunnel) stopRemoteCommand() {
	if t.config.DisconnectCmd != "" {
		t.LogFunc("Running disconnect command: %s", t.config.DisconnectCmd)
		session, err := t.sshClient.NewSession()
		if err != nil {
			t.LogFunc("Disconnect session error: %v", err)
		} else {
			if err := session.Run(t.config.DisconnectCmd); err != nil {
				t.LogFunc("Disconnect command exited: %v", err)
			}
			session.Close()
		}
	}

	if t.remoteSession != nil {
		t.remoteSession.Close()
		t.remoteSession = nil
	}
}

func (t *Tunnel) keepaliveLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_, _, err := t.sshClient.SendRequest("keepalive@openssh.com", true, nil)
			if err != nil {
				t.LogFunc("Keepalive failed, tunnel may be down: %v", err)
				t.Stop()
				return
			}
		case <-t.done:
			return
		}
	}
}

func (t *Tunnel) Stop() {
	select {
	case <-t.done:
		return
	default:
		close(t.done)
	}

	t.stopRemoteCommand()

	if t.listener != nil {
		t.listener.Close()
	}
	if t.sshClient != nil {
		t.sshClient.Close()
	}
	t.LogFunc("Tunnel stopped")
}

func (t *Tunnel) Wait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigChan:
		fmt.Println()
		t.LogFunc("Shutting down...")
		t.Stop()
	case <-t.done:
	}
}
