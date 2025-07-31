package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// FileInfo represents a file or directory
type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// Client wraps SSH and SFTP clients
type Client struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	host       SSHHost
}

// NewClient creates a new SSH/SFTP client
func NewClient(host SSHHost) (*Client, error) {
	// Try to read SSH private key
	key, err := readPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: host.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key checking
		Timeout:         30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host.Hostname, host.Port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return &Client{
		sshClient:  sshClient,
		sftpClient: sftpClient,
		host:       host,
	}, nil
}

// Close closes the SSH and SFTP connections
func (c *Client) Close() error {
	if c.sftpClient != nil {
		c.sftpClient.Close()
	}
	if c.sshClient != nil {
		return c.sshClient.Close()
	}
	return nil
}

// ListDir lists files in a remote directory
func (c *Client) ListDir(path string) ([]FileInfo, error) {
	files, err := c.sftpClient.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	result := make([]FileInfo, len(files))
	for i, file := range files {
		result[i] = FileInfo{
			Name:    file.Name(),
			Size:    file.Size(),
			ModTime: file.ModTime(),
			IsDir:   file.IsDir(),
		}
	}

	return result, nil
}

// readPrivateKey reads the SSH private key
func readPrivateKey() (ssh.Signer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	keyPaths := []string{
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
	}

	for _, keyPath := range keyPaths {
		if _, err := os.Stat(keyPath); err == nil {
			key, err := os.ReadFile(keyPath)
			if err != nil {
				continue
			}

			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				continue
			}

			return signer, nil
		}
	}

	return nil, fmt.Errorf("no valid private key found")
}

// ListLocalDir lists files in a local directory
func ListLocalDir(path string) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list local directory: %w", err)
	}

	result := make([]FileInfo, len(entries))
	for i, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		result[i] = FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
		}
	}

	return result, nil
}
