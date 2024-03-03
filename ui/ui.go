package ui

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
)

func NewProgram(cfg config.Value, b *browser.Api) *tea.Program {
	m := initialModel(cfg, b)
	return tea.NewProgram(m, tea.WithAltScreen())
}

func initialModel(cfg config.Value, b *browser.Api) model {
	m := model{
		browser: b,
	}

	stations := m.browser.TopStations()
	items := make([]list.Item, len(stations))
	for i := 0; i < len(stations); i++ {
		items[i] = stations[i]
	}

	delegate := newItemDelegate()
	uiList := list.New(items, delegate, 0, 0)
	uiList.SetShowStatusBar(true)
	uiList.Title = "Stations"
	uiList.Styles.Title = titleStyle
	uiList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			//
		}
	}
	m.list = uiList
	return m
}

type model struct {
	cfg     config.Value
	list    list.Model
	browser *browser.Api
	// delegateKeys  *delegateKeyMap
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		if msg.String() == "q" {
			slog.Info("----Quitting----")
			return m, tea.Quit
		}
		switch {
		// case key.Matches(msg, m.keys.toggleSpinner):
		// 	cmd := m.list.ToggleSpinner()
		// 	return m, cmd

		// case key.Matches(msg, m.keys.toggleTitleBar):
		// 	v := !m.list.ShowTitle()
		// 	m.list.SetShowTitle(v)
		// 	m.list.SetShowFilter(v)
		// 	m.list.SetFilteringEnabled(v)
		// 	return m, nil

		// case key.Matches(msg, m.keys.toggleStatusBar):
		// 	m.list.SetShowStatusBar(!m.list.ShowStatusBar())
		// 	return m, nil

		// case key.Matches(msg, m.keys.togglePagination):
		// 	m.list.SetShowPagination(!m.list.ShowPagination())
		// 	return m, nil

		// case key.Matches(msg, m.keys.toggleHelpMenu):
		// 	m.list.SetShowHelp(!m.list.ShowHelp())
		// 	return m, nil
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(m.list.View())
}
