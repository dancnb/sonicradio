package ui

import (
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
)

type uiTabIndex uint8

func (t uiTabIndex) String() string {
	switch t {
	case favoriteTabIx:
		return "1. Favorites"
	case browseTabIx:
		return "2. Browse"
	case historyTabIx:
		return "3. History"
	}
	return ""
}

const (
	favoriteTabIx uiTabIndex = iota
	browseTabIx
	historyTabIx
	// configTab
)

type uiTab interface {
	Init(m *model) tea.Cmd
	Update(m *model, msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
	List() *list.Model
}

func toNowPlaying(m *model, t uiTab) {
	if m.delegate.currPlaying != nil {
		selIndex := -1
		items := t.List().VisibleItems()
		for ix := range items {
			if items[ix].(browser.Station).Stationuuid == m.delegate.currPlaying.Stationuuid {
				selIndex = ix
				break
			}
		}
		slog.Debug("toNowPlaying", "selIndex", selIndex)
		if selIndex > -1 {
			t.List().Select(selIndex)
		}
	}
}

func createList(delegate *stationDelegate, width int, height int) list.Model {
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.InfiniteScrolling = true
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.StatusMessageLifetime = 3 * time.Second
	l.SetShowPagination(false)
	l.SetShowFilter(true)
	l.FilterInput.ShowSuggestions = true
	l.KeyMap.Quit.SetKeys("q")
	l.KeyMap.PrevPage.SetKeys("left", "h", "pgup", "u")
	l.KeyMap.NextPage.SetKeys("right", "l", "pgdown", "d")
	v, h := docStyle.GetFrameSize()
	l.SetSize(width-h, height-v)

	l.Help.ShortSeparator = "  "
	l.Help.Styles = help.Styles{
		ShortKey:       helpkeyStyle,
		ShortDesc:      helpDescStyle,
		ShortSeparator: helpDescStyle,
		Ellipsis:       helpDescStyle.Copy(),
		FullKey:        helpkeyStyle.Copy(),
		FullDesc:       helpDescStyle.Copy(),
		FullSeparator:  helpDescStyle.Copy(),
	}
	l.Styles.HelpStyle = helpStyle

	l.FilterInput.PromptStyle = filterPromptStyle
	l.FilterInput.TextStyle = filterTextStyle
	l.FilterInput.Cursor.Style = filterPromptStyle
	return l
}
