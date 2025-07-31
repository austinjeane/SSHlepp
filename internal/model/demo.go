package model

import (
	"time"

	"sshlepp/internal/ssh"
)

// CreateDemoFiles creates some demo files for testing
func CreateDemoFiles() []ssh.FileInfo {
	return []ssh.FileInfo{
		{Name: "..", IsDir: true, Size: 0, ModTime: time.Now()},
		{Name: "Documents", IsDir: true, Size: 0, ModTime: time.Now().Add(-time.Hour)},
		{Name: "Downloads", IsDir: true, Size: 0, ModTime: time.Now().Add(-2 * time.Hour)},
		{Name: "Pictures", IsDir: true, Size: 0, ModTime: time.Now().Add(-3 * time.Hour)},
		{Name: "README.md", IsDir: false, Size: 1024, ModTime: time.Now().Add(-time.Hour)},
		{Name: "config.json", IsDir: false, Size: 512, ModTime: time.Now().Add(-2 * time.Hour)},
		{Name: "script.sh", IsDir: false, Size: 256, ModTime: time.Now().Add(-3 * time.Hour)},
	}
}
