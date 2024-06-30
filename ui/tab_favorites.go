package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
)

type favoritesTab struct {
	stationsTab
}

func newFavoritesTab(infoModel *infoModel) *favoritesTab {
	k := newListKeymap()

	m := &favoritesTab{
		stationsTab: stationsTab{
			listKeymap: k,
			infoModel:  infoModel,
		},
	}
	return m
}

func (t *favoritesTab) createList(delegate *stationDelegate, width int, height int) list.Model {
	l := createList(delegate, width, height)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{t.listKeymap.search}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{t.listKeymap.search, t.listKeymap.toNowPlaying, t.listKeymap.toBrowser, t.listKeymap.prevTab, t.listKeymap.nextTab}
	}

	return l
}

func (t *favoritesTab) Init(m *model) tea.Cmd {
	t.viewMsg = loadingMsg
	t.list = t.createList(m.delegate, m.width, m.totHeight-m.headerHeight)
	return m.favoritesReqCmd
}

func (t *favoritesTab) Update(m *model, msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "ui.favoritesTab.Update")

	var cmds []tea.Cmd

	if t.IsInfoEnabled() {
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
		sm := msg.statusMsg
		if sm == "" && len(notFound) > 0 {
			sm = statusMsg(missingFavorites)
		}
		m.updateStatus(sm)
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
					t.list.RemoveItem(i)
					break
				}
			}
		}
		t.viewMsg = ""
		if len(t.list.Items()) == 0 {
			t.viewMsg = noFavoritesAddedMsg
		}

	case toggleInfoMsg:
		if msg.enable {
			cmds = append(cmds, t.initInfoModel(m, msg))
			return m, tea.Batch(cmds...)
		} else {
			t.listKeymap.setEnabled(true)
		}

	case tea.KeyMsg:
		if t.IsInfoEnabled() {
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
			m.activeTab = browseTabIx
			return m.tabs[browseTabIx].Update(m, msg)
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
	if t.IsInfoEnabled() {
		return t.infoModel.View()
	}
	return t.stationsTab.View()
}
