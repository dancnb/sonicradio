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
		// case historyTabIx:
		// 	return " History "
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
	List() *list.Model
	IsFiltering() bool
	IsSearchEnabled() bool
}

type baseTab struct {
	uiTab
	list       list.Model
	viewMsg    string
	listKeymap listKeymap
}

func (t *baseTab) View() string {
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

func (t *baseTab) List() *list.Model { return &t.list }

func (t *baseTab) IsFiltering() bool {
	return t.list.FilterState() == list.Filtering
}

func (t *baseTab) toNowPlaying(m *model) {
	uuid := ""
	if m.delegate.currPlaying != nil {
		uuid = m.delegate.currPlaying.Stationuuid
	} else if m.delegate.prevPlaying != nil {
		uuid = m.delegate.prevPlaying.Stationuuid
	} else {
		return
	}
	selIndex := -1
	items := t.List().VisibleItems()
	for ix := range items {
		if items[ix].(browser.Station).Stationuuid == uuid {
			selIndex = ix
			break
		}
	}
	slog.Debug("method", "ui.baseTab.toNowPlaying", "selIndex", selIndex)
	if selIndex > -1 {
		t.List().Select(selIndex)
	}
}

func (t *baseTab) IsSearchEnabled() bool {
	return false
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
