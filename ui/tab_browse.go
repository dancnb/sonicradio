package ui

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
)

type browseTab struct {
	list       list.Model
	viewMsg    string
	listKeymap listKeymap
}

func newBrowseTab() *browseTab {
	k := newListKeymap()

	m := &browseTab{
		listKeymap: k,
	}
	return m
}

func (t *browseTab) createList(delegate *stationDelegate, width int, height int) list.Model {
	l := createList(delegate, width, height)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{t.listKeymap.search}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{t.listKeymap.search, t.listKeymap.toNowPlaying, t.listKeymap.toFavorites}
	}

	return l
}

func (t *browseTab) Init(m *model) tea.Cmd {
	t.viewMsg = loadingMsg
	t.list = t.createList(m.delegate, m.width, m.totHeight-m.headerHeight)
	return m.topStationsCmd
}

func (t *browseTab) Update(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	slog.Debug("browse tab", "type", fmt.Sprintf("%T", msg), "value", msg)

	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		v, h := docStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-m.headerHeight-v)

	case topStationsRespMsg:
		t.viewMsg = msg.viewMsg
		items := make([]list.Item, len(msg.stations))
		for i := 0; i < len(msg.stations); i++ {
			items[i] = msg.stations[i]
		}
		cmd := t.list.SetItems(items)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if key.Matches(msg, t.listKeymap.toNowPlaying) {
			newListModel, cmd := t.list.Update(msg)
			t.list = newListModel
			cmds = append(cmds, cmd)
			if m.delegate.nowPlaying != nil {
				selIndex := 0
				items := t.list.Items()
				for ix := range items {
					if items[ix].(browser.Station).Stationuuid == m.delegate.nowPlaying.Stationuuid {
						selIndex = ix
						break
					}
				}
				t.list.Select(selIndex)
			}
		}

		// Don't match any of the keys below if we're actively filtering.
		if t.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, t.list.KeyMap.Quit, t.list.KeyMap.ForceQuit):
			m.stop()

		case key.Matches(msg, t.listKeymap.search):
			// TODO search stations; use cmd and msg
			cmd := t.list.NewStatusMessage(statusWarnMessageStyle("Not implemented yet!"))
			cmds = append(cmds, cmd)

		case key.Matches(msg, t.listKeymap.toFavorites):
			// favs := m.tabs[favoriteTabIx].Items()
			//    toAdd:=make([]string, 0)
			// for _, fav := range m.cfg.Favorites {
			//      if !slices.ContainsFunc(favs, func (it list.Item) bool  {
			//        s:=it.(browser.Station)
			//        return s.Stationuuid == fav
			//      }){
			//        toAdd = append(toAdd, s.S)
			//      }
			// }
			m.activeTab = favoriteTabIx
		}
	}

	newListModel, cmd := t.list.Update(msg)
	t.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (t *browseTab) View() string {
	if t.viewMsg != "" {
		return itemStyle.Render(t.viewMsg)
	}
	return t.list.View()
}
func (t *browseTab) Items() []list.Item { return t.list.Items() }
