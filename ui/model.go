package ui

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
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

	// header status
	noPlayingMsg     = "Nothing playing"
	missingFavorites = "Some stations not found"
	prevTermErr      = "Could not terminate previous playback!"
	voteSuccesful    = "Station was voted successfully"
	statusMsgTimeout = 1 * time.Second

	// metadata
	volumeFmt          = "%3d%%%s"
	volumeStep         = 5
	seekStepSec        = 10
	playerPollInterval = 500 * time.Millisecond
)

func NewModel(cfg *config.Value, b *browser.Api, p *player.Player) *Model {
	m := newModel(cfg, b, p)
	progr := tea.NewProgram(m, tea.WithAltScreen())
	m.Progr = progr
	trapSignal(progr)
	go getPlayerMetadata(progr, m)
	return m
}

func newModel(cfg *config.Value, b *browser.Api, p *player.Player) *Model {
	lipgloss.DefaultRenderer().SetHasDarkBackground(true)

	delegate := newStationDelegate(cfg, p, b)

	infoModel := newInfoModel(b)
	m := Model{
		cfg:      cfg,
		browser:  b,
		player:   p,
		delegate: delegate,
		tabs: []uiTab{
			newFavoritesTab(infoModel),
			newBrowseTab(b, infoModel),
		},
		statusUpdate: make(chan struct{}),

		volumeBar: getVolumeBar(),
	}

	if len(cfg.Favorites) > 0 {
		m.toFavoritesTab()
	} else {
		m.toBrowseTab()
	}

	go m.statusHandler()
	return &m
}

func getVolumeBar() progress.Model {
	b := progress.New([]progress.Option{
		progress.WithWidth(10),
		progress.WithFillCharacters(fullChar, emptyChar),
		progress.WithSolidFill(secondColor),
		progress.WithoutPercentage(),
	}...)
	b.EmptyColor = secondColor
	return b
}

func getPlayerMetadata(progr *tea.Program, m *Model) {
	log := slog.With("func", "getPlayerMetadata")
	tick := time.NewTicker(playerPollInterval)
	for range tick.C {
		if m.delegate.currPlaying == nil {
			continue
		}
		m := m.player.Metadata()
		if m == nil {
			continue
		} else if m.Err != nil {
			log.Error("", "metadata", m.Err)
			continue
		}
		msg := fromMetadata(*m)
		progr.Send(msg)
	}
}

type Model struct {
	Progr *tea.Program

	ready    bool
	cfg      *config.Value // use cfg.volume
	browser  *browser.Api
	player   *player.Player
	delegate *stationDelegate

	tabs         []uiTab
	activeTabIdx uiTabIndex

	// display currently performed action or encountered error
	statusMsg    string
	statusUpdate chan struct{}

	// display station metadata
	playbackTime time.Duration
	spinner      *spinner.Model
	songTitle    string
	volumeBar    progress.Model

	width        int
	totHeight    int
	headerHeight int
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "ui.model.Update")
	activeTab := m.tabs[m.activeTabIdx]

	switch msg := msg.(type) {
	//
	// messages that need to reach all tabs
	//
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.totHeight = msg.Height
		header := m.headerView(msg.Width)
		m.headerHeight = strings.Count(header, "\n")
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
		return nil, tea.Quit

	case statusMsg:
		m.updateStatus(string(msg))
		return m, nil

	case metadataMsg:
		m.songTitle = msg.songTitle
		if msg.playbackTime != nil {
			m.playbackTime = *msg.playbackTime
		}
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

	case pauseRespMsg:
		if msg.err != "" {
			m.updateStatus(msg.err)
		} else {
			m.songTitle = ""
			m.spinner = nil
			m.delegate.keymap.pause.SetHelp("space", "resume")
		}
		return m, nil
	case playRespMsg:
		if msg.err != "" {
			m.updateStatus(msg.err)
			m.spinner = nil
		}
		m.delegate.keymap.pause.SetHelp("space", " pause")
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		} else if activeTab, ok := activeTab.(stationTab); ok && (activeTab.IsSearchEnabled() || activeTab.IsFiltering()) {
			break
		}

		d := m.delegate

		if key.Matches(msg, d.keymap.volumeDown) {
			return m, m.volumeCmd(false)
		}
		if key.Matches(msg, d.keymap.volumeUp) {
			return m, m.volumeCmd(true)
		}
		if key.Matches(msg, d.keymap.seekBack) {
			return m, m.seekCmd(-seekStepSec)
		}
		if key.Matches(msg, d.keymap.seekFw) {
			return m, m.seekCmd(seekStepSec)
		}

		if key.Matches(msg, d.keymap.pause) {
			if d.currPlaying != nil {
				return m, d.pauseCmd()
			} else if d.prevPlaying != nil {
				cmds := []tea.Cmd{m.initSpinner(), d.resumeCmd()}
				return m, tea.Batch(cmds...)
			} else {
				activeTab, ok := activeTab.(stationTab)
				if !ok {
					// TODO handle pause key for other tabs if necessary
					return m, nil
				}
				selStation, ok := activeTab.Stations().list.SelectedItem().(browser.Station)
				if ok {
					m.updateStatus(fmt.Sprintf("Connecting to %s...", selStation.Name))
					cmds := []tea.Cmd{m.initSpinner(), d.playCmd(selStation)}
					return m, tea.Batch(cmds...)
				}
			}
		}

		if activeTab, ok := activeTab.(stationTab); ok && activeTab.IsInfoEnabled() {
			break
		}

		if key.Matches(msg, d.keymap.playSelected) {
			activeTab, ok := activeTab.(stationTab)
			if !ok {
				// TODO handle enter for other tabs if necessary
				return m, nil
			}
			selStation, ok := activeTab.Stations().list.SelectedItem().(browser.Station)
			if ok {
				m.songTitle = ""
				m.playbackTime = 0
				m.updateStatus(fmt.Sprintf("Connecting to %s...", selStation.Name))
				cmds := []tea.Cmd{m.initSpinner(), d.playCmd(selStation)}
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

func (m *Model) statusHandler() {
	t := time.NewTimer(math.MaxInt64)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			m.statusMsg = ""
		case <-m.statusUpdate:
			t.Stop()
			t.Reset(statusMsgTimeout)
		}
	}
}

func (m *Model) toFavoritesTab() {
	m.delegate.keymap.toggleFavorite.SetEnabled(false)
	m.activeTabIdx = favoriteTabIx
}
func (m *Model) toBrowseTab() {
	m.delegate.keymap.toggleFavorite.SetEnabled(true)
	m.activeTabIdx = browseTabIx
}

func (m *Model) updateStatus(msg string) {
	slog.Debug("updateStatus", "old", m.statusMsg, "new", msg)
	m.statusMsg = msg
	go func() {
		m.statusUpdate <- struct{}{}
	}()
}

func (m *Model) Quit() {
	log := slog.With("method", "ui.model.quit")
	log.Info("----------------------Quitting----------------------")
	err := m.player.Stop()
	if err != nil {
		log.Error("player stop", "error", err.Error())
	}
	err = m.player.Close()
	if err != nil {
		slog.Error(fmt.Sprintf("player close error: %v", err))
	}
	err = config.Save(*m.cfg)
	if err != nil {
		log.Error("config save", "error", err.Error())
	}
}

func newSpinner() *spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{"⡷", "⣧", "⣏", "⡟", "⡷", "⣧", "⣏", "⡟"},
		FPS:    time.Second / 10,
	}
	s.Style = playStatusStyle
	return &s
}

func (m *Model) initSpinner() tea.Cmd {
	m.spinner = newSpinner()
	return m.spinner.Tick
}

func (m *Model) headerView(width int) string {
	var res strings.Builder
	status := ""
	if len(m.statusMsg) > 0 {
		status = statusBarStyle.Render(strings.Repeat(" ", headerPadDist) + m.statusMsg)
	}
	res.WriteString(status)
	appNameVers := statusBarStyle.Render(fmt.Sprintf("sonicradio v%v  ", m.cfg.Version))
	fill := max(0, width-lipgloss.Width(status)-lipgloss.Width(appNameVers)-2*headerPadDist)
	res.WriteString(statusBarStyle.Render(strings.Repeat(" ", fill)))
	res.WriteString(appNameVers)
	res.WriteString("\n\n")

	metadata := m.metadataView(width)
	res.WriteString(metadata)

	res.WriteString("\n\n")

	var renderedTabs []string
	renderedTabs = append(renderedTabs, tabGap.Render(strings.Repeat(" ", tabGapDistance)))
	for i := range m.tabs {
		if i == int(m.activeTabIdx) {
			renderedTabs = append(renderedTabs, activeTab.Render(m.activeTabIdx.String()))
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
	hFill := width - lipgloss.Width(row) - 2*headerPadDist
	gap := tabGap.Render(strings.Repeat(" ", max(0, hFill)))
	res.WriteString(lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap) + "\n\n")

	return res.String()
}

func (m *Model) metadataView(width int) string {
	metadataParts := []string{"", "", ""}
	gap := strings.Repeat(" ", headerPadDist)

	playTime := fmt.Sprintf("%s%03d:%02d:%02d%s",
		gap,
		int(m.playbackTime.Hours()),
		int(m.playbackTime.Minutes())%60,
		int(m.playbackTime.Seconds())%60,
		gap,
	)
	playTimeView := italicStyle.Render(playTime)
	metadataParts[0] = playTimeView

	volumeView := gap +
		m.volumeBar.ViewAs(float64(m.cfg.GetVolume())/100) +
		italicStyle.Render(fmt.Sprintf(volumeFmt, m.cfg.GetVolume(), gap))
	metadataParts[2] = volumeView

	playTimeW := lipgloss.Width(playTimeView)
	volumeW := lipgloss.Width(volumeView)
	maxW := max(0, width-playTimeW-volumeW-2*headerPadDist)

	var songView strings.Builder
	if m.delegate.currPlaying != nil {
		if m.spinner == nil {
			m.spinner = newSpinner()
		}
		var line strings.Builder
		line.WriteString(m.spinner.View())
		line.WriteString(itemStyle.MaxWidth(maxW - 1).Render(" " + m.delegate.currPlaying.Name))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(itemStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	} else if m.delegate.prevPlaying != nil {
		var line strings.Builder
		line.WriteString(playStatusStyle.Render(pauseChar))
		line.WriteString(itemStyle.MaxWidth(maxW - 1).Render(" " + m.delegate.prevPlaying.Name))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(itemStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	} else {
		var line strings.Builder
		line.WriteString(playStatusStyle.MaxWidth(maxW).Render(lineChar + " " + noPlayingMsg))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(itemStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	}
	songView.WriteString("\n")
	if m.songTitle != "" {
		var line strings.Builder
		line.WriteString(playStatusStyle.MaxWidth(maxW).Render("  " + m.songTitle))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(itemStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	} else if m.delegate.currPlaying != nil {
		var line strings.Builder
		line.WriteString(playStatusStyle.MaxWidth(maxW).Render("  " + m.delegate.currPlaying.Homepage))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(itemStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	} else if m.delegate.prevPlaying != nil {
		var line strings.Builder
		line.WriteString(playStatusStyle.MaxWidth(maxW).Render("  " + m.delegate.prevPlaying.Homepage))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(itemStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	}
	metadataParts[1] = songView.String()

	metadataRows := lipgloss.JoinHorizontal(lipgloss.Top, metadataParts...)
	return metadataRows
}

func (m Model) View() string {
	if !m.ready {
		return loadingMsg
	}

	var doc strings.Builder
	header := m.headerView(m.width)
	doc.WriteString(header)
	tabView := m.tabs[m.activeTabIdx].View()
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
func (m *Model) favoritesReqCmd() tea.Msg {
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

func (m *Model) topStationsCmd() tea.Msg {
	stations, err := m.browser.TopStations()
	res := topStationsRespMsg{stations: stations}
	if err != nil {
		res.statusMsg = statusMsg(err.Error())
	} else if len(stations) == 0 {
		res.viewMsg = noStationsFound
	}
	return res
}

func (m *Model) volumeCmd(up bool) tea.Cmd {
	return func() tea.Msg {
		currVol := m.cfg.GetVolume()
		newVol := currVol + volumeStep
		if !up {
			newVol = currVol - volumeStep
		}
		setVol, err := m.player.SetVolume(newVol)
		if err != nil {
			return volumeMsg{err}
		}
		m.cfg.SetVolume(setVol)
		return volumeMsg{}
	}
}

func (m *Model) seekCmd(amtSec int) tea.Cmd {
	return func() tea.Msg {
		log := slog.With("method", "model.seekCmd")
		metadata := m.player.Seek(amtSec)
		if metadata == nil {
			return nil
		} else if metadata.Err != nil {
			log.Error("seek", "error", metadata.Err)
			return nil
		}
		msg := fromMetadata(*metadata)
		return msg
	}
}

func logTeaMsg(msg tea.Msg, tag string) {
	log := slog.With("method", tag)
	switch msg.(type) {
	case favoritesStationRespMsg, topStationsRespMsg, searchRespMsg, toggleInfoMsg:
		log.Debug("tea.Msg", "type", fmt.Sprintf("%T", msg))
	case cursor.BlinkMsg, spinner.TickMsg, list.FilterMatchesMsg:
		break
	default:
		log.Debug("tea.Msg", "type", fmt.Sprintf("%T", msg), "value", msg, "#", fmt.Sprintf("%#v", msg))
	}
}
