package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"sshlepp/internal/ssh"
	"sshlepp/internal/ui"
)

// serverSelectModel handles server selection
type serverSelectModel struct {
	hosts        []ssh.SSHHost
	cursor       int
	selectedHost *ssh.SSHHost
}

// newServerSelectModel creates a new server selection model
func newServerSelectModel(hosts []ssh.SSHHost) *serverSelectModel {
	return &serverSelectModel{
		hosts: hosts,
	}
}

// Init initializes the server selection model
func (m *serverSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for server selection
func (m *serverSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.hosts)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.hosts) {
				m.selectedHost = &m.hosts[m.cursor]
				return m, nil
			}
		}
	}
	return m, nil
}

// View renders the server selection
func (m *serverSelectModel) View() string {
	if len(m.hosts) == 0 {
		return ui.ErrorStyle.Render("No SSH hosts found in ~/.ssh/config")
	}

	var s strings.Builder
	s.WriteString("Select an SSH server:\n\n")

	for i, host := range m.hosts {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		style := ui.RegularRowStyle
		if m.cursor == i {
			style = ui.SelectedRowStyle
		}

		line := fmt.Sprintf("%s %s (%s)", cursor, host.Name, host.String())
		s.WriteString(style.Render(line) + "\n")
	}

	s.WriteString("\n")
	s.WriteString(ui.HelpStyle.Render("↑/↓: navigate • enter: select • q: quit"))

	return s.String()
}
