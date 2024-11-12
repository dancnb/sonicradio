package ui

import (
	"fmt"
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
	enabled bool

	browser   *browser.Api
	countries []string
	languages []string

	inputs      []textinput.Model
	idx         inputIdx
	orderActive bool
	oIdx        orderIx
	reverse     bool

	keymap searchKeymap
	help   help.Model
	width  int
	height int
}

type inputIdx byte

const (
	name inputIdx = iota
	tags
	country
	language
	limit
)

type orderIx uint8

const (
	orderRand orderIx = iota
	orderVotes
	orderClickcount
	orderClicktrend
	orderBitrate
	orderName
	orderTags
	orderCountry
	orderLang
	orderCodec
)

func (o orderIx) toSearchOrder() browser.OrderBy {
	return searchOrder[o]
}

var searchOrder = map[orderIx]browser.OrderBy{
	orderVotes:      browser.Votes,
	orderClickcount: browser.Clickcount,
	orderClicktrend: browser.Clicktrend,
	orderBitrate:    browser.Bitrate,
	orderName:       browser.Name,
	orderTags:       browser.Tags,
	orderCountry:    browser.CountryOrder,
	orderLang:       browser.LanguageOrder,
	orderCodec:      browser.Codec,
	orderRand:       browser.Random,
}

var orderView = map[orderIx]string{
	orderVotes:      "Votes",
	orderClickcount: "Clicks",
	orderClicktrend: "Recent trends",
	orderBitrate:    "Bitrate",
	orderName:       "Name",
	orderTags:       "Tags",
	orderCountry:    "Country",
	orderLang:       "Language",
	orderCodec:      "Codecs",
	orderRand:       "Random",
}

func newSearchModel(browser *browser.Api) *searchModel {
	k := newSearchKeymap()
	inputs := []textinput.Model{
		makeInput("Name          ", "leave empty for all", k),
		makeInput("Tags          ", "comma separated list", k),
		makeInput("Country       ", "---", k),
		makeInput("Language      ", "---", k),
		makeInput("Limit         ", "---", k),
	}
	inputs[limit].Validate = func(s string) error {
		_, err := strconv.Atoi(s)
		return err
	}

	h := help.New()
	h.ShowAll = false
	h.ShortSeparator = "   "
	h.Styles = helpStyles()

	sm := &searchModel{
		browser: browser,
		keymap:  k,
		help:    h,
		inputs:  inputs,
	}
	go sm.getSuggestions()
	return sm
}

func makeInput(prompt, placeholder string, keymap searchKeymap) textinput.Model {
	input := textinput.New()
	input.Cursor.SetMode(cursor.CursorBlink)
	prompt = padFieldName(prompt)
	textInputSyle(&input, prompt, placeholder)
	input.PromptStyle = searchPromptStyle
	input.KeyMap.NextSuggestion = keymap.nextSugg
	input.KeyMap.PrevSuggestion = keymap.prevSugg
	input.KeyMap.AcceptSuggestion = keymap.acceptSugg
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
	s.keymap.prevInput.SetHelp("↑/ctrl+k", "prev input")
	s.keymap.nextInput.SetHelp("↓/ctrl+j", "next input")
	return s.inputs[0].Focus()
}

func (s *searchModel) setSize(width, height int) {
	h, v := docStyle.GetFrameSize()
	s.width = width - h
	s.height = height - v
	s.help.Width = s.width
}

func (s *searchModel) isEnabled() bool {
	return s.enabled
}

func (s *searchModel) setEnabled(v bool) {
	s.enabled = v
	s.idx = name
	for i := range s.inputs {
		s.inputs[i].Blur()
		s.inputs[i].Reset()
	}
	s.inputs[limit].SetValue(fmt.Sprintf("%d", browser.DefLimit))
	s.orderActive = false
	s.oIdx = orderVotes
	s.reverse = true
	s.keymap.setEnable(v)
	s.help.ShowAll = false
}

func (s *searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "ui.searchModel.Update")
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

		case key.Matches(msg, s.keymap.order):
			if !s.orderActive {
				s.orderActive = true
				s.keymap.setEnable(false)
				s.keymap.cancel.SetEnabled(true)
				s.keymap.setEnableOrderKeys(true)
				cmds = s.updateInputs(cmds)
			}
		case key.Matches(msg, s.keymap.orderkeys...):
			ord, err := strconv.Atoi(msg.String())
			if err == nil {
				s.oIdx = orderIx(ord)
				s.orderActive = false
				s.keymap.setEnable(true)
				s.keymap.setEnableOrderKeys(false)
				cmds = s.updateInputs(cmds)
				return s, tea.Batch(cmds...)
			}

		case key.Matches(msg, s.keymap.reverse):
			s.reverse = !s.reverse

		case key.Matches(msg, s.keymap.cancel):
			if s.orderActive {
				s.orderActive = false
				s.keymap.setEnable(true)
				s.keymap.setEnableOrderKeys(false)
				cmds = s.updateInputs(cmds)
			} else {
				return s, func() tea.Msg {
					s.setEnabled(false)
					return searchRespMsg{cancelled: true}
				}
			}

		case key.Matches(msg, s.keymap.submit):
			return s, func() tea.Msg {
				defer s.setEnabled(false)

				params := browser.DefaultSearchParams()
				params.Name = strings.TrimSpace(s.inputs[name].Value())
				params.TagList = strings.TrimSpace(s.inputs[tags].Value())
				params.Country = strings.Title(strings.TrimSpace(s.inputs[country].Value()))
				params.Language = strings.TrimSpace(s.inputs[language].Value())
				limit, err := strconv.Atoi(strings.TrimSpace(s.inputs[limit].Value()))
				if err == nil {
					params.Limit = limit
				}
				params.Order = s.oIdx.toSearchOrder()
				params.Reverse = s.reverse

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
			if msg.String() == "tab" && strings.TrimSpace(s.inputs[s.idx].Value()) != "" && s.inputs[s.idx].ShowSuggestions {
				s.inputs[s.idx].SetValue(s.inputs[s.idx].CurrentSuggestion())
				s.inputs[s.idx].CursorEnd()
			}
			s.idx++
			s.idx = s.idx % inputIdx(len(s.inputs))
			cmds = s.updateInputs(cmds)
		case key.Matches(msg, s.keymap.prevInput):
			if s.idx == 0 {
				s.idx = limit
			}
			s.idx--
			cmds = s.updateInputs(cmds)
		}
	}

	for i := range s.inputs {
		var cmd tea.Cmd
		s.inputs[i], cmd = s.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return s, tea.Batch(cmds...)
}

func (s *searchModel) updateInputs(cmds []tea.Cmd) []tea.Cmd {
	for i := range s.inputs {
		if !s.orderActive && i == int(s.idx) {
			cmds = append(cmds, s.inputs[i].Focus())
			continue
		}
		s.inputs[i].Blur()
	}
	return cmds
}

func (s *searchModel) getOrderString(o orderIx) string {
	idx := o
	return fmt.Sprintf("%d. %s", idx, orderView[o])
}

func (s *searchModel) getOrderStyle(o orderIx) lipgloss.Style {
	if s.oIdx == o {
		return orderBySelStyle
	}
	return orderByStyle
}

func (s *searchModel) View() string {
	var b strings.Builder
	for i := range s.inputs {
		b.WriteString(s.inputs[i].View())
		b.WriteRune('\n')
	}

	b.WriteRune('\n')
	b.WriteRune('\n')
	orderPrompt := "Order by      "
	if s.orderActive {
		orderPrompt = "Enter #       "
	}
	b.WriteString(searchPromptStyle.Render(padFieldName(orderPrompt)))
	ordS := s.getOrderString(orderVotes)
	b.WriteString(s.getOrderStyle(orderVotes).Render(ordS))
	b.WriteRune('\n')
	for i := orderClickcount; i <= orderCodec; i++ {
		b.WriteString(searchPromptStyle.Render(padFieldName("")))
		ordS := s.getOrderString(i)
		b.WriteString(s.getOrderStyle(i).Render(ordS))
		b.WriteRune('\n')
	}
	b.WriteString(searchPromptStyle.Render(padFieldName("")))
	ordS = s.getOrderString(orderRand)
	b.WriteString(s.getOrderStyle(orderRand).Render(ordS))
	b.WriteRune('\n')

	b.WriteRune('\n')
	b.WriteString(searchPromptStyle.Render(padFieldName("Reverse       ")))
	rev := "off"
	if s.reverse {
		rev = "on"
	}
	b.WriteString(filterTextStyle.Render(rev))

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

type searchKeymap struct {
	submit        key.Binding
	cancel        key.Binding
	nextInput     key.Binding
	prevInput     key.Binding
	order         key.Binding
	orderkeys     []key.Binding
	reverse       key.Binding
	prevSugg      key.Binding
	nextSugg      key.Binding
	acceptSugg    key.Binding
	showFullHelp  key.Binding
	closeFullHelp key.Binding
}

func newSearchKeymap() searchKeymap {
	k := searchKeymap{
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
		order: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "order by"),
		),
		reverse: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reverse"),
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
	for i := orderRand; i <= orderCodec; i++ {
		x := fmt.Sprintf("%d", i)
		ordkey := key.NewBinding(key.WithKeys(x))
		ordkey.SetEnabled(false)
		k.orderkeys = append(k.orderkeys, ordkey)
	}
	return k
}

func (k *searchKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.prevInput, k.nextInput, k.order, k.reverse, k.submit, k.cancel, k.showFullHelp}
}

func (k *searchKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.prevInput, k.nextInput},
		{k.prevSugg, k.nextSugg, k.acceptSugg},
		{k.order, k.reverse},
		{k.submit, k.cancel, k.closeFullHelp},
	}
}

func (k *searchKeymap) setEnable(v bool) {
	k.submit.SetEnabled(v)
	k.cancel.SetEnabled(v)
	k.prevInput.SetEnabled(v)
	k.nextInput.SetEnabled(v)
	k.order.SetEnabled(v)
	k.reverse.SetEnabled(v)
	k.prevSugg.SetEnabled(v)
	k.nextSugg.SetEnabled(v)
	k.acceptSugg.SetEnabled(v)
	k.showFullHelp.SetEnabled(v)
	k.closeFullHelp.SetEnabled(false)
	k.setEnableOrderKeys(false)
}

func (k *searchKeymap) setEnableOrderKeys(v bool) {
	for i := range k.orderkeys {
		k.orderkeys[i].SetEnabled(v)
	}
}

func (k *searchKeymap) update(showAll bool) {
	if showAll {
		k.nextInput.SetHelp("↓/tab/ctrl+j", "next input")
		k.prevInput.SetHelp("↑/shift+tab/ctrl+k", "prev input")
	} else {
		k.nextInput.SetHelp("↓/ctrl+j", "next input")
		k.prevInput.SetHelp("↑/ctrl+k", "prev input")
	}
}
