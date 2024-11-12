package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	tabGapDistance = 2
	padDist        = 2
)

var (
	favChar   = "  ★  "
	playChar  = "\u2877"
	pauseChar = "\u28FF"
	lineChar  = "\u2847"

	basePrimaryColor     = lipgloss.Color("#ffb641")
	baseSecondColor      = lipgloss.Color("#bd862d")
	invertedPrimaryColor = lipgloss.Color("#12100d")
	invertedSecondColor  = lipgloss.Color("#4a4133")

	primaryColorStyle   = lipgloss.NewStyle().Foreground(basePrimaryColor)
	secondaryColorStyle = lipgloss.NewStyle().Foreground(baseSecondColor)

	prefixStyle           = primaryColorStyle.PaddingLeft(1)
	nowPlayingPrefixStyle = primaryColorStyle.PaddingLeft(0)

	nowPlayingStyle        = primaryColorStyle
	nowPlayingDescStyle    = secondaryColorStyle
	selNowPlayingStyle     = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedPrimaryColor)
	selNowPlayingDescStyle = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedSecondColor)

	playStatusStyle = lipgloss.NewStyle().Bold(true).Foreground(baseSecondColor)

	itemStyle    = primaryColorStyle
	descStyle    = secondaryColorStyle
	selItemStyle = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedPrimaryColor)
	selDescStyle = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedSecondColor)

	selectedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.BlockBorder(), false, false, false, true).
				BorderForeground(basePrimaryColor)

	viewStyle    = secondaryColorStyle.PaddingLeft(padDist)
	noItemsStyle = secondaryColorStyle.PaddingLeft(3)

	// header
	playTimeStyle = lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder(), false, true).
			Foreground(baseSecondColor).
			Italic(true).
			Padding(0, 0).Margin()
	volumeStyle = lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder(), false, true).
			Foreground(basePrimaryColor).
			Italic(true).
			Padding(0, 0).Margin()

	// tabs
	inactiveTab = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.HiddenBorder(), true).
			Foreground(basePrimaryColor).
			Padding(0, 0).Margin(0)
	activeTab = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(basePrimaryColor).
			Foreground(basePrimaryColor).
			Padding(0, 0).Margin(0)
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

	// general
	backgroundColor = termenv.RGBColor("#282c34")
	docStyle        = lipgloss.NewStyle().Padding(1, padDist, 0, padDist)
	statusBarStyle  = lipgloss.NewStyle().Background(baseSecondColor).Foreground(invertedPrimaryColor)
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
