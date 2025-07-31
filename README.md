# SSHlepp - Terminal SSH File Browser & Copy Tool

A powerful terminal UI application for browsing and copying files between local and remote servers via SSH, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![SSHlepp Demo](https://via.placeholder.com/800x400/1e1e1e/ffffff?text=SSHlepp+Demo)

## ✨ Features

- **Dual-Panel Interface**: Side-by-side local and remote file views
- **SSH Config Integration**: Automatically loads servers from `~/.ssh/config`
- **Multi-File Selection**: Select multiple files with space bar
- **Intuitive Navigation**: Tab to switch panels, arrow keys to navigate
- **File Copy Operations**: Copy files between local and remote with progress display
- **SSH Key Authentication**: Supports standard SSH key authentication
- **Responsive Design**: Adapts to terminal size

## 🚀 Quick Start

### Prerequisites

- Go 1.19 or later
- SSH client configured with key-based authentication
- Access to remote servers via SSH

### Installation

```bash
# Clone the repository
git clone https://github.com/austinjeane/SSHlepp.git
cd SSHlepp

# Build the application
go build -o sshlepp ./cmd/sshlepp

# Run SSHlepp
./sshlepp
```

### Setup SSH Configuration

Ensure you have SSH hosts configured in `~/.ssh/config`:

```
Host myserver
    HostName example.com
    User myuser
    Port 22

Host development
    HostName dev.company.com
    User developer
    Port 2222
```

## 🎮 Controls

| Key | Action |
|-----|--------|
| `↑/↓` or `k/j` | Navigate files |
| `Tab` | Switch between panels |
| `Enter` | Enter directory |
| `←/→` or `h/l` | Go up directory |
| `Space` | Select/deselect file |
| `c` | Copy selected files to other panel |
| `q` or `Ctrl+C` | Quit application |

## 🏗️ Architecture

The application is built with a clean, modular architecture:

```
├── cmd/sshlepp/          # Main application entry point
├── internal/
│   ├── model/            # Bubble Tea models
│   │   ├── main.go       # Main application state machine
│   │   ├── server_select.go # Server selection screen
│   │   ├── file_browser.go  # Dual-panel file browser
│   │   └── copy_progress.go # Copy progress display
│   ├── ssh/              # SSH and SFTP functionality
│   │   ├── config.go     # SSH config parsing
│   │   ├── sftp.go       # SFTP client operations
│   │   └── copy.go       # File copy operations
│   └── ui/               # UI styles and components
│       └── styles.go     # Lipgloss styling
```

### Key Components

- **Server Selection**: Parses SSH config and presents available hosts
- **File Browser**: Dual-panel interface with independent navigation
- **Copy Engine**: Handles file transfers with progress tracking
- **SSH Integration**: Manages connections and SFTP operations

## 🔧 Configuration

### SSH Key Support

SSHlepp automatically looks for SSH keys in the following order:
1. `~/.ssh/id_rsa`
2. `~/.ssh/id_ed25519`
3. `~/.ssh/id_ecdsa`

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SSH_CONFIG_PATH` | Path to SSH config file | `~/.ssh/config` |

## 🛠️ Development

### Running Tests

```bash
go test ./...
```

### Building for Different Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o sshlepp-linux ./cmd/sshlepp

# macOS
GOOS=darwin GOARCH=amd64 go build -o sshlepp-macos ./cmd/sshlepp

# Windows
GOOS=windows GOARCH=amd64 go build -o sshlepp.exe ./cmd/sshlepp
```

### Code Structure

The application follows the Bubble Tea pattern with three main components:

1. **Model**: Holds application state
2. **Update**: Handles messages and state transitions  
3. **View**: Renders the UI

Each screen (server selection, file browser) is implemented as a separate model that can be composed together.

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Commit your changes: `git commit -am 'Add feature'`
5. Push to the branch: `git push origin feature-name`
6. Submit a pull request

### Code Guidelines

- Follow Go best practices and conventions
- Add tests for new functionality
- Update documentation for user-facing changes
- Use meaningful commit messages

## 🐛 Troubleshooting

### Common Issues

**Connection Refused**
- Ensure SSH server is running on the target host
- Check firewall settings and port configuration
- Verify SSH key authentication is set up correctly

**Permission Denied**
- Ensure your SSH key is added to the server's authorized_keys
- Check file permissions on SSH keys (should be 600)
- Verify the correct username in SSH config

**Files Not Displaying**
- Check directory permissions on both local and remote
- Ensure SFTP subsystem is enabled on the SSH server

### Debug Mode

Run with verbose logging:
```bash
SSHLEPP_DEBUG=1 ./sshlepp
```

## 📋 Roadmap

- [ ] Resume interrupted transfers
- [ ] Background copy operations
- [ ] File preview functionality
- [ ] Directory synchronization
- [ ] Bookmark management
- [ ] Custom themes and styling
- [ ] File search and filtering
- [ ] Batch rename operations

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The TUI framework that powers SSHlepp
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling and layout
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Go SSH](https://golang.org/x/crypto/ssh) - SSH client implementation
- [SFTP](https://github.com/pkg/sftp) - SFTP client library

## 📞 Support

- Create an [issue](https://github.com/austinjeane/SSHlepp/issues) for bug reports
- Start a [discussion](https://github.com/austinjeane/SSHlepp/discussions) for questions
- Follow [@austinjeane](https://github.com/austinjeane) for updates

---

**Made with ❤️ and Go**
