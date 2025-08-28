package ui

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode"

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
	// view messages, nweline is important to sync with list no items view
	loadingMsg          = "\n  Fetching stations... \n"
	noFavoritesAddedMsg = "\n  No favorite stations added.\n"
	noStationsFound     = "\n  No stations found. \n"
	emptyHistoryMsg     = "\n  No playback history available. \n"

	// header status
	noPlayingMsg     = "Nothing playing"
	missingFavorites = "Some stations not found"
	prevTermErr      = "Could not terminate previous playback!"
	voteSuccesful    = "Station was voted successfully"
	statusMsgTimeout = 1 * time.Second

	// metadata
	volumeFmt          = "%3d%%%s"
	playerPollInterval = 500 * time.Millisecond
)

func NewModel(ctx context.Context, cfg *config.Value, b *browser.Api, p *player.Player) *Model {
	m := newModel(ctx, cfg, b, p)
	progr := tea.NewProgram(m, tea.WithAltScreen(), tea.WithContext(ctx))
	m.Progr = progr
	trapSignal(progr)
	go updatePlayerMetadata(ctx, progr, m)
	return m
}

func newModel(ctx context.Context, cfg *config.Value, b *browser.Api, p *player.Player) *Model {
	style := NewStyle(cfg.Theme)

	delegate := newStationDelegate(cfg, style, p, b)

	infoModel := newInfoModel(b, style)
	m := Model{
		cfg:          cfg,
		style:        style,
		browser:      b,
		player:       p,
		delegate:     delegate,
		statusUpdate: make(chan struct{}),

		volumeBar: getVolumeBar(style.GetSecondColor()),
	}
	m.tabs = []uiTab{
		newFavoritesTab(infoModel, style),
		newBrowseTab(ctx, b, infoModel, style),
		newHistoryTab(ctx, cfg, style),
		newSettingsTab(ctx, cfg, style, p.AvailablePlayerTypes(), m.changeTheme),
	}

	if len(cfg.Favorites) > 0 {
		m.toFavoritesTab()
	} else {
		m.toBrowseTab()
	}

	go m.statusHandler(ctx)
	return &m
}

func getVolumeBar(secondColor string) progress.Model {
	b := progress.New([]progress.Option{
		progress.WithWidth(10),
		progress.WithSolidFill(secondColor),
		progress.WithoutPercentage(),
	}...)
	b.EmptyColor = secondColor
	return b
}

func updatePlayerMetadata(ctx context.Context, progr *tea.Program, m *Model) {
	tick := time.NewTicker(playerPollInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			pollMetadata(m, progr)
		}
	}
}

func pollMetadata(m *Model, progr *tea.Program) {
	log := slog.With("method", "pollMetadata")

	m.delegate.playingMtx.RLock()
	defer m.delegate.playingMtx.RUnlock()

	if m.delegate.currPlaying == nil {
		return
	}
	metadata := m.player.Metadata()
	if metadata == nil {
		return
	} else if metadata.Err != nil {
		log.Error("", "metadata", metadata.Err)
		return
	}
	msg := getMetadataMsg(*m.delegate.currPlaying, *metadata)
	go progr.Send(msg)
}

type Model struct {
	Progr *tea.Program

	ready    bool
	cfg      *config.Value
	style    *Style
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
		go m.cfg.AddHistoryEntry(
			time.Now(),
			strings.TrimSpace(msg.stationUuid),
			strings.TrimSpace(msg.stationName),
			strings.TrimSpace(msg.songTitle),
		)
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
			m.spinner = nil
			m.delegate.keymap.pause.SetHelp("space", "resume")
		}
		return m, nil
	case playRespMsg:
		if msg.err != "" {
			m.updateStatus(msg.err)
			m.spinner = nil
		}
		m.delegate.keymap.pause.SetHelp("space", "pause")
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		} else if activeTab, ok := activeTab.(filteringTab); ok && activeTab.IsFiltering() {
			break
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
			if m.activeTabIdx == settingsTabIx {
				return m.tabs[settingsTabIx].Update(m, msg)
			}
			return m, m.seekCmd(-config.SeekStepSec)
		}
		if key.Matches(msg, d.keymap.seekFw) {
			if m.activeTabIdx == settingsTabIx {
				return m.tabs[settingsTabIx].Update(m, msg)
			}
			return m, m.seekCmd(config.SeekStepSec)
		}

		if key.Matches(msg, d.keymap.pause) {
			if m.activeTabIdx == settingsTabIx {
				return m.tabs[settingsTabIx].Update(m, msg)
			}

			if resM, resCmd := m.handlePauseKey(); resM != nil {
				return resM, resCmd
			}
			activeTab, ok := activeTab.(stationTab)
			if !ok {
				break
			}
			selStation, ok := activeTab.Stations().list.SelectedItem().(browser.Station)
			if ok {
				return m, m.playStationCmd(selStation)
			}
		}

		if activeTab, ok := activeTab.(stationTab); ok && activeTab.IsInfoEnabled() {
			break
		}

		if key.Matches(msg, d.keymap.playSelected) {
			if m.activeTabIdx == settingsTabIx {
				return m.tabs[settingsTabIx].Update(m, msg)
			}

			activeTab, ok := activeTab.(stationTab)
			if !ok {
				break
			}
			selStation, ok := activeTab.Stations().list.SelectedItem().(browser.Station)
			if ok {
				return m, m.playStationCmd(selStation)
			}
		}
	}

	//
	// messages that need to reach active tab
	//
	model, cmd := activeTab.Update(m, msg)
	return model, cmd
}

func (m *Model) handlePauseKey() (*Model, tea.Cmd) {
	log := slog.With("method", "ui.Model.handlePauseKey")
	log.Info("begin")
	defer log.Info("end")

	m.delegate.playingMtx.RLock()
	defer m.delegate.playingMtx.RUnlock()

	if m.delegate.currPlaying != nil {
		return m, m.delegate.pauseCmd()
	} else if m.delegate.prevPlaying != nil {
		cmds := []tea.Cmd{m.initSpinner(), m.delegate.resumeCmd()}
		return m, tea.Batch(cmds...)
	}
	return nil, nil
}

func (m *Model) statusHandler(ctx context.Context) {
	t := time.NewTimer(math.MaxInt64)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
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
	m.delegate.keymap.toggleAutoplay.SetEnabled(true)
	m.activeTabIdx = favoriteTabIx
}

func (m *Model) toBrowseTab() {
	m.delegate.keymap.toggleFavorite.SetEnabled(true)
	m.delegate.keymap.toggleAutoplay.SetEnabled(false)
	m.activeTabIdx = browseTabIx
}

func (m *Model) toHistoryTab() {
	m.activeTabIdx = historyTabIx
}

func (m *Model) toSettingsTab() tea.Cmd {
	m.activeTabIdx = settingsTabIx
	st := m.tabs[settingsTabIx].(*settingsTab)
	return st.onEnter()
}

func (m *Model) updateStatus(msg string) {
	slog.Info("updateStatus", "old", m.statusMsg, "new", msg)
	m.statusMsg = msg
	go func() {
		m.statusUpdate <- struct{}{}
	}()
}

func (m *Model) Quit() {
	log := slog.With("method", "ui.model.quit")
	log.Info("----------------------Quitting----------------------")

	// stop player
	err := m.player.Stop()
	if err != nil {
		log.Error("player stop", "error", err.Error())
	}
	err = m.player.Close()
	if err != nil {
		slog.Error(fmt.Sprintf("player close error: %v", err))
	}

	// save config
	autoplayFound := false
	for _, v := range m.cfg.Favorites {
		if v == m.cfg.AutoplayFavorite {
			autoplayFound = true
			break
		}
	}
	if !autoplayFound {
		m.cfg.AutoplayFavorite = ""
	}
	st := m.tabs[settingsTabIx].(*settingsTab)
	st.updateConfig()

	err = m.cfg.Save()
	if err != nil {
		log.Info(fmt.Sprintf("config save err: %v", err))
	}
	log.Info("config saved")
}

func (m *Model) newSpinner() *spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{"⡷", "⣧", "⣏", "⡟", "⡷", "⣧", "⣏", "⡟"},
		FPS:    time.Second / 10,
	}
	s.Style = m.style.SongTitleStyle
	return &s
}

func (m *Model) initSpinner() tea.Cmd {
	m.spinner = m.newSpinner()
	return m.spinner.Tick
}

func (m *Model) headerView(width int) string {
	var res strings.Builder
	status := ""
	if len(m.statusMsg) > 0 {
		status = m.style.StatusBarStyle.Render(strings.Repeat(" ", HeaderPadDist) + m.statusMsg)
	}
	res.WriteString(status)
	appNameVers := m.style.StatusBarStyle.Render(fmt.Sprintf("sonicradio v%v  ", m.cfg.Version))
	fill := max(0, width-lipgloss.Width(status)-lipgloss.Width(appNameVers)-2*HeaderPadDist)
	res.WriteString(m.style.StatusBarStyle.Render(strings.Repeat(" ", fill)))
	res.WriteString(appNameVers)
	res.WriteString("\n\n")

	metadata := m.metadataView(width)
	res.WriteString(metadata)

	res.WriteString("\n\n")

	var renderedTabs []string
	renderedTabs = append(renderedTabs, m.style.TabGap.Render(strings.Repeat(" ", TabGapDistance)))
	for i := range m.tabs {
		if i == int(m.activeTabIdx) {
			tabName := m.activeTabIdx.String()
			renderedTab := m.renderTabName(tabName, &m.style.ActiveTabInner, &m.style.ActiveTabInnerHighlight)
			renderedTabs = append(renderedTabs, m.style.ActiveTabBorder.Render(renderedTab.String()))
		} else {
			tabName := uiTabIndex(i).String()
			renderedTab := m.renderTabName(tabName, &m.style.InactiveTabInner, &m.style.InactiveTabInnerHighlight)
			renderedTabs = append(renderedTabs, m.style.InactiveTabBorder.Render(renderedTab.String()))
		}
		if i < len(m.tabs)-1 {
			renderedTabs = append(renderedTabs, m.style.TabGap.Render(strings.Repeat(" ", TabGapDistance)))
		}
	}
	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderedTabs...,
	)
	hFill := width - lipgloss.Width(row) - 2*HeaderPadDist
	gap := m.style.TabGap.Render(strings.Repeat(" ", max(0, hFill)))
	res.WriteString(lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap) + "\n\n")

	return res.String()
}

func (*Model) renderTabName(tabName string, tabInner *lipgloss.Style, tabInnerHighlight *lipgloss.Style) strings.Builder {
	highlight := false
	var renderTab strings.Builder
	for _, r := range tabName {
		rStr := fmt.Sprintf("%c", r)
		if unicode.IsSpace(r) {
			renderTab.WriteString(tabInner.Render(rStr))
		} else if !highlight {
			renderTab.WriteString(tabInnerHighlight.Render(rStr))
			highlight = true
		} else {
			renderTab.WriteString(tabInner.Render(rStr))
		}
	}
	return renderTab
}

func (m *Model) metadataView(width int) string {
	metadataParts := []string{"", "", ""}
	gap := strings.Repeat(" ", HeaderPadDist)

	playTime := fmt.Sprintf("%s%03d:%02d:%02d%s",
		gap,
		int(m.playbackTime.Hours()),
		int(m.playbackTime.Minutes())%60,
		int(m.playbackTime.Seconds())%60,
		gap,
	)
	playTimeView := m.style.ItalicStyle.Render(playTime)
	metadataParts[0] = playTimeView

	volumeView := gap +
		m.volumeBar.ViewAs(float64(m.cfg.GetVolume())/100) +
		m.style.ItalicStyle.Render(fmt.Sprintf(volumeFmt, m.cfg.GetVolume(), gap))
	metadataParts[2] = volumeView

	playTimeW := lipgloss.Width(playTimeView)
	volumeW := lipgloss.Width(volumeView)
	maxW := max(0, width-playTimeW-volumeW-2*HeaderPadDist)

	var songView strings.Builder

	m.delegate.playingMtx.RLock()
	defer m.delegate.playingMtx.RUnlock()

	if m.delegate.currPlaying != nil {
		if m.spinner == nil {
			m.spinner = m.newSpinner()
		}
		var line strings.Builder
		line.WriteString(m.spinner.View())
		line.WriteString(
			m.style.PrimaryColorStyle.MaxWidth(maxW - 1).Render(
				" " + m.delegate.currPlaying.Name))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(m.style.PrimaryColorStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	} else if m.delegate.prevPlaying != nil {
		var line strings.Builder
		line.WriteString(m.style.SongTitleStyle.Render(PauseChar))
		line.WriteString(
			m.style.PrimaryColorStyle.MaxWidth(maxW - 1).Render(
				" " + m.delegate.prevPlaying.Name))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(m.style.PrimaryColorStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	} else {
		var line strings.Builder
		line.WriteString(m.style.SongTitleStyle.MaxWidth(maxW).Render(LineChar + " " + noPlayingMsg))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(m.style.PrimaryColorStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	}
	songView.WriteString("\n")
	if m.songTitle != "" {
		var line strings.Builder
		line.WriteString(m.style.SongTitleStyle.MaxWidth(maxW).Render("  " + m.songTitle))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(m.style.PrimaryColorStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	} else if m.delegate.currPlaying != nil {
		var line strings.Builder
		line.WriteString(
			m.style.SongTitleStyle.MaxWidth(maxW).Render(
				"  " + m.delegate.currPlaying.Homepage))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(m.style.PrimaryColorStyle.Render(strings.Repeat(" ", fill)))
		songView.WriteString(line.String())
	} else if m.delegate.prevPlaying != nil {
		var line strings.Builder
		line.WriteString(
			m.style.SongTitleStyle.MaxWidth(maxW).Render(
				"  " + m.delegate.prevPlaying.Homepage))
		fill := max(0, maxW-lipgloss.Width(line.String()))
		line.WriteString(m.style.PrimaryColorStyle.Render(strings.Repeat(" ", fill)))
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
	return m.style.DocStyle.Render(doc.String())
}

func (m *Model) changeStationView() {
	log := slog.With("method", "ui.Model.changeStationView")
	m.cfg.StationView = (m.cfg.StationView + 1) % 3
	log.Info(fmt.Sprintf("new stationView=%s", m.cfg.StationView.String()))
	m.delegate.setStationView(m.cfg.StationView)
	for tIx := range m.tabs {
		if st, ok := m.tabs[tIx].(stationTab); ok && st.Stations() != nil {
			st.Stations().list.SetDelegate(m.delegate)
		}
	}
}

func (m *Model) changeTheme(themeIdx int) {
	m.style.SetThemeIdx(themeIdx)
	m.cfg.Theme = themeIdx
	if m.spinner != nil {
		m.spinner.Style = m.style.SongTitleStyle
	}
	m.volumeBar.FullColor = m.style.GetSecondColor()
	m.volumeBar.EmptyColor = m.style.GetSecondColor()

	helpStyle := m.style.HelpStyles()
	for i := range m.tabs {
		if t, ok := m.tabs[i].(stationTab); ok {
			m.style.TextInputSyle(&t.Stations().list.FilterInput, stationsFilterPrompt, stationsFilterPlaceholder)
			t.Stations().list.Help.Styles = helpStyle
			t.Stations().list.Styles.HelpStyle = m.style.HelpStyle
			t.Stations().list.Styles.NoItems = m.style.NoItemsStyle
			t.Stations().infoModel.help.Styles = helpStyle

			if browse, ok := t.(*browseTab); ok {
				for iIdx := range browse.searchModel.inputs {
					input := browse.searchModel.inputs[iIdx].TextInput()
					m.style.TextInputSyle(input, input.Prompt, input.Placeholder)
					input.PromptStyle = m.style.PromptStyle
				}
				browse.searchModel.help.Styles = helpStyle
			}

		} else if ht, ok := m.tabs[i].(*historyTab); ok {
			m.style.TextInputSyle(&ht.list.FilterInput, stationsFilterPrompt, historyFilterPlaceholder)
			ht.list.Help.Styles = helpStyle
			ht.list.Styles.HelpStyle = m.style.HelpStyle
			ht.list.Styles.NoItems = m.style.NoItemsStyle

		} else if st, ok := m.tabs[i].(*settingsTab); ok {
			for iIdx := range st.inputs {
				if st.inputs[iIdx] == nil || st.inputs[iIdx].TextInput() == nil {
					continue
				}
				input := st.inputs[iIdx].TextInput()
				m.style.TextInputSyle(input, input.Prompt, input.Placeholder)
				input.PromptStyle = m.style.PromptStyle
			}
			st.help.Styles = helpStyle
		}
	}
}

func trapSignal(p *tea.Program) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	go func() {
		osCall := <-signals
		slog.Info(fmt.Sprintf("received OS signal %+v", osCall))
		p.Send(quitMsg{})
	}()
}

func logTeaMsg(msg tea.Msg, tag string) {
	log := slog.With("method", tag)
	switch msg.(type) {
	case favoritesStationRespMsg, topStationsRespMsg, searchRespMsg, toggleInfoMsg:
		log.Info("tea.Msg", "type", fmt.Sprintf("%T", msg))
	case cursor.BlinkMsg, spinner.TickMsg, list.FilterMatchesMsg:
		break
	default:
		log.Info("tea.Msg", "type", fmt.Sprintf("%T", msg), "value", msg, "#", fmt.Sprintf("%#v", msg))
	}
}
