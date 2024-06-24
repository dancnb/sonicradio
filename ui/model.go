package ui

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player"
)

const (
	// view messages
	loadingMsg          = "\n Fetching stations... \n"
	noFavoritesAddedMsg = "\n No favorite stations added.\n"
	noStationsFound     = "\n No stations found. \n"
	// header status messages
	noPlayingMsg     = "Nothing playing"
	missingFavorites = "Some stations not found"

	playerPollInterval = 500 * time.Millisecond
)

func NewProgram(cfg *config.Value, b *browser.Api, p player.Player) *tea.Program {
	m := initialModel(cfg, b, p)
	progr := tea.NewProgram(m, tea.WithAltScreen())
	trapSignal(progr)
	go getPlayerMetadata(progr, m)
	return progr
}

func initialModel(cfg *config.Value, b *browser.Api, p player.Player) *model {
	lipgloss.DefaultRenderer().SetHasDarkBackground(true)

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
		tabs:      []uiTab{newFavoritesTab(), newBrowseTab(b)},
		activeTab: activeIx,
		statusMsg: noPlayingMsg,
	}
	return &m
}

func getPlayerMetadata(progr *tea.Program, m *model) {
	log := slog.With("func", "getPlayerMetadata")
	tick := time.NewTicker(playerPollInterval)
	for range tick.C {
		if m.delegate.currPlaying == nil {
			continue
		}
		log.Debug("", "currPlaying", m.delegate.currPlaying.URL)
		m := m.player.Metadata()
		log.Debug("", "metadata", m)
		if m == nil {
			continue
		} else if m.Err != nil {
			progr.Send(playRespMsg{m.Err.Error()})
			continue
		}
		progr.Send(titleMsg(m.Title))
	}
}

type model struct {
	ready    bool
	cfg      *config.Value
	browser  *browser.Api
	player   player.Player
	delegate *stationDelegate

	tabs      []uiTab
	activeTab uiTabIndex

	statusMsg string // display currently performed action or encountered error
	titleMsg  string // display station metadata (song name)
	spinner   *spinner.Model

	width        int
	totHeight    int
	headerHeight int
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "ui.model.Update")
	activeTab := m.tabs[m.activeTab]

	log := slog.With("method", "ui.model.Update")
	switch msg := msg.(type) {
	//
	// messages that need to reach all tabs
	//
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.totHeight = msg.Height
		header := m.headerView(msg.Width)
		m.headerHeight = strings.Count(header, "\n")
		log.Debug("width", "", m.width)
		log.Debug("totHeight", "", m.totHeight)
		log.Debug("headerHeight", "", m.headerHeight)
		var cmds []tea.Cmd
		if !m.ready {
			m.ready = true
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
		m.quit()
		return nil, tea.Quit

	case playRespMsg:
		m.statusMsg = msg.err
		if msg.err != "" {
			m.spinner = nil
			d := m.delegate
			if d.currPlaying != nil {
				_, err := d.stopStation(*d.currPlaying)
				if err != nil {
					m.statusMsg = "Could not terminate previous playback!"
					return m, nil
				}
			}
		}
		return m, nil
	case statusMsg:
		m.updateStatus(msg)
		return m, nil

	case titleMsg:
		m.titleMsg = string(msg)
		return m, nil

	case spinner.TickMsg:
		if m.spinner == nil {
			return m, nil
		}
		var cmd tea.Cmd
		s, cmd := m.spinner.Update(msg)
		m.spinner = &s
		return m, cmd

	//
	// messages that need to reach a particular tab
	//
	case topStationsRespMsg, searchRespMsg:
		return m.tabs[browseTabIx].Update(m, msg)

	case favoritesStationRespMsg:
		return m.tabs[favoriteTabIx].Update(m, msg)

	case toggleFavoriteMsg:
		return m.tabs[favoriteTabIx].Update(m, msg)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quit()
			return m, tea.Quit
		} else if activeTab.IsSearchEnabled() {
			break
		} else if activeTab.IsFiltering() {
			break
		}

		d := m.delegate

		switch {
		case key.Matches(msg, d.keymap.pause):
			m.titleMsg = ""
			m.spinner = nil
			if d.currPlaying != nil {
				_, err := d.stopStation(*d.currPlaying)
				if err != nil {
					m.statusMsg = "Could not terminate previous playback!"
					return m, nil
				}
			} else if d.prevPlaying != nil {
				m.statusMsg = fmt.Sprintf("Connecting to %s...", d.prevPlaying.Name)
				cmds := []tea.Cmd{m.initSpinner(), d.playCmd(d.prevPlaying)}
				return m, tea.Batch(cmds...)
			} else {
				if m.activeTab != favoriteTabIx && m.activeTab != browseTabIx {
					// TODO handle enter for other tabs if necessary
					return m, nil
				}
				selStation, ok := activeTab.List().SelectedItem().(browser.Station)
				if ok {
					m.statusMsg = fmt.Sprintf("Connecting to %s...", selStation.Name)
					cmds := []tea.Cmd{m.initSpinner(), d.playCmd(&selStation)}
					return m, tea.Batch(cmds...)
				}
			}

		case key.Matches(msg, d.keymap.playSelected):
			if m.activeTab != favoriteTabIx && m.activeTab != browseTabIx {
				// TODO handle enter for other tabs if necessary
				return m, nil
			}
			selStation, ok := activeTab.List().SelectedItem().(browser.Station)
			if ok {
				m.titleMsg = ""
				m.spinner = nil
				_, err := d.stopStation(selStation)
				if err != nil {
					m.statusMsg = "Could not terminate previous playback!"
					return m, nil
				}
				m.statusMsg = fmt.Sprintf("Connecting to %s...", selStation.Name)
				cmds := []tea.Cmd{m.initSpinner(), d.playCmd(&selStation)}
				return m, tea.Batch(cmds...)
			}
		}

	}

	//
	// messages that need to reach active tab
	//
	model, cmd := activeTab.Update(m, msg)
	return model, cmd
}

func (m *model) updateStatus(msg statusMsg) {
	if msg != "" {
		m.statusMsg = string(msg)
	}
}

func (m *model) quit() {
	log := slog.With("method", "ui.model.quit")
	log.Info("----------------------Quitting----------------------")
	err := m.player.Stop()
	if err != nil {
		log.Error("error stopping station at exit", "error", err.Error())
	}
	err = config.Save(*m.cfg)
	if err != nil {
		log.Error("error saving config", "error", err.Error())
	}
}

func (m *model) initSpinner() tea.Cmd {
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{"⡷", "⣧", "⣏", "⡟", "⡷", "⣧", "⣏", "⡟"},
		FPS:    time.Second / 10,
	}
	s.Style = playStatusStyle
	m.spinner = &s
	return m.spinner.Tick
}

func (m *model) headerView(width int) string {
	var res strings.Builder

	if m.statusMsg != "" {
		res.WriteString(playStatusStyle.Render(lineChar + " " + m.statusMsg))
	} else if m.delegate.currPlaying != nil {
		res.WriteString(m.spinner.View())
		res.WriteString(itemStyle.Render(" " + m.delegate.currPlaying.Name))
	} else if m.delegate.prevPlaying != nil {
		res.WriteString(playStatusStyle.Render(pauseChar))
		res.WriteString(itemStyle.Render(" " + m.delegate.prevPlaying.Name))
	}
	res.WriteString("\n")
	if m.titleMsg != "" {
		res.WriteString(playStatusStyle.Render("  " + m.titleMsg))
	} else if m.delegate.currPlaying != nil {
		res.WriteString(playStatusStyle.Render("  " + m.delegate.currPlaying.Homepage))
	} else if m.delegate.prevPlaying != nil {
		res.WriteString(playStatusStyle.Render("  " + m.delegate.prevPlaying.Homepage))
	}
	res.WriteString("\n\n")

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
	hFill := width - lipgloss.Width(row)
	gap := tabGap.Render(strings.Repeat(" ", max(0, hFill)))
	res.WriteString(lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap) + "\n\n")

	return res.String()
}

func (m model) View() string {
	log := slog.With("method", "ui.model.View")
	log.Debug("", "statusMsg", m.statusMsg, "titleMsg", m.titleMsg)
	if !m.ready {
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
			viewMsg: noFavoritesAddedMsg,
		}
	}

	stations, err := m.browser.GetStations(m.cfg.Favorites)
	res := favoritesStationRespMsg{stations: stations}
	if err != nil {
		res.statusMsg = statusMsg(err.Error())
	} else if len(stations) == 0 {
		res.viewMsg = noStationsFound
	}
	return res
}

func (m *model) topStationsCmd() tea.Msg {
	stations, err := m.browser.TopStations()
	res := topStationsRespMsg{stations: stations}
	if err != nil {
		res.statusMsg = statusMsg(err.Error())
	} else if len(stations) == 0 {
		res.viewMsg = noStationsFound
	}
	return res
}

func logTeaMsg(msg tea.Msg, tag string) {
	log := slog.With("method", tag)
	switch msg.(type) {
	case favoritesStationRespMsg, topStationsRespMsg, searchRespMsg:
		log.Debug("tea.Msg", "type", fmt.Sprintf("%T", msg))
	default:
		log.Debug("tea.Msg", "type", fmt.Sprintf("%T", msg), "value", msg, "#", fmt.Sprintf("%#v", msg))
	}
}
