package ui

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type searchModel struct {
	enabled bool
	content string
	keymap  searchKeymap
}

func (s *searchModel) toggle() {
	s.enabled = !s.enabled
}

func (s *searchModel) update(msg tea.Msg) tea.Cmd {
	slog.Debug("searchModel", "type", fmt.Sprintf("%T", msg), "value", msg, "#", fmt.Sprintf("%#v", msg))
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, s.keymap.cancelSearch) {
			s.enabled = false
		}
	}

	return tea.Batch(cmds...)
}

func (s *searchModel) view() string {
	return s.content
}

type searchKeymap struct {
	cancelSearch key.Binding
}

func newSearchKeymap() searchKeymap {
	return searchKeymap{
		cancelSearch: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}
