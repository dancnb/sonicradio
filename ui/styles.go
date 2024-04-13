package ui

import (
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
	// playChar  = "\u25B6"
	// pauseChar = "\u2225"

	// list items
	basePrimaryColor     = lipgloss.Color("#ffb641")
	baseSecondColor      = lipgloss.Color("#bd862d")
	invertedPrimaryColor = lipgloss.Color("#12100d")
	invertedSecondColor  = lipgloss.Color("#4a4133")

	prefixStyle           = lipgloss.NewStyle().Foreground(basePrimaryColor).PaddingLeft(1)
	nowPlayingPrefixStyle = lipgloss.NewStyle().Foreground(basePrimaryColor).PaddingLeft(0)

	nowPlayingStyle        = lipgloss.NewStyle().Foreground(basePrimaryColor)
	nowPlayingDescStyle    = lipgloss.NewStyle().Foreground(baseSecondColor)
	selNowPlayingStyle     = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedPrimaryColor)
	selNowPlayingDescStyle = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedSecondColor)

	playStatusStyle = lipgloss.NewStyle().Bold(true).Foreground(baseSecondColor)

	itemStyle    = lipgloss.NewStyle().Foreground(basePrimaryColor)
	descStyle    = lipgloss.NewStyle().Foreground(baseSecondColor)
	selItemStyle = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedPrimaryColor)
	selDescStyle = lipgloss.NewStyle().Background(basePrimaryColor).Foreground(invertedSecondColor)

	selectedBorderStyle = lipgloss.NewStyle().Border(lipgloss.BlockBorder(), false, false, false, true).BorderForeground(basePrimaryColor)

	// tabs
	inactiveTab = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.HiddenBorder(), true).
			Foreground(basePrimaryColor).
			Padding(0, 1).Margin(0)
	activeTab = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(basePrimaryColor).
			Foreground(basePrimaryColor).
			Padding(0, 1).Margin(0)
	tabGap = lipgloss.NewStyle().
		Border(lipgloss.Border{Left: " ", Right: " "}, true, false).
		Foreground(basePrimaryColor).
		BorderForeground(basePrimaryColor).
		Strikethrough(true).
		Margin(0).Padding(0)

	// header
	spinnerStyle = lipgloss.NewStyle().Foreground(basePrimaryColor)

	// help
	helpkeyStyle  = lipgloss.NewStyle().Foreground(basePrimaryColor)
	helpDescStyle = lipgloss.NewStyle().Foreground(baseSecondColor)
	helpStyle     = lipgloss.NewStyle().
			Padding(0, 1).Margin(0).
			Border(lipgloss.NormalBorder()).
			BorderForeground(basePrimaryColor)

	// filter
	filterPromptStyle = lipgloss.NewStyle().Foreground(basePrimaryColor).Bold(true)
	filterTextStyle   = lipgloss.NewStyle().Foreground(baseSecondColor).Bold(true)

	// general
	backgroundColor = termenv.RGBColor("#282c34")
	docStyle        = lipgloss.NewStyle().Padding(1, 2)

	// TODO replace list status
	statusWarnMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#eab676", Dark: "#eab676"}).
				Render
	statusErrMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#fc2c03", Dark: "#fc2c03"}).
				Render
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)
