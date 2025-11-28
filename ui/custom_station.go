package ui

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/model"
	"github.com/google/uuid"
)

type customStationModel struct {
	enabled bool
	style   *Style

	browser   *browser.API
	countries []string
	languages []string

	textInputs []FormElement
	idx        customStationInputIdx

	keymap customStationKeymap
	help   help.Model
	width  int
	height int
}

type customStationInputIdx byte

const (
	customStationInputIdxName customStationInputIdx = iota
	customStationInputIdxURL
	customStationInputIdxHomepage
	customStationInputIdxTags
	customStationInputIdxCountry
	customStationInputIdxLanguage
	customStationInputIdxBitrate
)

func newCustomStationModel(b *browser.API, s *Style) *customStationModel {
	k := newCustomStationKeymap()
	inputs := []textinput.Model{
		s.NewInputModel("Name", "required", &k.prevSugg, &k.nextSugg, &k.acceptSugg, nil),
		s.NewInputModel("URL", "required", &k.prevSugg, &k.nextSugg, &k.acceptSugg, nil),
		s.NewInputModel("Homepage", "---", &k.prevSugg, &k.nextSugg, &k.acceptSugg, nil),
		s.NewInputModel("Tags", "comma separated list", &k.prevSugg, &k.nextSugg, &k.acceptSugg, nil),
		s.NewInputModel("Country Code", "---", &k.prevSugg, &k.nextSugg, &k.acceptSugg, nil),
		s.NewInputModel("Language", "---", &k.prevSugg, &k.nextSugg, &k.acceptSugg, nil),
		s.NewInputModel("Bitrate", "128", &k.prevSugg, &k.nextSugg, &k.acceptSugg, NrInputValidator),
	}
	formElems := make([]FormElement, len(inputs))
	for ii := range inputs {
		formElems[ii] = *NewFormElement(WithTextInput(&inputs[ii]))
	}
	h := help.New()
	h.ShowAll = false
	h.ShortSeparator = "   "
	h.Styles = s.HelpStyles()

	m := &customStationModel{
		style:      s,
		browser:    b,
		textInputs: formElems,
		keymap:     k,
		help:       h,
	}
	go m.getSuggestions()
	return m
}

func (s *customStationModel) getSuggestions() {
	countries, err := s.browser.GetCountries()
	if err == nil && len(countries) > 0 {
		for i := range countries {
			s.countries = append(s.countries, countries[i].ISO3166_1)
		}
		s.textInputs[customStationInputIdxCountry].TextInput().ShowSuggestions = true
		s.textInputs[customStationInputIdxCountry].TextInput().SetSuggestions(s.countries)
	}

	langs, err := s.browser.GetLanguages()
	if err == nil && len(langs) > 0 {
		for i := range langs {
			s.languages = append(s.languages, langs[i].Name)
		}
		s.textInputs[customStationInputIdxLanguage].TextInput().ShowSuggestions = true
		s.textInputs[customStationInputIdxLanguage].TextInput().SetSuggestions(s.languages)
	}
}

func (s *customStationModel) Init() tea.Cmd {
	s.setEnabled(true)
	s.keymap.prevInput.SetHelp("↑/ctrl+k", "prev input")
	s.keymap.nextInput.SetHelp("↓/ctrl+j", "next input")
	return s.textInputs[0].Focus()
}

func (s *customStationModel) setSize(width, height int) {
	h, v := s.style.DocStyle.GetFrameSize()
	s.width = width - h
	s.height = height - v
	s.help.Width = s.width
}

func (s *customStationModel) isEnabled() bool {
	return s.enabled
}

// setEnabled is called on search page enter/exit only
func (s *customStationModel) setEnabled(v bool) {
	s.enabled = v
	s.idx = customStationInputIdxName
	for i := range s.textInputs {
		s.textInputs[i].Blur()
		s.textInputs[i].TextInput().Reset()
	}
	showAll := false
	s.help.ShowAll = showAll
	s.keymap.setEnable(v, showAll)
}

func (s *customStationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "ui.customStationModel.Update")
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.setSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch {

		case key.Matches(msg, s.keymap.showFullHelp):
			fallthrough
		case key.Matches(msg, s.keymap.closeFullHelp):
			s.help.ShowAll = !s.help.ShowAll
			s.keymap.showFullHelp.SetEnabled(!s.help.ShowAll)
			s.keymap.closeFullHelp.SetEnabled(s.help.ShowAll)
			s.keymap.update(s.help.ShowAll)
			return s, tea.Batch(cmds...)

		case key.Matches(msg, s.keymap.cancel):
			return s, func() tea.Msg {
				s.setEnabled(false)
				return customStationRespMsg{cancelled: true}
			}

		case key.Matches(msg, s.keymap.submit):
			name := strings.TrimSpace(s.textInputs[customStationInputIdxName].Value())
			url := strings.TrimSpace(s.textInputs[customStationInputIdxURL].Value())
			if len(name) == 0 {
				s.idx = customStationInputIdxName
				cmds = s.updateInputs(cmds)
				return s, tea.Batch(cmds...)
			} else if len(url) == 0 {
				s.idx = customStationInputIdxURL
				cmds = s.updateInputs(cmds)
				return s, tea.Batch(cmds...)
			}

			return s, func() tea.Msg {
				defer s.setEnabled(false)

				brS := strings.TrimSpace(s.textInputs[customStationInputIdxBitrate].Value())
				var br *int64
				if brVal, err := strconv.Atoi(brS); err != nil {
					slog.Error(fmt.Sprintf("invalid bitrate value %v: %v", brS, err))
				} else {
					brInt64 := int64(brVal)
					br = &brInt64
				}
				station := &model.Station{
					Stationuuid: uuid.NewString(),
					Name:        name,
					URL:         url,
					Homepage:    strings.TrimSpace(s.textInputs[customStationInputIdxHomepage].Value()),
					Tags:        strings.TrimSpace(s.textInputs[customStationInputIdxTags].Value()),
					Countrycode: strings.TrimSpace(s.textInputs[customStationInputIdxCountry].Value()),
					Language:    strings.TrimSpace(s.textInputs[customStationInputIdxLanguage].Value()),
					IsCustom:    true,
				}
				if br != nil {
					station.Bitrate = *br
				}
				return customStationRespMsg{station: station}
			}

		case key.Matches(msg, s.keymap.nextInput):
			if msg.String() == "tab" && strings.TrimSpace(s.textInputs[s.idx].Value()) != "" && s.textInputs[s.idx].TextInput().ShowSuggestions {
				s.textInputs[s.idx].SetValue(s.textInputs[s.idx].TextInput().CurrentSuggestion())
				s.textInputs[s.idx].TextInput().CursorEnd()
			}
			s.idx++
			s.idx = s.idx % customStationInputIdx(len(s.textInputs))
			cmds = s.updateInputs(cmds)
		case key.Matches(msg, s.keymap.prevInput):
			if s.idx == 0 {
				s.idx = customStationInputIdxBitrate
			}
			s.idx--
			cmds = s.updateInputs(cmds)
		}
	}

	for i := range s.textInputs {
		var cmd tea.Cmd
		fEl, cmd := s.textInputs[i].Update(msg)
		s.textInputs[i] = *fEl
		cmds = append(cmds, cmd)
	}

	return s, tea.Batch(cmds...)
}

func (s *customStationModel) updateInputs(cmds []tea.Cmd) []tea.Cmd {
	for i := range s.textInputs {
		if i == int(s.idx) {
			cmds = append(cmds, s.textInputs[i].Focus())
			continue
		}
		s.textInputs[i].Blur()
	}
	return cmds
}

func (s *customStationModel) View() string {
	var b strings.Builder
	for i := range s.textInputs {
		b.WriteString(s.textInputs[i].View())
		b.WriteRune('\n')
	}
	b.WriteRune('\n')

	availHeight := s.height
	help := s.style.HelpStyle.Render(s.help.View(&s.keymap))
	availHeight -= lipgloss.Height(help)

	inputs := b.String()
	inputsHeight := lipgloss.Height(inputs)
	for i := 0; i < availHeight-inputsHeight; i++ {
		b.WriteString("\n")
	}
	return b.String() + help
}

type customStationKeymap struct {
	submit        key.Binding
	cancel        key.Binding
	nextInput     key.Binding
	prevInput     key.Binding
	prevSugg      key.Binding
	nextSugg      key.Binding
	acceptSugg    key.Binding
	showFullHelp  key.Binding
	closeFullHelp key.Binding
}

func newCustomStationKeymap() customStationKeymap {
	k := customStationKeymap{
		submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		nextInput: key.NewBinding(
			key.WithKeys("down", "tab", "ctrl+j"),
			key.WithHelp("↓/ctrl+j", "next input"),
		),
		prevInput: key.NewBinding(
			key.WithKeys("up", "shift+tab", "ctrl+k"),
			key.WithHelp("↑/ctrl+k", "prev input"),
		),
		prevSugg: key.NewBinding(
			key.WithKeys("ctrl+p", "ctrl+up"),
			key.WithHelp("ctrl+↑/ctrl+p", "prev suggestion"),
		),
		nextSugg: key.NewBinding(
			key.WithKeys("ctrl+n", "ctrl+down"),
			key.WithHelp("ctrl+↓/ctrl+n", "next suggestion"),
		),
		acceptSugg: key.NewBinding(
			key.WithKeys("right", "ctrl+right", "ctrl+l"),
			key.WithHelp("→/ctrl+→/ctrl+l", "accept suggestion"),
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
	return k
}

func (k *customStationKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.prevInput, k.nextInput, k.submit, k.cancel, k.showFullHelp}
}

func (k *customStationKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.prevInput, k.nextInput},
		{k.prevSugg, k.nextSugg, k.acceptSugg},
		{k.submit, k.cancel, k.closeFullHelp},
	}
}

func (k *customStationKeymap) setEnable(enabled bool, showAll bool) {
	k.submit.SetEnabled(enabled)
	k.cancel.SetEnabled(enabled)
	k.prevInput.SetEnabled(enabled)
	k.nextInput.SetEnabled(enabled)
	k.prevSugg.SetEnabled(enabled)
	k.nextSugg.SetEnabled(enabled)
	k.acceptSugg.SetEnabled(enabled)
	if enabled {
		k.showFullHelp.SetEnabled(!showAll)
		k.closeFullHelp.SetEnabled(showAll)
	} else {
		k.showFullHelp.SetEnabled(false)
		k.closeFullHelp.SetEnabled(false)
	}
}

func (k *customStationKeymap) update(showAll bool) {
	if showAll {
		k.nextInput.SetHelp("↓/tab/ctrl+j", "next input")
		k.prevInput.SetHelp("↑/shift+tab/ctrl+k", "prev input")
	} else {
		k.nextInput.SetHelp("↓/ctrl+j", "next input")
		k.prevInput.SetHelp("↑/ctrl+k", "prev input")
	}
}
