package model

import (
	"fmt"

	"sshlepp/internal/ssh"
	"sshlepp/internal/ui"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type passwordInputModel struct {
	textInput textinput.Model
	err       error
	host      *ssh.SSHHost
}

type PasswordEnteredMsg struct {
	Password string
	Host     *ssh.SSHHost
}

type PasswordCancelledMsg struct{}

func newPasswordInputModel(host *ssh.SSHHost) *passwordInputModel {
	ti := textinput.New()
	ti.Placeholder = "Enter SSH key passphrase..."
	ti.Focus()
	ti.EchoMode = textinput.EchoPassword
	ti.Width = 50

	return &passwordInputModel{
		textInput: ti,
		host:      host,
	}
}

func (m *passwordInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *passwordInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			password := m.textInput.Value()
			return m, func() tea.Msg {
				return PasswordEnteredMsg{
					Password: password,
					Host:     m.host,
				}
			}
		case tea.KeyEscape:
			return m, func() tea.Msg {
				return PasswordCancelledMsg{}
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *passwordInputModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		ui.HeaderStyle.Render("SSH Key Passphrase Required"),
		"",
		fmt.Sprintf("Host: %s@%s", m.host.User, m.host.Hostname),
		"",
		m.textInput.View(),
		"",
		ui.HelpStyle.Render("Enter: confirm • Esc: cancel"),
	)
}
