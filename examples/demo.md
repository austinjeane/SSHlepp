# SSHlepp Demo

This document provides a walkthrough of SSHlepp's features and usage.

## Server Selection

When you first run SSHlepp, you'll see a list of SSH servers parsed from your `~/.ssh/config` file:

```
Select an SSH server:

> example (demo@example.com:22)
  localhost-test (dog@localhost:22)

↑/↓: navigate • enter: select • q: quit
```

Use the arrow keys to navigate and press Enter to connect to a server.

## File Browser

After selecting a server, you'll see the dual-panel file browser:

```
┌─ Local: /home/user/documents ──────────┐ ┌─ Remote: /home/demo ─────────────────┐
│                                        │ │                                      │
│ > ✓ [DIR] Documents                    │ │   [DIR] ..                           │
│     [DIR] Downloads                    │ │   [DIR] bin                          │
│     [DIR] Pictures                     │ │   [FILE] .bashrc (1024 bytes)       │
│     [FILE] README.md (1024 bytes)      │ │   [FILE] .profile (512 bytes)       │
│     [FILE] config.json (512 bytes)     │ │ > [FILE] script.sh (256 bytes)       │
│                                        │ │                                      │
└────────────────────────────────────────┘ └──────────────────────────────────────┘

tab: switch panel • ↑/↓: navigate • ←/→: go up/into dir • space: select • c: copy • q: quit
```

### Navigation

- **Tab**: Switch between left (local) and right (remote) panels
- **↑/↓**: Navigate through files and directories
- **Enter**: Enter a directory
- **←/→**: Go up one directory level
- **Space**: Select/deselect files for copying

### File Operations

- **c**: Copy selected files from the focused panel to the other panel
- Selected files are marked with a ✓ checkmark
- Progress is shown during copy operations

## Example Workflow

1. Start SSHlepp: `./sshlepp`
2. Select your server from the list
3. Navigate to the desired directories in both panels
4. Select files to copy using Space
5. Press 'c' to start the copy operation
6. Watch the progress as files are transferred

## Tips for Best Experience

- Set up SSH key authentication for seamless connections
- Use meaningful host names in your SSH config
- Test connections manually before using SSHlepp
- Keep file permissions in mind when copying
