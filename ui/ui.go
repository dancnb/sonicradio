package ui

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player"
)

func NewProgram(cfg config.Value, b *browser.Api, p player.Player) *tea.Program {
	m := initialModel(cfg, b, p)
	return tea.NewProgram(m, tea.WithAltScreen())
}

func initialModel(cfg config.Value, b *browser.Api, p player.Player) model {
	m := model{
		cfg:     cfg,
		browser: b,
		player:  p,
	}

	stations := m.browser.TopStations()
	items := make([]list.Item, len(stations))
	for i := 0; i < len(stations); i++ {
		items[i] = stations[i]
	}

	x := 0
	y := 0
	l := list.New(items, newStationDelegate(newDelegateKeyMap(), p), x, y)
	l.InfiniteScrolling = true
	// l.Paginator.PerPage = 50
	// l.Paginator.SetTotalPages(len(items))
	l.SetShowStatusBar(true)
	l.Title = "Stations"
	l.Styles.Title = titleStyle
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			//
		}
	}
	m.list = l
	return m
}

type model struct {
	list    list.Model
	browser *browser.Api
	player  player.Player
	cfg     config.Value
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
			err := m.player.Stop()
			if err != nil {
				errMsg := fmt.Sprintf("error stopping station at exit", err.Error())
				slog.Error(errMsg)
			}

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

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(m.list.View())
}
