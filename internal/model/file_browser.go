package model

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"sshlepp/internal/ssh"
	"sshlepp/internal/ui"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	ready          bool
	leftViewport   viewport.Model
	rightViewport  viewport.Model
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

// newFileBrowserModel creates a new file browser model with an existing SSH client
func newFileBrowserModel(client *ssh.Client, host *ssh.SSHHost, width, height int) (*fileBrowserModel, tea.Cmd) {
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
		remotePath:     "/",    // Start at root for remote
		sshClient:      client, // Use the provided client
		width:          width,
		height:         height,
		ready:          false,
	}

	// Initialize viewports immediately since we have width and height
	model.initializeViewports()

	// Load initial files since we already have the SSH client
	return model, loadFilesCmd(model)
}

// initializeViewports initializes the viewports with current dimensions
func (m *fileBrowserModel) initializeViewports() {
	headerHeight := 3                                         // Header with path info
	footerHeight := 2                                         // Footer with scroll info
	panelWidth := (m.width - 4) / 2                           // Account for borders and spacing
	panelHeight := m.height - headerHeight - footerHeight - 4 // Account for help text

	// Initialize viewports
	m.leftViewport = viewport.New(panelWidth-2, panelHeight) // Account for padding
	m.rightViewport = viewport.New(panelWidth-2, panelHeight)
	m.ready = true

	// Set initial content if files are already loaded
	if len(m.localFiles) > 0 || len(m.remoteFiles) > 0 {
		m.updateViewportContent()
	}
}

// Commands
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

// Init initializes the file browser
func (m *fileBrowserModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *fileBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case loadFilesMsg:
		m.localFiles = msg.localFiles
		m.remoteFiles = msg.remoteFiles
		if msg.err != nil {
			m.err = msg.err
		}
		// Clear selections when loading new files (changing directories)
		m.localSelected = make(map[int]bool)
		m.remoteSelected = make(map[int]bool)
		// Update viewport content after loading files
		if m.ready {
			m.updateViewportContent()
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// Initialize viewports if not ready yet
			m.initializeViewports()
			m.updateViewportContent()
		} else {
			// Resize existing viewports
			headerHeight := 3                                         // Header with path info
			footerHeight := 2                                         // Footer with scroll info
			panelWidth := (m.width - 4) / 2                           // Account for borders and spacing
			panelHeight := m.height - headerHeight - footerHeight - 4 // Account for help text

			m.leftViewport.Width = panelWidth - 2
			m.leftViewport.Height = panelHeight
			m.rightViewport.Width = panelWidth - 2
			m.rightViewport.Height = panelHeight
		}

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	// Handle viewport updates for scrolling
	if m.ready {
		if m.focusedPanel == LeftPanel {
			m.leftViewport, cmd = m.leftViewport.Update(msg)
		} else {
			m.rightViewport, cmd = m.rightViewport.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateViewportContent updates the content of both viewports
func (m *fileBrowserModel) updateViewportContent() {
	if !m.ready {
		return
	}

	m.leftViewport.SetContent(m.generatePanelContent(LeftPanel))
	m.rightViewport.SetContent(m.generatePanelContent(RightPanel))
}

// ensureCursorVisible scrolls the viewport to ensure the cursor is visible
func (m *fileBrowserModel) ensureCursorVisible(side PanelSide) {
	if !m.ready {
		return
	}

	var cursor int
	var viewport *viewport.Model

	if side == LeftPanel {
		cursor = m.localCursor
		viewport = &m.leftViewport
	} else {
		cursor = m.remoteCursor
		viewport = &m.rightViewport
	}

	// Calculate line position (cursor + header lines)
	linePos := cursor + 2 // Account for header lines

	// Scroll to make cursor visible
	if linePos < viewport.YOffset {
		viewport.YOffset = linePos
	} else if linePos >= viewport.YOffset+viewport.Height {
		viewport.YOffset = linePos - viewport.Height + 1
	}
}

// generatePanelContent generates the content for a panel that will be displayed in a viewport
func (m *fileBrowserModel) generatePanelContent(side PanelSide) string {
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
	if (side == LeftPanel && !isLocalRoot(m.localPath)) || (side == RightPanel && m.remotePath != "/") {
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
		if (side == LeftPanel && !isLocalRoot(m.localPath)) || (side == RightPanel && m.remotePath != "/") {
			displayIndex = i + 1
		}

		cursorIcon := " "
		if cursor == displayIndex {
			cursorIcon = ">"
			// DEBUG: Show which file has the cursor
			if side == LeftPanel {
				fmt.Printf("DEBUG CURSOR: Cursor on file[%d]: %s (displayIndex=%d, cursor=%d)\n", i, file.Name, displayIndex, cursor)
			}
		}

		selectIcon := " "
		if selected[i] {
			selectIcon = "✓"
			// DEBUG: Show which file is getting the checkmark
			if side == LeftPanel {
				fmt.Printf("DEBUG DISPLAY: Showing checkmark for file[%d]: %s (displayIndex=%d)\n", i, file.Name, displayIndex)
			}
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

	return content.String()
}

// headerView creates a header for a panel similar to the pager example
func (m *fileBrowserModel) headerView(side PanelSide) string {
	var title string
	var panelWidth int

	if side == LeftPanel {
		title = "Local Files"
		panelWidth = m.leftViewport.Width
	} else {
		title = "Remote Files"
		panelWidth = m.rightViewport.Width
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true).
		Padding(0, 1)

	renderedTitle := titleStyle.Render(title)
	line := strings.Repeat("─", max(0, panelWidth-lipgloss.Width(renderedTitle)))
	return lipgloss.JoinHorizontal(lipgloss.Center, renderedTitle, line)
}

// footerView creates a footer showing scroll position similar to the pager example
func (m *fileBrowserModel) footerView(side PanelSide) string {
	var viewport viewport.Model
	var panelWidth int

	if side == LeftPanel {
		viewport = m.leftViewport
		panelWidth = m.leftViewport.Width
	} else {
		viewport = m.rightViewport
		panelWidth = m.rightViewport.Width
	}

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 1)

	info := infoStyle.Render(fmt.Sprintf("%3.f%%", viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, panelWidth-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// remotePathDir returns the parent directory of a remote path using Unix-style separators
func remotePathDir(remotePath string) string {
	// Use path.Dir() which always uses forward slashes for Unix-style paths
	parent := path.Dir(remotePath)
	// Ensure we don't go above root
	if parent == "." || parent == "" {
		return "/"
	}
	return parent
}

// remotePathJoin joins remote path components using Unix-style separators
func remotePathJoin(elem ...string) string {
	// Use path.Join() which always uses forward slashes for Unix-style paths
	return path.Join(elem...)
}

// isLocalRoot checks if a local path is at the filesystem root
func isLocalRoot(localPath string) bool {
	parent := filepath.Dir(localPath)
	// If parent is the same as the path, we're at root
	// This works for both Unix ("/") and Windows ("C:\", "D:\", etc.)
	return parent == localPath
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
		// Update content to reflect focus change
		m.updateViewportContent()

	case "up", "k":
		if m.focusedPanel == LeftPanel {
			if m.localCursor > 0 {
				m.localCursor--
				m.updateViewportContent()
				m.ensureCursorVisible(LeftPanel)
			}
		} else {
			if m.remoteCursor > 0 {
				m.remoteCursor--
				m.updateViewportContent()
				m.ensureCursorVisible(RightPanel)
			}
		}

	case "down", "j":
		if m.focusedPanel == LeftPanel {
			maxCursor := len(m.localFiles) - 1
			if !isLocalRoot(m.localPath) {
				maxCursor = len(m.localFiles) // Account for ".." entry
			}
			if m.localCursor < maxCursor {
				m.localCursor++
				m.updateViewportContent()
				m.ensureCursorVisible(LeftPanel)
			}
		} else {
			maxCursor := len(m.remoteFiles) - 1
			if m.remotePath != "/" {
				maxCursor = len(m.remoteFiles) // Account for ".." entry
			}
			if m.remoteCursor < maxCursor {
				m.remoteCursor++
				m.updateViewportContent()
				m.ensureCursorVisible(RightPanel)
			}
		}

	case "enter":
		return m.handleEnterDirectory()

	case "left", "h":
		return m.handleGoUpDirectory()

	case "right", "l":
		return m.handleEnterDirectoryOnly()

	case " ":
		// Toggle selection
		if m.focusedPanel == LeftPanel {
			// Calculate the actual file index, accounting for ".." entry
			fileIndex := m.localCursor
			if !isLocalRoot(m.localPath) {
				fileIndex = m.localCursor - 1 // Subtract 1 to account for ".." entry
			}
			// DEBUG: Print debug info
			fmt.Printf("DEBUG: cursor=%d, isRoot=%v, fileIndex=%d, numFiles=%d\n",
				m.localCursor, isLocalRoot(m.localPath), fileIndex, len(m.localFiles))
			if fileIndex >= 0 && fileIndex < len(m.localFiles) {
				fmt.Printf("DEBUG: Toggling selection for file[%d]: %s\n", fileIndex, m.localFiles[fileIndex].Name)
			}
			// Only toggle selection for actual files, not ".." entry
			if fileIndex >= 0 && fileIndex < len(m.localFiles) {
				m.localSelected[fileIndex] = !m.localSelected[fileIndex]
			}
		} else {
			// Calculate the actual file index, accounting for ".." entry
			fileIndex := m.remoteCursor
			if m.remotePath != "/" {
				fileIndex = m.remoteCursor - 1 // Subtract 1 to account for ".." entry
			}
			// Only toggle selection for actual files, not ".." entry
			if fileIndex >= 0 && fileIndex < len(m.remoteFiles) {
				m.remoteSelected[fileIndex] = !m.remoteSelected[fileIndex]
			}
		}
		// Update content to reflect selection change
		m.updateViewportContent()

	case "c":
		// Copy selected files
		return m.handleCopy()
	}

	return m, nil
}

// handleEnterDirectory handles entering a directory
func (m *fileBrowserModel) handleEnterDirectory() (tea.Model, tea.Cmd) {
	if m.focusedPanel == LeftPanel {
		// Check if we're selecting ".." (go up directory)
		if m.localCursor == 0 && !isLocalRoot(m.localPath) {
			m.localPath = filepath.Dir(m.localPath)
			m.localCursor = 0
			return m, loadFilesCmd(m)
		}

		// Calculate the actual file index, accounting for ".." entry
		fileIndex := m.localCursor
		if !isLocalRoot(m.localPath) {
			fileIndex = m.localCursor - 1 // Subtract 1 to account for ".." entry
		}

		if fileIndex >= 0 && fileIndex < len(m.localFiles) && m.localFiles[fileIndex].IsDir {
			m.localPath = filepath.Join(m.localPath, m.localFiles[fileIndex].Name)
			m.localCursor = 0
			return m, loadFilesCmd(m)
		}
	} else {
		// Check if we're selecting ".." (go up directory)
		if m.remoteCursor == 0 && m.remotePath != "/" {
			m.remotePath = remotePathDir(m.remotePath)
			m.remoteCursor = 0
			return m, loadFilesCmd(m)
		}

		// Calculate the actual file index, accounting for ".." entry
		fileIndex := m.remoteCursor
		if m.remotePath != "/" {
			fileIndex = m.remoteCursor - 1 // Subtract 1 to account for ".." entry
		}

		if fileIndex >= 0 && fileIndex < len(m.remoteFiles) && m.remoteFiles[fileIndex].IsDir {
			m.remotePath = remotePathJoin(m.remotePath, m.remoteFiles[fileIndex].Name)
			m.remoteCursor = 0
			return m, loadFilesCmd(m)
		}
	}
	return m, nil
}

// handleEnterDirectoryOnly handles entering a directory with right arrow (only works on directories)
func (m *fileBrowserModel) handleEnterDirectoryOnly() (tea.Model, tea.Cmd) {
	if m.focusedPanel == LeftPanel {
		// Check if we're selecting ".." (go up directory)
		if m.localCursor == 0 && !isLocalRoot(m.localPath) {
			m.localPath = filepath.Dir(m.localPath)
			m.localCursor = 0
			return m, loadFilesCmd(m)
		}

		// Calculate the actual file index, accounting for ".." entry
		fileIndex := m.localCursor
		if !isLocalRoot(m.localPath) {
			fileIndex = m.localCursor - 1 // Subtract 1 to account for ".." entry
		}

		// Only enter if it's a directory (not a regular file)
		if fileIndex >= 0 && fileIndex < len(m.localFiles) && m.localFiles[fileIndex].IsDir {
			m.localPath = filepath.Join(m.localPath, m.localFiles[fileIndex].Name)
			m.localCursor = 0
			return m, loadFilesCmd(m)
		}
	} else {
		// Check if we're selecting ".." (go up directory)
		if m.remoteCursor == 0 && m.remotePath != "/" {
			m.remotePath = remotePathDir(m.remotePath)
			m.remoteCursor = 0
			return m, loadFilesCmd(m)
		}

		// Calculate the actual file index, accounting for ".." entry
		fileIndex := m.remoteCursor
		if m.remotePath != "/" {
			fileIndex = m.remoteCursor - 1 // Subtract 1 to account for ".." entry
		}

		// Only enter if it's a directory (not a regular file)
		if fileIndex >= 0 && fileIndex < len(m.remoteFiles) && m.remoteFiles[fileIndex].IsDir {
			m.remotePath = remotePathJoin(m.remotePath, m.remoteFiles[fileIndex].Name)
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
		m.remotePath = remotePathDir(m.remotePath)
		m.remoteCursor = 0
	}
	return m, loadFilesCmd(m)
}

// View renders the file browser
func (m *fileBrowserModel) View() string {
	if m.err != nil {
		return ui.ErrorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error()))
	}

	if !m.ready {
		return "\n  Initializing file browser..."
	}

	panelWidth := (m.width - 4) / 2 // Account for borders and spacing

	leftPanel := m.renderViewportPanel(LeftPanel, panelWidth)
	rightPanel := m.renderViewportPanel(RightPanel, panelWidth)

	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		rightPanel,
	)

	help := ui.HelpStyle.Render("tab: switch panel • ↑/↓/PgUp/PgDn: navigate • ←/→: go up/into dir • space: select • c: copy • q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, panels, help)
}

// renderViewportPanel renders a single panel using viewport with header and footer
func (m *fileBrowserModel) renderViewportPanel(side PanelSide, width int) string {
	var viewport viewport.Model

	if side == LeftPanel {
		viewport = m.leftViewport
	} else {
		viewport = m.rightViewport
	}

	header := m.headerView(side)
	footer := m.footerView(side)

	content := fmt.Sprintf("%s\n%s\n%s", header, viewport.View(), footer)

	// Apply panel style based on focus
	panelStyle := ui.UnfocusedPanelStyle
	if m.focusedPanel == side {
		panelStyle = ui.FocusedPanelStyle
	}

	return panelStyle.Width(width).Height(m.height - 4).Render(content)
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
					remotePathJoin(destPath, file),
				)
			} else {
				err = client.CopyFileToLocal(
					remotePathJoin(sourcePath, file),
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
