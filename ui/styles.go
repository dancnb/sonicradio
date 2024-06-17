package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	tabGapDistance = 2
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

	prefixStyle           = primaryColorStyle.Copy().PaddingLeft(1)
	nowPlayingPrefixStyle = primaryColorStyle.Copy().PaddingLeft(0)

	nowPlayingStyle        = primaryColorStyle.Copy()
	nowPlayingDescStyle    = secondaryColorStyle.Copy()
	selNowPlayingStyle     = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedPrimaryColor)
	selNowPlayingDescStyle = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedSecondColor)

	playStatusStyle = lipgloss.NewStyle().Bold(true).Foreground(baseSecondColor)

	itemStyle    = primaryColorStyle.Copy()
	descStyle    = secondaryColorStyle.Copy()
	selItemStyle = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedPrimaryColor)
	selDescStyle = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedSecondColor)

	selectedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.BlockBorder(), false, false, false, true).
				BorderForeground(basePrimaryColor)

	viewStyle    = secondaryColorStyle.Copy().PaddingLeft(2)
	noItemsStyle = secondaryColorStyle.Copy().PaddingLeft(3)

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
	helpkeyStyle  = primaryColorStyle.Copy()
	helpDescStyle = secondaryColorStyle.Copy()
	helpStyle     = lipgloss.NewStyle().
			Padding(0, 0).Margin(0).
			Border(lipgloss.NormalBorder()).
			BorderForeground(basePrimaryColor)

	// filter
	filterPromptStyle = primaryColorStyle.Copy().Bold(true).MarginLeft(1)
	filterTextStyle   = primaryColorStyle.Copy()

	//search
	searchPromptStyle = primaryColorStyle.Copy().Bold(true).MarginLeft(3)

	// general
	backgroundColor = termenv.RGBColor("#282c34")
	docStyle        = lipgloss.NewStyle().Padding(1, 2, 0, 2)
)

func textInputSyle(textInput *textinput.Model, prompt, placeholder string) {
	textInput.Prompt = prompt
	textInput.PromptStyle = filterPromptStyle
	textInput.TextStyle = filterTextStyle
	textInput.Cursor.Style = filterPromptStyle
	textInput.Placeholder = placeholder
	textInput.PlaceholderStyle = secondaryColorStyle.Copy()
}

func helpStyles() help.Styles {
	return help.Styles{
		ShortKey:       helpkeyStyle,
		ShortDesc:      helpDescStyle,
		ShortSeparator: helpDescStyle,
		Ellipsis:       helpDescStyle.Copy(),
		FullKey:        helpkeyStyle.Copy(),
		FullDesc:       helpDescStyle.Copy(),
		FullSeparator:  helpDescStyle.Copy(),
	}
}
