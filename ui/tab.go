package ui

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
)

type uiTabIndex uint8

func (t uiTabIndex) String() string {
	switch t {
	case favoriteTabIx:
		return "Favorites"
	case browseTabIx:
		return " Browse "
	}
	return ""
}

const (
	favoriteTabIx uiTabIndex = iota
	browseTabIx
	// historyTabIx
	// configTab
)

type uiTab interface {
	Init(m *model) tea.Cmd
	Update(m *model, msg tea.Msg) (tea.Model, tea.Cmd)
	View() string

	Stations() *stationsTab
	IsFiltering() bool
	IsSearchEnabled() bool
	IsInfoEnabled() bool
}

type stationsTab struct {
	uiTab
	list       list.Model
	viewMsg    string
	listKeymap listKeymap
	infoModel  *infoModel
}

func (t *stationsTab) View() string {
	if t.viewMsg != "" {
		var sections []string
		availHeight := t.list.Height()
		help := t.list.Styles.HelpStyle.Render(t.list.Help.View(t.list))
		availHeight -= lipgloss.Height(help)
		viewSection := viewStyle.Height(availHeight).Render(t.viewMsg)
		sections = append(sections, viewSection)
		sections = append(sections, help)
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}
	return t.list.View()
}

func (t *stationsTab) Stations() *stationsTab { return t }

func (t *stationsTab) IsFiltering() bool {
	return t.list.FilterState() == list.Filtering
}

func (t *stationsTab) toNowPlaying(m *model) {
	uuid := ""
	if m.delegate.currPlaying != nil {
		uuid = m.delegate.currPlaying.Stationuuid
	} else if m.delegate.prevPlaying != nil {
		uuid = m.delegate.prevPlaying.Stationuuid
	} else {
		return
	}
	selIndex := -1
	items := t.list.VisibleItems()
	for ix := range items {
		if items[ix].(browser.Station).Stationuuid == uuid {
			selIndex = ix
			break
		}
	}
	slog.Debug("method", "ui.baseTab.toNowPlaying", "selIndex", selIndex)
	if selIndex > -1 {
		t.list.Select(selIndex)
	}
}

func (t *stationsTab) IsSearchEnabled() bool {
	return false
}

func (t *stationsTab) IsInfoEnabled() bool {
	return t.infoModel != nil && t.infoModel.enabled
}

func (*stationsTab) newSizeMsg(sizeMsg tea.WindowSizeMsg, m *model) tea.WindowSizeMsg {
	availableHeight := sizeMsg.Height - m.headerHeight
	newSizeMsg := tea.WindowSizeMsg{Width: sizeMsg.Width, Height: availableHeight}
	return newSizeMsg
}

func (t *stationsTab) initInfoModel(m *model, msg toggleInfoMsg) tea.Cmd {
	t.listKeymap.setEnabled(false)
	t.infoModel.setSize(m.width, m.totHeight-m.headerHeight)
	return t.infoModel.Init(msg.station)
}

func createList(delegate *stationDelegate, width int, height int) list.Model {
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.InfiniteScrolling = true
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowFilter(true)
	l.SetStatusBarItemName("station", "stations")
	l.Styles.NoItems = noItemsStyle
	l.FilterInput.ShowSuggestions = true
	l.KeyMap.Quit.SetKeys("q")
	l.KeyMap.PrevPage.SetKeys("pgup", "u")
	l.KeyMap.PrevPage.SetHelp("u/pgup", "prev page")
	l.KeyMap.NextPage.SetKeys("pgdown", "d")
	l.KeyMap.NextPage.SetHelp("d/pgdn", "next page")
	h, v := docStyle.GetFrameSize()
	l.SetSize(width-h, height-v)

	l.Help.ShortSeparator = "   "
	l.Help.Styles = helpStyles()
	l.Styles.HelpStyle = helpStyle

	textInputSyle(&l.FilterInput, "Filter:       ", "name")
	return l
}
