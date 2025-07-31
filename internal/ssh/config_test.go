package ssh

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSSHConfig(t *testing.T) {
	// Create a temporary SSH config for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config")
	
	configContent := `
Host example
    HostName example.com
    User testuser
    Port 2222

Host localhost-test
    HostName localhost
    User root
    Port 22
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	
	// Set up environment to use test config
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	// Create .ssh directory
	sshDir := filepath.Join(tempDir, ".ssh")
	err = os.Mkdir(sshDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}
	
	// Move config to .ssh directory
	err = os.Rename(configPath, filepath.Join(sshDir, "config"))
	if err != nil {
		t.Fatalf("Failed to move config: %v", err)
	}
	
	// Test parsing
	hosts, err := ParseSSHConfig()
	if err != nil {
		t.Fatalf("ParseSSHConfig failed: %v", err)
	}
	
	if len(hosts) != 2 {
		t.Errorf("Expected 2 hosts, got %d", len(hosts))
	}
	
	// Check first host
	if hosts[0].Name != "example" {
		t.Errorf("Expected host name 'example', got '%s'", hosts[0].Name)
	}
	
	if hosts[0].Hostname != "example.com" {
		t.Errorf("Expected hostname 'example.com', got '%s'", hosts[0].Hostname)
	}
	
	if hosts[0].User != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", hosts[0].User)
	}
	
	if hosts[0].Port != 2222 {
		t.Errorf("Expected port 2222, got %d", hosts[0].Port)
	}
}

func TestParseSSHConfigNoFile(t *testing.T) {
	// Set up environment with no config file
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	hosts, err := ParseSSHConfig()
	if err != nil {
		t.Fatalf("ParseSSHConfig should not fail when no config exists: %v", err)
	}
	
	if len(hosts) != 0 {
		t.Errorf("Expected 0 hosts when no config exists, got %d", len(hosts))
	}
}
