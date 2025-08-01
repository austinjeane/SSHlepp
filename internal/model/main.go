package model

import (
	"fmt"

	"sshlepp/internal/ssh"
	"sshlepp/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AppState represents the current state of the application
type AppState int

const (
	StateServerSelect AppState = iota
	StatePasswordInput
	StateFileBrowser
	StateCopying
)

// mainModel is the main Bubble Tea model
type mainModel struct {
	state         AppState
	serverSelect  *serverSelectModel
	passwordInput *passwordInputModel
	fileBrowser   *fileBrowserModel
	copyProgress  *copyProgressModel
	width, height int
	error         error
}

// NewMainModel creates a new main model
func NewMainModel() (*mainModel, error) {
	hosts, err := ssh.ParseSSHConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH config: %w", err)
	}

	return &mainModel{
		state:        StateServerSelect,
		serverSelect: newServerSelectModel(hosts),
	}, nil
}

// Init initializes the main model
func (m *mainModel) Init() tea.Cmd {
	return m.serverSelect.Init()
}

// Update handles messages
func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Forward window size to active sub-models
		switch m.state {
		case StateFileBrowser:
			if m.fileBrowser != nil {
				newModel, newCmd := m.fileBrowser.Update(msg)
				m.fileBrowser = newModel.(*fileBrowserModel)
				cmd = newCmd
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case PasswordEnteredMsg:
		// Try to create SSH client with password
		client, err := ssh.NewClientWithPassphrase(*msg.Host, msg.Password)
		if err != nil {
			m.error = fmt.Errorf("authentication failed: %w", err)
			m.state = StateServerSelect
			return m, nil
		}

		m.state = StateFileBrowser
		m.fileBrowser, cmd = newFileBrowserModel(client, msg.Host, m.width, m.height)
		return m, cmd

	case PasswordCancelledMsg:
		m.state = StateServerSelect
		return m, nil
	}

	switch m.state {
	case StateServerSelect:
		newModel, newCmd := m.serverSelect.Update(msg)
		m.serverSelect = newModel.(*serverSelectModel)
		cmd = newCmd

		// Check if a server was selected
		if m.serverSelect.selectedHost != nil {
			// Try to connect without password first
			client, err := ssh.NewClient(*m.serverSelect.selectedHost)
			if err != nil {
				// If connection failed, might need password
				m.state = StatePasswordInput
				m.passwordInput = newPasswordInputModel(m.serverSelect.selectedHost)
				return m, m.passwordInput.Init()
			}

			m.state = StateFileBrowser
			m.fileBrowser, cmd = newFileBrowserModel(client, m.serverSelect.selectedHost, m.width, m.height)
			return m, cmd
		}

	case StatePasswordInput:
		newModel, newCmd := m.passwordInput.Update(msg)
		m.passwordInput = newModel.(*passwordInputModel)
		cmd = newCmd

	case StateFileBrowser:
		newModel, newCmd := m.fileBrowser.Update(msg)
		m.fileBrowser = newModel.(*fileBrowserModel)
		cmd = newCmd

	case StateCopying:
		// TODO: Add copy progress logic
	}

	return m, cmd
}

// View renders the main model
func (m *mainModel) View() string {
	if m.error != nil {
		return ui.ErrorStyle.Render(fmt.Sprintf("Error: %s", m.error.Error()))
	}

	switch m.state {
	case StateServerSelect:
		return m.serverSelect.View()
	case StatePasswordInput:
		return m.passwordInput.View()
	case StateFileBrowser:
		return m.fileBrowser.View()
	case StateCopying:
		return lipgloss.JoinVertical(lipgloss.Left, m.fileBrowser.View(), m.copyProgress.View())
	default:
		return ""
	}
}
