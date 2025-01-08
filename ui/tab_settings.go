package ui

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
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

	style  *style
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

func newSettingsTab(ctx context.Context, cfg *config.Value, s *style) *settingsTab {
	h := help.New()
	h.ShowAll = false
	h.ShortSeparator = "   "
	h.Styles = s.helpStyles()

	inputs := []textinput.Model{
		s.newInputModel("History max entries", "", nil, nil, nil, nrInputValidator),
	}

	return &settingsTab{
		cfg:    cfg,
		style:  s,
		inputs: inputs,
		keymap: newSettingsKeymap(),
		help:   h,
	}
}

func (s *settingsTab) Init(m *Model) tea.Cmd {
	s.setSize(m.width, m.totHeight-m.headerHeight)

	s.help.ShowAll = false
	s.keymap.showFullHelp.SetEnabled(true)
	s.keymap.closeFullHelp.SetEnabled(false)

	return nil
}

// onEnter: reads values from config file on tab enter
func (s *settingsTab) onEnter() tea.Cmd {
	s.inputsFocused = true
	s.idx = 0
	s.inputs[historySaveMaxIdx].SetValue(fmt.Sprintf("%d", s.cfg.HistorySaveMax))
	return s.inputs[0].Focus()
}

// onExit: writes values to config file on tab exit
func (s *settingsTab) onExit() {
	log := slog.With("method", "settingsTab.onExit")
	historySaveMaxval := s.inputs[historySaveMaxIdx].Value()
	intVal, err := strconv.Atoi(historySaveMaxval)
	if err != nil {
		log.Debug(fmt.Sprintf("invalid HistorySaveMax input value: %v", err))
	} else {
		s.cfg.HistorySaveMax = intVal
	}
	// .... other fields
	err = s.cfg.Save()
	if err != nil {
		log.Debug(fmt.Sprintf("config save err: %v", err))
	}
}

func (s *settingsTab) setSize(width, height int) {
	h, v := s.style.docStyle.GetFrameSize()
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
			go s.onExit()
			m.toFavoritesTab()
		case key.Matches(msg, s.keymap.browseTab):
			go s.onExit()
			m.toBrowseTab()
		case key.Matches(msg, s.keymap.prevTab, s.keymap.historyTab):
			go s.onExit()
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
	help := s.style.helpStyle.Render(s.help.View(&s.keymap))
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

func newSettingsKeymap() settingsKeymap {
	return settingsKeymap{
		nextInput: key.NewBinding(
			key.WithKeys("down", "ctrl+j"),
			key.WithHelp("↓/ctrl+j", "next input"),
		),
		prevInput: key.NewBinding(
			key.WithKeys("up", "ctrl+k"),
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
	}
}

func (k *settingsKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.showFullHelp}
}

func (k *settingsKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.closeFullHelp},
	}
}
