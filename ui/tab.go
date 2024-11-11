package ui

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
)

type uiTabIndex uint8

func (t uiTabIndex) String() string {
	switch t {
	case favoriteTabIx:
		return "Favorites"
	case browseTabIx:
		return " Browse "
	}
	return ""
}

const (
	favoriteTabIx uiTabIndex = iota
	browseTabIx
	// historyTabIx
	// configTab
)

type uiTab interface {
	Init(m *Model) tea.Cmd
	Update(m *Model, msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type stationTab interface {
	uiTab
	Stations() *stationsTabBase
	IsFiltering() bool
	IsSearchEnabled() bool
	IsInfoEnabled() bool
}

const jumpTimeout = 250 * time.Millisecond

type jumpInfo struct {
	position int
	last     time.Time
}

func (jump jumpInfo) isActive() bool {
	return jump.last.Add(jumpTimeout).After(time.Now())
}

type stationsTabBase struct {
	uiTab
	list       list.Model
	viewMsg    string
	listKeymap listKeymap
	jump       jumpInfo
	infoModel  *infoModel
}

func newStationsTab(k listKeymap, infoModel *infoModel) stationsTabBase {
	t := stationsTabBase{
		listKeymap: k,
		infoModel:  infoModel,
	}
	return t
}

func (t *stationsTabBase) doJump(msg tea.KeyMsg) {
	jumpIdx := t.jumpIdx(msg)
	if jumpIdx > 0 && jumpIdx <= len(t.list.Items()) {
		t.list.Select(jumpIdx - 1)
	}
}

func (t *stationsTabBase) jumpIdx(msg tea.Msg) int {
	log := slog.With("method", "ui.stattionsTab.getJumpIdx")
	digit, _ := strconv.Atoi(msg.(tea.KeyMsg).String())
	log.Debug("", "digit", digit, "oldPos", t.jump.position)
	if t.jump.isActive() {
		t.jump.position = t.jump.position*10 + digit
	} else {
		t.jump.position = digit
	}
	t.jump.last = time.Now()
	log.Debug("", "newPos", t.jump.position)
	return t.jump.position
}

func (t *stationsTabBase) View() string {
	if t.viewMsg != "" {
		var sections []string
		availHeight := t.list.Height()
		help := t.list.Styles.HelpStyle.Render(t.list.Help.View(t.list))
		availHeight -= lipgloss.Height(help)
		viewSection := viewStyle.Height(availHeight).Render(t.viewMsg)
		sections = append(sections, viewSection)
		sections = append(sections, help)
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}
	return t.list.View()
}

func (t *stationsTabBase) Stations() *stationsTabBase { return t }

func (t *stationsTabBase) IsFiltering() bool {
	return t.list.FilterState() == list.Filtering
}

func (t *stationsTabBase) toNowPlaying(m *Model) {
	uuid := ""
	if m.delegate.currPlaying != nil {
		uuid = m.delegate.currPlaying.Stationuuid
	} else if m.delegate.prevPlaying != nil {
		uuid = m.delegate.prevPlaying.Stationuuid
	} else {
		return
	}
	selIndex := -1
	items := t.list.VisibleItems()
	for ix := range items {
		if items[ix].(browser.Station).Stationuuid == uuid {
			selIndex = ix
			break
		}
	}
	slog.Debug("method", "ui.baseTab.toNowPlaying", "selIndex", selIndex)
	if selIndex > -1 {
		t.list.Select(selIndex)
	}
}

func (t *stationsTabBase) IsSearchEnabled() bool {
	return false
}

func (t *stationsTabBase) IsInfoEnabled() bool {
	return t.infoModel != nil && t.infoModel.enabled
}

func (*stationsTabBase) newSizeMsg(sizeMsg tea.WindowSizeMsg, m *Model) tea.WindowSizeMsg {
	availableHeight := sizeMsg.Height - m.headerHeight
	newSizeMsg := tea.WindowSizeMsg{Width: sizeMsg.Width, Height: availableHeight}
	return newSizeMsg
}

func (t *stationsTabBase) initInfoModel(m *Model, msg toggleInfoMsg) tea.Cmd {
	t.listKeymap.setEnabled(false)
	t.infoModel.setSize(m.width, m.totHeight-m.headerHeight)
	return t.infoModel.Init(msg.station)
}

func createList(delegate *stationDelegate, width int, height int) list.Model {
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.InfiniteScrolling = true
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowFilter(true)
	l.SetStatusBarItemName("station", "stations")
	l.Styles.NoItems = noItemsStyle
	l.FilterInput.ShowSuggestions = true
	l.KeyMap.Quit.SetKeys("q")
	l.KeyMap.PrevPage.SetKeys("pgup", "ctrl+b")
	l.KeyMap.PrevPage.SetHelp("ctrl+b/pgup", "prev page")
	l.KeyMap.NextPage.SetKeys("pgdown", "ctrl+f")
	l.KeyMap.NextPage.SetHelp("ctrl+f/pgdn", "next page")
	h, v := docStyle.GetFrameSize()
	l.SetSize(width-h, height-v)

	l.Help.ShortSeparator = "   "
	l.Help.Styles = helpStyles()
	l.Styles.HelpStyle = helpStyle

	textInputSyle(&l.FilterInput, "Filter:       ", "name")
	return l
}
