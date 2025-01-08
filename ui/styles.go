package ui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const (
	tabGapDistance = 2
	headerPadDist  = 2

	favChar   = "  ★  "
	playChar  = "\u2877"
	pauseChar = "\u28FF"
	lineChar  = "\u2847"
)

type style struct {
	primaryColor    string
	secondColor     string
	invPrimaryColor string
	invSecondColor  string

	basePrimaryColor     lipgloss.AdaptiveColor
	baseSecondColor      lipgloss.AdaptiveColor
	invertedPrimaryColor lipgloss.AdaptiveColor
	invertedSecondColor  lipgloss.AdaptiveColor

	primaryColorStyle   lipgloss.Style
	secondaryColorStyle lipgloss.Style

	// general
	docStyle       lipgloss.Style
	statusBarStyle lipgloss.Style
	viewStyle      lipgloss.Style
	noItemsStyle   lipgloss.Style

	// station delegate
	prefixStyle            lipgloss.Style
	nowPlayingPrefixStyle  lipgloss.Style
	selNowPlayingStyle     lipgloss.Style
	selNowPlayingDescStyle lipgloss.Style
	selItemStyle           lipgloss.Style
	selDescStyle           lipgloss.Style
	selectedBorderStyle    lipgloss.Style

	// header
	songTitleStyle lipgloss.Style
	italicStyle    lipgloss.Style

	// tabs
	inactiveTabBorder         lipgloss.Style
	inactiveTabInner          lipgloss.Style
	inactiveTabInnerHighlight lipgloss.Style
	activeTabBorder           lipgloss.Style
	activeTabInner            lipgloss.Style
	activeTabInnerHighlight   lipgloss.Style
	tabGap                    lipgloss.Style

	// help
	helpStyle lipgloss.Style
	// filter
	filterPromptStyle lipgloss.Style

	// search
	searchPromptStyle lipgloss.Style

	// station info
	infoFieldNameStyle lipgloss.Style

	// history
	historyDescStyle    lipgloss.Style
	historySelItemStyle lipgloss.Style
	historySelDescStyle lipgloss.Style
}

func newStyle(themeIdx int) *style {
	t := themes[themeIdx%len(themes)]
	u := style{
		docStyle: lipgloss.NewStyle().
			Padding(1, headerPadDist, 0, headerPadDist),
		inactiveTabBorder: lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder(), true).
			Padding(0, 0).Margin(0),
		activeTabBorder: lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder(), true).
			Padding(0, 0).Margin(0),
	}
	u.setTheme(t)
	return &u
}

func (s *style) setTheme(t theme) {
	s.primaryColor = t.primaryColor
	s.secondColor = t.secondColor
	s.invPrimaryColor = t.invPrimaryColor
	s.invSecondColor = t.invSecondColor

	s.basePrimaryColor = lipgloss.AdaptiveColor{Light: t.primaryColor, Dark: t.primaryColor}
	s.baseSecondColor = lipgloss.AdaptiveColor{Light: t.secondColor, Dark: t.secondColor}
	s.invertedPrimaryColor = lipgloss.AdaptiveColor{Light: t.invPrimaryColor, Dark: t.invPrimaryColor}
	s.invertedSecondColor = lipgloss.AdaptiveColor{Light: t.invSecondColor, Dark: t.invSecondColor}

	s.primaryColorStyle = lipgloss.NewStyle().Foreground(s.basePrimaryColor)
	s.secondaryColorStyle = lipgloss.NewStyle().Foreground(s.baseSecondColor)
	//
	// general
	// u.docStyle = lipgloss.NewStyle().Padding(1, headerPadDist, 0, headerPadDist)
	s.statusBarStyle = lipgloss.NewStyle().Background(s.baseSecondColor).Foreground(s.invertedPrimaryColor)
	s.viewStyle = s.secondaryColorStyle.PaddingLeft(headerPadDist)
	s.noItemsStyle = s.secondaryColorStyle.PaddingLeft(3)

	// station delegate
	s.prefixStyle = s.primaryColorStyle.PaddingLeft(1)
	s.nowPlayingPrefixStyle = s.primaryColorStyle.PaddingLeft(0)
	s.selNowPlayingStyle = lipgloss.NewStyle().Background(s.basePrimaryColor).Foreground(s.invertedPrimaryColor)
	s.selNowPlayingDescStyle = lipgloss.NewStyle().Background(s.basePrimaryColor).Foreground(s.invertedSecondColor)
	s.selItemStyle = lipgloss.NewStyle().Background(s.basePrimaryColor).Foreground(s.invertedPrimaryColor)
	s.selDescStyle = lipgloss.NewStyle().Background(s.basePrimaryColor).Foreground(s.invertedSecondColor)
	s.selectedBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.BlockBorder(), false, false, false, true).
		BorderForeground(s.basePrimaryColor)

	// header
	s.songTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(s.baseSecondColor)
	s.italicStyle = lipgloss.NewStyle().
		Border(lipgloss.HiddenBorder(), false, true).
		Foreground(s.baseSecondColor).
		Italic(true).
		Padding(0, 0).Margin()

	// tabs
	s.inactiveTabInner = lipgloss.NewStyle().
		Bold(false).
		Foreground(s.baseSecondColor)
	s.inactiveTabInnerHighlight = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.basePrimaryColor)
	s.activeTabInner = lipgloss.NewStyle().
		Bold(false).
		Background(s.baseSecondColor).
		Foreground(s.invertedPrimaryColor)
	s.activeTabInnerHighlight = lipgloss.NewStyle().
		Bold(true).
		Background(s.baseSecondColor).
		Foreground(s.invertedPrimaryColor)
	s.tabGap = lipgloss.NewStyle().
		Border(lipgloss.Border{Left: " ", Right: " "}, true, false).
		Foreground(s.basePrimaryColor).
		BorderForeground(s.basePrimaryColor).
		Strikethrough(true).
		Margin(0).Padding(0)

	// help
	s.helpStyle = lipgloss.NewStyle().
		Padding(0, 0).Margin(0).
		Border(lipgloss.NormalBorder()).
		BorderForeground(s.basePrimaryColor)

	// filter
	s.filterPromptStyle = s.primaryColorStyle.Bold(true).MarginLeft(1)

	// search
	s.searchPromptStyle = s.primaryColorStyle.Bold(true).MarginLeft(headerPadDist + tabGapDistance)

	// station info
	s.infoFieldNameStyle = s.primaryColorStyle.Bold(false).MarginLeft(headerPadDist + tabGapDistance)

	// history
	s.historyDescStyle = s.primaryColorStyle.Bold(true)
	s.historySelItemStyle = s.selDescStyle
	s.historySelDescStyle = s.selItemStyle.Bold(true)
}

func (s *style) helpStyles() help.Styles {
	return help.Styles{
		ShortKey:       s.primaryColorStyle,
		ShortDesc:      s.secondaryColorStyle,
		ShortSeparator: s.secondaryColorStyle,
		Ellipsis:       s.secondaryColorStyle,
		FullKey:        s.primaryColorStyle,
		FullDesc:       s.secondaryColorStyle,
		FullSeparator:  s.secondaryColorStyle,
	}
}

func (s *style) newInputModel(prompt, placeholder string,
	prevSugg *key.Binding,
	nextSugg *key.Binding,
	acceptSugg *key.Binding,
	validator textinput.ValidateFunc,
) textinput.Model {
	input := textinput.New()
	input.Cursor.SetMode(cursor.CursorBlink)
	prompt = padFieldName(prompt)
	s.textInputSyle(&input, prompt, placeholder)
	input.PromptStyle = s.searchPromptStyle
	if prevSugg != nil {
		input.KeyMap.NextSuggestion = *nextSugg
	}
	if nextSugg != nil {
		input.KeyMap.NextSuggestion = *nextSugg
	}
	if acceptSugg != nil {
		input.KeyMap.NextSuggestion = *nextSugg
	}
	if validator != nil {
		input.Validate = validator
	}
	return input
}

func (s *style) textInputSyle(textInput *textinput.Model, prompt, placeholder string) {
	textInput.Prompt = prompt
	textInput.PromptStyle = s.filterPromptStyle
	textInput.TextStyle = s.primaryColorStyle
	textInput.CompletionStyle = s.primaryColorStyle
	textInput.Cursor.Style = s.filterPromptStyle
	textInput.Cursor.TextStyle = s.primaryColorStyle
	textInput.Placeholder = placeholder
	textInput.PlaceholderStyle = s.secondaryColorStyle
}

func nrInputValidator(s string) error {
	_, err := strconv.Atoi(s)
	return err
}

const maxFieldLen = 26

func padFieldName(v string) string {
	var b strings.Builder
	b.WriteString(v)
	for i := len(v); i < maxFieldLen; i++ {
		b.WriteString(" ")
	}
	return b.String()
}
