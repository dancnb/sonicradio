package ui

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/config"
)

type settingsTab struct {
	cfg           *config.Value
	changeThemeFn func(int)

	style  *Style
	keymap settingsKeymap
	help   help.Model
	width  int
	height int

	idx    settingsInputIdx
	inputs []*FormElement
}

type settingsInputIdx byte

const (
	historySaveMaxIdx settingsInputIdx = iota
	themesIdx
	playerTypeIdx
	internalBufferSecIdx
	mpdHostIdx
	mpdPortIdx
	mpdPassIdx
)

var (
	descriptions = []string{
		`Maximum number of entries displayed in "History" tab.`,
		`Preview and select a theme.`,
		"Select a backend player (only those in PATH are shown: Mpv, FFplay, VLC, MPlayer, MPD), or use the experimental Internal player.\nChanges take effect after restart.\n",
		"Duration in seconds of the internal player's buffered samples (up to 5 minutes, but will increase memory usage). Set to 0 to disable buffering and seeking.\nTakes effect for the next played station.",
	}
	ffplayDesc  = "\nFFplay does not allow changing the volume during playback or seeking backward/forward."
	vlcDesc     = "\nFor VLC, pausing or seeking backward/forward may result in an invalid song title being displayed."
	mplayerDesc = "\nFor MPlayer, seeking backward/forward is not available."
	mpdDesc     = "\nFor MPD, a sound must be playing for the volume to be adjusted."

	mpdSettingsDesc = "The change will take effect after a restart."
)

func newSettingsTab(
	ctx context.Context,
	cfg *config.Value,
	s *Style,
	availablePlayerTypes []config.PlayerType,
	changeThemeFn func(int),
) *settingsTab {
	h := help.New()
	h.ShowAll = false
	h.ShortSeparator = "   "
	h.Styles = s.HelpStyles()

	// history max entries
	historySaveMax := s.NewInputModel("History max entries", "---", nil, nil, nil, NrInputValidator)

	// themes
	themeOpts := make([]OptionValue, len(Themes))
	for i := range Themes {
		themeOpts[i] = OptionValue{IdxView: i + 1, NameView: Themes[i].Name}
	}
	themeList := NewOptionList("Theme", themeOpts, cfg.Theme, s)
	// TODO: false if more than 10 themes
	themeList.SetQuick(true)
	themeList.PartialCallbackFn = changeThemeFn
	themeList.DoneCallbackFn = changeThemeFn

	// player
	playerOpts := make([]OptionValue, len(availablePlayerTypes))
	var startIdx int
	for i := range availablePlayerTypes {
		playerOpts[i] = OptionValue{IdxView: i + 1, NameView: availablePlayerTypes[i].String()}
		if availablePlayerTypes[i] == cfg.Player {
			startIdx = i
		}
	}
	playerList := NewOptionList("Player (requires restart)", playerOpts, startIdx, s)
	playerList.SetQuick(true)
	playerList.DoneCallbackFn = func(i int) {
		cfg.Player = availablePlayerTypes[i]
		slog.Info("change player type", "i", i, "new type", cfg.Player.String())
	}
	playerDesc := getPlayerDescription(availablePlayerTypes)

	//internal player settings
	internalBufferSec := s.NewInputModel("Internal buffer (seconds)", "0", nil, nil, nil, bufferDurationValidator)

	inputs := []*FormElement{
		NewFormElement(
			WithTextInput(&historySaveMax),
			WithDescription(descriptions[0])),
		NewFormElement(
			WithOptionList(&themeList),
			WithDescription(descriptions[1])),
		NewFormElement(
			WithOptionList(&playerList),
			WithDescription(playerDesc)),
		NewFormElement(
			WithTextInput(&internalBufferSec),
			WithDescription(descriptions[3])),
	}
	if slices.Contains(availablePlayerTypes, config.MPD) {
		mpdHost := s.NewInputModel("MPD hostname", "127.0.0.1", nil, nil, nil, nil)
		inputs = append(inputs, NewFormElement(
			WithTextInput(&mpdHost),
			WithDescription(mpdSettingsDesc)),
		)
		mpdPort := s.NewInputModel("MPD port", "6600", nil, nil, nil, portValidator)
		inputs = append(inputs, NewFormElement(
			WithTextInput(&mpdPort),
			WithDescription(mpdSettingsDesc)),
		)
		mpdPass := s.NewInputModel("MPD password", "---", nil, nil, nil, nil)
		mpdPass.EchoMode = textinput.EchoPassword
		inputs = append(inputs, NewFormElement(
			WithTextInput(&mpdPass),
			WithDescription(mpdSettingsDesc)),
		)
	}

	st := &settingsTab{
		cfg:           cfg,
		changeThemeFn: changeThemeFn,
		style:         s,
		inputs:        inputs,
		keymap:        newSettingsKeymap(),
		help:          h,
	}

	st.loadConfig()
	return st
}

func portValidator(v string) error {
	port, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("invalid port: %v", err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port out of range: %d", port)
	}
	return nil
}

func getPlayerDescription(playerTypes []config.PlayerType) string {
	playerDesc := descriptions[2]
	if slices.Contains(playerTypes, config.FFPlay) {
		playerDesc += ffplayDesc
	}
	if slices.Contains(playerTypes, config.Vlc) {
		playerDesc += vlcDesc
	}
	if slices.Contains(playerTypes, config.MPlayer) {
		playerDesc += mplayerDesc
	}
	if slices.Contains(playerTypes, config.MPD) {
		playerDesc += mpdDesc
	}
	return playerDesc
}

func (s *settingsTab) loadConfig() {
	s.inputs[historySaveMaxIdx].SetValue(fmt.Sprintf("%d", *s.cfg.HistorySaveMax))

	s.inputs[internalBufferSecIdx].SetValue(fmt.Sprintf("%d", s.cfg.Internal.BufferSeconds))

	if len(s.inputs) > int(internalBufferSecIdx)+1 {
		s.inputs[mpdHostIdx].SetValue(s.cfg.MpdHost)
		s.inputs[mpdPortIdx].SetValue(fmt.Sprintf("%d", s.cfg.MpdPort))
		if s.cfg.MpdPassword != nil {
			s.inputs[mpdPassIdx].SetValue(*s.cfg.MpdPassword)
		}
	}
}

func (s *settingsTab) Init(m *Model) tea.Cmd {
	s.setSize(m.width, m.totHeight-m.headerHeight)

	showAll := false
	s.help.ShowAll = showAll

	return nil
}

// onEnter: reads values from config file on tab enter
func (s *settingsTab) onEnter() tea.Cmd {
	slog.Info("settingsTab.onEnter")
	s.idx = 0
	s.keymap.setEnable(true, s.help.ShowAll)

	s.loadConfig()

	return s.inputs[historySaveMaxIdx].Focus()
}

func (s *settingsTab) onExit() {
	s.inputs[themesIdx].Blur()
	s.keymap.setEnable(false, false)

	s.updateConfig()
}

func (s *settingsTab) updateConfig() {
	log := slog.With("method", "settingsTab.saveConfig")

	historySaveMaxval := s.inputs[historySaveMaxIdx].Value()
	intVal, err := strconv.Atoi(historySaveMaxval)
	if err != nil {
		log.Info(fmt.Sprintf("invalid HistorySaveMax input value: %v", err))
	} else {
		s.cfg.HistorySaveMax = &intVal
	}

	internalBufferSecVal := s.inputs[internalBufferSecIdx].Value()
	bIntVal, err := strconv.Atoi(internalBufferSecVal)
	if err != nil {
		log.Info(fmt.Sprintf("invalid internalBufferSec input value: %v", err))
	} else {
		s.cfg.Internal.BufferSeconds = bIntVal
	}

	if len(s.inputs) > int(internalBufferSecIdx)+1 {
		mpdHost := strings.TrimSpace(s.inputs[mpdHostIdx].Value())
		s.cfg.MpdHost = mpdHost

		portV := strings.TrimSpace(s.inputs[mpdPortIdx].Value())
		mpdPort, err := strconv.Atoi(portV)
		if err != nil {
			log.Info(fmt.Sprintf("mpd port(%s) parse error: %v", portV, err))
		} else {
			s.cfg.MpdPort = mpdPort
		}

		passV := strings.TrimSpace(s.inputs[mpdPassIdx].Value())
		s.cfg.MpdPassword = &passV
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

	case OptionMsg:
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
		case key.Matches(msg, s.keymap.reset):
			s.resetSettings()
			return m, tea.Batch(cmds...)
		}
	}

	var cmd tea.Cmd
	s.inputs[s.idx], cmd = s.inputs[s.idx].Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (s *settingsTab) resetSettings() {
	defHistorySaveMax := config.DefHistorySaveMax
	s.cfg.HistorySaveMax = &defHistorySaveMax
	val := strconv.Itoa(defHistorySaveMax)
	s.inputs[historySaveMaxIdx].SetValue(val)

	s.cfg.Internal.BufferSeconds = config.DefInternalBufferSeconds
	bVal := strconv.Itoa(config.DefInternalBufferSeconds)
	s.inputs[internalBufferSecIdx].SetValue(bVal)

	if len(s.inputs) > int(internalBufferSecIdx)+1 {
		s.cfg.MpdHost = config.DefMpdHost
		s.inputs[mpdHostIdx].SetValue(config.DefMpdHost)

		s.cfg.MpdPort = config.DefMpdPort
		val := strconv.Itoa(config.DefMpdPort)
		s.inputs[mpdPortIdx].SetValue(val)

		s.inputs[mpdPassIdx].SetValue("")
	}

	s.changeThemeFn(0)
	s.inputs[themesIdx].SetValue(0)
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
		if i < int(mpdHostIdx) {
			b.WriteRune('\n')
		}
	}

	currInput := s.inputs[s.idx]
	availHeight := s.height

	// description
	desc := s.style.SettingDescription.Width(s.width).Render(currInput.Description()) + "\n"
	availHeight -= lipgloss.Height(desc) - 2

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
	reset         key.Binding
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
			key.WithKeys("down", "ctrl+j"),
			key.WithHelp("↓/ctrl+j", "next setting"),
		),
		prevInput: key.NewBinding(
			key.WithKeys("up", "ctrl+k"),
			key.WithHelp("↑/ctrl+k", "prev setting"),
		),
		enterInput: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("space/enter", "change setting"),
		),
		reset: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reset settings"),
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
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (k *settingsKeymap) setEnable(v bool, showAll bool) {
	k.nextInput.SetEnabled(v)
	k.prevInput.SetEnabled(v)
	k.enterInput.SetEnabled(v)
	k.reset.SetEnabled(v)
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
		{k.prevInput, k.nextInput, k.enterInput, k.reset},
		{k.prevTab, k.nextTab, k.favoritesTab, k.browseTab, k.historyTab},
		{k.quit, k.closeFullHelp},
	}
}

func bufferDurationValidator(s string) error {
	val, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if val < 0 || val > 300 {
		return fmt.Errorf("buffer duration out of bonds: %d", val)
	}
	return nil
}
