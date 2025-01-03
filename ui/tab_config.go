package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/config"
)

const configView = "config tab"

type configTab struct {
	cfg *config.Value
}

func newConfigTab(ctx context.Context, cfg *config.Value) *configTab {
	return &configTab{
		cfg: cfg,
	}
}

func (t *configTab) Init(m *Model) tea.Cmd { return nil }

func (t *configTab) Update(m *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (t *configTab) View() string {
	return configView
}
