package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
)

type searchModel struct {
	enabled bool

	style *Style

	browser   *browser.Api
	countries []string
	languages []string

	inputs []FormElement
	idx    inputIdx

	orderOptions OptionList
	oIdx         orderIx

	reverse bool

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
	orderVotes orderIx = iota
	orderClickcount
	orderClicktrend
	orderBitrate
	orderName
	orderTags
	orderCountry
	orderLang
	orderCodec
	orderRand
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

var orderView = []OptionValue{
	{IdxView: 1, NameView: "Votes            "},
	{IdxView: 2, NameView: "Clicks           "},
	{IdxView: 3, NameView: "Recent trends    "},
	{IdxView: 4, NameView: "Bitrate          "},
	{IdxView: 5, NameView: "Name             "},
	{IdxView: 6, NameView: "Tags             "},
	{IdxView: 7, NameView: "Country          "},
	{IdxView: 8, NameView: "Language         "},
	{IdxView: 9, NameView: "Codecs           "},
	{IdxView: 0, NameView: "Random           "},
}

func newSearchModel(ctx context.Context, browser *browser.Api, s *Style) *searchModel {
	k := newSearchKeymap()
	inputs := []textinput.Model{
		s.NewInputModel("Name          ", "leave empty for all", &k.prevSugg, &k.nextSugg, &k.acceptSugg, nil),
		s.NewInputModel("Tags          ", "comma separated list", &k.prevSugg, &k.nextSugg, &k.acceptSugg, nil),
		s.NewInputModel("Country       ", "---", &k.prevSugg, &k.nextSugg, &k.acceptSugg, nil),
		s.NewInputModel("Language      ", "---", &k.prevSugg, &k.nextSugg, &k.acceptSugg, nil),
		s.NewInputModel("Limit         ", "---", &k.prevSugg, &k.nextSugg, &k.acceptSugg, NrInputValidator),
	}
	formElems := make([]FormElement, len(inputs))
	for ii := range inputs {
		formElems[ii] = *NewFormElement(WithTextInput(&inputs[ii]))
	}
	h := help.New()
	h.ShowAll = false
	h.ShortSeparator = "   "
	h.Styles = s.HelpStyles()

	orderOpts := NewOptionList("Order by", orderView, 0, s)
	orderOpts.SetQuick(true)
	sm := &searchModel{
		browser:      browser,
		keymap:       k,
		help:         h,
		inputs:       formElems,
		orderOptions: orderOpts,
		style:        s,
	}
	go sm.getSuggestions()
	return sm
}

func (s *searchModel) getSuggestions() {
	countries, err := s.browser.GetCountries()
	if err == nil && len(countries) > 0 {
		for i := range countries {
			s.countries = append(s.countries, countries[i].Name)
		}
		s.inputs[country].TextInput().ShowSuggestions = true
		s.inputs[country].TextInput().SetSuggestions(s.countries)
	}

	langs, err := s.browser.GetLanguages()
	if err == nil && len(langs) > 0 {
		for i := range langs {
			s.languages = append(s.languages, langs[i].Name)
		}
		s.inputs[language].TextInput().ShowSuggestions = true
		s.inputs[language].TextInput().SetSuggestions(s.languages)
	}
}

func (s *searchModel) Init() tea.Cmd {
	s.setEnabled(true)
	s.keymap.prevInput.SetHelp("↑/ctrl+k", "prev input")
	s.keymap.nextInput.SetHelp("↓/ctrl+j", "next input")
	return s.inputs[0].Focus()
}

func (s *searchModel) setSize(width, height int) {
	h, v := s.style.DocStyle.GetFrameSize()
	s.width = width - h
	s.height = height - v
	s.help.Width = s.width
}

func (s *searchModel) isEnabled() bool {
	return s.enabled
}

// setEnabled is called on search page enter/exit only
func (s *searchModel) setEnabled(v bool) {
	s.enabled = v
	s.idx = name
	for i := range s.inputs {
		s.inputs[i].Blur()
		s.inputs[i].TextInput().Reset()
	}
	s.inputs[limit].SetValue(fmt.Sprintf("%d", browser.DefLimit))
	if !v {
		s.orderOptions.SetIdx(0)
	}
	s.oIdx = orderVotes
	s.reverse = true
	showAll := false
	s.help.ShowAll = showAll
	s.keymap.setEnable(v, showAll)
}

func (s *searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "ui.searchModel.Update")
	var cmds []tea.Cmd

	if s.orderOptions.IsActive() {
		newOptions, cmd := s.orderOptions.Update(msg)
		s.orderOptions = *newOptions.(*OptionList)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.setSize(msg.Width, msg.Height)

	case OptionMsg:
		if msg.Done {
			s.orderOptions.SetFocused(false)
			s.oIdx = orderIx(msg.SelIdx)
			s.keymap.setEnable(true, s.help.ShowAll)
			cmds = s.updateInputs(cmds)
			return s, tea.Batch(cmds...)
		}

	case tea.KeyMsg:
		switch {

		case key.Matches(msg, s.keymap.order):
			if !s.orderOptions.IsActive() {
				s.orderOptions.SetFocused(true)
				s.orderOptions.SetActive(true)
				s.keymap.setEnable(false, s.help.ShowAll)
				cmds = append(cmds, s.updateInputs(cmds)...)
				return s, tea.Batch(cmds...)
			}

		case key.Matches(msg, s.keymap.showFullHelp):
			fallthrough
		case key.Matches(msg, s.keymap.closeFullHelp):
			s.help.ShowAll = !s.help.ShowAll
			s.keymap.showFullHelp.SetEnabled(!s.help.ShowAll)
			s.keymap.closeFullHelp.SetEnabled(s.help.ShowAll)
			s.keymap.update(s.help.ShowAll)
			return s, tea.Batch(cmds...)

		case key.Matches(msg, s.keymap.reverse):
			s.reverse = !s.reverse

		case key.Matches(msg, s.keymap.cancel):
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
			if msg.String() == "tab" && strings.TrimSpace(s.inputs[s.idx].Value()) != "" && s.inputs[s.idx].TextInput().ShowSuggestions {
				s.inputs[s.idx].SetValue(s.inputs[s.idx].TextInput().CurrentSuggestion())
				s.inputs[s.idx].TextInput().CursorEnd()
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
		fEl, cmd := s.inputs[i].Update(msg)
		s.inputs[i] = *fEl
		cmds = append(cmds, cmd)
	}

	return s, tea.Batch(cmds...)
}

func (s *searchModel) updateInputs(cmds []tea.Cmd) []tea.Cmd {
	for i := range s.inputs {
		if !s.orderOptions.IsActive() && i == int(s.idx) {
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
		b.WriteRune('\n')
	}
	b.WriteRune('\n')

	b.WriteString(s.orderOptions.View())
	b.WriteRune('\n')
	b.WriteRune('\n')

	b.WriteString(s.style.PromptStyle.Render(PadFieldName("Reverse       ", nil)))
	rev := "off"
	if s.reverse {
		rev = "on"
	}
	b.WriteString(s.style.PrimaryColorStyle.Render(rev))

	availHeight := s.height
	var help string
	if !s.orderOptions.IsActive() {
		help = s.style.HelpStyle.Render(s.help.View(&s.keymap))
	} else {
		help = s.style.HelpStyle.Render(s.help.View(&s.orderOptions.Keymap))
	}
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
			key.WithHelp("ctrl+o", "order by "),
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

func (k *searchKeymap) setEnable(enabled bool, showAll bool) {
	k.submit.SetEnabled(enabled)
	k.cancel.SetEnabled(enabled)
	k.prevInput.SetEnabled(enabled)
	k.nextInput.SetEnabled(enabled)
	k.order.SetEnabled(enabled)
	k.reverse.SetEnabled(enabled)
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

func (k *searchKeymap) update(showAll bool) {
	if showAll {
		k.nextInput.SetHelp("↓/tab/ctrl+j", "next input")
		k.prevInput.SetHelp("↑/shift+tab/ctrl+k", "prev input")
	} else {
		k.nextInput.SetHelp("↓/ctrl+j", "next input")
		k.prevInput.SetHelp("↑/ctrl+k", "prev input")
	}
}
