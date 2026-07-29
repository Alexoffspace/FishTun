package main

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

func openURL(urlStr string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	return exec.Command(cmd, append(args, urlStr)...).Start()
}

func waitForLocalPort(localHost string, localPort int, timeoutSec int) error {
	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	addr := net.JoinHostPort(localHost, strconv.Itoa(localPort))
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s to be reachable", addr)
}
