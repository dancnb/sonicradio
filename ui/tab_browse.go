package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
)

type browseTab struct {
	baseTab
	defTopStations []browser.Station
	searchModel    searchModel
}

func newBrowseTab() *browseTab {
	k := newListKeymap()

	m := &browseTab{
		baseTab: baseTab{
			listKeymap: k,
		},
		searchModel: searchModel{
			content: "placeholder",
			keymap:  newSearchKeymap(),
		},
	}
	return m
}

func (t *browseTab) createList(delegate *stationDelegate, width int, height int) list.Model {
	l := createList(delegate, width, height)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{t.listKeymap.search}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{t.listKeymap.search, t.listKeymap.toNowPlaying, t.listKeymap.toFavorites, t.listKeymap.prevTab, t.listKeymap.nextTab}
	}

	return l
}

func (t *browseTab) Init(m *model) tea.Cmd {
	t.viewMsg = loadingMsg
	t.list = t.createList(m.delegate, m.width, m.totHeight-m.headerHeight)
	return m.topStationsCmd
}

func (t *browseTab) Update(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "update browseTab")

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-m.headerHeight-v)

	case topStationsRespMsg:
		t.viewMsg = string(msg.viewMsg)
		copy(t.defTopStations, msg.stations)
		cmd := t.setStations(msg.stations)
		cmds = append(cmds, cmd)

	case searchRespMsg:
		t.listKeymap.setEnabled(true)
		if msg.cancelled {
			// do nothing, list already has top stations
		} else {
			t.viewMsg = string(msg.viewMsg)
			// TODO handle errMsg
			cmd := t.setStations(msg.stations)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		// search > filter > list
		if t.IsSearch() {
			cmd := t.searchModel.update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		if key.Matches(msg, t.listKeymap.toNowPlaying) {
			newListModel, cmd := t.list.Update(msg)
			t.list = newListModel
			cmds = append(cmds, cmd)
			t.toNowPlaying(m)
		}

		// Don't match any of the keys below if we're actively filtering.
		if t.IsFiltering() {
			break
		}

		switch {
		case key.Matches(msg, t.list.KeyMap.Quit, t.list.KeyMap.ForceQuit):
			m.quit()
			return m, tea.Quit

		case key.Matches(msg, t.listKeymap.search):
			t.listKeymap.setEnabled(false)
			t.searchModel.setEnabled(true)
			return m, tea.Batch(cmds...)

		case key.Matches(msg, t.listKeymap.toFavorites):
			m.activeTab = favoriteTabIx
		case key.Matches(msg, t.listKeymap.nextTab):
			m.activeTab = favoriteTabIx
		case key.Matches(msg, t.listKeymap.prevTab):
			m.activeTab = favoriteTabIx
		}
	}

	newListModel, cmd := t.list.Update(msg)
	t.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (t *browseTab) setStations(stations []browser.Station) tea.Cmd {
	items := make([]list.Item, len(stations))
	for i := 0; i < len(stations); i++ {
		items[i] = stations[i]
	}
	cmd := t.list.SetItems(items)
	return cmd
}

func (t *browseTab) View() string {
	if t.IsSearch() {
		return t.searchModel.view()
	}
	return t.baseTab.View()
}

func (t *browseTab) IsSearch() bool {
	return t.searchModel.enabled
}
