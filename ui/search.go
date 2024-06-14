package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type searchModel struct {
	enabled bool
	content string
	keymap  searchKeymap
}

func (s *searchModel) setEnabled(v bool) {
	s.enabled = v
	s.keymap.setEnable(v)
}

func (s *searchModel) update(msg tea.Msg) tea.Cmd {
	logTeaMsg(msg, "update searchModel")
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, s.keymap.cancelSearch) {
			s.setEnabled(false)
			return func() tea.Msg { return searchRespMsg{cancelled: true} }
		} else if key.Matches(msg, s.keymap.submit) {
			return func() tea.Msg {
				s.setEnabled(false)
				return searchRespMsg{}
			}
		}

		// TODO imple search inputs + req

	}

	return tea.Batch(cmds...)
}

func (s *searchModel) view() string {
	return s.content
}

type searchKeymap struct {
	submit       key.Binding
	cancelSearch key.Binding
}

func newSearchKeymap() searchKeymap {
	return searchKeymap{
		submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		cancelSearch: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

func (k searchKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.submit, k.cancelSearch}
}

func (k searchKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(), //first column
		// {k.Help, k.Quit},                // second column
	}
}

func (k searchKeymap) setEnable(v bool) {
	k.submit.SetEnabled(v)
	k.cancelSearch.SetEnabled(v)
}
