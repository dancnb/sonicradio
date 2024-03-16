package ui

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
)

type favoritesTab struct {
	list    list.Model
	viewMsg string
	keymap  keymap
}

func newFavoritesTab() *favoritesTab {
	k := newKeymap()
	k.search.SetEnabled(false)

	m := &favoritesTab{
		keymap: k,
	}
	return m
}

func (t *favoritesTab) createList(delegate *stationDelegate, width int, height int) list.Model {
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.InfiniteScrolling = true
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.StatusMessageLifetime = 3 * time.Second
	l.SetShowPagination(false)
	l.SetShowFilter(true)
	l.FilterInput.ShowSuggestions = true

	l.KeyMap.Quit.SetKeys("q")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{t.keymap.toNowPlaying}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{t.keymap.toNowPlaying, t.keymap.toBrowser}
	}
	v, h := docStyle.GetFrameSize()
	l.SetSize(width-h, height-v)

	return l
}

func (t *favoritesTab) Init(m *model) tea.Cmd {
	t.viewMsg = loadingMsg
	t.list = t.createList(m.delegate, m.width, m.totHeight-m.headerHeight)
	return m.favoritesReqCmd
}

func (t *favoritesTab) Update(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	slog.Debug("favorites tab", "type", fmt.Sprintf("%T", msg), "value", msg)

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v, h := docStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-m.headerHeight-v)

	case favoritesStationRespMsg:
		t.viewMsg = msg.viewMsg
		items := make([]list.Item, len(msg.stations))
		for i := 0; i < len(msg.stations); i++ {
			items[i] = msg.stations[i]
		}
		t.list.SetItems(items)

	case quitMsg:
		m.stop()
		return nil, tea.Quit

	case tea.KeyMsg:
		if key.Matches(msg, t.keymap.toNowPlaying) {
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

		case key.Matches(msg, t.keymap.toBrowser):
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
