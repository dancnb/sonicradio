package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const (
	tabGapDistance = 2
	headerPadDist  = 2
)

var (
	favChar   = "  ★  "
	playChar  = "\u2877"
	pauseChar = "\u28FF"
	lineChar  = "\u2847"

	primaryColor    = "#ffb641"
	secondColor     = "#bd862d"
	invPrimaryColor = "#12100d"
	invSecondColor  = "#4a4133"

	basePrimaryColor     = lipgloss.Color(primaryColor)
	baseSecondColor      = lipgloss.Color(secondColor)
	invertedPrimaryColor = lipgloss.Color(invPrimaryColor)
	invertedSecondColor  = lipgloss.Color(invSecondColor)

	primaryColorStyle   = lipgloss.NewStyle().Foreground(basePrimaryColor)
	secondaryColorStyle = lipgloss.NewStyle().Foreground(baseSecondColor)

	// general
	// backgroundColor = termenv.RGBColor("#282c34")
	docStyle       = lipgloss.NewStyle().Padding(1, headerPadDist, 0, headerPadDist)
	statusBarStyle = lipgloss.NewStyle().Background(baseSecondColor).Foreground(invertedPrimaryColor)
	viewStyle      = secondaryColorStyle.PaddingLeft(headerPadDist)
	noItemsStyle   = secondaryColorStyle.PaddingLeft(3)

	// station delegate
	prefixStyle            = primaryColorStyle.PaddingLeft(1)
	nowPlayingPrefixStyle  = primaryColorStyle.PaddingLeft(0)
	nowPlayingStyle        = primaryColorStyle
	nowPlayingDescStyle    = secondaryColorStyle
	selNowPlayingStyle     = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedPrimaryColor)
	selNowPlayingDescStyle = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedSecondColor)
	itemStyle              = primaryColorStyle
	descStyle              = secondaryColorStyle
	selItemStyle           = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedPrimaryColor)
	selDescStyle           = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedSecondColor)
	selectedBorderStyle    = lipgloss.NewStyle().
				Border(lipgloss.BlockBorder(), false, false, false, true).
				BorderForeground(basePrimaryColor)

	// header
	songTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(baseSecondColor)
	italicStyle    = lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder(), false, true).
			Foreground(baseSecondColor).
			Italic(true).
			Padding(0, 0).Margin()

	// tabs
	inactiveTabBorder = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder(), true).
				Padding(0, 0).Margin(0)
	inactiveTabInner = lipgloss.NewStyle().
				Bold(false).
				Foreground(baseSecondColor)
	inactiveTabInnerHighlight = lipgloss.NewStyle().
					Bold(true).
					Foreground(basePrimaryColor)
	activeTabBorder = lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder(), true).
			Padding(0, 0).Margin(0)
	activeTabInner = lipgloss.NewStyle().
			Bold(false).
			Background(baseSecondColor).
			Foreground(invertedPrimaryColor)
	activeTabInnerHighlight = lipgloss.NewStyle().
				Bold(true).
				Background(baseSecondColor).
				Foreground(invertedPrimaryColor)
	tabGap = lipgloss.NewStyle().
		Border(lipgloss.Border{Left: " ", Right: " "}, true, false).
		Foreground(basePrimaryColor).
		BorderForeground(basePrimaryColor).
		Strikethrough(true).
		Margin(0).Padding(0)

	// help
	helpkeyStyle  = primaryColorStyle
	helpDescStyle = secondaryColorStyle
	helpStyle     = lipgloss.NewStyle().
			Padding(0, 0).Margin(0).
			Border(lipgloss.NormalBorder()).
			BorderForeground(basePrimaryColor)

	// filter
	filterPromptStyle = primaryColorStyle.Bold(true).MarginLeft(1)
	filterTextStyle   = primaryColorStyle

	//search
	searchPromptStyle = primaryColorStyle.Bold(true).MarginLeft(3)
	orderByStyle      = secondaryColorStyle
	orderBySelStyle   = primaryColorStyle

	// station info
	infoFieldNameStyle  = primaryColorStyle.Bold(false).MarginLeft(3)
	infoFieldValueStyle = secondaryColorStyle

	//history
	historyItemStyle    = secondaryColorStyle
	historyDescStyle    = primaryColorStyle.Bold(true)
	historySelItemStyle = selDescStyle
	historySelDescStyle = selItemStyle.Bold(true)
)

func padFieldName(v string) string {
	for i := len(v); i < 22; i++ {
		v += " "
	}
	return v
}

func textInputSyle(textInput *textinput.Model, prompt, placeholder string) {
	textInput.Prompt = prompt
	textInput.PromptStyle = filterPromptStyle
	textInput.TextStyle = filterTextStyle
	textInput.CompletionStyle = filterTextStyle
	textInput.Cursor.Style = filterPromptStyle
	textInput.Cursor.TextStyle = filterTextStyle
	textInput.Placeholder = placeholder
	textInput.PlaceholderStyle = secondaryColorStyle
}

func helpStyles() help.Styles {
	return help.Styles{
		ShortKey:       helpkeyStyle,
		ShortDesc:      helpDescStyle,
		ShortSeparator: helpDescStyle,
		Ellipsis:       helpDescStyle,
		FullKey:        helpkeyStyle,
		FullDesc:       helpDescStyle,
		FullSeparator:  helpDescStyle,
	}
}
