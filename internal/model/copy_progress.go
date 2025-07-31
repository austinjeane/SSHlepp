package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"sshlepp/internal/ui"
)

// copyProgressModel handles copy progress display
type copyProgressModel struct {
	progress     progress.Model
	currentFile  string
	totalFiles   int
	currentIndex int
	percent      float64
}

// newCopyProgressModel creates a new copy progress model
func newCopyProgressModel() *copyProgressModel {
	return &copyProgressModel{
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

// Init initializes the copy progress model
func (m *copyProgressModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for copy progress
func (m *copyProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	updatedProgress, cmd := m.progress.Update(msg)
	m.progress = updatedProgress.(progress.Model)
	return m, cmd
}

// View renders the copy progress
func (m *copyProgressModel) View() string {
	if m.totalFiles == 0 {
		return ""
	}

	status := fmt.Sprintf("Copying %s (%d/%d)", m.currentFile, m.currentIndex+1, m.totalFiles)
	progressBar := m.progress.ViewAs(m.percent)
	
	return ui.ProgressStyle.Render(status + "\n" + progressBar)
}

// Copy progress message types
type copyStartMsg struct {
	files []string
}

type copyProgressMsg struct {
	fileIndex int
	fileName  string
	percent   float64
}

type copyCompleteMsg struct{}
