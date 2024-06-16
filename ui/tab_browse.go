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
	searchModel    *searchModel
}

func newBrowseTab(browser *browser.Api) *browseTab {
	k := newListKeymap()

	m := &browseTab{
		baseTab: baseTab{
			listKeymap: k,
		},
		searchModel: newSearchModel(browser),
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

	if t.searchModel.isEnabled() {
		if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
			availableHeight := sizeMsg.Height - m.headerHeight
			newSizeMsg := tea.WindowSizeMsg{Width: sizeMsg.Width, Height: availableHeight}
			sm, cmd := t.searchModel.Update(newSizeMsg)
			t.searchModel = sm.(*searchModel)
			cmds = append(cmds, cmd)
		} else {
			sm, cmd := t.searchModel.Update(msg)
			t.searchModel = sm.(*searchModel)
			cmds = append(cmds, cmd)
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-m.headerHeight-v)

	case topStationsRespMsg:
		m.updateStatus(msg.statusMsg)
		t.viewMsg = string(msg.viewMsg)
		copy(t.defTopStations, msg.stations)
		cmd := t.setStations(msg.stations)
		cmds = append(cmds, cmd)

	case searchRespMsg:
		t.listKeymap.setEnabled(true)
		if msg.cancelled {
			// do nothing, list already has top stations
		} else {
			m.updateStatus(msg.statusMsg)
			t.viewMsg = string(msg.viewMsg)
			cmd := t.setStations(msg.stations)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		if t.searchModel.isEnabled() {
			return m, tea.Batch(cmds...)
		}

		if key.Matches(msg, t.listKeymap.toNowPlaying) {
			newListModel, cmd := t.list.Update(msg)
			t.list = newListModel
			cmds = append(cmds, cmd)
			t.toNowPlaying(m)
		}

		if t.IsFiltering() {
			break
		}

		switch {
		case key.Matches(msg, t.list.KeyMap.Quit, t.list.KeyMap.ForceQuit):
			m.quit()
			return m, tea.Quit

		case key.Matches(msg, t.listKeymap.search):
			t.listKeymap.setEnabled(false)
			t.searchModel.setSize(m.width, m.totHeight-m.headerHeight)
			cmds = append(cmds, t.searchModel.Init())
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
	t.List().Select(0)
	return cmd
}

func (t *browseTab) View() string {
	if t.searchModel.isEnabled() {
		return t.searchModel.View()
	}
	return t.baseTab.View()
}

func (t *browseTab) IsSearchEnabled() bool {
	return t.searchModel.isEnabled()
}
