package ui

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type browseTab struct {
	baseTab
}

func newBrowseTab() *browseTab {
	k := newListKeymap()

	m := &browseTab{
		baseTab: baseTab{
			listKeymap: k,
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
	slog.Debug("browse tab", "type", fmt.Sprintf("%T", msg), "value", msg, "#", fmt.Sprintf("%#v", msg))

	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		v, h := docStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-m.headerHeight-v)

	case topStationsRespMsg:
		t.viewMsg = string(msg.viewMsg)
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
			t.toNowPlaying(m)
		}

		// Don't match any of the keys below if we're actively filtering.
		if t.IsFiltering() {
			break
		}

		switch {
		case key.Matches(msg, t.list.KeyMap.Quit, t.list.KeyMap.ForceQuit):
			m.quit()

		case key.Matches(msg, t.listKeymap.search):
			// TODO search stations; use cmd and msg
			// cmd := t.list.NewStatusMessage(statusWarnMessageStyle("Not implemented yet!"))
			// cmds = append(cmds, cmd)

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
