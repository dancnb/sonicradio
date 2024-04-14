package ui

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
)

type favoritesTab struct {
	list       list.Model
	viewMsg    string
	listKeymap listKeymap
}

func newFavoritesTab() *favoritesTab {
	k := newListKeymap()
	k.search.SetEnabled(false)

	m := &favoritesTab{
		listKeymap: k,
	}
	return m
}

func (t *favoritesTab) createList(delegate *stationDelegate, width int, height int) list.Model {
	l := createList(delegate, width, height)

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{t.listKeymap.toNowPlaying, t.listKeymap.toBrowser, t.listKeymap.prevTab, t.listKeymap.nextTab}
	}

	return l
}

func (t *favoritesTab) Init(m *model) tea.Cmd {
	t.viewMsg = loadingMsg
	t.list = t.createList(m.delegate, m.width, m.totHeight-m.headerHeight)
	return m.favoritesReqCmd
}

func (t *favoritesTab) Update(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	slog.Debug("favorites tab", "type", fmt.Sprintf("%T", msg), "value", msg, "#", fmt.Sprintf("%#v", msg))

	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		v, h := docStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-m.headerHeight-v)

	case favoritesStationRespMsg:
		t.viewMsg = string(msg.viewMsg)
		items := make([]list.Item, 0)
		var notFound []string
		for j := 0; j < len(m.cfg.Favorites); j++ {
			found := false
			for i := 0; i < len(msg.stations); i++ {
				if msg.stations[i].Stationuuid == m.cfg.Favorites[j] {
					items = append(items, msg.stations[i])
					found = true
					break
				}
			}
			if !found {
				notFound = append(notFound, m.cfg.Favorites[j])
			}
		}
		if len(notFound) > 0 {
			// TODO status message: some stations could not be retrieved from the server
		}
		cmd := t.list.SetItems(items)
		cmds = append(cmds, cmd)

	case toggleFavoriteMsg:
		if msg.added {
			cmd := t.list.InsertItem(len(t.list.Items()), msg.station)
			cmds = append(cmds, cmd)
		} else {
			its := t.list.Items()
			for i := range its {
				s := its[i].(browser.Station)
				if s.Stationuuid == msg.station.Stationuuid {
					// _, _ = m.delegate.stopStation(s)
					t.list.RemoveItem(i)
					break
				}
			}
		}
		t.viewMsg = ""
		if len(t.list.Items()) == 0 {
			t.viewMsg = noFavoritesAddedMsg
		}

	case tea.KeyMsg:
		if key.Matches(msg, t.listKeymap.toNowPlaying) {
			newListModel, cmd := t.list.Update(msg)
			t.list = newListModel
			cmds = append(cmds, cmd)
			toNowPlaying(m, t)
		}

		// Don't match any of the keys below if we're actively filtering.
		if t.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, t.list.KeyMap.Quit, t.list.KeyMap.ForceQuit):
			m.stop()

		case key.Matches(msg, t.listKeymap.toBrowser):
			m.activeTab = browseTabIx
		case key.Matches(msg, t.listKeymap.nextTab):
			m.activeTab = browseTabIx
		case key.Matches(msg, t.listKeymap.prevTab):
			m.activeTab = browseTabIx
		}
	}

	newListModel, cmd := t.list.Update(msg)
	t.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (t *favoritesTab) View() string {
	if t.viewMsg != "" {
		return itemStyle.Render(t.viewMsg)
	}

	return t.list.View()
}

func (t *favoritesTab) List() *list.Model { return &t.list }
