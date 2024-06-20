package ui

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
)

type searchModel struct {
	browser *browser.Api
	enabled bool
	inputs  []textinput.Model
	idx     int

	keymap searchKeymap
	help   help.Model
	width  int
	height int

	countries []string
	languages []string
}

type inputIdx byte

const (
	name inputIdx = iota
	tags
	country
	state
	language
	limit
)

func newSearchModel(browser *browser.Api) *searchModel {
	inputs := []textinput.Model{
		makeInput("Name          ", "---"),
		makeInput("Tags          ", "comma separated list"),
		makeInput("Country       ", "---"),
		makeInput("State         ", "---"), //todo add suggestions from states by country req
		makeInput("Language      ", "---"), //todo add suggestions from languages req
		makeInput("Limit         ", "---"),
	}
	inputs[limit].Validate = func(s string) error {
		_, err := strconv.Atoi(s)
		return err
	}

	h := help.New()
	h.ShortSeparator = "   "
	h.Styles = helpStyles()

	sm := &searchModel{
		browser: browser,
		keymap:  newSearchKeymap(),
		help:    h,
		inputs:  inputs,
	}
	go sm.getSuggestions()
	return sm
}

func makeInput(prompt, placeholder string) textinput.Model {
	input := textinput.New()
	input.Cursor.SetMode(cursor.CursorBlink)
	textInputSyle(&input, prompt, placeholder)
	input.PromptStyle = searchPromptStyle
	return input
}

func (s *searchModel) getSuggestions() {
	countries, err := s.browser.GetCountries()
	if err == nil && len(countries) > 0 {
		for i := range countries {
			s.countries = append(s.countries, countries[i].Name)
		}
		s.inputs[country].ShowSuggestions = true
		s.inputs[country].SetSuggestions(s.countries)
	}

	langs, err := s.browser.GetLanguages()
	if err == nil && len(langs) > 0 {
		for i := range langs {
			s.languages = append(s.languages, langs[i].Name)
		}
		s.inputs[language].ShowSuggestions = true
		s.inputs[language].SetSuggestions(s.languages)
	}
}

func (s *searchModel) Init() tea.Cmd {
	s.setEnabled(true)
	return s.inputs[0].Focus()
}

func (s *searchModel) setSize(width, height int) {
	h, v := docStyle.GetFrameSize()
	s.width = width - h
	s.height = height - v
	s.help.Width = s.width
}

func (s *searchModel) isEnabled() bool {
	slog.Debug("searchModel", "enabled", s.enabled)
	return s.enabled
}

func (s *searchModel) setEnabled(v bool) {
	s.enabled = v
	s.idx = 0
	for i := range s.inputs {
		s.inputs[i].Blur()
		s.inputs[i].Reset()
	}
	s.inputs[limit].SetValue(fmt.Sprintf("%d", browser.DefLimit))
	s.keymap.setEnable(v)
}

func (s *searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "update searchModel")
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.setSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.cancelSearch):
			return s, func() tea.Msg {
				s.setEnabled(false)
				return searchRespMsg{cancelled: true}
			}

		case key.Matches(msg, s.keymap.submit):
			return s, func() tea.Msg {
				defer s.setEnabled(false)

				params := browser.DefaultSearchParams()
				params.Name = strings.TrimSpace(s.inputs[name].Value())
				params.TagList = strings.TrimSpace(s.inputs[tags].Value())
				params.Country = strings.Title(strings.TrimSpace(s.inputs[country].Value()))
				params.State = strings.TrimSpace(s.inputs[state].Value())
				params.Language = strings.TrimSpace(s.inputs[language].Value())
				limit, err := strconv.Atoi(strings.TrimSpace(s.inputs[limit].Value()))
				if err == nil {
					params.Limit = limit
				}

				stations, err := s.browser.Search(params)
				res := searchRespMsg{stations: stations}
				if err != nil {
					res.statusMsg = statusMsg(err.Error())
				} else if len(stations) == 0 {
					res.viewMsg = noStationsFound
				}
				return res
			}

		case key.Matches(msg, s.keymap.nextInput):
			s.idx++
			s.idx = s.idx % len(s.inputs)
			cmds = s.updateFocusedInput(cmds)
			return s, tea.Batch(cmds...)
		case key.Matches(msg, s.keymap.prevInput):
			s.idx--
			if s.idx < 0 {
				s.idx = len(s.inputs) - 1
			}
			cmds = s.updateFocusedInput(cmds)
			return s, tea.Batch(cmds...)
		}
	}

	for i := range s.inputs {
		var cmd tea.Cmd
		s.inputs[i], cmd = s.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return s, tea.Batch(cmds...)
}

func (s *searchModel) updateFocusedInput(cmds []tea.Cmd) []tea.Cmd {
	for i := range s.inputs {
		if i == s.idx {
			cmds = append(cmds, s.inputs[i].Focus())
			continue
		}
		s.inputs[i].Blur()
	}
	return cmds
}

func (s *searchModel) View() string {
	var b strings.Builder
	for i := range s.inputs {
		b.WriteString(s.inputs[i].View())
		if i < len(s.inputs)-1 {
			b.WriteRune('\n')
		}
	}
	inputs := b.String()

	availHeight := s.height
	help := helpStyle.Render(s.help.View(s.keymap))
	availHeight -= lipgloss.Height(help)

	for i := 0; i < availHeight-lipgloss.Height(inputs); i++ {
		b.WriteString("\n")
	}
	return b.String() + "\n" + help
}

type searchKeymap struct {
	submit       key.Binding
	cancelSearch key.Binding
	nextInput    key.Binding
	prevInput    key.Binding
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
		nextInput: key.NewBinding(
			key.WithKeys("down", "ctrl+n"),
			key.WithHelp("↓/ctrl+n", "next input"),
		),
		prevInput: key.NewBinding(
			key.WithKeys("up", "ctrl+p"),
			key.WithHelp("↑/ctrl+p", "prev input"),
		),
	}
}

func (k searchKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.prevInput, k.nextInput, k.submit, k.cancelSearch}
}

func (k searchKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(), //first column
		// second column ...
	}
}

func (k searchKeymap) setEnable(v bool) {
	k.prevInput.SetEnabled(v)
	k.nextInput.SetEnabled(v)
	k.submit.SetEnabled(v)
	k.cancelSearch.SetEnabled(v)
}
