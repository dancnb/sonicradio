package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/config"
)

const configView = "config tab"

type settingsTab struct {
	cfg *config.Value

	keymap settingsKeymap
	help   help.Model
	width  int
	height int

	inputsFocused bool
	idx           settingsInputIdx
	inputs        []textinput.Model
}
type settingsInputIdx byte

const (
	historySaveMaxIdx settingsInputIdx = iota
)

func newSettingsTab(ctx context.Context, cfg *config.Value) *settingsTab {
	h := help.New()
	h.ShowAll = false
	h.ShortSeparator = "   "
	h.Styles = helpStyles()

	inputs := []textinput.Model{
		newInputModel("History entries", "", nil, nil, nil, nrInputValidator),
	}

	return &settingsTab{
		cfg:    cfg,
		inputs: inputs,
		keymap: settingsKeymap{
			nextInput: key.NewBinding(
				key.WithKeys("down", "tab", "ctrl+j"),
				key.WithHelp("↓/ctrl+j", "next input"),
			),
			prevInput: key.NewBinding(
				key.WithKeys("up", "shift+tab", "ctrl+k"),
				key.WithHelp("↑/ctrl+k", "prev input"),
			),
			nextTab: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "go to next tab"),
			),
			prevTab: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "go to prev tab"),
			),
			historyTab: key.NewBinding(
				key.WithKeys("H"),
				key.WithHelp("H", "go to history tab"),
			),
			favoritesTab: key.NewBinding(
				key.WithKeys("F"),
				key.WithHelp("F", "go to favorites tab"),
			),
			browseTab: key.NewBinding(
				key.WithKeys("B"),
				key.WithHelp("B", "go to browse tab"),
			),
			showFullHelp: key.NewBinding(
				key.WithKeys("?"),
				key.WithHelp("?", "more"),
			),
			closeFullHelp: key.NewBinding(
				key.WithKeys("?"),
				key.WithHelp("?", "close help"),
			),
		},
		help: h,
	}
}

func (s *settingsTab) Init(m *Model) tea.Cmd {
	s.setSize(m.width, m.totHeight-m.headerHeight)

	s.help.ShowAll = false
	s.keymap.showFullHelp.SetEnabled(true)
	s.keymap.closeFullHelp.SetEnabled(false)
	s.inputsFocused = true

	s.idx = 0
	s.inputs[historySaveMaxIdx].SetValue(fmt.Sprintf("%d", s.cfg.HistorySaveMax))
	return nil
}

func (s *settingsTab) focus() tea.Cmd {
	return s.inputs[0].Focus()
}

func (s *settingsTab) setSize(width, height int) {
	h, v := docStyle.GetFrameSize()
	s.width = width - h
	s.height = height - v
	s.help.Width = s.width
}

func (s *settingsTab) Update(m *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "ui.settingsTab.Update")

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		availableHeight := msg.Height - m.headerHeight
		s.setSize(msg.Width, availableHeight)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.showFullHelp):
			fallthrough
		case key.Matches(msg, s.keymap.closeFullHelp):
			s.help.ShowAll = !s.help.ShowAll
			s.keymap.showFullHelp.SetEnabled(!s.help.ShowAll)
			s.keymap.closeFullHelp.SetEnabled(s.help.ShowAll)
			return m, tea.Batch(cmds...)

		case key.Matches(msg, s.keymap.nextTab, s.keymap.favoritesTab):
			m.toFavoritesTab()
		case key.Matches(msg, s.keymap.browseTab):
			m.toBrowseTab()
		case key.Matches(msg, s.keymap.prevTab, s.keymap.historyTab):
			m.toHistoryTab()

		case key.Matches(msg, s.keymap.nextInput):
			s.idx++
			s.idx = s.idx % settingsInputIdx(len(s.inputs))
			cmds = s.updateInputs(cmds)
		case key.Matches(msg, s.keymap.prevInput):
			if s.idx == 0 {
				s.idx = settingsInputIdx(len(s.inputs))
			}
			s.idx--
			cmds = s.updateInputs(cmds)

		}
	}

	// update all fields: inputs, etc
	for i := range s.inputs {
		var cmd tea.Cmd
		s.inputs[i], cmd = s.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (s *settingsTab) updateInputs(cmds []tea.Cmd) []tea.Cmd {
	for i := range s.inputs {
		if s.inputsFocused && i == int(s.idx) {
			cmds = append(cmds, s.inputs[i].Focus())
			continue
		}
		s.inputs[i].Blur()
	}
	return cmds
}

func (s *settingsTab) View() string {
	var b strings.Builder
	// content
	for i := range s.inputs {
		b.WriteString(s.inputs[i].View())
		b.WriteRune('\n')
	}
	b.WriteRune('\n')
	// help
	availHeight := s.height
	help := helpStyle.Render(s.help.View(&s.keymap))
	availHeight -= lipgloss.Height(help)
	inputs := b.String()
	inputsHeight := lipgloss.Height(inputs)
	for i := 0; i < availHeight-inputsHeight; i++ {
		b.WriteString("\n")
	}
	return b.String() + help
}

type settingsKeymap struct {
	nextInput     key.Binding
	prevInput     key.Binding
	nextTab       key.Binding
	prevTab       key.Binding
	favoritesTab  key.Binding
	browseTab     key.Binding
	historyTab    key.Binding
	showFullHelp  key.Binding
	closeFullHelp key.Binding
}

func (k *settingsKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.showFullHelp}
}

func (k *settingsKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.closeFullHelp},
	}
}
