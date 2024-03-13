package ui

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player"
)

var ready bool

func NewProgram(cfg *config.Value, b *browser.Api, p player.Player) *tea.Program {
	m := initialModel(cfg, b, p)
	progr := tea.NewProgram(m, tea.WithAltScreen())
	trapSignal(progr)
	return progr
}

func initialModel(cfg *config.Value, b *browser.Api, p player.Player) *model {
	delegate := newStationDelegate(p)
	activeIx := browseTabIx
	if len(cfg.Favorites) > 0 {
		activeIx = favoriteTabIx
	}
	m := model{
		cfg:       cfg,
		browser:   b,
		player:    p,
		delegate:  delegate,
		tabs:      []uiTab{newFavoritesTab(), newBrowseTab()},
		activeTab: activeIx,
	}
	return &m
}

type model struct {
	cfg      *config.Value
	browser  *browser.Api
	player   player.Player
	delegate *stationDelegate

	tabs         []uiTab
	activeTab    uiTabIndex
	width        int
	totHeight    int
	headerHeight int
}
type uiTabIndex uint8

func (t uiTabIndex) String() string {
	switch t {
	case favoriteTabIx:
		return "1. Favorites"
	case browseTabIx:
		return "2. Browse"
	case historyTabIx:
		return "3. History"
	}
	return ""
}

const (
	favoriteTabIx uiTabIndex = iota
	browseTabIx
	historyTabIx
	// configTab
)

type uiTab interface {
	Init(m *model) tea.Cmd
	Update(m *model, msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
	SetItems([]list.Item)
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.totHeight = msg.Height
		header := m.headerView(msg.Width)
		m.headerHeight = strings.Count(header, "\n")
		slog.Debug("width", "", m.width)
		slog.Debug("totHeight", "", m.totHeight)
		slog.Debug("headerHeight", "", m.headerHeight)

		var cmds []tea.Cmd
		if !ready {
			ready = true
			for i := range m.tabs {
				tcmd := m.tabs[i].Init(m)
				cmds = append(cmds, tcmd)
			}
		} else {
			for i := range m.tabs {
				_, tcmd := m.tabs[i].Update(m, msg)
				cmds = append(cmds, tcmd)
			}
		}
		return m, tea.Batch(cmds...)

	case topStationsResMsg:
		items := make([]list.Item, len(msg.stations))
		for i := 0; i < len(msg.stations); i++ {
			items[i] = msg.stations[i]
		}
		m.tabs[browseTabIx].SetItems(items)

	case favoritesStationResMsg:
		items := make([]list.Item, len(msg.stations))
		for i := 0; i < len(msg.stations); i++ {
			items[i] = msg.stations[i]
		}
		m.tabs[favoriteTabIx].SetItems(items)
	}

	model, cmd := m.tabs[m.activeTab].Update(m, msg)
	return model, cmd
}

func (m *model) stop() {
	slog.Info("----------------------Quitting----------------------")
	err := m.player.Stop()
	if err != nil {
		slog.Error("error stopping station at exit", "error", err.Error())
	}
}

func (m *model) headerView(width int) string {
	var renderedTabs []string

	for i := range m.tabs {
		if i == int(m.activeTab) {
			renderedTabs = append(renderedTabs, activeTab.Render(m.activeTab.String()))
		} else {
			renderedTabs = append(renderedTabs, tab.Render(uiTabIndex(i).String()))
		}
	}
	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderedTabs...,
	)
	hFill := width - lipgloss.Width(row) - 2
	gap := tabGap.Render(strings.Repeat(" ", max(0, hFill)))
	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap) + "\n\n"
}

func (m model) View() string {
	if !ready {
		return "\n  Fetching stations"
	}

	var doc strings.Builder
	header := m.headerView(m.width)
	doc.WriteString(header)

	tabView := m.tabs[m.activeTab].View()
	doc.WriteString(tabView)
	return docStyle.Render(doc.String())
}

func trapSignal(p *tea.Program) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	go func() {
		osCall := <-signals
		slog.Debug(fmt.Sprintf("received OS signal %+v", osCall))
		p.Send(quitMsg{})
	}()
}

// tea.Msg
type (
	// used for os signal quit not handled by the list model
	quitMsg struct{}

	favoritesStationResMsg struct {
		stations []browser.Station
	}
	topStationsResMsg struct {
		stations []browser.Station
	}
)

// tea.Cmd
func (m *model) favoritesReqCmd() tea.Msg {
	items := make([]browser.Station, 0)
	for i := range m.cfg.Favorites {
		slog.Debug("get station", "uuid", m.cfg.Favorites[i])
		s := m.browser.GetStation(m.cfg.Favorites[i])
		if s != nil {
			items = append(items, *s)
		}
	}
	slog.Debug("favorite stations", "lenght", len(items))
	return favoritesStationResMsg{items}
}

func (m *model) topStationsCmd() tea.Msg {
	stations := m.browser.TopStations()
	return topStationsResMsg{stations}
}
