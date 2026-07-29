package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	configDir  = ".config/FishTun"
	configFile = "config"
)

type Config struct {
	LocalPort     int
	RemotePort    int
	SSHHost       string
	SSHUser       string
	SSHPort       int
	LocalHost     string
	RemoteCmd     string
	DisconnectCmd string
}

func DefaultConfig() Config {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	if user == "" {
		user = "user"
	}
	return Config{
		LocalPort:  4096,
		RemotePort: 4096,
		SSHHost:    "",
		SSHUser:    user,
		SSHPort:    22,
		LocalHost:  "127.0.0.1",
	}
}

func configDirPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, configDir)
}

func configFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot get home directory: %v\n", err)
		os.Exit(1)
	}
	dir := filepath.Join(home, configDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot create config directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(dir, configFile)
}

func LoadConfig() Config {
	config := DefaultConfig()
	cfgPath := configFilePath()
	f, err := os.Open(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config
		}
		return config
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		switch key {
		case "LOCAL_PORT":
			if v, err := strconv.Atoi(val); err == nil {
				config.LocalPort = v
			}
		case "REMOTE_PORT":
			if v, err := strconv.Atoi(val); err == nil {
				config.RemotePort = v
			}
		case "SSH_HOST":
			config.SSHHost = val
		case "SSH_USER":
			config.SSHUser = val
		case "SSH_PORT":
			if v, err := strconv.Atoi(val); err == nil {
				config.SSHPort = v
			}
		case "LOCAL_HOST":
			config.LocalHost = val
		case "REMOTE_CMD":
			config.RemoteCmd = val
		case "DISCONNECT_CMD":
			config.DisconnectCmd = val
		}
	}
	return config
}

func SaveConfig(cfg Config) {
	cfgPath := configFilePath()
	f, err := os.Create(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Cannot save config: %v\n", err)
		return
	}
	defer f.Close()

	content := fmt.Sprintf(`# FishTun Configuration
# Updated: %s
LOCAL_PORT="%d"
REMOTE_PORT="%d"
SSH_HOST="%s"
SSH_USER="%s"
SSH_PORT="%d"
LOCAL_HOST="%s"
REMOTE_CMD="%s"
DISCONNECT_CMD="%s"
`, time.Now().Format(time.RFC3339),
		cfg.LocalPort,
		cfg.RemotePort,
		cfg.SSHHost,
		cfg.SSHUser,
		cfg.SSHPort,
		cfg.LocalHost,
		cfg.RemoteCmd,
		cfg.DisconnectCmd,
	)
	f.WriteString(content)
}
