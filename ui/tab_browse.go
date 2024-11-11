package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
)

type browseTab struct {
	stationsTabBase
	defTopStations []browser.Station
	searchModel    *searchModel
}

func newBrowseTab(browser *browser.Api, infoModel *infoModel) *browseTab {
	k := newListKeymap()

	m := &browseTab{
		stationsTabBase: newStationsTab(k, infoModel),
		searchModel:     newSearchModel(browser),
	}
	return m
}

func (t *browseTab) createList(delegate *stationDelegate, width int, height int) list.Model {
	l := createList(delegate, width, height)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{t.listKeymap.search}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			t.listKeymap.search,
			t.listKeymap.digitHelp,
			t.listKeymap.toNowPlaying,
			t.listKeymap.prevTab,
			t.listKeymap.nextTab,
			delegate.keymap.volumeDown,
			delegate.keymap.volumeUp,
		}
	}

	return l
}

func (t *browseTab) Init(m *Model) tea.Cmd {
	t.viewMsg = loadingMsg
	t.list = t.createList(m.delegate, m.width, m.totHeight-m.headerHeight)
	return m.topStationsCmd
}

func (t *browseTab) Update(m *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "ui.browseTab.Update")

	var cmds []tea.Cmd

	if t.IsSearchEnabled() {
		searchModelMsg := msg
		if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
			searchModelMsg = t.newSizeMsg(sizeMsg, m)
		}
		sm, cmd := t.searchModel.Update(searchModelMsg)
		t.searchModel = sm.(*searchModel)
		cmds = append(cmds, cmd)
	} else if t.IsInfoEnabled() {
		infoModelMsg := msg
		if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
			infoModelMsg = t.newSizeMsg(sizeMsg, m)
		}
		im, cmd := t.infoModel.Update(infoModelMsg)
		t.infoModel = im
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-m.headerHeight-v)

	case topStationsRespMsg:
		m.updateStatus(string(msg.statusMsg))
		t.viewMsg = string(msg.viewMsg)
		copy(t.defTopStations, msg.stations)
		cmd := t.setStations(msg.stations)
		cmds = append(cmds, cmd)

	case searchRespMsg:
		t.listKeymap.setEnabled(true)
		if msg.cancelled {
			// do nothing, list already has top stations
		} else {
			m.updateStatus(string(msg.statusMsg))
			t.viewMsg = string(msg.viewMsg)
			cmd := t.setStations(msg.stations)
			cmds = append(cmds, cmd)
		}

	case toggleInfoMsg:
		if msg.enable {
			cmds = append(cmds, t.initInfoModel(m, msg))
			return m, tea.Batch(cmds...)
		} else {
			t.listKeymap.setEnabled(true)
		}

	case tea.KeyMsg:
		if t.IsSearchEnabled() || t.IsInfoEnabled() {
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
			return m, tea.Quit

		case key.Matches(msg, t.listKeymap.search):
			t.listKeymap.setEnabled(false)
			t.searchModel.setSize(m.width, m.totHeight-m.headerHeight)
			cmds = append(cmds, t.searchModel.Init())
			return m, tea.Batch(cmds...)

		case key.Matches(msg, t.listKeymap.nextTab):
			m.toFavoritesTab()

		case key.Matches(msg, t.listKeymap.prevTab):
			m.toFavoritesTab()

		case key.Matches(msg, t.listKeymap.digits...):
			t.doJump(msg)
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
	t.list.Select(0)
	return cmd
}

func (t *browseTab) View() string {
	if t.IsSearchEnabled() {
		return t.searchModel.View()
	} else if t.IsInfoEnabled() {
		return t.infoModel.View()
	}
	return t.stationsTabBase.View()
}

func (t *browseTab) IsSearchEnabled() bool {
	return t.searchModel.isEnabled()
}
