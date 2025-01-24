package styles

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const (
	TabGapDistance = 2
	HeaderPadDist  = 2

	FavChar      = "  ★"
	AutoplayChar = " Auto"
	PlayChar     = "\u2877"
	PauseChar    = "\u28FF"
	LineChar     = "\u2847"
)

type Style struct {
	theme string

	basePrimaryColor       lipgloss.AdaptiveColor
	baseSecondaryColor     lipgloss.AdaptiveColor
	invertedPrimaryColor   lipgloss.AdaptiveColor
	invertedSecondaryColor lipgloss.AdaptiveColor

	PrimaryColorStyle   lipgloss.Style
	SecondaryColorStyle lipgloss.Style

	// general
	BaseBold       lipgloss.Style
	DocStyle       lipgloss.Style
	StatusBarStyle lipgloss.Style
	ViewStyle      lipgloss.Style
	NoItemsStyle   lipgloss.Style

	// station delegate
	PrefixStyle                 lipgloss.Style
	NowPlayingPrefixStyle       lipgloss.Style
	SelNowPlayingStyle          lipgloss.Style
	SelNowPlayingDescStyle      lipgloss.Style
	SelItemStyle                lipgloss.Style
	SelDescStyle                lipgloss.Style
	SelectedBorderStyle         lipgloss.Style
	SelectedBorderStyleInactive lipgloss.Style

	// header
	SongTitleStyle lipgloss.Style
	ItalicStyle    lipgloss.Style

	// tabs
	InactiveTabBorder         lipgloss.Style
	InactiveTabInner          lipgloss.Style
	InactiveTabInnerHighlight lipgloss.Style
	ActiveTabBorder           lipgloss.Style
	ActiveTabInner            lipgloss.Style
	ActiveTabInnerHighlight   lipgloss.Style
	TabGap                    lipgloss.Style

	// help
	HelpStyle lipgloss.Style
	// filter
	filterPromptStyle lipgloss.Style

	// textInput
	PromptStyle    lipgloss.Style
	SelPromptStyle lipgloss.Style

	// station info
	InfoFieldNameStyle lipgloss.Style

	// history
	HistoryDescStyle    lipgloss.Style
	HistorySelItemStyle lipgloss.Style
	HistorySelDescStyle lipgloss.Style
}

func NewStyle(themeIdx int) *Style {
	t := Themes[themeIdx%len(Themes)]
	u := Style{
		BaseBold: lipgloss.NewStyle().Bold(true),
		DocStyle: lipgloss.NewStyle().
			Padding(1, HeaderPadDist, 0, HeaderPadDist),
		InactiveTabBorder: lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder(), true).
			Padding(0, 0).Margin(0),
		ActiveTabBorder: lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder(), true).
			Padding(0, 0).Margin(0),
	}
	u.setTheme(t)
	return &u
}

func (s *Style) SetThemeIdx(themeIdx int) {
	t := Themes[themeIdx%len(Themes)]
	if s.theme == t.Name {
		return
	}
	s.setTheme(t)
}

func (s *Style) setTheme(t Theme) {
	s.theme = t.Name
	s.basePrimaryColor = lipgloss.AdaptiveColor{Light: t.Light.primaryColor, Dark: t.Dark.primaryColor}
	s.baseSecondaryColor = lipgloss.AdaptiveColor{Light: t.Light.secondaryColor, Dark: t.Dark.secondaryColor}
	s.invertedPrimaryColor = lipgloss.AdaptiveColor{Light: t.Light.invertedPrimaryColor, Dark: t.Dark.invertedPrimaryColor}
	s.invertedSecondaryColor = lipgloss.AdaptiveColor{Light: t.Light.invertedSecondaryColor, Dark: t.Dark.invertedSecondaryColor}

	s.PrimaryColorStyle = lipgloss.NewStyle().Foreground(s.basePrimaryColor)
	s.SecondaryColorStyle = lipgloss.NewStyle().Foreground(s.baseSecondaryColor)
	//
	// general
	s.StatusBarStyle = lipgloss.NewStyle().Background(s.baseSecondaryColor).Foreground(s.invertedPrimaryColor)
	s.ViewStyle = s.SecondaryColorStyle.PaddingLeft(HeaderPadDist)
	s.NoItemsStyle = s.SecondaryColorStyle.PaddingLeft(3)

	// station delegate
	s.PrefixStyle = s.PrimaryColorStyle.PaddingLeft(1)
	s.NowPlayingPrefixStyle = s.PrimaryColorStyle.PaddingLeft(0)
	s.SelNowPlayingStyle = lipgloss.NewStyle().Background(s.basePrimaryColor).Foreground(s.invertedPrimaryColor)
	s.SelNowPlayingDescStyle = lipgloss.NewStyle().Background(s.basePrimaryColor).Foreground(s.invertedSecondaryColor)
	s.SelItemStyle = lipgloss.NewStyle().Background(s.basePrimaryColor).Foreground(s.invertedPrimaryColor)
	s.SelDescStyle = lipgloss.NewStyle().Background(s.basePrimaryColor).Foreground(s.invertedSecondaryColor)
	s.SelectedBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.BlockBorder(), false, false, false, true).
		BorderForeground(s.basePrimaryColor)
	s.SelectedBorderStyleInactive = lipgloss.NewStyle().Inherit(s.SelectedBorderStyle).BorderForeground(s.baseSecondaryColor)

	// header
	s.SongTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(s.baseSecondaryColor)
	s.ItalicStyle = lipgloss.NewStyle().
		Border(lipgloss.HiddenBorder(), false, true).
		Foreground(s.baseSecondaryColor).
		Italic(true).
		Padding(0, 0).Margin()

	// tabs
	s.InactiveTabInner = lipgloss.NewStyle().
		Bold(false).
		Foreground(s.baseSecondaryColor)
	s.InactiveTabInnerHighlight = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.basePrimaryColor)
	s.ActiveTabInner = lipgloss.NewStyle().
		Bold(false).
		Background(s.baseSecondaryColor).
		Foreground(s.invertedPrimaryColor)
	s.ActiveTabInnerHighlight = lipgloss.NewStyle().
		Bold(true).
		Background(s.baseSecondaryColor).
		Foreground(s.invertedPrimaryColor)
	s.TabGap = lipgloss.NewStyle().
		Border(lipgloss.Border{Left: " ", Right: " "}, true, false).
		Foreground(s.basePrimaryColor).
		BorderForeground(s.basePrimaryColor).
		Strikethrough(true).
		Margin(0).Padding(0)

	// help
	s.HelpStyle = lipgloss.NewStyle().
		Padding(0, 0).Margin(0).
		Border(lipgloss.NormalBorder()).
		BorderForeground(s.basePrimaryColor)

	prompyStyleBase := s.PrimaryColorStyle.Bold(true)

	// filter
	s.filterPromptStyle = prompyStyleBase.MarginLeft(TabGapDistance)

	// search
	s.PromptStyle = prompyStyleBase.MarginLeft(HeaderPadDist + TabGapDistance)
	s.SelPromptStyle = lipgloss.NewStyle().Background(s.basePrimaryColor).Foreground(s.invertedPrimaryColor).
		Bold(true).MarginLeft(HeaderPadDist + TabGapDistance)

	// station info
	s.InfoFieldNameStyle = s.PrimaryColorStyle.Bold(false).MarginLeft(HeaderPadDist + TabGapDistance)

	// history
	s.HistoryDescStyle = s.PrimaryColorStyle.Bold(true)
	s.HistorySelItemStyle = s.SelDescStyle
	s.HistorySelDescStyle = s.SelItemStyle.Bold(true)
}

func (s *Style) GetSecondColor() string {
	hasDark := lipgloss.DefaultRenderer().HasDarkBackground()
	if hasDark {
		return s.baseSecondaryColor.Dark
	} else {
		return s.baseSecondaryColor.Light
	}
}

func (s *Style) HelpStyles() help.Styles {
	return help.Styles{
		ShortKey:       s.PrimaryColorStyle,
		ShortDesc:      s.SecondaryColorStyle,
		ShortSeparator: s.SecondaryColorStyle,
		Ellipsis:       s.SecondaryColorStyle,
		FullKey:        s.PrimaryColorStyle,
		FullDesc:       s.SecondaryColorStyle,
		FullSeparator:  s.SecondaryColorStyle,
	}
}

func (s *Style) NewInputModel(
	prompt, placeholder string,
	prevSugg *key.Binding,
	nextSugg *key.Binding,
	acceptSugg *key.Binding,
	validator textinput.ValidateFunc,
) textinput.Model {
	input := textinput.New()
	input.Cursor.SetMode(cursor.CursorBlink)
	prompt = PadFieldName(prompt, nil)
	s.TextInputSyle(&input, prompt, placeholder)
	input.PromptStyle = s.PromptStyle
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

func (s *Style) TextInputSyle(textInput *textinput.Model, prompt, placeholder string) {
	textInput.Prompt = prompt
	textInput.PromptStyle = s.filterPromptStyle
	textInput.TextStyle = s.PrimaryColorStyle
	textInput.CompletionStyle = s.PrimaryColorStyle
	textInput.Cursor.Style = s.filterPromptStyle
	textInput.Cursor.TextStyle = s.PrimaryColorStyle
	textInput.Placeholder = placeholder
	textInput.PlaceholderStyle = s.SecondaryColorStyle
}

func NrInputValidator(s string) error {
	_, err := strconv.Atoi(s)
	return err
}

const MaxFieldLen = 30

func PadFieldName(v string, padAmt *int) string {
	amt := MaxFieldLen
	if padAmt != nil {
		amt = *padAmt
	}
	var b strings.Builder
	b.WriteString(v)
	for i := len(v); i < amt; i++ {
		b.WriteString(" ")
	}
	return b.String()
}

const IndexStringPadAmt = 3

func IndexString(index int) string {
	prefix := fmt.Sprintf("%d. ", index)
	if index < 10 {
		prefix = fmt.Sprintf("%s%s", strings.Repeat(" ", IndexStringPadAmt), prefix)
	} else if index < 100 {
		prefix = fmt.Sprintf("%s%s", strings.Repeat(" ", IndexStringPadAmt-1), prefix)
	} else if index < 1000 {
		prefix = fmt.Sprintf("%s%s", strings.Repeat(" ", IndexStringPadAmt-2), prefix)
	}
	return prefix
}
