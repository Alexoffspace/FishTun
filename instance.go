package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// instanceFileName is the single-instance lock file, stored alongside the
// config. It holds the IPC port of the running instance.
const instanceFileName = "instance.lock"

// instanceLock holds the open lock file for the lifetime of the process.
// Releasing it (on graceful exit) allows the next launch to take over.
var instanceLock struct {
	mu sync.Mutex
	f  *os.File
}

func instanceFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(home, configDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ""
	}
	return filepath.Join(dir, instanceFileName)
}

// acquireSingleInstance tries to become the primary instance. It returns
// (true, nil) when this process owns the lock and must keep running, or
// (false, nil) when another live instance was asked to show its window and
// this process should exit immediately.
func acquireSingleInstance() (bool, error) {
	path := instanceFilePath()
	if path == "" {
		// Without a stable path we cannot coordinate; run as before.
		return true, nil
	}

	// Two attempts: the first may race with a stale lock left by a crashed
	// instance, which we recover by removing the file and retrying.
	for attempt := 0; attempt < 2; attempt++ {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
		if err == nil {
			// We own the lock: start the IPC listener and record its port.
			port, lerr := startInstanceServer()
			if lerr != nil {
				f.Close()
				os.Remove(path)
				return true, nil
			}
			f.Truncate(0)
			f.Seek(0, 0)
			f.WriteString(strconv.Itoa(port))
			f.Sync()
			instanceLock.mu.Lock()
			instanceLock.f = f
			instanceLock.mu.Unlock()
			return true, nil
		}
		if !os.IsExist(err) {
			return true, err
		}

		// Lock exists: ask the running instance to surface its window.
		port, rerr := readInstancePort(path)
		if rerr == nil && notifyInstance(port) {
			return false, nil
		}

		// Stale lock (dead process or unreadable port): drop it and retry.
		os.Remove(path)
		time.Sleep(50 * time.Millisecond)
	}

	// Could not coordinate cleanly; don't block the user's launch.
	return true, nil
}

func readInstancePort(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

// startInstanceServer listens on a loopback port that accepts "show" commands
// from secondary launches. Each command is forwarded to the tray event pump.
func startInstanceServer() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.SetReadDeadline(time.Now().Add(2 * time.Second))
				line, _ := bufio.NewReader(c).ReadString('\n')
				if strings.TrimSpace(line) == "show" {
					instanceShowCh <- struct{}{}
				}
			}(conn)
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port, nil
}

// instanceShowCh delivers "show window" requests from secondary launches into
// the App's tray event loop (buffered and non-blocking).
var instanceShowCh = make(chan struct{}, 10)

// notifyInstance connects to the primary instance and asks it to show its
// window. It returns true when the command was delivered.
func notifyInstance(port int) bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if _, err := fmt.Fprintln(conn, "show"); err != nil {
		return false
	}
	return true
}

// releaseInstance removes the lock so a future launch can take over cleanly.
func releaseInstance() {
	instanceLock.mu.Lock()
	f := instanceLock.f
	instanceLock.f = nil
	instanceLock.mu.Unlock()
	if f == nil {
		return
	}
	f.Close()
	if path := instanceFilePath(); path != "" {
		os.Remove(path)
	}
}
