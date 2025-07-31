package model

import (
	"fmt"
	"path/filepath"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"sshlepp/internal/ssh"
	"sshlepp/internal/ui"
)

// PanelSide represents which panel is focused
type PanelSide int

const (
	LeftPanel PanelSide = iota
	RightPanel
)

// fileBrowserModel handles dual-panel file browsing
type fileBrowserModel struct {
	localFiles     []ssh.FileInfo
	remoteFiles    []ssh.FileInfo
	localCursor    int
	remoteCursor   int
	localSelected  map[int]bool
	remoteSelected map[int]bool
	focusedPanel   PanelSide
	localPath      string
	remotePath     string
	sshClient      *ssh.Client
	width, height  int
	err            error
}

// Messages
type loadFilesMsg struct {
	localFiles  []ssh.FileInfo
	remoteFiles []ssh.FileInfo
	err         error
}

type errMsg struct {
	err error
}

// newFileBrowserModel creates a new file browser model
func newFileBrowserModel(host *ssh.SSHHost, width, height int) (*fileBrowserModel, tea.Cmd) {
	// Get current working directory
	localPath, err := os.Getwd()
	if err != nil {
		localPath = "."
	}

	model := &fileBrowserModel{
		localSelected:  make(map[int]bool),
		remoteSelected: make(map[int]bool),
		focusedPanel:   LeftPanel,
		localPath:      localPath,
		remotePath:     "/", // Start at root for remote
		width:          width,
		height:         height,
	}

	// Connect to SSH and load initial files
	return model, tea.Batch(
		connectSSHCmd(host),
		loadFilesCmd(model),
	)
}

// Commands
func connectSSHCmd(host *ssh.SSHHost) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		client, err := ssh.NewClient(*host)
		if err != nil {
			return errMsg{err}
		}
		return sshConnectedMsg{client}
	})
}

func loadFilesCmd(m *fileBrowserModel) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Load local files
		localFiles, localErr := ssh.ListLocalDir(m.localPath)
		if localErr != nil {
			return errMsg{localErr}
		}

		var remoteFiles []ssh.FileInfo
		var remoteErr error
		if m.sshClient != nil {
			remoteFiles, remoteErr = m.sshClient.ListDir(m.remotePath)
			if remoteErr != nil {
				return errMsg{remoteErr}
			}
		}

		return loadFilesMsg{
			localFiles:  localFiles,
			remoteFiles: remoteFiles,
		}
	})
}

type sshConnectedMsg struct {
	client *ssh.Client
}

// Init initializes the file browser
func (m *fileBrowserModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *fileBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sshConnectedMsg:
		m.sshClient = msg.client
		return m, loadFilesCmd(m)

	case loadFilesMsg:
		m.localFiles = msg.localFiles
		m.remoteFiles = msg.remoteFiles
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *fileBrowserModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		// Switch focused panel
		if m.focusedPanel == LeftPanel {
			m.focusedPanel = RightPanel
		} else {
			m.focusedPanel = LeftPanel
		}

	case "up", "k":
		if m.focusedPanel == LeftPanel {
			if m.localCursor > 0 {
				m.localCursor--
			}
		} else {
			if m.remoteCursor > 0 {
				m.remoteCursor--
			}
		}

	case "down", "j":
		if m.focusedPanel == LeftPanel {
			if m.localCursor < len(m.localFiles)-1 {
				m.localCursor++
			}
		} else {
			if m.remoteCursor < len(m.remoteFiles)-1 {
				m.remoteCursor++
			}
		}

	case "enter":
		return m.handleEnterDirectory()

	case "left", "h":
		return m.handleGoUpDirectory()

	case " ":
		// Toggle selection
		if m.focusedPanel == LeftPanel {
			m.localSelected[m.localCursor] = !m.localSelected[m.localCursor]
		} else {
			m.remoteSelected[m.remoteCursor] = !m.remoteSelected[m.remoteCursor]
		}

	case "c":
		// Copy selected files
		return m.handleCopy()
	}

	return m, nil
}

// handleEnterDirectory handles entering a directory
func (m *fileBrowserModel) handleEnterDirectory() (tea.Model, tea.Cmd) {
	if m.focusedPanel == LeftPanel {
		if m.localCursor < len(m.localFiles) && m.localFiles[m.localCursor].IsDir {
			if m.localFiles[m.localCursor].Name == ".." {
				m.localPath = filepath.Dir(m.localPath)
			} else {
				m.localPath = filepath.Join(m.localPath, m.localFiles[m.localCursor].Name)
			}
			m.localCursor = 0
			return m, loadFilesCmd(m)
		}
	} else {
		if m.remoteCursor < len(m.remoteFiles) && m.remoteFiles[m.remoteCursor].IsDir {
			if m.remoteFiles[m.remoteCursor].Name == ".." {
				m.remotePath = filepath.Dir(m.remotePath)
			} else {
				m.remotePath = filepath.Join(m.remotePath, m.remoteFiles[m.remoteCursor].Name)
			}
			m.remoteCursor = 0
			return m, loadFilesCmd(m)
		}
	}
	return m, nil
}

// handleGoUpDirectory handles going up one directory level
func (m *fileBrowserModel) handleGoUpDirectory() (tea.Model, tea.Cmd) {
	if m.focusedPanel == LeftPanel {
		m.localPath = filepath.Dir(m.localPath)
		m.localCursor = 0
	} else {
		m.remotePath = filepath.Dir(m.remotePath)
		m.remoteCursor = 0
	}
	return m, loadFilesCmd(m)
}

// View renders the file browser
func (m *fileBrowserModel) View() string {
	if m.err != nil {
		return ui.ErrorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error()))
	}

	panelWidth := (m.width - 4) / 2 // Account for borders and spacing

	leftPanel := m.renderPanel(LeftPanel, panelWidth)
	rightPanel := m.renderPanel(RightPanel, panelWidth)

	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		rightPanel,
	)

	help := ui.HelpStyle.Render("tab: switch panel • ↑/↓: navigate • ←/→: go up/into dir • space: select • c: copy • q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, panels, help)
}

// renderPanel renders a single panel (local or remote)
func (m *fileBrowserModel) renderPanel(side PanelSide, width int) string {
	var files []ssh.FileInfo
	var cursor int
	var selected map[int]bool
	var path string
	var title string

	if side == LeftPanel {
		files = m.localFiles
		cursor = m.localCursor
		selected = m.localSelected
		path = m.localPath
		title = "Local"
	} else {
		files = m.remoteFiles
		cursor = m.remoteCursor
		selected = m.remoteSelected
		path = m.remotePath
		title = "Remote"
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("%s: %s\n\n", title, path))

	// Add .. entry if not at root
	if (side == LeftPanel && m.localPath != "/") || (side == RightPanel && m.remotePath != "/") {
		cursorIcon := " "
		if cursor == 0 {
			cursorIcon = ">"
		}
		style := ui.RegularRowStyle
		if cursor == 0 && m.focusedPanel == side {
			style = ui.SelectedRowStyle
		}
		content.WriteString(style.Render(fmt.Sprintf("%s [DIR] ..", cursorIcon)) + "\n")
	}

	// Render files
	for i, file := range files {
		displayIndex := i
		if (side == LeftPanel && m.localPath != "/") || (side == RightPanel && m.remotePath != "/") {
			displayIndex = i + 1
		}

		cursorIcon := " "
		if cursor == displayIndex {
			cursorIcon = ">"
		}

		selectIcon := " "
		if selected[i] {
			selectIcon = "✓"
		}

		fileType := "FILE"
		if file.IsDir {
			fileType = "DIR"
		}

		style := ui.RegularRowStyle
		if i%2 == 0 {
			style = ui.DimRowStyle
		}
		if cursor == displayIndex && m.focusedPanel == side {
			style = ui.SelectedRowStyle
		}

		line := fmt.Sprintf("%s %s [%s] %s (%d bytes)",
			cursorIcon, selectIcon, fileType, file.Name, file.Size)
		
		content.WriteString(style.Render(line) + "\n")
	}

	// Apply panel style based on focus
	panelStyle := ui.UnfocusedPanelStyle
	if m.focusedPanel == side {
		panelStyle = ui.FocusedPanelStyle
	}

	return panelStyle.Width(width).Height(m.height-4).Render(content.String())
}

// handleCopy handles copying selected files
func (m *fileBrowserModel) handleCopy() (tea.Model, tea.Cmd) {
	var selectedFiles []string
	var sourcePath, destPath string
	var isLocalToRemote bool

	if m.focusedPanel == LeftPanel {
		// Copy from local to remote
		isLocalToRemote = true
		sourcePath = m.localPath
		destPath = m.remotePath
		for i, selected := range m.localSelected {
			if selected && i < len(m.localFiles) {
				selectedFiles = append(selectedFiles, m.localFiles[i].Name)
			}
		}
	} else {
		// Copy from remote to local
		isLocalToRemote = false
		sourcePath = m.remotePath
		destPath = m.localPath
		for i, selected := range m.remoteSelected {
			if selected && i < len(m.remoteFiles) {
				selectedFiles = append(selectedFiles, m.remoteFiles[i].Name)
			}
		}
	}

	if len(selectedFiles) == 0 {
		return m, nil // No files selected
	}

	// Start copy operation
	return m, copyFilesCmd(m.sshClient, selectedFiles, sourcePath, destPath, isLocalToRemote)
}

// copyFilesCmd creates a command to copy files
func copyFilesCmd(client *ssh.Client, files []string, sourcePath, destPath string, isLocalToRemote bool) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		for i, file := range files {
			var err error
			if isLocalToRemote {
				err = client.CopyFileFromLocal(
					filepath.Join(sourcePath, file),
					filepath.Join(destPath, file),
				)
			} else {
				err = client.CopyFileToLocal(
					filepath.Join(sourcePath, file),
					filepath.Join(destPath, file),
				)
			}
			
			if err != nil {
				return errMsg{fmt.Errorf("failed to copy %s: %w", file, err)}
			}

			// Send progress update
			progress := copyProgressMsg{
				fileIndex: i,
				fileName:  file,
				percent:   float64(i+1) / float64(len(files)),
			}
			
			// For now, just continue - in a real implementation, 
			// you'd send progress updates through a channel
			_ = progress
		}

		return copyCompleteMsg{}
	})
}
