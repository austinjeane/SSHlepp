package ssh

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SSHHost represents a single SSH host configuration
type SSHHost struct {
	Name     string
	Hostname string
	User     string
	Port     int
}

// ParseSSHConfig parses the SSH config file and returns available hosts
func ParseSSHConfig() ([]SSHHost, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".ssh", "config")
	file, err := os.Open(configPath)
	if err != nil {
		// If config doesn't exist, return empty list
		if os.IsNotExist(err) {
			return []SSHHost{}, nil
		}
		return nil, fmt.Errorf("failed to open SSH config: %w", err)
	}
	defer file.Close()

	var hosts []SSHHost
	var currentHost *SSHHost

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		switch key {
		case "host":
			// Save previous host if exists
			if currentHost != nil && currentHost.Name != "" {
				hosts = append(hosts, *currentHost)
			}
			// Start new host
			currentHost = &SSHHost{
				Name: value,
				Port: 22, // default port
				User: "root", // default user
			}

		case "hostname":
			if currentHost != nil {
				currentHost.Hostname = value
			}

		case "user":
			if currentHost != nil {
				currentHost.User = value
			}

		case "port":
			if currentHost != nil {
				if port, err := strconv.Atoi(value); err == nil {
					currentHost.Port = port
				}
			}
		}
	}

	// Add the last host
	if currentHost != nil && currentHost.Name != "" {
		hosts = append(hosts, *currentHost)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading SSH config: %w", err)
	}

	return hosts, nil
}

// String returns a formatted string representation of the host
func (h SSHHost) String() string {
	return fmt.Sprintf("%s@%s:%d", h.User, h.Hostname, h.Port)
}
