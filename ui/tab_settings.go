package ui

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/dancnb/sonicradio/ui/components"
	"github.com/dancnb/sonicradio/ui/styles"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/config"
)

type settingsTab struct {
	cfg *config.Value

	style  *styles.Style
	keymap settingsKeymap
	help   help.Model
	width  int
	height int

	idx    settingsInputIdx
	inputs []*components.FormElement
}

type settingsInputIdx byte

const (
	historySaveMaxIdx settingsInputIdx = iota
	themesIdx
)

var (
	descriptions = []string{
		`Maximum number of entries displayed in "History" tab.`,
		`Preview and select a theme.`,
		`Choose one of the available backend players (only those found in PATH are displayed): Mpv, FFplay. The choice will take effect after a restart.`,
	}
	ffplayDesc = "\nFFplay does not allow changing the volume during playback or seeking backward/forward."
)

func newSettingsTab(
	ctx context.Context,
	cfg *config.Value,
	s *styles.Style,
	playerTypes []config.PlayerType,
	changeThemeFn func(int),
) *settingsTab {
	h := help.New()
	h.ShowAll = false
	h.ShortSeparator = "   "
	h.Styles = s.HelpStyles()

	// history max entries
	historySaveMax := s.NewInputModel("History max entries", "---", nil, nil, nil, styles.NrInputValidator)

	// themes
	themeOpts := make([]components.OptionValue, len(styles.Themes))
	for i := range styles.Themes {
		themeOpts[i] = components.OptionValue{IdxView: i + 1, NameView: styles.Themes[i].Name}
	}
	themeList := components.NewOptionList("Theme", themeOpts, cfg.Theme, s)
	// TODO: false if more than 10 themes
	themeList.SetQuick(true)
	themeList.PartialCallbackFn = changeThemeFn
	themeList.DoneCallbackFn = changeThemeFn

	// player
	playerOpts := make([]components.OptionValue, len(playerTypes))
	var startIdx int
	for i := range playerTypes {
		playerOpts[i] = components.OptionValue{IdxView: i + 1, NameView: playerTypes[i].String()}
		if playerTypes[i] == cfg.Player {
			startIdx = i
		}
	}
	playerList := components.NewOptionList("Player (requires restart)", playerOpts, startIdx, s)
	playerList.SetQuick(true)
	playerList.DoneCallbackFn = func(i int) {
		cfg.Player = playerTypes[i]
		slog.Debug("change player type", "i", i, "new type", cfg.Player.String())
	}

	playerDesc := descriptions[2]
	if slices.Contains(playerTypes, config.FFPlay) {
		playerDesc += ffplayDesc
	}
	st := &settingsTab{
		cfg:   cfg,
		style: s,
		inputs: []*components.FormElement{
			components.NewFormElement(
				components.WithTextInput(&historySaveMax),
				components.WithDescription(descriptions[0])),
			components.NewFormElement(
				components.WithOptionList(&themeList),
				components.WithDescription(descriptions[1])),
			components.NewFormElement(
				components.WithOptionList(&playerList),
				components.WithDescription(playerDesc)),
		},
		keymap: newSettingsKeymap(),
		help:   h,
	}

	st.loadConfig()
	return st
}

func (s *settingsTab) loadConfig() {
	s.inputs[historySaveMaxIdx].SetValue(fmt.Sprintf("%d", *s.cfg.HistorySaveMax))
}

func (s *settingsTab) Init(m *Model) tea.Cmd {
	s.setSize(m.width, m.totHeight-m.headerHeight)

	showAll := false
	s.help.ShowAll = showAll

	return nil
}

// onEnter: reads values from config file on tab enter
func (s *settingsTab) onEnter() tea.Cmd {
	slog.Debug("settingsTab.onEnter")
	s.idx = 0
	s.keymap.setEnable(true, s.help.ShowAll)

	s.loadConfig()

	return s.inputs[historySaveMaxIdx].Focus()
}

func (s *settingsTab) onExit() {
	slog.Debug("settingsTab.onExit")
	s.inputs[themesIdx].Blur()
	s.keymap.setEnable(false, false)
	go s.saveConfig()
}

// saveConfig: writes values to config file on tab exit
func (s *settingsTab) saveConfig() {
	log := slog.With("method", "settingsTab.onExit")
	historySaveMaxval := s.inputs[historySaveMaxIdx].Value()
	intVal, err := strconv.Atoi(historySaveMaxval)
	if err != nil {
		log.Debug(fmt.Sprintf("invalid HistorySaveMax input value: %v", err))
	} else {
		s.cfg.HistorySaveMax = &intVal
	}

	err = s.cfg.Save()
	if err != nil {
		log.Debug(fmt.Sprintf("config save err: %v", err))
	}
}

func (s *settingsTab) setSize(width, height int) {
	h, v := s.style.DocStyle.GetFrameSize()
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

	case components.OptionMsg:
		var idx int
		if msg.Done {
			idx = msg.SelIdx
			currInput := s.inputs[s.idx]
			if !currInput.IsActive() {
				s.keymap.setEnable(true, s.help.ShowAll)
			}
		} else {
			idx = msg.PreviewIdx
		}
		if msg.CallbackFn != nil {
			msg.CallbackFn(idx)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.quit):
			return m, tea.Quit

		case key.Matches(msg, s.keymap.showFullHelp):
			fallthrough
		case key.Matches(msg, s.keymap.closeFullHelp):
			s.help.ShowAll = !s.help.ShowAll
			s.keymap.showFullHelp.SetEnabled(!s.help.ShowAll)
			s.keymap.closeFullHelp.SetEnabled(s.help.ShowAll)
			return m, tea.Batch(cmds...)

		// case key.Matches(msg, s.keymap.search):
		// 	s.onExit()
		// 	m.toBrowseTab()
		// 	return m.tabs[browseTabIx].Update(m, msg)
		case key.Matches(msg, s.keymap.nextTab, s.keymap.favoritesTab):
			s.onExit()
			m.toFavoritesTab()
		case key.Matches(msg, s.keymap.browseTab):
			s.onExit()
			m.toBrowseTab()
		case key.Matches(msg, s.keymap.prevTab, s.keymap.historyTab):
			s.onExit()
			m.toHistoryTab()

		case key.Matches(msg, s.keymap.nextInput):
			s.idx++
			s.idx = s.idx % settingsInputIdx(len(s.inputs))
			cmds = s.changeInput(cmds)
			return m, tea.Batch(cmds...)
		case key.Matches(msg, s.keymap.prevInput):
			if s.idx == 0 {
				s.idx = settingsInputIdx(len(s.inputs))
			}
			s.idx--
			cmds = s.changeInput(cmds)
			return m, tea.Batch(cmds...)
		case key.Matches(msg, s.keymap.enterInput):
			s.keymap.setEnable(s.inputs[s.idx].Keymap() == nil, s.help.ShowAll)
			s.inputs[s.idx].SetActive()
			return m, tea.Batch(cmds...)
		}
	}

	var cmd tea.Cmd
	s.inputs[s.idx], cmd = s.inputs[s.idx].Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (s *settingsTab) changeInput(cmds []tea.Cmd) []tea.Cmd {
	for i := range s.inputs {
		if i == int(s.idx) {
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
		b.WriteRune('\n')
	}

	currInput := s.inputs[s.idx]
	availHeight := s.height

	// description
	desc := s.style.SettingDescription.Width(s.width).Render(currInput.Description()) + "\n"
	availHeight -= lipgloss.Height(desc)

	// help
	var elemKeymap help.KeyMap
	var help string
	if currInput.Keymap() != nil && currInput.IsActive() {
		elemKeymap = currInput.Keymap()
	} else {
		elemKeymap = &s.keymap
	}
	help = s.style.HelpStyle.Render(s.help.View(elemKeymap))
	availHeight -= lipgloss.Height(help)

	inputs := b.String()
	inputsHeight := lipgloss.Height(inputs)
	for i := 0; i < availHeight-inputsHeight; i++ {
		b.WriteString("\n")
	}
	return b.String() + desc + help
}

type settingsKeymap struct {
	nextInput     key.Binding
	prevInput     key.Binding
	enterInput    key.Binding
	nextTab       key.Binding
	prevTab       key.Binding
	favoritesTab  key.Binding
	browseTab     key.Binding
	historyTab    key.Binding
	showFullHelp  key.Binding
	closeFullHelp key.Binding
	quit          key.Binding
}

func newSettingsKeymap() settingsKeymap {
	return settingsKeymap{
		nextInput: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "next setting"),
		),
		prevInput: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "prev setting"),
		),
		enterInput: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("space/enter", "change setting"),
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
		quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k *settingsKeymap) setEnable(v bool, showAll bool) {
	k.nextInput.SetEnabled(v)
	k.prevInput.SetEnabled(v)
	k.enterInput.SetEnabled(v)
	k.nextTab.SetEnabled(v)
	k.prevTab.SetEnabled(v)
	k.favoritesTab.SetEnabled(v)
	k.browseTab.SetEnabled(v)
	k.historyTab.SetEnabled(v)
	if v {
		k.showFullHelp.SetEnabled(!showAll)
		k.closeFullHelp.SetEnabled(showAll)
	} else {
		k.showFullHelp.SetEnabled(false)
		k.closeFullHelp.SetEnabled(false)
	}
	k.quit.SetEnabled(v)
}

func (k *settingsKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.prevInput, k.nextInput, k.enterInput, k.quit, k.showFullHelp}
}

func (k *settingsKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.prevInput, k.nextInput, k.enterInput},
		{k.prevTab, k.nextTab, k.favoritesTab, k.browseTab, k.historyTab},
		{k.quit, k.closeFullHelp},
	}
}
