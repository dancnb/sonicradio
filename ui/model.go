package ui

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player"
)

const (
	loadingMsg          = "\n  Fetching stations... \n"
	noFavoritesAddedMsg = "\n No favorites added\n"
)

var ready bool

func NewProgram(cfg *config.Value, b *browser.Api, p player.Player) *tea.Program {
	m := initialModel(cfg, b, p)
	progr := tea.NewProgram(m, tea.WithAltScreen())
	trapSignal(progr)
	return progr
}

func initialModel(cfg *config.Value, b *browser.Api, p player.Player) *model {
	lipgloss.DefaultRenderer().SetHasDarkBackground(true)
	// lipgloss.DefaultRenderer().Output().SetBackgroundColor(backgroundColor)

	delegate := newStationDelegate(cfg, p)
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

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	//
	// messages that need to reach all tabs
	//
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

	case quitMsg:
		m.stop()
		return nil, tea.Quit

	//
	// messages that need to reach a particular tab
	//
	case topStationsRespMsg:
		// TODO handle errMsg
		return m.tabs[browseTabIx].Update(m, msg)

	case favoritesStationRespMsg:
		// TODO handle errMsg
		return m.tabs[favoriteTabIx].Update(m, msg)

	case toggleFavoriteMsg:
		return m.tabs[favoriteTabIx].Update(m, msg)
	}

	//
	// messages that need to reach active tab
	//
	model, cmd := m.tabs[m.activeTab].Update(m, msg)
	return model, cmd
}

func (m *model) stop() {
	slog.Info("----------------------Quitting----------------------")
	err := m.player.Stop()
	if err != nil {
		slog.Error("error stopping station at exit", "error", err.Error())
	}
	err = config.Save(*m.cfg)
	if err != nil {
		slog.Error("error saving config", "error", err.Error())
	}
}

func (m *model) headerView(width int) string {
	var renderedTabs []string

	renderedTabs = append(renderedTabs, tabGap.Render(strings.Repeat(" ", tabGapDistance)))
	for i := range m.tabs {
		if i == int(m.activeTab) {
			renderedTabs = append(renderedTabs, activeTab.Render(m.activeTab.String()))
		} else {
			renderedTabs = append(renderedTabs, inactiveTab.Render(uiTabIndex(i).String()))
		}
		if i < len(m.tabs)-1 {
			renderedTabs = append(renderedTabs, tabGap.Render(strings.Repeat(" ", tabGapDistance)))
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
		return loadingMsg
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

// tea.Cmd
func (m *model) favoritesReqCmd() tea.Msg {
	if len(m.cfg.Favorites) == 0 {
		return favoritesStationRespMsg{
			respMsg: respMsg{viewMsg: noFavoritesAddedMsg},
		}
	}

	stations, err := m.browser.GetStations(m.cfg.Favorites)
	res := respMsg{}
	if err != nil {
		res.viewMsg = "No stations found"
		res.errMsg = err
	}
	return favoritesStationRespMsg{
		respMsg:  res,
		stations: stations,
	}
}

func (m *model) topStationsCmd() tea.Msg {
	stations, err := m.browser.TopStations()
	res := respMsg{}
	if err != nil {
		res.viewMsg = "No stations found"
		res.errMsg = err
	}
	return topStationsRespMsg{
		respMsg:  res,
		stations: stations,
	}
}
