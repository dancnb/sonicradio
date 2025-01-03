package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/config"
)

const configView = "config tab"

type settingsTab struct {
	cfg *config.Value
}

func newSettingsTab(ctx context.Context, cfg *config.Value) *settingsTab {
	return &settingsTab{
		cfg: cfg,
	}
}

func (t *settingsTab) Init(m *Model) tea.Cmd { return nil }

func (t *settingsTab) Update(m *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (t *settingsTab) View() string {
	return configView
}
