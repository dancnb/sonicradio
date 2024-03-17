package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	tabGapDistance = 5
)

var (
	// list items
	baseColor     = lipgloss.Color("#ffb641")
	selectedColor = lipgloss.Color("#12100d")
	selDescColor  = lipgloss.Color("#4a4133")

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

	nowPlayingStyle        = lipgloss.NewStyle().Background(baseColor).Foreground(selectedColor)
	nowPlayingDescStyle    = lipgloss.NewStyle().Background(baseColor).Foreground(selDescColor)
	selNowPlayingStyle     = lipgloss.NewStyle().Background(baseColor).Foreground(selectedColor)
	selNowPlayingDescStyle = lipgloss.NewStyle().Background(baseColor).Foreground(selDescColor)

	itemStyle    = lipgloss.NewStyle().Foreground(baseColor).PaddingLeft(4)
	descStyle    = itemStyle.Copy().Faint(true)
	selItemStyle = lipgloss.NewStyle().Foreground(baseColor).PaddingLeft(3)
	selDescStyle = selItemStyle.Copy().Faint(true)

	selectedBorderStyle = lipgloss.NewStyle().Border(lipgloss.BlockBorder(), false, false, false, true).BorderForeground(baseColor)

	// tabs
	inactiveTab = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.HiddenBorder(), true).
			Foreground(baseColor).
			Padding(0, 1).Margin(0)
	activeTab = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.NormalBorder(), true).
			BorderForeground(baseColor).
			Foreground(baseColor).
			Padding(0, 1).Margin(0)
	tabGap = lipgloss.NewStyle().
		Border(lipgloss.Border{Left: " ", Right: " "}, true, false).
		Foreground(baseColor).
		BorderForeground(baseColor).
		Strikethrough(true).
		Margin(0).Padding(0)

	// help
	helpkeyStyle  = lipgloss.NewStyle().Foreground(baseColor)
	helpDescStyle = lipgloss.NewStyle().Foreground(baseColor).Faint(true)
	helpStyle     = lipgloss.NewStyle().
			Padding(0, 1).Margin(0).
			Border(lipgloss.NormalBorder()).
			BorderForeground(baseColor)

	// filter
	filterPromptStyle = lipgloss.NewStyle().Foreground(baseColor).Bold(true)
	filterTextStyle   = filterPromptStyle.Copy().Faint(true)

	// general
	backgroundColor = termenv.RGBColor("#282c34")
	docStyle        = lipgloss.NewStyle().Padding(1, 2)
)
